"use client";

import { usePathname } from "next/navigation";
import { Breadcrumbs } from "@heroui/react";

interface Crumb {
  label: string;
  href?: string;
}

interface PageMeta {
  description: string;
  breadcrumbs: Crumb[];
}

function getPageMeta(pathname: string): PageMeta | null {
  if (pathname === "/") {
    return {
      description: "Search for a user to view their transaction history and trust score.",
      breadcrumbs: [{ label: "Users" }],
    };
  }

  if (pathname.startsWith("/users/")) {
    const userID = decodeURIComponent(pathname.slice("/users/".length));
    return {
      description: `Viewing transactions and trust score for ${userID}`,
      breadcrumbs: [
        { label: "Users", href: "/" },
        { label: userID },
      ],
    };
  }

  if (pathname === "/frauds") {
    return {
      description: "Browse and filter fraud activity across all users by date range.",
      breadcrumbs: [{ label: "Fraud Transactions" }],
    };
  }

  return null;
}

export function PageHeader() {
  const pathname = usePathname();
  const meta = getPageMeta(pathname);

  if (!meta) return null;

  return (
    <header className="border-b border-border px-6 py-4">
      <Breadcrumbs className="mb-1 text-xs">
        {meta.breadcrumbs.map((crumb) => (
          <Breadcrumbs.Item key={crumb.label} href={crumb.href}>
            {crumb.label}
          </Breadcrumbs.Item>
        ))}
      </Breadcrumbs>
      <p className="text-sm text-muted">{meta.description}</p>
    </header>
  );
}
