package generation

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/williamprado/foto-magica-profissional/internal/credits"
	"github.com/williamprado/foto-magica-profissional/internal/notifications"
	"github.com/williamprado/foto-magica-profissional/internal/providers/ai"
	"github.com/williamprado/foto-magica-profissional/internal/providers/storage"
)

var (
	ErrInvalidReferenceImage = errors.New("invalid reference image")
	ErrInvalidUserImage      = errors.New("invalid user image")
)

type Job struct {
	ID             uuid.UUID          `json:"id"`
	Status         string             `json:"status"`
	Progress       int                `json:"progress"`
	CostCredits    int                `json:"costCredits"`
	ResultURL      string             `json:"resultUrl,omitempty"`
	PromptSections []ai.PromptSection `json:"promptSections"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

type Service struct {
	db            *pgxpool.Pool
	credits       credits.Service
	storage       storage.Provider
	aiProvider    ai.Provider
	notifications notifications.Service
}

func NewService(db *pgxpool.Pool, credits credits.Service, storage storage.Provider, aiProvider ai.Provider, notifications notifications.Service) Service {
	return Service{
		db:            db,
		credits:       credits,
		storage:       storage,
		aiProvider:    aiProvider,
		notifications: notifications,
	}
}

type CreateJobInput struct {
	Instructions       string `json:"instructions"`
	ReferenceImageB64  string `json:"referenceImageB64" binding:"required"`
	ReferenceMimeType  string `json:"referenceMimeType" binding:"required"`
	UserImageB64       string `json:"userImageB64" binding:"required"`
	UserMimeType       string `json:"userMimeType" binding:"required"`
}

func (s Service) ListJobs(ctx context.Context, tenantID uuid.UUID) ([]Job, error) {
	rows, err := s.db.Query(ctx, `
		select j.id, j.status, j.progress, j.cost_credits, coalesce(a.object_key, ''), p.sections, j.created_at, j.updated_at
		from generation_jobs j
		left join result_assets a on a.job_id = j.id
		left join prompts p on p.id = j.prompt_id
		where j.tenant_id = $1
		order by j.created_at desc
		limit 20
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var (
			job         Job
			objectKey   string
			sectionsRaw []byte
		)
		if err := rows.Scan(&job.ID, &job.Status, &job.Progress, &job.CostCredits, &objectKey, &sectionsRaw, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}
		if len(sectionsRaw) > 0 {
			_ = json.Unmarshal(sectionsRaw, &job.PromptSections)
		}
		if objectKey != "" {
			url, _ := s.storage.SignedGetURL(ctx, objectKey, 15*time.Minute)
			job.ResultURL = url
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s Service) CreateJob(ctx context.Context, tenantID, userID uuid.UUID, input CreateJobInput) (Job, error) {
	refBytes, err := decodeBase64(input.ReferenceImageB64)
	if err != nil {
		return Job{}, ErrInvalidReferenceImage
	}
	userBytes, err := decodeBase64(input.UserImageB64)
	if err != nil {
		return Job{}, ErrInvalidUserImage
	}

	if err := s.credits.DebitForGeneration(ctx, tenantID, userID, 1, "AI portrait generation"); err != nil {
		return Job{}, err
	}

	referenceKey := filepath.Join(tenantID.String(), "references", uuid.NewString()+extensionFor(input.ReferenceMimeType))
	userKey := filepath.Join(tenantID.String(), "source", uuid.NewString()+extensionFor(input.UserMimeType))

	refUpload, err := s.storage.Upload(ctx, referenceKey, input.ReferenceMimeType, refBytes)
	if err != nil {
		return Job{}, err
	}
	userUpload, err := s.storage.Upload(ctx, userKey, input.UserMimeType, userBytes)
	if err != nil {
		return Job{}, err
	}

	analysis, err := s.aiProvider.AnalyzeReference(ctx, input.ReferenceMimeType, refBytes)
	if err != nil {
		return Job{}, err
	}
	promptSections, err := s.aiProvider.GeneratePrompt(ctx, analysis, input.Instructions)
	if err != nil {
		return Job{}, err
	}
	sectionsRaw, _ := json.Marshal(promptSections)
	analysisRaw, _ := json.Marshal(analysis)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Job{}, err
	}
	defer tx.Rollback(ctx)

	var referenceID uuid.UUID
	if err := tx.QueryRow(ctx, `
		insert into reference_images (tenant_id, created_by, object_key, mime_type, size_bytes, analysis, status)
		values ($1, $2, $3, $4, $5, $6, 'analyzed')
		returning id
	`, tenantID, userID, refUpload.Key, input.ReferenceMimeType, len(refBytes), analysisRaw).Scan(&referenceID); err != nil {
		return Job{}, err
	}

	var promptID uuid.UUID
	if err := tx.QueryRow(ctx, `
		insert into prompts (tenant_id, reference_image_id, created_by, sections, raw_text)
		values ($1, $2, $3, $4, $5)
		returning id
	`, tenantID, referenceID, userID, sectionsRaw, analysis.Summary).Scan(&promptID); err != nil {
		return Job{}, err
	}

	job := Job{}
	if err := tx.QueryRow(ctx, `
		insert into generation_jobs (tenant_id, created_by, reference_image_id, prompt_id, source_image_key, status, progress, attempts, cost_credits)
		values ($1, $2, $3, $4, $5, 'queued', 10, 0, 1)
		returning id, status, progress, cost_credits, created_at, updated_at
	`, tenantID, userID, referenceID, promptID, userUpload.Key).Scan(&job.ID, &job.Status, &job.Progress, &job.CostCredits, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return Job{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Job{}, err
	}

	job.PromptSections = promptSections
	s.notifications.Publish(ctx, "generation.job.queued", map[string]any{
		"tenantId": tenantID.String(),
		"jobId":    job.ID.String(),
	})
	return job, nil
}

func (s Service) ProcessNextQueuedJob(ctx context.Context, maxAttempts int) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var (
		jobID         uuid.UUID
		tenantID      uuid.UUID
		referenceKey  string
		sourceKey     string
		referenceMime string
		userMime      string
		sectionsRaw   []byte
		attempts      int
	)
	err = tx.QueryRow(ctx, `
		select j.id, j.tenant_id, r.object_key, j.source_image_key, r.mime_type, split_part(j.source_image_key, '.', 2), p.sections, j.attempts
		from generation_jobs j
		join reference_images r on r.id = j.reference_image_id
		join prompts p on p.id = j.prompt_id
		where j.status in ('queued', 'retrying')
		order by j.created_at asc
		for update skip locked
		limit 1
	`).Scan(&jobID, &tenantID, &referenceKey, &sourceKey, &referenceMime, &userMime, &sectionsRaw, &attempts)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}

	if _, err := tx.Exec(ctx, `update generation_jobs set status = 'processing', progress = 45, attempts = attempts + 1, updated_at = now() where id = $1`, jobID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	var promptSections []ai.PromptSection
	_ = json.Unmarshal(sectionsRaw, &promptSections)

	refBytes, refErr := s.storage.Read(ctx, referenceKey)
	userBytes, userErr := s.storage.Read(ctx, sourceKey)
	if refErr != nil || userErr != nil {
		return s.failJob(ctx, jobID, attempts+1, maxAttempts, "storage read failed")
	}

	result, err := s.aiProvider.GenerateImage(ctx, promptSections, userMime, userBytes, referenceMime, refBytes)
	if err != nil {
		return s.failJob(ctx, jobID, attempts+1, maxAttempts, err.Error())
	}

	resultKey := filepath.Join(tenantID.String(), "results", jobID.String()+extensionFor(result.ContentType))
	upload, err := s.storage.Upload(ctx, resultKey, result.ContentType, result.Bytes)
	if err != nil {
		return s.failJob(ctx, jobID, attempts+1, maxAttempts, err.Error())
	}

	if _, err := s.db.Exec(ctx, `
		insert into result_assets (tenant_id, job_id, object_key, mime_type, favorite)
		values ($1, $2, $3, $4, false)
	`, tenantID, jobID, upload.Key, result.ContentType); err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `
		update generation_jobs set status = 'done', progress = 100, updated_at = now() where id = $1
	`, jobID); err != nil {
		return err
	}

	s.notifications.Publish(ctx, "generation.job.completed", map[string]any{
		"tenantId": tenantID.String(),
		"jobId":    jobID.String(),
	})
	return nil
}

func (s Service) failJob(ctx context.Context, jobID uuid.UUID, attempts, maxAttempts int, reason string) error {
	status := "failed"
	progress := 0
	if attempts < maxAttempts {
		status = "retrying"
		progress = 15
	}
	_, err := s.db.Exec(ctx, `
		update generation_jobs
		set status = $2, progress = $3, failure_reason = $4, updated_at = now()
		where id = $1
	`, jobID, status, progress, reason)
	return err
}

func extensionFor(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
