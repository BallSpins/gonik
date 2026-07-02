# gonik

[![Go Reference](https://pkg.go.dev/badge/github.com/ballspins/gonik.svg)](https://pkg.go.dev/github.com/ballspins/gonik)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**gonik** adalah pustaka (*library*) parser NIK (Nomor Induk Kependudukan) KTP Indonesia berkinerja tinggi (*high-performance*) yang ditulis menggunakan bahasa Go. Dirancang khusus untuk skenario industri yang membutuhkan kecepatan pemrosesan super kilat dengan efisiensi memori ekstrem murni **Zero-Allocation (0 B/op, 0 allocs/op)**.

Pustaka ini mampu mengurai dan memvalidasi lebih dari **5-6 juta data NIK per detik** pada perangkat keras kelas standar berkat optimasi arsitektur memori di level *compiler stack* dan peniadaan *pointer chasing*.

---

## Mengapa Pendekatan Ini?

Berbeda dengan pustaka parser NIK konvensional atau versi porting dari bahasa dinamis (seperti PHP/Node.js) yang sering kali memicu alokasi memori berulang di *heap*, `gonik` memaksimalkan kapabilitas runtime Go melalui pendekatan:

* **Memory-Resident Preheated Map:** Dataset wilayah se-Indonesia dimuat sekali di awal (*startup*) ke dalam RAM (`dbCache`). Ini menjamin kompleksitas pencarian konstan $O(1)$ yang sangat cepat dibandingkan metode *disk-backed binary search*.
* **Murni Zero-Allocation ($0\text{ B/op}$):** Konstruktor dan metode parser menggunakan *Value Type* (bukan *pointer*). Seluruh siklus hidup objek dikunci di dalam **Stack Memory**, menghilangkan ketergantungan pada *Garbage Collector* (GC) dan mencegah degradasi performa akibat *cache miss*.
* **Matematika Kering:** Mengonversi string tanggal lahir dan penentuan jenis kelamin secara langsung lewat kalkulasi numerik karakter indeks byte (`nik[i] - '0'`), sepenuhnya menyingkirkan fungsi mahal seperti `strconv.Atoi` dan penanganan zona waktu lokal (`time.Local`).
* **Single-Pass Map Lookup:** Memangkas frekuensi operasi *hashing map* wilayah dari yang awalnya 7 kali redundan menjadi maksimal **3 kali lookup** sekuensial ter-cache di stack untuk menyusun informasi data KTP secara utuh.

---

## API Reference

### 1. Parser Instance (`Parser`)

Metode instansiasi *read-only* untuk mengekstrak informasi terstruktur dari string NIK.

| Metode | Jenis Return | Deskripsi |
| :--- | :--- | :--- |
| `New(nik string) Parser` | `Parser` | Konstruktor *Value Type* untuk menginisialisasi objek parser di Stack. |
| `Province() string` | `string` | Mendapatkan nama provinsi berdasarkan 2 digit pertama NIK. |
| `RegencyCity() string` | `string` | Mendapatkan nama kabupaten/kota berdasarkan 4 digit pertama NIK. |
| `District() string` | `string` | Mendapatkan nama kecamatan berdasarkan 6 digit pertama NIK. |
| `PostalCode() string` | `string` | Mendapatkan kode pos yang melekat pada level kecamatan. |
| `Gender() string` | `string` | Mendeteksi gender (`Make` / `Female`) dengan penanganan otomatis *offset* 40. |
| `BirthDate() time.Time` | `time.Time` | Mengembalikan objek tanggal lahir tervalidasi. |
| `GetDetails() Details` | `Details` | Mengonversi seluruh informasi NIK ke dalam satu struct tunggal `Details`. |

### 2. Generator (`Generate`)

| Metode | Deskripsi |
| :--- | :--- |
| `GenerateNIK(...) string` | Membuat 16 digit NIK tiruan tervalidasi untuk kebutuhan *seeding* data otomatis. |

---

## Cara Penggunaan

### Inisialisasi & Parsing Satuan

```go
package main

import (
	"fmt"
	"github.com/ballspins/gonik"
)

func main() {
	// 1. Muat dataset wilayah ke memori sekali saja di awal aplikasi
	if err := gonik.InitDatabase(); err != nil {
		panic(err)
	}

	// 2. Instansiasi parser (Murni alokasi Stack)
	parser := gonik.New("3578201503990001")

	// 3. Ambal data parsial atau sekaligus
	fmt.Println("Kecamatan:", parser.District())
	fmt.Println("Tanggal Lahir:", parser.BirthDate().Format("2006-01-02"))

	// 4. Ambil seluruh detail terstruktur (Ukuran struct Details: 192 bytes)
	details := parser.GetDetails()
	if details.IsValid {
		fmt.Printf("%+v\n", details)
	}
}

```

---

## Perbandingan Benchmark Riil (1 Juta Iterasi)

Pengujian dilakukan secara objektif dengan membandingkan eksekusi ketat subsistem pengujian Go (`go test -bench`) antara pustaka **([fanchann/nik-parser](https://github.com/fanchann/nik-parser))** melawan **gonik ([ballspins/gonik](https://github.com/BallSpins/gonik))** pada arsitektur mesin yang sama.

| Metrik Kinerja | [fanchann/nik-parser](https://github.com/fanchann/nik-parser) | [ballspins/gonik](https://github.com/BallSpins/gonik) | Keunggulan `gonik` |
| :--- | :---: | :---: | :---: |
| **Kecepatan rata-rata (`ns/op`)** | ~643.4 ns/op | **~193.2 ns/op** | **~3.3x Lebih Cepat** |
| **Alokasi Memori (`B/op`)** | 210 B/op | **0 B/op** | **Mutlak (Zero Allocation)** |
| **Jumlah Alokasi Heap (`allocs/op`)** | 4 allocs/op | **0 allocs/op** | **Murni Bebas Sampah Heap** |
| **Throughput Data (per detik)** | ~1.55 Juta NIK/detik | **~5.17 Juta NIK/detik** | **Memproses ~3.6 Juta Lebih Banyak** |

### Mengapa `gonik` Bisa Menang Telak?

1. **Peniadaan Alokasi Heap (0 B/op):** Pustaka [fanchann/nik-parser](https://github.com/fanchann/nik-parser) menghasilkan sampah memori sebesar 210 byte dan memicu 4 kali operasi alokasi heap (`4 allocs/op`) pada setiap satu kali proses eksekusi NIK. `gonik` mengunci seluruh siklus hidup objek di dalam **Stack Memory** sehingga CPU tidak perlu membuang siklus untuk berinteraksi dengan runtime allocator.
2. **Bebas Degradasi Garbage Collector (GC):** Akibat dari alokasi kumulatif [fanchann/nik-parser](https://github.com/fanchann/nik-parser), memproses 1 juta data secara berurutan akan memaksa sistem meminjam memori total hingga **~200 MB** sebelum disapu oleh GC. Di sisi lain, `gonik` mempertahankan penggunaan memori kumulatif yang stabil dan bersih sejak iterasi pertama hingga terakhir.
3. **Single-Pass Map Lookup:** Jika parser lain melakukan pencarian map berulang kali (redundant lookup) untuk mengambil data Provinsi, Kabupaten, dan Kecamatan secara terpisah, `gonik` hanya melakukan maksimal 3 kali operasi *hashing map* sekuensial ter-cache untuk menyusun objek data KTP secara instan.

### Hasil `go test -bench` Internal

Berikut adalah hasil pengujian performa bawaan subsistem pengujian Go pada pustaka `gonik`:

```bash
λ go test -bench=. -benchmem
Ukuran total struct Details: 192 bytes
goos: windows
goarch: amd64
pkg: github.com/ballspins/gonik
cpu: AMD Ryzen 3 7320U with Radeon Graphics
BenchmarkGenerateNIK-8                  21824054                59.65 ns/op            0 B/op          0 allocs/op
BenchmarkNikParser_GetDetails-8          6397884               195.0 ns/op             0 B/op          0 allocs/op
BenchmarkParser_Province-8              51141302                22.80 ns/op            0 B/op          0 allocs/op
BenchmarkParser_RegencyCity-8           55107804                22.75 ns/op            0 B/op          0 allocs/op
BenchmarkParser_District-8              52825503                22.48 ns/op            0 B/op          0 allocs/op
BenchmarkParser_PostalCode-8            45795932                22.73 ns/op            0 B/op          0 allocs/op
BenchmarkParser_Gender-8                136785369                8.779 ns/op           0 B/op          0 allocs/op
BenchmarkParser_BirthDate-8             17517122                67.33 ns/op            0 B/op          0 allocs/op
BenchmarkParser_getSubstring-8          1000000000               0.8793 ns/op          0 B/op          0 allocs/op
PASS
ok      github.com/ballspins/gonik      11.954s

```

> **Analisis Angka:** Operasi komparasi tercepat dicatat oleh `getSubString` sebesar **~0.87 ns/op** yang menandakan fungsi berhasil di-*inline* penuh oleh compiler ke tingkat register CPU. Kecepatan *single-pass detail extraction* (`GetDetails`) kokoh berada pada level **~193 ns/op** murni tanpa alokasi heap tunggal pun (`0 B/op`).

---

## Lingkungan Pengujian (Benchmark Environment)

Untuk menjaga akurasi konteks data di atas, berikut adalah spesifikasi mesin eksekusi lokal yang digunakan selama proses standarisasi metrik:

| Komponen | Spesifikasi Perangkat |
| --- | --- |
| **Sistem Operasi** | Windows 11 Home Single Language (Build 26100) |
| **Prosesor** | AMD Ryzen 3 7320U (4 Cores, 8 Threads, Base 2.4GHz) |
| **Memori Utama** | 8GB LPDDR5 Dual-Channel @ 5500 MT/s |
| **Arsitektur Compiler** | Go 1.25.0 `amd64` (CLI Environment) |

---

## Catatan Penting untuk Produksi

1. **Efek Startup Awal:** Saat memanggil `gonik.InitDatabase()`, sistem akan memakan waktu beberapa milidetik untuk membangun peta *hash map* internal di Heap RAM (`Peak Heap Alloc` awal berkisar $\approx 2.14\text{ MB}$). Pemuatan ini disarankan dieksekusi di fungsi `init()` atau blok awal `main()` sebelum server HTTP mendengarkan *request*.
2. **Hindari Variabel Pointer:** Untuk mempertahankan performa *Zero-Allocation* di sistem Anda sendiri, pastikan tidak mengubah variabel instansiasi `gonik.New()` menjadi tipe pointer (`*Parser`) secara manual atau melemparkannya ke fungsi luar yang memicu *escape analysis* ke heap.

---

## Lisensi

Proyek ini dilisensikan di bawah ketentuan [MIT License](https://www.google.com/search?q=LICENSE).

---

*Dataset wilayah dan referensi arsitektur data diadaptasi secara radikal dari basis struktur data [mul14/nik_parser](https://github.com/mul14/nik_parser.ts).*
