package follow_grpc

import (
	"context"
	"github.com/go-playground/validator/v10"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/service"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
)

var validate = validator.New()

type FollowGRPCService struct {
	pb.UnimplementedRelationServiceServer
	relationService     service.FollowService
	log                 *logger.Logger
	followHandler       *FollowHandler
	unfollowHandler     *UnfollowHandler
	getFollowersHandler *GetFollowersHandler
	getFolloweesHandler *GetFolloweesHandler
}

func NewFollowGRPCService(relationService service.FollowService, log *logger.Logger) *FollowGRPCService {
	followHandler := NewFollowHandler(relationService, validate)
	unfollowHandler := NewUnfollowHandler(relationService, validate)
	getFollowersHandler := NewGetFollowersHandler(relationService, validate)
	getFolloweesHandler := NewGetFolloweesHandler(relationService, validate)

	return &FollowGRPCService{
		relationService:     relationService,
		log:                 log,
		followHandler:       followHandler,
		unfollowHandler:     unfollowHandler,
		getFollowersHandler: getFollowersHandler,
		getFolloweesHandler: getFolloweesHandler,
	}
}

func (s *FollowGRPCService) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	return s.followHandler.Follow(ctx, req)
}

func (s *FollowGRPCService) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	return s.unfollowHandler.Unfollow(ctx, req)
}

func (s *FollowGRPCService) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	return s.getFollowersHandler.GetFollowers(ctx, req)
}

func (s *FollowGRPCService) GetFollowees(ctx context.Context, req *pb.GetFolloweesRequest) (*pb.GetFolloweesResponse, error) {
	return s.getFolloweesHandler.GetFollowees(ctx, req)
}
