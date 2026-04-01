package models

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "pending"
	StatusApproved   TransactionStatus = "approved"
	StatusSuspicious TransactionStatus = "suspicious"
	StatusFraud      TransactionStatus = "fraud"
)

func (s TransactionStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusSuspicious, StatusFraud:
		return true
	}

	return false
}

func (s TransactionStatus) String() string {
	return string(s)
}
