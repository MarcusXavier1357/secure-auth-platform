DELETE FROM user_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('users.manage', 'permissions.manage')
);

DELETE FROM permissions WHERE code IN ('users.manage', 'permissions.manage');
