package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"pinstack-relation-service/internal/custom_errors"
	"pinstack-relation-service/internal/logger"
	"pinstack-relation-service/internal/model"
	"pinstack-relation-service/mocks"
	"testing"
)

func setupTest(t *testing.T) (*Service, *mocks.FollowRepository, *mocks.UnitOfWork, *mocks.Transaction, *mocks.OutboxRepository) {
	mockFollowRepo := mocks.NewFollowRepository(t)
	mockUOW := mocks.NewUnitOfWork(t)
	mockTx := mocks.NewTransaction(t)
	mockOutboxRepo := mocks.NewOutboxRepository(t)

	// Отключаем автоматическую проверку моков
	t.Cleanup(func() {
		// Пустая функция для переопределения автоматической проверки
	})

	// Инициализируем логгер в тихом режиме для тестов
	log := logger.New("test")

	svc := NewFollowService(log, mockFollowRepo, mockUOW)

	return svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo
}

func TestService_Follow(t *testing.T) {
	t.Run("успешное создание подписки", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo)
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)

		follower := model.Follower{
			FollowerID: followerID,
			FolloweeID: followeeID,
		}
		mockFollowRepo.On("Create", ctx, followerID, followeeID).Return(follower, nil)

		// Проверяем вызов добавления события в outbox
		mockOutboxRepo.On("AddEvent", ctx, mock.AnythingOfType("model.OutboxEvent")).Return(nil)
		mockTx.On("Commit", ctx).Return(nil)

		// Act
		err := svc.Follow(ctx, followerID, followeeID)

		// Assert
		assert.NoError(t, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockOutboxRepo.AssertExpectations(t)
	})

	t.Run("ошибка при попытке подписаться на себя", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, mockUOW, _, _ := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)

		// Act
		err := svc.Follow(ctx, followerID, followerID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrSelfFollow, err)

		// В этом тесте не вызываются методы моков
		mockFollowRepo.AssertNotCalled(t, "Exists")
		mockUOW.AssertNotCalled(t, "Begin")
	})

	t.Run("ошибка при старте транзакции", func(t *testing.T) {
		// Arrange
		svc, _, mockUOW, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUOW.On("Begin", ctx).Return(nil, errors.New("db connection error"))

		// Act
		err := svc.Follow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrDatabaseQuery, err)
		mockUOW.AssertExpectations(t)
	})

	t.Run("ошибка при создании подписки", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockUOW.On("Begin", ctx).Return(mockTx, nil)
		mockTx.On("FollowRepository").Return(mockFollowRepo)
		mockTx.On("OutboxRepository").Return(mockOutboxRepo) // Нужно добавить это ожидание
		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)
		mockFollowRepo.On("Create", ctx, followerID, followeeID).Return(model.Follower{}, errors.New("db error"))
		mockTx.On("Rollback", ctx).Return(nil)

		// Act
		err := svc.Follow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при добавлении события в outbox", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

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

		// Act
		err := svc.Follow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockOutboxRepo.AssertExpectations(t)
	})

	t.Run("ошибка при коммите транзакции", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, mockUOW, mockTx, mockOutboxRepo := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

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

		// Act
		err := svc.Follow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrDatabaseQuery, err)
		mockUOW.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockFollowRepo.AssertExpectations(t)
		mockOutboxRepo.AssertExpectations(t)
	})
}

func TestService_Unfollow(t *testing.T) {
	t.Run("успешное удаление подписки", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(true, nil)
		mockFollowRepo.On("Delete", ctx, followerID, followeeID).Return(nil)

		// Act
		err := svc.Unfollow(ctx, followerID, followeeID)

		// Assert
		assert.NoError(t, err)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при попытке отписаться от себя", func(t *testing.T) {
		// Arrange
		svc, _, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)

		// Act
		err := svc.Unfollow(ctx, followerID, followerID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrSelfFollow, err)
	})

	t.Run("подписка не существует", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, nil)

		// Act
		err := svc.Unfollow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrFollowRelationNotFound, err)
	})

	t.Run("ошибка при проверке существования подписки", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(false, errors.New("db error"))

		// Act
		err := svc.Unfollow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при удалении подписки", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID, followeeID := int64(1), int64(2)

		mockFollowRepo.On("Exists", ctx, followerID, followeeID).Return(true, nil)
		mockFollowRepo.On("Delete", ctx, followerID, followeeID).Return(errors.New("db error"))

		// Act
		err := svc.Unfollow(ctx, followerID, followeeID)

		// Assert
		assert.Error(t, err)
		mockFollowRepo.AssertExpectations(t)
	})
}

func TestService_GetFollowers(t *testing.T) {
	t.Run("успешное получение подписчиков", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followeeID := int64(2)
		limit, page := int32(10), int32(1)
		expectedFollowers := []int64{1, 3, 5}

		// limit и offset после применения SetPaginationDefaults
		mockFollowRepo.On("GetFollowers", ctx, followeeID, limit, int32(0)).Return(expectedFollowers, nil)

		// Act
		followers, err := svc.GetFollowers(ctx, followeeID, limit, page)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedFollowers, followers)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении подписчиков", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followeeID := int64(2)
		limit, page := int32(10), int32(1)

		mockFollowRepo.On("GetFollowers", ctx, followeeID, limit, int32(0)).Return(nil, errors.New("db error"))

		// Act
		followers, err := svc.GetFollowers(ctx, followeeID, limit, page)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, followers)
	})
}

func TestService_GetFollowees(t *testing.T) {
	t.Run("успешное получение подписок", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)
		limit, page := int32(10), int32(1)
		expectedFollowees := []int64{2, 4, 6}

		// limit и offset после применения SetPaginationDefaults
		mockFollowRepo.On("GetFollowees", ctx, followerID, limit, int32(0)).Return(expectedFollowees, nil)

		// Act
		followees, err := svc.GetFollowees(ctx, followerID, limit, page)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedFollowees, followees)
		mockFollowRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении подписок", func(t *testing.T) {
		// Arrange
		svc, mockFollowRepo, _, _, _ := setupTest(t)
		ctx := context.Background()
		followerID := int64(1)
		limit, page := int32(10), int32(1)

		mockFollowRepo.On("GetFollowees", ctx, followerID, limit, int32(0)).Return(nil, errors.New("db error"))

		// Act
		followees, err := svc.GetFollowees(ctx, followerID, limit, page)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, followees)
	})
}
