-- reverse -- CREATE INDEX `idx_articles_deleted_at` ON `articles`(`deleted_at`);
DROP INDEX IF EXISTS `idx_articles_deleted_at`;

-- reverse -- CREATE TABLE `articles` (`id` integer PRIMARY KEY AUTOINCREMENT,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`title` varchar(200) NOT NULL,`content` text,`author` varchar(100),`status` varchar(20) DEFAULT "draft");
DROP TABLE IF EXISTS `articles`;
