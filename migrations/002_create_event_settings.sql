CREATE TABLE IF NOT EXISTS event_settings (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE,
    event_title VARCHAR(255) NOT NULL DEFAULT 'Graduation Ceremony 2026',
    school_name VARCHAR(255) NOT NULL DEFAULT 'Nama Sekolah',
    graduation_year VARCHAR(20) NOT NULL DEFAULT '2026',
    recipient_greeting TEXT NOT NULL DEFAULT 'Yth. Siswa/i dan Orang Tua/Wali',
    opening_text TEXT NOT NULL DEFAULT 'Dengan hormat, kami mengundang Anda untuk menghadiri acara wisuda sebagai bentuk apresiasi atas pencapaian dan perjalanan pendidikan siswa/i.',
    event_date VARCHAR(100) NOT NULL DEFAULT 'Sabtu, 20 Juni 2026',
    event_time VARCHAR(100) NOT NULL DEFAULT '08.00 WIB - selesai',
    venue_name VARCHAR(255) NOT NULL DEFAULT 'Aula Utama Sekolah',
    venue_address TEXT NOT NULL DEFAULT 'Jl. Pendidikan No. 1',
    maps_url TEXT NOT NULL DEFAULT '',
    dress_code VARCHAR(150) NOT NULL DEFAULT 'Formal rapi',
    additional_note TEXT NOT NULL DEFAULT 'Mohon hadir 30 menit sebelum acara dimulai dan tunjukkan QR Code kepada petugas registrasi.',
    whatsapp_template TEXT NOT NULL DEFAULT 'Assalamu''alaikum Wr. Wb.

{{recipient_greeting}},

Dengan hormat, kami mengundang {{student_name}} untuk menghadiri:

*{{event_title}}*
{{school_name}}

Hari/Tanggal: {{event_date}}
Waktu: {{event_time}}
Tempat: {{venue_name}}
Alamat: {{venue_address}}
Dress Code: {{dress_code}}

Data Undangan:
Nama: {{student_name}}
Kelas: {{class_name}}
Jurusan: {{major}}
Nomor Siswa: {{student_seat_number}}
Nomor Pendamping: {{companion_seat_number}}

Undangan digital dan QR Code:
{{invitation_link}}

{{additional_note}}

Wassalamu''alaikum Wr. Wb.',
    email_subject VARCHAR(255) NOT NULL DEFAULT 'Undangan {{event_title}} - {{student_name}}',
    email_template TEXT NOT NULL DEFAULT '{{recipient_greeting}},

{{opening_text}}

Acara:
{{event_title}}
{{school_name}}

Hari/Tanggal: {{event_date}}
Waktu: {{event_time}}
Tempat: {{venue_name}}
Alamat: {{venue_address}}
Dress Code: {{dress_code}}

Data Undangan:
Nama: {{student_name}}
Kelas: {{class_name}}
Jurusan: {{major}}
Nomor Siswa: {{student_seat_number}}
Nomor Pendamping: {{companion_seat_number}}

Silakan buka undangan digital dan QR Code melalui tautan berikut:
{{invitation_link}}

{{additional_note}}

Hormat kami,
Panitia {{event_title}}',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT event_settings_singleton CHECK (id = TRUE)
);

INSERT INTO event_settings (id)
VALUES (TRUE)
ON CONFLICT (id) DO NOTHING;
