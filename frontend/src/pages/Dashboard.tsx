import { useNavigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";
import { Can } from "../components/Can";

export function Dashboard() {
  const { user, permissions, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate("/login", { replace: true });
  }

  return (
    <div className="min-h-screen bg-slate-950 text-white">
      <header className="border-b border-slate-800 bg-slate-900">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <h1 className="text-lg font-semibold">Auth System</h1>
          <div className="flex items-center gap-4">
            <span className="text-sm text-slate-400">
              {user?.name}
              {user?.role && (
                <span className="ml-2 rounded-full bg-indigo-950 px-2 py-0.5 text-xs text-indigo-300">
                  {user.role.name}
                </span>
              )}
            </span>
            <button
              onClick={handleLogout}
              className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm text-slate-300 transition hover:bg-slate-800"
            >
              Sair
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-5xl space-y-6 px-6 py-8">
        <section className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
          <h2 className="mb-1 text-base font-semibold">Bem-vindo, {user?.name}</h2>
          <p className="text-sm text-slate-400">
            Sessão ativa com renovação automática. O access token vive apenas em memória.
          </p>
        </section>

        <section className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
          <h2 className="mb-3 text-base font-semibold">Suas permissões</h2>
          {permissions.length === 0 ? (
            <p className="text-sm text-slate-400">Nenhuma permissão atribuída.</p>
          ) : (
            <ul className="flex flex-wrap gap-2">
              {permissions.map((code) => (
                <li
                  key={code}
                  className="rounded-full border border-slate-700 bg-slate-800 px-3 py-1 font-mono text-xs text-slate-300"
                >
                  {code}
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Can permission="users.manage">
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
              <h3 className="font-semibold">Usuários</h3>
              <p className="mt-1 text-sm text-slate-400">Gerenciar contas e acessos.</p>
            </div>
          </Can>
          <Can permission="permissions.manage">
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
              <h3 className="font-semibold">Permissões</h3>
              <p className="mt-1 text-sm text-slate-400">Conceder e revogar permissões.</p>
            </div>
          </Can>
          <Can permission="audit_logs.read">
            <div className="rounded-2xl border border-slate-800 bg-slate-900 p-5">
              <h3 className="font-semibold">Auditoria</h3>
              <p className="mt-1 text-sm text-slate-400">Histórico de ações críticas.</p>
            </div>
          </Can>
        </section>
      </main>
    </div>
  );
}
