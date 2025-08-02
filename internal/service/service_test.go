package service

import (
	"context"
	"errors"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
	"pinstack-relation-service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Service, *mocks.FollowRepository, *mocks.UnitOfWork, *mocks.Transaction, *mocks.OutboxRepository, *mocks.Client) {
	mockFollowRepo := mocks.NewFollowRepository(t)
	mockUOW := mocks.NewUnitOfWork(t)
	mockTx := mocks.NewTransaction(t)
	mockOutboxRepo := mocks.NewOutboxRepository(t)
	mockUserClient := mocks.NewClient(t)

	t.Cleanup(func() {
	})

	log := logger.New("test")

	svc := NewFollowService(log, mockFollowRepo, mockUOW, mockUserClient)

	return svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo, mockUserClient
}

func TestService_Follow(t *testing.T) {
	t.Run("успешное создание подписки", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo)
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)

		follower := model.Follower{
			FollowerID: followerID,
			FolloweeID: followeeID,
		}
		mockFollowRepo.On("Create", ctx, followerID, followeeID).Return(follower, nil)

		mockOutboxRepo.On("AddEvent", ctx, mock.AnythingOfType("model.OutboxEvent")).Return(nil)
		mockTx.On("Commit", ctx).Return(nil)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.NoError(t, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockOutboxRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при попытке подписаться на себя", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)

		err := svc.Follow(ctx, followerID, followerID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrSelfFollow, err)

		mockFollowRepo.AssertNotCalled(t, "Exists")
		mockUOW.AssertNotCalled(t, "Begin")
		mockUserClient.AssertNotCalled(t, "GetUser")
	})

	t.Run("ошибка при несуществующем пользователе", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(nil, custom_errors.ErrUserNotFound)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrUserNotFound, err)

		mockFollowRepo.AssertNotCalled(t, "Exists")
		mockUOW.AssertNotCalled(t, "Begin")
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка сети при проверке пользователя в Follow", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		networkErr := errors.New("network timeout")
		mockUserClient.On("GetUser", ctx, followeeID).Return(nil, networkErr)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		assert.Equal(t, networkErr, err)

		mockFollowRepo.AssertNotCalled(t, "Exists")
		mockUOW.AssertNotCalled(t, "Begin")
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при старте транзакции", func(t *testing.T) {
		svc, _, mockUOW, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockUOW.On("Begin", ctx).Return(nil, errors.New("db connection error"))

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrDatabaseQuery, err)
		mockUOW.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при попытке создать уже существующую подписку", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo)
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(true, nil)
		mockTx.On("Rollback", ctx).Return(nil)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrAlreadyFollowing, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
		mockFollowRepo.AssertNotCalled(t, "Create")
	})

	t.Run("ошибка при создании подписки", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo)
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)
		mockFollowRepo.On("Create", ctx, followerID, followeeID).Return(model.Follower{}, errors.New("db error"))
		mockTx.On("Rollback", ctx).Return(nil)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при добавлении события в outbox", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo)
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)

		follower := model.Follower{
			FollowerID: followerID,
			FolloweeID: followeeID,
		}
		mockFollowRepo.On("Create", ctx, followerID, followeeID).Return(follower, nil)

		mockOutboxRepo.On("AddEvent", ctx, mock.AnythingOfType("model.OutboxEvent")).Return(errors.New("outbox error"))
		mockTx.On("Rollback", ctx).Return(nil)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockOutboxRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при коммите транзакции", func(t *testing.T) {
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo)
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)

		follower := model.Follower{
			FollowerID: followerID,
			FolloweeID: followeeID,
		}
		mockFollowRepo.On("Create", ctx, followerID, followeeID).Return(follower, nil)

		mockOutboxRepo.On("AddEvent", ctx, mock.AnythingOfType("model.OutboxEvent")).Return(nil)
		mockTx.On("Commit", ctx).Return(errors.New("commit error"))
		mockTx.On("Rollback", ctx).Return(nil)

		err := svc.Follow(ctx, followerID, followeeID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrDatabaseQuery, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockOutboxRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})
}

func TestService_Unfollow(t *testing.T) {
	t.Run("успешное удаление подписки", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(true, nil)
		mockFollowRepo.On("Delete", ctx, followerID, followeeID).Return(nil)

		err := svc.Unfollow(ctx, followerID, followeeID)

		assert.NoError(t, err)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при попытке отписаться от себя", func(t *testing.T) {
		svc, _, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)

		err := svc.Unfollow(ctx, followerID, followerID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrSelfUnfollow, err)
	})

	t.Run("подписка не существует", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)

		err := svc.Unfollow(ctx, followerID, followeeID)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrFollowRelationNotFound, err)
	})

	t.Run("ошибка при проверке существования подписки", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, errors.New("db error"))

		err := svc.Unfollow(ctx, followerID, followeeID)

		assert.Error(t, err)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при удалении подписки", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(true, nil)
		mockFollowRepo.On("Delete", ctx, followerID, followeeID).Return(errors.New("db error"))

		err := svc.Unfollow(ctx, followerID, followeeID)

		assert.Error(t, err)
		mockFollowRepo.AssertExpectations(t)
	})
}

