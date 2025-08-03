package repository_postgres

import (
	"context"
	"log/slog"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	log *logger.Logger
	db  PgDB
}

func NewFollowRepository(db PgDB, log *logger.Logger) *Repository {
	return &Repository{db: db, log: log}
}

func (r *Repository) Create(ctx context.Context, followerID, followeeID int64) (model.Follower, error) {
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

	var follower model.Follower
	err := r.db.QueryRow(ctx, query, args).Scan(&follower.ID, &follower.FollowerID, &follower.FolloweeID, &follower.CreatedAt)
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
	return follower, nil
}

func (r *Repository) Delete(ctx context.Context, followerID, followeeID int64) error {
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

func (r *Repository) GetFollowers(ctx context.Context, followeeID int64, limit, offset int32) ([]int64, int64, error) {
	r.log.Info("Getting followers", slog.Int64("followee_id", followeeID))

	countArgs := pgx.NamedArgs{
		"followee_id": followeeID,
	}

	countQuery := `
		SELECT COUNT(*) 
		FROM followers 
		WHERE followee_id = @followee_id
	`

	var total int64
	err := r.db.QueryRow(ctx, countQuery, countArgs).Scan(&total)
	if err != nil {
		r.log.Error("Failed to count followers",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}

	args := pgx.NamedArgs{
		"followee_id": followeeID,
		"limit":       limit,
		"offset":      offset,
	}

	query := `
		SELECT follower_id 
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

	followers := make([]int64, 0)
	for rows.Next() {
		var followerID int64
		if err := rows.Scan(&followerID); err != nil {
			r.log.Error("Failed to scan follower row",
				slog.Int64("followee_id", followeeID),
				slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseScan
		}
		followers = append(followers, followerID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during followers iteration",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}

	r.log.Info("Successfully retrieved followers",
		slog.Int64("followee_id", followeeID),
		slog.Int("count", len(followers)),
		slog.Int64("total", total))
	return followers, total, nil
}

func (r *Repository) GetFollowees(ctx context.Context, followerID int64, limit, offset int32) ([]int64, int64, error) {
	r.log.Info("Getting followees", slog.Int64("follower_id", followerID))

	countArgs := pgx.NamedArgs{
		"follower_id": followerID,
	}

	countQuery := `
		SELECT COUNT(*) 
		FROM followers 
		WHERE follower_id = @follower_id
	`

	var total int64
	err := r.db.QueryRow(ctx, countQuery, countArgs).Scan(&total)
	if err != nil {
		r.log.Error("Failed to count followees",
			slog.Int64("follower_id", followerID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}

	args := pgx.NamedArgs{
		"follower_id": followerID,
		"limit":       limit,
		"offset":      offset,
	}

	query := `
		SELECT followee_id 
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

	followees := make([]int64, 0)
	for rows.Next() {
		var followeeID int64
		if err := rows.Scan(&followeeID); err != nil {
			r.log.Error("Failed to scan followee row",
				slog.Int64("follower_id", followerID),
				slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseScan
		}
		followees = append(followees, followeeID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during followees iteration",
			slog.Int64("follower_id", followerID),
			slog.String("error", err.Error()))
		return nil, 0, custom_errors.ErrDatabaseQuery
	}

	r.log.Info("Successfully retrieved followees",
		slog.Int64("follower_id", followerID),
		slog.Int("count", len(followees)),
		slog.Int64("total", total))
	return followees, total, nil
}

func (r *Repository) Exists(ctx context.Context, followerID, followeeID int64) (bool, error) {
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

	var exists bool
	err := r.db.QueryRow(ctx, query, args).Scan(&exists)
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
		slog.Bool("exists", exists))
	return exists, nil
}
