-- Chatbot is a new Module for the AI chatbot POC (see
-- docs/chatbot-poc-plan.md). It is read-only single-turn Q&A over the POC
-- Oracle ADB, so only the 'view' Action is meaningful - the other four
-- Actions are not created for this Module.
INSERT INTO permissions (module, action)
SELECT 'chatbot', 'view';

-- Grant chatbot:view to every staff Role - the chatbot only reads synthetic
-- demo data and enforces its own SELECT-only guardrails.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.module = 'chatbot' AND p.action = 'view'
WHERE r.name IN ('Admin', 'Lab Manager', 'Scientist', 'QA', 'General');
