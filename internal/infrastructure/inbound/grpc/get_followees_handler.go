package follow_grpc

import (
	"context"
	"errors"
	model "pinstack-relation-service/internal/domain/models"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
)

type FolloweesGetter interface {
	GetFollowees(ctx context.Context, followerID int64, limit, page int32) ([]*model.User, int64, error)
}

type GetFolloweesHandler struct {
	pb.UnimplementedRelationServiceServer
	relationService FolloweesGetter
	validate        *validator.Validate
}

func NewGetFolloweesHandler(relationService FolloweesGetter, validate *validator.Validate) *GetFolloweesHandler {
	return &GetFolloweesHandler{
		relationService: relationService,
		validate:        validate,
	}
}

type GetFolloweesRequestInternal struct {
	FollowerID int64 `validate:"required,gt=0"`
	Limit      int32 `validate:"required,gt=0,lte=100"`
	Page       int32 `validate:"required,gte=1"`
}

func (h *GetFolloweesHandler) GetFollowees(ctx context.Context, req *pb.GetFolloweesRequest) (*pb.GetFolloweesResponse, error) {
	validationReq := &GetFolloweesRequestInternal{
		FollowerID: req.GetFollowerId(),
		Limit:      req.GetLimit(),
		Page:       req.GetPage(),
	}

	if err := h.validate.Struct(validationReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	followees, total, err := h.relationService.GetFollowees(ctx, req.GetFollowerId(), req.GetLimit(), req.GetPage())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		case errors.Is(err, custom_errors.ErrDatabaseQuery):
			return nil, status.Error(codes.Internal, custom_errors.ErrDatabaseQuery.Error())
		default:
			return nil, status.Error(codes.Internal, custom_errors.ErrExternalServiceError.Error())
		}
	}

	pbFollowees := make([]*pb.User, 0, len(followees))
	for _, followee := range followees {
		pbUser := &pb.User{
			FollowerId: followee.ID,
			Username:   followee.Username,
		}
		if followee.AvatarURL != nil {
			pbUser.AvatarUrl = followee.AvatarURL
		}
		pbFollowees = append(pbFollowees, pbUser)
	}

	return &pb.GetFolloweesResponse{
		Followees: pbFollowees,
		Total:     total,
	}, nil
}
