"use client";

import { useMemo, useState } from "react";

import type { Transaction } from "@/features/transactions/types";
import { useFraudTransactions } from "../hooks/useFrauds";
import {
  Button,
  Chip,
  ListBox,
  Pagination,
  Select,
  Table,
  toast,
} from "@heroui/react";

type ChipColor = "default" | "accent" | "success" | "warning" | "danger";

const REASON_CONFIG: Record<string, { label: string; color: ChipColor }> = {
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

const ALL_REASON_KEYS = Object.keys(REASON_CONFIG);

function getReasonConfig(reason: string): { label: string; color: ChipColor } {
  return (
    REASON_CONFIG[reason] ?? {
      label: reason.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase()),
      color: "default",
    }
  );
}

// ── Icons ──────────────────────────────────────────────────────────────────

function CopyIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function CopyableId({ id }: { id: string }) {
  const [copied, setCopied] = useState(false);

  function handleCopy() {
    navigator.clipboard.writeText(id);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
    toast.success("Copied!", { description: id, timeout: 2000 });
  }

  return (
    <button
      onClick={handleCopy}
      title={`Copy: ${id}`}
      className="group flex cursor-pointer items-center gap-1.5 rounded px-1 py-0.5 transition-colors hover:bg-border/50"
    >
      <span className="font-mono text-xs text-muted">…{id.slice(-8)}</span>
      {copied ? (
        <CheckIcon className="shrink-0 text-success" />
      ) : (
        <CopyIcon className="shrink-0 text-muted opacity-50 transition-opacity group-hover:opacity-100" />
      )}
    </button>
  );
}

// ── Pagination helpers ─────────────────────────────────────────────────────

function getPageRange(current: number, total: number): (number | "…")[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  const pages: (number | "…")[] = [1];
  if (current > 3) pages.push("…");
  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);
  for (let i = start; i <= end; i++) pages.push(i);
  if (current < total - 2) pages.push("…");
  pages.push(total);
  return pages;
}

// ── Types ──────────────────────────────────────────────────────────────────

type SortCol = "amount" | "created_at";
type SortDir = "ascending" | "descending";

const PAGE_SIZE_OPTIONS = [10, 20, 50] as const;
type PageSizeOption = (typeof PAGE_SIZE_OPTIONS)[number];

// ── Main component ─────────────────────────────────────────────────────────

