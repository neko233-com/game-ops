package ingest

import (
	"archive/zip"
	"encoding/xml"
	"path/filepath"
	"sort"
	"strings"

	"github.com/neko233/game-ops/num"
)

type xlsxSST struct {
	SI []struct {
		T string `xml:"t"`
	} `xml:"si"`
}

type xlsxCell struct {
	R  string `xml:"r,attr"`
	T  string `xml:"t,attr"`
	V  string `xml:"v"`
	IS struct {
		T string `xml:"t"`
	} `xml:"is"`
}

type xlsxSheet struct {
	SheetData struct {
		Rows []struct {
			Cells []xlsxCell `xml:"c"`
		} `xml:"row"`
	} `xml:"sheetData"`
}

func readSharedStrings(f *zip.File) ([]string, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var sst xlsxSST
	if err := xml.NewDecoder(r).Decode(&sst); err != nil {
		return nil, err
	}
	out := make([]string, len(sst.SI))
	for i, si := range sst.SI {
		out[i] = si.T
	}
	return out, nil
}

func cellString(c xlsxCell, shared []string) string {
	switch c.T {
	case "s":
		idx := 0
		for _, ch := range c.V {
			if ch < '0' || ch > '9' {
				break
			}
			idx = idx*10 + int(ch-'0')
		}
		if idx < len(shared) {
			return shared[idx]
		}
		return ""
	case "inlineStr":
		return c.IS.T
	default:
		return c.V
	}
}

func parseXLSX(path string) ([]Record, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var shared []string
	var sheets []*zip.File
	for _, f := range zr.File {
		switch {
		case f.Name == "xl/sharedStrings.xml":
			shared, err = readSharedStrings(f)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml"):
			sheets = append(sheets, f)
		}
	}
	sort.Slice(sheets, func(i, j int) bool { return sheets[i].Name < sheets[j].Name })

	var records []Record
	for _, sf := range sheets {
		r, err := sf.Open()
		if err != nil {
			return nil, err
		}
		var sheet xlsxSheet
		err = xml.NewDecoder(r).Decode(&sheet)
		r.Close()
		if err != nil {
			return nil, err
		}
		if len(sheet.SheetData.Rows) == 0 {
			continue
		}
		sheetName := strings.TrimSuffix(filepath.Base(sf.Name), ".xml")

		// First row is the header: map column letters to names.
		header := map[string]string{}
		for _, c := range sheet.SheetData.Rows[0].Cells {
			col := parseColumnLetter(c.R)
			if col != "" {
				header[col] = cellString(c, shared)
			}
		}

		for i, row := range sheet.SheetData.Rows[1:] {
			rec := Record{
				File: path, Sheet: sheetName, Row: int64(i + 2),
				Values: map[string]num.Num{}, Raw: map[string]string{},
			}
			for _, c := range row.Cells {
				name, ok := header[parseColumnLetter(c.R)]
				if !ok || name == "" {
					continue
				}
				value := cellString(c, shared)
				rec.Raw[name] = value
				if n, err := num.FromString(value); err == nil {
					rec.Values[name] = n
				}
			}
			records = append(records, rec)
		}
	}
	return records, nil
}
