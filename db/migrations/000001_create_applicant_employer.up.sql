CREATE TYPE applicant_status_type AS ENUM (
    'actively_searching',      -- Активно ищу работу
    'open_to_offers',          -- Рассматриваю предложения
    'considering_offer',       -- Предложили работу, пока думаю
    'starting_soon',           -- Уже выхожу на новое место
    'not_searching'            -- Не ищу работу
);

-- Создание таблицы city
CREATE TABLE IF NOT EXISTS city
(
    id   INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name TEXT
    CONSTRAINT name_length CHECK (LENGTH(name) <= 30) NOT NULL UNIQUE
);

-- Создание таблицы static
CREATE TABLE IF NOT EXISTS static
(
    id         INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    file_path       TEXT
    CONSTRAINT upload_file_path_length CHECK (LENGTH(file_path) <= 255) NOT NULL,
    file_name       TEXT
    CONSTRAINT upload_file_name_length CHECK (LENGTH(file_name) <= 255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
    birth_date     TIMESTAMP WITH TIME ZONE,
    sex            CHAR(1)
    CONSTRAINT person_gender CHECK (sex = 'M' OR sex = 'F'),
    email          TEXT
    CONSTRAINT valid_email_format CHECK (
        email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}(?:\.[A-Za-z]{2,})?$'
    )
    CONSTRAINT email_length CHECK (LENGTH(email) <= 255) NOT NULL UNIQUE,
    password_hashed  bytea
    CONSTRAINT password_hashed_length CHECK (OCTET_LENGTH(password_hashed) <= 32) NOT NULL,
    password_salt  bytea
    CONSTRAINT password_salt_length CHECK (OCTET_LENGTH(password_salt) <= 8) NOT NULL,
    status         applicant_status_type,
    quote          TEXT
    CONSTRAINT applicant_quote_length CHECK (LENGTH(quote) <= 255),
    avatar_id      INT REFERENCES static (id) ON DELETE CASCADE,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы employer
CREATE TABLE IF NOT EXISTS employer
(
    id              INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    company_name            TEXT
    CONSTRAINT employer_company_name_length CHECK (LENGTH(company_name) <= 64) NOT NULL UNIQUE,
    slogan          TEXT,
    website         TEXT
    CONSTRAINT employer_website_url_length CHECK (LENGTH(website) <= 255) UNIQUE,
    description     TEXT,
    legal_address   TEXT
    CONSTRAINT employer_legal_address_length CHECK (LENGTH(legal_address) <= 255),
    email          TEXT
    CONSTRAINT valid_email_format CHECK (
         email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}(?:\.[A-Za-z]{2,})?$'
    )
    CONSTRAINT email_length CHECK (LENGTH(email) <= 255) NOT NULL UNIQUE,
    password_hashed  bytea
    CONSTRAINT password_hashed_length CHECK (OCTET_LENGTH(password_hashed) <= 32) NOT NULL,
    password_salt  bytea
    CONSTRAINT password_salt_length CHECK (OCTET_LENGTH(password_salt) <= 8) NOT NULL,
    logo_id         INT REFERENCES static (id) ON DELETE CASCADE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
-- static
CREATE TRIGGER update_static_timestamp
    BEFORE UPDATE
    ON static
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- applicant
CREATE TRIGGER update_applicant_timestamp
    BEFORE UPDATE
    ON applicant
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- employer
CREATE TRIGGER update_employer_timestamp
    BEFORE UPDATE
    ON employer
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp();