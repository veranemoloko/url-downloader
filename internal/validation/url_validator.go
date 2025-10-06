package validation

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("safe_url", validateSafeURL)
}

func ValidateURLs(urls []string) error {
	for _, u := range urls {
		if err := validate.Var(u, "required,safe_url"); err != nil {
			return fmt.Errorf("invalid URL %q: %w", u, err)
		}
	}
	return nil
}

func validateSafeURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	host := u.Hostname()

	forbiddenHosts := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
		"169.254.169.254",
	}

	for _, forbidden := range forbiddenHosts {
		if strings.EqualFold(host, forbidden) {
			return false
		}
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsPrivate() || ip.IsLoopback() {
			return false
		}
	}

	return true
}
