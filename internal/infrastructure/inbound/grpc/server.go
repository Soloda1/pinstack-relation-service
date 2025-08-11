package follow_grpc

import (
	"fmt"
	"log/slog"
	"net"
	ports "pinstack-relation-service/internal/domain/ports/output"
	"pinstack-relation-service/internal/infrastructure/inbound/middleware"
	"runtime/debug"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"google.golang.org/grpc"
)

type Server struct {
	followGRPCService *FollowGRPCService
	server            *grpc.Server
	address           string
	port              int
	log               ports.Logger
}

func NewServer(grpcServer *FollowGRPCService, address string, port int, log ports.Logger) *Server {
	return &Server{
		followGRPCService: grpcServer,
		address:           address,
		port:              port,
		log:               log,
	}
}

func (s *Server) Run() error {
	address := fmt.Sprintf("%s:%d", s.address, s.port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			s.log.Error("panic recovered", slog.Any("panic", p), slog.String("stack", string(debug.Stack())))
			return status.Errorf(codes.Internal, "internal server error")
		}),
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.UnaryLoggerInterceptor(s.log),
			grpc_recovery.UnaryServerInterceptor(opts...),
		)),
	)

	pb.RegisterRelationServiceServer(s.server, s.followGRPCService)

	s.log.Info("Starting gRPC server", slog.Int("port", s.port))
	return s.server.Serve(lis)
}

func (s *Server) Shutdown() error {
	if s.server != nil {
		s.server.GracefulStop()
	}
	return nil
}
