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
        setLocalError(
          "A senha deve conter ao menos uma letra maiúscula, uma minúscula e um número.",
        );
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
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-fade-in cursor-pointer"
      onClick={onClose}
    >
      <div
        className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-3xl border border-app-border bg-app-card p-8 shadow-2xl backdrop-blur-xl animate-scale-in cursor-default"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between pb-4 border-b border-app-border/60">
          <h3 className="text-lg font-bold text-app-text">
            {isEdit ? "Editar Usuário" : "Criar Novo Usuário"}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-app-muted hover:text-app-text transition-colors cursor-pointer"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-6 space-y-5">
          {showFields ? (
            <>
              <div className="space-y-1.5">
                <label
                  htmlFor="user-name"
                  className="block text-xs font-semibold uppercase tracking-wider text-app-muted"
                >
                  Nome Completo
                </label>
                <input
                  id="user-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 text-sm text-app-text placeholder-app-muted/50 outline-none transition-all focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
                  placeholder="Nome do usuário"
                />
              </div>
              <div className="space-y-1.5">
                <label
                  htmlFor="user-email"
                  className="block text-xs font-semibold uppercase tracking-wider text-app-muted"
                >
                  Endereço de E-mail
                </label>
                <input
                  id="user-email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 text-sm text-app-text placeholder-app-muted/50 outline-none transition-all focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
                  placeholder="exemplo@empresa.com"
                />
              </div>
              <div className="space-y-1.5">
                <label
                  htmlFor="user-role"
                  className="block text-xs font-semibold uppercase tracking-wider text-app-muted"
                >
                  Perfil de Acesso (Role)
                </label>
                <select
                  id="user-role"
                  value={roleId ?? ""}
                  onChange={(e) => setRoleId(e.target.value ? Number(e.target.value) : null)}
                  className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 text-sm text-app-text outline-none transition-all focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
                >
                  <option value="" className="bg-app-bg text-app-text">
                    Nenhum
                  </option>
                  {roles?.map((role) => (
                    <option key={role.id} value={role.id} className="bg-app-bg text-app-text">
                      {role.name}
                    </option>
                  ))}
                </select>
              </div>
            </>
          ) : null}

          {(!isEdit || canResetPassword) &&
            (() => {
              const strength = checkPasswordStrength(password, name, email);
              return (
                <div className="space-y-4 rounded-2xl border border-app-border bg-app-bg/40 p-5">
                  <p className="block text-xs font-semibold uppercase tracking-wider text-app-muted">
                    {isEdit ? "Redefinir senha (opcional)" : "Definir senha de acesso"}
                  </p>
                  <div className="space-y-3">
                    <input
                      type="password"
                      placeholder={isEdit ? "Nova senha" : "Senha (mínimo de 12 caracteres)"}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 text-sm text-app-text placeholder-app-muted/50 outline-none transition-all focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
                    />
                    <input
                      type="password"
                      placeholder="Confirmar nova senha"
                      value={passwordConfirm}
                      onChange={(e) => setPasswordConfirm(e.target.value)}
                      className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 text-sm text-app-text placeholder-app-muted/50 outline-none transition-all focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
                    />
                  </div>

                  {password && (
                    <div className="space-y-3 mt-3 pt-3 border-t border-app-border/60">
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-app-muted font-medium">Complexidade da senha:</span>
                        <span
                          className={
                            strength.score <= 1
                              ? "text-red-450 font-bold"
                              : strength.score <= 3
                                ? "text-yellow-500 dark:text-yellow-400 font-bold"
                                : "text-emerald-500 dark:text-emerald-400 font-bold"
                          }
                        >
                          {strength.score <= 1 ? "Fraca" : strength.score <= 3 ? "Média" : "Forte"}
                        </span>
                      </div>
                      <div className="h-1.5 w-full rounded-full bg-app-bg overflow-hidden border border-app-border/20">
                        <div
                          className={`h-full transition-all duration-300 ${
                            strength.score <= 1
                              ? "bg-red-500 w-1/3"
                              : strength.score <= 3
                                ? "bg-yellow-500 w-2/3"
                                : "bg-emerald-500 w-full"
                          }`}
                        />
                      </div>

                      <ul className="text-[11px] space-y-1.5 text-app-muted mt-2">
                        <li className="flex items-center gap-2">
                          <span
                            className={
                              strength.hasLength
                                ? "text-emerald-500 dark:text-emerald-400 font-bold"
                                : "text-app-muted/40"
                            }
                          >
                            {strength.hasLength ? "●" : "○"}
                          </span>
                          <span className={strength.hasLength ? "text-app-text" : ""}>
                            Mínimo de 12 caracteres
                          </span>
                        </li>
                        <li className="flex items-center gap-2">
                          <span
                            className={
                              strength.hasLower && strength.hasUpper && strength.hasDigit
                                ? "text-emerald-500 dark:text-emerald-400 font-bold"
                                : "text-app-muted/40"
                            }
                          >
                            {strength.hasLower && strength.hasUpper && strength.hasDigit
                              ? "●"
                              : "○"}
                          </span>
                          <span
                            className={
                              strength.hasLower && strength.hasUpper && strength.hasDigit
                                ? "text-app-text"
                                : ""
                            }
                          >
                            Letra maiúscula, minúscula e número
                          </span>
                        </li>
                        <li className="flex items-center gap-2">
                          <span
                            className={
                              strength.noSequenceOrRepetition
                                ? "text-emerald-500 dark:text-emerald-400 font-bold"
                                : "text-app-muted/40"
                            }
                          >
                            {strength.noSequenceOrRepetition ? "●" : "○"}
                          </span>
                          <span className={strength.noSequenceOrRepetition ? "text-app-text" : ""}>
                            Sem padrões previsíveis, sequências ou dados pessoais
                          </span>
                        </li>
                      </ul>
                    </div>
                  )}
                </div>
              );
            })()}

          {(localError || error) && (
            <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-xs text-red-500 dark:text-red-400">
              <span className="font-semibold">Erro: </span> {localError ?? error}
            </div>
          )}

          <div className="flex justify-end gap-3 pt-3 border-t border-app-border/60">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-app-border bg-app-bg/40 px-5 py-2.5 text-sm font-semibold text-app-text transition-all hover:bg-app-card-hover cursor-pointer"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={pending}
              className="rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-bold text-white shadow-lg transition-all hover:bg-indigo-500 active:scale-95 disabled:cursor-not-allowed disabled:opacity-60 cursor-pointer"
            >
              {pending ? "Salvando..." : isEdit ? "Salvar alterações" : "Criar usuário"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
