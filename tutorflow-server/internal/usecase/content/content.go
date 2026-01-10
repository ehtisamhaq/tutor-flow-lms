package content

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type contentUseCase struct {
	lessonRepo repository.LessonRepository
}

// NewContentUseCase creates a new content use case
func NewUseCase(lessonRepo repository.LessonRepository) domain.ContentUseCase {
	return &contentUseCase{
		lessonRepo: lessonRepo,
	}
}

func (uc *contentUseCase) AddAttachment(ctx context.Context, lessonID uuid.UUID, attachment domain.Attachment) error {
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return err
	}

	var attachments []domain.Attachment
	if lesson.Attachments != nil && *lesson.Attachments != "" && *lesson.Attachments != "[]" {
		if err := json.Unmarshal([]byte(*lesson.Attachments), &attachments); err != nil {
			return fmt.Errorf("failed to parse attachments: %w", err)
		}
	}

	attachments = append(attachments, attachment)

	bytes, err := json.Marshal(attachments)
	if err != nil {
		return err
	}

	attachmentsStr := string(bytes)
	lesson.Attachments = &attachmentsStr

	return uc.lessonRepo.Update(ctx, lesson)
}

func (uc *contentUseCase) RemoveAttachment(ctx context.Context, lessonID uuid.UUID, fileURL string) error {
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return err
	}

	if lesson.Attachments == nil || *lesson.Attachments == "" || *lesson.Attachments == "[]" {
		return nil
	}

	var attachments []domain.Attachment
	if err := json.Unmarshal([]byte(*lesson.Attachments), &attachments); err != nil {
		return err
	}

	newAttachments := make([]domain.Attachment, 0)
	for _, a := range attachments {
		if a.FileURL != fileURL {
			newAttachments = append(newAttachments, a)
		}
	}

	bytes, err := json.Marshal(newAttachments)
	if err != nil {
		return err
	}

	attachmentsStr := string(bytes)
	lesson.Attachments = &attachmentsStr

	return uc.lessonRepo.Update(ctx, lesson)
}

func (uc *contentUseCase) GetAttachments(ctx context.Context, lessonID uuid.UUID) ([]domain.Attachment, error) {
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	if lesson.Attachments == nil || *lesson.Attachments == "" || *lesson.Attachments == "[]" {
		return []domain.Attachment{}, nil
	}

	var attachments []domain.Attachment
	if err := json.Unmarshal([]byte(*lesson.Attachments), &attachments); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (uc *contentUseCase) UpdateResourceInfo(ctx context.Context, lessonID uuid.UUID, info domain.ResourceInfo) error {
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return err
	}

	if lesson.LessonType != domain.LessonTypeResource {
		return fmt.Errorf("lesson is not of type resource")
	}

	// For resource lessons, we can store the info in the Content field as JSON
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	contentStr := string(bytes)
	lesson.Content = &contentStr

	return uc.lessonRepo.Update(ctx, lesson)
}
