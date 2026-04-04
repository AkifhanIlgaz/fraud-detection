import { Card, Chip } from "@heroui/react";

import type { TrustScore } from "../types";

type RiskConfig = {
  color: "success" | "warning" | "danger";
  label: string;
  barColor: string;
};

const RISK_CONFIG: Record<TrustScore["risk_level"], RiskConfig> = {
  low: { color: "success", label: "Low Risk", barColor: "var(--success)" },
  medium: {
    color: "warning",
    label: "Medium Risk",
    barColor: "var(--warning)",
  },
  high: { color: "danger", label: "High Risk", barColor: "var(--danger)" },
};

function ScoreBar({ score, color }: { score: number; color: string }) {
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-border">
      <div
        className="h-full rounded-full transition-all duration-500"
        style={{ width: `${score}%`, backgroundColor: color }}
      />
    </div>
  );
}

function Stat({
  label,
  value,
  color,
}: {
  label: string;
  value: React.ReactNode;
  color?: string;
}) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-muted">{label}</span>
      <span
        className="text-xl font-semibold tabular-nums"
        style={color ? { color } : undefined}
      >
        {value}
      </span>
    </div>
  );
}

export function TrustScoreCard({ data }: { data: TrustScore }) {
  const risk = RISK_CONFIG[data.risk_level];
  const fraudRate =
    data.total > 0 ? ((data.fraud_count / data.total) * 100).toFixed(1) : "0.0";

  return (
    <Card>
      <Card.Header>
        <Card.Title>Trust Score</Card.Title>
        <Card.Description>Based on {data.total} transactions</Card.Description>
      </Card.Header>
      <Card.Content className="flex flex-col gap-5">
        <div className="flex items-end justify-between">
          <div className="flex items-end gap-2">
            <span className="text-5xl font-bold tabular-nums">
              {data.score.toFixed(0)}
            </span>
            <span className="mb-1 text-lg text-muted">/ 100</span>
          </div>
          <Chip color={risk.color} variant="soft">
            {risk.label}
          </Chip>
        </div>

        <ScoreBar score={data.score} color={risk.barColor} />

        <div className="grid grid-cols-3 gap-3 border-t border-border pt-4">
          <Stat label="Total" value={data.total} />
          <Stat
            label="Fraud"
            value={data.fraud_count}
            color={data.fraud_count > 0 ? "var(--danger)" : undefined}
          />
          <Stat label="Risk Rate" value={`${fraudRate}%`} />
        </div>
      </Card.Content>
    </Card>
  );
}
