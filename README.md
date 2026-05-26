# ИС «Дверной Достык» (doordostyk-is)

Веб-приложение для склад-магазина дверей: каталог, заказы, поступления, продажи, остатки и отчёты. Курсовой проект, ЛГТУ, 2026.

Репозиторий: [github.com/Slazzzer/doordostyk-is](https://github.com/Slazzzer/doordostyk-is)

## Стек

| Слой | Технологии |
|------|------------|
| Клиент | Vue 3 + Vite, vue-router, Pinia, axios |
| API | Go 1.22, Gin, JWT, pgx |
| БД | PostgreSQL 18 (PL/pgSQL: процедура + 2 триггера + 2 VIEW) |
| Инфра | Docker Compose: `web` (Nginx + SPA), `api`, `db` |

## Быстрый старт

```bash
cd doordostyk
docker compose up -d --build
```

После сборки приложение доступно по адресу: **http://localhost:9093**

API напрямую: `http://localhost:9093/api/v1/...` (всё через nginx-проксирование, контейнер `api` снаружи не виден).

Логи: `docker compose logs -f`. Остановка: `docker compose down`. Полный сброс (включая данные БД): `docker compose down -v`.

## Учётные записи (сотрудники)

| Роль | Логин | Пароль |
|------|--------|--------|
| Администратор | `admin` | `admin123` |
| Продавец | `seller` | `seller123` |
| Кладовщик | `storekeeper` | `store123` |

Клиенты в seed не создаются — регистрация на сайте или вкладка «Клиенты» в админ-панели.

## Структура каталогов

```
doordostyk/
├── docker-compose.yml      # 3 сервиса: db / api / web
├── .env                    # переменные окружения
├── db/                     # init-скрипты PostgreSQL
│   ├── 01_schema.sql       # таблицы, FK, индексы
│   ├── 02_objects.sql      # процедура, триггеры, представления
│   └── 03_seed.sql         # тестовые данные
├── backend/                # Go REST API
│   ├── Dockerfile
│   ├── go.mod
│   ├── cmd/server/main.go
│   └── internal/
│       ├── config/         # парсинг env
│       ├── db/             # пул pgx
│       ├── middleware/     # JWT, RBAC
│       ├── model/          # DTO
│       └── handler/        # маршруты Gin
└── frontend/               # Vue 3 SPA
    ├── Dockerfile          # multi-stage: node build → nginx
    ├── nginx.conf          # статика + /api → api:8080
    ├── package.json
    ├── vite.config.js
    └── src/
        ├── fonts/              # Roboto (woff2, локально)
        ├── main.js
        ├── App.vue
        ├── router.js
        ├── api.js
        ├── stores/auth.js
        └── views/*.vue
```

## Основные эндпоинты API

Все запросы: `http://localhost:9093/api/v1/...`, JSON, `Authorization: Bearer <JWT>`.

### Публичные
- `POST /auth/login/user`: вход сотрудника
- `POST /auth/login/customer`: вход клиента
- `POST /auth/register`: регистрация клиента
- `GET  /catalog/products`: каталог (фильтр `category_id`, `q`)
- `GET  /catalog/categories`: категории

### Клиент
- `POST /orders`: создать заказ
- `GET  /orders/my`: мои заказы

### Продавец
- `GET  /orders?status=новый`: очередь заказов
- `POST /orders/:id/execute`: выполнить заказ (вызов `sp_execute_order`)
- `POST /orders/:id/reject`: отклонить
- `POST /sales`: продажа «с полки»
- `GET  /reports/sales?from=&to=&category_id=`: отчёт о продажах

### Кладовщик
- `POST /receipts`: поступление
- `GET  /stock?max_balance=`: остатки (VIEW `v_stock_balance`)
- `GET  /reports/receipts?from=&to=&supplier_id=`: отчёт по поставщикам

### Администратор
- CRUD на `/admin/users`, `/admin/customers`, `/admin/categories`, `/admin/products`, `/admin/suppliers`

## Хранимые объекты БД

| Объект | Тип | Назначение |
|--------|-----|------------|
| `sp_execute_order(order_id, user_id, OUT sale_id)` | PROCEDURE | выполнение заказа (продажа + смена статуса) |
| `trg_sale_check_stock` | TRIGGER на `sale` | контроль остатка перед INSERT/UPDATE |
| `trg_order_status_guard` | TRIGGER на `"order"` | контроль допустимых статусов |
| `v_sales_by_category` | VIEW | продажи по категориям |
| `v_stock_balance` | VIEW | остатки на складе |
| `fn_product_balance(product_id, exclude_sale_id)` | FUNCTION | вспомогательная: остаток |

## Замечание о расширении PDM

Для аутентификации к таблицам `"user"` и `customer` добавлены поля `*_password_hash VARCHAR(72)` (bcrypt). Это **минимальное** расширение PDM; на спецификацию сущностей и связей не влияет.

## Локальная разработка с IDE

Если IDE (VS Code / Cursor / GoLand) ругается на ненайденные пакеты `github.com/gin-gonic/gin`, `github.com/jackc/pgx/v5` и др.: нужно один раз подтянуть модули локально:

```bash
cd backend
go mod tidy
```

После этого перезапустите Go language server (`Ctrl+Shift+P → Go: Restart Language Server` в VS Code/Cursor). Файл `go.sum` коммитьте в git вместе с `go.mod`.

## Разработка без Docker

```bash
# 1. БД (Docker или локальная PostgreSQL 18)
psql -U dostyk -d doordostyk -f db/01_schema.sql
psql -U dostyk -d doordostyk -f db/02_objects.sql
psql -U dostyk -d doordostyk -f db/03_seed.sql

# 2. API
cd backend
go mod download
DATABASE_URL="postgres://dostyk:dostyk_pass_2026@localhost:5432/doordostyk?sslmode=disable" \
JWT_SECRET="dev_secret" HTTP_PORT=8080 \
go run ./cmd/server

# 3. SPA
cd frontend
npm install
npm run dev  # http://localhost:5173 c прокси на :8080
```

## Лицензия

Учебный проект (LGTU, ПИ-23-2). Использовать свободно.
