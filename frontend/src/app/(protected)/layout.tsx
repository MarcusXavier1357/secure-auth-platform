import { Suspense } from "react";
import { Outlet } from "react-router-dom";
import { RequireAuth } from "../../components/RequireAuth";
import { AppShell } from "../../components/AppShell";

function PageSkeleton() {
  return <div className="flex items-center justify-center py-20 text-slate-400">Carregando...</div>;
}

// Layout das rotas protegidas: exige sessão e envolve tudo no AppShell.
// O Suspense cobre as páginas admin carregadas via lazy().
export function ProtectedLayout() {
  return (
    <RequireAuth>
      <AppShell>
        <Suspense fallback={<PageSkeleton />}>
          <Outlet />
        </Suspense>
      </AppShell>
    </RequireAuth>
  );
}
