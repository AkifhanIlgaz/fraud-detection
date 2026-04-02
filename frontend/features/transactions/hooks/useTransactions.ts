import { useQuery } from "@tanstack/react-query";

import type { PageParams } from "@/shared/types/api";
import { TransactionService } from "../services/transactionService";

export const transactionKeys = {
  all: ["transactions"] as const,
  byUser: (userID: string, params?: PageParams) =>
    [...transactionKeys.all, "user", userID, params] as const,
  trustScore: (userID: string) =>
    [...transactionKeys.all, "trust-score", userID] as const,
};

export function useUserTransactions(userID: string, params?: PageParams) {
  return useQuery({
    queryKey: transactionKeys.byUser(userID, params),
    queryFn: () => TransactionService.getByUser(userID, params),
    enabled: !!userID,
  });
}

export function useTrustScore(userID: string) {
  return useQuery({
    queryKey: transactionKeys.trustScore(userID),
    queryFn: () => TransactionService.getTrustScore(userID),
    enabled: !!userID,
  });
}
