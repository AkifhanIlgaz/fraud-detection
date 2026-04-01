import { Chip } from "@heroui/react";

import type { Transaction } from "../types";

type StatusConfig = {
  color: "default" | "success" | "warning" | "danger";
  label: string;
};

const STATUS_CONFIG: Record<Transaction["status"], StatusConfig> = {
  pending: { color: "default", label: "Pending" },
  approved: { color: "success", label: "Approved" },
  suspicious: { color: "warning", label: "Suspicious" },
  fraud: { color: "danger", label: "Fraud" },
};

export function StatusChip({ status }: { status: Transaction["status"] }) {
  const config = STATUS_CONFIG[status] ?? STATUS_CONFIG.pending;
  return (
    <Chip color={config.color} size="sm" variant="soft">
      {config.label}
    </Chip>
  );
}
