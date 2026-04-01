export interface Transaction {
  id: string;
  user_id: string;
  amount: number;
  lat: number;
  lon: number;
  status: "pending" | "approved" | "flagged" | "rejected";
  created_at: string;
  fraud_reasons?: string[];
}

export interface TrustScore {
  user_id: string;
  score: number;
  risk_level: "low" | "medium" | "high";
  total: number;
  fraud_count: number;
}
