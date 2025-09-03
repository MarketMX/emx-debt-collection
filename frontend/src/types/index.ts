export interface User {
  id: string;
  email: string;
  name?: string;
}

export interface Upload {
  id: string;
  user_id: string;
  filename: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  created_at: string;
  processed_count?: number;
  total_count?: number;
}

export interface Account {
  id: string;
  upload_id: string;
  account_code: string;
  customer_name: string;
  contact_person?: string;
  telephone: string;
  amount_current: number;
  amount_30d: number;
  amount_60d: number;
  amount_90d: number;
  amount_120d: number;
  total_balance: number;
  selected?: boolean;
}

export interface MessageLog {
  id: string;
  account_id: string;
  status: 'sent' | 'failed';
  sent_at: string;
  response_from_service?: string;
  account?: Account;
}

export interface MessageJob {
  id: string;
  user_id: string;
  total_accounts: number;
  successful_sends: number;
  failed_sends: number;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  created_at: string;
  completed_at?: string;
}

export interface AuthConfig {
  auth_url: string;
  realm: string;
  client_id: string;
}