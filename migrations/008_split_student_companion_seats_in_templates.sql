UPDATE event_settings
SET whatsapp_template = replace(whatsapp_template, 'Bangku Siswa & Pendamping: {{seat_number}}', 'Nomor Siswa: {{student_seat_number}}' || chr(10) || 'Nomor Pendamping: {{companion_seat_number}}'),
    email_template = replace(email_template, 'Bangku Siswa & Pendamping: {{seat_number}}', 'Nomor Siswa: {{student_seat_number}}' || chr(10) || 'Nomor Pendamping: {{companion_seat_number}}')
WHERE whatsapp_template LIKE '%Bangku Siswa & Pendamping: {{seat_number}}%'
   OR email_template LIKE '%Bangku Siswa & Pendamping: {{seat_number}}%';

UPDATE event_templates
SET whatsapp_template = replace(whatsapp_template, 'Bangku Siswa & Pendamping: {{seat_number}}', 'Nomor Siswa: {{student_seat_number}}' || chr(10) || 'Nomor Pendamping: {{companion_seat_number}}'),
    email_template = replace(email_template, 'Bangku Siswa & Pendamping: {{seat_number}}', 'Nomor Siswa: {{student_seat_number}}' || chr(10) || 'Nomor Pendamping: {{companion_seat_number}}')
WHERE whatsapp_template LIKE '%Bangku Siswa & Pendamping: {{seat_number}}%'
   OR email_template LIKE '%Bangku Siswa & Pendamping: {{seat_number}}%';
