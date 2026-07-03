package gonik

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Type       string
	PostalCode string
	Name       string
}

type Details struct {
	BirthDate       time.Time `json:"birth_date"`
	NIK             string    `json:"nik"`
	ProvinceID      string    `json:"province_id"`
	Province        string    `json:"province"`
	RegencyCityID 	string    `json:"regency_city_id"`
	RegencyCity   	string    `json:"regency_city"`
	DistrictID      string    `json:"District_id"`
	District        string    `json:"District"`
	PostalCode      string    `json:"postal_code"`
	Gender          string    `json:"gender"`
	UniqueCode      string    `json:"unique_code"`
	IsValid         bool      `json:"is_valid"`
}

//go:embed data/wilayah.bin
var defaultBinaryData []byte

var dbCache map[string]Result
var districtIDs []string

func InitDatabase() error {
	if len(defaultBinaryData) == 0 {
		return errors.New("embedded region binary data is empty")
	}

	dbCache = make(map[string]Result, 7600)

	scanner := bufio.NewScanner(bytes.NewReader(defaultBinaryData))

	for scanner.Scan() {
		row := scanner.Bytes()
		if len(row) == 0 {
			continue
		}

		parts := bytes.SplitN(row, []byte("|"), 4)
		if len(parts) < 4 {
			continue
		}

		code := strings.TrimRight(string(parts[0]), "-")
		
		dbCache[code] = Result{
			Type:       string(parts[1]),
			PostalCode: string(parts[2]),
			Name:       string(parts[3]),
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	districtIDs = make([]string, 0, len(dbCache))
	for k := range dbCache {
		if len(k) == 6 {
			districtIDs = append(districtIDs, k)
		}
	}

	return nil
}

type Parser struct {
	nik string
}

func New(nik string) Parser {
	var sb strings.Builder
	sb.Grow(16)
	for i := 0; i < len(nik); i++ {
		if nik[i] >= '0' && nik[i] <= '9' {
			sb.WriteByte(nik[i])
		}
	}
	return Parser{nik: sb.String()}
}

func (p Parser) IsValid() bool {
	return len(p.nik) == 16 &&
		p.Province() != "" &&
		p.RegencyCity() != "" &&
		p.District() != "" &&
		!p.BirthDate().IsZero()
}

func (p Parser) ProvinceID() string    { return p.getSubstring(0, 2) }
func (p Parser) RegencyCityID() string { return p.getSubstring(0, 4) }
func (p Parser) DistrictID() string    { return p.getSubstring(0, 6) }
func (p Parser) UniqueCode() string    { return p.getSubstring(12, 16) }

func (p Parser) search(code string) (Result, bool) {
	if dbCache == nil {
		return Result{}, false
	}
	res, exists := dbCache[code]
	return res, exists
}

func (p Parser) Province() string {
	if len(p.nik) < 2 {
		return ""
	}
	if res, ok := p.search(p.nik[0:2]); ok {
		return res.Name
	}
	return ""
}

func (p Parser) RegencyCity() string {
	if len(p.nik) < 4 {
		return ""
	}
	if res, ok := p.search(p.nik[0:4]); ok {
		return res.Name
	}
	return ""
}

func (p Parser) District() string {
	if len(p.nik) < 6 {
		return ""
	}
	if res, ok := p.search(p.nik[0:6]); ok {
		return res.Name
	}
	return ""
}

func (p Parser) PostalCode() string {
	if len(p.nik) < 6 {
		return ""
	}
	if res, ok := p.search(p.nik[0:6]); ok {
		return res.PostalCode
	}
	return ""
}

func (p Parser) Gender() string {
	if len(p.nik) != 16 {
		return ""
	}
	day, _ := strconv.Atoi(p.getSubstring(6, 8))
	if day > 40 {
		return "female"
	}
	return "male"
}

func (p Parser) BirthDate() time.Time {
	if len(p.nik) != 16 {
		return time.Time{}
	}

	// Use getSubstring or slice the string directly by hand
	day := int(p.nik[6]-'0')*10 + int(p.nik[7]-'0')
	month := int(p.nik[8]-'0')*10 + int(p.nik[9]-'0')
	year := int(p.nik[10]-'0')*10 + int(p.nik[11]-'0')

	if day > 40 {
		day -= 40
	}

	// Apply basic mathematical validation before entering the heavier path
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return time.Time{}
	}

	// Use a static two-digit current year (2026)
	const currentYear2Digit = 26
	fullYear := 1900 + year
	if year <= currentYear2Digit {
		fullYear = 2000 + year
	}

	// Use time.UTC to avoid local timezone overhead
	t := time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Day() != day || t.Month() != time.Month(month) || t.Year() != fullYear {
		return time.Time{}
	}

	return t
}

func (p Parser) GetDetails() Details {
	if len(p.nik) != 16 {
		return Details{}
	}

	// 1. Compute the birth date quickly using the mathematical method above
	birthDate := p.BirthDate()
	if birthDate.IsZero() {
		return Details{}
	}

	// 2. Extract regional code slices directly using slicing (0 allocations)
	provID := p.nik[0:2]
	kabID := p.nik[0:4]
	kecID := p.nik[0:6]

	// 3. Perform map lookups sequentially and store results in local variables
	provRes, provOK := dbCache[provID]
	kabRes, kabOK := dbCache[kabID]
	kecRes, kecOK := dbCache[kecID]

	// If any regional level is missing, treat the NIK as invalid
	if !provOK || !kabOK || !kecOK {
		return Details{}
	}

	// 4. Build the Details struct directly from cached stack data
	return Details{
		NIK:             p.nik,
		IsValid:         true,
		ProvinceID:      provID,
		Province:        provRes.Name,
		RegencyCityID: 	 kabID,
		RegencyCity:   	 kabRes.Name,
		DistrictID:      kecID,
		District:        kecRes.Name,
		PostalCode:      kecRes.PostalCode,
		Gender:          p.Gender(),
		BirthDate:       birthDate,
		UniqueCode:      p.nik[12:16],
	}
}

func (p Parser) getSubstring(start, end int) string {
	if len(p.nik) < end {
		return ""
	}
	return p.nik[start:end]
}
