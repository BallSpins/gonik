# gonik

[![Go Reference](https://pkg.go.dev/badge/github.com/ballspins/gonik.svg)](https://pkg.go.dev/github.com/ballspins/gonik)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**gonik** adalah pustaka (*library*) parser NIK (Nomor Induk Kependudukan) KTP Indonesia berkinerja tinggi (*high-performance*) yang ditulis menggunakan bahasa Go. Dirancang khusus untuk skenario industri yang membutuhkan kecepatan pemrosesan super kilat dengan efisiensi memori ekstrem murni **Zero-Allocation (0 B/op, 0 allocs/op)**.

Library ini mampu mengurai dan memvalidasi lebih dari **5-6 juta data NIK per detik** pada perangkat keras kelas standar berkat optimasi arsitektur memori di level *compiler stack* dan peniadaan *pointer chasing*.

---

## Mengapa Pendekatan Ini?

Berbeda dengan library parser NIK konvensional atau versi porting dari bahasa dinamis (seperti PHP/Node.js) yang sering kali memicu alokasi memori berulang di *heap*, `gonik` memaksimalkan kapabilitas runtime Go melalui pendekatan:

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
| `Gender() string` | `string` | Mendeteksi gender (`Male` / `Female`) dengan penanganan otomatis *offset* 40. |
| `BirthDate() time.Time` | `time.Time` | Mengembalikan objek tanggal lahir tervalidasi. |
| `GetDetails() Details` | `Details` | Mengonversi seluruh informasi NIK ke dalam satu struct tunggal `Details`. |

### 2. Generator (`Generate`)

| Metode | Deskripsi |
| :--- | :--- |
| `GenerateNIK(dst []byte, kecamatanID string, birthDate time.Time, gender string, uniqueCode string) (string, error)` | Membuat 16 digit NIK tiruan tervalidasi dengan menulis hasil ke buffer yang sudah disediakan. |
| `GenerateRandomNIK(dst []byte) (string, error)` | Membuat 16 digit NIK acak tervalidasi dengan menulis hasil ke buffer yang sudah disediakan. | 

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

	// 3. Ambil data parsial atau sekaligus
	fmt.Println("District:", parser.District())
	fmt.Println("Birth Date:", parser.BirthDate().Format("2006-01-02"))

	// 4. Ambil seluruh detail terstruktur (Ukuran struct Details: 192 bytes)
	details := parser.GetDetails()
	if details.IsValid {
		fmt.Printf("%+v\n", details)
	}
}
```

### Generate NIK

Berikut contoh pemanfaatan fungsi generator untuk membuat NIK tiruan secara efisien:

```go
package main

import (
	"fmt"
	"time"

	"github.com/ballspins/gonik"
)

func main() {
	var buf [16]byte
	birthDate := time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC)

	nik, err := gonik.GenerateNIK(buf[:], "357820", birthDate, "pria", "0001")
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated NIK:", nik)
}
```

Parameter:
- `dst`: buffer byte dengan panjang minimal 16 untuk menampung hasil NIK
- `kecamatanID`: kode kecamatan 6 digit
- `birthDate`: tanggal lahir yang akan dikonversi
- `gender`: `pria` atau `wanita`
- `uniqueCode`: kode unik 4 digit opsional


### Generate Random NIK

Berikut contoh pemanfaatan fungsi generator untuk membuat NIK secara acak dengan efisien:

```go
package main

import (
	"fmt"

	"github.com/ballspins/gonik"
)

func main() {
	var buf [16]byte

	err := gonik.InitDatabase()
	if err != nil {
		panic(err)
	}

	nik, err := gonik.GenerateRandomNIK(buf[:])
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated NIK:", nik)
}
```

Parameter:
- `dst`: buffer byte dengan panjang minimal 16 untuk menampung hasil NIK

## Panduan Manajemen Memori & Optimasi Performa

Library gonik didesain dengan tanda tangan fungsi (function signature) menerima parameter buffer dari luar (dst []byte). Pendekatan ini sengaja diambil untuk memberikan kendali penuh kepada pengembang dalam mengatur siklus hidup memori dan menghindari tekanan berlebih pada Garbage Collector (GC).

Berikut adalah dua pola implementasi profesional untuk memanfaatkan efisiensi murni Zero-Allocation di berbagai skenario:

### Skenario 1: Batch Processing & Seeding Data (Loop Ketat)

Jika Anda perlu menghasilkan jutaan data NIK acak secara berurutan (misalnya untuk kebutuhan database seeding atau pengujian beban), alokasikan buffer array sekali saja di luar perulangan.

Eksekusi berikutnya akan langsung menimpa (overwrite) memori pada indeks yang sama secara instan tanpa memicu alokasi heap baru.

```go
package main

import (
	"github.com/ballspins/gonik"
)

