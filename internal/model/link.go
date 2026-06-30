package model

import "time"

type Link struct {
	ID           int64      `json:"id"`
	ShortCode    string     `json:"short_code"`
	OriginalURL  string     `json:"original_url"`
	Title        string     `json:"title,omitempty"`
	Description  string     `json:"description,omitempty"`
	UserID       *int64     `json:"user_id,omitempty"`
	Domain       string     `json:"domain"`
	ExpireAt     *time.Time `json:"expire_at,omitempty"`
	PasswordHash string     `json:"-"`
	Rules        string     `json:"rules,omitempty"`
	RedirectType string     `json:"redirect_type"`
	IsActive     bool       `json:"is_active"`
	ClickCount   int64      `json:"click_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ClickLog struct {
	ID         int64     `json:"id"`
	LinkID     int64     `json:"link_id"`
	IP         string    `json:"ip,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Referer    string    `json:"referer,omitempty"`
	Country    string    `json:"country,omitempty"`
	City       string    `json:"city,omitempty"`
	DeviceType string    `json:"device_type,omitempty"`
	Browser    string    `json:"browser,omitempty"`
	OS         string    `json:"os,omitempty"`
	ClickedAt  time.Time `json:"clicked_at"`
}

type ClickStatsHourly struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	Hour      time.Time `json:"hour"`
	Clicks    int       `json:"clicks"`
	UniqueIPs int       `json:"unique_ips"`
}

type APIKey struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	KeyHash    string     `json:"-"`
	RateLimit  int        `json:"rate_limit"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Request/Response DTOs

type CreateLinkReq struct {
	URL          string `json:"url"`
	CustomCode   string `json:"custom_code,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	Domain       string `json:"domain,omitempty"`
	ExpireAt     string `json:"expire_at,omitempty"`
	Password     string `json:"password,omitempty"`
	RedirectType string `json:"redirect_type,omitempty"`
	Rules        []Rule `json:"rules,omitempty"`
}

type UpdateLinkReq struct {
	Title        *string `json:"title,omitempty"`
	Description  *string `json:"description,omitempty"`
	ExpireAt     *string `json:"expire_at,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	RedirectType *string `json:"redirect_type,omitempty"`
}

type LinkResp struct {
	Link
	ShortURL string `json:"short_url"`
}

type Rule struct {
	Priority int        `json:"priority"`
	Conditions []Condition `json:"conditions"`
	Action    Action     `json:"action"`
}

type Condition struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type Action struct {
	Type   string `json:"type"`
	Target string `json:"target"`
}

type PaginatedResp struct {
	Data       any   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterReq struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

type AuthResp struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	UserID   int64  `json:"user_id"`
}

type DashboardOverview struct {
	TotalLinks     int64 `json:"total_links"`
	TotalClicks    int64 `json:"total_clicks"`
	TodayClicks    int64 `json:"today_clicks"`
	ActiveLinks    int64 `json:"active_links"`
	ExpiredLinks   int64 `json:"expired_links"`
	AvgClicksPerDay float64 `json:"avg_clicks_per_day"`
}
