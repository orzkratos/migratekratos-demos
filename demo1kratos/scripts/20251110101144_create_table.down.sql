-- reverse -- CREATE INDEX `idx_records_deleted_at` ON `records`(`deleted_at`);
DROP INDEX IF EXISTS `idx_records_deleted_at`;

-- reverse -- CREATE TABLE `records` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`message` varchar(255));
DROP TABLE IF EXISTS `records`;
