package grpc

import (
	"context"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-relation-service/internal/custom_errors"
)

type FollowersGetter interface {
	GetFollowers(ctx context.Context, followeeID int64, limit, page int32) ([]int64, error)
}

type GetFollowersHandler struct {
	pb.UnimplementedRelationServiceServer
	relationService FollowersGetter
	validate        *validator.Validate
}

func NewGetFollowersHandler(relationService FollowersGetter, validate *validator.Validate) *GetFollowersHandler {
	return &GetFollowersHandler{
		relationService: relationService,
		validate:        validate,
	}
}

type GetFollowersRequestInternal struct {
	FolloweeID int64 `validate:"required,gt=0"`
	Limit      int32 `validate:"required,gt=0,lte=100"`
	Page       int32 `validate:"required,gte=1"`
}

func (h *GetFollowersHandler) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	validationReq := &GetFollowersRequestInternal{
		FolloweeID: req.GetFolloweeId(),
		Limit:      req.GetLimit(),
		Page:       req.GetPage(),
	}

	if err := h.validate.Struct(validationReq); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	followerIDs, err := h.relationService.GetFollowers(ctx, req.GetFolloweeId(), req.GetLimit(), req.GetPage())
	if err != nil {
		switch err {
		case custom_errors.ErrUserNotFound:
			return nil, status.Errorf(codes.NotFound, "user not found")
		case custom_errors.ErrDatabaseQuery:
			return nil, status.Errorf(codes.Internal, "failed to fetch followers")
		default:
			return nil, status.Errorf(codes.Internal, "failed to get followers: %v", err)
		}
	}

	return &pb.GetFollowersResponse{
		FollowerIds: followerIDs,
	}, nil
}
