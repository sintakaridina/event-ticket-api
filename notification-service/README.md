# Notification Service

## Deskripsi
Notification Service adalah bagian dari sistem microservices Event Ticket Platform yang bertanggung jawab untuk mengirim notifikasi kepada pengguna melalui berbagai saluran seperti email, SMS, dan push notification. Layanan ini menerima permintaan notifikasi dari layanan lain melalui RabbitMQ dan mengirimkannya ke pengguna menggunakan provider yang sesuai.

## Fitur Utama
- Pengiriman notifikasi melalui email, SMS, dan push notification
- Manajemen template notifikasi
- Pelacakan status pengiriman notifikasi
- Integrasi dengan layanan lain melalui RabbitMQ
- Penjadwalan notifikasi

## Teknologi
- Go (Golang)
- Gin Web Framework
- GORM (ORM untuk PostgreSQL)
- PostgreSQL
- RabbitMQ
- Prometheus (untuk metrik)
- Logrus (untuk logging)
- SMTP (untuk email)
- Twilio (untuk SMS)
- Firebase Cloud Messaging (untuk push notification)

## Struktur Proyek
```
notification-service/
├── config/             # Konfigurasi database, RabbitMQ, dll
├── handler/            # HTTP handlers
├── middleware/         # Middleware Gin (JWT, logging, metrics)
├── model/              # Model data dan struktur request/response
├── provider/           # Provider untuk email, SMS, dan push notification
├── repository/         # Akses database
├── service/            # Logika bisnis
├── utils/              # Utilitas umum
├── .env                # File konfigurasi lingkungan
├── Dockerfile          # Untuk containerization
├── go.mod              # Dependensi Go
├── go.sum              # Checksum dependensi
├── main.go             # Entry point aplikasi
└── README.md           # Dokumentasi
```

## Instalasi

### Prasyarat
- Go 1.19 atau lebih tinggi
- PostgreSQL
- RabbitMQ
- Docker (opsional)

### Langkah-langkah
1. Clone repositori
   ```bash
   git clone https://github.com/yourusername/trae.git
   cd trae/notification-service
   ```

2. Salin file .env.example ke .env dan sesuaikan konfigurasi
   ```bash
   cp .env.example .env
   # Edit file .env sesuai kebutuhan
   ```

3. Instal dependensi
   ```bash
   go mod download
   ```

4. Jalankan aplikasi
   ```bash
   go run main.go
   ```

### Menggunakan Docker
1. Build image Docker
   ```bash
   docker build -t notification-service .
   ```

2. Jalankan container
   ```bash
   docker run -p 8083:8083 --env-file .env notification-service
   ```

## API Endpoints

### Notifikasi
- `POST /api/v1/notifications` - Membuat notifikasi baru
- `POST /api/v1/notifications/send` - Mengirim notifikasi
- `GET /api/v1/notifications/:id` - Mendapatkan notifikasi berdasarkan ID
- `GET /api/v1/notifications/user/:userId` - Mendapatkan notifikasi berdasarkan ID pengguna
- `GET /api/v1/notifications/code/:code` - Mendapatkan notifikasi berdasarkan kode
- `PUT /api/v1/notifications/:id/status` - Memperbarui status notifikasi
- `DELETE /api/v1/notifications/:id` - Menghapus notifikasi

### Template
- `POST /api/v1/templates` - Membuat template baru
- `GET /api/v1/templates/:id` - Mendapatkan template berdasarkan ID
- `GET /api/v1/templates/code/:code` - Mendapatkan template berdasarkan kode
- `PUT /api/v1/templates/:id` - Memperbarui template
- `DELETE /api/v1/templates/:id` - Menghapus template

### Lainnya
- `GET /health` - Health check
- `GET /metrics` - Metrik Prometheus

## Integrasi dengan Layanan Lain

### Payment Service
Menerima event pembayaran untuk mengirim notifikasi tentang status pembayaran (berhasil, gagal, refund).

### Event & Ticket Service
Menerima event tiket untuk mengirim notifikasi tentang pemesanan tiket, pembatalan, dan pengingat acara.

### User Service
Menerima event pengguna untuk mengirim notifikasi tentang pendaftaran, reset password, dan perubahan profil.

## Pengembangan

### Menambahkan Provider Notifikasi Baru
1. Buat interface baru di `provider/` jika diperlukan
2. Implementasikan provider baru
3. Perbarui factory method untuk membuat instance provider

### Menambahkan Template Baru
Template dapat ditambahkan melalui API atau secara langsung di database. Template menggunakan format placeholder `{{variable_name}}` yang akan diganti dengan nilai sebenarnya saat notifikasi dikirim.

## Monitoring

### Logging
Layanan menggunakan Logrus untuk logging. Level log dapat dikonfigurasi melalui variabel lingkungan `LOG_LEVEL`.

### Metrik
Metrik Prometheus tersedia di endpoint `/metrics` dan mencakup:
- Jumlah permintaan HTTP
- Durasi permintaan HTTP
- Jumlah notifikasi yang dikirim (berdasarkan saluran dan status)
- Ukuran permintaan dan respons

## Lisensi
Proyek ini dilisensikan di bawah [MIT License](LICENSE).