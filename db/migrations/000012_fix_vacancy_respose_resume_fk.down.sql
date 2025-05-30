ALTER TABLE vacancy_response
DROP CONSTRAINT IF EXISTS vacancy_response_resume_id_fkey,
ADD CONSTRAINT vacancy_response_resume_id_fkey
    FOREIGN KEY (resume_id)
    REFERENCES resume(id)
    ON DELETE SET NULL;