-- Partner Device is a new Module for the third-party Partner API
-- integration (see CONTEXT.md#environment, ADR 0011). No 'approve' Action -
-- there is no review/sign-off step for a Serial<->Location mapping.
INSERT INTO permissions (module, action)
SELECT 'partnerdevice', a
FROM unnest(ARRAY['view', 'create', 'edit', 'delete']) AS a;

-- Admin only, by design: environment:view is already broadly granted (Lab
-- Manager/QA/Scientist/General via migration 000020), but Partner Device
-- mapping/data should default to Admin-only until deliberately extended.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'partnerdevice'
WHERE r.name = 'Admin';
