import { useMutation, useQueryClient } from "@tanstack/react-query";

import type { CreateTransactionInput, UpdateStatusInput } from "../schemas";
import { TransactionService } from "../services/transactionService";
import { transactionKeys } from "./useTransactions";

export function useCreateTransaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTransactionInput) => TransactionService.create(data),
    onSuccess: (transaction) => {
      queryClient.invalidateQueries({
        queryKey: transactionKeys.byUser(transaction.user_id),
      });
      queryClient.invalidateQueries({
        queryKey: transactionKeys.trustScore(transaction.user_id),
      });
    },
  });
}

export function useUpdateTransactionStatus(transactionId: string, userID: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateStatusInput) =>
      TransactionService.updateStatus(transactionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: transactionKeys.byUser(userID),
      });
      queryClient.invalidateQueries({
        queryKey: transactionKeys.all,
      });
    },
  });
}
