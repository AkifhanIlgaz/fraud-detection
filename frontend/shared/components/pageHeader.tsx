"use client";

import { usePathname } from "next/navigation";
import { Breadcrumbs } from "@heroui/react";
import { Menu } from "lucide-react";

import { useSidebar } from "@/shared/providers/sidebarProvider";

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
  const { openMobile } = useSidebar();
  const meta = getPageMeta(pathname);

  if (!meta) return null;

  return (
    <header className="flex items-center gap-3 border-b border-border px-4 py-4 md:px-6">
      {/* Hamburger — mobile only */}
      <button
        onClick={openMobile}
        aria-label="Open menu"
        className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted transition-colors hover:bg-default hover:text-foreground md:hidden"
      >
        <Menu size={18} aria-hidden />
      </button>

      <div>
        <Breadcrumbs className="mb-0.5 text-xs">
          {meta.breadcrumbs.map((crumb) => (
            <Breadcrumbs.Item key={crumb.label} href={crumb.href}>
              {crumb.label}
            </Breadcrumbs.Item>
          ))}
        </Breadcrumbs>
        <p className="text-sm text-muted">{meta.description}</p>
      </div>
    </header>
  );
}
