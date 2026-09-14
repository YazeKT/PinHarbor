package webrunner

import (
	"context"
	"encoding/csv"
	"reflect"
	"sync"

	"github.com/gosom/scrapemate"
)

// resilientCSVWriter is a local replacement for Scrapemate's CSV writer. Some
// optional enrichment workers legitimately return a nil Result.Data when no
// public email is found. Ignoring empty results keeps the collection alive and
// preserves every real place row.
type resilientCSVWriter struct {
	w    *csv.Writer
	once sync.Once
}

func newResilientCSVWriter(w *csv.Writer) scrapemate.ResultWriter {
	return &resilientCSVWriter{w: w}
}

func (c *resilientCSVWriter) Run(ctx context.Context, in <-chan scrapemate.Result) error {
	for {
		select {
		case <-ctx.Done():
			return c.w.Error()
		case result, ok := <-in:
			if !ok {
				return c.w.Error()
			}

			for _, element := range csvCapableResults(result.Data) {
				c.once.Do(func() { _ = c.w.Write(element.CsvHeaders()) })
				if err := c.w.Write(element.CsvRow()); err != nil {
					return err
				}
			}

			c.w.Flush()
			if err := c.w.Error(); err != nil {
				return err
			}
		}
	}
}

func csvCapableResults(data any) []scrapemate.CsvCapable {
	if data == nil {
		return nil
	}

	if element, ok := data.(scrapemate.CsvCapable); ok {
		return []scrapemate.CsvCapable{element}
	}

	value := reflect.ValueOf(data)
	if value.Kind() != reflect.Slice {
		return nil
	}

	elements := make([]scrapemate.CsvCapable, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		if !item.IsValid() || ((item.Kind() == reflect.Interface || item.Kind() == reflect.Ptr) && item.IsNil()) {
			continue
		}
		if element, ok := item.Interface().(scrapemate.CsvCapable); ok {
			elements = append(elements, element)
		}
	}

	return elements
}
