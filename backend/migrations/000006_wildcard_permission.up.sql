INSERT INTO permissions (code, description) VALUES
    ('*', 'Acesso total ao sistema'),
    ('users.*', 'Wildcard: módulo de usuários')
ON CONFLICT (code) DO NOTHING;
