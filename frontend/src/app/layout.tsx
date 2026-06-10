import { Outlet } from "react-router-dom";

// Layout raiz — ponto único para error boundaries e providers de UI futuros.
export function RootLayout() {
  return <Outlet />;
}
