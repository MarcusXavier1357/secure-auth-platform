import { paths } from "./paths";

export const adminRoutes = [
  {
    id: "users",
    path: paths.users(),
    permission: "users.manage",
    navLabel: "Usuários",
    cardTitle: "Usuários",
    cardDescription: "Gerenciar contas e acessos.",
  },
  {
    id: "permissions",
    path: paths.permissions(),
    permission: "permissions.manage",
    navLabel: "Permissões",
    cardTitle: "Permissões",
    cardDescription: "Conceder e revogar permissões.",
  },
  {
    id: "audit",
    path: paths.audit(),
    permission: "audit_logs.read",
    navLabel: "Auditoria",
    cardTitle: "Auditoria",
    cardDescription: "Histórico de ações críticas.",
  },
] as const;

export type AdminRouteId = (typeof adminRoutes)[number]["id"];
