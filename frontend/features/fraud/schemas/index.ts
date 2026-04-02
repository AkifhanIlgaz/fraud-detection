import { z } from "zod";

export const fraudsQuerySchema = z.object({
  from: z
    .string()
    .min(1, "Start date is required")
    .regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/, "Must be RFC3339"),
  to: z
    .string()
    .min(1, "End date is required")
    .regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/, "Must be RFC3339"),
  page: z.number().int().positive().optional(),
  limit: z.number().int().positive().max(100).optional(),
});

export type FraudsQueryInput = z.infer<typeof fraudsQuerySchema>;

const dateStr = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Invalid date");

export const dateRangeSchema = z
  .object({ start: dateStr, end: dateStr })
  .refine((d) => d.start <= d.end, {
    message: "Start date must be before end date",
    path: ["end"],
  });

export type DateRangeValues = z.infer<typeof dateRangeSchema>;
