package gonik

import (
	"sync"
	"testing"
	"time"
)

func TestGenerateNIK_Pria(t *testing.T) {
	birthDate := time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC)
	var buf [16]byte
	nik, err := GenerateNIK(buf[:], "357820", birthDate, "pria", "0001")
	if err != nil {
		t.Fatalf("Gagal generate NIK: %v", err)
	}

	expected := "3578201503990001"
	if nik != expected {
		t.Errorf("Ekspektasi NIK %s, got %s", expected, nik)
	}
}

func TestGenerateNIK_Wanita(t *testing.T) {
	birthDate := time.Date(2001, 8, 12, 0, 0, 0, 0, time.UTC)
	var buf [16]byte
	nik, err := GenerateNIK(buf[:], "357820", birthDate, "wanita", "0002")
	if err != nil {
		t.Fatalf("Gagal generate NIK: %v", err)
	}

	// Tanggal 12 + 40 = 52
	expected := "3578205208010002"
	if nik != expected {
		t.Errorf("Ekspektasi NIK %s, got %s", expected, nik)
	}
}

func TestGenerateRandomNIK(t *testing.T) {
	var buf [16]byte

	nikAcak, err := GenerateRandomNIK(buf[:])
	if err != nil {
		t.Errorf("Gagal membuat NIK acak: %v", err)
	}

	parser := New(nikAcak)
	if !parser.IsValid() {
		t.Errorf("Ekspektasi NIK acak valid, namun hasil menyatakan tidak valid: %s", nikAcak)
	}
}

func BenchmarkGenerateNIK_BatchLoop(b *testing.B) {
	birthDate := time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC)
	var buf [16]byte

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateNIK(buf[:], "357820", birthDate, "pria", "0001")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateRandomNIK_BatchLoop(b *testing.B) {
	_ = InitDatabase()
	b.ResetTimer()

	var buf [16]byte

	for i := 0; i < b.N; i++ {
		_, err := GenerateRandomNIK(buf[:])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateNIK_SyncPool(b *testing.B) {
	var bufferPool = sync.Pool{
		New: func() any {
			buf := make([]byte, 16)
			return &buf
		},
	}

	birthDate := time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bufPtr := bufferPool.Get().(*[]byte)
		buf := *bufPtr
	
		_, err := GenerateNIK(buf, "357820", birthDate, "pria", "0001")
		if err != nil {
			b.Fatal(err)
		}
	
		bufferPool.Put(bufPtr)
	}
}

func BenchmarkGenerateRandomNIK_SyncPool(b *testing.B) {
	_ = InitDatabase()

	var bufferPool = sync.Pool{
		New: func() any {
			buf := make([]byte, 16)
			return &buf
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bufPtr := bufferPool.Get().(*[]byte)
		buf := *bufPtr
	
		_, err := GenerateRandomNIK(buf)
		if err != nil {
			b.Fatal(err)
		}
	
		bufferPool.Put(bufPtr)
	}
}
