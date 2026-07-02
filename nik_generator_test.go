package main

import (
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

func BenchmarkGenerateNIK(b *testing.B) {
	birthDate := time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC)
	var buf [16]byte
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateNIK(buf[:], "357820", birthDate, "pria", "0001")
	}
}