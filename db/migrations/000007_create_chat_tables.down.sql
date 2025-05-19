DROP TRIGGER IF EXISTS trigger_update_chat_updated_at ON message;
DROP FUNCTION IF EXISTS update_chat_updated_at();


DROP INDEX IF EXISTS idx_message_chat;
DROP INDEX IF EXISTS idx_chat_employer;
DROP INDEX IF EXISTS idx_chat_applicant;

DROP TABLE IF EXISTS message;
DROP TABLE IF EXISTS chat;