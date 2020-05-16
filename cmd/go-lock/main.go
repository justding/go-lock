package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/stoex/go-lock/internal/config"
	pb "github.com/stoex/go-lock/internal/generated"
	"github.com/stoex/go-lock/pkg/service"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// Version denotes the program version
	Version string
	// BuildDate denotes the build date
	BuildDate     string
	configuration = config.NewManager()
	tls           = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile      = flag.String("cert_file", "", "The TLS cert file")
	keyFile       = flag.String("key_file", "", "The TLS key file")
	port          = flag.Int("port", 10000, "The server port")
	grpcServer    *grpc.Server
)

func main() {
	log.Printf("go-lock :: version %s :: build date %s", Version, BuildDate)
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))

		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		var opts []grpc.ServerOption

		if *tls {
			if *certFile == "" {
				*certFile = testdata.Path("server1.pem")
			}
			if *keyFile == "" {
				*keyFile = testdata.Path("server1.key")
			}
			creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
			if err != nil {
				log.Fatalf("Failed to generate credentials %v", err)
			}
			opts = []grpc.ServerOption{grpc.Creds(creds)}
		}

		grpcServer = grpc.NewServer(opts...)
		svc, err := service.NewLockService(configuration.Redlock.Clients)

		if err != nil {
			log.Fatalf("failed to create lock service: %v", err)
		}

		pb.RegisterLockServer(grpcServer, svc)
		return grpcServer.Serve(lis)
	})

	select {
	case <-interrupt:
		log.Println("received shutdown signal")
		break
	case <-ctx.Done():
		break
	}

	cancel()

	_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	err := g.Wait()

	if err != nil {
		log.Printf("server returning an error: %v", err.Error())
		os.Exit(2)
	}
}
