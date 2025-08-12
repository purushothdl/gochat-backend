package upload

// ProfileUploadJob defines the data structure for an image processing job.
type ProfileUploadJob struct {
	JobID        string `json:"job_id"`
	UserID       string `json:"user_id"`
	StagingKey   string `json:"staging_key"`
	OriginalName string `json:"original_name"`
}

// JobResponse is the immediate response after initiating an upload.
type JobResponse struct {
	JobID string `json:"job_id"`
}