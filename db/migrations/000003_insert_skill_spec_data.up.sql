-- Заполнение таблицы specialization
INSERT INTO specialization (name) VALUES
    ('Программист'),
    ('Аналитик'),
    ('Дизайнер'),
    ('Менеджер проектов'),
    ('Тестировщик'),
    ('DevOps инженер'),
    ('Системный администратор'),
    ('Маркетолог'),
    ('HR специалист'),
    ('Продуктовый менеджер'),
    ('UX/UI дизайнер'),
    ('Бизнес-аналитик'),
    ('Data Scientist'),
    ('Frontend разработчик'),
    ('Backend разработчик'),
    ('Fullstack разработчик'),
    ('Mobile разработчик'),
    ('QA инженер'),
    ('Технический писатель'),
    ('SEO специалист')
ON CONFLICT (name) DO NOTHING;

-- Заполнение таблицы skill
INSERT INTO skill (name) VALUES
    -- Языки программирования
    ('JavaScript'),
    ('Python'),
    ('Java'),
    ('C#'),
    ('C++'),
    ('Go'),
    ('Ruby'),
    ('PHP'),
    ('Swift'),
    ('Kotlin'),
    ('TypeScript'),
    ('Rust'),
    ('Scala'),
    
    -- Фреймворки и библиотеки
    ('React'),
    ('Angular'),
    ('Vue.js'),
    ('Node.js'),
    ('Django'),
    ('Flask'),
    ('Spring'),
    ('ASP.NET'),
    ('Laravel'),
    ('Ruby on Rails'),
    ('Express.js'),
    
    -- Базы данных
    ('SQL'),
    ('PostgreSQL'),
    ('MySQL'),
    ('MongoDB'),
    ('Redis'),
    ('Elasticsearch'),
    ('Cassandra'),
    ('Oracle'),
    ('MS SQL Server'),
    
    -- DevOps и инфраструктура
    ('Docker'),
    ('Kubernetes'),
    ('AWS'),
    ('Azure'),
    ('Google Cloud'),
    ('CI/CD'),
    ('Jenkins'),
    ('Terraform'),
    ('Ansible'),
    
    -- Тестирование
    ('Автоматизированное тестирование'),
    ('Ручное тестирование'),
    ('Selenium'),
    ('JUnit'),
    ('TestNG'),
    ('Cypress'),
    ('Jest'),
    
    -- Дизайн
    ('Figma'),
    ('Adobe Photoshop'),
    ('Adobe Illustrator'),
    ('Sketch'),
    ('InVision'),
    ('Adobe XD'),
    
    -- Аналитика
    ('SQL'),
    ('Power BI'),
    ('Tableau'),
    ('Excel'),
    ('Google Analytics'),
    ('A/B тестирование'),
    
    -- Soft skills
    ('Коммуникабельность'),
    ('Работа в команде'),
    ('Управление проектами'),
    ('Agile'),
    ('Scrum'),
    ('Kanban'),
    ('Лидерство'),
    ('Критическое мышление'),
    ('Решение проблем'),
    ('Тайм-менеджмент')
ON CONFLICT (name) DO NOTHING;