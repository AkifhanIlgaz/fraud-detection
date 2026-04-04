"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Toast } from "@heroui/react";

import { queryClient } from "@/shared/lib/queryClient";
import { TransactionStreamListener } from "@/shared/components/transactionStreamListener";
import { SidebarProvider } from "@/shared/providers/sidebarProvider";
import { ThemeProvider } from "next-themes";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <SidebarProvider>
        <QueryClientProvider client={queryClient}>
          {children}
          <TransactionStreamListener />
          <Toast.Provider placement="bottom" />
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      </SidebarProvider>
    </ThemeProvider>
  );
}
