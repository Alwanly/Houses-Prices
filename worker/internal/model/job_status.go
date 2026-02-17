package model

import "time"

// JobStatus represents the status of a scheduled job
type JobStatus struct {
	JobID     string    `json:"job_id"`
	SiteName  string    `json:"site_name"`
	Status    string    `json:"status"` // running, completed, failed
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Duration  float64   `json:"duration_seconds"`
	Scraped   int       `json:"scraped_count"`
	Saved     int       `json:"saved_count"`
	Errors    int       `json:"error_count"`
	Message   string    `json:"message,omitempty"`
}
