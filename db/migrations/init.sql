CREATE TYPE work_format_type AS ENUM (
    'office',
    'remote',
    'hybrid',
    'traveling'
);

CREATE TYPE employment_type AS ENUM (
    'full_time',
    'part_time',
    'contract',
    'internship',
    'freelance',
    'watch'
);

CREATE TYPE schedule_type AS ENUM (
    '5/2',
    '2/2',
    '6/1',
    '3/3',
    'on_weekend',
    'by_agreement'
);

CREATE TYPE experience_type AS ENUM (
    'no_matter',
    'no_experience',
    '1_3_years',
    '3_6_years',
    '6_plus_years'
);

CREATE TYPE working_hours_type AS ENUM (
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    '10',
    '11',
    '12',
    '24',
    'by_agreement'
);

CREATE TYPE education_type AS ENUM (
    'secondary_school',
    'incomplete_higher',
    'higher',
    'bachelor',
    'master',
    'phd'
);

-- Создание таблицы static
CREATE TABLE IF NOT EXISTS static
(
    id         INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    path       TEXT
        CONSTRAINT upload_path_length CHECK (LENGTH(path) <= 255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы city
CREATE TABLE IF NOT EXISTS city
(
    id   INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT
        CONSTRAINT name_length CHECK (LENGTH(name) <= 64) NOT NULL UNIQUE
);

-- Создание таблицы specialization
CREATE TABLE IF NOT EXISTS specialization
(
    id   INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT
        CONSTRAINT name_length CHECK (LENGTH(name) <= 255) NOT NULL UNIQUE
);

-- Создание таблицы skill
CREATE TABLE IF NOT EXISTS skill
(
    id   INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT
        CONSTRAINT name_length CHECK (LENGTH(name) <= 64) NOT NULL UNIQUE
);

-- Создание таблицы supplementary_conditions
CREATE TABLE IF NOT EXISTS supplementary_conditions
(
    id   INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT
        CONSTRAINT name_length CHECK (LENGTH(name) <= 255) NOT NULL UNIQUE
);

-- Создание таблицы applicant
CREATE TABLE IF NOT EXISTS applicant
(
    id             INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    first_name     TEXT
        CONSTRAINT applicant_first_name_length CHECK (LENGTH(first_name) <= 30) NOT NULL,
    last_name      TEXT
        CONSTRAINT applicant_last_name_length CHECK (LENGTH(last_name) <= 30) NOT NULL,
    middle_name    TEXT
        CONSTRAINT applicant_middle_name_length CHECK (LENGTH(middle_name) <= 30),
    city_id        INT REFERENCES city (id),
    birth_date     DATE,
    sex            CHAR(1)
        CONSTRAINT person_gender CHECK (sex = 'M' OR sex = 'F') NOT NULL,
    email          TEXT
        CONSTRAINT valid_email_format CHECK (
            email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}(?:\.[A-Za-z]{2,})?$'
        ),
        CONSTRAINT email_length CHECK (LENGTH(email) <= 255) NOT NULL UNIQUE,
    password_hashed  bytea
        CONSTRAINT password_hashed_length CHECK (OCTET_LENGTH(password_hashed) <= 32) NOT NULL,
    password_salt  bytea,
        CONSTRAINT password_salt_length CHECK (OCTET_LENGTH(password_salt) <= 8) NOT NULL,
    status         INT,
    avatar_id      INT REFERENCES static (id) ON DELETE CASCADE,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы employer
CREATE TABLE IF NOT EXISTS employer
(
    id              INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name            TEXT
        CONSTRAINT name_length CHECK (LENGTH(name) <= 64) NOT NULL UNIQUE,
    slogan          TEXT,
    website         TEXT
        CONSTRAINT website_url_length CHECK (LENGTH(website) <= 255) UNIQUE,
    description     TEXT,
    legal_address   TEXT
        CONSTRAINT legal_address_length CHECK (LENGTH(legal_address) <= 255),
    email          TEXT
        CONSTRAINT valid_email_format CHECK (
            email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}(?:\.[A-Za-z]{2,})?$'
        ),
        CONSTRAINT email_length CHECK (LENGTH(email) <= 255) NOT NULL UNIQUE,
    password_hashed  bytea
        CONSTRAINT password_hashed_length CHECK (OCTET_LENGTH(password_hashed) <= 32) NOT NULL,
    password_salt  bytea,
        CONSTRAINT password_salt_length CHECK (OCTET_LENGTH(password_salt) <= 8) NOT NULL,
    logo_id         INT REFERENCES static (id) ON DELETE CASCADE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для хранения внешних ссылок пользователей
CREATE TABLE IF NOT EXISTS user_info_link (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    user_type TEXT
        CHECK (user_type IN ('applicant', 'employer')) NOT NULL,
    user_id INT NOT NULL,
    url TEXT
        CONSTRAINT url_length CHECK (LENGTH(url) <= 255) NOT NULL,
    image_id INT REFERENCES static(id) ON DELETE SET NULL
);

-- Создание таблицы vacancy
CREATE TABLE IF NOT EXISTS vacancy
(
    id                    INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    title                 TEXT
        CONSTRAINT url_length CHECK (LENGTH(url) <= 64) NOT NULL,
    is_active             BOOLEAN NOT NULL DEFAULT TRUE,
    employer_id           INT NOT NULL REFERENCES employer (id) ON DELETE CASCADE,
    specialization_id     INT REFERENCES specialization (id),
    work_format           work_format_type,
    employment            employment_type,
    schedule              schedule_type,
    working_hours         working_hours_type,
    salary_from           INT,
    salary_to             INT,
    taxes_included        BOOLEAN,
    experience            experience_type,
    description           TEXT NOT NULL,
    tasks                 TEXT,
    requirements          TEXT,
    optional_requirements TEXT,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы vacancy_city
CREATE TABLE IF NOT EXISTS vacancy_city
(
    id        INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id INT NOT NULL REFERENCES vacancy (id) ON DELETE CASCADE,
    city_id   INT NOT NULL REFERENCES city (id) ON DELETE CASCADE
);


-- Создание таблицы vacancy_supplementary_conditions
CREATE TABLE IF NOT EXISTS vacancy_supplementary_conditions
(
    id          INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id  INT NOT NULL REFERENCES vacancy (id) ON DELETE CASCADE,
    condition_id INT NOT NULL REFERENCES supplementary_conditions (id) ON DELETE CASCADE
);

-- Создание таблицы vacancy_skill
CREATE TABLE IF NOT EXISTS vacancy_skill
(
    id        INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id INT NOT NULL REFERENCES vacancy (id) ON DELETE CASCADE,
    skill_id  INT NOT NULL REFERENCES skill (id) ON DELETE CASCADE
);

-- Создание таблицы resume
CREATE TABLE IF NOT EXISTS resume
(
    id                      INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    applicant_id            INT NOT NULL REFERENCES applicant (id) ON DELETE CASCADE,
    about_me                TEXT,
    specialization_id       INT REFERENCES specialization (id),
    education               education_type,
    educational_institution TEXT,
    graduation_year         DATE,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы work_experience
CREATE TABLE IF NOT EXISTS work_experience
(
    id             INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id      INT NOT NULL REFERENCES resume (id) ON DELETE CASCADE,
    employer_name  TEXT
        CONSTRAINT employer_name_length CHECK (LENGTH(employer_name) <= 64) NOT NULL,
    position       TEXT
        CONSTRAINT position_length CHECK (LENGTH(position) <= 64) NOT NULL,
    duties         TEXT,
    achievements   TEXT,
    start_date     DATE NOT NULL,
    end_date       DATE,
    until_now      BOOLEAN NOT NULL DEFAULT FALSE
);

-- Создание таблицы resume_specialization
CREATE TABLE IF NOT EXISTS resume_specialization
(
    id                INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id         INT NOT NULL REFERENCES resume (id) ON DELETE CASCADE,
    specialization_id INT NOT NULL REFERENCES specialization (id) ON DELETE CASCADE
);

-- Создание таблицы resume_skill
CREATE TABLE IF NOT EXISTS resume_skill
(
    id        INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    resume_id INT NOT NULL REFERENCES resume (id) ON DELETE CASCADE,
    skill_id  INT NOT NULL REFERENCES skill (id) ON DELETE CASCADE
);

-- Создание таблицы vacancy_response
CREATE TABLE IF NOT EXISTS vacancy_response
(
    id           INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id   INT NOT NULL REFERENCES vacancy (id) ON DELETE CASCADE,
    applicant_id INT NOT NULL REFERENCES applicant (id) ON DELETE CASCADE,
    applied_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы vacancy_like
CREATE TABLE IF NOT EXISTS vacancy_like
(
    id           INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id   INT NOT NULL REFERENCES vacancy (id) ON DELETE CASCADE,
    applicant_id INT NOT NULL REFERENCES applicant (id) ON DELETE CASCADE,
    liked_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Триггер для обновления поля updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Добавление триггера к каждой таблице, которая имеет поле updated_at
-- applicant
CREATE TRIGGER update_applicant_timestamp
    BEFORE UPDATE
    ON applicant
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- static
CREATE TRIGGER update_static_timestamp
    BEFORE UPDATE
    ON static
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- employer
CREATE TRIGGER update_employer_timestamp
    BEFORE UPDATE
    ON employer
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- vacancy
CREATE TRIGGER update_vacancy_timestamp
    BEFORE UPDATE
    ON vacancy
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- resume
CREATE TRIGGER update_resume_timestamp
    BEFORE UPDATE
    ON resume
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();


-- Индексы для таблицы vacancy
CREATE INDEX idx_vacancy_employer ON vacancy(employer_id);
CREATE INDEX idx_vacancy_specialization ON vacancy(specialization_id);

-- Индексы для таблицы resume
CREATE INDEX idx_resume_applicant ON resume(applicant_id);
CREATE INDEX idx_resume_specialization ON resume(specialization_id);