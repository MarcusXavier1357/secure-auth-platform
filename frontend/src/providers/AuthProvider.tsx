import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import { api, setAccessToken, tryRefresh, type UserWithPermissions } from "../services/api";
import { matchPermission } from "../utils/permission";

interface AuthState {
  user: UserWithPermissions | null;
  permissions: string[];
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  hasPermission: (code: string) => boolean;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserWithPermissions | null>(null);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  const loadMe = useCallback(async () => {
    const me = await api.me();
    setUser(me.user);
    const codes =
      me.permissions.length > 0
        ? me.permissions
        : (me.user.permissions?.map((p) => p.code) ?? []);
    setPermissions(codes);
  }, []);

  // Na carga da página, tenta restaurar a sessão pelo cookie de refresh.
  useEffect(() => {
    (async () => {
      try {
        if (await tryRefresh()) {
          await loadMe();
        }
      } finally {
        setLoading(false);
      }
    })();
  }, [loadMe]);

  const login = useCallback(
    async (email: string, password: string) => {
      const { accessToken } = await api.login(email, password);
      setAccessToken(accessToken);
      await loadMe();
    },
    [loadMe],
  );

  const logout = useCallback(async () => {
    try {
      await api.logout();
    } finally {
      setAccessToken(null);
      setUser(null);
      setPermissions([]);
    }
  }, []);

  const hasPermission = useCallback(
    (code: string) => matchPermission(permissions, code),
    [permissions],
  );

  return (
    <AuthContext.Provider value={{ user, permissions, loading, login, logout, hasPermission }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
