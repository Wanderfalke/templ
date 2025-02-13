package htmldiff

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/a-h/htmlformat"
	"github.com/a-h/templ"
	"github.com/google/go-cmp/cmp"
)

func Diff(input templ.Component, expected string) (diff string, err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	var errs []error

	// Format the expected value.
	go func() {
		defer wg.Done()
		e := new(strings.Builder)
		err := htmlformat.Fragment(e, strings.NewReader(expected))
		if err != nil {
			errs = append(errs, fmt.Errorf("expected html formatting error: %w", err))
		}
		expected = e.String()
	}()

	// Pipe via the HTML formatter.
	actual := new(strings.Builder)
	r, w := io.Pipe()
	go func() {
		defer wg.Done()
		err := htmlformat.Fragment(actual, r)
		if err != nil {
			errs = append(errs, fmt.Errorf("actual html formatting error: %w", err))
		}
	}()

	// Render the component.
	err = input.Render(context.Background(), w)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to render component: %w", err))
	}
	w.Close()

	// Wait for processing.
	wg.Wait()

	return cmp.Diff(expected, actual.String()), errors.Join(errs...)
}
