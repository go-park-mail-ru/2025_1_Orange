CREATE ROLE orange_user WITH LOGIN PASSWORD 'Orange12345' CONNECTION LIMIT 1;

ALTER ROLE orange_user WITH NOSUPERUSER NOCREATEDB NOCREATEROLE;

ALTER ROLE orange_user SET statement_timeout = '10s';  -- Ограничиваем общее время выполнения любого SQL-запроса
ALTER ROLE orange_user SET lock_timeout = '5s';       -- Ограничиваем время ожидания блокировок при выполнении запроса

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO orange_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO orange_user;
