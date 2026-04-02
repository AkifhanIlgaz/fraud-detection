import type { Metadata } from "next";
import { Outfit } from "next/font/google";
import "./globals.css";

import { PageHeader } from "@/shared/components/pageHeader";
import { Sidebar } from "@/shared/components/sidebar";
import { Providers } from "./providers";

const outfit = Outfit({
  variable: "--font-outfit",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Fraud Detection",
  description: "Fraud Detection Dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${outfit.className} antialiased`}
      suppressHydrationWarning
    >
      <body className="flex min-h-screen bg-background">
        <Providers>
          <Sidebar />
          <main className="flex flex-1 flex-col">
            <PageHeader />
            {children}
          </main>
        </Providers>
      </body>
    </html>
  );
}
