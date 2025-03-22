## ER-диаграмма

```mermaid
erDiagram
    STATIC {
        INT id PK "Идентификатор файла"
        TEXT path "Путь к файлу в MinIO/S3"
        TIMESTAMP created_at "Дата и время создания файла"
        TIMESTAMP updated_at "Дата и время последнего обновления файла"
    }

    APPLICANT {
        INT id PK "Идентификатор соискателя"
        TEXT first_name "Имя соискателя"
        TEXT last_name "Фамилия соискателя"
        TEXT middle_name "Отчество соискателя"
        INT city_id FK "Идентификатор города проживания"
        DATE birth_date "Дата рождения"
        TEXT sex "Пол соискателя"
        TEXT email "Электронная почта"
        TEXT password "Пароль"
        INT status_id FK "Идентификатор статуса поиска работы"
        INT avatar_id FK "Идентификатор фото"
        TIMESTAMP created_at "Дата и время создания профиля"
        TIMESTAMP updated_at "Дата и время последнего обновления профиля"
    }

    CITY {
        INT id PK "Идентификатор города"
        TEXT name "Название города"
    }

    VACANCY_CITY {
        INT id PK "Идентификатор связи"
        INT vacancy_id FK "Идентификатор вакансии"
        INT city_id FK "Идентификатор города"
    }

    JOB_SEARCH_STATUS {
        INT id PK "Идентификатор статуса"
        TEXT name "Статус"
    }

    EMPLOYER {
        INT id PK "Идентификатор работодателя"
        TEXT name "Название работодателя"
        TEXT slogan "Слоган работодателя"
        TEXT website "Адрес сайта работодателя"
        TEXT description "Описание работодателя"
        TEXT legal_address "Юридический адрес"
        TEXT email "Электронная почта"
        INT logo_id FK "Идентификатор логотипа"
        TIMESTAMP created_at "Дата и время создания профиля работодателя"
        TIMESTAMP updated_at "Дата и время последнего обновления профиля работодателя"
    }

    EMPLOYER_INFO_LINK {
        INT id PK "Идентификатор связи"
        INT employer_id FK "Идентификатор работодателя"
        INT image_id FK "Идентификатор логотипа"
        TEXT url "URL ссылки на ресурс"
    }

    VACANCY {
        INT id PK "Идентификатор вакансии"
        TEXT title "Название вакансии"
        BOOLEAN is_active "Статус вакансии (активна/неактивна)"
        INT employer_id FK "Идентификатор работодателя"
        INT specialization_id FK "Идентификатор специализации"
        INT work_format_id FK "Идентификатор формата работы"
        INT employment_type_id FK "Идентификатор типа занятости"
        INT schedule_id FK "Идентификатор графика работы"
        INT working_hours "Рабочие часы"
        INT salary_from "Минимальная зарплата"
        INT salary_to "Максимальная зарплата"
        INT payment_frequency_id FK "Идентификатор частоты выплат"
        TEXT taxes "Условия по налогам"
        INT experience_id FK "Идентификатор уровня опыта"
        TEXT description "Описание вакансии"
        TEXT tasks "Задачи по вакансии"
        TEXT requirements "Требования к кандидату"
        TEXT optional_requirements "Будет плюсом"
        TIMESTAMP created_at "Дата и время создания вакансии"
        TIMESTAMP updated_at "Дата и время последнего обновления вакансии"
    }

    SPECIALIZATION {
        INT id PK "Идентификатор специализации"
        TEXT name "Название специализации"
    }

    WORK_FORMAT {
        INT id PK "Идентификатор формата работы"
        TEXT name "Название формата работы"
    }

    EMPLOYMENT_TYPE {
        INT id PK "Идентификатор типа занятости"
        TEXT name "Название типа занятости"
    }

    SCHEDULE {
        INT id PK "Идентификатор графика работы"
        TEXT name "Название графика работы"
    }

    PAYMENT_FREQUENCY {
        INT id PK "Идентификатор частоты выплат"
        TEXT name "Название частоты выплат"
    }

    EXPERIENCE {
        INT id PK "Идентификатор уровня опыта"
        TEXT name "Название уровня опыта"
    }

    SUPPLEMENTARY_CONDITIONS {
        INT id PK "Идентификатор условия"
        TEXT name "Название условия"
    }

    VACANCY_SUPPLEMENTARY_CONDITIONS {
        INT id PK "Идентификатор связи"
        INT vacancy_id FK "Идентификатор вакансии"
        INT condition_id FK "Идентификатор условия"
    }

    SKILL {
        INT id PK "Идентификатор навыка"
        TEXT name "Название навыка"
    }

    VACANCY_SKILL {
        INT id PK "Идентификатор связи"
        INT vacancy_id FK "Идентификатор вакансии"
        INT skill_id FK "Идентификатор навыка"
    }

    VACANCY_RESPONSE {
        INT id PK "Идентификатор отклика"
        INT vacancy_id FK "Идентификатор вакансии"
        INT applicant_id FK "Идентификатор соискателя"
        TIMESTAMP applied_at "Дата и время отклика"
    }

    RESUME_SPECIALIZATION {
        INT id PK "Идентификатор связи"
        INT resume_id FK "Идентификатор резюме"
        INT specialization_id FK "Идентификатор специализации"
    }

    RESUME {
        INT id PK "Идентификатор резюме"
        INT applicant_id FK "Идентификатор соискателя"
        TEXT about_me "Информация о себе"
        INT specialization_id FK "Идентификатор специализации"
        INT education_type_id FK "Идентификатор типа образования"
        TEXT educational_institution "Название учебного заведения"
        DATE graduation_year "Год окончания"
        TIMESTAMP created_at "Дата и время создания резюме"
        TIMESTAMP updated_at "Дата и время последнего обновления резюме"
    }

    EDUCATION_TYPE {
        INT id PK "Идентификатор типа образования"
        TEXT name "Название типа образования"
    }

    WORK_EXPERIENCE {
        INT id PK "Идентификатор записи об опыте работы"
        INT resume_id FK "Идентификатор резюме"
        TEXT employer_name "Название работодателя"
        TEXT position "Должность"
        TEXT duties "Обязанности"
        TEXT achievements "Достижения"
        DATE start_date "Дата начала работы"
        DATE end_date "Дата окончания работы"
        BOOLEAN until_now "Текущее место работы"
    }

    RESUME_SKILL {
        INT id PK "Идентификатор связи"
        INT resume_id FK "Идентификатор резюме"
        INT skill_id FK "Идентификатор навыка"
    }

    STATIC ||--|| APPLICANT : "avatar_id"
    CITY ||--o{ APPLICANT : "city_id"
    JOB_SEARCH_STATUS ||--|| APPLICANT : "status_id"
    STATIC ||--|| EMPLOYER : "logo_id"
    EMPLOYER ||--o{ EMPLOYER_INFO_LINK : "employer_id"
    STATIC ||--o{ EMPLOYER_INFO_LINK : "image_id"
    EMPLOYER ||--o{ VACANCY : "employer_id"
    SPECIALIZATION ||--o{ VACANCY : "specialization_id"
    WORK_FORMAT ||--o{ VACANCY : "work_format_id"
    EMPLOYMENT_TYPE ||--o{ VACANCY : "employment_type_id"
    SCHEDULE ||--o{ VACANCY : "schedule_id"
    PAYMENT_FREQUENCY ||--o{ VACANCY : "payment_frequency_id"
    EXPERIENCE ||--o{ VACANCY : "experience_id"
    VACANCY ||--o{ VACANCY_CITY : "vacancy_id"
    CITY ||--o{ VACANCY_CITY : "city_id"
    VACANCY ||--o{ VACANCY_SUPPLEMENTARY_CONDITIONS : "vacancy_id"
    SUPPLEMENTARY_CONDITIONS ||--o{ VACANCY_SUPPLEMENTARY_CONDITIONS : "condition_id"
    VACANCY ||--o{ VACANCY_SKILL : "vacancy_id"
    SKILL ||--o{ VACANCY_SKILL : "skill_id"
    VACANCY ||--o{ VACANCY_RESPONSE : "vacancy_id"
    APPLICANT ||--o{ VACANCY_RESPONSE : "applicant_id"
    RESUME ||--o{ RESUME_SPECIALIZATION : "resume_id"
    SPECIALIZATION ||--o{ RESUME_SPECIALIZATION : "specialization_id"
    APPLICANT ||--o{ RESUME : "applicant_id"
    EDUCATION_TYPE ||--o{ RESUME : "education_type_id"
    RESUME ||--o{ WORK_EXPERIENCE : "resume_id"
    RESUME ||--o{ RESUME_SKILL : "resume_id"
    SKILL ||--o{ RESUME_SKILL : "skill_id"
```