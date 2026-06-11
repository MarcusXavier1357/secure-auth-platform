// Espelha matchPermission do backend — controle visual apenas.
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
