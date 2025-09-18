package cookie

import (
	"net/http"
	"time"
)

const (
	// Cookie names for storing tokens
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

type CookieOptions struct {
	Domain   string
	Path     string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

func GetDefaultCookieOptions() CookieOptions {
	// Check if we're in production
	// cfg := config.Get()
	// isProduction := cfg.IsProduction()

	return CookieOptions{
		Domain:   "", // Empty means current domain
		Path:     "/",
		Secure:   false,                // Only secure in production (HTTPS)
		HTTPOnly: true,                 // Prevent XSS attacks
		SameSite: http.SameSiteLaxMode, // CSRF protection
	}
}

func SetAccessTokenCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	options := GetDefaultCookieOptions()
	options.MaxAge = int(time.Until(expiresAt).Seconds())

	cookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    token,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}

	http.SetCookie(w, cookie)
}

func SetRefreshTokenCookie(w http.ResponseWriter, token string, isRememberMe bool) {
	options := GetDefaultCookieOptions()

	var maxAge int
	if isRememberMe {
		maxAge = int((30 * 24 * time.Hour).Seconds()) // 30 days
	} else {
		maxAge = int((30 * time.Minute).Seconds()) // 30 minutes
	}
	options.MaxAge = maxAge

	cookie := &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    token,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}

	http.SetCookie(w, cookie)
}

func GetAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AccessTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func ClearAuthCookies(w http.ResponseWriter) {
	options := GetDefaultCookieOptions()

	accessCookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   -1, // Expire immediately
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}
	http.SetCookie(w, accessCookie)

	refreshCookie := &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   -1, // Expire immediately
		Secure:   options.Secure,
		HttpOnly: options.HTTPOnly,
		SameSite: options.SameSite,
	}
	http.SetCookie(w, refreshCookie)
}

func SetUserPreferenceCookie(w http.ResponseWriter, name, value string, maxAge int) {
	options := GetDefaultCookieOptions()

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		Domain:   options.Domain,
		MaxAge:   maxAge,
		Secure:   options.Secure,
		HttpOnly: false, // Allow JavaScript access for preferences
		SameSite: options.SameSite,
	}

	http.SetCookie(w, cookie)
}
