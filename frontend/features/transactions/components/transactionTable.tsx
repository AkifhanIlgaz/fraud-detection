"use client";

import { useMemo, useState } from "react";

import {
  Button,
  Chip,
  ListBox,
  Pagination,
  Select,
  Table,
  toast,
} from "@heroui/react";

import {
  ALL_REASON_KEYS,
  getReasonConfig,
} from "@/shared/constants/fraudReasons";
import type { PageSizeOption } from "@/shared/constants/pagination";
import { PAGE_SIZE_OPTIONS } from "@/shared/constants/pagination";
import { Check, Copy } from "lucide-react";
import { ALL_STATUSES } from "../constants";
import { useUserTransactions } from "../hooks/useTransactions";
import { StatusChip } from "./statusChip";

function CopyableId({ id }: { id: string }) {
  const [copied, setCopied] = useState(false);

  function handleCopy() {
    navigator.clipboard.writeText(id);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
    toast.success("Copied!", {
      description: id,
      timeout: 2000,
    });
  }

  return (
    <button
      onClick={handleCopy}
      title={`Copy: ${id}`}
      aria-label={`Copy transaction ID: ${id}`}
      className="group flex cursor-pointer items-center gap-1.5 rounded px-1 py-0.5 transition-colors hover:bg-border/50"
    >
      <span className="font-mono text-xs text-muted">…{id.slice(-8)}</span>
      {copied ? (
        <Check className="size-3 shrink-0 text-success" />
      ) : (
        <Copy className="shrink-0 size-3 text-muted opacity-50 transition-opacity group-hover:opacity-100" />
      )}
    </button>
  );
}

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

type SortCol = "amount" | "created_at";
type SortDir = "ascending" | "descending";

