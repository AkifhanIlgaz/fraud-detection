import type { Transaction } from "@/features/transactions/types";
import { apiGet } from "@/shared/lib/http";
import type { PageParams, PaginatedResponse } from "@/shared/types/api";

export class FraudService {
  private static readonly base = "/transactions";

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
}
