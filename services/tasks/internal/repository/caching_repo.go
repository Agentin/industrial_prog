package repository

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/student/tech-ip-sem2/services/tasks/internal/service"
	"github.com/student/tech-ip-sem2/shared/cache"
	"go.uber.org/zap"
)

type CachingTaskRepository struct {
	repo      TaskRepository
	redis     *cache.RedisClient
	logger    *zap.Logger
	ttlBase   int
	ttlJitter int
}

func NewCachingTaskRepository(repo TaskRepository, redis *cache.RedisClient, logger *zap.Logger) *CachingTaskRepository {
	return &CachingTaskRepository{
		repo:      repo,
		redis:     redis,
		logger:    logger,
		ttlBase:   120,
		ttlJitter: 30,
	}
}

func (r *CachingTaskRepository) GetByID(ctx context.Context, id string) (service.Task, error) {
	if r.redis == nil {
		r.logger.Debug("redis client unavailable, using db directly")
		return r.repo.GetByID(ctx, id)
	}
	key := fmt.Sprintf("tasks:task:%s", id)
	var task service.Task
	hit, err := r.redis.Get(ctx, key, &task)
	if err != nil {
		r.logger.Warn("redis get failed, fallback to db", zap.Error(err))
	}
	if hit {
		r.logger.Info("cache HIT", zap.String("key", key))
		return task, nil
	}
	r.logger.Info("cache MISS", zap.String("key", key))
	task, err = r.repo.GetByID(ctx, id)
	if err != nil {
		return task, err
	}
	ttl := time.Duration(r.ttlBase+rand.Intn(r.ttlJitter)) * time.Second
	if err := r.redis.Set(ctx, key, task, ttl); err != nil {
		r.logger.Warn("redis set failed", zap.Error(err))
	} else {
		r.logger.Debug("cached task", zap.String("key", key), zap.Duration("ttl", ttl))
	}
	return task, nil
}

func (r *CachingTaskRepository) Create(ctx context.Context, task service.Task) error {
	return r.repo.Create(ctx, task)
}

func (r *CachingTaskRepository) GetAll(ctx context.Context) ([]service.Task, error) {
	return r.repo.GetAll(ctx)
}

func (r *CachingTaskRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	err := r.repo.Update(ctx, id, updates)
	if err == nil && r.redis != nil {
		key := fmt.Sprintf("tasks:task:%s", id)
		if delErr := r.redis.Del(ctx, key); delErr != nil {
			r.logger.Warn("failed to invalidate cache", zap.Error(delErr))
		}
	}
	return err
}

func (r *CachingTaskRepository) Delete(ctx context.Context, id string) error {
	err := r.repo.Delete(ctx, id)
	if err == nil && r.redis != nil {
		key := fmt.Sprintf("tasks:task:%s", id)
		_ = r.redis.Del(ctx, key)
	}
	return err
}

func (r *CachingTaskRepository) SearchByTitle(ctx context.Context, title string) ([]service.Task, error) {
	return r.repo.SearchByTitle(ctx, title)
}