func main() {
	_ = gonik.InitDatabase()

	// 1. Alokasikan buffer 16 byte SEKALI SAJA di Stack (Luar Loop)
	var buf [16]byte 

	for i := 0; i < 1000000; i++ {
		// 2. Data lama di dalam buf otomatis tertimpa bersih (0 allocs/op)
		nik, _ := gonik.GenerateRandomNIK(buf[:])
		
		// Proses data nik Anda di sini...
		_ = nik 
	}
}
```

Kenapa ini efisien? RAM aplikasi Anda akan tetap stabil dan tidak akan naik sama sekali dari iterasi pertama hingga ke sejuta, karena tidak ada objek baru yang diciptakan di dalam lingkaran eksekusi.

### Skenario 2: Implementasi pada Server HTTP Concurrent (sync.Pool)
Jika fungsi generator ditaruh di dalam skenario konkuren tinggi seperti server HTTP, performa terbaik dicapai dengan mengadopsi mekanisme daur ulang (recycle) memori memanfaatkan sync.Pool.

Mekanisme ini mencegah array memicu escape analysis ke memori Heap akibat pembuatan variabel lokal yang berulang di setiap goroutine masuk.

```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"github.com/ballspins/gonik"
)

// Sediakan pool khusus buffer 16 byte
var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 16)
		return &b // Menyimpan pointer ke slice untuk efisiensi pool
	},
}

func handleGenerateNIK(w http.ResponseWriter, r *http.Request) {
	// 1. Pinjam buffer dari pool pusat
	bufPtr := bufferPool.Get().(*[]byte)
	buf := *bufPtr

	// 2. Tulis data NIK langsung ke buffer pinjaman
	nik, err := gonik.GenerateRandomNIK(buf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "NIK:", nik)

	// 3. Kembalikan buffer ke pool untuk digunakan request berikutnya
	bufferPool.Put(bufPtr)
}
```

Mengapa Buffer Tidak Perlu Dibersihkan (Zeroing Out)?
Mungkin Anda bertanya-tanya mengapa kita tidak membersihkan isi buffer (mengisi ulang elemennya dengan angka 0) sebelum dikembalikan ke sync.Pool atau di dalam perulangan.

Pembersihan memori (zeroing out) hanya wajib dilakukan pada dua kondisi:

Keamanan Data Transaksional: Menangani data sensitif (seperti password plaintext, token JWT, atau kunci enkripsi) guna mencegah kebocoran data di lapisan memori lain.

Panjang Data Dinamis: Jika fungsi berikutnya menulis data dengan panjang bervariasi (misal menulis 5 byte di atas sisa memori lama sepanjang 11 byte, yang akan menghasilkan residu data rusak).

Karena struktur data NIK pasti tepat berukuran 16 byte, algoritma internal gonik dijamin selalu menimpa indeks koordinat 0 sampai 15 secara utuh dan sempurna. Tidak ada residu data lama yang akan bocor atau merusak hasil generation berikutnya.

---

## Perbandingan Benchmark Riil (1 Juta Iterasi)

Pengujian dilakukan secara objektif dengan membandingkan eksekusi ketat subsistem pengujian Go (`go test -bench`) antara library **([fanchann/nik-parser](https://github.com/fanchann/nik-parser))** melawan **gonik ([ballspins/gonik](https://github.com/BallSpins/gonik))** pada arsitektur mesin yang sama.

| Metrik Kinerja | [fanchann/nik-parser](https://github.com/fanchann/nik-parser) | [ballspins/gonik](https://github.com/BallSpins/gonik) | Keunggulan `gonik` |
| :--- | :---: | :---: | :---: |
| **Kecepatan rata-rata (`ns/op`)** | ~643.4 ns/op | **~222.4 ns/op** | **~2.9x Lebih Cepat** |
| **Alokasi Memori (`B/op`)** | 210 B/op | **0 B/op** | **Mutlak (Zero Allocation)** |
| **Jumlah Alokasi Heap (`allocs/op`)** | 4 allocs/op | **0 allocs/op** | **Murni Bebas Sampah Heap** |
| **Throughput Data (per detik)** | ~1.55 Juta NIK/detik | **~4.49 Juta NIK/detik** | **Memproses ~2.9 Juta Lebih Banyak** |

### Mengapa `gonik` Bisa Menang Telak?

1. **Peniadaan Alokasi Heap (0 B/op):** Library [fanchann/nik-parser](https://github.com/fanchann/nik-parser) menghasilkan sampah memori sebesar 210 byte dan memicu 4 kali operasi alokasi heap (`4 allocs/op`) pada setiap satu kali proses eksekusi NIK. `gonik` mengunci seluruh siklus hidup objek di dalam **Stack Memory** sehingga CPU tidak perlu membuang siklus untuk berinteraksi dengan runtime allocator.
2. **Bebas Degradasi Garbage Collector (GC):** Akibat dari alokasi kumulatif [fanchann/nik-parser](https://github.com/fanchann/nik-parser), memproses 1 juta data secara berurutan akan memaksa sistem meminjam memori total hingga **~200 MB** sebelum disapu oleh GC. Di sisi lain, `gonik` mempertahankan penggunaan memori kumulatif yang stabil dan bersih sejak iterasi pertama hingga terakhir.
3. **Single-Pass Map Lookup:** Jika parser lain melakukan pencarian map berulang kali (redundant lookup) untuk mengambil data Provinsi, Kabupaten, dan Kecamatan secara terpisah, `gonik` hanya melakukan maksimal 3 kali operasi *hashing map* sekuensial ter-cache untuk menyusun objek data KTP secara instan.

### Hasil `go test -bench` Internal

Berikut adalah hasil pengujian performa bawaan subsistem pengujian Go pada library `gonik`:

```bash
λ go test -bench=. -benchmem
Ukuran total struct Details: 192 bytes
goos: windows
goarch: amd64
pkg: github.com/ballspins/gonik
cpu: AMD Ryzen 3 7320U with Radeon Graphics
BenchmarkGenerateNIK_BatchLoop-8                20584990                54.25 ns/op            0 B/op          0 allocs/op
BenchmarkGenerateRandomNIK_BatchLoop-8           6471228               184.5 ns/op             0 B/op          0 allocs/op
BenchmarkGenerateNIK_SyncPool-8                 16970053                70.99 ns/op            0 B/op          0 allocs/op
BenchmarkGenerateRandomNIK_SyncPool-8            5950116               199.6 ns/op             0 B/op          0 allocs/op
BenchmarkNikParser_GetDetails-8                  6629397               184.0 ns/op             0 B/op          0 allocs/op
BenchmarkParser_Province-8                      56872306                22.51 ns/op            0 B/op          0 allocs/op
BenchmarkParser_RegencyCity-8                   50162190                23.00 ns/op            0 B/op          0 allocs/op
BenchmarkParser_District-8                      54330355                22.79 ns/op            0 B/op          0 allocs/op
BenchmarkParser_PostalCode-8                    52199331                22.40 ns/op            0 B/op          0 allocs/op
BenchmarkParser_Gender-8                        137527414                8.740 ns/op           0 B/op          0 allocs/op
BenchmarkParser_BirthDate-8                     17812029                68.46 ns/op            0 B/op          0 allocs/op
BenchmarkParser_getSubstring-8                  1000000000               0.4357 ns/op          0 B/op          0 allocs/op
PASS
ok      github.com/ballspins/gonik      15.581s

