import { Button } from "@heroui/react";
import { Card } from "@heroui/react";

export default function Home() {
  return (
    <div className="flex flex-col flex-1 items-center justify-center min-h-screen bg-background p-8">
      <Card className="max-w-md w-full">
        <Card.Header>
          <Card.Title>Fraud Detection Dashboard</Card.Title>
          <Card.Description>
            HeroUI v3 + Next.js App Router ile hazır.
          </Card.Description>
        </Card.Header>
        <Card.Content className="flex gap-3">
          <Button variant="primary">Başla</Button>
          <Button variant="outline">Daha Fazla</Button>
        </Card.Content>
      </Card>
    </div>
  );
}
