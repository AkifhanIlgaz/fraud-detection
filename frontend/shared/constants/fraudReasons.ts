export type ChipColor = "default" | "accent" | "success" | "warning" | "danger";

export const REASON_CONFIG: Record<string, { label: string; color: ChipColor }> = {
  velocity_limit_exceeded: { label: "Velocity Limit Exceeded", color: "warning" },
  amount_anomaly:          { label: "Amount Anomaly",          color: "danger" },
  impossible_travel:       { label: "Impossible Travel",       color: "danger" },
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
