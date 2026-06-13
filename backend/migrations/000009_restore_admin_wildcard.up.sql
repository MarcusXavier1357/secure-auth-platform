-- Após 000008, admins seedados antes de 000006 podem ter granulares sem `*`.
-- Concede `*` a usuários que já possuem todos os codes granulares de usuários e permissões.
INSERT INTO user_permissions (user_id, permission_id)
SELECT up.user_id, star.id
FROM (
    SELECT user_id
    FROM user_permissions
    WHERE permission_id IN (
        SELECT id FROM permissions WHERE code IN (
            'users.read', 'users.create', 'users.update', 'users.password.reset', 'users.deactivate',
            'permissions.read', 'permissions.create', 'permissions.update',
            'permissions.delete', 'permissions.grant', 'permissions.revoke'
        )
    )
    GROUP BY user_id
    HAVING COUNT(DISTINCT permission_id) = 11
) AS up
CROSS JOIN permissions AS star
WHERE star.code = '*'
ON CONFLICT DO NOTHING;
