ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS school_id UUID;

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS position VARCHAR(150) NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS whatsapp_number VARCHAR(30) NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS email VARCHAR(150);

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS invitation_code VARCHAR(50);

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS qr_payload VARCHAR(120);

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS attendance_status VARCHAR(20) NOT NULL DEFAULT 'belum_hadir';

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS attendance_time TIMESTAMP NULL;

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE IF EXISTS teacher_invites
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE teacher_invites
SET
    position = COALESCE(position, ''),
    whatsapp_number = COALESCE(whatsapp_number, ''),
    attendance_status = COALESCE(NULLIF(attendance_status, ''), 'belum_hadir'),
    created_at = COALESCE(created_at, CURRENT_TIMESTAMP),
    updated_at = COALESCE(updated_at, CURRENT_TIMESTAMP)
WHERE TRUE;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'teacher_invites'
          AND column_name = 'school_id'
          AND is_nullable = 'YES'
    ) THEN
        DELETE FROM teacher_invites WHERE school_id IS NULL;
        ALTER TABLE teacher_invites ALTER COLUMN school_id SET NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'teacher_invites_school_id_fkey'
    ) THEN
        ALTER TABLE teacher_invites
            ADD CONSTRAINT teacher_invites_school_id_fkey
            FOREIGN KEY (school_id) REFERENCES schools(id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_invites_invitation_code ON teacher_invites(invitation_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_invites_qr_payload ON teacher_invites(qr_payload);
CREATE INDEX IF NOT EXISTS idx_teacher_invites_school_id ON teacher_invites(school_id);
CREATE INDEX IF NOT EXISTS idx_teacher_invites_school_name ON teacher_invites(school_id, name);
CREATE INDEX IF NOT EXISTS idx_teacher_invites_school_attendance ON teacher_invites(school_id, attendance_status);
