DELETE FROM user_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN (
        'users.read', 'users.create', 'users.update',
        'users.password.reset', 'users.deactivate',
        'permissions.read', 'permissions.create', 'permissions.update',
        'permissions.delete', 'permissions.grant', 'permissions.revoke'
    )
);

DELETE FROM permissions WHERE code IN (
    'users.read', 'users.create', 'users.update',
    'users.password.reset', 'users.deactivate',
    'permissions.read', 'permissions.create', 'permissions.update',
    'permissions.delete', 'permissions.grant', 'permissions.revoke'
);
