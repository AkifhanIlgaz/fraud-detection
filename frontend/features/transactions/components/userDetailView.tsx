"use client";

import { useRouter } from "next/navigation";

import { ThemeToggle } from "@/shared/components/themeToggle";
import { Button, Card } from "@heroui/react";

import { ChevronLeft } from "lucide-react";
import { useTrustScore, useUserTransactions } from "../hooks/useTransactions";
import { TransactionTable } from "./transactionTable";
import { TrustScoreCard } from "./trustScoreCard";

function TrustScoreSkeleton() {
  return (
    <Card>
      <Card.Header>
        <Card.Title>Trust Score</Card.Title>
      </Card.Header>
      <Card.Content className="flex flex-col gap-4">
        <div className="h-14 animate-pulse rounded-lg bg-border" />
        <div className="h-2 animate-pulse rounded-full bg-border" />
        <div className="grid grid-cols-3 gap-3 border-t border-border pt-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-10 animate-pulse rounded-lg bg-border" />
          ))}
        </div>
      </Card.Content>
    </Card>
  );
}

export function UserDetailView({ userID }: { userID: string }) {
  const router = useRouter();
  const {
    data: trustScore,
    isLoading: scoreLoading,
    isError: scoreError,
  } = useTrustScore(userID);
  const { data: txData } = useUserTransactions(userID, { page: 1, limit: 1 });

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6 px-4 py-8">
      <div className="flex items-center gap-4">
        <Button variant="ghost" onPress={() => router.push("/")}>
          <ChevronLeft />
          Back
        </Button>
        <div className="flex-1">
          <h1 className="text-xl font-semibold">User Detail</h1>
          <p className="font-mono text-sm text-muted">{userID}</p>
        </div>
        <ThemeToggle />
      </div>

      {scoreLoading ? (
        <TrustScoreSkeleton />
      ) : scoreError || !trustScore ? (
        <Card>
          <Card.Content>
            <p className="text-sm text-muted">
              Could not load trust score for this user.
            </p>
          </Card.Content>
        </Card>
      ) : (
        <TrustScoreCard data={trustScore} />
      )}

      <div className="flex flex-col gap-3">
        <h2 className="text-base font-semibold">Transaction History</h2>
        <TransactionTable userID={userID} />
      </div>
    </div>
  );
}
