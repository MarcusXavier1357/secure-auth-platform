DELETE FROM permissions WHERE code IN (
    'clients.read', 'clients.write', 'clients.delete',
    'contracts.read', 'contracts.write', 'contracts.cancel',
    'financial.read', 'financial.export',
    'users.manage', 'permissions.manage', 'audit_logs.read'
);

DELETE FROM roles WHERE name IN ('Admin', 'Supervisor', 'Operador', 'Financeiro', 'Vendedor');
