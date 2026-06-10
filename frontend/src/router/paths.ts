// Fonte única dos paths da aplicação. Sempre use paths.x() em links e
// navegação — nunca strings soltas. Funções para acomodar rotas dinâmicas
// futuras (ex.: users(id) => `/users/${id}`).
export const paths = {
  login: () => "/login",
  home: () => "/",
  users: () => "/users",
  permissions: () => "/permissions",
  audit: () => "/audit",
} as const;
