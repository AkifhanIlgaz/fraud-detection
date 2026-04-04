package fraud

import "time"

const (
	// son 1 dakikada max 5 tx
	maxTxPerMinute = 5
	velocityPeriod = time.Minute

	// son 24 saat ortalamasi * 3' ten buyukse supheli
	suspiciousAmountFactor = 3.0
	amountHistoryTTL       = 24 * time.Hour

	// ucakla gitse bile imkansiz
	maxTravelSpeedKmH = 900.0
	locationCacheTTL  = 24 * time.Hour
)
