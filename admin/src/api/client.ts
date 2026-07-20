const BASE_URL = import.meta.env.VITE_API_BASE_URL;
const TOKEN_KEY = "beetrack_admin_token";

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearStoredToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

type RequestOptions = {
  method?: string;
  body?: unknown;
  query?: Record<string, string | number | undefined>;
};

function buildUrl(path: string, query?: RequestOptions["query"]): string {
  const url = new URL(BASE_URL + path);
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined) url.searchParams.set(key, String(value));
    }
  }
  return url.toString();
}

// request attaches the bearer token, JSON encodes/decodes, and throws ApiError on non-2xx.
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const token = getStoredToken();
  const headers: Record<string, string> = { Accept: "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;

  let body: string | undefined;
  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(options.body);
  }

  const res = await fetch(buildUrl(path, options.query), {
    method: options.method ?? "GET",
    headers,
    body,
  });

  if (!res.ok) {
    let code = "UNKNOWN_ERROR";
    let message = `request failed with status ${res.status}`;
    try {
      const data = await res.json();
      code = data.code ?? code;
      message = data.message ?? message;
    } catch {
      // response body wasn't JSON — fall back to the generic message above
    }
    throw new ApiError(res.status, code, message);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

// resourceUrl resolves a server-provided path (already including /api/v1) against the API host.
export function resourceUrl(path: string): string {
  return new URL(path, BASE_URL).toString();
}
