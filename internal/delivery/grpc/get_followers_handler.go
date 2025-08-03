package follow_grpc

import (
	"context"
	"errors"
	"pinstack-relation-service/internal/model"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-relation-service/internal/custom_errors"
)

type FollowersGetter interface {
	GetFollowers(ctx context.Context, followeeID int64, limit, page int32) ([]*model.User, int64, error)
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
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	followers, total, err := h.relationService.GetFollowers(ctx, req.GetFolloweeId(), req.GetLimit(), req.GetPage())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		case errors.Is(err, custom_errors.ErrDatabaseQuery):
			return nil, status.Error(codes.Internal, custom_errors.ErrDatabaseQuery.Error())
		default:
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	pbFollowers := make([]*pb.User, 0, len(followers))
	for _, follower := range followers {
		pbUser := &pb.User{
			FollowerId: follower.ID,
			Username:   follower.Username,
		}
		if follower.AvatarURL != nil {
			pbUser.AvatarUrl = follower.AvatarURL
		}
		pbFollowers = append(pbFollowers, pbUser)
	}

	return &pb.GetFollowersResponse{
		Followers: pbFollowers,
		Total:     total,
	}, nil
}
