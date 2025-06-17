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

type FolloweesGetter interface {
	GetFollowees(ctx context.Context, followerID int64, limit, page int32) ([]int64, error)
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	followeeIDs, err := h.relationService.GetFollowees(ctx, req.GetFollowerId(), req.GetLimit(), req.GetPage())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			return nil, status.Errorf(codes.NotFound, "user not found")
		case errors.Is(err, custom_errors.ErrDatabaseQuery):
			return nil, status.Errorf(codes.Internal, "failed to fetch followees")
		default:
			return nil, status.Errorf(codes.Internal, "failed to get followees: %v", err)
		}
	}

	return &pb.GetFolloweesResponse{
		FolloweeIds: followeeIDs,
	}, nil
}
