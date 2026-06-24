import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../../../services/api";

const PAGE_SIZE = 25;

export default function AuditPage() {
  const [page, setPage] = useState(0);

  const {
    data: logs,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["audit-logs", page],
    queryFn: () => api.audit.list(PAGE_SIZE, page * PAGE_SIZE),
  });

  if (isLoading) {
    return <p className="text-sm text-slate-400">Carregando auditoria...</p>;
  }
  if (isError) {
    return <p className="text-sm text-red-400">Erro ao carregar logs de auditoria.</p>;
  }

  return (
    <div className="space-y-8 animate-fade-in font-sans">
      <div>
        <h2 className="text-2xl font-extrabold tracking-tight text-app-text">Auditoria</h2>
        <p className="text-sm text-app-muted">
          Rastreabilidade completa de todas as alterações críticas realizadas no sistema.
        </p>
      </div>

      <div className="overflow-hidden rounded-2xl border border-app-border bg-app-card backdrop-blur-md shadow-xl">
        <table className="w-full text-left text-sm border-collapse">
          <thead>
            <tr className="border-b border-app-border bg-app-bg/40 text-app-muted">
              <th className="px-6 py-4 font-semibold">Quando</th>
              <th className="px-6 py-4 font-semibold">Ação</th>
              <th className="px-6 py-4 font-semibold">Entidade</th>
              <th className="px-6 py-4 font-semibold">Usuário</th>
              <th className="px-6 py-4 font-semibold">Detalhes (JSON)</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-app-border/60">
            {logs?.map((log) => {
              // Color helper for action tags
              let actionColor = "bg-app-bg text-app-text border border-app-border";
              if (log.action.includes("login"))
                actionColor =
                  "bg-indigo-500/10 border border-indigo-500/20 text-indigo-650 dark:text-indigo-300";
              else if (log.action.includes("create") || log.action.includes("grant"))
                actionColor =
                  "bg-emerald-500/10 border border-emerald-500/20 text-emerald-650 dark:text-emerald-400";
              else if (
                log.action.includes("delete") ||
                log.action.includes("revoke") ||
                log.action.includes("deactivate")
              )
                actionColor =
                  "bg-red-500/10 border border-red-500/20 text-red-600 dark:text-red-400";

              return (
                <tr key={log.id} className="transition-colors hover:bg-app-card-hover/40">
                  <td className="whitespace-nowrap px-6 py-4 text-app-muted">
                    {new Date(log.createdAt).toLocaleString("pt-BR")}
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`inline-flex rounded-lg px-2.5 py-0.5 font-mono text-[10px] font-semibold uppercase tracking-wider ${actionColor}`}
                    >
                      {log.action}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-app-text">
                    <span className="font-semibold text-app-text">{log.entity}</span>
                    {log.entityId != null && (
                      <span className="ml-1.5 text-xs text-app-muted font-mono">
                        #{log.entityId}
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 text-app-muted font-mono text-xs">
                    {log.userId ?? "—"}
                  </td>
                  <td className="max-w-md truncate px-6 py-4 font-mono text-xs text-app-muted hover:text-app-text transition-colors">
                    {log.newData ? JSON.stringify(log.newData) : "—"}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className="flex items-center gap-3">
        <button
          disabled={page === 0}
          onClick={() => setPage((p) => p - 1)}
          className="rounded-xl border border-app-border bg-app-bg/40 px-4 py-2 text-sm font-semibold text-app-text transition-all hover:bg-app-card-hover disabled:opacity-30 disabled:cursor-not-allowed"
        >
          Anterior
        </button>
        <span className="text-sm font-medium text-app-muted">Página {page + 1}</span>
        <button
          disabled={(logs?.length ?? 0) < PAGE_SIZE}
          onClick={() => setPage((p) => p + 1)}
          className="rounded-xl border border-app-border bg-app-bg/40 px-4 py-2 text-sm font-semibold text-app-text transition-all hover:bg-app-card-hover disabled:opacity-30 disabled:cursor-not-allowed"
        >
          Próxima
        </button>
      </div>
    </div>
  );
}
