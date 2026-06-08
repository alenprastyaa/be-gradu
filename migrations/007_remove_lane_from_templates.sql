UPDATE event_settings
SET whatsapp_template = replace(replace(whatsapp_template, 'Lorong: {{lane_code}}' || chr(10), ''), 'Lorong: {{lane_code}}' || chr(13) || chr(10), ''),
    email_template = replace(replace(email_template, 'Lorong: {{lane_code}}' || chr(10), ''), 'Lorong: {{lane_code}}' || chr(13) || chr(10), '')
WHERE whatsapp_template LIKE '%{{lane_code}}%' OR email_template LIKE '%{{lane_code}}%';

UPDATE event_templates
SET whatsapp_template = replace(replace(whatsapp_template, 'Lorong: {{lane_code}}' || chr(10), ''), 'Lorong: {{lane_code}}' || chr(13) || chr(10), ''),
    email_template = replace(replace(email_template, 'Lorong: {{lane_code}}' || chr(10), ''), 'Lorong: {{lane_code}}' || chr(13) || chr(10), '')
WHERE whatsapp_template LIKE '%{{lane_code}}%' OR email_template LIKE '%{{lane_code}}%';
