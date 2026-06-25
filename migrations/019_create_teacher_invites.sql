CREATE TABLE IF NOT EXISTS teacher_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    position VARCHAR(150) NOT NULL DEFAULT '',
    whatsapp_number VARCHAR(30) NOT NULL DEFAULT '',
    email VARCHAR(150),
    invitation_code VARCHAR(50) NOT NULL UNIQUE,
    qr_payload VARCHAR(120) NOT NULL UNIQUE,
    attendance_status VARCHAR(20) NOT NULL DEFAULT 'belum_hadir',
    attendance_time TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_teacher_invites_school_id ON teacher_invites(school_id);
CREATE INDEX IF NOT EXISTS idx_teacher_invites_school_name ON teacher_invites(school_id, name);
CREATE INDEX IF NOT EXISTS idx_teacher_invites_school_attendance ON teacher_invites(school_id, attendance_status);
