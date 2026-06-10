import { lazy } from "react";
import { Link, Route, Routes } from "react-router-dom";
import { RootLayout } from "../app/layout";
import { PublicLayout } from "../app/(public)/layout";
import { ProtectedLayout } from "../app/(protected)/layout";
import { RequirePermission } from "../components/RequirePermission";
import { paths } from "./paths";
import LoginPage from "../app/(public)/login/page";
import DashboardPage from "../app/(protected)/page";

// Páginas admin via lazy(): só entram no bundle de quem navega até elas.
const UsersPage = lazy(() => import("../app/(protected)/users/page"));
const PermissionsPage = lazy(() => import("../app/(protected)/permissions/page"));
const AuditPage = lazy(() => import("../app/(protected)/audit/page"));

function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-3 bg-slate-950 text-white">
      <h1 className="text-3xl font-semibold">404</h1>
      <p className="text-sm text-slate-400">Página não encontrada.</p>
      <Link to={paths.home()} className="text-sm text-indigo-400 hover:underline">
        Voltar ao início
      </Link>
    </div>
  );
}

export function AppRouter() {
  return (
    <Routes>
      <Route element={<RootLayout />}>
        <Route element={<PublicLayout />}>
          <Route path={paths.login()} element={<LoginPage />} />
        </Route>

        <Route element={<ProtectedLayout />}>
          <Route index element={<DashboardPage />} />

          <Route element={<RequirePermission code="users.manage" />}>
            <Route path={paths.users()} element={<UsersPage />} />
          </Route>
          <Route element={<RequirePermission code="permissions.manage" />}>
            <Route path={paths.permissions()} element={<PermissionsPage />} />
          </Route>
          <Route element={<RequirePermission code="audit_logs.read" />}>
            <Route path={paths.audit()} element={<AuditPage />} />
          </Route>
        </Route>
      </Route>

      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
