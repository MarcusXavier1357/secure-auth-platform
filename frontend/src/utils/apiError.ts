export function parseApiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) {
    const msg = error.message;
    const map: Record<string, string> = {
      "email already in use": "Este e-mail já está em uso.",
      "cannot deactivate your own account": "Você não pode desativar a própria conta.",
      "cannot deactivate the last admin": "Não é possível desativar o último administrador.",
      "permission code already exists": "Este código de permissão já existe.",
      "invalid permission code": "Código de permissão inválido.",
      "protected permission": "Esta permissão é protegida e não pode ser removida.",
      "permission in use": "Remova a permissão de todos os usuários antes de excluí-la.",
      "cannot grant permission you do not have": "Você não pode conceder uma permissão que não possui.",
      "missing permission for this update": "Você não tem permissão para esta alteração.",
    };
    for (const [key, pt] of Object.entries(map)) {
      if (msg.includes(key)) return pt;
    }
    return msg;
  }
  return fallback;
}
