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

interface PasswordStrength {
  score: number;
  hasLength: boolean;
  hasUpper: boolean;
  hasLower: boolean;
  hasDigit: boolean;
  noSequenceOrRepetition: boolean;
}

function checkPasswordStrength(password: string, name?: string, email?: string): PasswordStrength {
  const hasLength = password.length >= 12;
  const hasUpper = /[A-Z]/.test(password);
  const hasLower = /[a-z]/.test(password);
  const hasDigit = /[0-9]/.test(password);

  let noSequenceOrRepetition = true;
  const lower = password.toLowerCase();

  // 1. Sequence checks (length >= 8)
  if (lower.length >= 8) {
    for (let i = 0; i <= lower.length - 8; i++) {
      let isAsc = true;
      let isDesc = true;
      for (let j = 0; j < 7; j++) {
        if (lower.charCodeAt(i + j + 1) !== lower.charCodeAt(i + j) + 1) isAsc = false;
        if (lower.charCodeAt(i + j + 1) !== lower.charCodeAt(i + j) - 1) isDesc = false;
      }
      if (isAsc || isDesc) {
        noSequenceOrRepetition = false;
        break;
      }
    }
  }

  // 2. Keyboard sequences
  if (noSequenceOrRepetition && lower.length >= 8) {
    const rows = ["qwertyuiop", "asdfghjkl", "zxcvbnm"];
    for (const row of rows) {
      const revRow = row.split("").reverse().join("");
      for (let i = 0; i <= lower.length - 8; i++) {
        const sub = lower.substring(i, i + 8);
        if (row.includes(sub) || revRow.includes(sub)) {
          noSequenceOrRepetition = false;
          break;
        }
      }
      if (!noSequenceOrRepetition) break;
    }
  }

  // 3. Repetition checks (length >= 10 total)
  if (noSequenceOrRepetition) {
    for (let k = 1; k <= 3; k++) {
      for (let i = 0; i <= password.length - k * 3; i++) {
        const pattern = password.substring(i, i + k);
        let repeats = 1;
        for (let j = i + k; j + k <= password.length; j += k) {
          if (password.substring(j, j + k) === pattern) {
            repeats++;
          } else {
            break;
          }
        }
        if (repeats * k >= 10) {
          noSequenceOrRepetition = false;
          break;
        }
      }
      if (!noSequenceOrRepetition) break;
    }
  }

  // 4. Personal data checks
  let containsPersonal = false;
  if (name) {
    const parts = name.toLowerCase().split(/[\s\-_.]+/);
    for (const part of parts) {
      if (part.length >= 3 && lower.includes(part)) {
        containsPersonal = true;
      }
    }
  }
  if (email) {
    const eLower = email.toLowerCase();
    if (lower.includes(eLower)) {
      containsPersonal = true;
    }
    const idx = eLower.indexOf("@");
    if (idx > 0) {
      const prefix = eLower.substring(0, idx);
      if (prefix.length >= 3 && lower.includes(prefix)) {
        containsPersonal = true;
      }
    }
  }

  let score = 0;
  if (hasLength) score++;
  if (hasUpper && hasLower) score++;
  if (hasDigit) score++;
  if (noSequenceOrRepetition && !containsPersonal) score++;

  return {
    score,
    hasLength,
    hasUpper,
    hasLower,
    hasDigit,
    noSequenceOrRepetition: noSequenceOrRepetition && !containsPersonal,
  };
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

    if (password) {
      if (password.length < 12) {
        setLocalError("A senha deve possuir pelo menos 12 caracteres.");
        return;
      }
      const hasLower = /[a-z]/.test(password);
      const hasUpper = /[A-Z]/.test(password);
      const hasDigit = /[0-9]/.test(password);
      if (!hasLower || !hasUpper || !hasDigit) {
        setLocalError("A senha deve conter ao menos uma letra maiúscula, uma minúscula e um número.");
        return;
      }
      
      const strength = checkPasswordStrength(password, name, email);
      if (strength.score < 4) {
        setLocalError("A senha é considerada fraca. Escolha uma senha mais forte.");
        return;
      }
    } else if (!isEdit) {
      setLocalError("Senha é obrigatória.");
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

          {(!isEdit || canResetPassword) && (() => {
            const strength = checkPasswordStrength(password, name, email);
            return (
              <div className="space-y-3 rounded-lg border border-slate-800 bg-slate-950/50 p-4">
                <p className="text-xs font-medium text-slate-400">
                  {isEdit ? "Redefinir senha (opcional)" : "Senha"}
                </p>
                <input
                  type="password"
                  placeholder={isEdit ? "Nova senha" : "Senha (mín. 12 caracteres)"}
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

                {password && (
                  <div className="space-y-2 mt-2 pt-2 border-t border-slate-800/60">
                    {/* Força da Senha */}
                    <div className="flex items-center justify-between text-[11px]">
                      <span className="text-slate-400">Força da senha:</span>
                      <span className={
                        strength.score <= 1 ? "text-red-400 font-semibold" :
                        strength.score <= 3 ? "text-yellow-400 font-semibold" :
                        "text-emerald-400 font-semibold"
                      }>
                        {strength.score <= 1 ? "Fraca" : strength.score <= 3 ? "Média" : "Forte"}
                      </span>
                    </div>
                    <div className="h-1.5 w-full rounded-full bg-slate-800 overflow-hidden">
                      <div
                        className={`h-full transition-all duration-300 ${
                          strength.score <= 1 ? "bg-red-500 w-1/3" :
                          strength.score <= 3 ? "bg-yellow-500 w-2/3" :
                          "bg-emerald-500 w-full"
                        }`}
                      />
                    </div>

                    {/* Requisitos Checklist */}
                    <ul className="text-[11px] space-y-1 text-slate-400 mt-2">
                      <li className="flex items-center gap-1.5">
                        <span className={strength.hasLength ? "text-emerald-400 font-bold" : "text-slate-500"}>
                          {strength.hasLength ? "✓" : "○"} Mínimo de 12 caracteres
                        </span>
                      </li>
                      <li className="flex items-center gap-1.5">
                        <span className={(strength.hasLower && strength.hasUpper && strength.hasDigit) ? "text-emerald-400 font-bold" : "text-slate-500"}>
                          {(strength.hasLower && strength.hasUpper && strength.hasDigit) ? "✓" : "○"} Letra maiúscula, minúscula e número
                        </span>
                      </li>
                      <li className="flex items-center gap-1.5">
                        <span className={strength.noSequenceOrRepetition ? "text-emerald-400 font-bold" : "text-slate-500"}>
                          {strength.noSequenceOrRepetition ? "✓" : "○"} Sem padrões previsíveis ou dados pessoais
                        </span>
                      </li>
                    </ul>
                  </div>
                )}
              </div>
            );
          })()}

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
