package repository_postgres

import (
	"context"
	"log/slog"
	model "pinstack-relation-service/internal/domain/models"
	ports "pinstack-relation-service/internal/domain/ports/output"
	"time"

	"github.com/soloda1/pinstack-proto-definitions/custom_errors"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	log     ports.Logger
	db      PgDB
	metrics ports.MetricsProvider
}

func NewFollowRepository(db PgDB, log ports.Logger, metrics ports.MetricsProvider) *Repository {
	return &Repository{db: db, log: log, metrics: metrics}
}

func (r *Repository) Create(ctx context.Context, followerID, followeeID int64) (follower model.Follower, err error) {
	start := time.Now()
	defer func() {
		r.metrics.IncrementDatabaseQueries("create_follow_relation", err == nil)
		r.metrics.RecordDatabaseQueryDuration("create_follow_relation", time.Since(start))
	}()

	r.log.Info("Creating follow relation", slog.Int64("follower_id", followerID), slog.Int64("followee_id", followeeID))

	if followerID == followeeID {
		r.log.Error("Attempt to follow yourself",
			slog.Int64("user_id", followerID))
		return model.Follower{}, custom_errors.ErrSelfFollow
	}

	args := pgx.NamedArgs{
		"follower_id": followerID,
		"followee_id": followeeID,
	}

	query := `
		INSERT INTO followers (follower_id, followee_id, created_at)
		VALUES (@follower_id, @followee_id, NOW())
		ON CONFLICT (follower_id, followee_id) DO NOTHING
		RETURNING id, follower_id, followee_id, created_at
	`

	var followerData model.Follower
	err = r.db.QueryRow(ctx, query, args).Scan(&followerData.ID, &followerData.FollowerID, &followerData.FolloweeID, &followerData.CreatedAt)
	if err != nil {
		r.log.Error("Failed to create follow relation",
			slog.Int64("follower_id", followerID),
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return model.Follower{}, custom_errors.ErrFollowRelationCreateFail
	}

	r.log.Info("Follow relation created successfully",
		slog.Int64("follower_id", followerID),
		slog.Int64("followee_id", followeeID))
	return followerData, nil
}

func (r *Repository) Delete(ctx context.Context, followerID, followeeID int64) (err error) {
	start := time.Now()
	defer func() {
		r.metrics.IncrementDatabaseQueries("delete_follow_relation", err == nil)
		r.metrics.RecordDatabaseQueryDuration("delete_follow_relation", time.Since(start))
	}()

	r.log.Info("Deleting follow relation", slog.Int64("follower_id", followerID), slog.Int64("followee_id", followeeID))

	args := pgx.NamedArgs{
		"follower_id": followerID,
		"followee_id": followeeID,
	}

	query := `
		DELETE FROM followers 
		WHERE follower_id = @follower_id AND followee_id = @followee_id
	`

	result, err := r.db.Exec(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to delete follow relation",
			slog.Int64("follower_id", followerID),
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return custom_errors.ErrFollowRelationDeleteFail
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.log.Warn("Follow relation not found",
			slog.Int64("follower_id", followerID),
			slog.Int64("followee_id", followeeID))
		return custom_errors.ErrFollowRelationNotFound
	}

	r.log.Info("Follow relation deleted successfully",
		slog.Int64("follower_id", followerID),
		slog.Int64("followee_id", followeeID))
	return nil
}

func (r *Repository) GetFollowers(ctx context.Context, followeeID int64, limit, offset int32) (followers []int64, total int64, err error) {
	start := time.Now()
	defer func() {
		r.metrics.IncrementDatabaseQueries("get_followers", err == nil)
		r.metrics.RecordDatabaseQueryDuration("get_followers", time.Since(start))
	}()

	r.log.Info("Getting followers", slog.Int64("followee_id", followeeID))

	args := pgx.NamedArgs{
		"followee_id": followeeID,
		"limit":       limit,
		"offset":      offset,
	}

	query := `
		SELECT 
			follower_id,
			COUNT(*) OVER() as total_count
		FROM followers 
		WHERE followee_id = @followee_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to query followers",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}
	defer rows.Close()

	followersList := make([]int64, 0)
	var totalCount int64

	for rows.Next() {
		var followerID int64
		if err := rows.Scan(&followerID, &totalCount); err != nil {
			r.log.Error("Failed to scan follower row",
				slog.Int64("followee_id", followeeID),
				slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseQuery
		}
		followersList = append(followersList, followerID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during followers iteration",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}

	if len(followersList) == 0 {
		countArgs := pgx.NamedArgs{
			"followee_id": followeeID,
		}

		countQuery := `SELECT COUNT(*) FROM followers WHERE followee_id = @followee_id`
		err := r.db.QueryRow(ctx, countQuery, countArgs).Scan(&totalCount)
		if err != nil {
			r.log.Error("Failed to count followers for empty result",
				slog.Int64("followee_id", followeeID),
				slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseQuery
		}
	}

	r.log.Info("Successfully retrieved followers",
		slog.Int64("followee_id", followeeID),
		slog.Int("count", len(followersList)),
		slog.Int64("total", totalCount))

	return followersList, totalCount, nil
}

func (r *Repository) GetFollowees(ctx context.Context, followerID int64, limit, offset int32) (followees []int64, total int64, err error) {
	start := time.Now()
	defer func() {
		r.metrics.IncrementDatabaseQueries("get_followees", err == nil)
		r.metrics.RecordDatabaseQueryDuration("get_followees", time.Since(start))
	}()

	r.log.Info("Getting followees", slog.Int64("follower_id", followerID))

	args := pgx.NamedArgs{
		"follower_id": followerID,
		"limit":       limit,
		"offset":      offset,
	}

	query := `
		SELECT 
			followee_id,
			COUNT(*) OVER() as total_count
		FROM followers 
		WHERE follower_id = @follower_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to query followees",
			slog.Int64("follower_id", followerID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}
	defer rows.Close()

	followeesList := make([]int64, 0)
	var totalCount int64

	for rows.Next() {
		var followeeID int64
		if err := rows.Scan(&followeeID, &totalCount); err != nil {
			r.log.Error("Failed to scan followee row",
				slog.Int64("follower_id", followerID),
				slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseQuery
		}
		followeesList = append(followeesList, followeeID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during followees iteration",
			slog.Int64("follower_id", followerID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}

	if len(followeesList) == 0 {
		countArgs := pgx.NamedArgs{
			"follower_id": followerID,
		}

		countQuery := `SELECT COUNT(*) FROM followers WHERE follower_id = @follower_id`
		err := r.db.QueryRow(ctx, countQuery, countArgs).Scan(&totalCount)
		if err != nil {
			r.log.Error("Failed to count followees for empty result",
				slog.Int64("follower_id", followerID),
				slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseQuery
		}
	}

	r.log.Info("Successfully retrieved followees",
		slog.Int64("follower_id", followerID),
		slog.Int("count", len(followeesList)),
		slog.Int64("total", totalCount))

	return followeesList, totalCount, nil
}

func (r *Repository) Exists(ctx context.Context, followerID, followeeID int64) (exists bool, err error) {
	start := time.Now()
	defer func() {
		r.metrics.IncrementDatabaseQueries("check_follow_relation_exists", err == nil)
		r.metrics.RecordDatabaseQueryDuration("check_follow_relation_exists", time.Since(start))
	}()

	r.log.Info("Checking if follow relation exists",
		slog.Int64("follower_id", followerID),
		slog.Int64("followee_id", followeeID))

	args := pgx.NamedArgs{
		"follower_id": followerID,
		"followee_id": followeeID,
	}

	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM followers 
			WHERE follower_id = @follower_id AND followee_id = @followee_id
		)
	`

	var existsResult bool
	err = r.db.QueryRow(ctx, query, args).Scan(&existsResult)
	if err != nil {
		r.log.Error("Failed to check follow relation existence",
			slog.Int64("follower_id", followerID),
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return false, custom_errors.ErrDatabaseQuery
	}

	r.log.Info("Follow relation check completed",
		slog.Int64("follower_id", followerID),
		slog.Int64("followee_id", followeeID),
		slog.Bool("exists", existsResult))
	return existsResult, nil
}
