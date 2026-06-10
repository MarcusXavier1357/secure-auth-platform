import type { ReactNode } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";
import { Can } from "./Can";
import { paths } from "../router/paths";

const navItems = [
  { to: paths.home(), label: "Início", permission: null },
  { to: paths.users(), label: "Usuários", permission: "users.manage" },
  { to: paths.permissions(), label: "Permissões", permission: "permissions.manage" },
  { to: paths.audit(), label: "Auditoria", permission: "audit_logs.read" },
] as const;

function SidebarLink({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      end={to === paths.home()}
      className={({ isActive }) =>
        `block rounded-lg px-3 py-2 text-sm transition ${
          isActive
            ? "bg-indigo-600/20 font-medium text-indigo-300"
            : "text-slate-400 hover:bg-slate-800 hover:text-slate-200"
        }`
      }
    >
      {label}
    </NavLink>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate(paths.login(), { replace: true });
  }

  return (
    <div className="flex min-h-screen bg-slate-950 text-white">
      <aside className="flex w-60 shrink-0 flex-col border-r border-slate-800 bg-slate-900">
        <div className="border-b border-slate-800 px-5 py-5">
          <Link to={paths.home()} className="text-lg font-semibold tracking-tight">
            Auth System
          </Link>
        </div>

        <nav className="flex-1 space-y-1 px-3 py-4">
          {navItems.map((item) =>
            item.permission ? (
              <Can key={item.to} permission={item.permission}>
                <SidebarLink to={item.to} label={item.label} />
              </Can>
            ) : (
              <SidebarLink key={item.to} to={item.to} label={item.label} />
            ),
          )}
        </nav>

        <div className="border-t border-slate-800 px-4 py-4">
          <p className="truncate text-sm font-medium">{user?.name}</p>
          <p className="truncate text-xs text-slate-500">{user?.email}</p>
          {user?.role && (
            <span className="mt-2 inline-block rounded-full bg-indigo-950 px-2 py-0.5 text-xs text-indigo-300">
              {user.role.name}
            </span>
          )}
          <button
            onClick={handleLogout}
            className="mt-4 w-full rounded-lg border border-slate-700 px-3 py-2 text-sm text-slate-300 transition hover:bg-slate-800"
          >
            Sair
          </button>
        </div>
      </aside>

      <main className="min-w-0 flex-1 overflow-auto px-8 py-8">{children}</main>
    </div>
  );
}
