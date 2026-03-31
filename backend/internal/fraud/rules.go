package fraud

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"fraud-detection/internal/queue"
)

type rule func(ctx context.Context, msg queue.TransactionMessage) (FraudReason, bool)

func (a *Analyzer) rules() []rule {
	return []rule{
		a.checkVelocity,
		a.checkAmountAnomaly,
		a.checkImpossibleTravel,
	}
}

func (a *Analyzer) runRules(ctx context.Context, msg queue.TransactionMessage) []FraudReason {
	var reasons []FraudReason
	for _, r := range a.rules() {
		if reason, ok := r(ctx, msg); ok {
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

func (a *Analyzer) checkVelocity(ctx context.Context, msg queue.TransactionMessage) (FraudReason, bool) {
	count, err := a.cache.IncrVelocity(ctx, msg.UserID, velocityPeriod)
	if err != nil {
		log.Printf("[fraud] velocity check hatası: %v — kural atlanıyor", err)
		return "", false
	}
	if count > maxTxPerMinute {
		return ReasonVelocity, true
	}
	return "", false
}

func (a *Analyzer) checkAmountAnomaly(ctx context.Context, msg queue.TransactionMessage) (FraudReason, bool) {
	members, err := a.cache.ApprovedAmountsInWindow(ctx, msg.UserID, amountHistoryTTL)
	if err != nil {
		log.Printf("[fraud] amount anomaly check hatası: %v — kural atlanıyor", err)
		return "", false
	}
	if len(members) == 0 {
		return "", false
	}

	var total float64
	var count int
	for _, m := range members {
		parts := strings.SplitN(m, "|", 2)
		if len(parts) != 2 {
			continue
		}
		amount, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		total += amount
		count++
	}
	if count == 0 {
		return "", false
	}

	if msg.Amount > (total/float64(count))*suspiciousAmountFactor {
		return ReasonAmountAnomaly, true
	}
	return "", false
}

type lastLocation struct {
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *Analyzer) checkImpossibleTravel(ctx context.Context, msg queue.TransactionMessage) (FraudReason, bool) {
	createdAt, err := time.Parse(time.RFC3339, msg.CreatedAt)
	if err != nil {
		log.Printf("[fraud] impossible travel check: createdAt parse hatası: %v", err)
		return "", false
	}
	current := lastLocation{Lat: msg.Lat, Lon: msg.Lon, CreatedAt: createdAt}
	defer func() {
		data, err := json.Marshal(current)
		if err == nil {
			a.cache.SetLastLocation(ctx, msg.UserID, string(data), locationCacheTTL)
		}
	}()

	val, err := a.cache.GetLastLocation(ctx, msg.UserID)
	if err != nil {
		log.Printf("[fraud] impossible travel check hatası: %v — kural atlanıyor", err)
		return "", false
	}
	if val == "" {
		return "", false
	}

	var prev lastLocation
	if err := json.Unmarshal([]byte(val), &prev); err != nil {
		return "", false
	}

	distanceKm := haversineKm(prev.Lat, prev.Lon, current.Lat, current.Lon)
	elapsedHours := current.CreatedAt.Sub(prev.CreatedAt).Hours()
	if elapsedHours <= 0 {
		return "", false
	}

	if elapsedHours < distanceKm/maxTravelSpeedKmH {
		return ReasonImpossibleTravel, true
	}
	return "", false
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
