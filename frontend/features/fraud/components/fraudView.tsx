"use client";

import { Button } from "@heroui/react";
import { ChevronLeft } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { ThemeToggle } from "@/shared/components/themeToggle";
import { FraudTable } from "./fraudTable";

function toDateInput(date: Date): string {
  return date.toISOString().slice(0, 10);
}

function toFromISO(dateStr: string): string {
  return `${dateStr}T00:00:00.000Z`;
}

function toToISO(dateStr: string): string {
  return `${dateStr}T23:59:59.999Z`;
}

function today() {
  return toDateInput(new Date());
}

function daysAgo(n: number) {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return toDateInput(d);
}

const PRESETS = [
  { label: "Today", from: () => today(), to: () => today() },
  { label: "Last 7 days", from: () => daysAgo(7), to: () => today() },
  { label: "Last 30 days", from: () => daysAgo(30), to: () => today() },
  { label: "Last 90 days", from: () => daysAgo(90), to: () => today() },
] as const;

export function FraudView() {
  const router = useRouter();
  const [fromDate, setFromDate] = useState(() => daysAgo(7));
  const [toDate, setToDate] = useState(() => today());

  const activePreset =
    PRESETS.find((p) => p.from() === fromDate && p.to() === toDate)?.label ??
    null;

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6 px-4 py-8">
      <div className="flex items-center gap-4">
        <Button variant="ghost" onPress={() => router.push("/")}>
          <ChevronLeft aria-hidden />
          Back
        </Button>
        <div className="flex-1">
          <h1 className="text-xl font-semibold">Fraud Transactions</h1>
          <p className="text-sm text-muted">
            Browse fraud activity by date range
          </p>
        </div>
        <ThemeToggle />
      </div>

      {/* Date range controls */}
      <div className="flex flex-wrap items-end gap-4 rounded-xl border border-border bg-surface p-4">
        {/* Preset buttons */}
        <div className="flex flex-wrap gap-2">
          {PRESETS.map((preset) => (
            <Button
              key={preset.label}
              size="sm"
              variant={activePreset === preset.label ? "primary" : "outline"}
              onPress={() => {
                setFromDate(preset.from());
                setToDate(preset.to());
              }}
            >
              {preset.label}
            </Button>
          ))}
        </div>

        {/* Custom date inputs */}
        <div className="flex items-end gap-3 ml-auto">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-muted">From</label>
            <input
              type="date"
              value={fromDate}
              max={toDate}
              onChange={(e) => setFromDate(e.target.value)}
              className="h-9 rounded-lg border border-border bg-field-background px-3 text-sm text-field-foreground focus:outline-none focus:ring-2 focus:ring-focus"
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-muted">To</label>
            <input
              type="date"
              value={toDate}
              min={fromDate}
              max={today()}
              onChange={(e) => setToDate(e.target.value)}
              className="h-9 rounded-lg border border-border bg-field-background px-3 text-sm text-field-foreground focus:outline-none focus:ring-2 focus:ring-focus"
            />
          </div>
        </div>
      </div>

      {/* Table */}
      <FraudTable from={toFromISO(fromDate)} to={toToISO(toDate)} />
    </div>
  );
}
