import { useAuth } from "../providers/AuthProvider";

// Controle visual apenas — a autorização real é sempre validada no backend.
export function usePermission(code: string): boolean {
  const { hasPermission } = useAuth();
  return hasPermission(code);
}
