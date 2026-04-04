"use client";

import { useTransactionStream } from "@/shared/hooks/useTransactionStream";

export function TransactionStreamListener() {
  useTransactionStream();
  return null;
}
