
# Конфигурация PostgreSQL

## Структура директорий

```plaintext
/internal/db/
├── postgresql.conf         # Основной конфигурационный файл PostgreSQL
├── pg_hba.conf             # Конфигурация аутентификации
├── create_role.sql         # Скрипт создания роли `orange_user`
```

## Инициализация и запуск PostgreSQL

Сервис `postgres` в `docker-compose.yml` использует кастомную конфигурацию и SQL-скрипт при инициализации:

```yaml
volumes:
  - ./internal/db/postgresql.conf:/etc/postgresql/postgresql.conf
  - ./internal/db/create_role.sql:/docker-entrypoint-initdb.d/create_role.sql
  - ./logs:/var/log/postgresql

command: postgres -c config_file=/etc/postgresql/postgresql.conf
```

### Назначение

- Загрузка пользовательского конфига (`postgresql.conf`)
- Инициализация роли (`create_role.sql`)
- Сбор логов PostgreSQL для анализа через pgBadger (`./logs`)

## Роли и права доступа

### SQL-скрипт `create_role.sql`

```sql
CREATE ROLE orange_user WITH LOGIN PASSWORD 'Orange12345' CONNECTION LIMIT 1;
ALTER ROLE orange_user WITH NOSUPERUSER NOCREATEDB NOCREATEROLE;
ALTER ROLE orange_user SET statement_timeout = '5s';
ALTER ROLE orange_user SET lock_timeout = '3s';

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO orange_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO orange_user;
```

#### Обоснование

- `CONNECTION LIMIT 1` — ограничение количества соединений для защиты от перегрузки
- `NOSUPERUSER` и отсутствие других административных прав — исключение возможности эскалации
- `statement_timeout`, `lock_timeout` — защита от долгих и зависших запросов
- Принцип минимально необходимых привилегий — доступ только к операциям `SELECT`, `INSERT`, `UPDATE`, `DELETE` в `public`

## Аутентификация: `pg_hba.conf`

```conf
# TYPE  DATABASE    USER            ADDRESS             METHOD
local   all         all                                 trust
host    all         all             127.0.0.1/32        trust
host    postgres    orange_user     217.16.23.61        md5
```

#### Обоснование

- `trust` для `local` и `127.0.0.1/32` допустим в среде разработки
- `md5` — безопасный метод пароля для удалённого подключения

## Основной конфигурационный файл: `postgresql.conf`

```conf
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
```

#### Обоснование

- Ограничение на число соединений + резерв для администратора
- Подробное логирование для отладки и анализа производительности
- Подключение `pg_stat_statements` и `auto_explain` — профилирование запросов и анализ «узких мест»

## Пул соединений

```go
db.SetMaxOpenConns(50)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(10 * time.Minute)
```

#### Расчёт

```text
(4 микросервиса × 10 соединений) + 10 = 50 соединений
```

#### Обоснование

- Каждый микросервис использует до 10 соединений
- Резерв в 10 соединений — для мониторинга, администрирования и логирования
- Значения соответствуют `max_connections` в PostgreSQL и обеспечивают стабильную работу

## Логирование и pgBadger

### Сбор логов

- Логи PostgreSQL сохраняются в `./logs`
- Форматирование и ротация логов обеспечивают читаемость и управляемость

### Генерация отчёта

#### Через CLI

```bash
pgbadger ./logs/postgresql-*.log -o report.html
```

#### Через Docker

```bash
docker run --rm \
  -v $(pwd)/logs:/logs \
  -v $(pwd)/report:/output \
  pgbadger/pgbadger \
  /logs/postgresql-*.log -o /output/report.html
```

#### Обоснование

- Анализ логов выполняется офлайн, без нагрузки на базу
- Поддерживается анализ производительности, выявление медленных запросов и неэффективных планов
