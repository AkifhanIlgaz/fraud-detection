package fraud

import (
	"context"
	"log"
	"math"
	"time"

	"fraud-detection/internal/cache"
	"fraud-detection/internal/queue"
)

type violation struct {
	reason FraudReason
	count  int64 // velocity ihlalinde işaretlenecek işlem sayısı; diğer kurallarda 0
}

type rule func(ctx context.Context, msg queue.TransactionMessage) (violation, bool)

func (a *Analyzer) rules() []rule {
	return []rule{
		a.checkVelocity,
		a.checkAmountAnomaly,
		a.checkImpossibleTravel,
	}
}

func (a *Analyzer) runRules(ctx context.Context, msg queue.TransactionMessage) []violation {
	var violations []violation
	for _, r := range a.rules() {
		if v, ok := r(ctx, msg); ok {
			violations = append(violations, v)
		}
	}
	return violations
}

func (a *Analyzer) checkVelocity(ctx context.Context, msg queue.TransactionMessage) (violation, bool) {
	count, err := a.cache.IncrVelocity(ctx, msg.UserID, velocityPeriod)
	if err != nil {
		log.Printf("[fraud] velocity check hatası: %v — kural atlanıyor", err)
		return violation{}, false
	}

	if count > maxTxPerMinute {
		return violation{reason: ReasonVelocity, count: count}, true
	}

	return violation{}, false
}

func (a *Analyzer) checkAmountAnomaly(ctx context.Context, msg queue.TransactionMessage) (violation, bool) {
	avg, exists, err := a.cache.GetAmountAverage(ctx, msg.UserID)
	if err != nil {
		log.Printf("[fraud] amount anomaly check hatası: %v — kural atlanıyor", err)
		return violation{}, false
	}

	if !exists {
		return violation{}, false
	}

	if msg.Amount > avg*suspiciousAmountFactor {
		return violation{reason: ReasonAmountAnomaly}, true
	}
	return violation{}, false
}

func (a *Analyzer) checkImpossibleTravel(ctx context.Context, msg queue.TransactionMessage) (violation, bool) {
	createdAt, err := time.Parse(time.RFC3339, msg.CreatedAt)
	if err != nil {
		log.Printf("[fraud] impossible travel check: createdAt parse hatası: %v", err)
		return violation{}, false
	}
	current := cache.LastLocation{Lat: msg.Lat, Lon: msg.Lon, CreatedAt: createdAt.Unix()}

	prev, err := a.cache.GetLastLocation(ctx, msg.UserID)
	if err != nil {
		log.Printf("[fraud] impossible travel check hatası: %v — kural atlanıyor", err)
		return violation{}, false
	}
	if prev == nil {
		return violation{}, false
	}

	distanceKm := haversineKm(prev.Lat, prev.Lon, current.Lat, current.Lon)
	elapsedHours := time.Unix(current.CreatedAt, 0).Sub(time.Unix(prev.CreatedAt, 0)).Hours()
	if elapsedHours <= 0 {
		return violation{}, false
	}

	if elapsedHours < distanceKm/maxTravelSpeedKmH {
		return violation{reason: ReasonImpossibleTravel}, true
	}

	// eger fraud degilse last location guncelle
	a.cache.SetLastLocation(ctx, msg.UserID, current, locationCacheTTL)
	return violation{}, false
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
