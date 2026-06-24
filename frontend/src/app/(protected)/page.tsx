import { Link } from "react-router-dom";
import { useAuth } from "../../providers/AuthProvider";
import { Can } from "../../components/Can";
import { adminRoutes } from "../../router/admin-routes";

export default function DashboardPage() {
  const { user, permissions } = useAuth();

  return (
    <div className="space-y-8 animate-fade-in font-sans">
      {/* Welcome Banner */}
      <section className="relative overflow-hidden rounded-3xl border border-app-border bg-app-card p-8 shadow-xl backdrop-blur-md">
        <div className="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-indigo-500/10 blur-3xl pointer-events-none" />
        <div className="relative z-10 space-y-2">
          <span className="inline-flex rounded-lg bg-indigo-500/10 border border-indigo-500/20 px-2.5 py-0.5 text-xs font-semibold uppercase tracking-wider text-indigo-400">
            Painel Geral
          </span>
          <h2 className="text-2xl font-extrabold tracking-tight text-app-text sm:text-3xl">
            Bem-vindo, {user?.name}
          </h2>
          <p className="text-sm text-app-muted max-w-xl leading-relaxed">Sessão ativa.</p>
        </div>
      </section>

      {/* Permissions List */}
      <section className="rounded-3xl border border-app-border bg-app-card/60 p-8 backdrop-blur-md space-y-4">
        <div className="flex items-center gap-2.5">
          <div className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
          <h2 className="text-sm font-semibold uppercase tracking-wider text-app-text/90">
            Suas permissões ativas
          </h2>
        </div>

        {permissions.length === 0 ? (
          <p className="text-sm text-app-muted">
            Nenhuma permissão especial atribuída ao seu perfil.
          </p>
        ) : (
          <ul className="flex flex-wrap gap-2.5">
            {permissions.map((code) => {
              const permObj = user?.permissions?.find(
                (p: { code: string; description: string }) => p.code === code,
              );
              const displayLabel = permObj?.description || code;
              return (
                <li
                  key={code}
                  className="rounded-xl border border-app-border bg-app-bg/60 px-3.5 py-1.5 text-xs font-semibold text-app-text shadow-inner"
                >
                  {displayLabel}
                </li>
              );
            })}
          </ul>
        )}
      </section>

      {/* Navigation Grid */}
      <section className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {adminRoutes.map((card) => (
          <Can key={card.path} permission={card.permission}>
            <Link
              to={card.path}
              className="group relative block rounded-3xl border border-app-border bg-app-card/30 p-6 shadow-md transition-all duration-300 hover:-translate-y-1 hover:border-app-border/80 hover:bg-app-card-hover"
            >
              <div className="absolute inset-0 rounded-3xl bg-gradient-to-br from-indigo-500/0 to-indigo-500/5 opacity-0 transition-opacity duration-300 group-hover:opacity-100" />

              <div className="relative z-10 space-y-3">
                <div className="inline-flex h-10 w-10 items-center justify-center rounded-xl bg-app-bg border border-app-border text-indigo-400 group-hover:text-indigo-300 transition-colors">
                  {card.navLabel === "Usuários" && "👥"}
                  {card.navLabel === "Permissões" && "🔑"}
                  {card.navLabel === "Auditoria" && "📑"}
                </div>
                <div>
                  <h3 className="font-bold text-app-text group-hover:text-indigo-400 transition-colors">
                    {card.cardTitle}
                  </h3>
                  <p className="mt-2 text-xs text-app-muted leading-relaxed">
                    {card.cardDescription}
                  </p>
                </div>
              </div>
            </Link>
          </Can>
        ))}
      </section>
    </div>
  );
}
