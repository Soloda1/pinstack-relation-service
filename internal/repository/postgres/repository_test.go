package repository_postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	repository_postgres "pinstack-relation-service/internal/repository/postgres"
	"pinstack-relation-service/mocks"
)

func createSuccessCommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("INSERT 0 1")
}

func createEmptyCommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("DELETE 0")
}

func setupMockRows(t *testing.T, ids []int64) *mocks.Rows {
	mockRows := mocks.NewRows(t)
	callsCount := len(ids)
	for i := 0; i < callsCount; i++ {
		mockRows.On("Next").Return(true).Once()
	}
	mockRows.On("Next").Return(false).Once()
	for _, id := range ids {
		mockRows.On("Scan", mock.AnythingOfType("*int64")).
			Run(func(args mock.Arguments) {
				arg := args.Get(0).(*int64)
				*arg = id
			}).
			Return(nil).
			Once()
	}
	mockRows.On("Err").Return(nil).Maybe()
	mockRows.On("Close").Return()
	return mockRows
}

func TestRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		followeeID  int64
		mockSetup   func(*mocks.PgDB)
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful follow",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createSuccessCommandTag(), nil)
			},
			wantErr: false,
		},
		{
			name:        "self follow error",
			followerID:  1,
			followeeID:  1,
			mockSetup:   func(db *mocks.PgDB) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrSelfFollow,
		},
		{
			name:       "database error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationCreateFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := repository_postgres.NewFollowRepository(mockDB, log)
			err := repo.Create(context.Background(), tt.followerID, tt.followeeID)
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

func TestRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		followeeID  int64
		mockSetup   func(*mocks.PgDB)
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful unfollow",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createSuccessCommandTag(), nil)
			},
			wantErr: false,
		},
		{
			name:       "relation not found",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createEmptyCommandTag(), nil)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationNotFound,
		},
		{
			name:       "database error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrFollowRelationDeleteFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := repository_postgres.NewFollowRepository(mockDB, log)
			err := repo.Delete(context.Background(), tt.followerID, tt.followeeID)
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

func TestRepository_GetFollowers(t *testing.T) {
	tests := []struct {
		name        string
		followeeID  int64
		mockSetup   func(*mocks.PgDB)
		want        []int64
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful get followers",
			followeeID: 1,
			mockSetup: func(db *mocks.PgDB) {
				rows := setupMockRows(t, []int64{2, 3, 4})
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(rows, nil)
			},
			want:    []int64{2, 3, 4},
			wantErr: false,
		},
		{
			name:       "empty followers list",
			followeeID: 1,
			mockSetup: func(db *mocks.PgDB) {
				rows := setupMockRows(t, []int64{})
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(rows, nil)
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name:       "database query error",
			followeeID: 1,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(nil, errors.New("db error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:       "scan error",
			followeeID: 1,
			mockSetup: func(db *mocks.PgDB) {
				mockRows := mocks.NewRows(t)
				mockRows.On("Next").Return(true).Once()
				mockRows.On("Scan", mock.AnythingOfType("*int64")).Return(errors.New("scan error"))
				mockRows.On("Close").Return()
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRows, nil)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseScan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := repository_postgres.NewFollowRepository(mockDB, log)
			got, err := repo.GetFollowers(context.Background(), tt.followeeID)
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

func TestRepository_GetFollowees(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		mockSetup   func(*mocks.PgDB)
		want        []int64
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful get followees",
			followerID: 1,
			mockSetup: func(db *mocks.PgDB) {
				rows := setupMockRows(t, []int64{2, 3, 4})
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(rows, nil)
			},
			want:    []int64{2, 3, 4},
			wantErr: false,
		},
		{
			name:       "empty followees list",
			followerID: 1,
			mockSetup: func(db *mocks.PgDB) {
				rows := setupMockRows(t, []int64{})
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(rows, nil)
			},
			want:    []int64{},
			wantErr: false,
		},
		{
			name:       "database query error",
			followerID: 1,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(nil, errors.New("db error"))
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := repository_postgres.NewFollowRepository(mockDB, log)
			got, err := repo.GetFollowees(context.Background(), tt.followerID)
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

func TestRepository_Exists(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		followeeID  int64
		mockSetup   func(*mocks.PgDB)
		want        bool
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "relation exists",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan", mock.AnythingOfType("*bool")).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*bool)
					*arg = true
				}).Return(nil)
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockRow)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:       "relation does not exist",
			followerID: 1,
			followeeID: 3,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan", mock.AnythingOfType("*bool")).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*bool)
					*arg = false
				}).Return(nil)
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockRow)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:       "database error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan", mock.AnythingOfType("*bool")).Return(errors.New("db error"))
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockRow)
			},
			want:        false,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := repository_postgres.NewFollowRepository(mockDB, log)
			got, err := repo.Exists(context.Background(), tt.followerID, tt.followeeID)
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
