import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type Permission } from "../../../services/api";
import { Can } from "../../../components/Can";
import { parseApiErrorMessage } from "../../../utils/apiError";

const PROTECTED_CODES = new Set(["*", "users.*", "audit_logs.read"]);

export default function PermissionsPage() {
  const queryClient = useQueryClient();
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
    return <p className="text-sm text-slate-400">Carregando permissões...</p>;
  }
  if (permissionsError || usersError) {
    return <p className="text-sm text-red-400">Erro ao carregar permissões.</p>;
  }

  return (
    <div className="space-y-8">
      <h2 className="text-xl font-semibold">Permissões</h2>

      <section className="space-y-4">
        <div className="flex items-center justify-between gap-4">
          <h3 className="text-base font-medium text-slate-200">Catálogo</h3>
          <Can permission="permissions.create">
            <button
              type="button"
              disabled={pending}
              onClick={() => {
                setCatalogError(null);
                setCatalogModal("create");
              }}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
            >
              Nova permissão
            </button>
          </Can>
        </div>

        <div className="overflow-hidden rounded-2xl border border-slate-800">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-900 text-slate-400">
              <tr>
                <th className="px-4 py-3 font-medium">Código</th>
                <th className="px-4 py-3 font-medium">Descrição</th>
                <th className="px-4 py-3 font-medium">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800 bg-slate-900/50">
              {permissions?.map((permission) => (
                <tr key={permission.id}>
                  <td className="px-4 py-3 font-mono text-xs">{permission.code}</td>
                  <td className="px-4 py-3 text-slate-400">{permission.description || "—"}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <Can permission="permissions.update">
                        <button
                          type="button"
                          disabled={pending}
                          onClick={() => {
                            setCatalogError(null);
                            setCatalogModal({ edit: permission });
                          }}
                          className="rounded-lg border border-slate-700 px-2.5 py-1 text-xs text-slate-300 hover:bg-slate-800 disabled:opacity-50"
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
                            className="rounded-lg border border-red-900/60 px-2.5 py-1 text-xs text-red-400 hover:bg-red-950/40 disabled:opacity-50"
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

      <section className="space-y-4">
        <h3 className="text-base font-medium text-slate-200">Por usuário</h3>

        <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
          <label htmlFor="user" className="mb-1.5 block text-sm font-medium text-slate-300">
            Usuário
          </label>
          <select
            id="user"
            value={selectedUserId ?? ""}
            onChange={(e) => setSelectedUserId(e.target.value ? Number(e.target.value) : null)}
            className="w-full max-w-sm rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
          >
            <option value="">Selecione um usuário...</option>
            {users?.map((user) => (
              <option key={user.id} value={user.id}>
                {user.name} ({user.email})
              </option>
            ))}
          </select>
        </div>

        {selectedUser && (
          <div className="rounded-2xl border border-slate-800 bg-slate-900 p-6">
            <h4 className="mb-4 text-sm font-medium text-slate-300">
              Permissões de {selectedUser.name}
            </h4>
            <ul className="space-y-2">
              {permissions?.map((permission) => {
                const granted = grantedIds.has(permission.id);
                return (
                  <li
                    key={permission.id}
                    className="flex items-center justify-between rounded-lg border border-slate-800 bg-slate-950/50 px-4 py-2.5"
                  >
                    <div>
                      <span className="font-mono text-sm">{permission.code}</span>
                      <p className="text-xs text-slate-500">{permission.description}</p>
                    </div>
                    {granted ? (
                      <Can permission="permissions.revoke">
                        <button
                          disabled={pending}
                          onClick={() =>
                            setRevokeTarget({ userId: selectedUser.id, permission })
                          }
                          className="rounded-lg border border-red-900/60 px-3 py-1.5 text-xs font-semibold text-red-400 transition hover:bg-red-950/40 disabled:opacity-50"
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
                          className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
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

      {deleteTarget && (
        <ConfirmDialog
          title="Excluir permissão"
          message={`Excluir "${deleteTarget.code}"? Só é possível se nenhum usuário a possuir.`}
          pending={deletePerm.isPending}
          onCancel={() => setDeleteTarget(null)}
          onConfirm={() => deletePerm.mutate(deleteTarget.id)}
        />
      )}

      {revokeTarget && (
        <ConfirmDialog
          title="Revogar permissão"
          message={`Revogar "${revokeTarget.permission.code}" de ${selectedUser?.name}?`}
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
      setLocalError("Código é obrigatório.");
      return;
    }
    onSubmit(code.trim(), description.trim());
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-800 bg-slate-900 p-6">
        <h3 className="text-lg font-semibold">
          {isCreate ? "Nova permissão" : "Editar descrição"}
        </h3>
        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          {isCreate ? (
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-300">Código</label>
              <input
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="ex.: reports.read"
                className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 font-mono text-sm text-white outline-none focus:border-indigo-500"
              />
              <p className="mt-1 text-xs text-slate-500">Formato: recurso.acao (minúsculas)</p>
            </div>
          ) : (
            <p className="font-mono text-sm text-slate-400">{code}</p>
          )}
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-300">Descrição</label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
            />
          </div>
          {(localError || error) && <p className="text-sm text-red-400">{localError ?? error}</p>}
          <div className="flex justify-end gap-3">
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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-800 bg-slate-900 p-6">
        <h3 className="text-base font-semibold">{title}</h3>
        <p className="mt-2 text-sm text-slate-400">{message}</p>
        <div className="mt-6 flex justify-end gap-3">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300"
          >
            Cancelar
          </button>
          <button
            type="button"
            disabled={pending}
            onClick={onConfirm}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
          >
            Confirmar
          </button>
        </div>
      </div>
    </div>
  );
}
