"use client";

import { useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Card, FieldError, Input, Label, TextField } from "@heroui/react";
import { z } from "zod";

const schema = z.object({
  userID: z.string().min(1, "User ID is required"),
});

type FormValues = z.infer<typeof schema>;

export default function Home() {
  const router = useRouter();
  const {
    control,
    handleSubmit,
    formState: { isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { userID: "" },
  });

  function onSubmit({ userID }: FormValues) {
    router.push(`/users/${encodeURIComponent(userID.trim())}`);
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-6">
      <Card className="w-full max-w-sm">
        <Card.Header>
          <Card.Title>Fraud Detection</Card.Title>
          <Card.Description>
            Enter a user ID to view their transaction history and trust score.
          </Card.Description>
        </Card.Header>
        <Card.Content>
          <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
            <Controller
              name="userID"
              control={control}
              render={({ field, fieldState }) => (
                <TextField
                  fullWidth
                  isInvalid={!!fieldState.error}
                  name={field.name}
                  onChange={field.onChange}
                  value={field.value}
                >
                  <Label>User ID</Label>
                  <Input placeholder="e.g. user_abc123" />
                  {fieldState.error && (
                    <FieldError>{fieldState.error.message}</FieldError>
                  )}
                </TextField>
              )}
            />
            <Button
              className="w-full"
              isDisabled={isSubmitting}
              type="submit"
              variant="primary"
            >
              View User
            </Button>
          </form>
        </Card.Content>
      </Card>
    </div>
  );
}
