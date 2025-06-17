package follow_grpc_test

import (
	"context"
	"errors"
	follow_grpc "pinstack-relation-service/internal/delivery/grpc"
	"testing"

	"github.com/go-playground/validator/v10"
	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/relation/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/mocks"
)

func TestFollowHandler_Follow(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.FollowRequest
		mockSetup      func(*mocks.FollowService)
		wantErr        bool
		expectedCode   codes.Code
		expectedErrMsg string
	}{
		{
			name: "successful follow",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 2,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("Follow", mock.Anything, int64(1), int64(2)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "validation error - follower ID zero",
			req: &pb.FollowRequest{
				FollowerId: 0,
				FolloweeId: 2,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "invalid request",
		},
		{
			name: "validation error - followee ID zero",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 0,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "invalid request",
		},
		{
			name: "validation error - self follow",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 1,
			},
			mockSetup:      func(mockService *mocks.FollowService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "invalid request",
		},
		{
			name: "self follow error from service",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 2,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("Follow", mock.Anything, int64(1), int64(2)).Return(custom_errors.ErrSelfFollow)
			},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "cannot follow yourself",
		},
		{
			name: "already following error",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 2,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("Follow", mock.Anything, int64(1), int64(2)).Return(custom_errors.ErrAlreadyFollowing)
			},
			wantErr:        true,
			expectedCode:   codes.AlreadyExists,
			expectedErrMsg: "already following this user",
		},
		{
			name: "user not found error",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 2,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("Follow", mock.Anything, int64(1), int64(2)).Return(custom_errors.ErrUserNotFound)
			},
			wantErr:        true,
			expectedCode:   codes.NotFound,
			expectedErrMsg: "user not found",
		},
		{
			name: "database error",
			req: &pb.FollowRequest{
				FollowerId: 1,
				FolloweeId: 2,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("Follow", mock.Anything, int64(1), int64(2)).Return(errors.New("database error"))
			},
			wantErr:        true,
			expectedCode:   codes.Internal,
			expectedErrMsg: "failed to follow user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			mockService := mocks.NewFollowService(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := follow_grpc.NewFollowHandler(mockService, validate)
			resp, err := handler.Follow(context.Background(), tt.req)

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
			}

			mockService.AssertExpectations(t)
		})
	}
}
