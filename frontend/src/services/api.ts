const BASE_URL = "/api";

// Access token mantido apenas em memória — nunca em localStorage.
let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken() {
  return accessToken;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

async function rawRequest(path: string, options: RequestInit = {}): Promise<Response> {
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`);
  }
  return fetch(BASE_URL + path, { ...options, headers, credentials: "include" });
}

// Renova o access token via cookie HttpOnly. Retorna false se a sessão expirou.
export async function tryRefresh(): Promise<boolean> {
  const res = await fetch(BASE_URL + "/auth/refresh", {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) {
    setAccessToken(null);
    return false;
  }
  const data = await res.json();
  setAccessToken(data.accessToken);
  return true;
}

// request faz a chamada e, em 401, tenta renovar a sessão uma vez antes de falhar.
export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  let res = await rawRequest(path, options);

  if (res.status === 401 && (await tryRefresh())) {
    res = await rawRequest(path, options);
  }

  if (!res.ok) {
    const body = await res.text();
    throw new ApiError(res.status, body || res.statusText);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return res.json();
}

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  roleId?: number | null;
}

export interface UpdateUserInput {
  name?: string;
  email?: string;
  password?: string;
  roleId?: number | null;
  active?: boolean;
}

export interface CreatePermissionInput {
  code: string;
  description: string;
}

export interface UpdatePermissionInput {
  description: string;
}

export const api = {
  login: (email: string, password: string) =>
    request<{ accessToken: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  logout: () => request<void>("/auth/logout", { method: "POST" }),

  me: () => request<MeResponse>("/me"),

  users: {
    list: () => request<UserWithPermissions[]>("/users"),
    get: (id: number) => request<UserWithPermissions>(`/users/${id}`),
    create: (input: CreateUserInput) =>
      request<User>("/users", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    update: (id: number, input: UpdateUserInput) =>
      request<User>(`/users/${id}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
  },

  roles: {
    list: () => request<Role[]>("/roles"),
  },

  permissions: {
    list: () => request<Permission[]>("/permissions"),
    create: (input: CreatePermissionInput) =>
      request<Permission>("/permissions", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    update: (id: number, input: UpdatePermissionInput) =>
      request<Permission>(`/permissions/${id}`, {
        method: "PATCH",
        body: JSON.stringify(input),
      }),
    delete: (id: number) =>
      request<void>(`/permissions/${id}`, {
        method: "DELETE",
      }),
    grant: (userId: number, permissionId: number) =>
      request<void>(`/users/${userId}/permissions`, {
        method: "POST",
        body: JSON.stringify({ permissionId }),
      }),
    revoke: (userId: number, permissionId: number) =>
      request<void>(`/users/${userId}/permissions/${permissionId}`, {
        method: "DELETE",
      }),
  },

  audit: {
    list: (limit = 50, offset = 0) =>
      request<AuditLog[]>(`/audit-logs?limit=${limit}&offset=${offset}`),
  },
};

export interface Role {
  id: number;
  name: string;
  description: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  active: boolean;
  roleId: number | null;
  role?: { id: number; name: string };
}

export interface UserWithPermissions extends User {
  permissions?: Permission[];
}

export interface Permission {
  id: number;
  code: string;
  description: string;
}

export interface AuditLog {
  id: number;
  userId: number | null;
  action: string;
  entity: string;
  entityId: number | null;
  oldData: Record<string, unknown> | null;
  newData: Record<string, unknown> | null;
  createdAt: string;
}

export interface MeResponse {
  user: UserWithPermissions;
  permissions: string[];
}
