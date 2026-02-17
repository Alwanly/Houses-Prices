package errors

import "fmt"

// ScrapingError represents an error during scraping
type ScrapingError struct {
	Site    string
	URL     string
	Attempt int
	Err     error
}

func (e *ScrapingError) Error() string {
	return fmt.Sprintf("scraping %s failed at %s (attempt %d): %v",
		e.Site, e.URL, e.Attempt, e.Err)
}

func (e *ScrapingError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error
type ValidationError struct {
	Field string
	Value interface{}
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %s (value: %v): %v",
		e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ExtractionError represents a data extraction error
type ExtractionError struct {
	Field    string
	Selector string
	Err      error
}

func (e *ExtractionError) Error() string {
	return fmt.Sprintf("extraction failed for field %s (selector: %s): %v",
		e.Field, e.Selector, e.Err)
}

func (e *ExtractionError) Unwrap() error {
	return e.Err
}
