DELETE FROM audit_logs
WHERE tenant_id IN ('tenant-acme', 'tenant-view', 'tenant-beta');
