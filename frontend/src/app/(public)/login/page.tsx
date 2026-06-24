import { useState, type FormEvent } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { useAuth } from "../../../providers/AuthProvider";
import { ApiError } from "../../../services/api";
import { paths } from "../../../router/paths";

export default function LoginPage() {
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  if (user) {
    return <Navigate to={paths.home()} replace />;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      await login(email, password);
      navigate(paths.home(), { replace: true });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError("E-mail ou senha incorretos.");
      } else if (err instanceof ApiError && err.status === 429) {
        setError("Muitas tentativas. Aguarde alguns minutos e tente novamente.");
      } else if (err instanceof ApiError && err.status === 503) {
        setError("Login temporariamente indisponível. Tente novamente em instantes.");
      } else {
        setError("Erro ao entrar. Tente novamente.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="w-full max-w-md px-4">
      {/* Header */}
      <div className="mb-8 text-center space-y-3">
        <div className="mx-auto h-12 w-12 rounded-2xl bg-gradient-to-tr from-indigo-500 to-purple-600 flex items-center justify-center font-bold text-lg tracking-wider shadow-[0_0_30px_rgba(99,102,241,0.3)] animate-pulse">
          AS
        </div>
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white">Auth System</h1>
          <p className="mt-2 text-sm text-slate-400">Entre com sua conta administrativa</p>
        </div>
      </div>

      {/* Glass Card Form */}
      <form
        onSubmit={handleSubmit}
        className="space-y-5 rounded-3xl border border-slate-800 bg-slate-900/40 p-8 shadow-[0_20px_50px_rgba(0,0,0,0.5)] backdrop-blur-xl"
      >
        <div className="space-y-1.5">
          <label
            htmlFor="email"
            className="block text-xs font-semibold uppercase tracking-wider text-slate-400"
          >
            E-mail
          </label>
          <input
            id="email"
            type="email"
            required
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-xl border border-slate-800/80 bg-slate-950/60 px-4 py-3 text-sm text-white placeholder-slate-600 outline-none transition-all duration-200 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
            placeholder="exemplo@empresa.com"
          />
        </div>

        <div className="space-y-1.5">
          <label
            htmlFor="password"
            className="block text-xs font-semibold uppercase tracking-wider text-slate-400"
          >
            Senha
          </label>
          <input
            id="password"
            type="password"
            required
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-xl border border-slate-800/80 bg-slate-950/60 px-4 py-3 text-sm text-white placeholder-slate-600 outline-none transition-all duration-200 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
            placeholder="••••••••••••"
          />
        </div>

        {error && (
          <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-xs text-red-400 animate-shake">
            <span className="font-semibold">Erro: </span> {error}
          </div>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 py-3 text-sm font-bold text-white shadow-lg transition-all duration-200 hover:from-indigo-500 hover:to-indigo-400 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-65"
        >
          {submitting ? "Autenticando..." : "Entrar"}
        </button>
      </form>
    </div>
  );
}
