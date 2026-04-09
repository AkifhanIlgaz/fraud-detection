"use client";

import { useQueryClient } from "@tanstack/react-query";
import { toast } from "@heroui/react";
import { useEffect, useRef } from "react";

export interface TransactionEvent {
  transaction_id: string;
  user_id: string;
  status: "approved" | "fraud" | "suspicious";
  amount: number;
  fraud_reasons: string[];
  created_at: string;
}

export const LIVE_FEED_KEY = ["live-feed"] as const;
export const ALERT_FEED_KEY = ["alert-feed"] as const;

const FEED_LIMIT = 25;
const ALERT_LIMIT = 10;
const BATCH_MS = 100;

export function useTransactionStream() {
  const queryClient = useQueryClient();
  const buffer = useRef<TransactionEvent[]>([]);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const wsUrl =
      process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8081/transactions";
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event: MessageEvent) => {
      let tx: TransactionEvent;
      try {
        tx = JSON.parse(event.data as string) as TransactionEvent;
      } catch {
        return;
      }

      buffer.current.push(tx);

      if (!timerRef.current) {
        timerRef.current = setTimeout(() => {
          const batch = buffer.current.splice(0);
          timerRef.current = null;

          // Live feed — tüm statuslar, son 25
          queryClient.setQueryData<TransactionEvent[]>(
            LIVE_FEED_KEY,
            (prev = []) => [...batch, ...prev].slice(0, FEED_LIMIT),
          );

          // Alert feed — sadece fraud/suspicious, son 10
          const critical = batch.filter(
            (t) => t.status === "fraud" || t.status === "suspicious",
          );
          if (critical.length > 0) {
            queryClient.setQueryData<TransactionEvent[]>(
              ALERT_FEED_KEY,
              (prev = []) => [...critical, ...prev].slice(0, ALERT_LIMIT),
            );
          }

          // Toast bildirimleri
          for (const t of critical) {
            const shortUser = `…${t.user_id.slice(-8)}`;
            const amount = `$${t.amount.toFixed(2)}`;
            const opts = {
              description: `User ${shortUser}`,
              timeout: t.status === "fraud" ? 8000 : 6000,
              actionProps: {
                children: "View User",
                size: "sm" as const,
                onPress: () => {
                  window.location.href = `/users/${t.user_id}`;
                },
              },
            };

            if (t.status === "fraud") {
              toast.danger(`Fraud Detected — ${amount}`, opts);
            } else {
              toast.warning(`Suspicious Transaction — ${amount}`, opts);
            }
          }
        }, BATCH_MS);
      }
    };

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
      ws.close();
    };
  }, [queryClient]);
}
