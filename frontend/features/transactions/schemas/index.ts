import { z } from "zod";

export const createTransactionSchema = z.object({
  user_id: z.string().min(1, "User ID is required"),
  amount: z.number({ error: "Amount must be a number" }).positive("Amount must be greater than 0"),
  lat: z.number().min(-90, "Lat must be ≥ -90").max(90, "Lat must be ≤ 90"),
  lon: z.number().min(-180, "Lon must be ≥ -180").max(180, "Lon must be ≤ 180"),
});

export const updateStatusSchema = z.object({
  status: z.enum(["pending", "approved", "suspicious", "fraud"], "Invalid status"),
});

export const fraudsFilterSchema = z.object({
  from: z.string().regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/, "Must be RFC3339 (e.g. 2024-01-01T00:00:00Z)"),
  to: z.string().regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/, "Must be RFC3339 (e.g. 2024-01-01T00:00:00Z)"),
});

export type CreateTransactionInput = z.infer<typeof createTransactionSchema>;
export type UpdateStatusInput = z.infer<typeof updateStatusSchema>;
export type FraudsFilterInput = z.infer<typeof fraudsFilterSchema>;
