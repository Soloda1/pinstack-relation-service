package repository_postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
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
		name           string
		followerID     int64
		followeeID     int64
		mockSetup      func(*mocks.PgDB)
		wantErr        bool
		expectedErr    error
		expectedResult model.Follower
	}{
		{
			name:       "successful follow",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*time.Time")).
					Run(func(args mock.Arguments) {
						idArg := args.Get(0).(*int64)
						followerIDArg := args.Get(1).(*int64)
						followeeIDArg := args.Get(2).(*int64)
						createdAtArg := args.Get(3).(*time.Time)

						*idArg = 1
						*followerIDArg = 1
						*followeeIDArg = 2
						*createdAtArg = time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
					}).
					Return(nil)

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			wantErr: false,
			expectedResult: model.Follower{
				ID:         1,
				FollowerID: 1,
				FolloweeID: 2,
				CreatedAt:  time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "self follow error",
			followerID:     1,
			followeeID:     1,
			mockSetup:      func(db *mocks.PgDB) {},
			wantErr:        true,
			expectedErr:    custom_errors.ErrSelfFollow,
			expectedResult: model.Follower{},
		},
		{
			name:       "database error",
			followerID: 1,
			followeeID: 2,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*time.Time")).
					Return(errors.New("db error"))

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			wantErr:        true,
			expectedErr:    custom_errors.ErrFollowRelationCreateFail,
			expectedResult: model.Follower{},
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
			result, err := repo.Create(context.Background(), tt.followerID, tt.followeeID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.ID, result.ID)
				assert.Equal(t, tt.expectedResult.FollowerID, result.FollowerID)
				assert.Equal(t, tt.expectedResult.FolloweeID, result.FolloweeID)
				assert.Equal(t, tt.expectedResult.CreatedAt, result.CreatedAt)
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
		limit       int32
		offset      int32
		mockSetup   func(*mocks.PgDB)
		want        []int64
		wantTotal   int64
		wantErr     bool
		expectedErr error
		checkQuery  bool
	}{
		{
			name:       "successful get followers with default pagination",
			followeeID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 5
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT COUNT(*) 
		FROM followers 
		WHERE followee_id = @followee_id
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["followee_id"] == int64(1)
					})).Return(mockCountRow)

				// Mock data query
				rows := setupMockRows(t, []int64{2, 3, 4})
				db.On("Query",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT follower_id 
		FROM followers 
		WHERE followee_id = @followee_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["followee_id"] == int64(1) &&
							args["limit"] == int32(10) &&
							args["offset"] == int32(0)
					})).Return(rows, nil)
			},
			want:       []int64{2, 3, 4},
			wantTotal:  5,
			wantErr:    false,
			checkQuery: true,
		},
		{
			name:       "get followers with custom pagination",
			followeeID: 1,
			limit:      5,
			offset:     10,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 15
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT COUNT(*) 
		FROM followers 
		WHERE followee_id = @followee_id
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["followee_id"] == int64(1)
					})).Return(mockCountRow)

				// Mock data query
				rows := setupMockRows(t, []int64{6, 7})
				db.On("Query",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT follower_id 
		FROM followers 
		WHERE followee_id = @followee_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["followee_id"] == int64(1) &&
							args["limit"] == int32(5) &&
							args["offset"] == int32(10)
					})).Return(rows, nil)
			},
			want:       []int64{6, 7},
			wantTotal:  15,
			wantErr:    false,
			checkQuery: true,
		},
		{
			name:       "empty followers list",
			followeeID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 0
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)

				// Mock data query
				rows := setupMockRows(t, []int64{})
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(rows, nil)
			},
			want:      []int64{},
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:       "database query error",
			followeeID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query error
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).Return(errors.New("db error"))
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)
			},
			want:        nil,
			wantTotal:   0,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:       "scan error",
			followeeID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query success
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 1
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)

				// Mock data query with scan error
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
			wantTotal:   0,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseScan,
		},
		{
			name:       "data query error",
			followeeID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query success
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 1
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)

				// Mock data query error
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(nil, errors.New("db error"))
			},
			want:        nil,
			wantTotal:   0,
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
			got, total, err := repo.GetFollowers(context.Background(), tt.followeeID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestRepository_GetFollowees(t *testing.T) {
	tests := []struct {
		name        string
		followerID  int64
		limit       int32
		offset      int32
		mockSetup   func(*mocks.PgDB)
		want        []int64
		wantTotal   int64
		wantErr     bool
		expectedErr error
		checkQuery  bool
	}{
		{
			name:       "successful get followees with default pagination",
			followerID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 5
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT COUNT(*) 
		FROM followers 
		WHERE follower_id = @follower_id
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["follower_id"] == int64(1)
					})).Return(mockCountRow)

				// Mock data query
				rows := setupMockRows(t, []int64{2, 3, 4})
				db.On("Query",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT followee_id 
		FROM followers 
		WHERE follower_id = @follower_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["follower_id"] == int64(1) &&
							args["limit"] == int32(10) &&
							args["offset"] == int32(0)
					})).Return(rows, nil)
			},
			want:       []int64{2, 3, 4},
			wantTotal:  5,
			wantErr:    false,
			checkQuery: true,
		},
		{
			name:       "get followees with custom pagination",
			followerID: 1,
			limit:      5,
			offset:     10,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 15
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT COUNT(*) 
		FROM followers 
		WHERE follower_id = @follower_id
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["follower_id"] == int64(1)
					})).Return(mockCountRow)

				// Mock data query
				rows := setupMockRows(t, []int64{8, 9})
				db.On("Query",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return query == `
		SELECT followee_id 
		FROM followers 
		WHERE follower_id = @follower_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["follower_id"] == int64(1) &&
							args["limit"] == int32(5) &&
							args["offset"] == int32(10)
					})).Return(rows, nil)
			},
			want:       []int64{8, 9},
			wantTotal:  15,
			wantErr:    false,
			checkQuery: true,
		},
		{
			name:       "empty followees list",
			followerID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 0
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)

				// Mock data query
				rows := setupMockRows(t, []int64{})
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(rows, nil)
			},
			want:      []int64{},
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:       "database query error",
			followerID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query error
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).Return(errors.New("db error"))
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)
			},
			want:        nil,
			wantTotal:   0,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:       "data query error",
			followerID: 1,
			limit:      10,
			offset:     0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query success
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int64")).
					Run(func(args mock.Arguments) {
						arg := args.Get(0).(*int64)
						*arg = 1
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockCountRow)

				// Mock data query error
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(nil, errors.New("db error"))
			},
			want:        nil,
			wantTotal:   0,
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
			got, total, err := repo.GetFollowees(context.Background(), tt.followerID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantTotal, total)
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
