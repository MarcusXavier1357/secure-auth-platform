import { lazy } from "react";
import { Link, Route, Routes } from "react-router-dom";
import { RootLayout } from "../app/layout";
import { PublicLayout } from "../app/(public)/layout";
import { ProtectedLayout } from "../app/(protected)/layout";
import { RequirePermission } from "../components/RequirePermission";
import { adminRoutes, type AdminRouteId } from "./admin-routes";
import { paths } from "./paths";
import LoginPage from "../app/(public)/login/page";
import DashboardPage from "../app/(protected)/page";

const adminPages: Record<AdminRouteId, ReturnType<typeof lazy>> = {
  users: lazy(() => import("../app/(protected)/users/page")),
  permissions: lazy(() => import("../app/(protected)/permissions/page")),
  audit: lazy(() => import("../app/(protected)/audit/page")),
};

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

          {adminRoutes.map((route) => {
            const Page = adminPages[route.id];
            return (
              <Route key={route.id} element={<RequirePermission code={route.permission} />}>
                <Route path={route.path} element={<Page />} />
              </Route>
            );
          })}
        </Route>
      </Route>

      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
