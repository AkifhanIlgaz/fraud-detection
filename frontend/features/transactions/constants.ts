import type { Transaction } from "./types";

type StatusValue = Transaction["status"];

export const ALL_STATUSES: StatusValue[] = [
  "pending",
  "approved",
  "suspicious",
  "fraud",
];
