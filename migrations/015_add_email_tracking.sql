ALTER TABLE students
ADD COLUMN IF NOT EXISTS email_sent_at TIMESTAMP NULL,
ADD COLUMN IF NOT EXISTS email_brevo_message_id TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_students_email_sent_at ON students(email_sent_at);
CREATE INDEX IF NOT EXISTS idx_students_email_brevo_message_id ON students(email_brevo_message_id);

CREATE TABLE IF NOT EXISTS email_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NULL REFERENCES students(id) ON DELETE SET NULL,
    email VARCHAR(255) NOT NULL,
    message_id TEXT NULL,
    subject TEXT NULL,
    event VARCHAR(50) NOT NULL,
    event_time TIMESTAMP NULL,
    reason TEXT NULL,
    link TEXT NULL,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_logs_student_id ON email_logs(student_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_email ON email_logs(email);
CREATE INDEX IF NOT EXISTS idx_email_logs_message_id ON email_logs(message_id);
CREATE INDEX IF NOT EXISTS idx_email_logs_event ON email_logs(event);
CREATE INDEX IF NOT EXISTS idx_email_logs_event_time ON email_logs(event_time);
