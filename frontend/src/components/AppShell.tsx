import { type ReactNode, useState, useEffect } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../providers/AuthProvider";
import { Can } from "./Can";
import { paths } from "../router/paths";
import { adminRoutes } from "../router/admin-routes";

const navItems = [
  { to: paths.home(), label: "Início", permission: null as string | null },
  ...adminRoutes.map((route) => ({
    to: route.path,
    label: route.navLabel,
    permission: route.permission,
  })),
];

function SidebarLink({ to, label }: { to: string; label: string }) {
  return (
    <NavLink
      to={to}
      end={to === paths.home()}
      className={({ isActive }) =>
        `block rounded-xl px-4 py-2.5 text-sm font-medium transition-all duration-200 ${
          isActive
            ? "bg-indigo-500/15 border-l-2 border-indigo-500 text-indigo-300 shadow-[0_0_15px_rgba(99,102,241,0.1)]"
            : "text-app-muted hover:bg-app-card-hover hover:text-app-text border-l-2 border-transparent"
        }`
      }
    >
      {label}
    </NavLink>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [theme, setTheme] = useState(() => localStorage.getItem("theme") || "dark");
  const [isCollapsed, setIsCollapsed] = useState(false);

  useEffect(() => {
    const root = document.documentElement;
    if (theme === "light") {
      root.classList.add("light");
    } else {
      root.classList.remove("light");
    }
    localStorage.setItem("theme", theme);
  }, [theme]);

  async function handleLogout() {
    await logout();
    navigate(paths.login(), { replace: true });
  }

  const toggleTheme = () => {
    setTheme((prev) => (prev === "light" ? "dark" : "light"));
  };

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-app-bg text-app-text font-sans">
      {/* Fixed/Collapsible Sidebar */}
      <aside className={`flex flex-col border-app-border bg-app-card backdrop-blur-xl h-screen justify-between transition-all duration-300 ease-in-out ${
        isCollapsed ? "w-0 overflow-hidden border-r-0" : "w-64 border-r shrink-0"
      }`}>
        <div>
          <div className="border-b border-app-border px-6 py-6 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-tr from-indigo-500 to-purple-600 flex items-center justify-center font-bold text-sm tracking-wider shadow-[0_0_20px_rgba(99,102,241,0.4)] text-white">
                AS
              </div>
              <Link to={paths.home()} className="text-lg font-bold tracking-tight text-app-text">
                Auth System
              </Link>
            </div>
            
            <button
              onClick={() => setIsCollapsed(true)}
              className="p-1.5 rounded-lg hover:bg-app-card-hover text-app-muted hover:text-app-text transition cursor-pointer"
              title="Recolher Sidebar"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect width="18" height="18" x="3" y="3" rx="2" />
                <path d="M9 3v18" />
              </svg>
            </button>
          </div>

          <nav className="space-y-1.5 px-4 py-6">
            {navItems.map((item) =>
              item.permission ? (
                <Can key={item.to} permission={item.permission}>
                  <SidebarLink to={item.to} label={item.label} />
                </Can>
              ) : (
                <SidebarLink key={item.to} to={item.to} label={item.label} />
              ),
            )}
          </nav>
        </div>

        {/* User Card */}
        <div className="border-t border-app-border p-5 bg-app-bg/10">
          <div className="rounded-2xl bg-app-card border border-app-border p-4 space-y-4 shadow-sm">
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold text-app-text">{user?.name}</p>
              <p className="truncate text-xs text-app-muted">{user?.email}</p>
            </div>
            
            {user?.role && (
              <div>
                <span className="inline-flex rounded-lg bg-indigo-500/10 border border-indigo-500/20 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-indigo-400">
                  {user.role.name}
                </span>
              </div>
            )}

            <div className="flex items-center gap-2">
              <button
                onClick={handleLogout}
                className="flex-1 rounded-xl border border-app-border bg-app-bg/40 py-2.5 text-xs font-semibold text-app-text transition-all duration-200 hover:bg-red-500/10 hover:text-red-500 hover:border-red-500/20 cursor-pointer"
              >
                Sair da conta
              </button>
              
              {/* Theme Toggle inside user box */}
              <button
                onClick={toggleTheme}
                className="p-2.5 rounded-xl border border-app-border bg-app-bg/40 hover:bg-app-card-hover text-sm transition-all active:scale-90 cursor-pointer"
                title={theme === "light" ? "Modo Escuro" : "Modo Claro"}
              >
                {theme === "light" ? "🌙" : "☀️"}
              </button>
            </div>
          </div>
        </div>
      </aside>

      {/* Independently Scrolling Content Area */}
      <main className="min-w-0 flex-1 overflow-y-auto px-10 py-10 h-screen relative">
        {isCollapsed && (
          <button
            onClick={() => setIsCollapsed(false)}
            className="fixed top-6 left-3 z-50 flex h-10 w-10 items-center justify-center rounded-full border border-app-border bg-app-card backdrop-blur-md text-app-text shadow-lg hover:bg-app-card-hover transition-all duration-200 active:scale-95 cursor-pointer"
            title="Expandir Sidebar"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect width="18" height="18" x="3" y="3" rx="2" />
              <path d="M9 3v18" />
            </svg>
          </button>
        )}
        <div className="max-w-7xl mx-auto space-y-8">
          {children}
        </div>
      </main>
    </div>
  );
}
