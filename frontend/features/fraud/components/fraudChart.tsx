"use client";

import { useMemo } from "react";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from "chart.js";
import { Line } from "react-chartjs-2";
import { useFraudTransactions } from "../hooks/useFrauds";

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  Legend,
  Filler,
);

// ── Helpers ────────────────────────────────────────────────────────────────

function diffDays(from: string, to: string): number {
  const a = new Date(from).getTime();
  const b = new Date(to).getTime();
  return Math.round((b - a) / 86_400_000);
}

function bucketKey(date: Date, granularity: "hour" | "day"): string {
  if (granularity === "hour") {
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

function buildLabelsAndCounts(
  from: string,
  to: string,
  granularity: "hour" | "day",
): { labels: string[]; counts: Record<string, number> } {
  const labels: string[] = [];
  const counts: Record<string, number> = {};
  const cursor = new Date(from);
  const end = new Date(to);

  while (cursor <= end) {
    const key = bucketKey(cursor, granularity);
    if (!counts[key]) {
      labels.push(key);
      counts[key] = 0;
    }
    if (granularity === "hour") {
      cursor.setHours(cursor.getHours() + 1);
    } else {
      cursor.setDate(cursor.getDate() + 1);
    }
  }

  return { labels, counts };
}

// ── Component ──────────────────────────────────────────────────────────────

interface FraudChartProps {
  from: string;
  to: string;
}

export function FraudChart({ from, to }: FraudChartProps) {
  const days = diffDays(from, to);
  const granularity: "hour" | "day" = days <= 1 ? "hour" : "day";

  const { data, isLoading, isError } = useFraudTransactions(from, to, {
    page: 1,
    limit: 1000,
  });

  const { labels, values } = useMemo(() => {
    const { labels, counts } = buildLabelsAndCounts(from, to, granularity);

    (data?.items ?? []).forEach((tx) => {
      const key = bucketKey(new Date(tx.created_at), granularity);
      if (key in counts) counts[key]++;
    });

    return { labels, values: labels.map((l) => counts[l]) };
  }, [data, from, to, granularity]);

  const chartData = {
    labels,
    datasets: [
      {
        label: "Fraud Transactions",
        data: values,
        backgroundColor: "rgba(239, 68, 68, 0.15)",
        borderColor: "rgba(239, 68, 68, 0.9)",
        borderWidth: 2,
        pointBackgroundColor: "rgba(239, 68, 68, 0.9)",
        pointRadius: 3,
        pointHoverRadius: 5,
        tension: 0.3,
        fill: true,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (ctx: import("chart.js").TooltipItem<"line">) =>
            ` ${ctx.parsed.y} fraud${ctx.parsed.y !== 1 ? "s" : ""}`,
        },
      },
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: {
          maxRotation: 45,
          maxTicksLimit: 20,
          font: { size: 11 },
        },
      },
      y: {
        beginAtZero: true,
        ticks: {
          stepSize: 1,
          font: { size: 11 },
        },
      },
    },
  };

  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-foreground">
          Fraud Activity Over Time
        </h2>
        <span className="text-xs text-muted">
          {granularity === "hour" ? "Hourly" : "Daily"} breakdown
          {data?.meta?.total !== undefined && ` · ${data.meta.total} total`}
        </span>
      </div>
      <div className="h-52">
        {isLoading ? (
          <div className="flex h-full items-center justify-center text-sm text-muted">
            Loading chart…
          </div>
        ) : isError ? (
          <div className="flex h-full items-center justify-center text-sm text-danger">
            Failed to load chart data.
          </div>
        ) : (
          <Line data={chartData} options={options} />
        )}
      </div>
    </div>
  );
}
