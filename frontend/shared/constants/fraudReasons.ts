export type ChipColor = "default" | "accent" | "success" | "warning" | "danger";

export const REASON_CONFIG: Record<string, { label: string; color: ChipColor }> = {
  amount_exceeds_daily_limit: { label: "Exceeds Daily Limit", color: "danger" },
  unusual_location: { label: "Unusual Location", color: "warning" },
  multiple_transactions_short_interval: {
    label: "Multiple Rapid Txns",
    color: "warning",
  },
  high_risk_merchant: { label: "High Risk Merchant", color: "danger" },
  card_not_present: { label: "Card Not Present", color: "accent" },
  ip_country_mismatch: { label: "IP / Country Mismatch", color: "danger" },
  velocity_check_failed: { label: "Velocity Check Failed", color: "warning" },
  suspicious_device_fingerprint: {
    label: "Suspicious Device",
    color: "accent",
  },
  foreign_currency_mismatch: { label: "Currency Mismatch", color: "default" },
};

export const ALL_REASON_KEYS = Object.keys(REASON_CONFIG);

export function getReasonConfig(reason: string): { label: string; color: ChipColor } {
  return (
    REASON_CONFIG[reason] ?? {
      label: reason.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase()),
      color: "default",
    }
  );
}
