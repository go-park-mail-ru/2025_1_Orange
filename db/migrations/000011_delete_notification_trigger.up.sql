ALTER TABLE notification
ADD CONSTRAINT fk_notification_resume
FOREIGN KEY (resume_id) REFERENCES resume(id) ON DELETE CASCADE;
