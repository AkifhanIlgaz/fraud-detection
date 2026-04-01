import { apiGet, apiPatch, apiPost } from "@/shared/lib/http";
import type { PageParams, PaginatedResponse } from "@/shared/types/api";
import type { CreateTransactionInput, UpdateStatusInput } from "../schemas";
import type { Transaction, TrustScore } from "../types";

export class TransactionService {
  private static readonly base = "/transactions";

  static async create(data: CreateTransactionInput): Promise<Transaction> {
    return apiPost<Transaction>(this.base, data);
  }

  static async getByUser(
    userID: string,
    params?: PageParams,
  ): Promise<PaginatedResponse<Transaction>> {
    return apiGet<PaginatedResponse<Transaction>>(
      `${this.base}/user/${userID}`,
      params as Record<string, unknown>,
    );
  }

  static async getTrustScore(userID: string): Promise<TrustScore> {
    return apiGet<TrustScore>(`${this.base}/user/${userID}/trust-score`);
  }

  static async getFrauds(
    from: string,
    to: string,
    params?: PageParams,
  ): Promise<PaginatedResponse<Transaction>> {
    return apiGet<PaginatedResponse<Transaction>>(`${this.base}/frauds`, {
      from,
      to,
      ...params,
    });
  }

  static async updateStatus(id: string, data: UpdateStatusInput): Promise<void> {
    return apiPatch<void>(`${this.base}/${id}/status`, data);
  }
}
