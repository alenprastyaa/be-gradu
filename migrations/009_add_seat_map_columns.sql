ALTER TABLE event_templates
ADD COLUMN IF NOT EXISTS seat_map_columns INT NOT NULL DEFAULT 20;

UPDATE event_templates
SET seat_map_columns = 20
WHERE seat_map_columns IS NULL OR seat_map_columns < 4 OR seat_map_columns > 40;
