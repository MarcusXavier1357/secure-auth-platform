import { Link } from "react-router-dom";
import { useAuth } from "../../providers/AuthProvider";
import { Can } from "../../components/Can";
import { adminRoutes } from "../../router/admin-routes";

export default function DashboardPage() {
  const { user, permissions } = useAuth();

  return (
    <div className="space-y-6">
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
        {adminRoutes.map((card) => (
          <Can key={card.path} permission={card.permission}>
            <Link
              to={card.path}
              className="block rounded-2xl border border-slate-800 bg-slate-900 p-5 transition hover:border-slate-600 hover:bg-slate-800/60"
            >
              <h3 className="font-semibold">{card.cardTitle}</h3>
              <p className="mt-1 text-sm text-slate-400">{card.cardDescription}</p>
            </Link>
          </Can>
        ))}
      </section>
    </div>
  );
}
