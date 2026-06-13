import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api, type UserWithPermissions } from "../services/api";
import { usePermission } from "../hooks/usePermission";

export interface UserFormData {
  name: string;
  email: string;
  password?: string;
  passwordConfirm?: string;
  roleId: number | null;
}

interface UserFormModalProps {
  user: UserWithPermissions | null;
  pending: boolean;
  error: string | null;
  onClose: () => void;
  onSubmit: (data: UserFormData) => void;
}

export function UserFormModal({ user, pending, error, onClose, onSubmit }: UserFormModalProps) {
  const isEdit = user !== null;
  const canResetPassword = usePermission("users.password.reset");
  const canUpdate = usePermission("users.update");

  const [name, setName] = useState(user?.name ?? "");
  const [email, setEmail] = useState(user?.email ?? "");
  const [password, setPassword] = useState("");
  const [passwordConfirm, setPasswordConfirm] = useState("");
  const [roleId, setRoleId] = useState<number | null>(user?.roleId ?? null);
  const [localError, setLocalError] = useState<string | null>(null);

  const { data: roles } = useQuery({
    queryKey: ["roles"],
    queryFn: api.roles.list,
    enabled: canUpdate || !isEdit,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null);

    if (!name.trim() || !email.trim()) {
      setLocalError("Nome e e-mail são obrigatórios.");
      return;
    }

    if (!isEdit && (!password || password.length < 8)) {
      setLocalError("Senha deve ter pelo menos 8 caracteres.");
      return;
    }

    if (password && password !== passwordConfirm) {
      setLocalError("As senhas não coincidem.");
      return;
    }

    const data: UserFormData = {
      name: name.trim(),
      email: email.trim(),
      roleId,
    };

    if (password) {
      if (!canResetPassword && isEdit) {
        setLocalError("Você não tem permissão para redefinir senha.");
        return;
      }
      data.password = password;
    }

    onSubmit(data);
  };

  const showFields = !isEdit || canUpdate;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl border border-slate-800 bg-slate-900 p-6">
        <h3 className="text-lg font-semibold">{isEdit ? "Editar usuário" : "Novo usuário"}</h3>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          {showFields ? (
            <>
              <div>
                <label htmlFor="user-name" className="mb-1.5 block text-sm font-medium text-slate-300">
                  Nome
                </label>
                <input
                  id="user-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                />
              </div>
              <div>
                <label htmlFor="user-email" className="mb-1.5 block text-sm font-medium text-slate-300">
                  E-mail
                </label>
                <input
                  id="user-email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                />
              </div>
              <div>
                <label htmlFor="user-role" className="mb-1.5 block text-sm font-medium text-slate-300">
                  Role
                </label>
                <select
                  id="user-role"
                  value={roleId ?? ""}
                  onChange={(e) => setRoleId(e.target.value ? Number(e.target.value) : null)}
                  className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                  <option value="">Nenhuma</option>
                  {roles?.map((role) => (
                    <option key={role.id} value={role.id}>
                      {role.name}
                    </option>
                  ))}
                </select>
              </div>
            </>
          ) : null}

          {(!isEdit || canResetPassword) && (
            <div className="space-y-3 rounded-lg border border-slate-800 bg-slate-950/50 p-4">
              <p className="text-xs font-medium text-slate-400">
                {isEdit ? "Redefinir senha (opcional)" : "Senha"}
              </p>
              <input
                type="password"
                placeholder={isEdit ? "Nova senha" : "Senha (mín. 8 caracteres)"}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
              />
              <input
                type="password"
                placeholder="Confirmar senha"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
              />
            </div>
          )}

          {(localError || error) && (
            <p className="text-sm text-red-400">{localError ?? error}</p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={pending}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
            >
              {pending ? "Salvando..." : isEdit ? "Salvar" : "Criar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
