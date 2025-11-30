package service

import (
	"context"
	"encoding/json"
	"time"

	"gin-mongo-aws/internal/database"
	"gin-mongo-aws/internal/models"
	"gin-mongo-aws/internal/repository"
	"gin-mongo-aws/internal/logger"

	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	return s.repo.Create(ctx, user)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.FindAll(ctx)
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	// Try to get from cache
	val, err := database.RedisClient.Get(ctx, "user:"+id).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			logger.Log.Info("Cache hit for user", zap.String("id", id))
			return &user, nil
		}
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Set to cache
	data, _ := json.Marshal(user)
	database.RedisClient.Set(ctx, "user:"+id, data, 10*time.Minute)

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, user *models.User) error {
	err := s.repo.Update(ctx, id, user)
	if err == nil {
		// Invalidate cache
		database.RedisClient.Del(ctx, "user:"+id)
	}
	return err
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err == nil {
		// Invalidate cache
		database.RedisClient.Del(ctx, "user:"+id)
	}
	return err
}
