CREATE TABLE files (
   id VARCHAR(36) PRIMARY KEY,

   student_id VARCHAR(50) NOT NULL,
   task_id VARCHAR(50) NOT NULL,

   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
   status VARCHAR(20) DEFAULT 'uploaded',

   UNIQUE(student_id, task_id)
);