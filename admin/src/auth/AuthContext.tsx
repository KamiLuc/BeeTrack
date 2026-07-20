import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import { ApiError, clearStoredToken, getStoredToken } from "../api/client";
import { getMe, login as apiLogin, logout as apiLogout, type CurrentUser } from "../api/auth";
import { useI18n } from "../i18n/I18nContext";

type AuthState = {
  user: CurrentUser | null;
  loading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthState | null>(null);

// AuthProvider confirms role === 'admin' via /users/me before considering the user signed in.
export function AuthProvider({ children }: { children: ReactNode }) {
  const { t } = useI18n();
  const [user, setUser] = useState<CurrentUser | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadCurrentUser = useCallback(async () => {
    if (!getStoredToken()) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      const me = await getMe();
      setUser(me);
    } catch {
      clearStoredToken();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadCurrentUser();
  }, [loadCurrentUser]);

  const login = useCallback(async (email: string, password: string) => {
    setError(null);
    try {
      await apiLogin(email, password);
      const me = await getMe();
      if (me.role !== "admin") {
        clearStoredToken();
        setError(t("login.notAdmin"));
        return;
      }
      setUser(me);
    } catch (err) {
      clearStoredToken();
      setError(err instanceof ApiError ? err.message : t("login.failed"));
    }
  }, [t]);

  const logout = useCallback(() => {
    apiLogout();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, error, login, logout }}>{children}</AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within an AuthProvider");
  return ctx;
}
