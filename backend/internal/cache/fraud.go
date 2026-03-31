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
func (c *FraudCache) IncrVelocity(ctx context.Context, userID string, ttl time.Duration) (int64, error) {
	key := fraudKey("velocity", userID)
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 {
		c.rdb.Expire(ctx, key, ttl)
	}
	return count, nil
}

// ApprovedAmountsInWindow, kullanıcının son window süresi içindeki onaylı
// işlem tutarlarını döndürür. Member formatı: "{txID}|{amount}".
func (c *FraudCache) ApprovedAmountsInWindow(ctx context.Context, userID string, window time.Duration) ([]string, error) {
	key := fraudKey("amounts", userID)
	cutoff := time.Now().Add(-window)
	return c.rdb.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     key,
		Start:   strconv.FormatInt(cutoff.UnixMilli(), 10),
		Stop:    "+inf",
		ByScore: true,
	}).Result()
}

// RecordApprovedAmount, fraud olmayan işlemin tutarını geçmişe ekler
// ve window dışına çıkan eski kayıtları temizler.
func (c *FraudCache) RecordApprovedAmount(ctx context.Context, userID, txID string, amount float64, window time.Duration) {
	key := fraudKey("amounts", userID)
	now := time.Now()
	cutoff := now.Add(-window)

	pipe := c.rdb.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%s|%.6f", txID, amount),
	})
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(cutoff.UnixMilli(), 10))
	pipe.Expire(ctx, key, window+time.Hour)
	pipe.Exec(ctx) //nolint:errcheck
}

// GetLastLocation, kullanıcının cache'deki son konumunu döndürür.
// Konum yoksa ("", nil) döner.
func (c *FraudCache) GetLastLocation(ctx context.Context, userID string) (string, error) {
	key := fraudKey("location", userID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *FraudCache) SetLastLocation(ctx context.Context, userID string, data string, ttl time.Duration) {
	c.rdb.Set(ctx, fraudKey("location", userID), data, ttl)
}

func fraudKey(parts ...string) string {
	return "fraud:" + strings.Join(parts, ":")
}
