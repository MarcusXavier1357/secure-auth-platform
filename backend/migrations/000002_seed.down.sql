DELETE FROM permissions WHERE code IN (
    'users.manage', 'permissions.manage', 'audit_logs.read'
);

DELETE FROM roles WHERE name IN ('Admin', 'Supervisor', 'Operador', 'Financeiro', 'Vendedor');
