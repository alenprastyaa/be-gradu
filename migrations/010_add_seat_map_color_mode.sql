ALTER TABLE event_templates
ADD COLUMN IF NOT EXISTS seat_map_color_mode VARCHAR(20) NOT NULL DEFAULT 'attendance';

UPDATE event_templates
SET seat_map_color_mode = 'attendance'
WHERE seat_map_color_mode NOT IN ('attendance', 'class');
