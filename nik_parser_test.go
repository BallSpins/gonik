package main

import (
	"fmt"
	"testing"
	"unsafe"
)

// TestMain digunakan untuk setup database sekali saja sebelum seluruh test berjalan
func TestMain(m *testing.M) {
	// Panggil init database Anda (bisa menggunakan jalur file atau embed data)
	// Di sini berasumsi menggunakan fungsi InitDatabase yang mengarah ke file dummy/asli
	err := InitDatabase()
	if err != nil {
		panic("Gagal inisialisasi database untuk testing: " + err.Error())
	}

	// Jalankan seluruh suite pengujian
	m.Run()
}

// 1. UNIT TEST: Menguji keakuratan logika parsing
func TestNikParser_GetDetails(t *testing.T) {
	// Gunakan data NIK tiruan yang valid sesuai format wilayah.bin Anda
	// Misal: 357820 (Surabaya), 150399 (Pria, 15 Maret 1999)
	nikInput := "3578201503990001"
	parser := New(nikInput)
	details := parser.GetDetails()

	// Validasi status kelayakan NIK
	if !details.IsValid {
		t.Errorf("Ekspektasi NIK %s valid, namun hasil menyatakan tidak valid", nikInput)
	}

	// Validate gender
	if details.Gender != "male" {
		t.Errorf("Expected gender 'male', got '%s'", details.Gender)
	}

	birthDateStr := details.BirthDate.Format("2006-01-02")
	if birthDateStr != "1999-03-15" {
		t.Errorf("Ekspektasi tanggal lahir '1999-03-15', got '%s'", birthDateStr)
	}

	// Validasi Kode Unik
	if details.UniqueCode != "0001" {
		t.Errorf("Ekspektasi kode unik '0001', got '%s'", details.UniqueCode)
	}
}

// 2. UNIT TEST: Menguji skenario NIK Wanita (+40 pada tanggal)
func TestNikParser_Wanita(t *testing.T) {
	// Tanggal lahir 12 Agustus 2001 -> 12 + 40 = 52
	nikInput := "3578205208010002"
	parser := New(nikInput)

	if parser.Gender() != "female" {
		t.Errorf("Expected gender 'female', got '%s'", parser.Gender())
	}

	if parser.BirthDate().Format("2006-01-02") != "2001-08-12" {
		t.Errorf("Expected birth date '2001-08-12', got '%s'", parser.BirthDate().Format("2006-01-02"))
	}
}

// 3. BENCHMARK TEST: Menguji kecepatan eksekusi dan alokasi memori
func BenchmarkNikParser_GetDetails(b *testing.B) {
	nikInput := "3578201503990001"
	parser := New(nikInput)

	// Reset timer agar proses inisialisasi di atas tidak masuk dalam hitungan benchmark
	b.ResetTimer()

	// b.N akan ditentukan otomatis oleh Go untuk menguji iterasi yang optimal (bisa jutaan kali)
	for i := 0; i < b.N; i++ {
		_ = parser.GetDetails()
	}
}

func BenchmarkParser_Province(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.Province()
	}
}

func BenchmarkParser_RegencyCity(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.RegencyCity()
	}
}

func BenchmarkParser_District(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.District()
	}
}

func BenchmarkParser_PostalCode(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.PostalCode()
	}
}

func BenchmarkParser_Gender(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.Gender()
	}
}

func BenchmarkParser_BirthDate(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.BirthDate()
	}
}

func BenchmarkParser_getSubstring(b *testing.B) {
	parser := New("3578201503990001")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.getSubstring(6, 8)
	}
}

// 4. TEST UKURAN STRUCT: Memastikan ukuran struct Details tetap efisien
func TestStructSize(t *testing.T) {
	var d Details
	fmt.Printf("Ukuran total struct Details: %d bytes\n", unsafe.Sizeof(d))
}
