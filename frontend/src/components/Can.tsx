import type { ReactNode } from "react";
import { usePermission } from "../hooks/usePermission";

// Renderiza children apenas se o usuário possui a permissão.
// Controle visual — o backend revalida toda autorização.
export function Can({ permission, children }: { permission: string; children: ReactNode }) {
  const allowed = usePermission(permission);
  return allowed ? children : null;
}
