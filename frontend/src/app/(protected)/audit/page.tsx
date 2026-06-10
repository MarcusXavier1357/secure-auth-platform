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
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Auditoria</h2>

      <div className="overflow-hidden rounded-2xl border border-slate-800">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-900 text-slate-400">
            <tr>
              <th className="px-4 py-3 font-medium">Quando</th>
              <th className="px-4 py-3 font-medium">Ação</th>
              <th className="px-4 py-3 font-medium">Entidade</th>
              <th className="px-4 py-3 font-medium">Usuário</th>
              <th className="px-4 py-3 font-medium">Detalhes</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800 bg-slate-900/50">
            {logs?.map((log) => (
              <tr key={log.id}>
                <td className="whitespace-nowrap px-4 py-3 text-slate-400">
                  {new Date(log.createdAt).toLocaleString("pt-BR")}
                </td>
                <td className="px-4 py-3">
                  <span className="rounded-full bg-slate-800 px-2 py-0.5 font-mono text-xs">
                    {log.action}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-400">
                  {log.entity}
                  {log.entityId != null && ` #${log.entityId}`}
                </td>
                <td className="px-4 py-3 text-slate-400">{log.userId ?? "—"}</td>
                <td className="max-w-md truncate px-4 py-3 font-mono text-xs text-slate-500">
                  {log.newData ? JSON.stringify(log.newData) : "—"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center gap-3">
        <button
          disabled={page === 0}
          onClick={() => setPage((p) => p - 1)}
          className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm text-slate-300 transition hover:bg-slate-800 disabled:opacity-40"
        >
          Anterior
        </button>
        <span className="text-sm text-slate-400">Página {page + 1}</span>
        <button
          disabled={(logs?.length ?? 0) < PAGE_SIZE}
          onClick={() => setPage((p) => p + 1)}
          className="rounded-lg border border-slate-700 px-3 py-1.5 text-sm text-slate-300 transition hover:bg-slate-800 disabled:opacity-40"
        >
          Próxima
        </button>
      </div>
    </div>
  );
}
