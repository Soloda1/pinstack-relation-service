package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/mocks"
)

func TestService_Follow(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		followeeID  int64
		mockSetup   func(*mocks.FollowRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful follow",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(false, nil)
				repo.On("Create", mock.Anything, int64(1), int64(2)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "self follow error",
			followerID:  1,
			followeeID:  1,
			mockSetup:   func(repo *mocks.FollowRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrSelfFollow,
		},
		{
			name:       "relationship already exists",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(true, nil)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationExists,
		},
		{
			name:       "exists check error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(false, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:       "create relationship error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(false, nil)
				repo.On("Create", mock.Anything, int64(1), int64(2)).Return(custom_errors.ErrFollowRelationCreateFail)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationCreateFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewFollowRepository(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewFollowService(log, mockRepo)
			err := service.Follow(context.Background(), tt.followerID, tt.followeeID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Unfollow(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		followeeID  int64
		mockSetup   func(*mocks.FollowRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful unfollow",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(true, nil)
				repo.On("Delete", mock.Anything, int64(1), int64(2)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "self unfollow error",
			followerID:  1,
			followeeID:  1,
			mockSetup:   func(repo *mocks.FollowRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrSelfFollow,
		},
		{
			name:       "relationship doesn't exist",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(false, nil)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationNotFound,
		},
		{
			name:       "exists check error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(false, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:       "delete relationship error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("Exists", mock.Anything, int64(1), int64(2)).Return(true, nil)
				repo.On("Delete", mock.Anything, int64(1), int64(2)).Return(custom_errors.ErrFollowRelationDeleteFail)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationDeleteFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewFollowRepository(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewFollowService(log, mockRepo)
			err := service.Unfollow(context.Background(), tt.followerID, tt.followeeID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetFollowers(t *testing.T) {
	tests := []struct {
		name        string
		followeeID  int64
		mockSetup   func(*mocks.FollowRepository)
		want        []int64
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful get followers",
			followeeID: 1,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1)).Return([]int64{2, 3, 4}, nil)
			},
			want:    []int64{2, 3, 4},
			wantErr: false,
		},
		{
			name:       "empty followers list",
			followeeID: 1,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1)).Return([]int64{}, nil)
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name:       "database error",
			followeeID: 1,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1)).Return(nil, custom_errors.ErrDatabaseQuery)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewFollowRepository(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewFollowService(log, mockRepo)
			got, err := service.GetFollowers(context.Background(), tt.followeeID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_GetFollowees(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		mockSetup   func(*mocks.FollowRepository)
		want        []int64
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful get followees",
			followerID: 1,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1)).Return([]int64{2, 3, 4}, nil)
			},
			want:    []int64{2, 3, 4},
			wantErr: false,
		},
		{
			name:       "empty followees list",
			followerID: 1,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1)).Return([]int64{}, nil)
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name:       "database error",
			followerID: 1,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1)).Return(nil, custom_errors.ErrDatabaseQuery)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewFollowRepository(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := NewFollowService(log, mockRepo)
			got, err := service.GetFollowees(context.Background(), tt.followerID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
