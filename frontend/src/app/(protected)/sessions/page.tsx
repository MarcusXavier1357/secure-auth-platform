import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type ActiveSession } from "../../../services/api";
import { parseApiErrorMessage } from "../../../utils/apiError";
import { useToast } from "../../../hooks/useToast";

// Helper simples para extrair Navegador e SO do User-Agent
function parseUserAgent(uaString?: string) {
  if (!uaString) return { browser: "Desconhecido", os: "Desconhecido" };

  let browser = "Outro Navegador";
  let os = "Outro SO";

  const ua = uaString.toLowerCase();

  // Detecção de SO
  if (ua.includes("windows")) os = "Windows";
  else if (ua.includes("macintosh") || ua.includes("mac os")) os = "macOS";
  else if (ua.includes("linux")) os = "Linux";
  else if (ua.includes("android")) os = "Android";
  else if (ua.includes("iphone") || ua.includes("ipad")) os = "iOS";

  // Detecção de Navegador
  if (ua.includes("edg/")) browser = "Edge";
  else if (ua.includes("chrome") && !ua.includes("chromium")) browser = "Chrome";
  else if (ua.includes("firefox")) browser = "Firefox";
  else if (ua.includes("safari") && !ua.includes("chrome")) browser = "Safari";
  else if (ua.includes("opera") || ua.includes("opr/")) browser = "Opera";

  return { browser, os };
}

// Retorna emoji representativo para o SO
function getOSEmoji(os: string) {
  switch (os) {
    case "Windows":
      return "💻";
    case "macOS":
      return "🍎";
    case "Linux":
      return "🐧";
    case "Android":
      return "🤖";
    case "iOS":
      return "📱";
    default:
      return "🖥️";
  }
}

export default function SessionsPage() {
  const queryClient = useQueryClient();
  const toast = useToast();

  const {
    data: sessions,
    isLoading,
    isError,
  } = useQuery({ queryKey: ["sessions"], queryFn: api.sessions.list });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["sessions"] });

  const revokeMutation = useMutation({
    mutationFn: (id: number) => api.sessions.revoke(id),
    onSuccess: () => {
      invalidate();
      toast.success("Sessão revogada com sucesso!");
    },
    onError: (err) => {
      toast.error(parseApiErrorMessage(err, "Erro ao revogar sessão."));
    },
  });

  const revokeAllMutation = useMutation({
    mutationFn: api.sessions.revokeAll,
    onSuccess: () => {
      invalidate();
      toast.success("Todas as outras sessões foram encerradas!");
    },
    onError: (err) => {
      toast.error(parseApiErrorMessage(err, "Erro ao encerrar outras sessões."));
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20 text-app-muted">
        Carregando sessões...
      </div>
    );
  }

  if (isError) {
    return (
      <div className="rounded-2xl border border-red-500/10 bg-red-500/5 p-6 text-center text-red-400">
        Erro ao carregar sessões ativas.
      </div>
    );
  }

  const otherSessionsCount = sessions?.filter((s) => !s.isCurrent).length || 0;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-app-text">Sessões Ativas</h1>
          <p className="text-sm text-app-muted">
            Gerencie os dispositivos e navegadores que estão conectados à sua conta.
          </p>
        </div>

        {otherSessionsCount > 0 && (
          <button
            onClick={() => {
              if (confirm("Tem certeza que deseja encerrar todas as outras sessões?")) {
                revokeAllMutation.mutate();
              }
            }}
            disabled={revokeAllMutation.isPending}
            className="w-full sm:w-auto rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-2.5 text-sm font-semibold text-red-400 transition-all duration-200 hover:bg-red-500/20 active:scale-95 disabled:opacity-50 cursor-pointer"
          >
            {revokeAllMutation.isPending ? "Encerrando..." : "Desconectar outros dispositivos"}
          </button>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {sessions?.map((session: ActiveSession) => {
          const { browser, os } = parseUserAgent(session.userAgent);
          return (
            <div
              key={session.id}
              className={`relative rounded-2xl border bg-app-card p-6 transition-all duration-300 hover:bg-app-card-hover ${
                session.isCurrent
                  ? "border-indigo-500/30 shadow-[0_0_20px_rgba(99,102,241,0.05)]"
                  : "border-app-border"
              }`}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-3.5">
                  <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-app-bg border border-app-border text-2xl">
                    {getOSEmoji(os)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-app-text flex items-center gap-2">
                      {os} • {browser}
                      {session.isCurrent && (
                        <span className="inline-flex rounded-full bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 text-[10px] font-semibold text-emerald-400 uppercase tracking-wider">
                          Este dispositivo
                        </span>
                      )}
                    </h3>
                    <p className="text-xs text-app-muted mt-0.5">
                      IP: {session.ipAddress || "Desconhecido"}
                    </p>
                  </div>
                </div>

                {!session.isCurrent && (
                  <button
                    onClick={() => {
                      if (confirm("Deseja desconectar esta sessão?")) {
                        revokeMutation.mutate(session.id);
                      }
                    }}
                    disabled={revokeMutation.isPending}
                    className="rounded-lg border border-app-border bg-app-bg/50 p-2 text-app-muted hover:text-red-400 hover:border-red-500/20 hover:bg-red-500/5 transition cursor-pointer"
                    title="Desconectar dispositivo"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    >
                      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                      <polyline points="16 17 21 12 16 7" />
                      <line x1="21" x2="9" y1="12" y2="12" />
                    </svg>
                  </button>
                )}
              </div>

              <div className="mt-6 pt-4 border-t border-app-border/40 grid grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-app-muted block">Primeiro acesso</span>
                  <span className="font-medium text-app-text">
                    {new Date(session.createdAt).toLocaleString("pt-BR")}
                  </span>
                </div>
                <div>
                  <span className="text-app-muted block">Última atividade</span>
                  <span className="font-medium text-app-text">
                    {session.lastActivityAt
                      ? new Date(session.lastActivityAt).toLocaleString("pt-BR")
                      : "Sem registro"}
                  </span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