export function FraudTable({ from, to }: { from: string; to: string }) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState<PageSizeOption>(20);
  const [sortCol, setSortCol] = useState<SortCol>("created_at");
  const [sortDir, setSortDir] = useState<SortDir>("descending");
  const [selectedReasons, setSelectedReasons] = useState<string[]>([
    ...ALL_REASON_KEYS,
  ]);

  const { data, isLoading, isError, error } = useFraudTransactions(from, to, {
    page,
    limit: pageSize,
  });

  const rawItems: Transaction[] = data?.items ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  const pageReasons = useMemo(() => {
    const set = new Set<string>();
    rawItems.forEach((tx) => tx.fraud_reasons?.forEach((r) => set.add(r)));
    return Array.from(set);
  }, [rawItems]);

  const displayItems = useMemo(() => {
    const filtered = rawItems.filter((tx) => {
      if (!tx.fraud_reasons?.length) return true;
      return tx.fraud_reasons.some((r) => selectedReasons.includes(r));
    });
    return [...filtered].sort((a, b) => {
      const cmp =
        sortCol === "amount"
          ? a.amount - b.amount
          : new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      return sortDir === "ascending" ? cmp : -cmp;
    });
  }, [rawItems, selectedReasons, sortCol, sortDir]);

  const isFiltered = selectedReasons.length < ALL_REASON_KEYS.length;

  const filteredInfo = isFiltered
    ? `${displayItems.length} filtered / ${meta?.total ?? 0} total`
    : meta
      ? `${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, meta.total)} of ${meta.total}`
      : null;

  return (
    <div className="flex flex-col gap-4">
      <div className="overflow-hidden rounded-xl border">
        {/* Filter bar */}
        <div className="flex flex-wrap items-center gap-3 border-b border-border bg-default/60 px-4 py-2">
          <Select
            selectionMode="multiple"
            value={selectedReasons}
            onChange={(val) => {
              setSelectedReasons((val as string[]) ?? []);
              setPage(1);
            }}
            className="w-40"
          >
            <Select.Trigger className="h-8 text-xs">
              <Select.Value>
                {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                {({ selectedItems }: any) =>
                  selectedReasons.length === ALL_REASON_KEYS.length ||
                  !selectedItems?.length ? (
                    <span className="text-muted">Reason</span>
                  ) : (
                    <span>Reason ({selectedItems.length})</span>
                  )
                }
              </Select.Value>
              <Select.Indicator />
            </Select.Trigger>
            <Select.Popover>
              <ListBox selectionMode="multiple">
                {pageReasons.length > 0 ? (
                  pageReasons.map((reason) => {
                    const cfg = getReasonConfig(reason);
                    return (
                      <ListBox.Item
                        key={reason}
                        id={reason}
                        textValue={cfg.label}
                      >
                        <div className="flex items-center justify-between gap-3 py-0.5">
                          <Chip color={cfg.color} size="sm" variant="soft">
                            {cfg.label}
                          </Chip>
                          <ListBox.ItemIndicator />
                        </div>
                      </ListBox.Item>
                    );
                  })
                ) : (
                  <ListBox.Item id="__empty" isDisabled textValue="No reasons">
                    <span className="text-xs text-muted">
                      No fraud reasons on this page
                    </span>
                  </ListBox.Item>
                )}
              </ListBox>
            </Select.Popover>
          </Select>

          {filteredInfo && (
            <span className="text-xs text-muted">{filteredInfo}</span>
          )}

          <div className="ml-auto flex items-center gap-3">
            {isFiltered && (
              <Button
                size="sm"
                variant="outline"
                onPress={() => {
                  setSelectedReasons([...ALL_REASON_KEYS]);
                  setPage(1);
                }}
              >
                Clear
              </Button>
            )}
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted">Per page</span>
              <Select
                value={String(pageSize)}
                onChange={(val) => {
                  if (val) {
                    setPageSize(Number(val) as PageSizeOption);
                    setPage(1);
                  }
                }}
                className="w-20"
              >
                <Select.Trigger className="h-8 text-xs">
                  <Select.Value />
                  <Select.Indicator />
                </Select.Trigger>
                <Select.Popover>
                  <ListBox>
                    {PAGE_SIZE_OPTIONS.map((n) => (
                      <ListBox.Item
                        key={n}
                        id={String(n)}
                        textValue={String(n)}
                      >
                        {n}
                      </ListBox.Item>
                    ))}
                  </ListBox>
                </Select.Popover>
              </Select>
            </div>
          </div>
        </div>

        {/* Table */}
        <Table className="rounded-none border-0">
          <Table.ScrollContainer>
            <Table.Content
              aria-label="Fraud transactions"
              className="w-full table-fixed"
              sortDescriptor={{ column: sortCol, direction: sortDir }}
              onSortChange={(d) => {
                setSortCol(d.column as SortCol);
                setSortDir(d.direction as SortDir);
              }}
            >
              <Table.Header>
                <Table.Column id="id" isRowHeader className="w-36">
                  ID
                </Table.Column>
                <Table.Column id="user_id" className="w-36">
                  User
                </Table.Column>
                <Table.Column id="amount" allowsSorting className="w-28">
                  Amount
                </Table.Column>
                <Table.Column id="location" className="w-36">
                  Location
                </Table.Column>
                <Table.Column id="created_at" allowsSorting className="w-36">
                  Date
                </Table.Column>
                <Table.Column id="fraud_reasons" className="w-56">
                  Reasons
                </Table.Column>
              </Table.Header>

              <Table.Body
                renderEmptyState={() => (
                  <div className="py-10 text-center text-sm text-muted">
                    {isLoading
                      ? "Loading…"
                      : isError
                        ? `Error: ${error?.message}`
                        : "No fraud transactions found in this date range."}
                  </div>
                )}
              >
                {displayItems.map((tx) => (
                  <Table.Row key={tx.id}>
                    <Table.Cell>
                      <CopyableId id={tx.id} />
                    </Table.Cell>
                    <Table.Cell>
                      <span className="font-mono text-xs text-muted">
                        …{tx.user_id.slice(-8)}
                      </span>
                    </Table.Cell>
                    <Table.Cell>
                      <span className="font-semibold tabular-nums">
                        ${tx.amount.toFixed(2)}
                      </span>
                    </Table.Cell>
                    <Table.Cell>
                      <span className="font-mono text-xs text-muted">
                        {tx.lat.toFixed(4)}, {tx.lon.toFixed(4)}
                      </span>
                    </Table.Cell>
                    <Table.Cell>
                      <span className="text-sm">
                        {new Date(tx.created_at).toLocaleString("tr-TR", {
                          dateStyle: "short",
                          timeStyle: "short",
                        })}
                      </span>
                    </Table.Cell>
                    <Table.Cell>
                      <div className="flex flex-wrap gap-1">
                        {tx.fraud_reasons?.map((reason) => {
                          const cfg = getReasonConfig(reason);
                          return (
                            <Chip
                              key={reason}
                              color={cfg.color}
                              size="sm"
                              variant="soft"
                            >
                              {cfg.label}
                            </Chip>
                          );
                        })}
                      </div>
                    </Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table.Content>
          </Table.ScrollContainer>
        </Table>
      </div>

      {totalPages > 1 && (
        <Pagination className="flex w-full items-center justify-center">
          <Pagination.Content>
            <Pagination.Item>
              <Pagination.Previous
                isDisabled={page === 1}
                onPress={() => setPage((p) => Math.max(1, p - 1))}
              >
                <Pagination.PreviousIcon />
              </Pagination.Previous>
            </Pagination.Item>
            {getPageRange(page, totalPages).map((p, i) =>
              p === "…" ? (
                <Pagination.Item key={`ellipsis-${i}`}>
                  <Pagination.Ellipsis />
                </Pagination.Item>
              ) : (
                <Pagination.Item key={p}>
                  <Pagination.Link
                    isActive={p === page}
                    onPress={() => setPage(p as number)}
                  >
                    {p}
                  </Pagination.Link>
                </Pagination.Item>
              ),
            )}
            <Pagination.Item>
              <Pagination.Next
                isDisabled={page === totalPages}
                onPress={() => setPage((p) => Math.min(totalPages, p + 1))}
              >
                <Pagination.NextIcon />
              </Pagination.Next>
            </Pagination.Item>
          </Pagination.Content>
        </Pagination>
      )}
    </div>
  );
}
