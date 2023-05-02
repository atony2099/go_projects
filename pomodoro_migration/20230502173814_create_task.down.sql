
ALTER TABLE task_logs ADD COLUMN task varchar(255) NOT NULL;

UPDATE task_logs, tasks SET task_logs.task = tasks.task_name WHERE task_logs.task_id = tasks.id;

ALTER TABLE task_logs DROP COLUMN task_id;

DROP TABLE `tasks`;