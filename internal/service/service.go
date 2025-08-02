package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	user_client "pinstack-relation-service/internal/clients/user"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
	"pinstack-relation-service/internal/repository"
	"pinstack-relation-service/internal/uow"
	"pinstack-relation-service/internal/utils"
	"time"

	"github.com/soloda1/pinstack-proto-definitions/events"
)

type Service struct {
	followRepo repository.FollowRepository
	userClient user_client.Client
	uow        uow.UnitOfWork
	log        *logger.Logger
}

func NewFollowService(log *logger.Logger, followRepo repository.FollowRepository, uow uow.UnitOfWork, userClient user_client.Client) *Service {
	return &Service{
		log:        log,
		followRepo: followRepo,
		userClient: userClient,
		uow:        uow,
	}
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID int64) (err error) {
	s.log.Info("Follow request received", slog.Int64("followerID", followerID), slog.Int64("followeeID", followeeID))

	if followerID == followeeID {
		return custom_errors.ErrSelfFollow
	}

	_, err = s.userClient.GetUser(ctx, followeeID)
	if err != nil {
		s.log.Error("Failed to get user", slog.Int64("followeeID", followeeID))
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			s.log.Debug("User not found in follow", slog.Int64("followeeID", followeeID), slog.String("error", err.Error()))
			return custom_errors.ErrUserNotFound
		default:
			s.log.Error("Failed to get user", slog.Int64("followeeID", followeeID))
			return err
		}
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
		return custom_errors.ErrAlreadyFollowing
	}

	follower, err := followRepo.Create(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error creating follow relationship", slog.String("error", err.Error()))
		return err
	}

	payload, err := json.Marshal(events.FollowCreatedPayload{
		FollowerID:  follower.FollowerID,
		FolloweeID:  follower.FolloweeID,
		Timestamptz: time.Now(),
	})
	if err != nil {
		s.log.Error("Failed to marshal payload", slog.String("error", err.Error()))
		return err
	}

	event := model.OutboxEvent{
		EventType:   events.EventTypeFollowCreated,
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
		return custom_errors.ErrSelfUnfollow
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
	_, err := s.userClient.GetUser(ctx, followeeID)
	if err != nil {
		s.log.Error("Failed to get user", slog.Int64("followeeID", followeeID))
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			s.log.Debug("User not found in follow", slog.Int64("followeeID", followeeID), slog.String("error", err.Error()))
			return nil, custom_errors.ErrUserNotFound
		default:
			s.log.Error("Failed to get user", slog.Int64("followeeID", followeeID))
			return nil, err
		}
	}
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
	_, err := s.userClient.GetUser(ctx, followerID)
	if err != nil {
		s.log.Error("Failed to get user", slog.Int64("followerID", followerID))
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			s.log.Debug("User not found in follow", slog.Int64("followerID", followerID), slog.String("error", err.Error()))
			return nil, custom_errors.ErrUserNotFound
		default:
			s.log.Error("Failed to get user", slog.Int64("followerID", followerID))
			return nil, err
		}
	}
	limit, offset := utils.SetPaginationDefaults(limit, page)
	followees, err := s.followRepo.GetFollowees(ctx, followerID, limit, offset)
	if err != nil {
		s.log.Error("Error getting followees", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("Followees retrieved successfully", slog.Int64("followerID", followerID), slog.Int("count", len(followees)))
	return followees, nil
}
