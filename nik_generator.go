package gonik

import (
	"errors"
	"math/rand"
	"time"
	"unsafe"
)

var (
	ErrDistrictCodeLength     = errors.New("district code must be 6 digits")
	ErrInvalidGender          = errors.New("gender must be 'male' or 'female'")
	ErrUniqueCodeLength       = errors.New("unique code must be 4 digits when provided")
	ErrBufferTooSmall         = errors.New("buffer must have a minimum length of 16 bytes")
	ErrDatabaseNotInitialized = errors.New("database cache is empty, call InitDatabase() first")
)

// GenerateNIK menghasilkan string NIK 16 digit dengan performa 0 alokasi memori.
func GenerateNIK(dst []byte, kecamatanID string, birthDate time.Time, gender string, uniqueCode string) (string, error) {
	if len(kecamatanID) != 6 {
		return "", ErrDistrictCodeLength
	}
	if gender != "pria" && gender != "wanita" {
		return "", ErrInvalidGender
	}
	if uniqueCode != "" && len(uniqueCode) != 4 {
		return "", ErrUniqueCodeLength
	}
	if len(dst) < 16 {
		return "", ErrBufferTooSmall
	}

	// 3. Salin Kode Kecamatan (Indeks 0-5)
	copy(dst[0:6], kecamatanID)

	// 4. Hitung Komponen Tanggal Lahir
	day := birthDate.Day()
	if gender == "wanita" {
		day += 40
	}
	month := int(birthDate.Month())
	year := birthDate.Year() % 100 // Ambil 2 digit terakhir

	// 5. Tulis Tanggal, Bulan, Tahun ke Buffer tanpa alokasi string/sprintf
	dst[6] = byte('0' + day/10)
	dst[7] = byte('0' + day%10)

	dst[8] = byte('0' + month/10)
	dst[9] = byte('0' + month%10)

	dst[10] = byte('0' + year/10)
	dst[11] = byte('0' + year%10)

	// 6. Isi Kode Unik (Indeks 12-15)
	if uniqueCode != "" {
		copy(dst[12:16], uniqueCode)
	} else {
		// Jika kosong, buat 4 digit acak secara efisien
		// Menggunakan rand.Intn dari global pseudo-random generator (0 allocs)
		num := rand.Intn(10000)
		dst[12] = byte('0' + num/1000)
		dst[13] = byte('0' + (num/100)%10)
		dst[14] = byte('0' + (num/10)%10)
		dst[15] = byte('0' + num%10)
	}

	// 7. Konversi byte array ke string secara aman dan 0 alokasi
	res := unsafe.String(&dst[0], len(dst))

	return res, nil
}

// GenerateRandomNIK menghasilkan NIK acak yang dijamin valid (terdaftar di dbCache) dengan murni 0 alokasi memori.
func GenerateRandomNIK(dst []byte) (string, error) {
	if len(dst) < 16 {
		return "", ErrBufferTooSmall
	}

	if len(dbCache) == 0 {
		return "", ErrDatabaseNotInitialized
	}

	if len(districtIDs) == 0 {
		return "", ErrDatabaseNotInitialized
	}
	randIdx := rand.Intn(len(districtIDs))
	copy(dst[0:6], districtIDs[randIdx])

	isWanita := rand.Intn(2) == 1

	// Acak Tanggal Lahir yang Valid
	year := 1950 + rand.Intn(2026-1950+1)
	month := 1 + rand.Intn(12)

	// Tentukan batas hari maksimal berdasarkan bulan & tahun kabisat
	maxDays := 31
	switch month {
	case 4, 6, 9, 11:
		maxDays = 30
	case 2:
		if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
			maxDays = 29
		} else {
			maxDays = 28
		}
	}
	day := 1 + rand.Intn(maxDays)

	// Terapkan offset +40 jika gender wanita
	if isWanita {
		day += 40
	}

	// Ambil 2 digit belakang untuk format NIK
	yearShort := year % 100

	// Tulis data waktu ke buffer
	dst[6] = byte('0' + day/10)
	dst[7] = byte('0' + day%10)

	dst[8] = byte('0' + month/10)
	dst[9] = byte('0' + month%10)

	dst[10] = byte('0' + yearShort/10)
	dst[11] = byte('0' + yearShort%10)

	// Buat 4 Digit Kode Unik Acak Berurutan (0001 - 9999)
	num := 1 + rand.Intn(9999)
	dst[12] = byte('0' + num/1000)
	dst[13] = byte('0' + (num/100)%10)
	dst[14] = byte('0' + (num/10)%10)
	dst[15] = byte('0' + num%10)

	// Ikat ke String secara aman dengan Unsafe
	res := unsafe.String(&dst[0], 16)

	return res, nil
}
