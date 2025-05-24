ALTER TABLE notification DROP COLUMN sender_role;
ALTER TABLE notification DROP COLUMN receiver_role;

ALTER TABLE notification DROP COLUMN resume_id;

DROP TYPE user_type;