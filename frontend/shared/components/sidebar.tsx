"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronLeft, ChevronRight, ShieldAlert, Users } from "lucide-react";

import { useSidebar } from "@/shared/providers/sidebarProvider";
import { ThemeToggle } from "./themeToggle";

const NAV_ITEMS = [
  { href: "/", label: "Users", icon: Users },
  { href: "/frauds", label: "Fraud Transactions", icon: ShieldAlert },
] as const;

export function Sidebar() {
  const pathname = usePathname();
  const { collapsed, toggle } = useSidebar();

  function isActive(href: string) {
    if (href === "/") return pathname === "/" || pathname.startsWith("/users");
    return pathname.startsWith(href);
  }

  return (
    <aside
      className={`sticky top-0 flex h-screen shrink-0 flex-col border-r border-border bg-surface transition-[width] duration-200 ${
        collapsed ? "w-14" : "w-56"
      }`}
    >
      {/* Logo + collapse toggle */}
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

      {/* Nav */}
      <nav className="flex flex-1 flex-col gap-1 px-2 py-3">
        {NAV_ITEMS.map(({ href, label, icon: Icon }) => {
          const active = isActive(href);
          return (
            <Link
              key={href}
              href={href}
              title={collapsed ? label : undefined}
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

      {/* Footer */}
      <div
        className={`border-t border-border p-3 ${collapsed ? "flex justify-center" : ""}`}
      >
        <ThemeToggle />
      </div>
    </aside>
  );
}
