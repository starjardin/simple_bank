package gapi

import (
	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		Email:             user.Email,
		FullName:          user.FullName,
		PasswordChangedAt: timestamppb.New(user.PasswordChangeAt),
		CreatedAd:         timestamppb.New(user.CreatedAt),
	}
}
