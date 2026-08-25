CREATE TABLE role_permissions (
    role_id       BIGINT NOT NULL REFERENCES roles (id),
    permission_id BIGINT NOT NULL REFERENCES permissions (id),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);

-- Seed the Role -> Permission grants per the matrix in CONTEXT.md#access-control.

-- Admin: every Action on every Module.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Admin';

-- Lab Manager: full CRUD+approve on every operational Module...
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module IN (
    'sample', 'testresult', 'location', 'equipment', 'inventory',
    'purchaseorder', 'document', 'environment', 'notification'
)
WHERE r.name = 'Lab Manager';

-- ...plus user:view only...
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'user' AND p.action = 'view'
WHERE r.name = 'Lab Manager';

-- ...plus audit:view only (no audit:delete/edit/create/approve).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'audit' AND p.action = 'view'
WHERE r.name = 'Lab Manager';

-- QA: view on every Module...
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.action = 'view'
WHERE r.name = 'QA';

-- ...approve on sample/testresult/equipment/purchaseorder...
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module IN ('sample', 'testresult', 'equipment', 'purchaseorder') AND p.action = 'approve'
WHERE r.name = 'QA';

-- ...and edit on document.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'document' AND p.action = 'edit'
WHERE r.name = 'QA';

-- ...plus audit:export (Admin already has it via the full cross join above;
-- Lab Manager deliberately does not - see the audit:view grant above).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'audit' AND p.action = 'export'
WHERE r.name = 'QA';

-- Scientist: view/create/edit (no delete/approve) on the core lab Modules...
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module IN ('sample', 'testresult', 'location', 'equipment', 'inventory')
    AND p.action IN ('view', 'create', 'edit')
WHERE r.name = 'Scientist';

-- ...view-only on the rest.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module IN ('purchaseorder', 'document', 'environment', 'notification', 'user')
    AND p.action = 'view'
WHERE r.name = 'Scientist';

-- General: view-only on every operational Module (not user or audit).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module IN (
    'sample', 'testresult', 'location', 'equipment', 'inventory',
    'purchaseorder', 'document', 'environment', 'notification'
) AND p.action = 'view'
WHERE r.name = 'General';
