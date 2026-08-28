-- Vendor is a new Module (see task.md Phase 1 / CONTEXT.md#vendors). The
-- five Actions apply uniformly across every Module - see
-- CONTEXT.md#access-control - so all five Permission rows are created even
-- though only view/create/edit/delete are granted to any Role today
-- (approve on a Vendor is meaningless and stays ungranted; delete has no
-- use case yet but the Permission is in place).
INSERT INTO permissions (module, action)
SELECT 'vendor', a
FROM unnest(ARRAY['view', 'create', 'edit', 'delete', 'approve']) AS a;

-- Admin: every Action (the 000020 seed cross-join ran before this Module
-- existed, so grant explicitly).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'vendor'
WHERE r.name = 'Admin';

-- Lab Manager: full CRUD on this operational master-data Module.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'vendor' AND p.action IN ('view', 'create', 'edit', 'delete')
WHERE r.name = 'Lab Manager';

-- Scientist: view/create/edit (may quick-add a Vendor from an
-- Equipment/Inventory form), no delete.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'vendor' AND p.action IN ('view', 'create', 'edit')
WHERE r.name = 'Scientist';

-- QA and General: view only.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'vendor' AND p.action = 'view'
WHERE r.name IN ('QA', 'General');
