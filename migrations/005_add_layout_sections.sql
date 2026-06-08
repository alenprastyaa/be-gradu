-- Ordered, drag-and-drop layout sections for invitation templates (JSON array string)
ALTER TABLE event_templates
    ADD COLUMN IF NOT EXISTS layout_sections TEXT NOT NULL DEFAULT '';
