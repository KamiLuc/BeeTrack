import { clearStoredToken, request, setStoredToken } from "./client";

export type CurrentUser = {
  id: number;
  email: string;
  name: string;
  role: string;
  verified: boolean;
};

type LoginResponse = {
  access_token: string;
  refresh_token: string;
  name: string;
};

export async function login(email: string, password: string): Promise<void> {
  const res = await request<LoginResponse>("/auth/login", {
    method: "POST",
    body: { email, password },
  });
  setStoredToken(res.access_token);
}

export function logout(): void {
  clearStoredToken();
}

export function getMe(): Promise<CurrentUser> {
  return request<CurrentUser>("/users/me");
}
