import { Navigate, Outlet } from "react-router-dom";
import { usePermission } from "../hooks/usePermission";
import { paths } from "../router/paths";

// Guard de rota: bloqueia a subárvore inteira se o usuário não tem a
// permissão. Diferente do Can (que só esconde UI), este impede a navegação.
// Lembrete: isso é cosmético — o backend revalida via RequirePermission.
export function RequirePermission({ code }: { code: string }) {
  const allowed = usePermission(code);

  if (!allowed) {
    return <Navigate to={paths.home()} replace />;
  }
  return <Outlet />;
}
