ALTER TABLE tasks
DROP CONSTRAINT tasks_context_id_fkey,
ADD CONSTRAINT tasks_context_id_fkey
FOREIGN KEY (context_id)
REFERENCES contexts(id)
ON DELETE SET NULL;
