package service

import "strings"

// matchPermission verifica se alguma permissão concedida cobre a requerida.
// Suporta match exato, wildcard de módulo (users.*) e wildcard global (*).
// Manter em sync com frontend/src/utils/permission.ts — testes espelhados.
func matchPermission(granted []string, required string) bool {
	for _, g := range granted {
		if g == "*" || g == required {
			return true
		}
		if strings.HasSuffix(g, ".*") {
			prefix := strings.TrimSuffix(g, ".*") + "."
			if strings.HasPrefix(required, prefix) {
				return true
			}
		}
	}
	return false
}
