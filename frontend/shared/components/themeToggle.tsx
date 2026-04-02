"use client";

import { useSyncExternalStore } from "react";
import { useTheme } from "next-themes";
import { Button } from "@heroui/react";
import { MoonIcon, Sun } from "lucide-react";

const mounted = {
  subscribe: () => () => {},
  getSnapshot: () => true,
  getServerSnapshot: () => false,
};

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  const isMounted = useSyncExternalStore(
    mounted.subscribe,
    mounted.getSnapshot,
    mounted.getServerSnapshot,
  );

  if (!isMounted) return <Button variant="outline" aria-label="Toggle theme" />;

  return (
    <Button
      variant="outline"
      onPress={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
      aria-label="Toggle theme"
    >
      {resolvedTheme === "dark" ? <Sun aria-hidden /> : <MoonIcon aria-hidden />}
    </Button>
  );
}
