CREATE TABLE `records`
(
    `id`         integer PRIMARY KEY AUTOINCREMENT,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `message`    varchar(255)
);

CREATE INDEX `idx_records_deleted_at` ON `records` (`deleted_at`);
