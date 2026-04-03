export type ApiEnvelope<T> = {
  data: T;
  meta?: Record<string, unknown>;
};

export type AuthSession = {
  token: string;
  expiresAt: string;
  user: User;
  tenant: TenantSummary;
};

export type User = {
  id: string;
  email: string;
  fullName: string;
  role: string;
  credits: number;
};

export type TenantSummary = {
  id: string;
  slug: string;
  name: string;
  planCode: string;
};

export type PromptSection = {
  key: string;
  title: string;
  content: string;
};

export type GenerationJob = {
  id: string;
  status: "queued" | "processing" | "done" | "failed";
  progress: number;
  costCredits: number;
  resultUrl?: string;
  createdAt: string;
  updatedAt: string;
  promptSections: PromptSection[];
};

export type WalletSummary = {
  balance: number;
  pending: number;
  lifetimeSpent: number;
};

export type Plan = {
  id: string;
  code: string;
  name: string;
  creditAmount: number;
  priceCents: number;
  currency: string;
  active: boolean;
};

export type GalleryItem = {
  id: string;
  title: string;
  previewUrl: string;
  favorite: boolean;
  createdAt: string;
};

