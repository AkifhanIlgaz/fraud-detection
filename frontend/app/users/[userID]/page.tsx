import { UserDetailView } from "@/features/transactions/components/userDetailView";

interface Props {
  params: Promise<{ userID: string }>;
}

export default async function UserPage({ params }: Props) {
  const { userID } = await params;
  return <UserDetailView userID={decodeURIComponent(userID)} />;
}
