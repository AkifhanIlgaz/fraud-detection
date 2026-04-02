import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useForm } from "react-hook-form";

import type { PageParams } from "@/shared/types/api";
import { fraudsQuerySchema, type FraudsQueryInput } from "../schemas";
import { FraudService } from "../services/fraudService";

export const fraudKeys = {
  all: ["frauds"] as const,
  list: (from: string, to: string, params?: PageParams) =>
    [...fraudKeys.all, from, to, params] as const,
};

export function useFraudTransactions(from: string, to: string, params?: PageParams) {
  return useQuery({
    queryKey: fraudKeys.list(from, to, params),
    queryFn: () => FraudService.getFrauds(from, to, params),
    enabled: !!from && !!to,
  });
}

export function useFraudsFilter(params?: PageParams) {
  const form = useForm<FraudsQueryInput>({
    resolver: zodResolver(fraudsQuerySchema),
    defaultValues: { from: "", to: "" },
  });

  const { from, to } = form.watch();
  const isValid = form.formState.isValid && !!from && !!to;

  const query = useFraudTransactions(
    isValid ? from : "",
    isValid ? to : "",
    params,
  );

  return { form, query };
}
