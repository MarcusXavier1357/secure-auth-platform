-- Remove permissões do plano futuro (clients/contracts/financial) sem rotas implementadas.
DELETE FROM permissions WHERE code IN (
    'clients.read', 'clients.write', 'clients.delete',
    'contracts.read', 'contracts.write', 'contracts.cancel',
    'financial.read', 'financial.export'
);
