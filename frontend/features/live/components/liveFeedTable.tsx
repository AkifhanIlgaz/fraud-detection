"use client";

import { useQuery } from "@tanstack/react-query";
import { Chip, Table } from "@heroui/react";
import { Activity } from "lucide-react";

import {
  LIVE_FEED_KEY,
  type TransactionEvent,
} from "@/shared/hooks/useTransactionStream";
import { getReasonConfig } from "@/shared/constants/fraudReasons";

const STATUS_CONFIG: Record<
  TransactionEvent["status"],
  { label: string; color: "success" | "warning" | "danger"; animation: string }
> = {
  approved:   { label: "Approved",   color: "success", animation: "animate-enter-approved" },
  suspicious: { label: "Suspicious", color: "warning", animation: "animate-enter-suspicious" },
  fraud:      { label: "Fraud",      color: "danger",  animation: "animate-enter-fraud" },
};

export function LiveFeedTable() {
  const { data: transactions = [] } = useQuery<TransactionEvent[]>({
    queryKey: LIVE_FEED_KEY,
    queryFn: () => [],
    staleTime: Infinity,
  });

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <Activity size={15} className="text-accent" />
        <h2 className="text-sm font-semibold">Live Transaction Feed</h2>
        <span className="ml-auto rounded-full bg-default px-2 py-0.5 text-xs text-muted">
          {transactions.length} / {25}
        </span>
      </div>

      <div className="overflow-hidden rounded-xl border border-border">
        <Table className="rounded-none border-0">
          <Table.ScrollContainer>
            <Table.Content
              aria-label="Live transaction feed"
              className="w-full table-fixed"
            >
              <Table.Header>
                <Table.Column id="id" isRowHeader className="w-32">
                  ID
                </Table.Column>
                <Table.Column id="amount" className="w-24">
                  Amount
                </Table.Column>
                <Table.Column id="status" className="w-28">
                  Status
                </Table.Column>
                <Table.Column id="user" className="w-32">
                  User
                </Table.Column>
                <Table.Column id="time" className="w-24">
                  Time
                </Table.Column>
                <Table.Column id="reasons" className="w-52">
                  Reasons
                </Table.Column>
              </Table.Header>

              <Table.Body
                renderEmptyState={() => (
                  <div className="py-12 text-center text-sm text-muted">
                    Waiting for transactions…
                  </div>
                )}
              >
                {transactions.map((tx) => {
                  const cfg = STATUS_CONFIG[tx.status];
                  return (
                    <Table.Row
                      key={tx.transaction_id}
                      className={cfg.animation}
                    >
                      <Table.Cell>
                        <span className="font-mono text-xs text-muted">
                          …{tx.transaction_id.slice(-8)}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        <span className="font-semibold tabular-nums">
                          ${tx.amount.toFixed(2)}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        <Chip color={cfg.color} size="sm" variant="soft">
                          {cfg.label}
                        </Chip>
                      </Table.Cell>
                      <Table.Cell>
                        <span className="font-mono text-xs text-muted">
                          …{tx.user_id.slice(-8)}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        <span className="text-xs text-muted">
                          {new Date(tx.created_at).toLocaleTimeString("tr-TR")}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        {tx.fraud_reasons?.length > 0 ? (
                          <div className="flex flex-wrap gap-1">
                            {tx.fraud_reasons.map((r) => {
                              const rc = getReasonConfig(r);
                              return (
                                <Chip
                                  key={r}
                                  color={rc.color}
                                  size="sm"
                                  variant="soft"
                                >
                                  {rc.label}
                                </Chip>
                              );
                            })}
                          </div>
                        ) : (
                          <span className="text-xs text-success">Clean</span>
                        )}
                      </Table.Cell>
                    </Table.Row>
                  );
                })}
              </Table.Body>
            </Table.Content>
          </Table.ScrollContainer>
        </Table>
      </div>
    </div>
  );
}
