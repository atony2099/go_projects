CREATE TABLE `tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,

  `project_name` varchar(255) NOT NULL,
  `task_name` varchar(255) NOT NULL,
  `parent_name` varchar(255) NOT NULL,

  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,

  PRIMARY KEY (`id`),

  UNIQUE KEY `idx_project_name_task_name_parent_name` (`project_name`,`task_name`,`parent_name`),

  KEY `idx_deleted_at` (`deleted_at`)
);

UPDATE task_logs SET task='rpc' WHERE task='micro_rpc';
UPDATE task_logs SET task='architecture' where task='micro_architecture';

INSERT INTO tasks (project_name, task_name, parent_name, created_at, updated_at, deleted_at) VALUES ('interview1', 'basic', 'go', NOW(), NOW(), NULL);
INSERT INTO tasks (project_name, task_name, parent_name, created_at, updated_at, deleted_at) VALUES ('interview1', 'rpc', 'micro', NOW(), NOW(), NULL);
INSERT INTO tasks (project_name, task_name, parent_name, created_at, updated_at, deleted_at) VALUES ('interview1', 'architecture', 'micro', NOW(), NOW(), NULL);

ALTER TABLE task_logs ADD COLUMN task_id bigint unsigned NOT NULL;

UPDATE task_logs, tasks SET task_logs.task_id = tasks.id WHERE task_logs.task = tasks.task_name;


ALTER TABLE task_logs DROP COLUMN `task` ;
ALTER TABLE task_logs DROP COLUMN `project`;