package service

import (
	"context"
	"log/slog"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/repository"
)

type Service struct {
	followRepo repository.FollowRepository
	log        *logger.Logger
}

func NewFollowService(log *logger.Logger, followRepo repository.FollowRepository) *Service {
	return &Service{
		log:        log,
		followRepo: followRepo,
	}
}

func (s *Service) Follow(ctx context.Context, followerID, followeeID int64) error {
	s.log.Info("Follow request received", slog.Int64("followerID", followerID), slog.Int64("followeeID", followeeID))

	if followerID == followeeID {
		return custom_errors.ErrSelfFollow
	}

	exists, err := s.followRepo.Exists(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error checking follow existence", slog.String("error", err.Error()))
		return err
	}
	if exists {
		return custom_errors.ErrFollowRelationExists
	}

	err = s.followRepo.Create(ctx, followerID, followeeID)
	if err != nil {
		s.log.Error("Error creating follow relationship", slog.String("error", err.Error()))
		return err
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

func (s *Service) GetFollowers(ctx context.Context, followeeID int64) ([]int64, error) {
	s.log.Info("GetFollowers request received", slog.Int64("followeeID", followeeID))

	followers, err := s.followRepo.GetFollowers(ctx, followeeID)
	if err != nil {
		s.log.Error("Error getting followers", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("Followers retrieved successfully", slog.Int64("followeeID", followeeID), slog.Int("count", len(followers)))
	return followers, nil
}

func (s *Service) GetFollowees(ctx context.Context, followerID int64) ([]int64, error) {
	s.log.Info("GetFollowees request received", slog.Int64("followerID", followerID))

	followees, err := s.followRepo.GetFollowees(ctx, followerID)
	if err != nil {
		s.log.Error("Error getting followees", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("Followees retrieved successfully", slog.Int64("followerID", followerID), slog.Int("count", len(followees)))
	return followees, nil
}
