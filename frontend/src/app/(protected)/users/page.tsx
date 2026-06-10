import { useQuery } from "@tanstack/react-query";
import { api } from "../../../services/api";

export default function UsersPage() {
  const {
    data: users,
    isLoading,
    isError,
  } = useQuery({ queryKey: ["users"], queryFn: api.users.list });

  if (isLoading) {
    return <p className="text-sm text-slate-400">Carregando usuários...</p>;
  }
  if (isError) {
    return <p className="text-sm text-red-400">Erro ao carregar usuários.</p>;
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Usuários</h2>

      <div className="overflow-hidden rounded-2xl border border-slate-800">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-900 text-slate-400">
            <tr>
              <th className="px-4 py-3 font-medium">Nome</th>
              <th className="px-4 py-3 font-medium">Email</th>
              <th className="px-4 py-3 font-medium">Role</th>
              <th className="px-4 py-3 font-medium">Permissões</th>
              <th className="px-4 py-3 font-medium">Status</th>
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
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
