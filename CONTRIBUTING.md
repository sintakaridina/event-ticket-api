# Panduan Kontribusi

Terima kasih telah mempertimbangkan untuk berkontribusi pada proyek Trae! Berikut adalah panduan untuk membantu Anda berkontribusi.

## Proses Kontribusi

1. Fork repositori ini
2. Clone fork Anda ke mesin lokal Anda
3. Buat branch fitur baru (`git checkout -b feature/amazing-feature`)
4. Commit perubahan Anda (`git commit -m 'Menambahkan fitur amazing'`)
5. Push ke branch (`git push origin feature/amazing-feature`)
6. Buat Pull Request baru

## Standar Kode

### Go

- Ikuti [Effective Go](https://golang.org/doc/effective_go.html) dan [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Gunakan `gofmt` untuk memformat kode Anda
- Pastikan kode Anda lulus `golint` dan `go vet`
- Tulis unit test untuk kode baru
- Dokumentasikan fungsi dan metode publik

### Pesan Commit

- Gunakan bentuk imperatif ("Tambahkan fitur" bukan "Menambahkan fitur")
- Buat pesan commit yang singkat dan deskriptif
- Referensikan issue atau PR jika relevan

## Pengembangan Lokal

### Prasyarat

- Go 1.21 atau lebih baru
- Docker dan Docker Compose
- Make (opsional, untuk menggunakan Makefile)

### Menyiapkan Lingkungan Pengembangan

1. Clone repositori
2. Jalankan `make build` untuk membangun semua layanan
3. Jalankan `make up` untuk memulai semua layanan

## Pengujian

- Tulis unit test untuk semua kode baru
- Pastikan semua test lulus sebelum mengirimkan PR
- Gunakan `go test ./...` untuk menjalankan semua test

## Dokumentasi

- Perbarui README.md jika diperlukan
- Dokumentasikan semua endpoint API baru
- Perbarui diagram arsitektur jika Anda membuat perubahan signifikan

## Pelaporan Bug

Jika Anda menemukan bug, buat issue baru dengan informasi berikut:

- Deskripsi singkat tentang bug
- Langkah-langkah untuk mereproduksi
- Perilaku yang diharapkan
- Perilaku aktual
- Tangkapan layar (jika relevan)
- Lingkungan (OS, versi Go, dll.)

## Permintaan Fitur

Jika Anda memiliki ide untuk fitur baru, buat issue baru dengan label "enhancement" dan jelaskan fitur yang Anda usulkan secara detail.

## Pertanyaan

Jika Anda memiliki pertanyaan tentang proyek ini, silakan buat issue baru dengan label "question".

Terima kasih atas kontribusi Anda!