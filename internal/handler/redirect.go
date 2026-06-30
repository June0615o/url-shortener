package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/panhao/url-shortener/internal/cache"
	"github.com/panhao/url-shortener/internal/model"
	"github.com/panhao/url-shortener/internal/service"
	"github.com/panhao/url-shortener/internal/util"
)

type RedirectHandler struct {
	linkService *service.LinkService
	clickChan   chan<- model.ClickLog
	redisCache  *cache.RedisCache
	bloomFilter *cache.BloomFilter
}

func NewRedirectHandler(linkService *service.LinkService, clickChan chan<- model.ClickLog,
	redisCache *cache.RedisCache, bloomFilter *cache.BloomFilter) *RedirectHandler {
	return &RedirectHandler{
		linkService: linkService,
		clickChan:   clickChan,
		redisCache:  redisCache,
		bloomFilter: bloomFilter,
	}
}

// ServeRedirect handles GET /:short_code
func (h *RedirectHandler) ServeRedirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "shortCode")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	// Layer 0: Bloom filter — fast rejection of non-existent codes
	if h.bloomFilter != nil {
		mightExist, err := h.bloomFilter.MightExist(r.Context(), code)
		if err == nil && !mightExist {
			http.NotFound(w, r)
			return
		}
	}

	// Layer 1: Redis cache — fast lookup
	if h.redisCache != nil {
		cachedURL, err := h.redisCache.GetURL(r.Context(), code)
		if err == nil && cachedURL != "" {
			h.logClick(-1, r) // We don't have linkID from cache
			http.Redirect(w, r, cachedURL, http.StatusFound)
			return
		}
	}

	// Layer 2: PostgreSQL — authoritative source
	link, err := h.linkService.Get(r.Context(), code)
	if err != nil {
		log.Printf("Error fetching link %s: %v", code, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if link == nil {
		http.NotFound(w, r)
		return
	}

	// Backfill cache
	if h.redisCache != nil {
		if err := h.redisCache.SetURL(r.Context(), code, link.OriginalURL, link.ExpireAt); err != nil {
			log.Printf("Warning: failed to cache URL for %s: %v", code, err)
		}
	}

	// Check expiration
	if link.ExpireAt != nil && time.Now().After(*link.ExpireAt) {
		http.Error(w, "This link has expired", http.StatusGone)
		return
	}

	// Check if active
	if !link.IsActive {
		http.Error(w, "This link is no longer active", http.StatusGone)
		return
	}

	// Check password protection
	if link.PasswordHash != "" {
		password := r.URL.Query().Get("password")
		if password == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(passwordProtectedHTML(code)))
			return
		}
		if err := checkPassword(link.PasswordHash, password); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(passwordWrongHTML(code)))
			return
		}
	}

	// Apply smart routing rules
	targetURL := link.OriginalURL
	if link.Rules != "" && link.Rules != "[]" {
		if resolved := applyRules(link.Rules, r); resolved != "" {
			targetURL = resolved
		}
	}

	// Log click asynchronously
	h.logClick(link.ID, r)

	// Redirect
	statusCode := http.StatusFound // 302
	if link.RedirectType == "301" {
		statusCode = http.StatusMovedPermanently
	}
	http.Redirect(w, r, targetURL, statusCode)
}

func (h *RedirectHandler) logClick(linkID int64, r *http.Request) {
	click := model.ClickLog{
		LinkID:    linkID,
		IP:        getClientIP(r),
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
		ClickedAt: time.Now(),
	}

	select {
	case h.clickChan <- click:
	default:
		// Channel full, drop this click to avoid blocking the redirect
		log.Printf("Click log channel full, dropping click for link %d", linkID)
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Strip port from RemoteAddr
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

func checkPassword(hash, password string) error {
	if !util.CheckPassword(password, hash) {
		return errWrongPassword
	}
	return nil
}

var errWrongPassword = &wrongPasswordError{}

type wrongPasswordError struct{}

func (e *wrongPasswordError) Error() string { return "wrong password" }

func applyRules(rulesJSON string, r *http.Request) string {
	// TODO: Implement rules engine in Phase 6
	return ""
}

const passwordProtectedHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>需要密码</title>
<style>
  body { font-family: -apple-system, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #f5f5f5; }
  .box { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 12px rgba(0,0,0,0.1); text-align: center; max-width: 400px; width: 90%; }
  input { width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 16px; margin: 12px 0; box-sizing: border-box; }
  button { width: 100%; padding: 12px; background: #4f46e5; color: white; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; }
  button:hover { background: #4338ca; }
  h2 { margin: 0 0 8px; color: #1f2937; }
  p { color: #6b7280; margin: 0 0 16px; }
</style></head>
<body>
<div class="box"><h2>🔒 需要密码</h2><p>此链接受到密码保护</p>
<form method="GET"><input type="password" name="password" placeholder="请输入密码" required autofocus><button type="submit">访问</button></form></div>
</body></html>`

func passwordProtectedHTML(code string) string {
	return passwordProtectedHTMLTemplate
}

const passwordWrongHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>密码错误</title>
<style>
  body { font-family: -apple-system, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; background: #fef2f2; }
  .box { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 2px 12px rgba(0,0,0,0.1); text-align: center; max-width: 400px; width: 90%; }
  input { width: 100%; padding: 12px; border: 1px solid #fca5a5; border-radius: 6px; font-size: 16px; margin: 12px 0; box-sizing: border-box; }
  button { width: 100%; padding: 12px; background: #4f46e5; color: white; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; }
  button:hover { background: #4338ca; }
  h2 { margin: 0 0 8px; color: #dc2626; }
  p { color: #6b7280; margin: 0 0 16px; }
  .error { color: #dc2626; margin: 8px 0; }
</style></head>
<body>
<div class="box"><h2>❌ 密码错误</h2><p>请重试</p>
<form method="GET"><input type="password" name="password" placeholder="请输入密码" required autofocus><p class="error">密码不正确</p><button type="submit">重试</button></form></div>
</body></html>`

func passwordWrongHTML(code string) string {
	return passwordWrongHTMLTemplate
}
