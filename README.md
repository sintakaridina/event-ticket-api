# Sistem Pemesanan Tiket Event (Event Ticket Booking System)

Sistem pemesanan tiket event dengan arsitektur microservices menggunakan Golang dan Gin framework.

## Arsitektur Sistem

Sistem ini menggunakan arsitektur microservices dengan komponen-komponen berikut:

```
+----------------+     +----------------+     +----------------+
|                |     |                |     |                |
|  API Gateway   |---->|  User Service  |---->|  PostgreSQL    |
|                |     |                |     |  Database      |
+----------------+     +----------------+     +----------------+
        |                      |
        |                      |                +----------------+
        |                      |                |                |
        |              +----------------+       |   RabbitMQ     |
        |              |                |       |   Message      |
        |              |  Notification  |<------|   Queue       |
        |              |  Service       |       |                |
        |              +----------------+       +----------------+
        |                                              ^
        v                                              |
+----------------+     +----------------+               |
|                |     |                |               |
|  Event &       |---->|  Payment       |---------------+
|  Ticket Service|     |  Service       |
+----------------+     +----------------+
```

## Komponen Utama

1. **User Service**: Mengelola registrasi pengguna, autentikasi, dan profil
2. **Event & Ticket Service**: Mengelola informasi event dan pemesanan tiket
3. **Payment Service**: Mengelola proses pembayaran
4. **Notification Service**: Mengirim notifikasi ke pengguna
5. **API Gateway**: Menyediakan titik akses tunggal untuk semua layanan
6. **PostgreSQL**: Database untuk menyimpan data
7. **RabbitMQ**: Message queue untuk komunikasi asinkron antar layanan

## Fitur Utama

- Registrasi dan login pengguna
- Pencarian event
- Pemesanan tiket
- Pembayaran
- Notifikasi (email, SMS)

## Teknologi yang Digunakan

- **Backend**: Go (Golang), Gin Framework
- **Database**: PostgreSQL
- **Message Queue**: RabbitMQ
- **Containerization**: Docker, Docker Compose
- **Monitoring**: Prometheus, Grafana
- **Logging**: Logrus
- **Authentication**: JWT (JSON Web Tokens)
- **Documentation**: Swagger/OpenAPI

## Struktur Proyek

```
trae/
├── api-gateway/            # API Gateway service
├── user-service/           # User management service
├── event-ticket-service/   # Event and ticket management service
├── payment-service/        # Payment processing service
├── notification-service/   # Notification service
├── docker-compose.yml      # Docker Compose configuration
├── prometheus.yml          # Prometheus configuration
└── README.md               # Project documentation
```

## Instalasi dan Penggunaan

### Prasyarat

- Go 1.19 atau lebih tinggi
- Docker dan Docker Compose
- PostgreSQL
- RabbitMQ

### Langkah-langkah Instalasi

1. Clone repositori
   ```bash
   git clone https://github.com/yourusername/trae.git
   cd trae
   ```

2. Konfigurasi file .env untuk setiap service
   ```bash
   # Contoh untuk user-service
   cd user-service
   cp .env.example .env
   # Edit file .env sesuai kebutuhan
   ```

3. Build dan jalankan dengan Docker Compose
   ```bash
   # Build dan jalankan semua service untuk production
   docker-compose up -d
   
   # Atau gunakan Makefile
   make up
   
   # Untuk development (hanya infrastruktur)
   docker-compose -f docker-compose.dev.yml up -d
   ```

4. Akses API Gateway
   ```
   http://localhost:8080
   ```

### Pengembangan Lokal

1. Jalankan PostgreSQL dan RabbitMQ
   ```bash
   docker-compose up -d postgres rabbitmq
   ```

2. Jalankan service yang ingin dikembangkan
   ```bash
   cd user-service
   go run main.go
   ```

## Dokumentasi API

Dokumentasi API tersedia melalui Swagger UI di endpoint berikut setelah menjalankan sistem:

```
http://localhost:8090
```

Swagger UI menyediakan antarmuka komprehensif untuk menjelajahi dan menguji semua endpoint API yang tersedia.

## Monitoring

Metrik Prometheus tersedia di endpoint berikut untuk setiap service:

- API Gateway: http://localhost:8080/metrics
- User Service: http://localhost:8081/metrics
- Event & Ticket Service: http://localhost:8082/metrics
- Payment Service: http://localhost:8083/metrics
- Notification Service: http://localhost:8084/metrics

Grafana dashboard tersedia di:

```
http://localhost:3000
```

## Kontribusi

Kontribusi selalu diterima! Silakan buat pull request atau buka issue untuk diskusi.

## Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE).

- **Backend**: Golang dengan Gin framework
- **Database**: PostgreSQL
- **Message Queue**: RabbitMQ
- **Containerization**: Docker
- **Testing**: Go testing framework
- **Logging & Monitoring**: Prometheus, Grafana

## Setup dan Instalasi

Instruksi setup akan ditambahkan setelah implementasi.

## API Documentation

Dokumentasi API akan ditambahkan setelah implementasi.

## Pengujian

Untuk menjalankan unit tests:

```bash
# Menjalankan semua test dengan Docker
make test

# Atau menjalankan test untuk service tertentu
cd <service-directory>
go test ./... -v
```

## Test Coverage

Informasi test coverage akan ditambahkan setelah implementasi.