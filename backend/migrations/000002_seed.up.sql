INSERT INTO roles (name, description) VALUES
    ('Admin', 'Administrador do sistema'),
    ('Supervisor', 'Supervisor de equipe'),
    ('Operador', 'Operador padrão'),
    ('Financeiro', 'Equipe financeira'),
    ('Vendedor', 'Equipe de vendas');

INSERT INTO permissions (code, description) VALUES
    ('clients.read', 'Visualizar clientes'),
    ('clients.write', 'Criar e editar clientes'),
    ('clients.delete', 'Excluir clientes'),
    ('contracts.read', 'Visualizar contratos'),
    ('contracts.write', 'Criar e editar contratos'),
    ('contracts.cancel', 'Cancelar contratos'),
    ('financial.read', 'Visualizar dados financeiros'),
    ('financial.export', 'Exportar dados financeiros'),
    ('users.manage', 'Gerenciar usuários'),
    ('permissions.manage', 'Gerenciar permissões'),
    ('audit_logs.read', 'Visualizar logs de auditoria');
