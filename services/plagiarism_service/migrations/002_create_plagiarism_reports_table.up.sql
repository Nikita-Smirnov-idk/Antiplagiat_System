CREATE TABLE plagiarism_reports (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(50) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    student_a VARCHAR(50) NOT NULL,
    student_b VARCHAR(50) NOT NULL,
    similarity FLOAT PRECISION NOT NULL ,
    file_a_handed_over_at TIMESTAMP NOT NULL,
    file_b_handed_over_at TIMESTAMP NOT NULL,

    UNIQUE(task_id, student, student_with_similar_file)
);

CREATE INDEX idx_plagiarism_reports_task_id ON plagiarism_reports(task_id);
CREATE INDEX idx_plagiarism_reports_student_a ON plagiarism_reports(student_a);
CREATE INDEX idx_plagiarism_reports_student_b ON plagiarism_reports(student_b);