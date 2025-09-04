# Payment Service

## Deskripsi
Payment Service adalah bagian dari sistem tiket event yang menangani semua operasi terkait pembayaran. Layanan ini bertanggung jawab untuk memproses pembayaran, mengelola refund, dan menyimpan catatan transaksi pembayaran.

## Fitur Utama
- Pembuatan dan pengelolaan transaksi pembayaran
- Pemrosesan pembayaran melalui berbagai metode (kartu kredit, PayPal, dll)
- Pengelolaan status pembayaran
- Refund pembayaran
- Integrasi dengan penyedia pembayaran (Stripe, PayPal, dll)
- Komunikasi dengan layanan lain melalui RabbitMQ

## Teknologi
- Go (Golang)
- Gin Web Framework
- GORM (ORM)
- PostgreSQL
- RabbitMQ
- Prometheus (Metrics)
- Logrus (Logging)

## Struktur Proyek
```
/payment-service
├── config/             # Konfigurasi database, RabbitMQ, dll
├── handler/            # HTTP handlers
├── middleware/         # Middleware (JWT, logging, metrics)
├── model/              # Model data
├── provider/           # Penyedia pembayaran (Stripe, PayPal, dll)
├── repository/         # Akses database
├── service/            # Logika bisnis
├── utils/              # Utilitas
├── .env                # Variabel lingkungan
├── Dockerfile          # Konfigurasi Docker
├── go.mod              # Dependensi Go
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
2. Salin `.env.example` ke `.env` dan sesuaikan nilai-nilainya
3. Jalankan `go mod download` untuk mengunduh dependensi
4. Jalankan `go run main.go` untuk memulai layanan

## API Endpoints

### Pembayaran
- `POST /api/payments` - Membuat pembayaran baru
- `POST /api/payments/process` - Memproses pembayaran
- `GET /api/payments/:id` - Mendapatkan detail pembayaran
- `GET /api/payments/booking/:bookingId` - Mendapatkan pembayaran berdasarkan ID booking
- `GET /api/payments/user` - Mendapatkan semua pembayaran pengguna
- `PUT /api/payments/:id/status` - Memperbarui status pembayaran (admin)
- `POST /api/payments/:id/refund` - Refund pembayaran (admin)

## Integrasi dengan Layanan Lain

### Event & Ticket Service
- Menerima notifikasi pembuatan booking
- Mengirim notifikasi status pembayaran

### Notification Service
- Mengirim notifikasi pembayaran berhasil/gagal
- Mengirim notifikasi refund

### User Service
- Validasi pengguna melalui JWT
- Menerima notifikasi penghapusan pengguna

## Pengembangan

### Menambahkan Penyedia Pembayaran Baru
1. Buat implementasi baru dari interface `PaymentProvider` di package `provider`
2. Daftarkan penyedia baru di `provider/payment_provider.go`

### Menjalankan Test
```
go test ./...
```

## Monitoring
- Metrics tersedia di endpoint `/metrics` (Prometheus)
- Health check tersedia di endpoint `/health`

## Lisensi
Hak Cipta © 2023