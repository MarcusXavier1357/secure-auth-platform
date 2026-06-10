import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";

export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-400">
        Carregando...
      </div>
    );
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  return children;
}
