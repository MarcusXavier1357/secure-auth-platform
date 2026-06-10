import { Outlet } from "react-router-dom";

// Layout das rotas públicas: conteúdo centralizado, sem shell de navegação.
export function PublicLayout() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-950 px-4">
      <Outlet />
    </div>
  );
}
