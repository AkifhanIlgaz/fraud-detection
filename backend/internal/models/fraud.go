package models

type FraudReason string

const (
	ReasonVelocity         FraudReason = "velocity_limit_exceeded"
	ReasonAmountAnomaly    FraudReason = "amount_anomaly"
	ReasonImpossibleTravel FraudReason = "impossible_travel"
	ReasonNone             FraudReason = "none"
)
