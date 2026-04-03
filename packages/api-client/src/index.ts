import type {
  ApiEnvelope,
  AuthSession,
  GalleryItem,
  GenerationJob,
  Plan,
  WalletSummary
} from "@foto-magica/types";

type RequestOptions = RequestInit & {
  token?: string;
};

export class ApiClient {
  constructor(private readonly baseUrl: string) {}

  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    const headers = new Headers(options.headers ?? {});
    headers.set("Content-Type", "application/json");
    if (options.token) {
      headers.set("Authorization", `Bearer ${options.token}`);
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers
    });

    if (!response.ok) {
      const payload = await response.json().catch(() => ({}));
      const message = payload.error?.message ?? `Request failed with ${response.status}`;
      throw new Error(message);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    const payload = (await response.json()) as ApiEnvelope<T>;
    return payload.data;
  }

  login(payload: { email: string; password: string; tenantSlug?: string }) {
    return this.request<AuthSession>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  register(payload: {
    companyName: string;
    companySlug: string;
    fullName: string;
    email: string;
    password: string;
  }) {
    return this.request<AuthSession>("/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  me(token: string) {
    return this.request<AuthSession>("/v1/profile/me", { token });
  }

  plans(token: string) {
    return this.request<Plan[]>("/v1/billing/plans", { token });
  }

  wallet(token: string) {
    return this.request<WalletSummary>("/v1/credits/wallet", { token });
  }

  listJobs(token: string) {
    return this.request<GenerationJob[]>("/v1/generation/jobs", { token });
  }

  createGeneration(token: string, payload: Record<string, unknown>) {
    return this.request<GenerationJob>("/v1/generation/jobs", {
      method: "POST",
      token,
      body: JSON.stringify(payload)
    });
  }

  gallery(token: string) {
    return this.request<GalleryItem[]>("/v1/gallery", { token });
  }
}

export const createApiClient = () =>
  new ApiClient(import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080");

