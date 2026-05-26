-- ИС «Дверной Достык»: структура БД
-- PostgreSQL 18 (совместимо с 9.x)
-- Сгенерировано из PDM «Дверной_Достык_PDM» (PowerDesigner) с минимальным
-- расширением: добавлены customer_password_hash и user_password_hash для JWT.

SET client_encoding = 'UTF8';
SET client_min_messages = WARNING;

/*========== Таблицы (без FK) ==========*/

create table category (
   category_id          SERIAL               not null,
   category_name        VARCHAR(100)         not null,
   category_description VARCHAR(500)         null,
   constraint PK_CATEGORY primary key (category_id)
);

create table customer (
   customer_id          SERIAL               not null,
   customer_full_name   VARCHAR(150)         not null,
   customer_email       VARCHAR(100)         null,
   customer_phone_number VARCHAR(20)         null,
   customer_password_hash VARCHAR(72)        null,
   constraint PK_CUSTOMER primary key (customer_id)
);

create table supplier (
   supplier_id          SERIAL               not null,
   organization_name    VARCHAR(200)         not null,
   supplier_address     VARCHAR(300)         null,
   supplier_phone_number VARCHAR(20)         null,
   constraint PK_SUPPLIER primary key (supplier_id)
);

create table "user" (
   user_id              SERIAL               not null,
   user_full_name       VARCHAR(150)         not null,
   user_login           VARCHAR(50)          not null,
   user_role            VARCHAR(30)          not null,
   user_password_hash   VARCHAR(72)          null,
   constraint PK_USER primary key (user_id)
);

create table product (
   product_id           SERIAL               not null,
   category_id          INT4                 not null,
   product_name         VARCHAR(200)         not null,
   product_description  TEXT                 null,
   product_dimensions   VARCHAR(100)         null,
   product_purchase_price DECIMAL(12,2)      null,
   product_retail_price DECIMAL(12,2)        null,
   constraint PK_PRODUCT primary key (product_id)
);

create table sale (
   sale_id              SERIAL               not null,
   order_id             INT4                 null,
   product_id           INT4                 not null,
   user_id              INT4                 not null,
   sale_date            DATE                 not null,
   sale_quantity        INT4                 not null,
   sale_price           DECIMAL(12,2)        not null,
   constraint PK_SALE primary key (sale_id)
);

create table "order" (
   order_id             SERIAL               not null,
   customer_id          INT4                 not null,
   product_id           INT4                 not null,
   sale_id              INT4                 null,
   order_date           DATE                 not null,
   order_quantity       INT4                 not null,
   order_status         VARCHAR(30)          not null,
   constraint PK_ORDER primary key (order_id)
);

create table receipt (
   receipt_id           SERIAL               not null,
   supplier_id          INT4                 not null,
   user_id              INT4                 not null,
   product_id           INT4                 not null,
   receipt_date         DATE                 not null,
   receipt_quantity     INT4                 not null,
   receipt_purchase_price DECIMAL(12,2)      not null,
   constraint PK_RECEIPT primary key (receipt_id)
);

/*========== Первичные индексы ==========*/

create unique index category_PK on category (category_id);
create unique index customer_PK on customer (customer_id);
create unique index supplier_PK on supplier (supplier_id);
create unique index user_PK on "user" (user_id);
create unique index product_PK on product (product_id);
create unique index sale_PK on sale (sale_id);
create unique index order_PK on "order" (order_id);
create unique index receipt_PK on receipt (receipt_id);

/*========== Индексы FK ==========*/

create index groups_FK on product (category_id);
create index places_FK on "order" (customer_id);
create index selected_in_FK on "order" (product_id);
create index becomes_FK on "order" (sale_id);
create index becomes2_FK on sale (order_id);
create index specified_in_FK on sale (product_id);
create index processes_FK on sale (user_id);
create index ships_FK on receipt (supplier_id);
create index records_FK on receipt (user_id);
create index stocked_FK on receipt (product_id);

/*========== Внешние ключи ==========*/

alter table product
   add constraint FK_PRODUCT_GROUPS_CATEGORY foreign key (category_id)
      references category (category_id) on delete restrict on update restrict;

alter table "order"
   add constraint FK_ORDER_PLACES_CUSTOMER foreign key (customer_id)
      references customer (customer_id) on delete restrict on update restrict;

alter table "order"
   add constraint FK_ORDER_SELECTED__PRODUCT foreign key (product_id)
      references product (product_id) on delete restrict on update restrict;

alter table "order"
   add constraint FK_ORDER_BECOMES_SALE foreign key (sale_id)
      references sale (sale_id) on delete restrict on update restrict;

alter table sale
   add constraint FK_SALE_BECOMES2_ORDER foreign key (order_id)
      references "order" (order_id) on delete restrict on update restrict;

alter table sale
   add constraint FK_SALE_PROCESSES_USER foreign key (user_id)
      references "user" (user_id) on delete restrict on update restrict;

alter table sale
   add constraint FK_SALE_SPECIFIED_PRODUCT foreign key (product_id)
      references product (product_id) on delete restrict on update restrict;

alter table receipt
   add constraint FK_RECEIPT_SHIPS_SUPPLIER foreign key (supplier_id)
      references supplier (supplier_id) on delete restrict on update restrict;

alter table receipt
   add constraint FK_RECEIPT_STOCKED_PRODUCT foreign key (product_id)
      references product (product_id) on delete restrict on update restrict;

alter table receipt
   add constraint FK_RECEIPT_RECORDS_USER foreign key (user_id)
      references "user" (user_id) on delete restrict on update restrict;

/*========== Вторичные индексы ==========*/

-- 1. Вход сотрудника
CREATE UNIQUE INDEX idx_user_login
    ON "user" (user_login);

-- 2. Вход / регистрация клиента
CREATE UNIQUE INDEX idx_customer_email
    ON customer (customer_email)
    WHERE customer_email IS NOT NULL;

-- 3. Каталог: товары в категории
CREATE INDEX idx_product_category_name
    ON product (category_id, product_name);

-- 4. Поиск товара по наименованию
CREATE INDEX idx_product_name
    ON product (product_name);

-- 5. Очередь заказов (продавец)
CREATE INDEX idx_order_status_date
    ON "order" (order_status, order_date);

-- 6. История заказов клиента
CREATE INDEX idx_order_customer_date
    ON "order" (customer_id, order_date DESC);

-- 7. Отчёт о продажах за период
CREATE INDEX idx_sale_date
    ON sale (sale_date);

-- 8. Продажи по товару за период
CREATE INDEX idx_sale_product_date
    ON sale (product_id, sale_date);

-- 9. Отчёт по поставщикам
CREATE INDEX idx_receipt_supplier_date
    ON receipt (supplier_id, receipt_date);

-- 10. Поступления за период
CREATE INDEX idx_receipt_date
    ON receipt (receipt_date);
