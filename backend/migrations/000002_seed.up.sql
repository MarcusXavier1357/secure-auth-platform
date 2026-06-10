INSERT INTO roles (name, description) VALUES
    ('Admin', 'Administrador do sistema'),
    ('Supervisor', 'Supervisor de equipe'),
    ('Operador', 'Operador padrão'),
    ('Financeiro', 'Equipe financeira'),
    ('Vendedor', 'Equipe de vendas');

INSERT INTO permissions (code, description) VALUES
    ('users.manage', 'Gerenciar usuários'),
    ('permissions.manage', 'Gerenciar permissões'),
    ('audit_logs.read', 'Visualizar logs de auditoria');
