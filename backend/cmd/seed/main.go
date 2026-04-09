package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func apiEndpoint() string {
	if u := os.Getenv("API_URL"); u != "" {
		return u
	}
	return "http://localhost:8080/api/v1/transactions"
}

type createTxRequest struct {
	UserID    string     `json:"user_id"`
	Amount    float64    `json:"amount"`
	Lat       float64    `json:"lat"`
	Lon       float64    `json:"lon"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// ── Locations ──────────────────────────────────────────────────────────────

type loc struct{ lat, lon float64 }

var (
	istanbul   = loc{41.0082, 28.9784}
	newYork    = loc{40.7128, -74.0060}
	tokyo      = loc{35.6762, 139.6503}
	sydney     = loc{-33.8688, 151.2093}
	losAngeles = loc{34.0522, -118.2437}
	singapore  = loc{1.3521, 103.8198}
	saoPaulo   = loc{-23.5505, -46.6333}
)

// ── Helpers ────────────────────────────────────────────────────────────────

// at returns a pointer to a UTC time N days ago at the given hour:minute.
func at(daysAgo, hour, min int) *time.Time {
	d := time.Now().UTC().Truncate(24*time.Hour).
		AddDate(0, 0, -daysAgo).
		Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
	return &d
}

func build(userID string, amount float64, l loc, createdAt *time.Time) createTxRequest {
	return createTxRequest{
		UserID:    userID,
		Amount:    amount,
		Lat:       l.lat,
		Lon:       l.lon,
		CreatedAt: createdAt,
	}
}

// ── Baseline transactions ──────────────────────────────────────────────────
//
// Sent first to populate Redis amount-average cache for each user.
// All use Istanbul (sets initial location cache too).
// Amounts ~$80–220 → average ≈ $130 per user → fraud threshold ≈ $390.

var baseline = []createTxRequest{
	build("user-001", 110.00, istanbul, nil),
	build("user-001", 135.50, istanbul, nil),
	build("user-001", 98.75, istanbul, nil),
	build("user-002", 145.00, istanbul, nil),
	build("user-002", 120.00, istanbul, nil),
	build("user-002", 160.25, istanbul, nil),
	build("user-003", 80.00, istanbul, nil),
	build("user-003", 95.00, istanbul, nil),
	build("user-003", 110.50, istanbul, nil),
	build("user-004", 200.00, istanbul, nil),
	build("user-004", 185.00, istanbul, nil),
	build("user-004", 215.50, istanbul, nil),
	build("user-005", 90.00, istanbul, nil),
	build("user-005", 115.00, istanbul, nil),
	build("user-005", 88.50, istanbul, nil),
	build("user-006", 160.00, istanbul, nil),
	build("user-006", 140.00, istanbul, nil),
	build("user-006", 175.00, istanbul, nil),
}

// ── Fraud pairs ────────────────────────────────────────────────────────────
//
// Each pair = (setup tx) + (fraud tx).
//   setup tx : user at Istanbul, normal amount, created_at = DayN H:00
//              → approved; updates location cache to {Istanbul, DayN H:00}
//   fraud tx : same user, distant city, high amount (>3× avg ≈ $390),
//              created_at = DayN H:05  (only 5 min later → impossible travel)
//              → triggers AmountAnomaly + ImpossibleTravel → status=fraud
//
// Distribution chosen to produce an ascending trend for the last 30 days.

type fraudPair struct {
	user   string
	day    int // days ago
	hour   int // setup hour (fraud is hour:05)
	from   loc
	to     loc
	amount float64 // fraud tx amount — must be > ~$390
}

var pairs = []fraudPair{
	// Day -29 · 2 frauds
	{"user-001", 29, 9, istanbul, newYork, 950.00},
	{"user-002", 29, 14, istanbul, tokyo, 1200.00},

	// Day -26 · 1 fraud
	{"user-003", 26, 10, istanbul, sydney, 800.00},

	// Day -23 · 2 frauds
	{"user-001", 23, 11, istanbul, losAngeles, 1100.00},
	{"user-004", 23, 15, istanbul, singapore, 1850.00},

	// Day -20 · 1 fraud
	{"user-002", 20, 9, istanbul, newYork, 900.00},

	// Day -18 · 2 frauds
	{"user-005", 18, 10, istanbul, sydney, 850.00},
	{"user-003", 18, 13, istanbul, tokyo, 650.00},

	// Day -15 · 1 fraud
	{"user-006", 15, 11, istanbul, saoPaulo, 1500.00},

	// Day -13 · 2 frauds
	{"user-001", 13, 8, istanbul, singapore, 750.00},
	{"user-004", 13, 16, istanbul, losAngeles, 2000.00},

	// Day -10 · 2 frauds
	{"user-002", 10, 9, istanbul, sydney, 1100.00},
	{"user-005", 10, 14, istanbul, newYork, 880.00},

	// Day -8 · 1 fraud
	{"user-006", 8, 12, istanbul, tokyo, 1300.00},

	// Day -6 · 3 frauds
	{"user-003", 6, 8, istanbul, newYork, 720.00},
	{"user-001", 6, 11, istanbul, sydney, 950.00},
	{"user-004", 6, 15, istanbul, tokyo, 1650.00},

	// Day -4 · 2 frauds
	{"user-002", 4, 10, istanbul, losAngeles, 1250.00},
	{"user-005", 4, 13, istanbul, singapore, 800.00},

	// Day -3 · 3 frauds
	{"user-006", 3, 9, istanbul, newYork, 1100.00},
	{"user-001", 3, 12, istanbul, tokyo, 900.00},
	{"user-003", 3, 15, istanbul, saoPaulo, 780.00},

	// Day -2 · 4 frauds
	{"user-004", 2, 8, istanbul, sydney, 1900.00},
	{"user-002", 2, 11, istanbul, singapore, 1050.00},
	{"user-005", 2, 14, istanbul, newYork, 920.00},
	{"user-006", 2, 16, istanbul, losAngeles, 1200.00},

	// Day -1 · 5 frauds
	{"user-001", 1, 9, istanbul, sydney, 850.00},
	{"user-003", 1, 11, istanbul, newYork, 690.00},
	{"user-004", 1, 13, istanbul, singapore, 2100.00},
	{"user-005", 1, 15, istanbul, tokyo, 760.00},
	{"user-006", 1, 17, istanbul, saoPaulo, 1350.00},
}

// ── HTTP ───────────────────────────────────────────────────────────────────

func post(r createTxRequest) error {
	body, err := json.Marshal(r)
	if err != nil {
		return err
	}
	resp, err := http.Post(apiEndpoint(), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

// ── Main ───────────────────────────────────────────────────────────────────

func main() {
	// ── Phase 1: baseline ─────────────────────────────────────────────────
	fmt.Printf("Phase 1: %d baseline transactions (establishing amount averages)...\n\n", len(baseline))

	for i, r := range baseline {
		if err := post(r); err != nil {
			log.Printf("  [B%02d] FAIL %-10s — %v", i+1, r.UserID, err)
		} else {
			fmt.Printf("  [B%02d] OK   %-10s $%7.2f  istanbul\n", i+1, r.UserID, r.Amount)
		}
	}

	fmt.Println("\nWaiting 4s for queue to drain baseline txs...")
	time.Sleep(4 * time.Second)

	// ── Phase 2: fraud pairs ───────────────────────────────────────────────
	fmt.Printf("\nPhase 2: %d fraud pairs (impossible travel + amount anomaly)...\n\n", len(pairs))

	totalFraud := 0
	for i, p := range pairs {
		// Setup tx: Istanbul, normal amount → sets location cache
		setup := build(p.user, 130.00, p.from, at(p.day, p.hour, 0))
		if err := post(setup); err != nil {
			log.Printf("  [P%02d] setup FAIL %-10s — %v", i+1, p.user, err)
			continue
		}
		fmt.Printf("  [P%02d] setup  %-10s day=-%2d %02d:00  istanbul\n",
			i+1, p.user, p.day, p.hour)

		// Wait for worker to process setup tx before sending fraud tx
		time.Sleep(2 * time.Second)

		// Fraud tx: distant city, high amount, created_at only 5 min later
		fraud := build(p.user, p.amount, p.to, at(p.day, p.hour, 5))
		if err := post(fraud); err != nil {
			log.Printf("  [P%02d] fraud  FAIL %-10s — %v", i+1, p.user, err)
		} else {
			fmt.Printf("  [P%02d] fraud  %-10s day=-%2d %02d:05  distant $%.2f\n",
				i+1, p.user, p.day, p.hour+0, p.amount)
			totalFraud++
		}

		time.Sleep(300 * time.Millisecond)
	}

	fmt.Printf("\nDone. Sent %d fraud transactions across 14 days.\n", totalFraud)
	fmt.Println("Chart should show an ascending trend from day -29 to day -1.")
}
