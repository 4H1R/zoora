package tutorials

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// aparatHashRe guards the value we splice into the outbound oEmbed URL.
var aparatHashRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// aparatHTTPClient is a short-timeout client for the author-time oEmbed proxy.
var aparatHTTPClient = &http.Client{Timeout: 10 * time.Second}

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context) ([]domain.Tutorial, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.repo.AdminList(ctx)
}

func (s *service) AdminGet(ctx context.Context, id uuid.UUID) (*domain.Tutorial, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *service) AdminCreate(ctx context.Context, dto domain.CreateTutorialDTO) (*domain.Tutorial, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	// New drafts append to the end of the curated order.
	max, err := s.repo.MaxPosition(ctx)
	if err != nil {
		return nil, err
	}
	tu := &domain.Tutorial{
		TitleEn:       dto.TitleEn,
		TitleFa:       dto.TitleFa,
		DescriptionEn: dto.DescriptionEn,
		DescriptionFa: dto.DescriptionFa,
		AparatHash:    dto.AparatHash,
		ThumbnailURL:  dto.ThumbnailURL,
		Position:      max + 1,
		// PublishedAt stays nil → draft.
	}
	if err := s.repo.Create(ctx, tu); err != nil {
		return nil, err
	}
	return tu, nil
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.UpdateTutorialDTO) (*domain.Tutorial, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	tu, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.TitleEn != nil {
		tu.TitleEn = *dto.TitleEn
	}
	if dto.TitleFa != nil {
		tu.TitleFa = *dto.TitleFa
	}
	if dto.DescriptionEn != nil {
		tu.DescriptionEn = *dto.DescriptionEn
	}
	if dto.DescriptionFa != nil {
		tu.DescriptionFa = *dto.DescriptionFa
	}
	if dto.AparatHash != nil {
		tu.AparatHash = *dto.AparatHash
	}
	if dto.ThumbnailURL != nil {
		tu.ThumbnailURL = *dto.ThumbnailURL
	}
	if err := s.repo.Update(ctx, tu); err != nil {
		return nil, err
	}
	return tu, nil
}

func (s *service) AdminPublish(ctx context.Context, id uuid.UUID) (*domain.Tutorial, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	tu, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tu.PublishedAt == nil {
		now := time.Now()
		tu.PublishedAt = &now
		if err := s.repo.Update(ctx, tu); err != nil {
			return nil, err
		}
	}
	return tu, nil
}

func (s *service) AdminUnpublish(ctx context.Context, id uuid.UUID) (*domain.Tutorial, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	tu, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tu.PublishedAt = nil
	if err := s.repo.Update(ctx, tu); err != nil {
		return nil, err
	}
	return tu, nil
}

func (s *service) AdminDelete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	// The video lives on Aparat — only the metadata row is dropped here.
	return s.repo.Delete(ctx, id)
}

func (s *service) AdminReorder(ctx context.Context, dto domain.ReorderTutorialsDTO) error {
	if _, err := s.requireAdmin(ctx); err != nil {
		return err
	}
	ids := make([]uuid.UUID, 0, len(dto.IDs))
	for _, raw := range dto.IDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return domain.NewValidationError(map[string]string{"ids": "must be valid UUIDs"})
		}
		ids = append(ids, id)
	}
	return s.repo.Reorder(ctx, ids)
}

func (s *service) AdminAparatOEmbed(ctx context.Context, hash string) (*domain.AparatOEmbedResponse, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if !aparatHashRe.MatchString(hash) {
		return nil, domain.NewValidationError(map[string]string{"hash": "must be an Aparat video hash"})
	}
	watch := "https://www.aparat.com/v/" + hash
	endpoint := "https://www.aparat.com/oembed?url=" + url.QueryEscape(watch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("tutorials.AdminAparatOEmbed: build request: %w", err)
	}
	resp, err := aparatHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tutorials.AdminAparatOEmbed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		// Most often a 404 for an unknown hash — surface as not found.
		return nil, domain.ErrNotFound
	}
	var body struct {
		Title        string `json:"title"`
		ThumbnailURL string `json:"thumbnail_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("tutorials.AdminAparatOEmbed: decode: %w", err)
	}
	return &domain.AparatOEmbedResponse{Title: body.Title, ThumbnailURL: body.ThumbnailURL}, nil
}
