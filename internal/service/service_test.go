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
		name           string
		followeeID     int64
		limit          int32
		page           int32
		expectedLimit  int32
		expectedOffset int32
		mockSetup      func(*mocks.FollowRepository)
		want           []int64
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "successful get followers with default pagination",
			followeeID:     1,
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(0)).Return([]int64{2, 3, 4}, nil)
			},
			want:    []int64{2, 3, 4},
			wantErr: false,
		},
		{
			name:           "get second page of followers",
			followeeID:     1,
			limit:          5,
			page:           2,
			expectedLimit:  5,
			expectedOffset: 5,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(5), int32(5)).Return([]int64{6, 7, 8}, nil)
			},
			want:    []int64{6, 7, 8},
			wantErr: false,
		},
		{
			name:           "use default limit when limit is zero",
			followeeID:     1,
			limit:          0,
			page:           1,
			expectedLimit:  20, // defaultLimit
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(20), int32(0)).Return([]int64{2, 3}, nil)
			},
			want:    []int64{2, 3},
			wantErr: false,
		},
		{
			name:           "use default page when page is zero",
			followeeID:     1,
			limit:          10,
			page:           0,
			expectedLimit:  10,
			expectedOffset: 0, // (1-1)*10
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(0)).Return([]int64{2, 3}, nil)
			},
			want:    []int64{2, 3},
			wantErr: false,
		},
		{
			name:           "negative values for limit and page use defaults",
			followeeID:     1,
			limit:          -5,
			page:           -2,
			expectedLimit:  20, // defaultLimit
			expectedOffset: 0,  // (1-1)*20
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(20), int32(0)).Return([]int64{2, 3}, nil)
			},
			want:    []int64{2, 3},
			wantErr: false,
		},
		{
			name:           "empty followers list",
			followeeID:     1,
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(0)).Return([]int64{}, nil)
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name:           "database error",
			followeeID:     1,
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowers", mock.Anything, int64(1), int32(10), int32(0)).Return(nil, custom_errors.ErrDatabaseQuery)
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
			got, err := service.GetFollowers(context.Background(), tt.followeeID, tt.limit, tt.page)
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
		name           string
		followerID     int64
		limit          int32
		page           int32
		expectedLimit  int32
		expectedOffset int32
		mockSetup      func(*mocks.FollowRepository)
		want           []int64
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "successful get followees with default pagination",
			followerID:     1,
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(0)).Return([]int64{2, 3, 4}, nil)
			},
			want:    []int64{2, 3, 4},
			wantErr: false,
		},
		{
			name:           "get third page of followees",
			followerID:     1,
			limit:          3,
			page:           3,
			expectedLimit:  3,
			expectedOffset: 6,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(3), int32(6)).Return([]int64{7, 8, 9}, nil)
			},
			want:    []int64{7, 8, 9},
			wantErr: false,
		},
		{
			name:           "use default limit when limit is zero",
			followerID:     1,
			limit:          0,
			page:           1,
			expectedLimit:  20, // defaultLimit
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(20), int32(0)).Return([]int64{2, 3}, nil)
			},
			want:    []int64{2, 3},
			wantErr: false,
		},
		{
			name:           "use default page when page is zero",
			followerID:     1,
			limit:          10,
			page:           0,
			expectedLimit:  10,
			expectedOffset: 0, // (1-1)*10
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(0)).Return([]int64{2, 3}, nil)
			},
			want:    []int64{2, 3},
			wantErr: false,
		},
		{
			name:           "negative values for limit and page use defaults",
			followerID:     1,
			limit:          -5,
			page:           -2,
			expectedLimit:  20, // defaultLimit
			expectedOffset: 0,  // (1-1)*20
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(20), int32(0)).Return([]int64{2, 3}, nil)
			},
			want:    []int64{2, 3},
			wantErr: false,
		},
		{
			name:           "empty followees list",
			followerID:     1,
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(0)).Return([]int64{}, nil)
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name:           "database error",
			followerID:     1,
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
			mockSetup: func(repo *mocks.FollowRepository) {
				repo.On("GetFollowees", mock.Anything, int64(1), int32(10), int32(0)).Return(nil, custom_errors.ErrDatabaseQuery)
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
			got, err := service.GetFollowees(context.Background(), tt.followerID, tt.limit, tt.page)
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
