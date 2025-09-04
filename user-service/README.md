# User Service

## Deskripsi
User Service adalah bagian dari sistem microservices Event Ticket Platform yang bertanggung jawab untuk mengelola pengguna, autentikasi, dan otorisasi. Layanan ini menyediakan API untuk pendaftaran, login, manajemen profil, dan pengelolaan peran pengguna.

## Fitur Utama
- Pendaftaran dan login pengguna
- Autentikasi dengan JWT
- Manajemen profil pengguna
- Pengelolaan peran dan izin
- Reset dan perubahan password
- Verifikasi email

## Teknologi
- Go (Golang)
- Gin Web Framework
- GORM (ORM untuk PostgreSQL)
- PostgreSQL
- RabbitMQ
- Prometheus (untuk metrik)
- Logrus (untuk logging)
- JWT untuk autentikasi

## Struktur Proyek
```
user-service/
├── config/             # Konfigurasi database, RabbitMQ, dll
├── handler/            # HTTP handlers
├── middleware/         # Middleware Gin (JWT, logging, metrics)
├── model/              # Model data dan struktur request/response
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
   cd trae/user-service
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
   docker build -t user-service .
   ```

2. Jalankan container
   ```bash
   docker run -p 8081:8081 --env-file .env user-service
   ```

## API Endpoints

### Autentikasi
- `POST /api/v1/auth/register` - Mendaftarkan pengguna baru
- `POST /api/v1/auth/login` - Login pengguna
- `POST /api/v1/auth/refresh` - Memperbaharui token JWT
- `POST /api/v1/auth/logout` - Logout pengguna
- `POST /api/v1/auth/forgot-password` - Meminta reset password
- `POST /api/v1/auth/reset-password` - Reset password
- `POST /api/v1/auth/verify-email` - Verifikasi email

### Pengguna
- `GET /api/v1/users/me` - Mendapatkan profil pengguna saat ini
- `PUT /api/v1/users/me` - Memperbarui profil pengguna saat ini
- `PUT /api/v1/users/me/password` - Mengubah password
- `GET /api/v1/users/:id` - Mendapatkan pengguna berdasarkan ID
- `GET /api/v1/users` - Mendapatkan daftar pengguna (admin)
- `PUT /api/v1/users/:id` - Memperbarui pengguna (admin)
- `DELETE /api/v1/users/:id` - Menghapus pengguna (admin)

### Peran
- `POST /api/v1/roles` - Membuat peran baru (admin)
- `GET /api/v1/roles/:id` - Mendapatkan peran berdasarkan ID (admin)
- `GET /api/v1/roles` - Mendapatkan daftar peran (admin)
- `PUT /api/v1/roles/:id` - Memperbarui peran (admin)
- `DELETE /api/v1/roles/:id` - Menghapus peran (admin)
- `POST /api/v1/users/:id/roles` - Menetapkan peran ke pengguna (admin)
- `DELETE /api/v1/users/:id/roles/:roleId` - Menghapus peran dari pengguna (admin)

### Lainnya
- `GET /health` - Health check
- `GET /metrics` - Metrik Prometheus

## Integrasi dengan Layanan Lain

### Payment Service
Menyediakan informasi pengguna untuk verifikasi pembayaran dan otorisasi.

### Event & Ticket Service
Menyediakan informasi pengguna untuk pemesanan tiket dan manajemen acara.

### Notification Service
Mengirim event pengguna untuk notifikasi seperti pendaftaran, reset password, dan perubahan profil.

## Pengembangan

### Menambahkan Endpoint Baru
1. Definisikan model request/response di `model/`
2. Implementasikan logika bisnis di `service/`
3. Tambahkan handler di `handler/`
4. Daftarkan rute baru di `handler/setup.go`

### Menambahkan Peran Baru
Peran dapat ditambahkan melalui API atau secara langsung di database. Setiap peran memiliki set izin yang menentukan tindakan apa yang dapat dilakukan oleh pengguna dengan peran tersebut.

## Monitoring

### Logging
Layanan menggunakan Logrus untuk logging. Level log dapat dikonfigurasi melalui variabel lingkungan `LOG_LEVEL`.

### Metrik
Metrik Prometheus tersedia di endpoint `/metrics` dan mencakup:
- Jumlah permintaan HTTP
- Durasi permintaan HTTP
- Jumlah login berhasil/gagal
- Jumlah pendaftaran pengguna
- Ukuran permintaan dan respons

## Keamanan
- Semua password di-hash menggunakan bcrypt
- Autentikasi menggunakan JWT dengan refresh token
- Validasi input untuk semua permintaan
- Rate limiting untuk mencegah brute force
- Proteksi CSRF untuk endpoint sensitif

## Lisensi
Proyek ini dilisensikan di bawah [MIT License](LICENSE).