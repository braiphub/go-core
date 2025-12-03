package gohorizon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Job represents a queueable job
type Job interface {
	// Handle executes the job logic
	Handle(ctx context.Context) error

	// Name returns the unique job type name
	Name() string
}

// JobWithTags allows jobs to define tags for filtering
type JobWithTags interface {
	Job
	Tags() []string
}

// JobWithRetry allows custom retry behavior
type JobWithRetry interface {
	Job
	MaxRetries() int
	RetryDelay() time.Duration
}

// JobWithTimeout allows custom timeout
type JobWithTimeout interface {
	Job
	Timeout() time.Duration
}

// JobWithQueue specifies a custom queue
type JobWithQueue interface {
	Job
	Queue() string
}

// Status represents job processing status
type Status string

const (
	StatusPending   Status = "pending"
	StatusReserved  Status = "reserved"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Payload represents a serialized job in Redis
type Payload struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Queue       string                 `json:"queue"`
	Data        json.RawMessage        `json:"data"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	AvailableAt time.Time              `json:"available_at"`
	ReservedAt  *time.Time             `json:"reserved_at,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	RetryDelay  time.Duration          `json:"retry_delay"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewPayload creates a new payload from a job
func NewPayload(job Job, queue string) (*Payload, error) {
	data, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	payload := &Payload{
		ID:          uuid.New().String(),
		Name:        job.Name(),
		Queue:       queue,
		Data:        data,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   now,
		AvailableAt: now,
		Timeout:     60 * time.Second,
		RetryDelay:  5 * time.Second,
		Metadata:    make(map[string]interface{}),
	}

	// Apply job-specific settings
	if jt, ok := job.(JobWithTags); ok {
		payload.Tags = jt.Tags()
	}

	if jr, ok := job.(JobWithRetry); ok {
		payload.MaxAttempts = jr.MaxRetries()
		payload.RetryDelay = jr.RetryDelay()
	}

	if jto, ok := job.(JobWithTimeout); ok {
		payload.Timeout = jto.Timeout()
	}

	return payload, nil
}

// Serialize converts payload to JSON bytes
func (p *Payload) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// DeserializePayload creates a payload from JSON bytes
func DeserializePayload(data []byte) (*Payload, error) {
	var payload Payload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// FailedJob represents a job that failed processing
type FailedJob struct {
	ID        string    `json:"id"`
	Queue     string    `json:"queue"`
	Payload   *Payload  `json:"payload"`
	Exception string    `json:"exception"`
	FailedAt  time.Time `json:"failed_at"`
}

// RecentJob represents a recently processed job
type RecentJob struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Queue       string        `json:"queue"`
	Status      Status        `json:"status"`
	Attempts    int           `json:"attempts"`
	Runtime     time.Duration `json:"runtime"`
	CompletedAt time.Time     `json:"completed_at"`
	Tags        []string      `json:"tags,omitempty"`
}

// JobRegistry holds registered job types for deserialization
type JobRegistry struct {
	jobs map[string]func() Job
}

// NewJobRegistry creates a new job registry
func NewJobRegistry() *JobRegistry {
	return &JobRegistry{
		jobs: make(map[string]func() Job),
	}
}

// Register adds a job type to the registry
func (r *JobRegistry) Register(factory func() Job) {
	job := factory()
	r.jobs[job.Name()] = factory
}

// Get returns a new instance of the job type
func (r *JobRegistry) Get(name string) (Job, error) {
	factory, ok := r.jobs[name]
	if !ok {
		return nil, ErrJobNotRegistered
	}
	return factory(), nil
}

// Has checks if a job type is registered
func (r *JobRegistry) Has(name string) bool {
	_, ok := r.jobs[name]
	return ok
}

// Names returns all registered job names
func (r *JobRegistry) Names() []string {
	names := make([]string, 0, len(r.jobs))
	for name := range r.jobs {
		names = append(names, name)
	}
	return names
}

// Hydrate creates a job instance from a payload
func (r *JobRegistry) Hydrate(payload *Payload) (Job, error) {
	job, err := r.Get(payload.Name)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payload.Data, job); err != nil {
		return nil, err
	}

	return job, nil
}