```

```bash
λ go run ./cmd/sampler
Running Benchmark sampling 50x

================ AVG SAMPLING RESULT ================
BenchmarkGenerateNIK_BatchLoop           : 66.50 ns/op (50 samples)
BenchmarkGenerateRandomNIK_BatchLoop     : 237.52 ns/op (50 samples)
BenchmarkGenerateNIK_SyncPool            : 89.09 ns/op (50 samples)
BenchmarkGenerateRandomNIK_SyncPool      : 257.32 ns/op (50 samples)
BenchmarkNikParser_GetDetails            : 222.45 ns/op (50 samples)
BenchmarkParser_Province                 : 27.63 ns/op (50 samples)
BenchmarkParser_RegencyCity              : 27.77 ns/op (50 samples)
BenchmarkParser_District                 : 24.05 ns/op (50 samples)
BenchmarkParser_PostalCode               : 23.29 ns/op (50 samples)
BenchmarkParser_Gender                   : 9.09 ns/op (50 samples)
BenchmarkParser_BirthDate                : 68.48 ns/op (50 samples)
BenchmarkParser_getSubstring             : 0.46 ns/op (50 samples)
==========================================================
Total Sampling Time Execution: 13m35.818s
==========================================================
```

> **Analisis Angka:** Operasi komparasi tercepat dicatat oleh `getSubString` sebesar **~0.48 ns/op** yang menandakan fungsi berhasil di-*inline* penuh oleh compiler ke tingkat register CPU. Kecepatan *single-pass detail extraction* (`GetDetails`) kokoh berada pada level **~222 ns/op** murni tanpa alokasi heap tunggal pun (`0 B/op`).

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

## Kontribusi

Kontribusi dalam bentuk perbaikan *bug*, optimasi performa murni, peningkatan cakupan pengujian (*test coverage*), maupun pembaruan dataset wilayah sangat diapresiasi. Library ini menganut prinsip efisiensi ekstrem, jadi pastikan setiap perubahan kode tetap menjaga status **Zero-Allocation**.

Silahkan berkontribusi dengan langkah-langkah berikut:

1. **Fork** repositori ini ke akun GitHub Anda.
2. **Clone** hasil fork tersebut ke lingkungan lokal Anda:
   ```bash
   git clone https://github.com/USERNAME/gonik.git
   cd gonik
	```

3. Buat *branch* baru untuk fitur atau perbaikan Anda:
	```bash
	git checkout -b fitur-atau-perbaikan-anda
	```

4. Lakukan perubahan kode dan pastikan seluruh unit test serta benchmark lolos tanpa memicu alokasi memori baru:
	```bash
	go test -v ./...
	go test -bench=. -benchmem
	```


5. Commit perubahan Anda, push ke GitHub, dan buka sebuah **Pull Request** ke *branch* utama (`main`) repositori ini.


## Lisensi

Proyek ini dilisensikan di bawah ketentuan [MIT License](./LICENSE).

---

*Dataset wilayah dan referensi arsitektur data diadaptasi secara radikal dari basis struktur data [mul14/nik_parser](https://github.com/mul14/nik_parser.ts).*
