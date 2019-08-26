package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	v1 "github.com/medis/grpc/pkg/api/v1"
	"google.golang.org/grpc"
)

func RunServer(ctx context.Context, v1API v1.ToDoServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// Register service.
	server := grpc.NewServer()
	v1.RegisterToDoServiceServer(server, v1API)

	// Graceful shutdown.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("Shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	// Start gRPC server.
	return server.Serve(listen)
}
