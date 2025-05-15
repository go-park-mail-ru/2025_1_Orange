📦 Структура конфигурации
Конфиги и скрипты лежат в следующих директориях:

/internal/db/
├── postgresql.conf         # Основной конфиг PostgreSQL
├── pg_hba.conf             # Конфиг авторизации пользователей
├── create_role.sql         # Скрипт создания роли orange_user с ограничениями

🔌 Подключение PostgreSQL
В docker-compose.yml сервис postgres подключается с указанием кастомного конфига и инициализационного SQL:

volumes:
  - ./internal/db/postgresql.conf:/etc/postgresql/postgresql.conf
  - ./internal/db/create_role.sql:/docker-entrypoint-initdb.d/create_role.sql
  - ./logs:/var/log/postgresql
command: postgres -c config_file=/etc/postgresql/postgresql.conf
📌 Это позволяет:

Использовать свою конфигурацию (postgresql.conf)

Установить необходимые роли при старте (create_role.sql)

Собрать логи PostgreSQL в каталог ./logs для последующего анализа через pgBadger.

👤 Роли и безопасность
Файл create_role.sql:

CREATE ROLE orange_user WITH LOGIN PASSWORD 'Orange12345' CONNECTION LIMIT 1;
ALTER ROLE orange_user WITH NOSUPERUSER NOCREATEDB NOCREATEROLE;
ALTER ROLE orange_user SET statement_timeout = '5s';
ALTER ROLE orange_user SET lock_timeout = '3s';

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO orange_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO orange_user;
✅ Обоснование:

CONNECTION LIMIT 1 — ограничиваем ресурсы на пользователя (DDoS-защита).

Без SUPERUSER — исключаем эскалацию прав.

Timeout'ы — защита от висящих/долгих запросов.

Привилегии только на нужные операции — принцип наименьших привилегий.

🔐 pg_hba.conf

# TYPE  DATABASE    USER            ADDRESS             METHOD
local   all         all                                 trust
host    all         all             127.0.0.1/32        trust
host    postgres    orange_user     217.16.23.61        md5
✅ Обоснование:

Используется md5 для подключения снаружи (а не trust).

Указан конкретный IP, с которого разрешено подключение пользователю orange_user.

Локальные соединения разрешены trust — допустимо в dev-окружении.

⚙️ Параметры PostgreSQL (postgresql.conf)

listen_addresses = 'resumatch,auth,static,localhost'
max_connections = 50
superuser_reserved_connections = 3
shared_preload_libraries = 'pg_stat_statements,auto_explain'

# Логирование
log_min_duration_statement = 1000
log_statement = 'all'
log_duration = on
logging_collector = on
log_directory = '/var/log/postgresql'
log_filename = 'postgresql-%Y-%m-%d.log'
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_rotation_age = 1d
log_rotation_size = 100MB
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on
log_temp_files = 0

# pg_stat_statements
pg_stat_statements.max = 10000
pg_stat_statements.track = all
pg_stat_statements.track_utility = on

# auto_explain
auto_explain.log_min_duration = '1s'
auto_explain.log_analyze = true
auto_explain.log_verbose = true
auto_explain.log_buffers = true
auto_explain.log_timing = true
auto_explain.log_triggers = true
auto_explain.log_format = 'json'
✅ Обоснование:

Включено подробное логирование, включая медленные запросы.

Сбор расширенной статистики (для pgBadger и анализа производительности).

Используются pg_stat_statements и auto_explain — это инструменты профилирования производительности.

🏗️ Пул соединений
Пул создаётся вручную:

db.SetMaxOpenConns(50)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(10 * time.Minute)
🔹 Каждый сервис имеет собственный пул. Размер пула сбалансирован:

max_connections = 10 * 4 + 10 = 50
✅ Обоснование:

Каждый из 4 микросервисf выделяет по 10 соединений.

10 соединений — резерв для админа или метрик.

Это соответствует заданию и не перегружает PostgreSQL.

📊 Логирование с pgBadger
pgBadger подключён следующим образом:

Логи PostgreSQL собираются в ./logs

Через команду:

pgbadger ./logs/postgresql-*.log -o report.html
или через Docker:

docker run --rm -v $(pwd)/logs:/logs -v $(pwd)/report:/output pgbadger/pgbadger /logs/postgresql-*.log -o /output/report.html