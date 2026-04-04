"use client";

import { LiveFeedTable } from "./liveFeedTable";
import { AlertPanel } from "./alertPanel";

export function LiveFeedView() {
  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8">
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-[1fr_320px]">
        <LiveFeedTable />
        <AlertPanel />
      </div>
    </div>
  );
}
