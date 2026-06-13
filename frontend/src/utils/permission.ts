// Espelha matchPermission do backend (internal/service/permission_match.go) —
// controle visual apenas. Manter em sync; testes espelhados.
export function matchPermission(granted: string[], required: string): boolean {
  for (const g of granted) {
    if (g === "*" || g === required) {
      return true;
    }
    if (g.endsWith(".*")) {
      const prefix = g.slice(0, -1);
      if (required.startsWith(prefix)) {
        return true;
      }
    }
  }
  return false;
}
