package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/panhao/url-shortener/internal/model"
	"github.com/panhao/url-shortener/internal/repository"
	"github.com/panhao/url-shortener/internal/util"
)

type LinkService struct {
	linkRepo  *repository.LinkRepo
	codeSvc   *ShortCodeService
	baseURL   string
}

func NewLinkService(linkRepo *repository.LinkRepo, codeSvc *ShortCodeService, baseURL string) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
		codeSvc:  codeSvc,
		baseURL:  baseURL,
	}
}

func (s *LinkService) Create(ctx context.Context, req model.CreateLinkReq, userID *int64) (*model.LinkResp, error) {
	// Validate URL
	targetURL := strings.TrimSpace(req.URL)
	if targetURL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	parsed, err := url.Parse(targetURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	// Block redirect loops
	if strings.Contains(parsed.Host, strings.TrimPrefix(s.baseURL, "http://")) ||
		strings.Contains(parsed.Host, strings.TrimPrefix(s.baseURL, "https://")) {
		return nil, fmt.Errorf("cannot shorten URLs from this service")
	}

	// Validate or generate short code
	customCode := strings.TrimSpace(req.CustomCode)
	if customCode != "" {
		if !util.IsValidCustomCode(customCode) {
			return nil, fmt.Errorf("invalid custom code: must be 1-20 chars, alphanumeric, hyphens or underscores")
		}
		exists, err := s.linkRepo.IsShortCodeExists(ctx, customCode)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("custom code '%s' already taken", customCode)
		}
	}

	domain := req.Domain
	if domain == "" {
		domain = strings.TrimPrefix(s.baseURL, "http://")
		domain = strings.TrimPrefix(domain, "https://")
	}

	redirectType := req.RedirectType
	if redirectType == "" {
		redirectType = "302"
	}

	rulesJSON := "[]"
	if len(req.Rules) > 0 {
		data, _ := json.Marshal(req.Rules)
		rulesJSON = string(data)
	}

	// Generate short code: use custom code or generate a random one
	shortCode := customCode
	if shortCode == "" {
		var err error
		shortCode, err = s.codeSvc.Generate(ctx)
		if err != nil {
			return nil, fmt.Errorf("generate short code: %w", err)
		}
	}

	link := &model.Link{
		ShortCode:    shortCode,
		OriginalURL:  targetURL,
		Title:        req.Title,
		Description:  req.Description,
		UserID:       userID,
		Domain:       domain,
		RedirectType: redirectType,
		Rules:        rulesJSON,
		IsActive:     true,
	}

	if req.ExpireAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpireAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expire_at format, use RFC3339: %w", err)
		}
		link.ExpireAt = &t
	}

	if req.Password != "" {
		hash, err := util.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		link.PasswordHash = hash
	}

	if err := s.linkRepo.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}

	return &model.LinkResp{
		Link:     *link,
		ShortURL: s.baseURL + "/" + shortCode,
	}, nil
}

func (s *LinkService) Get(ctx context.Context, code string) (*model.Link, error) {
	return s.linkRepo.GetByShortCode(ctx, code)
}

func (s *LinkService) GetUserLink(ctx context.Context, code string, userID int64) (*model.Link, error) {
	return s.linkRepo.GetUserLink(ctx, code, userID)
}

func (s *LinkService) List(ctx context.Context, userID int64, page, pageSize int) (*model.PaginatedResp, error) {
	links, total, err := s.linkRepo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var resp []model.LinkResp
	for _, l := range links {
		resp = append(resp, model.LinkResp{
			Link:     l,
			ShortURL: s.baseURL + "/" + l.ShortCode,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.PaginatedResp{
		Data:       resp,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *LinkService) Update(ctx context.Context, code string, req model.UpdateLinkReq) (*model.Link, error) {
	updates := make(map[string]any)

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.RedirectType != nil {
		updates["redirect_type"] = *req.RedirectType
	}
	if req.ExpireAt != nil {
		if *req.ExpireAt != "" {
			t, err := time.Parse(time.RFC3339, *req.ExpireAt)
			if err != nil {
				return nil, fmt.Errorf("invalid expire_at: %w", err)
			}
			updates["expire_at"] = t
		} else {
			updates["expire_at"] = nil
		}
	}

	if len(updates) == 0 {
		return s.linkRepo.GetByShortCode(ctx, code)
	}

	if err := s.linkRepo.Update(ctx, code, updates); err != nil {
		return nil, err
	}

	return s.linkRepo.GetByShortCode(ctx, code)
}

func (s *LinkService) Delete(ctx context.Context, code string) error {
	return s.linkRepo.Delete(ctx, code)
}

func (s *LinkService) GetStats(ctx context.Context, code string, userID int64) (map[string]any, error) {
	link, err := s.linkRepo.GetUserLink(ctx, code, userID)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, fmt.Errorf("link not found")
	}

	stats := map[string]any{
		"link": model.LinkResp{
			Link:     *link,
			ShortURL: s.baseURL + "/" + link.ShortCode,
		},
	}

	return stats, nil
}

func (s *LinkService) DashboardOverview(ctx context.Context, userID int64) (*model.DashboardOverview, error) {
	return s.linkRepo.GetDashboardOverview(ctx, userID)
}

func (s *LinkService) DashboardTrend(ctx context.Context, userID int64, hours int) ([]map[string]any, error) {
	return s.linkRepo.GetTrendData(ctx, userID, hours)
}

func (s *LinkService) DashboardGeo(ctx context.Context, userID int64) ([]map[string]any, error) {
	return s.linkRepo.GetGeoData(ctx, userID)
}

func (s *LinkService) DashboardDevices(ctx context.Context, userID int64) ([]map[string]any, error) {
	return s.linkRepo.GetDeviceData(ctx, userID)
}
