# API Gateway

## Deskripsi
API Gateway adalah komponen utama dalam sistem microservices Event Ticket Platform yang bertindak sebagai titik masuk tunggal untuk semua permintaan klien. Gateway ini mengarahkan permintaan ke layanan yang sesuai, menangani autentikasi, dan menyediakan fungsionalitas umum seperti rate limiting, caching, dan logging.

## Fitur Utama
- Routing permintaan ke microservices yang sesuai
- Autentikasi dan otorisasi terpusat
- Rate limiting untuk mencegah penyalahgunaan API
- Caching untuk meningkatkan performa
- Logging dan monitoring terpusat
- Load balancing
- Circuit breaking untuk ketahanan sistem

## Teknologi
- Go (Golang)
- Gin Web Framework
- Redis (untuk caching dan rate limiting)
- Prometheus (untuk metrik)
- Logrus (untuk logging)
- JWT untuk autentikasi

## Struktur Proyek
```
api-gateway/
├── config/             # Konfigurasi layanan, routing, dll
├── handler/            # HTTP handlers
├── middleware/         # Middleware Gin (JWT, logging, metrics, rate limiting)
├── model/              # Model data dan struktur request/response
├── proxy/              # Proxy untuk microservices
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
- Redis
- Docker (opsional)

### Langkah-langkah
1. Clone repositori
   ```bash
   git clone https://github.com/yourusername/trae.git
   cd trae/api-gateway
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
   docker build -t api-gateway .
   ```

2. Jalankan container
   ```bash
   docker run -p 8080:8080 --env-file .env api-gateway
   ```

## Konfigurasi Routing

API Gateway mengarahkan permintaan ke microservices berdasarkan konfigurasi yang ditentukan di `config/routes.go`. Berikut adalah contoh konfigurasi routing:

```go
var Routes = []RouteConfig{
    {
        Path:        "/api/v1/users/*",
        ServiceName: "user-service",
        ServiceURL:  "http://user-service:8081",
        Methods:     []string{"GET", "POST", "PUT", "DELETE"},
        Protected:   true,
    },
    {
        Path:        "/api/v1/auth/*",
        ServiceName: "user-service",
        ServiceURL:  "http://user-service:8081",
        Methods:     []string{"POST"},
        Protected:   false,
    },
    {
        Path:        "/api/v1/events/*",
        ServiceName: "event-ticket-service",
        ServiceURL:  "http://event-ticket-service:8082",
        Methods:     []string{"GET", "POST", "PUT", "DELETE"},
        Protected:   true,
    },
    // ... dan seterusnya
}
```

## Integrasi dengan Microservices

### User Service
Menangani permintaan terkait pengguna, autentikasi, dan otorisasi.

### Event & Ticket Service
Menangani permintaan terkait acara, tiket, dan pemesanan.

### Payment Service
Menangani permintaan terkait pembayaran dan transaksi.

### Notification Service
Menangani permintaan terkait notifikasi dan preferensi notifikasi.

## Pengembangan

### Menambahkan Rute Baru
1. Tambahkan konfigurasi rute baru di `config/routes.go`
2. Jika diperlukan, tambahkan middleware khusus di `middleware/`
3. Jika diperlukan, tambahkan handler khusus di `handler/`

### Menambahkan Microservice Baru
1. Tambahkan konfigurasi layanan baru di `config/services.go`
2. Tambahkan konfigurasi rute untuk layanan baru di `config/routes.go`
3. Jika diperlukan, tambahkan proxy khusus di `proxy/`

## Monitoring

### Logging
API Gateway menggunakan Logrus untuk logging. Level log dapat dikonfigurasi melalui variabel lingkungan `LOG_LEVEL`.

### Metrik
Metrik Prometheus tersedia di endpoint `/metrics` dan mencakup:
- Jumlah permintaan HTTP per layanan
- Durasi permintaan HTTP per layanan
- Jumlah error per layanan
- Rate limiting metrics
- Ukuran permintaan dan respons

## Keamanan
- Autentikasi menggunakan JWT
- Rate limiting untuk mencegah penyalahgunaan API
- Validasi input untuk semua permintaan
- Proteksi CSRF untuk endpoint sensitif
- CORS konfigurasi

## Lisensi
Proyek ini dilisensikan di bawah [MIT License](LICENSE).