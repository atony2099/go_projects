CREATE TABLE `task_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `duration` int NOT NULL,
  `task` varchar(255) NOT NULL,
  `project` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_task_logs_deleted_at` (`deleted_at`)
)