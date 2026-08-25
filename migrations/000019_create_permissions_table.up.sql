CREATE TABLE permissions (
    id         BIGSERIAL PRIMARY KEY,
    module     VARCHAR(30) NOT NULL,
    action     VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (module, action)
);

-- 11 Modules x 5 Actions = 55 Permissions - see CONTEXT.md#access-control
-- and ADR 0002.
INSERT INTO permissions (module, action)
SELECT m, a
FROM unnest(ARRAY[
    'user', 'sample', 'testresult', 'location', 'equipment', 'inventory',
    'purchaseorder', 'document', 'environment', 'notification', 'audit'
]) AS m
CROSS JOIN unnest(ARRAY['view', 'create', 'edit', 'delete', 'approve']) AS a;

-- audit:export is a 12th, module-specific Permission: distinct from
-- audit:view (Admin and QA can pull the PDF compliance export, Lab
-- Manager can only browse the JSON audit trail) - see CONTEXT.md#access-control.
INSERT INTO permissions (module, action) VALUES ('audit', 'export');
