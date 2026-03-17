INSERT INTO memberships (user_id, tenant_id, role) VALUES
    ('123', 'tenant-acme', 'admin'),
    ('123', 'tenant-view', 'viewer'),
    ('123', 'tenant-beta', 'admin')
ON CONFLICT (user_id, tenant_id) DO UPDATE SET role = EXCLUDED.role;
