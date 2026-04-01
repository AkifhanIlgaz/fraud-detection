"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Toast } from "@heroui/react";

import { queryClient } from "@/shared/lib/query-client";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toast.Provider placement="bottom" />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}
