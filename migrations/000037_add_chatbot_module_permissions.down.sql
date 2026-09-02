DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE module = 'chatbot');

DELETE FROM permissions WHERE module = 'chatbot';
