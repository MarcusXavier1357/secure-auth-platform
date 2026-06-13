import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../../../services/api";

export default function PermissionsPage() {
  const queryClient = useQueryClient();
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);

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

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["users"] });

  const grant = useMutation({
    mutationFn: ({ userId, permissionId }: { userId: number; permissionId: number }) =>
      api.permissions.grant(userId, permissionId),
    onSuccess: invalidate,
  });
  const revoke = useMutation({
    mutationFn: ({ userId, permissionId }: { userId: number; permissionId: number }) =>
      api.permissions.revoke(userId, permissionId),
    onSuccess: invalidate,
  });

  const pending = grant.isPending || revoke.isPending;

  if (permissionsLoading || usersLoading) {
    return <p className="text-sm text-slate-400">Carregando permissões...</p>;
  }
  if (permissionsError || usersError) {
    return <p className="text-sm text-red-400">Erro ao carregar permissões.</p>;
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Permissões</h2>

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
          <h3 className="mb-4 text-sm font-medium text-slate-300">
            Permissões de {selectedUser.name}
          </h3>
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
                  <button
                    disabled={pending}
                    onClick={() =>
                      (granted ? revoke : grant).mutate({
                        userId: selectedUser.id,
                        permissionId: permission.id,
                      })
                    }
                    className={`rounded-lg px-3 py-1.5 text-xs font-semibold transition disabled:opacity-50 ${
                      granted
                        ? "border border-red-900/60 text-red-400 hover:bg-red-950/40"
                        : "bg-indigo-600 text-white hover:bg-indigo-500"
                    }`}
                  >
                    {granted ? "Revogar" : "Conceder"}
                  </button>
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </div>
  );
}
