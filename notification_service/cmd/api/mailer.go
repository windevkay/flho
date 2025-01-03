package main

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/windevkay/flho/notification_service/proto"
)

func (s *server) SendWelcomeEmail(ctx context.Context, in *pb.WelcomeEmailRequest) (*empty.Empty, error) {
	err := s.mailer.Send(in.GetRecipient(), in.GetFile(), in.GetData())
	if err != nil {
		return nil, err
	}
	return new(empty.Empty), nil
}
