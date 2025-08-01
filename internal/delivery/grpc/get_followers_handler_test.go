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

func TestGetFollowersHandler_GetFollowers(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.GetFollowersRequest
		mockSetup      func(*mocks.FollowService)
		wantErr        bool
		expectedCode   codes.Code
		expectedErrMsg string
		expectedIDs    []int64
	}{
		{
			name: "successful get followers",
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]int64{2, 3, 4}, nil)
			},
			wantErr:     false,
			expectedIDs: []int64{2, 3, 4},
		},
		{
			name: "empty result",
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]int64{}, nil)
			},
			wantErr:     false,
			expectedIDs: []int64{},
		},
		{
			name: "validation error - followee ID zero",
			req: &pb.GetFollowersRequest{
				FolloweeId: 0,
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
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
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
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
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
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
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
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]int64{}, custom_errors.ErrUserNotFound)
			},
			wantErr:        true,
			expectedCode:   codes.NotFound,
			expectedErrMsg: custom_errors.ErrUserNotFound.Error(),
		},
		{
			name: "database query error",
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]int64{}, custom_errors.ErrDatabaseQuery)
			},
			wantErr:        true,
			expectedCode:   codes.Internal,
			expectedErrMsg: custom_errors.ErrDatabaseQuery.Error(),
		},
		{
			name: "generic error",
			req: &pb.GetFollowersRequest{
				FolloweeId: 1,
				Limit:      10,
				Page:       1,
			},
			mockSetup: func(mockService *mocks.FollowService) {
				mockService.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(1)).
					Return([]int64{}, errors.New("unexpected error"))
			},
			wantErr:        true,
			expectedCode:   codes.Internal,
			expectedErrMsg: custom_errors.ErrInternalServiceError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := validator.New()
			mockService := mocks.NewFollowService(t)

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := follow_grpc.NewGetFollowersHandler(mockService, validate)
			resp, err := handler.GetFollowers(context.Background(), tt.req)

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
				assert.Equal(t, len(tt.expectedIDs), len(resp.FollowerIds))
				assert.ElementsMatch(t, tt.expectedIDs, resp.FollowerIds)
			}

			mockService.AssertExpectations(t)
		})
	}
}
