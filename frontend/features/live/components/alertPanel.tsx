"use client";

import { Chip } from "@heroui/react";
import { useQuery } from "@tanstack/react-query";
import { ArrowRight, ShieldAlert } from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { getReasonConfig } from "@/shared/constants/fraudReasons";
import {
  ALERT_FEED_KEY,
  type TransactionEvent,
} from "@/shared/hooks/useTransactionStream";

const STATUS_CONFIG = {
  suspicious: {
    label: "Suspicious",
    color: "warning" as const,
    animation: "animate-card-enter-suspicious",
  },
  fraud: {
    label: "Fraud",
    color: "danger" as const,
    animation: "animate-card-enter-fraud",
  },
};

function AlertCard({ tx }: { tx: TransactionEvent }) {
  const cfg = STATUS_CONFIG[tx.status as keyof typeof STATUS_CONFIG];

  return (
    <div
      className={`rounded-lg border border-border bg-surface p-3 ${cfg.animation}`}
    >
      <div className="flex items-center justify-between gap-2">
        <Chip color={cfg.color} size="sm" variant="soft">
          {cfg.label}
        </Chip>
        <span className="font-semibold tabular-nums text-sm">
          ${tx.amount.toFixed(2)}
        </span>
      </div>

      <div className="mt-1.5 flex items-center justify-between">
        <span className="font-mono text-xs text-muted">
          …{tx.user_id.slice(-8)}
        </span>
        <span className="text-xs text-muted">
          {new Date(tx.created_at).toLocaleTimeString("tr-TR")}
        </span>
      </div>

      {tx.fraud_reasons?.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {tx.fraud_reasons.map((r) => {
            const rc = getReasonConfig(r);
            return (
              <Chip key={r} color={rc.color} size="sm" variant="soft">
                {rc.label}
              </Chip>
            );
          })}
        </div>
      )}

      <div className="mt-2.5">
        <Link
          href={`/users/${tx.user_id}`}
          className="inline-flex items-center gap-1 rounded-md border border-border px-2.5 py-1 text-xs font-medium text-foreground transition-colors hover:bg-default"
        >
          Go to User
          <ArrowRight size={11} />
        </Link>
      </div>
    </div>
  );
}

export function AlertPanel() {
  const { data: alerts = [] } = useQuery<TransactionEvent[]>({
    queryKey: ALERT_FEED_KEY,
    queryFn: () => [],
    staleTime: Infinity,
  });

  const [fraudPulsing, setFraudPulsing] = useState(false);
  const prevFirstIdRef = useRef<string | undefined>(null);

  useEffect(() => {
    const first = alerts[0];
    if (
      first?.status === "fraud" &&
      first.transaction_id !== prevFirstIdRef.current
    ) {
      prevFirstIdRef.current = first.transaction_id;
      setFraudPulsing(true);
    }
  }, [alerts]);

  const hasFraud = alerts.some((a) => a.status === "fraud");

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <ShieldAlert size={15} className="text-danger" />
        <h2 className="text-sm font-semibold">Alert Panel</h2>
        {hasFraud && (
          <span className="h-2 w-2 rounded-full bg-danger animate-pulse" />
        )}
        <span className="ml-auto rounded-full bg-default px-2 py-0.5 text-xs text-muted">
          {alerts.length} / {10}
        </span>
      </div>

      <div
        className={`rounded-xl border transition-[border-color,box-shadow] duration-300 ${
          fraudPulsing
            ? "animate-fraud-pulse border-danger/40"
            : "border-border"
        }`}
        onAnimationEnd={() => setFraudPulsing(false)}
      >
        {alerts.length === 0 ? (
          <div className="py-12 text-center text-sm text-muted">
            No alerts yet
          </div>
        ) : (
          <div className="flex flex-col gap-2 p-3">
            {alerts.map((tx) => (
              <AlertCard key={tx.transaction_id} tx={tx} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
