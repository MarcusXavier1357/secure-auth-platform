import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { api, type UserWithPermissions } from "../../../services/api";
import { Can } from "../../../components/Can";
import { UserFormModal } from "../../../components/UserFormModal";
import { parseApiErrorMessage } from "../../../utils/apiError";
import { useToast } from "../../../hooks/useToast";

export default function UsersPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
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
      toast.success("Usuário criado com sucesso!");
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
      toast.success("Usuário atualizado com sucesso!");
    },
    onError: (err) => setFormError(parseApiErrorMessage(err, "Erro ao atualizar usuário.")),
  });

  const toggleActiveMutation = useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) =>
      api.users.update(id, { active }),
    onSuccess: () => {
      invalidate();
      toast.success(
        toggleTarget?.active ? "Usuário desativado com sucesso!" : "Usuário ativado com sucesso!",
      );
      setToggleTarget(null);
    },
    onError: (err) => {
      toast.error(parseApiErrorMessage(err, "Erro ao alterar status do usuário."));
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
    <div className="space-y-8 animate-fade-in font-sans">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h2 className="text-2xl font-extrabold tracking-tight text-app-text">Usuários</h2>
          <p className="text-sm text-app-muted">
            Gerencie contas, papéis e status de acesso dos colaboradores.
          </p>
        </div>
        <Can permission="users.create">
          <button
            type="button"
            onClick={openCreate}
            disabled={pending}
            className="rounded-xl bg-indigo-600 px-5 py-2.5 text-sm font-bold text-white shadow-lg transition-all hover:bg-indigo-500 active:scale-95 disabled:opacity-50"
          >
            Novo usuário
          </button>
        </Can>
      </div>

      <div className="overflow-hidden rounded-2xl border border-app-border bg-app-card backdrop-blur-md shadow-xl">
        <table className="w-full text-left text-sm border-collapse">
          <thead>
            <tr className="border-b border-app-border bg-app-bg/40 text-app-muted">
              <th className="px-6 py-4 font-semibold">Nome</th>
              <th className="px-6 py-4 font-semibold">E-mail</th>
              <th className="px-6 py-4 font-semibold">Perfil (Role)</th>
              <th className="px-6 py-4 font-semibold text-center">Permissões</th>
              <th className="px-6 py-4 font-semibold">Status</th>
              <th className="px-6 py-4 font-semibold text-right">Ações</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-app-border/60">
            {users?.map((user) => (
              <tr key={user.id} className="transition-colors hover:bg-app-card-hover/40">
                <td className="px-6 py-4 font-semibold text-app-text">{user.name}</td>
                <td className="px-6 py-4 text-app-muted">{user.email}</td>
                <td className="px-6 py-4">
                  {user.role ? (
                    <span className="inline-flex rounded-lg bg-indigo-500/10 border border-indigo-500/20 px-2 py-0.5 text-xs text-indigo-550 dark:text-indigo-300 font-medium">
                      {user.role.name}
                    </span>
                  ) : (
                    <span className="text-app-muted">—</span>
                  )}
                </td>
                <td className="px-6 py-4 text-center">
                  <span className="inline-flex items-center justify-center rounded-lg bg-app-bg border border-app-border px-2 py-0.5 font-mono text-xs text-app-text">
                    {user.permissions?.length ?? 0}
                  </span>
                </td>
                <td className="px-6 py-4">
                  {user.active ? (
                    <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 px-2.5 py-0.5 text-xs font-medium text-emerald-600 dark:text-emerald-400">
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                      Ativo
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1.5 rounded-full bg-app-bg border border-app-border px-2.5 py-0.5 text-xs font-medium text-app-muted">
                      <span className="h-1.5 w-1.5 rounded-full bg-slate-500" />
                      Inativo
                    </span>
                  )}
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex justify-end gap-2">
                    <Can permission="users.update">
                      <button
                        type="button"
                        disabled={pending}
                        onClick={() => openEdit(user)}
                        className="rounded-lg border border-app-border bg-app-bg/40 px-3 py-1.5 text-xs font-semibold text-app-text transition hover:bg-app-card-hover disabled:opacity-50"
                      >
                        Editar
                      </button>
                    </Can>
                    <Can permission="users.deactivate">
                      <button
                        type="button"
                        disabled={pending}
                        onClick={() => setToggleTarget(user)}
                        className={`rounded-lg border px-3 py-1.5 text-xs font-semibold transition disabled:opacity-50 ${
                          user.active
                            ? "border-red-500/40 text-red-500 hover:bg-red-500/10"
                            : "border-emerald-500/40 text-emerald-500 hover:bg-emerald-500/10"
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
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-fade-in cursor-pointer"
          onClick={() => setToggleTarget(null)}
        >
          <div
            className="w-full max-w-md rounded-3xl border border-app-border bg-app-card p-8 shadow-2xl backdrop-blur-xl animate-scale-in cursor-default"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-base font-bold text-app-text">
              {toggleTarget.active ? "Desativar usuário" : "Ativar usuário"}
            </h3>
            <p className="mt-3 text-sm text-app-muted leading-relaxed">
              {toggleTarget.active
                ? `Você tem certeza que deseja desativar a conta de ${toggleTarget.name}? Esta ação revogará imediatamente todas as sessões ativas do usuário.`
                : `Deseja reativar o acesso de ${toggleTarget.name} ao sistema?`}
            </p>
            <div className="mt-6 flex justify-end gap-3 pt-4 border-t border-app-border">
              <button
                type="button"
                onClick={() => setToggleTarget(null)}
                className="rounded-xl border border-app-border bg-app-bg/60 px-4 py-2 text-sm font-semibold text-app-text hover:bg-app-card-hover transition cursor-pointer"
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
                className="rounded-xl bg-indigo-600 px-4 py-2 text-sm font-bold text-white shadow-lg hover:bg-indigo-500 transition active:scale-95 disabled:opacity-50 cursor-pointer"
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
