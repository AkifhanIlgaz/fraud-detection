import { UserDetailView } from "@/features/transactions/components/user-detail-view";

interface Props {
  params: Promise<{ userID: string }>;
}

export default async function UserPage({ params }: Props) {
  const { userID } = await params;
  return <UserDetailView userID={decodeURIComponent(userID)} />;
}
