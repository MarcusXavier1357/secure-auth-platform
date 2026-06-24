import { createContext, useContext, useState, useCallback, type ReactNode } from "react";

export type ToastType = "success" | "error" | "info";

export interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

interface ToastContextType {
  toast: (message: string, type?: ToastType) => void;
  success: (message: string) => void;
  error: (message: string) => void;
  info: (message: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = useCallback((message: string, type: ToastType = "info") => {
    const id = Math.random().toString(36).substring(2, 9);
    setToasts((prev) => [...prev, { id, message, type }]);

    // Auto-remove após 4 segundos
    setTimeout(() => {
      removeToast(id);
    }, 4000);
  }, [removeToast]);

  const success = useCallback((msg: string) => toast(msg, "success"), [toast]);
  const error = useCallback((msg: string) => toast(msg, "error"), [toast]);
  const info = useCallback((msg: string) => toast(msg, "info"), [toast]);

  return (
    <ToastContext.Provider value={{ toast, success, error, info }}>
      {children}
      
      {/* Toast Render Area */}
      <div className="fixed top-6 right-6 z-[9999] flex flex-col gap-3 w-full max-w-sm pointer-events-none">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`pointer-events-auto flex items-center justify-between gap-3 rounded-2xl border p-4 shadow-xl backdrop-blur-md transition-all duration-300 animate-slide-in ${
              t.type === "success"
                ? "bg-emerald-950/85 dark:bg-emerald-950/85 text-emerald-250 dark:text-emerald-300 border-emerald-500/30"
                : t.type === "error"
                ? "bg-red-950/85 dark:bg-red-950/85 text-red-250 dark:text-red-300 border-red-500/30"
                : "bg-app-card/90 text-app-text border-app-border"
            }`}
          >
            <div className="flex items-center gap-3">
              {t.type === "success" && (
                <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-emerald-500/20 text-xs font-bold text-emerald-400">
                  ✓
                </span>
              )}
              {t.type === "error" && (
                <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-red-500/20 text-xs font-bold text-red-400">
                  ✕
                </span>
              )}
              {t.type === "info" && (
                <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-indigo-500/20 text-xs font-bold text-indigo-400">
                  ℹ
                </span>
              )}
              <p className="text-sm font-semibold leading-snug">{t.message}</p>
            </div>
            
            <button
              onClick={() => removeToast(t.id)}
              className="text-current opacity-60 hover:opacity-100 transition-opacity p-1 text-xs shrink-0 cursor-pointer"
            >
              ✕
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast deve ser usado dentro de um ToastProvider");
  }
  return context;
}