func TestService_GetFollowers(t *testing.T) {
	t.Run("успешное получение подписчиков", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followeeID := int64(2)
		limit, page := int32(10), int32(1)
		expectedFollowers := []int64{1, 3, 5}

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockFollowRepo.On("GetFollowers", ctx, followeeID, limit, int32(0)).Return(expectedFollowers, nil)

		followers, err := svc.GetFollowers(ctx, followeeID, limit, page)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowers, followers)
		mockFollowRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при несуществующем пользователе в GetFollowers", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followeeID := int64(2)
		limit, page := int32(10), int32(1)

		mockUserClient.On("GetUser", ctx, followeeID).Return(nil, custom_errors.ErrUserNotFound)

		followers, err := svc.GetFollowers(ctx, followeeID, limit, page)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrUserNotFound, err)
		assert.Nil(t, followers)
		mockFollowRepo.AssertNotCalled(t, "GetFollowers")
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка сети при проверке пользователя в GetFollowers", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followeeID := int64(2)
		limit, page := int32(10), int32(1)

		networkErr := errors.New("network timeout")
		mockUserClient.On("GetUser", ctx, followeeID).Return(nil, networkErr)

		followers, err := svc.GetFollowers(ctx, followeeID, limit, page)

		assert.Error(t, err)
		assert.Equal(t, networkErr, err)
		assert.Nil(t, followers)
		mockFollowRepo.AssertNotCalled(t, "GetFollowers")
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при получении подписчиков", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followeeID := int64(2)
		limit, page := int32(10), int32(1)

		mockUserClient.On("GetUser", ctx, followeeID).Return(&model.User{ID: followeeID}, nil)

		mockFollowRepo.On("GetFollowers", ctx, followeeID, limit, int32(0)).Return(nil, errors.New("db error"))

		followers, err := svc.GetFollowers(ctx, followeeID, limit, page)

		assert.Error(t, err)
		assert.Nil(t, followers)
		mockFollowRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})
}

func TestService_GetFollowees(t *testing.T) {
	t.Run("успешное получение подписок", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)
		limit, page := int32(10), int32(1)
		expectedFollowees := []int64{2, 4, 6}

		mockUserClient.On("GetUser", ctx, followerID).Return(&model.User{ID: followerID}, nil)

		mockFollowRepo.On("GetFollowees", ctx, followerID, limit, int32(0)).Return(expectedFollowees, nil)

		followees, err := svc.GetFollowees(ctx, followerID, limit, page)

		require.NoError(t, err)
		assert.Equal(t, expectedFollowees, followees)
		mockFollowRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при несуществующем пользователе в GetFollowees", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)
		limit, page := int32(10), int32(1)

		mockUserClient.On("GetUser", ctx, followerID).Return(nil, custom_errors.ErrUserNotFound)

		followees, err := svc.GetFollowees(ctx, followerID, limit, page)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrUserNotFound, err)
		assert.Nil(t, followees)
		mockFollowRepo.AssertNotCalled(t, "GetFollowees")
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка сети при проверке пользователя в GetFollowees", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)
		limit, page := int32(10), int32(1)

		networkErr := errors.New("network timeout")
		mockUserClient.On("GetUser", ctx, followerID).Return(nil, networkErr)

		followees, err := svc.GetFollowees(ctx, followerID, limit, page)

		assert.Error(t, err)
		assert.Equal(t, networkErr, err)
		assert.Nil(t, followees)
		mockFollowRepo.AssertNotCalled(t, "GetFollowees")
		mockUserClient.AssertExpectations(t)
	})

	t.Run("ошибка при получении подписок", func(t *testing.T) {
		svc, mockFollowRepo, _, _, _, mockUserClient := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)
		limit, page := int32(10), int32(1)

		mockUserClient.On("GetUser", ctx, followerID).Return(&model.User{ID: followerID}, nil)

		mockFollowRepo.On("GetFollowees", ctx, followerID, limit, int32(0)).Return(nil, errors.New("db error"))

		followees, err := svc.GetFollowees(ctx, followerID, limit, page)

		assert.Error(t, err)
		assert.Nil(t, followees)
		mockFollowRepo.AssertExpectations(t)
		mockUserClient.AssertExpectations(t)
	})
}
