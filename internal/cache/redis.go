package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis 客户端封装
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(addr, password string, db, poolSize int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})

	return &RedisClient{client: rdb}
}

// Ping 测试连接
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Get 获取缓存值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set 设置缓存值
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Delete 删除缓存值
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	return r.client.Exists(ctx, keys...).Result()
}

// Incr 递增计数器
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr 递减计数器
func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// Expire 设置过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Pipeline 获取管道用于批量操作
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient 获取原始 Redis 客户端（用于高级操作）
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// HGet 获取哈希字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HSet 设置哈希字段
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HDel 删除哈希字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}
	return r.client.HDel(ctx, key, fields...).Err()
}

// HGetAll 获取哈希所有字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// TTL 获取剩余过期时间
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// SetNX 设置值（仅当键不存在时）
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// MGet 批量获取
func (r *RedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return []interface{}{}, nil
	}
	return r.client.MGet(ctx, keys...).Result()
}

// MSet 批量设置
func (r *RedisClient) MSet(ctx context.Context, pairs ...interface{}) error {
	if len(pairs) == 0 {
		return nil
	}
	return r.client.MSet(ctx, pairs...).Err()
}

// FlushDB 清空当前数据库（谨慎使用）
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Info 获取 Redis 信息
func (r *RedisClient) Info(ctx context.Context, section ...string) (string, error) {
	return r.client.Info(ctx, section...).Result()
}

// DBSize 获取当前数据库键数量
func (r *RedisClient) DBSize(ctx context.Context) (int64, error) {
	return r.client.DBSize(ctx).Result()
}

// Keys 获取匹配的键（生产环境慎用）
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Scan 扫描键（推荐用于生产环境）
func (r *RedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}

// PipelineExec 执行管道命令
func (r *RedisClient) PipelineExec(ctx context.Context, pipe redis.Pipeliner) error {
	_, err := pipe.Exec(ctx)
	return err
}

// CacheMiss 缓存未命中错误
type CacheMiss struct {
	Key string
}

func (e *CacheMiss) Error() string {
	return fmt.Sprintf("cache miss: %s", e.Key)
}

// IsCacheMiss 检查是否为缓存未命中
func IsCacheMiss(err error) bool {
	return err == redis.Nil
}
