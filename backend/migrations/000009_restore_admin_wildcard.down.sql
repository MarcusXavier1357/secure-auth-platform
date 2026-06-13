-- Rollback manual: revogar `*` concedido por 000009 se necessário.
DELETE FROM user_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE code = '*')
  AND user_id IN (
    SELECT user_id FROM user_permissions
    WHERE permission_id IN (
        SELECT id FROM permissions WHERE code IN (
            'users.read', 'users.create', 'users.update', 'users.password.reset', 'users.deactivate',
            'permissions.read', 'permissions.create', 'permissions.update',
            'permissions.delete', 'permissions.grant', 'permissions.revoke'
        )
    )
    GROUP BY user_id
    HAVING COUNT(DISTINCT permission_id) = 11
  );
