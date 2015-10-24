package kamaji

import (
	"code.google.com/p/go-uuid/uuid"
)

// Job is the structure that holds tasks.
type Job struct {
	ID     uuid.UUID
	Name   string
	Status Status
}

// NewJob create a new Job struct, generates a uuid for it and returns the job.
func NewJob(name string) *Job {
	j := new(Job)
	j.ID = uuid.NewRandom()
	j.Name = name
	j.Status = UNKNOWN
	return j
}
