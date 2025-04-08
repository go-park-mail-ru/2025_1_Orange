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
    applicant_id INT NOT NULL REFERENCES applicant(id),
    about_me TEXT NOT NULL,
    specialization_id INT NOT NULL REFERENCES specialization(id),
    education INT NOT NULL,
    educational_institution TEXT NOT NULL,
    graduation_year DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создание таблицы resume_specialization
CREATE TABLE IF NOT EXISTS resume_specialization (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
    specialization_id INT NOT NULL REFERENCES specialization(id),
    UNIQUE(resume_id, specialization_id)
);

-- Создание таблицы resume_skill
CREATE TABLE IF NOT EXISTS resume_skill (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skill(id),
    UNIQUE(resume_id, skill_id)
);

-- Создание таблицы work_experience
CREATE TABLE IF NOT EXISTS work_experience (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
    employer_name TEXT NOT NULL,
    position TEXT NOT NULL,
    duties TEXT NOT NULL,
    achievements TEXT,
    start_date DATE NOT NULL,
    end_date DATE,
    until_now BOOLEAN NOT NULL DEFAULT FALSE,
    CHECK (until_now OR end_date IS NOT NULL)
);