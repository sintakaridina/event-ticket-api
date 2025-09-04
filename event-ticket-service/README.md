# Event & Ticket Service

Event & Ticket Service adalah bagian dari sistem Trae yang menangani manajemen acara dan pemesanan tiket. Layanan ini memungkinkan pengguna untuk mencari acara, melihat detail acara, dan memesan tiket untuk acara tersebut.

## Fitur Utama

- Manajemen acara (membuat, mengupdate, menghapus, mencari)
- Manajemen tiket (membuat, mengupdate, menghapus)
- Pemesanan tiket (membuat, membatalkan, mengupdate status)
- Integrasi dengan layanan lain melalui RabbitMQ

## Teknologi

- Go (Golang)
- Gin Web Framework
- GORM (ORM)
- PostgreSQL
- RabbitMQ
- Prometheus (untuk metrik)
- Logrus (untuk logging)

## Struktur Proyek

```
.
├── config/             # Konfigurasi database, RabbitMQ, dll.
├── handler/            # HTTP handlers
├── middleware/         # Middleware (JWT, logging, metrics)
├── model/              # Model data
├── repository/         # Akses database
├── service/            # Logika bisnis
├── utils/              # Utilitas
├── .env                # File konfigurasi lingkungan
├── main.go             # Entry point aplikasi
└── README.md           # Dokumentasi
```

## Instalasi

### Prasyarat

- Go 1.16 atau lebih tinggi
- PostgreSQL
- RabbitMQ

### Langkah-langkah

1. Clone repositori
   ```bash
   git clone https://github.com/yourusername/ticket-system.git
   cd ticket-system/event-ticket-service
   ```

2. Instal dependensi
   ```bash
   go mod download
   ```

3. Konfigurasi lingkungan
   ```bash
   cp .env.example .env
   # Edit .env sesuai dengan konfigurasi lokal Anda
   ```

4. Jalankan aplikasi
   ```bash
   go run main.go
   ```

## API Endpoints

### Acara

- `GET /api/events` - Mendapatkan semua acara
- `GET /api/events/search` - Mencari acara
- `GET /api/events/:id` - Mendapatkan detail acara
- `POST /api/events` - Membuat acara baru (admin)
- `PUT /api/events/:id` - Mengupdate acara (admin)
- `DELETE /api/events/:id` - Menghapus acara (admin)

### Pemesanan

- `POST /api/bookings` - Membuat pemesanan baru
- `GET /api/bookings/user` - Mendapatkan pemesanan pengguna
- `GET /api/bookings/:id` - Mendapatkan detail pemesanan
- `PUT /api/bookings/:id/status` - Mengupdate status pemesanan (admin)
- `POST /api/bookings/:id/cancel` - Membatalkan pemesanan

## Integrasi dengan Layanan Lain

Event & Ticket Service terintegrasi dengan layanan lain melalui RabbitMQ:

- **User Service**: Menerima peristiwa terkait pengguna
- **Payment Service**: Menerima peristiwa terkait pembayaran
- **Notification Service**: Mengirim peristiwa untuk notifikasi

## Pengembangan

### Menjalankan Test

```bash
go test ./...
```

### Membangun untuk Produksi

```bash
go build -o event-ticket-service
```

## Kontribusi

Kontribusi selalu diterima! Silakan buat pull request atau buka issue untuk diskusi.

## Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE).