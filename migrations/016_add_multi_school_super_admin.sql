CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS schools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    address TEXT NOT NULL DEFAULT '',
    logo_url TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schools (name, slug)
VALUES ('Sekolah Default', 'default-school')
ON CONFLICT (slug) DO NOTHING;

ALTER TABLE admins
ADD COLUMN IF NOT EXISTS role VARCHAR(30) NOT NULL DEFAULT 'school_admin',
ADD COLUMN IF NOT EXISTS school_id UUID NULL REFERENCES schools(id) ON DELETE RESTRICT;

UPDATE admins
SET role = 'school_admin'
WHERE role IS NULL OR role = '' OR role = 'admin';

UPDATE admins
SET school_id = (SELECT id FROM schools WHERE slug = 'default-school' LIMIT 1)
WHERE role = 'school_admin' AND school_id IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'admins_role_check'
    ) THEN
        ALTER TABLE admins
        ADD CONSTRAINT admins_role_check CHECK (role IN ('super_admin', 'school_admin'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_admins_role ON admins(role);
CREATE INDEX IF NOT EXISTS idx_admins_school_id ON admins(school_id);

ALTER TABLE students
ADD COLUMN IF NOT EXISTS school_id UUID NULL REFERENCES schools(id) ON DELETE RESTRICT;

UPDATE students
SET school_id = (SELECT id FROM schools WHERE slug = 'default-school' LIMIT 1)
WHERE school_id IS NULL;

ALTER TABLE students
ALTER COLUMN school_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_students_school_id ON students(school_id);

ALTER TABLE seating_lanes
ADD COLUMN IF NOT EXISTS school_id UUID NULL REFERENCES schools(id) ON DELETE RESTRICT;

UPDATE seating_lanes
SET school_id = (SELECT id FROM schools WHERE slug = 'default-school' LIMIT 1)
WHERE school_id IS NULL;

ALTER TABLE seating_lanes
ALTER COLUMN school_id SET NOT NULL;

ALTER TABLE seating_lanes DROP CONSTRAINT IF EXISTS seating_lanes_lane_code_key;
ALTER TABLE seating_lanes DROP CONSTRAINT IF EXISTS seating_lanes_class_name_major_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_seating_lanes_school_lane_code ON seating_lanes(school_id, lane_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_seating_lanes_school_class_major ON seating_lanes(school_id, class_name, major);

ALTER TABLE event_templates
ADD COLUMN IF NOT EXISTS school_id UUID NULL REFERENCES schools(id) ON DELETE RESTRICT;

UPDATE event_templates
SET school_id = (SELECT id FROM schools WHERE slug = 'default-school' LIMIT 1)
WHERE school_id IS NULL;

ALTER TABLE event_templates
ALTER COLUMN school_id SET NOT NULL;

DROP INDEX IF EXISTS idx_event_templates_single_active;
CREATE UNIQUE INDEX IF NOT EXISTS idx_event_templates_school_single_active
ON event_templates (school_id)
WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_event_templates_school_id ON event_templates(school_id);

ALTER TABLE event_settings
ADD COLUMN IF NOT EXISTS school_id UUID NULL REFERENCES schools(id) ON DELETE RESTRICT;

UPDATE event_settings
SET school_id = (SELECT id FROM schools WHERE slug = 'default-school' LIMIT 1)
WHERE school_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_event_settings_school_id ON event_settings(school_id);

ALTER TABLE email_logs
ADD COLUMN IF NOT EXISTS school_id UUID NULL REFERENCES schools(id) ON DELETE SET NULL;

UPDATE email_logs el
SET school_id = students.school_id
FROM students
WHERE el.student_id = students.id
  AND el.school_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_email_logs_school_id ON email_logs(school_id);
