package repository_postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
)

type Repository struct {
	log *logger.Logger
	db  PgDB
}

func NewFollowRepository(db PgDB, log *logger.Logger) *Repository {
	return &Repository{db: db, log: log}
}

func (r *Repository) Create(ctx context.Context, followerID, followeeID int64) error {
	r.log.Info("Creating follow relation", slog.Int64("follower_id", followerID), slog.Int64("followee_id", followeeID))

	if followerID == followeeID {
		r.log.Error("Attempt to follow yourself",
			slog.Int64("user_id", followerID))
		return custom_errors.ErrSelfFollow
	}

	args := pgx.NamedArgs{
		"follower_id": followerID,
		"followee_id": followeeID,
	}

	query := `
		INSERT INTO followers (follower_id, followee_id, created_at)
		VALUES (@follower_id, @followee_id, NOW())
		ON CONFLICT (follower_id, followee_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to create follow relation",
			slog.Int64("follower_id", followerID),
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return custom_errors.ErrFollowRelationCreateFail
	}

	r.log.Info("Follow relation created successfully",
		slog.Int64("follower_id", followerID),
		slog.Int64("followee_id", followeeID))
	return nil
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

func (r *Repository) GetFollowers(ctx context.Context, followeeID int64) ([]int64, error) {
	r.log.Info("Getting followers", slog.Int64("followee_id", followeeID))

	args := pgx.NamedArgs{
		"followee_id": followeeID,
	}

	query := `
		SELECT follower_id 
		FROM followers 
		WHERE followee_id = @followee_id
	`

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to query followers",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return nil, custom_errors.ErrDatabaseQuery
	}
	defer rows.Close()

	followers := make([]int64, 0)
	for rows.Next() {
		var followerID int64
		if err := rows.Scan(&followerID); err != nil {
			r.log.Error("Failed to scan follower row",
				slog.Int64("followee_id", followeeID),
				slog.String("error", err.Error()))
			return nil, custom_errors.ErrDatabaseScan
		}
		followers = append(followers, followerID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during followers iteration",
			slog.Int64("followee_id", followeeID),
			slog.String("error", err.Error()))
		return nil, custom_errors.ErrDatabaseQuery
	}

	r.log.Info("Successfully retrieved followers",
		slog.Int64("followee_id", followeeID),
		slog.Int("count", len(followers)))
	return followers, nil
}

func (r *Repository) GetFollowees(ctx context.Context, followerID int64) ([]int64, error) {
	r.log.Info("Getting followees", slog.Int64("follower_id", followerID))

	args := pgx.NamedArgs{
		"follower_id": followerID,
	}

	query := `
		SELECT followee_id 
		FROM followers 
		WHERE follower_id = @follower_id
	`

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		r.log.Error("Failed to query followees",
			slog.Int64("follower_id", followerID),
			slog.String("error", err.Error()))
		return nil, custom_errors.ErrDatabaseQuery
	}
	defer rows.Close()

	followees := make([]int64, 0)
	for rows.Next() {
		var followeeID int64
		if err := rows.Scan(&followeeID); err != nil {
			r.log.Error("Failed to scan followee row",
				slog.Int64("follower_id", followerID),
				slog.String("error", err.Error()))
			return nil, custom_errors.ErrDatabaseScan
		}
		followees = append(followees, followeeID)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during followees iteration",
			slog.Int64("follower_id", followerID),
			slog.String("error", err.Error()))
		return nil, custom_errors.ErrDatabaseQuery
	}

	r.log.Info("Successfully retrieved followees",
		slog.Int64("follower_id", followerID),
		slog.Int("count", len(followees)))
	return followees, nil
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
