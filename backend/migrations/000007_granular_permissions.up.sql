INSERT INTO permissions (code, description) VALUES
    ('users.read', 'Listar usuários'),
    ('users.create', 'Criar usuários'),
    ('users.update', 'Editar nome, email e role de usuários'),
    ('users.password.reset', 'Redefinir senha de usuários'),
    ('users.deactivate', 'Ativar ou desativar usuários'),
    ('permissions.read', 'Listar permissões'),
    ('permissions.create', 'Criar permissões no catálogo'),
    ('permissions.update', 'Editar descrição de permissões'),
    ('permissions.delete', 'Remover permissões do catálogo'),
    ('permissions.grant', 'Conceder permissões a usuários'),
    ('permissions.revoke', 'Revogar permissões de usuários')
ON CONFLICT (code) DO NOTHING;

-- Quem tem users.manage recebe as permissões granulares de usuários.
INSERT INTO user_permissions (user_id, permission_id)
SELECT up.user_id, np.id
FROM user_permissions up
JOIN permissions p ON p.id = up.permission_id AND p.code = 'users.manage'
CROSS JOIN permissions np
WHERE np.code IN (
    'users.read', 'users.create', 'users.update',
    'users.password.reset', 'users.deactivate'
)
ON CONFLICT DO NOTHING;

-- Quem tem permissions.manage recebe as permissões granulares de permissões.
INSERT INTO user_permissions (user_id, permission_id)
SELECT up.user_id, np.id
FROM user_permissions up
JOIN permissions p ON p.id = up.permission_id AND p.code = 'permissions.manage'
CROSS JOIN permissions np
WHERE np.code IN (
    'permissions.read', 'permissions.create', 'permissions.update',
    'permissions.delete', 'permissions.grant', 'permissions.revoke'
)
ON CONFLICT DO NOTHING;
