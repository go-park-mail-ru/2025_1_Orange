ALTER TABLE applicant
DROP CONSTRAINT applicant_avatar_id_fkey,
ADD CONSTRAINT applicant_avatar_id_fkey
FOREIGN KEY (avatar_id) REFERENCES static (id) ON DELETE CASCADE;

ALTER TABLE employer
DROP CONSTRAINT employer_logo_id_fkey,
ADD CONSTRAINT employer_logo_id_fkey
FOREIGN KEY (logo_id) REFERENCES static (id) ON DELETE CASCADE;