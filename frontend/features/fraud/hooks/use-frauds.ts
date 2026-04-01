import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

import { useFraudTransactions } from "@/features/transactions/hooks/use-transactions";
import type { PageParams } from "@/shared/types/api";
import { fraudsQuerySchema, type FraudsQueryInput } from "../schemas";

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
