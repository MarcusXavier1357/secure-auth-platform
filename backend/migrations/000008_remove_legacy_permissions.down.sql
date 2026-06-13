INSERT INTO permissions (code, description) VALUES
    ('users.manage', 'Gerenciar usuários'),
    ('permissions.manage', 'Gerenciar permissões')
ON CONFLICT (code) DO NOTHING;
