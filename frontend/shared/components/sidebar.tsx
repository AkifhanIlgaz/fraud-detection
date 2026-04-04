"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Activity, ChevronLeft, ChevronRight, ShieldAlert, Users, X } from "lucide-react";

import { useSidebar } from "@/shared/providers/sidebarProvider";
import { ThemeToggle } from "./themeToggle";

const NAV_ITEMS = [
  { href: "/", label: "Users", icon: Users },
  { href: "/frauds", label: "Fraud Transactions", icon: ShieldAlert },
  { href: "/live", label: "Live Feed", icon: Activity },
] as const;

function NavItems({ collapsed, onNavigate }: { collapsed: boolean; onNavigate?: () => void }) {
  const pathname = usePathname();

  function isActive(href: string) {
    if (href === "/") return pathname === "/" || pathname.startsWith("/users");
    return pathname.startsWith(href);
  }

  return (
    <nav className="flex flex-1 flex-col gap-1 px-2 py-3">
      {NAV_ITEMS.map(({ href, label, icon: Icon }) => {
        const active = isActive(href);
        return (
          <Link
            key={href}
            href={href}
            title={collapsed ? label : undefined}
            onClick={onNavigate}
            className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
              active
                ? "bg-accent/10 font-medium text-accent"
                : "text-muted hover:bg-default hover:text-foreground"
            } ${collapsed ? "justify-center" : ""}`}
          >
            <Icon size={15} aria-hidden />
            {!collapsed && <span>{label}</span>}
          </Link>
        );
      })}
    </nav>
  );
}

// ── Desktop sidebar ────────────────────────────────────────────────────────

function DesktopSidebar() {
  const { collapsed, toggle } = useSidebar();

  return (
    <aside
      className={`sticky top-0 hidden h-screen shrink-0 flex-col border-r border-border bg-surface transition-[width] duration-200 md:flex ${
        collapsed ? "w-14" : "w-56"
      }`}
    >
      <div className="flex h-14 items-center justify-between border-b border-border px-3">
        {!collapsed && (
          <span className="truncate text-sm font-semibold">Fraud Detection</span>
        )}
        <button
          onClick={toggle}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted transition-colors hover:bg-default hover:text-foreground ${
            collapsed ? "mx-auto" : ""
          }`}
        >
          {collapsed ? <ChevronRight size={14} aria-hidden /> : <ChevronLeft size={14} aria-hidden />}
        </button>
      </div>

      <NavItems collapsed={collapsed} />

      <div className={`border-t border-border p-3 ${collapsed ? "flex justify-center" : ""}`}>
        <ThemeToggle />
      </div>
    </aside>
  );
}

// ── Mobile sidebar (drawer) ────────────────────────────────────────────────

function MobileSidebar() {
  const { mobileOpen, closeMobile } = useSidebar();

  return (
    <>
      {/* Backdrop */}
      <div
        aria-hidden
        className={`fixed inset-0 z-40 bg-black/50 transition-opacity duration-200 md:hidden ${
          mobileOpen ? "opacity-100" : "pointer-events-none opacity-0"
        }`}
        onClick={closeMobile}
      />

      {/* Drawer */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-border bg-surface transition-transform duration-200 md:hidden ${
          mobileOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex h-14 items-center justify-between border-b border-border px-4">
          <span className="text-sm font-semibold">Fraud Detection</span>
          <button
            onClick={closeMobile}
            aria-label="Close menu"
            className="flex h-7 w-7 items-center justify-center rounded-md text-muted transition-colors hover:bg-default hover:text-foreground"
          >
            <X size={15} aria-hidden />
          </button>
        </div>

        <NavItems collapsed={false} onNavigate={closeMobile} />

        <div className="border-t border-border p-3">
          <ThemeToggle />
        </div>
      </aside>
    </>
  );
}

// ── Export ─────────────────────────────────────────────────────────────────

export function Sidebar() {
  return (
    <>
      <DesktopSidebar />
      <MobileSidebar />
    </>
  );
}
