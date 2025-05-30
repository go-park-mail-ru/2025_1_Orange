CREATE TYPE work_format_type AS ENUM ('office', 'remote', 'hybrid', 'traveling');
CREATE TYPE employment_type AS ENUM ('full_time', 'part_time', 'contract', 'internship', 'freelance', 'watch');
CREATE TYPE schedule_type AS ENUM ('5/2', '2/2', '6/1', '3/3', 'on_weekend', 'by_agreement');
CREATE TYPE experience_type AS ENUM ('no_matter', 'no_experience', '1_3_years' , '3_6_years', '6_plus_years');

-- Таблица вакансий
CREATE TABLE vacancy (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    employer_id INTEGER NOT NULL REFERENCES employer(id) ON DELETE CASCADE,
    title VARCHAR(128) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    specialization_id INT REFERENCES specialization(id),
    work_format work_format_type NOT NULL,
    employment employment_type NOT NULL,
    schedule schedule_type,
    working_hours INTEGER CHECK (working_hours > 0 AND working_hours <= 168), -- часов в неделю
    salary_from INTEGER CHECK (salary_from >= 0),
    salary_to INTEGER CHECK (salary_to >= 0),
    taxes_included BOOLEAN DEFAULT TRUE,
    experience experience_type,
    description TEXT,
    city TEXT,
    tasks TEXT,
    requirements TEXT,
    optional_requirements TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT salary_check CHECK (salary_to >= salary_from)
);

CREATE TABLE IF NOT EXISTS vacancy_specialization (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id INT NOT NULL REFERENCES resume(id) ON DELETE CASCADE,
    specialization_id INT NOT NULL REFERENCES specialization(id) ON DELETE CASCADE,
    UNIQUE(vacancy_id, specialization_id)
);

-- Таблица связи вакансий и навыков
CREATE TABLE IF NOT EXISTS vacancy_skill (
    id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    vacancy_id INT NOT NULL REFERENCES vacancy(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skill(id) ON DELETE CASCADE,
    UNIQUE(vacancy_id, skill_id)
);

-- Таблица связи вакансий и городов
CREATE TABLE vacancy_city (
    vacancy_id INTEGER NOT NULL REFERENCES vacancy(id) ON DELETE CASCADE,
    city_id INTEGER NOT NULL REFERENCES city(id) ON DELETE CASCADE,
    PRIMARY KEY (vacancy_id, city_id)
);

-- Таблица откликов на вакансии
CREATE TABLE IF NOT EXISTS vacancy_response (
    id SERIAL PRIMARY KEY,
    vacancy_id INTEGER NOT NULL REFERENCES vacancy(id) ON DELETE CASCADE,
    applicant_id INTEGER NOT NULL REFERENCES applicant(id) ON DELETE CASCADE,
    resume_id INTEGER REFERENCES resume(id) ON DELETE CASCADE,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(vacancy_id, applicant_id)
);

-- Таблица лайков вакансий
CREATE TABLE vacancy_like (
    id SERIAL PRIMARY KEY,
    vacancy_id INTEGER NOT NULL REFERENCES vacancy(id) ON DELETE CASCADE,
    applicant_id INTEGER NOT NULL REFERENCES applicant(id) ON DELETE CASCADE,
    liked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (vacancy_id, applicant_id) -- один лайк от соискателя на вакансию
);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_vacancy_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для вакансий
CREATE TRIGGER update_vacancy_timestamp
BEFORE UPDATE ON vacancy
FOR EACH ROW
EXECUTE FUNCTION update_vacancy_timestamp();