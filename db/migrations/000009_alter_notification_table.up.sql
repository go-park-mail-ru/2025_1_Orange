ALTER TABLE notification ADD COLUMN resume_id INT NOT NULL;

CREATE TYPE user_type AS ENUM ('applicant', 'employer');

ALTER TABLE notification ADD COLUMN sender_role user_type;
ALTER TABLE notification ADD COLUMN receiver_role user_type;
