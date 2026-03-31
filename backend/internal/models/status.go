package models

type TransactionStatus string

const (
	StatusPending  TransactionStatus = "pending"
	StatusApproved TransactionStatus = "approved"
	StatusFraud    TransactionStatus = "fraud"
)

func (s TransactionStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusFraud:
		return true
	}

	return false
}

func (s TransactionStatus) String() string {
	return string(s)
}
