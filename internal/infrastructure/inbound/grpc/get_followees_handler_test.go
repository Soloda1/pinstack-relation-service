package follow_grpc_test

import (
	"context"
	"errors"
	model "pinstack-relation-service/internal/domain/models"
	follow_grpc "pinstack-relation-service/internal/infrastructure/inbound/grpc"
	"pinstack-relation-service/internal/infrastructure/utils"
	"testing"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-relation-service/mocks"

	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
)

func TestGetFolloweesHandler_GetFollowees(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.GetFolloweesRequest
		mockSetup      func(*mocks.FollowService)
		wantErr        bool
		expectedCode   codes.Code
		expectedErrMsg string
		expectedUsers  []*model.User
		expectedTotal  int64
	}{
		{
			name: "successful get followees",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				users := []*model.User{
					{ID: 2, Username: "user2", AvatarURL: utils.StringPtr("avatar2.jpg")},
					{ID: 3, Username: "user3", AvatarURL: nil},
					{ID: 4, Username: "user4", AvatarURL: utils.StringPtr("avatar4.jpg")},
				}
				mockService.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(1)).
					Return(users, int64(15), nil)
			},
			wantErr: false,
			expectedUsers: []*model.User{
				{ID: 2, Username: "user2", AvatarURL: utils.StringPtr("avatar2.jpg")},
				{ID: 3, Username: "user3", AvatarURL: nil},
				{ID: 4, Username: "user4", AvatarURL: utils.StringPtr("avatar4.jpg")},
			},
			expectedTotal: 15,
		},
		{
			name: "empty result",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]*model.User{}, int64(0), nil)
			},
			wantErr:       false,
			expectedUsers: []*model.User{},
			expectedTotal: 0,
		},
		{
			name: "validation error - follower ID zero",
			req: &pb.GetFolloweesRequest{
				FollowerId: 0,
				Limit:      10,
				Page:       1,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "validation error - limit zero",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      0,
				Page:       1,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "validation error - limit too large",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      101,
				Page:       1,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "validation error - page zero",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      10,
				Page:       0,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: custom_errors.ErrValidationFailed.Error(),
		},
		{
			name: "user not found error",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]*model.User{}, int64(0), custom_errors.ErrUserNotFound)
			},
			wantErr:        true,
			expectedCode:   codes.NotFound,
			expectedErrMsg: custom_errors.ErrUserNotFound.Error(),
		},
		{
			name: "database query error",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]*model.User{}, int64(0), custom_errors.ErrDatabaseQuery)
			},
			wantErr:        true,
			expectedCode:   codes.Internal,
			expectedErrMsg: custom_errors.ErrDatabaseQuery.Error(),
		},
		{
			name: "generic error",
			req: &pb.GetFolloweesRequest{
				FollowerId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]*model.User{}, int64(0), errors.New("unexpected error"))
			},
			wantErr:        true,
			expectedCode:   codes.Internal,
			expectedErrMsg: custom_errors.ErrExternalServiceError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			mockService := mocks.NewFollowService(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := follow_grpc.NewGetFolloweesHandler(mockService, validate)
			resp, err := handler.GetFollowees(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				statusErr, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tt.expectedErrMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, len(tt.expectedUsers), len(resp.Followees))
				assert.Equal(t, tt.expectedTotal, resp.Total)

				for i, expectedUser := range tt.expectedUsers {
					assert.Equal(t, expectedUser.ID, resp.Followees[i].FollowerId)
					assert.Equal(t, expectedUser.Username, resp.Followees[i].Username)
					if expectedUser.AvatarURL != nil {
						assert.Equal(t, *expectedUser.AvatarURL, *resp.Followees[i].AvatarUrl)
					} else {
						assert.Nil(t, resp.Followees[i].AvatarUrl)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
