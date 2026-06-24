import { useState, useRef } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type Permission } from "../../../services/api";
import { Can } from "../../../components/Can";
import { parseApiErrorMessage } from "../../../utils/apiError";

const PROTECTED_CODES = new Set(["*", "users.*", "audit_logs.read"]);

export default function PermissionsPage() {
  const queryClient = useQueryClient();
  const userSectionRef = useRef<HTMLDivElement>(null);
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const [catalogModal, setCatalogModal] = useState<"create" | { edit: Permission } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Permission | null>(null);
  const [revokeTarget, setRevokeTarget] = useState<{
    userId: number;
    permission: Permission;
  } | null>(null);
  const [catalogError, setCatalogError] = useState<string | null>(null);

  const {
    data: permissions,
    isLoading: permissionsLoading,
    isError: permissionsError,
  } = useQuery({
    queryKey: ["permissions"],
    queryFn: api.permissions.list,
  });
  const {
    data: users,
    isLoading: usersLoading,
    isError: usersError,
  } = useQuery({ queryKey: ["users"], queryFn: api.users.list });

  const selectedUser = users?.find((u) => u.id === selectedUserId) ?? null;
  const grantedIds = new Set(selectedUser?.permissions?.map((p) => p.id) ?? []);

  const invalidateUsers = () => queryClient.invalidateQueries({ queryKey: ["users"] });
  const invalidatePerms = () => queryClient.invalidateQueries({ queryKey: ["permissions"] });

  const createPerm = useMutation({
    mutationFn: api.permissions.create,
    onSuccess: () => {
      invalidatePerms();
      setCatalogModal(null);
      setCatalogError(null);
    },
    onError: (err) => setCatalogError(parseApiErrorMessage(err, "Erro ao criar permissão.")),
  });

  const updatePerm = useMutation({
    mutationFn: ({ id, description }: { id: number; description: string }) =>
      api.permissions.update(id, { description }),
    onSuccess: () => {
      invalidatePerms();
      setCatalogModal(null);
      setCatalogError(null);
    },
    onError: (err) => setCatalogError(parseApiErrorMessage(err, "Erro ao atualizar permissão.")),
  });

  const deletePerm = useMutation({
    mutationFn: api.permissions.delete,
    onSuccess: () => {
      invalidatePerms();
      setDeleteTarget(null);
    },
    onError: (err) => {
      alert(parseApiErrorMessage(err, "Erro ao excluir permissão."));
      setDeleteTarget(null);
    },
  });

  const grant = useMutation({
    mutationFn: ({ userId, permissionId }: { userId: number; permissionId: number }) =>
      api.permissions.grant(userId, permissionId),
    onSuccess: invalidateUsers,
    onError: (err) => alert(parseApiErrorMessage(err, "Erro ao conceder permissão.")),
  });

  const revoke = useMutation({
    mutationFn: ({ userId, permissionId }: { userId: number; permissionId: number }) =>
      api.permissions.revoke(userId, permissionId),
    onSuccess: () => {
      invalidateUsers();
      setRevokeTarget(null);
    },
    onError: (err) => {
      alert(parseApiErrorMessage(err, "Erro ao revogar permissão."));
      setRevokeTarget(null);
    },
  });

  const pending =
    grant.isPending ||
    revoke.isPending ||
    createPerm.isPending ||
    updatePerm.isPending ||
    deletePerm.isPending;

  if (permissionsLoading || usersLoading) {
    return <p className="text-sm text-app-muted">Carregando permissões...</p>;
  }
  if (permissionsError || usersError) {
    return <p className="text-sm text-red-500">Erro ao carregar permissões.</p>;
  }

  return (
    <div className="space-y-10 animate-fade-in font-sans">
      <div>
        <h2 className="text-2xl font-extrabold tracking-tight text-app-text">Permissões</h2>
        <p className="text-sm text-app-muted">Gerencie regras de autorização granulares e atribuições de acessos aos usuários.</p>
      </div>

      {/* Catalog section */}
      <section className="space-y-4">
        <div className="flex items-center justify-between gap-4">
          <h3 className="text-sm font-semibold uppercase tracking-wider text-app-muted">Catálogo de Permissões</h3>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => userSectionRef.current?.scrollIntoView({ behavior: "smooth" })}
              className="rounded-xl border border-app-border bg-app-bg/40 px-5 py-2.5 text-sm font-bold text-app-text hover:bg-app-card-hover transition"
            >
              Permissões por Usuário
            </button>
            <Can permission="permissions.create">
              <button
                type="button"
                disabled={pending}
                onClick={() => {
                  setCatalogError(null);
                  setCatalogModal("create");
                }}
                className="rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-bold text-white shadow-lg transition hover:bg-indigo-500 disabled:opacity-50"
              >
                Nova permissão
              </button>
            </Can>
          </div>
        </div>

        <div className="overflow-hidden rounded-2xl border border-app-border bg-app-card backdrop-blur-md shadow-xl">
          <table className="w-full text-left text-sm border-collapse">
            <thead>
              <tr className="border-b border-app-border bg-app-bg/40 text-app-muted">
                <th className="px-6 py-4 font-semibold">Código</th>
                <th className="px-6 py-4 font-semibold">Descrição</th>
                <th className="px-6 py-4 font-semibold text-right">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-app-border/60">
              {permissions?.map((permission) => (
                <tr key={permission.id} className="transition-colors hover:bg-app-card-hover/40">
                  <td className="px-6 py-4">
                    <span className="inline-flex rounded-lg bg-app-bg border border-app-border px-3 py-1 font-mono text-xs font-semibold code-pill">
                      {permission.code}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-app-text">{permission.description || <span className="text-app-muted">—</span>}</td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex justify-end gap-2">
                      <Can permission="permissions.update">
                        <button
                          type="button"
                          disabled={pending}
                          onClick={() => {
                            setCatalogError(null);
                            setCatalogModal({ edit: permission });
                          }}
                          className="rounded-lg border border-app-border bg-app-bg/40 px-3 py-1.5 text-xs font-semibold text-app-text hover:bg-app-card-hover disabled:opacity-50"
                        >
                          Editar
                        </button>
                      </Can>
                      <Can permission="permissions.delete">
                        {!PROTECTED_CODES.has(permission.code) && (
                          <button
                            type="button"
                            disabled={pending}
                            onClick={() => setDeleteTarget(permission)}
                            className="rounded-lg border border-red-500/40 px-3 py-1.5 text-xs font-semibold text-red-500 hover:bg-red-500/10 disabled:opacity-50"
                          >
                            Excluir
                          </button>
                        )}
                      </Can>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* User permissions section */}
      <section ref={userSectionRef} className="space-y-5">
        <h3 className="text-sm font-semibold uppercase tracking-wider text-app-muted">Permissões por Usuário</h3>

        <div className="rounded-3xl border border-app-border bg-app-card backdrop-blur-md p-6 max-w-xl">
          <label htmlFor="user" className="block text-xs font-semibold uppercase tracking-wider text-app-muted mb-2">
            Selecionar Usuário
          </label>
          <select
            id="user"
            value={selectedUserId ?? ""}
            onChange={(e) => setSelectedUserId(e.target.value ? Number(e.target.value) : null)}
            className="w-full rounded-xl border border-app-border bg-app-input px-4 py-3 text-sm text-app-text outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20"
          >
            <option value="" className="bg-app-bg text-app-text">Selecione um colaborador...</option>
            {users?.map((user) => (
              <option key={user.id} value={user.id} className="bg-app-bg text-app-text">
                {user.name} ({user.email})
              </option>
            ))}
          </select>
        </div>

        {selectedUser && (
          <div className="rounded-3xl border border-app-border bg-app-card backdrop-blur-md p-6 animate-scale-in">
            <h4 className="text-sm font-bold text-app-text mb-4">
              Acessos concedidos para {selectedUser.name}
            </h4>
            <ul className="grid gap-3 sm:grid-cols-2">
              {permissions?.map((permission) => {
                const granted = grantedIds.has(permission.id);
                return (
                  <li
                    key={permission.id}
                    className="flex items-center justify-between rounded-2xl border border-app-border bg-app-bg/40 p-4 transition-all hover:border-app-border"
                  >
                    <div className="space-y-1">
                      <span className="font-mono text-xs text-app-text font-semibold">{permission.code}</span>
                      <p className="text-xs text-app-muted">{permission.description}</p>
                    </div>
                    {granted ? (
                      <Can permission="permissions.revoke">
                        <button
                          disabled={pending}
                          onClick={() =>
                            setRevokeTarget({ userId: selectedUser.id, permission })
                          }
                          className="rounded-xl border border-red-500/40 px-3.5 py-2 text-xs font-bold text-red-550 dark:text-red-400 hover:bg-red-500/10 transition active:scale-95 disabled:opacity-50"
                        >
                          Revogar
                        </button>
                      </Can>
                    ) : (
                      <Can permission="permissions.grant">
                        <button
                          disabled={pending}
                          onClick={() =>
                            grant.mutate({
                              userId: selectedUser.id,
                              permissionId: permission.id,
                            })
                          }
                          className="rounded-xl bg-indigo-600 px-3.5 py-2 text-xs font-bold text-white hover:bg-indigo-500 transition active:scale-95 disabled:opacity-50"
                        >
                          Conceder
                        </button>
                      </Can>
                    )}
                  </li>
                );
              })}
            </ul>
          </div>
        )}
      </section>

      {/* Catalog modal */}
      {catalogModal && (
        <CatalogModal
          mode={catalogModal}
          pending={createPerm.isPending || updatePerm.isPending}
          error={catalogError}
          onClose={() => {
            setCatalogModal(null);
            setCatalogError(null);
          }}
          onSubmit={(code, description) => {
            if (catalogModal === "create") {
              createPerm.mutate({ code, description });
            } else {
              updatePerm.mutate({ id: catalogModal.edit.id, description });
            }
          }}
        />
      )}

      {/* Delete Confirmation */}
      {deleteTarget && (
        <ConfirmDialog
          title="Excluir permissão"
          message={`Excluir definitivamente "${deleteTarget.code}"? Esta ação só será executada se nenhum usuário possuir esta permissão ativa.`}
          pending={deletePerm.isPending}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={() => deletePerm.mutate(deleteTarget.id)}
        />
      )}

      {/* Revoke Confirmation */}
      {revokeTarget && (
        <ConfirmDialog
          title="Revogar permissão"
          message={`Revogar a credencial "${revokeTarget.permission.code}" de ${selectedUser?.name}?`}
          pending={revoke.isPending}
          onCancel={() => setRevokeTarget(null)}
          onConfirm={() =>
            revoke.mutate({
              userId: revokeTarget.userId,
              permissionId: revokeTarget.permission.id,
            })
          }
        />
      )}
    </div>
  );
}

function CatalogModal({
  mode,
  pending,
  error,
  onClose,
  onSubmit,
}: {
  mode: "create" | { edit: Permission };
  pending: boolean;
  error: string | null;
  onClose: () => void;
  onSubmit: (code: string, description: string) => void;
}) {
  const isCreate = mode === "create";
  const [code, setCode] = useState(isCreate ? "" : mode.edit.code);
  const [description, setDescription] = useState(isCreate ? "" : mode.edit.description);
  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null);
    if (isCreate && !code.trim()) {
      setLocalError("O código é obrigatório.");
      return;
    }
    onSubmit(code.trim(), description.trim());
  };

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-fade-in cursor-pointer"
      onClick={onClose}
    >
      <div 
        className="w-full max-w-md rounded-3xl border border-app-border bg-app-card p-8 shadow-2xl backdrop-blur-xl animate-scale-in cursor-default"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between pb-4 border-b border-app-border/60">
          <h3 className="text-lg font-bold text-app-text font-sans">
            {isCreate ? "Nova Permissão" : "Editar Permissão"}
          </h3>
          <button type="button" onClick={onClose} className="text-app-muted hover:text-app-text transition-colors cursor-pointer">
            ✕
          </button>
        </div>
        
        <form onSubmit={handleSubmit} className="mt-6 space-y-5">
          {isCreate ? (
            <div className="space-y-1.5">
              <label className="block text-xs font-semibold uppercase tracking-wider text-app-muted">Código identificador</label>
              <input
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="ex.: logs.visualizar"
                className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 font-mono text-sm text-app-text outline-none focus:border-indigo-500"
              />
              <p className="text-[10px] text-app-muted/80">Padrão do sistema: recurso.acao (apenas letras minúsculas)</p>
            </div>
          ) : (
            <div className="rounded-xl bg-app-bg/40 p-4 border border-app-border">
              <p className="font-mono text-sm text-indigo-400 dark:text-indigo-300">{code}</p>
            </div>
          )}
          
          <div className="space-y-1.5">
            <label className="block text-xs font-semibold uppercase tracking-wider text-app-muted">Descrição detalhada</label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Descreva o escopo deste acesso..."
              className="w-full rounded-xl border border-app-border bg-app-input px-4 py-2.5 text-sm text-app-text outline-none focus:border-indigo-500"
            />
          </div>
          
          {(localError || error) && (
            <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-xs text-red-500 dark:text-red-400">
              {localError ?? error}
            </div>
          )}
          
          <div className="flex justify-end gap-3 pt-4 border-t border-app-border/60">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-app-border bg-app-bg/40 px-5 py-2.5 text-sm font-semibold text-app-text hover:bg-app-card-hover transition cursor-pointer"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={pending}
              className="rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-bold text-white shadow-lg hover:bg-indigo-500 transition disabled:opacity-50 cursor-pointer"
            >
              {pending ? "Salvando..." : "Salvar"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function ConfirmDialog({
  title,
  message,
  pending,
  onCancel,
  onConfirm,
}: {
  title: string;
  message: string;
  pending: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-fade-in cursor-pointer"
      onClick={onCancel}
    >
      <div 
        className="w-full max-w-md rounded-3xl border border-app-border bg-app-card p-8 shadow-2xl backdrop-blur-xl animate-scale-in cursor-default"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-base font-bold text-app-text">{title}</h3>
        <p className="mt-3 text-sm text-app-muted leading-relaxed">{message}</p>
        <div className="mt-6 flex justify-end gap-3 pt-4 border-t border-app-border/60">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-xl border border-app-border bg-app-bg/40 px-4 py-2 text-sm font-semibold text-app-text hover:bg-app-card-hover transition cursor-pointer"
          >
            Cancelar
          </button>
          <button
            type="button"
            disabled={pending}
            onClick={onConfirm}
            className="rounded-xl bg-indigo-600 px-4 py-2 text-sm font-bold text-white shadow-lg hover:bg-indigo-500 transition active:scale-95 disabled:opacity-50 cursor-pointer"
          >
            Confirmar
          </button>
        </div>
      </div>
    </div>
  );
}
