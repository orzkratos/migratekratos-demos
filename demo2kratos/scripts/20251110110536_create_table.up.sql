CREATE TABLE `products`
(
    `id`          integer PRIMARY KEY AUTOINCREMENT,
    `created_at`  datetime,
    `updated_at`  datetime,
    `deleted_at`  datetime,
    `name`        varchar(150)   NOT NULL,
    `price`       decimal(10, 2) NOT NULL,
    `stock`       integer DEFAULT 0,
    `description` text
);

CREATE INDEX `idx_products_deleted_at` ON `products` (`deleted_at`);
