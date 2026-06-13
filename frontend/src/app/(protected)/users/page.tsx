import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { api, type UserWithPermissions } from "../../../services/api";
import { Can } from "../../../components/Can";
import { UserFormModal } from "../../../components/UserFormModal";
import { parseApiErrorMessage } from "../../../utils/apiError";

export default function UsersPage() {
  const queryClient = useQueryClient();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserWithPermissions | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [toggleTarget, setToggleTarget] = useState<UserWithPermissions | null>(null);

  const {
    data: users,
    isLoading,
    isError,
  } = useQuery({ queryKey: ["users"], queryFn: api.users.list });

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["users"] });

  const createMutation = useMutation({
    mutationFn: api.users.create,
    onSuccess: () => {
      invalidate();
      setModalOpen(false);
      setFormError(null);
    },
    onError: (err) => setFormError(parseApiErrorMessage(err, "Erro ao criar usuário.")),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: number; body: Parameters<typeof api.users.update>[1] }) =>
      api.users.update(id, body),
    onSuccess: () => {
      invalidate();
      setModalOpen(false);
      setEditingUser(null);
      setFormError(null);
    },
    onError: (err) => setFormError(parseApiErrorMessage(err, "Erro ao atualizar usuário.")),
  });

  const toggleActiveMutation = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) =>
      api.users.update(id, { active }),
    onSuccess: () => {
      invalidate();
      setToggleTarget(null);
    },
    onError: (err) => {
      alert(parseApiErrorMessage(err, "Erro ao alterar status do usuário."));
      setToggleTarget(null);
    },
  });

  const openCreate = () => {
    setEditingUser(null);
    setFormError(null);
    setModalOpen(true);
  };

  const openEdit = (user: UserWithPermissions) => {
    setEditingUser(user);
    setFormError(null);
    setModalOpen(true);
  };

  const pending =
    createMutation.isPending || updateMutation.isPending || toggleActiveMutation.isPending;

  if (isLoading) {
    return <p className="text-sm text-slate-400">Carregando usuários...</p>;
  }
  if (isError) {
    return <p className="text-sm text-red-400">Erro ao carregar usuários.</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-xl font-semibold">Usuários</h2>
        <Can permission="users.create">
          <button
            type="button"
            onClick={openCreate}
            disabled={pending}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
          >
            Novo usuário
          </button>
        </Can>
      </div>

      <div className="overflow-hidden rounded-2xl border border-slate-800">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-900 text-slate-400">
            <tr>
              <th className="px-4 py-3 font-medium">Nome</th>
              <th className="px-4 py-3 font-medium">Email</th>
              <th className="px-4 py-3 font-medium">Role</th>
              <th className="px-4 py-3 font-medium">Permissões</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Ações</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800 bg-slate-900/50">
            {users?.map((user) => (
              <tr key={user.id}>
                <td className="px-4 py-3">{user.name}</td>
                <td className="px-4 py-3 text-slate-400">{user.email}</td>
                <td className="px-4 py-3 text-slate-400">{user.role?.name ?? "—"}</td>
                <td className="px-4 py-3 text-slate-400">{user.permissions?.length ?? 0}</td>
                <td className="px-4 py-3">
                  {user.active ? (
                    <span className="rounded-full bg-emerald-950 px-2 py-0.5 text-xs text-emerald-400">
                      Ativo
                    </span>
                  ) : (
                    <span className="rounded-full bg-red-950 px-2 py-0.5 text-xs text-red-400">
                      Inativo
                    </span>
                  )}
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    <Can permission="users.update">
                      <button
                        type="button"
                        disabled={pending}
                        onClick={() => openEdit(user)}
                        className="rounded-lg border border-slate-700 px-2.5 py-1 text-xs text-slate-300 transition hover:bg-slate-800 disabled:opacity-50"
                      >
                        Editar
                      </button>
                    </Can>
                    <Can permission="users.deactivate">
                      <button
                        type="button"
                        disabled={pending}
                        onClick={() => setToggleTarget(user)}
                        className={`rounded-lg border px-2.5 py-1 text-xs transition disabled:opacity-50 ${
                          user.active
                            ? "border-red-900/60 text-red-400 hover:bg-red-950/40"
                            : "border-emerald-900/60 text-emerald-400 hover:bg-emerald-950/40"
                        }`}
                      >
                        {user.active ? "Desativar" : "Ativar"}
                      </button>
                    </Can>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {modalOpen && (
        <UserFormModal
          key={editingUser?.id ?? "create"}
          user={editingUser}
          pending={createMutation.isPending || updateMutation.isPending}
          error={formError}
          onClose={() => {
            setModalOpen(false);
            setEditingUser(null);
            setFormError(null);
          }}
          onSubmit={(data) => {
            if (editingUser) {
              const body: Parameters<typeof api.users.update>[1] = {
                name: data.name,
                email: data.email,
                roleId: data.roleId,
              };
              if (data.password) {
                body.password = data.password;
              }
              updateMutation.mutate({ id: editingUser.id, body });
            } else {
              if (!data.password) {
                setFormError("Senha é obrigatória para novo usuário.");
                return;
              }
              createMutation.mutate({
                name: data.name,
                email: data.email,
                password: data.password,
                roleId: data.roleId,
              });
            }
          }}
        />
      )}

      {toggleTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-2xl border border-slate-800 bg-slate-900 p-6">
            <h3 className="text-base font-semibold">
              {toggleTarget.active ? "Desativar usuário" : "Ativar usuário"}
            </h3>
            <p className="mt-2 text-sm text-slate-400">
              {toggleTarget.active
                ? `Desativar ${toggleTarget.name}? Todas as sessões serão revogadas.`
                : `Reativar ${toggleTarget.name}?`}
            </p>
            <div className="mt-6 flex justify-end gap-3">
              <button
                type="button"
                onClick={() => setToggleTarget(null)}
                className="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300"
              >
                Cancelar
              </button>
              <button
                type="button"
                disabled={toggleActiveMutation.isPending}
                onClick={() =>
                  toggleActiveMutation.mutate({
                    id: toggleTarget.id,
                    active: !toggleTarget.active,
                  })
                }
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
              >
                Confirmar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
