-- Create dag table
CREATE TABLE dag (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'created',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create executor table
CREATE TABLE executor (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    config JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create task table
CREATE TABLE task (
    id SERIAL PRIMARY KEY,
    dag_id INT REFERENCES dag(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'created',
    parent_id INT REFERENCES task(id) ON DELETE SET NULL,
    executor_id INT REFERENCES executor(id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    definition JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_dag FOREIGN KEY (dag_id) REFERENCES dag (id),
    CONSTRAINT fk_parent_task FOREIGN KEY (parent_id) REFERENCES task (id),
    CONSTRAINT fk_executor FOREIGN KEY (executor_id) REFERENCES executor (id)
);

-- Add index to the status column of the task table
CREATE INDEX idx_task_status ON task (status);

-- Add index to the type column of the task table
CREATE INDEX idx_task_type ON task (type);

