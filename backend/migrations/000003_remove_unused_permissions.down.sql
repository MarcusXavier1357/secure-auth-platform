INSERT INTO permissions (code, description) VALUES
    ('clients.read', 'Visualizar clientes'),
    ('clients.write', 'Criar e editar clientes'),
    ('clients.delete', 'Excluir clientes'),
    ('contracts.read', 'Visualizar contratos'),
    ('contracts.write', 'Criar e editar contratos'),
    ('contracts.cancel', 'Cancelar contratos'),
    ('financial.read', 'Visualizar dados financeiros'),
    ('financial.export', 'Exportar dados financeiros')
ON CONFLICT (code) DO NOTHING;
