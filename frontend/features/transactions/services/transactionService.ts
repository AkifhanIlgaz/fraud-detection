import { apiGet } from "@/shared/lib/http";
import type { PageParams, PaginatedResponse } from "@/shared/types/api";
import type { Transaction, TrustScore } from "../types";

export class TransactionService {
  private static readonly base = "/transactions";

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
}
