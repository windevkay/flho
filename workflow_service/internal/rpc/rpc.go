package rpc

import (
	pb "github.com/windevkay/flho/mailer_service/proto"
	"google.golang.org/grpc"
)

type Clients struct {
	MailerClient pb.MailerClient
}

type Connections struct {
	MailerConn *grpc.ClientConn
}

func (c *Connections) Close() {
	c.MailerConn.Close()
}

func GetClients(conn Connections) Clients {
	return Clients{
		MailerClient: pb.NewMailerClient(conn.MailerConn),
	}
}
