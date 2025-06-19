package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starjardin/simplebank/api"
	db "github.com/starjardin/simplebank/db/sqlc"
	_ "github.com/starjardin/simplebank/docs/statik"
	"github.com/starjardin/simplebank/gapi"
	"github.com/starjardin/simplebank/mail"
	"github.com/starjardin/simplebank/pb"
	"github.com/starjardin/simplebank/utils"
	"github.com/starjardin/simplebank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Info().Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)

	defer stop()

	connPool, err := pgxpool.New(ctx, config.DBSource)

	if err != nil {
		log.Info().Msg("cannot connect to db")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runTaskProcessor(ctx, waitGroup, config, redisOpt, store)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)

	err = waitGroup.Wait()

	if err != nil {
		log.Fatal().Err(err).Msgf("error from wait group")
	}
}

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config utils.Config,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
) {
	mailer := mail.NewGmailSender(
		config.EmailSenderName,
		config.EmailSenderAddress,
		config.EmailSenderPassword,
	)

	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("starting task processor...")

	if err := taskProcessor.Start(); err != nil {
		log.Info().Msg("cannot start task processor")
	}

	log.Info().Msg("task processor started successfully")

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down task processor...")

		taskProcessor.ShutDown()

		log.Info().Msg("task processor stopped successfully")
		return nil

	})

}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)

	if err != nil {
		log.Info().Msg("cannot create migration")
	}

	if err = migration.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Info().Msg("failed to run migration")
		}
	}
	log.Info().Msg("db migration completed successfully")
	migration.Close()
}

func runGatewayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config utils.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {

	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Info().Msg("cannot create server")
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Info().Msg("cannot register handle server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()

	if err != nil {
		log.Info().Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))

	mux.Handle("/swagger/", swaggerHandler)

	httpServer := &http.Server{
		Handler: gapi.HttpLogger(mux),
		Addr:    config.HTTPServerAddress,
	}

	waitGroup.Go(func() error {
		log.Printf("start gateway server at %s", httpServer.Addr)

		err = httpServer.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("Http gateway server failed to start")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down gateway server...")

		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to shutdown gateway server")
			return err
		}

		log.Info().Msg("gateway server stopped successfully")
		return nil
	})

}

func runGrpcServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config utils.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {

	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Info().Msg("cannot create server")
	}
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Info().Msg("cannot create listener")
	}

	waitGroup.Go(func() error {
		log.Printf("start gRPC server at %s", config.GRPCServerAddress)

		err = grpcServer.Serve(listener)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Error().Err(err).Msg("gRPC server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down gRPC server...")

		grpcServer.GracefulStop()

		log.Info().Msg("gRPC server stopped successfully")
		return nil
	})

}

func runGinServer(config utils.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Info().Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Info().Msg("cannot start server")
	}
}
