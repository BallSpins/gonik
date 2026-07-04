package gonik

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestMain(m *testing.M) {
	err := InitDatabase()
	if err != nil {
		panic("Gagal inisialisasi database untuk testing: " + err.Error())
	}

	m.Run()
}

func TestNikParser_InvalidCharLength_TooShort(t *testing.T) {
	invalidNIK := "123456789012345" // NIK less than 16 chars
	parser := New(invalidNIK)

	if parser.IsValid() {
		t.Errorf("Expected isValid 'false', got '%t'", parser.IsValid())
	}
}

func TestNikParser_InvalidCharLength_TooLong(t *testing.T) {
	invalidNIK := "12345678901234567" // NIK more than 16 chars
	parser := New(invalidNIK)

	if parser.IsValid() {
		t.Errorf("Expected isValid 'false', got '%t'", parser.IsValid())
	}
}

func TestNikParser_InvalidCharLength_NonNumeric(t *testing.T) {
	invalidNIK := "12345678901234AB"
	parser := New(invalidNIK)

	if parser.IsValid() {
		t.Errorf("Expected isValid 'false', got '%t'", parser.IsValid())
	}
}

func TestNikParser_GetDetails(t *testing.T) {
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

	// Validate birth date
	birthDateStr := details.BirthDate.Format("2006-01-02")
	if birthDateStr != "1999-03-15" {
		t.Errorf("Ekspektasi tanggal lahir '1999-03-15', got '%s'", birthDateStr)
	}

	// Validasi Kode Unik
	if details.UniqueCode != "0001" {
		t.Errorf("Ekspektasi kode unik '0001', got '%s'", details.UniqueCode)
	}
}

func TestNikParser_Wanita(t *testing.T) {
	nikInput := "3578205208010002"
	parser := New(nikInput)

	if parser.Gender() != "female" {
		t.Errorf("Expected gender 'female', got '%s'", parser.Gender())
	}

	if parser.BirthDate().Format("2006-01-02") != "2001-08-12" {
		t.Errorf("Expected birth date '2001-08-12', got '%s'", parser.BirthDate().Format("2006-01-02"))
	}
}

func BenchmarkGetDetails_NewInstance(b *testing.B) {
	nikInput := "3578201503990001"
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		parser := New(nikInput)
		_ = parser.GetDetails()
	}
}

func BenchmarkGetDetails_ReusedInstance(b *testing.B) {
	nikInput := "3578201503990001"
	
	parser := New(nikInput)
	b.ResetTimer()
	
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

func TestStructSize(t *testing.T) {
	var d Details
	fmt.Printf("Ukuran total struct Details: %d bytes\n", unsafe.Sizeof(d))
}
