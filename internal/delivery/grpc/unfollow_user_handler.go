package follow_grpc

import (
	"context"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-relation-service/internal/custom_errors"
)

type UnfollowDeleter interface {
	Unfollow(ctx context.Context, followerID, followeeID int64) error
}

type UnfollowHandler struct {
	pb.UnimplementedRelationServiceServer
	relationService UnfollowDeleter
	validate        *validator.Validate
}

func NewUnfollowHandler(relationService UnfollowDeleter, validate *validator.Validate) *UnfollowHandler {
	return &UnfollowHandler{
		relationService: relationService,
		validate:        validate,
	}
}

type UnfollowRequestInternal struct {
	FollowerID int64 `validate:"required,gt=0"`
	FolloweeID int64 `validate:"required,gt=0,nefield=FollowerID"`
}

func (h *UnfollowHandler) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	validationReq := &UnfollowRequestInternal{
		FollowerID: req.GetFollowerId(),
		FolloweeID: req.GetFolloweeId(),
	}

	if err := h.validate.Struct(validationReq); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	err := h.relationService.Unfollow(ctx, req.GetFollowerId(), req.GetFolloweeId())
	if err != nil {
		switch err {
		case custom_errors.ErrFollowRelationNotFound:
			return nil, status.Errorf(codes.NotFound, custom_errors.ErrFollowRelationNotFound.Error())
		case custom_errors.ErrUserNotFound:
			return nil, status.Errorf(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		case custom_errors.ErrSelfFollow:
			return nil, status.Errorf(codes.InvalidArgument, custom_errors.ErrSelfUnfollow.Error())
		default:
			return nil, status.Errorf(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	return &pb.UnfollowResponse{}, nil
}
