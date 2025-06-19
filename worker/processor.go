package worker

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/hibiken/asynq"
	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/mail"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ShutDown()
	ProcessTaskSendVerifyEmail(
		ctx context.Context,
		task *asynq.Task,
	) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			QueueCritical: 10,
			QueueDefault:  5,
		},

		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Err(err).Str("type", task.Type()).
				Bytes("payload", task.Payload()).
				Msg("task processing failed")
		}),
		Logger: NewLogger(),
	})
	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) ShutDown() {
	processor.server.Shutdown()
}
