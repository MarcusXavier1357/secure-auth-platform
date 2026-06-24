import { Outlet } from "react-router-dom";

// Layout das rotas públicas: conteúdo centralizado, sem shell de navegação.
export function PublicLayout() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-950 px-4 relative overflow-hidden font-sans">
      {/* Glow effects */}
      <div className="absolute -top-40 -left-40 w-96 h-96 rounded-full bg-indigo-500/10 blur-[120px] pointer-events-none" />
      <div className="absolute -bottom-40 -right-40 w-96 h-96 rounded-full bg-purple-500/10 blur-[120px] pointer-events-none" />
      
      <div className="z-10 w-full flex items-center justify-center">
        <Outlet />
      </div>
    </div>
  );
}
