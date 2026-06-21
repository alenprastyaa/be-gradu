ALTER TABLE event_templates
ADD COLUMN IF NOT EXISTS dress_code_student VARCHAR(150) NOT NULL DEFAULT 'Seragam sekolah / formal rapi',
ADD COLUMN IF NOT EXISTS dress_code_parent VARCHAR(150) NOT NULL DEFAULT 'Batik / formal rapi';

UPDATE event_templates
SET
  dress_code_student = CASE
    WHEN trim(coalesce(dress_code_student, '')) = '' THEN coalesce(NULLIF(trim(dress_code), ''), 'Seragam sekolah / formal rapi')
    ELSE dress_code_student
  END,
  dress_code_parent = CASE
    WHEN trim(coalesce(dress_code_parent, '')) = '' THEN coalesce(NULLIF(trim(dress_code), ''), 'Batik / formal rapi')
    ELSE dress_code_parent
  END;