function FilterSelect({
  triggerLabel,
  totalCount,
  selected,
  onChange,
  children,
}: {
  triggerLabel: string;
  totalCount: number;
  selected: string[];
  onChange: (val: string[]) => void;
  children: React.ReactNode;
}) {
  const isAll = selected.length === totalCount;

  return (
    <Select
      selectionMode="multiple"
      value={selected}
      onChange={(val) => onChange((val as string[]) ?? [])}
      aria-label={triggerLabel}
      className="w-40"
    >
      <Select.Trigger className="h-8 text-xs">
        <Select.Value>
          {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
          {({ selectedItems }: any) =>
            isAll || !selectedItems?.length ? (
              <span className="text-muted">{triggerLabel}</span>
            ) : (
              <span>
                {triggerLabel} ({selectedItems.length})
              </span>
            )
          }
        </Select.Value>
        <Select.Indicator />
      </Select.Trigger>
      <Select.Popover>
        <ListBox selectionMode="multiple">{children}</ListBox>
      </Select.Popover>
    </Select>
  );
}

export function TransactionTable({ userID }: { userID: string }) {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState<PageSizeOption>(20);
  const [sortCol, setSortCol] = useState<SortCol>("created_at");
  const [sortDir, setSortDir] = useState<SortDir>("descending");
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([
    ...ALL_STATUSES,
  ]);
  const [selectedReasons, setSelectedReasons] = useState<string[]>([
    ...ALL_REASON_KEYS,
  ]);

  const { data, isLoading, isError, error } = useUserTransactions(userID, {
    page,
    limit: pageSize,
  });

  const rawItems = data?.items ?? [];
  const meta = data?.meta;
  const totalPages = meta?.total_pages ?? 1;

  // Unique reasons visible on current page — only show known ones
  const pageReasons = useMemo(() => {
    const set = new Set<string>();
    rawItems.forEach((tx) => tx.fraud_reasons?.forEach((r) => set.add(r)));
    return Array.from(set);
  }, [rawItems]);

  // Filter → Sort
  const displayItems = useMemo(() => {
    const filtered = rawItems.filter((tx) => {
      if (!selectedStatuses.includes(tx.status)) return false;
      // Transactions without fraud reasons always pass the reason filter
      if (tx.fraud_reasons && tx.fraud_reasons.length > 0) {
        if (!tx.fraud_reasons.some((r) => selectedReasons.includes(r)))
          return false;
      }
      return true;
    });

    return [...filtered].sort((a, b) => {
      const cmp =
        sortCol === "amount"
          ? a.amount - b.amount
          : new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      return sortDir === "ascending" ? cmp : -cmp;
    });
  }, [rawItems, selectedStatuses, selectedReasons, sortCol, sortDir]);

  function handleSortChange(descriptor: {
    column: React.Key;
    direction: SortDir;
  }) {
    setSortCol(descriptor.column as SortCol);
    setSortDir(descriptor.direction);
  }

  function handlePageSizeChange(val: React.Key | null) {
    if (val) {
      setPageSize(Number(val) as PageSizeOption);
      setPage(1);
    }
  }

  const isFiltered =
    selectedStatuses.length < ALL_STATUSES.length ||
    selectedReasons.length < ALL_REASON_KEYS.length;

  const filteredInfo = isFiltered
    ? `${displayItems.length} filtered / ${meta?.total ?? 0} total`
    : meta
      ? `${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, meta.total)} of ${meta.total}`
      : null;

  return (
    <div className="flex flex-col gap-4">
      <div className="overflow-hidden rounded-xl border">
        <div className="flex flex-wrap items-center gap-3 border-b border-border bg-default/60 px-4 py-2">
          {/* Status dropdown */}
          <FilterSelect
            triggerLabel="Status"
            totalCount={ALL_STATUSES.length}
            selected={selectedStatuses}
            onChange={(v) => {
              setSelectedStatuses(v);
              setPage(1);
            }}
          >
            {ALL_STATUSES.map((s) => (
              <ListBox.Item key={s} id={s} textValue={s}>
                <div className="flex items-center justify-between gap-3 py-0.5">
                  <StatusChip status={s} />
                  <ListBox.ItemIndicator />
                </div>
              </ListBox.Item>
            ))}
          </FilterSelect>

          {/* Reason dropdown */}
          <FilterSelect
            triggerLabel="Reason"
            totalCount={ALL_REASON_KEYS.length}
            selected={selectedReasons}
            onChange={(v) => {
              setSelectedReasons(v);
              setPage(1);
            }}
          >
            {pageReasons.length > 0 ? (
              pageReasons.map((reason) => {
                const cfg = getReasonConfig(reason);
                return (
                  <ListBox.Item key={reason} id={reason} textValue={cfg.label}>
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
          </FilterSelect>

          {/* Filtered count — center */}
          {filteredInfo && (
            <span className="text-xs text-muted">{filteredInfo}</span>
          )}

          {/* Right: clear + per page */}
          <div className="ml-auto flex items-center gap-3">
            {isFiltered && (
              <Button
                size="sm"
                variant="outline"
                onPress={() => {
                  setSelectedStatuses([...ALL_STATUSES]);
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
                onChange={handlePageSizeChange}
                aria-label="Items per page"
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

        {/* ── Table ── */}
        <Table className="rounded-none border-0">
          <Table.ScrollContainer>
            <Table.Content
              aria-label="Transaction history"
              className="w-full table-fixed"
              sortDescriptor={{ column: sortCol, direction: sortDir }}
              onSortChange={handleSortChange}
            >
              <Table.Header>
                <Table.Column id="id" isRowHeader className="w-36">
                  ID
                </Table.Column>
                <Table.Column id="amount" allowsSorting className="w-28">
                  Amount
                </Table.Column>
                <Table.Column id="status" className="w-28">
                  Status
                </Table.Column>
                <Table.Column id="location" className="w-36">
                  Location
                </Table.Column>
                <Table.Column id="created_at" allowsSorting className="w-36">
                  Date
                </Table.Column>
                <Table.Column id="fraud_reasons" className="w-48">
                  Fraud Reasons
                </Table.Column>
              </Table.Header>

              <Table.Body
                renderEmptyState={() => (
                  <div className="py-10 text-center text-sm text-muted">
                    {isLoading
                      ? "Loading transactions…"
                      : isError
                        ? `Error: ${error?.message}`
                        : isFiltered
                          ? "No transactions match the current filters."
                          : "No transactions found."}
                  </div>
                )}
              >
                {displayItems.map((tx) => (
                  <Table.Row key={tx.id}>
                    <Table.Cell>
                      <CopyableId id={tx.id} />
                    </Table.Cell>
                    <Table.Cell>
                      <span className="font-semibold tabular-nums">
                        ${tx.amount.toFixed(2)}
                      </span>
                    </Table.Cell>
                    <Table.Cell>
                      <StatusChip status={tx.status} />
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
                      {tx.fraud_reasons && tx.fraud_reasons.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {tx.fraud_reasons.map((reason) => {
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
                      ) : (
                        <span className="text-xs text-success">No flags</span>
                      )}
                    </Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table.Content>
          </Table.ScrollContainer>
        </Table>
      </div>

      {/* ── Pagination — outside wrapper, centered ── */}
      {totalPages > 1 && (
        <Pagination className="flex w-full justify-center items-center">
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
