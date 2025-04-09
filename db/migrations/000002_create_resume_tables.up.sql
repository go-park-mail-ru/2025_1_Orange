-- Create custom ENUM types
CREATE TYPE education_type AS ENUM (
    'secondary_school',
    'incomplete_higher',
    'higher',
    'bachelor',
    'master',
    'phd'
);

-- Создание таблицы specialization (если она еще не существует)
CREATE TABLE IF NOT EXISTS specialization (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Создание таблицы skill
CREATE TABLE IF NOT EXISTS skill (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- Создание таблицы resume
CREATE TABLE IF NOT EXISTS resume (
   id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
   applicant_id INT NOT NULL REFERENCES applicant(id) ON DELETE CASCADE,
   about_me TEXT,
   specialization_id INT REFERENCES specialization(id),
   education education_type,
   educational_institution TEXT,
   graduation_year DATE,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы resume_specialization
CREATE TABLE IF NOT EXISTS resume_specialization (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
    specialization_id INT NOT NULL REFERENCES specialization(id) ON DELETE CASCADE,
    UNIQUE(resume_id, specialization_id)
);

-- Создание таблицы resume_skill
CREATE TABLE IF NOT EXISTS resume_skill (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
    UNIQUE(resume_id, skill_id)
);

-- Создание таблицы work_experience
CREATE TABLE IF NOT EXISTS work_experience (
   id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
   resume_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
   employer_name TEXT CONSTRAINT employer_name_length CHECK (LENGTH(employer_name) <= 64) NOT NULL,
   position TEXT CONSTRAINT position_length CHECK (LENGTH(position) <= 64) NOT NULL,
   duties TEXT,
   achievements TEXT,
   start_date DATE NOT NULL,
   end_date DATE,
   until_now BOOLEAN NOT NULL DEFAULT FALSE,
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Триггер для обновления поля updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Добавление триггера к таблицам
CREATE TRIGGER update_resume_timestamp
    BEFORE UPDATE
    ON resume
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_work_experience_timestamp
    BEFORE UPDATE
    ON work_experience
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();