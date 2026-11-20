// Package ingest streams numeric game config files (xlsx/json/jsonl/csv/tsv)
// with bounded memory. It uses only the Go standard library; xlsx is read as a
// zip of XML parts, so no cgo and no external dependencies are required.
package ingest

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/neko233-com/game-ops/num"
)

const maxFileCeiling = 1_000_000

// Aggregate is an exact per-column summary.
type Aggregate struct {
	Count    int64
	Min, Max num.Num
	Sum      num.Num
}

// Stats is the result of a scan.
type Stats struct {
	Files   int64
	Rows    int64
	ByExt   map[string]int64
	Columns map[string]Aggregate
	Errors  int64
}

func newStats() Stats {
	return Stats{ByExt: map[string]int64{}, Columns: map[string]Aggregate{}}
}

func (s *Stats) add(records []Record, ext string, err error) {
	if err != nil {
		s.Errors++
		return
	}
	s.Files++
	s.ByExt[ext]++
	s.Rows += int64(len(records))
	for _, rec := range records {
		for key, value := range rec.Values {
			agg, ok := s.Columns[key]
			if !ok {
				agg = Aggregate{Min: value, Max: value, Sum: value, Count: 1}
			} else {
				agg.Count++
				agg.Sum = agg.Sum.Add(value)
				agg.Min = agg.Min.Min(value)
				agg.Max = agg.Max.Max(value)
			}
			s.Columns[key] = agg
		}
	}
}

// Record is one numeric row/object from a config file.
type Record struct {
	File, Sheet string
	Row         int64
	Values      map[string]num.Num
	Raw         map[string]string
}

// Options controls a scan. MaxFiles <= 0 means the 1,000,000 ceiling.
type Options struct {
	Roots      []string
	MaxFiles   int64
	Workers    int
	Extensions []string
	OnRecord   func(Record) error
}

// Scan walks the roots with a bounded worker pool and aggregates numeric cells.
func Scan(ctx context.Context, opt Options) (Stats, error) {
	stats := newStats()
	if err := ctx.Err(); err != nil {
		return stats, err
	}
	exts := opt.Extensions
	if len(exts) == 0 {
		exts = []string{".json", ".jsonl", ".csv", ".tsv", ".xlsx"}
	}
	allowed := map[string]bool{}
	for _, e := range exts {
		allowed[strings.ToLower(e)] = true
	}
	workers := opt.Workers
	if workers < 1 {
		workers = 4
	}
	limit := opt.MaxFiles
	if limit <= 0 || limit > maxFileCeiling {
		limit = maxFileCeiling
	}

	paths := make(chan string, workers*2)
	type result struct {
		records []Record
		ext     string
		err     error
		onErr   error
	}
	results := make(chan result, workers*2)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				if err := ctx.Err(); err != nil {
					return
				}
				records, err := parseFile(path)
				res := result{records: records, ext: strings.ToLower(filepath.Ext(path)), err: err}
				if err == nil && opt.OnRecord != nil {
					for _, rec := range records {
						if e := opt.OnRecord(rec); e != nil {
							res.onErr = e
							break
						}
					}
				}
				results <- res
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var count int64
	var walkErr error
	walkDone := make(chan struct{})
	go func() {
		defer close(walkDone)
		for _, root := range opt.Roots {
			err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if err := ctx.Err(); err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if !allowed[strings.ToLower(filepath.Ext(path))] {
					return nil
				}
				if count >= limit {
					return io.EOF // stop walking once the cap is hit
				}
				count++
				paths <- path
				return nil
			})
			if err != nil && err != io.EOF {
				walkErr = err
				break
			}
			if count >= limit {
				break
			}
		}
		close(paths)
	}()

	var firstErr error
	for res := range results {
		stats.add(res.records, res.ext, res.err)
		if res.onErr != nil && firstErr == nil {
			firstErr = res.onErr
		}
	}
	<-walkDone
	if firstErr != nil {
		return stats, firstErr
	}
	if walkErr != nil {
		return stats, walkErr
	}
	if err := ctx.Err(); err != nil {
		return stats, err
	}
	return stats, nil
}

func parseFile(path string) ([]Record, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv", ".tsv":
		return parseDelimited(path)
	case ".json", ".jsonl":
		return parseJSON(path)
	case ".xlsx":
		return parseXLSX(path)
	default:
		return nil, fmt.Errorf("ingest: unsupported extension: %s", path)
	}
}

func parseDelimited(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	if strings.EqualFold(filepath.Ext(path), ".tsv") {
		r.Comma = '\t'
	}
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	headers := rows[0]
	var records []Record
	for i, row := range rows[1:] {
		rec := Record{File: path, Row: int64(i + 2), Values: map[string]num.Num{}, Raw: map[string]string{}}
		for j, h := range headers {
			if j >= len(row) || h == "" {
				continue
			}
			rec.Raw[h] = row[j]
			if v, err := num.FromString(row[j]); err == nil {
				rec.Values[h] = v
			}
		}
		records = append(records, rec)
	}
	return records, nil
}

func parseJSON(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.UseNumber()

	var first interface{}
	if err := dec.Decode(&first); err != nil {
		return nil, err
	}

	var objects []map[string]interface{}
	switch v := first.(type) {
	case []interface{}:
		for _, item := range v {
			obj, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("ingest: array element is not an object: %s", path)
			}
			objects = append(objects, obj)
		}
	case map[string]interface{}:
		objects = append(objects, v)
		// NDJSON: more objects follow the first one.
		for {
			var obj map[string]interface{}
			if err := dec.Decode(&obj); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			objects = append(objects, obj)
		}
	default:
		return nil, fmt.Errorf("ingest: top-level JSON is neither array nor object: %s", path)
	}

	var records []Record
	for i, obj := range objects {
		rec := Record{File: path, Row: int64(i + 1), Values: map[string]num.Num{}, Raw: map[string]string{}}
		flatten(obj, "", &rec)
		records = append(records, rec)
	}
	return records, nil
}

func flatten(obj map[string]interface{}, prefix string, rec *Record) {
	for key, value := range obj {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			flatten(v, path, rec)
		case json.Number:
			s := v.String()
			rec.Raw[path] = s
			if n, err := num.FromString(s); err == nil {
				rec.Values[path] = n
			}
		case string:
			rec.Raw[path] = v
			if n, err := num.FromString(v); err == nil {
				rec.Values[path] = n
			}
		default:
			rec.Raw[path] = fmt.Sprint(value)
		}
	}
}

// parseColumnLetter turns a cell reference like "B12" into "B".
func parseColumnLetter(ref string) string {
	i := 0
	for i < len(ref) && ref[i] >= 'A' && ref[i] <= 'Z' {
		i++
	}
	return ref[:i]
}
