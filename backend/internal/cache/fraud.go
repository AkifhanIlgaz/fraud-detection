package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type FraudCache struct {
	rdb *redis.Client
}

func NewFraudCache(rdb *redis.Client) *FraudCache {
	return &FraudCache{rdb: rdb}
}

// eger son 1 dakikada islem yapmamissa 0 dan baslar ve 1 artirilir
// son 1 dakikada yapilan ilk islem oldugu icin ttl baslatilir
// her islem yapildiginda ttl yenilenir. ornegin fraud kullanici 1 dakika icinde 5 islem yapar ve fraud olarak gozukmez. her dakika 5 islem yapip bu checki bypass edebilir
func (c *FraudCache) IncrVelocity(ctx context.Context, userID string, ttl time.Duration) (int64, error) {
	key := fraudKey("velocity", userID)

	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		return count, fmt.Errorf("velocity ttl set: %w", err)
	}

	return count, nil
}

// GetAmountAverage, kullanıcının onaylı işlemlerinden hesaplanan ortalama tutarı döndürür.
// Geçmiş yoksa (exists=false) döner.
func (c *FraudCache) GetAmountAverage(ctx context.Context, userID string) (avg float64, exists bool, err error) {
	key := fraudKey("avg_amount", userID)
	vals, err := c.rdb.HMGet(ctx, key, "sum", "count").Result()
	if err != nil {
		return 0, false, err
	}

	if vals[0] == nil || vals[1] == nil {
		return 0, false, nil
	}

	sum, err := strconv.ParseFloat(vals[0].(string), 64)
	if err != nil {
		return 0, false, nil
	}

	count, err := strconv.ParseInt(vals[1].(string), 10, 64)
	if err != nil || count == 0 {
		return 0, false, nil
	}

	return sum / float64(count), true, nil
}

// UpdateAmountAverage, fraud olmayan işlemin tutarını sum ve count'a ekler.
// İlk işlemde window süresiyle TTL başlatılır; pencere sabit kalır (kayar değil).
func (c *FraudCache) UpdateAmountAverage(ctx context.Context, userID string, amount float64, window time.Duration) {
	key := fraudKey("avg_amount", userID)

	count, _ := c.rdb.HIncrBy(ctx, key, "count", 1).Result()

	c.rdb.HIncrByFloat(ctx, key, "sum", amount)

	if count == 1 {
		c.rdb.Expire(ctx, key, window)
	}
}

type LastLocation struct {
	Lat       float64 `redis:"lat"`
	Lon       float64 `redis:"lon"`
	CreatedAt int64   `redis:"created_at"` // Unix seconds
}

// GetLastLocation, kullanıcının cache'deki son konumunu döndürür.
// Konum yoksa (nil, nil) döner.
func (c *FraudCache) GetLastLocation(ctx context.Context, userID string) (*LastLocation, error) {
	key := fraudKey("location", userID)

	var loc LastLocation
	if err := c.rdb.HGetAll(ctx, key).Scan(&loc); err != nil {
		return nil, err
	}

	if loc.CreatedAt == 0 {
		return nil, nil
	}

	return &loc, nil
}

func (c *FraudCache) SetLastLocation(ctx context.Context, userID string, loc LastLocation, ttl time.Duration) {
	key := fraudKey("location", userID)

	c.rdb.HSet(ctx, key, loc)
	c.rdb.Expire(ctx, key, ttl)
}

func fraudKey(parts ...string) string {
	return "fraud:" + strings.Join(parts, ":")
}
