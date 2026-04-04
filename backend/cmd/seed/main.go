package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const apiURL = "http://localhost:8080/api/v1/transactions"

type createTxRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

var transactions = []createTxRequest{
	// --- user-001 ---
	{UserID: "user-001", Amount: 120.50, Lat: 41.0082, Lon: 28.9784},
	{UserID: "user-001", Amount: 4875.00, Lat: 41.0136, Lon: 28.9550},
	{UserID: "user-001", Amount: 38.99, Lat: 41.0195, Lon: 29.0047},
	{UserID: "user-001", Amount: 980.00, Lat: 40.9874, Lon: 28.8456},
	{UserID: "user-001", Amount: 3200.75, Lat: 41.0451, Lon: 29.0123},
	{UserID: "user-001", Amount: 55.20, Lat: 41.0072, Lon: 28.9741},
	{UserID: "user-001", Amount: 215.00, Lat: 41.0310, Lon: 28.9801},
	{UserID: "user-001", Amount: 1450.00, Lat: 41.0890, Lon: 29.0512},
	{UserID: "user-001", Amount: 760.40, Lat: 40.9923, Lon: 29.1023},

	// --- user-002 ---
	{UserID: "user-002", Amount: 89.90, Lat: 39.9208, Lon: 32.8541},
	{UserID: "user-002", Amount: 14500.00, Lat: 39.8765, Lon: 32.7412},
	{UserID: "user-002", Amount: 430.00, Lat: 39.9334, Lon: 32.8597},
	{UserID: "user-002", Amount: 17.50, Lat: 39.9104, Lon: 32.8022},
	{UserID: "user-002", Amount: 6750.00, Lat: 39.8901, Lon: 32.7788},
	{UserID: "user-002", Amount: 320.00, Lat: 39.9450, Lon: 32.8745},
	{UserID: "user-002", Amount: 2890.00, Lat: 39.8542, Lon: 32.7103},
	{UserID: "user-002", Amount: 530.75, Lat: 39.9012, Lon: 32.8301},
	{UserID: "user-002", Amount: 1100.25, Lat: 39.9271, Lon: 32.8643},

	// --- user-003 ---
	{UserID: "user-003", Amount: 250.00, Lat: 38.4189, Lon: 27.1287},
	{UserID: "user-003", Amount: 67.30, Lat: 38.4341, Lon: 27.1402},
	{UserID: "user-003", Amount: 9200.00, Lat: 38.3982, Lon: 27.0991},
	{UserID: "user-003", Amount: 510.80, Lat: 38.4512, Lon: 27.1653},
	{UserID: "user-003", Amount: 135.00, Lat: 38.4078, Lon: 27.1101},
	{UserID: "user-003", Amount: 3750.00, Lat: 38.3701, Lon: 27.0644},
}

func post(tx createTxRequest) error {
	body, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}

func main() {
	fmt.Printf("Sending %d transactions to %s (5s interval)...\n\n", len(transactions), apiURL)

	ok, fail := 0, 0
	for i, tx := range transactions {
		if err := post(tx); err != nil {
			log.Printf("  [%2d] FAIL %-10s $%8.2f  — %v", i+1, tx.UserID, tx.Amount, err)
			fail++
		} else {
			fmt.Printf("  [%2d] OK   %-10s $%8.2f  lat=%.4f lon=%.4f\n",
				i+1, tx.UserID, tx.Amount, tx.Lat, tx.Lon)
			ok++
		}

		if i < len(transactions)-1 {
			time.Sleep(15 * time.Second)
		}
	}

	fmt.Printf("\nDone: %d ok, %d failed.\n", ok, fail)
}
