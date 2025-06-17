package follow_grpc

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-relation-service/internal/custom_errors"
)

type FollowCreator interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
}

type FollowHandler struct {
	pb.UnimplementedRelationServiceServer
	relationService FollowCreator
	validate        *validator.Validate
}

func NewFollowHandler(relationService FollowCreator, validate *validator.Validate) *FollowHandler {
	return &FollowHandler{
		relationService: relationService,
		validate:        validate,
	}
}

type FollowRequestInternal struct {
	FollowerID int64 `validate:"required,gt=0"`
	FolloweeID int64 `validate:"required,gt=0,nefield=FollowerID"`
}

func (h *FollowHandler) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	validationReq := &FollowRequestInternal{
		FollowerID: req.GetFollowerId(),
		FolloweeID: req.GetFolloweeId(),
	}

	if err := h.validate.Struct(validationReq); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	err := h.relationService.Follow(ctx, req.GetFollowerId(), req.GetFolloweeId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrSelfFollow):
			return nil, status.Errorf(codes.InvalidArgument, "cannot follow yourself")
		case errors.Is(err, custom_errors.ErrAlreadyFollowing):
			return nil, status.Errorf(codes.AlreadyExists, "already following this user")
		case errors.Is(err, custom_errors.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to follow user: %v", err)
		}
	}

	return &pb.FollowResponse{}, nil
}
