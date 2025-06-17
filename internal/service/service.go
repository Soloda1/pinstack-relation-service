package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
	"pinstack-relation-service/internal/repository"
	"pinstack-relation-service/internal/uow"
	"pinstack-relation-service/internal/utils"
	"time"
)

type Service struct {
	followRepo repository.FollowRepository
	uow        uow.UnitOfWork
	log        *logger.Logger
}

func NewFollowService(log *logger.Logger, followRepo repository.FollowRepository, uow uow.UnitOfWork) *Service {
	return &Service{
		log:        log,
		followRepo: followRepo,
		uow:        uow,
	}
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID int64) error {
	s.log.Info("Follow request received", slog.Int64("followerID", followerID), slog.Int64("followeeID", followeeID))

	if followerID == followeeID {
		return custom_errors.ErrSelfFollow
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error("Failed to start transaction", slog.String("error", err.Error()))
		return custom_errors.ErrDatabaseQuery
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	followRepo := tx.FollowRepository()
	outboxRepo := tx.OutboxRepository()

	exists, err := followRepo.Exists(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error checking follow existence", slog.String("error", err.Error()))
		return err
	}
	if exists {
		return custom_errors.ErrFollowRelationExists
	}

	follower, err := followRepo.Create(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error creating follow relationship", slog.String("error", err.Error()))
		return err
	}

	payload, err := json.Marshal(model.FollowCreatedPayload{
		FollowerID:  follower.FollowerID,
		FolloweeID:  follower.FolloweeID,
		Timestamptz: time.Now(),
	})
	if err != nil {
		s.log.Error("Failed to marshal payload", slog.String("error", err.Error()))
		return err
	}

	event := model.OutboxEvent{
		EventType:   model.EventTypeFollowCreated,
		Payload:     payload,
		AggregateID: follower.ID,
	}

	err = outboxRepo.AddEvent(ctx, event)
	if err != nil {
		s.log.Error("Error adding event to outbox", slog.String("error", err.Error()))
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.log.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return custom_errors.ErrDatabaseQuery
	}

	s.log.Info("Follow relationship created successfully", slog.Int64("followerID", followerID), slog.Int64("followeeID", followeeID))
	return nil
}

func (s *Service) Unfollow(ctx context.Context, followerID, followeeID int64) error {
	s.log.Info("Unfollow request received", slog.Int64("followerID", followerID), slog.Int64("followeeID", followeeID))

	if followerID == followeeID {
		return custom_errors.ErrSelfFollow
	}

	exists, err := s.followRepo.Exists(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error checking follow existence", slog.String("error", err.Error()))
		return err
	}
	if !exists {
		return custom_errors.ErrFollowRelationNotFound
	}

	err = s.followRepo.Delete(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error deleting follow relationship", slog.String("error", err.Error()))
		return err
	}

	s.log.Info("Follow relationship deleted successfully", slog.Int64("followerID", followerID), slog.Int64("followeeID", followeeID))
	return nil
}

func (s *Service) GetFollowers(ctx context.Context, followeeID int64, limit, page int32) ([]int64, error) {
	s.log.Info("GetFollowers request received", slog.Int64("followeeID", followeeID))
	limit, offset := utils.SetPaginationDefaults(limit, page)
	followers, err := s.followRepo.GetFollowers(ctx, followeeID, limit, offset)
	if err != nil {
		s.log.Error("Error getting followers", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("Followers retrieved successfully", slog.Int64("followeeID", followeeID), slog.Int("count", len(followers)))
	return followers, nil
}

func (s *Service) GetFollowees(ctx context.Context, followerID int64, limit, page int32) ([]int64, error) {
	s.log.Info("GetFollowees request received", slog.Int64("followerID", followerID))
	limit, offset := utils.SetPaginationDefaults(limit, page)
	followees, err := s.followRepo.GetFollowees(ctx, followerID, limit, offset)
	if err != nil {
		s.log.Error("Error getting followees", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("Followees retrieved successfully", slog.Int64("followerID", followerID), slog.Int("count", len(followees)))
	return followees, nil
}
