package collector

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)

// Common secret patterns to mask
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|pwd)`),
	regexp.MustCompile(`(?i)(secret|token|key|api_key|apikey)`),
	regexp.MustCompile(`(?i)(auth|credential|cred)`),
	regexp.MustCompile(`(?i)(private|priv)`),
	regexp.MustCompile(`(?i)(access_token|refresh_token)`),
	regexp.MustCompile(`(?i)(connection_string|conn_str)`),
}

func (c *Collector) collectEnvVars(ctx context.Context) (map[string]models.EnvVar, error) {
	envVars := make(map[string]models.EnvVar)

	// Compile include/exclude patterns
	var includePatterns, excludePatterns []*regexp.Regexp
	
	for _, pattern := range c.config.EnvVars.Include {
		if re, err := regexp.Compile(pattern); err == nil {
			includePatterns = append(includePatterns, re)
		}
	}
	
	for _, pattern := range c.config.EnvVars.Exclude {
		if re, err := regexp.Compile(pattern); err == nil {
			excludePatterns = append(excludePatterns, re)
		}
	}

	for _, env := range os.Environ() {
		select {
		case <-ctx.Done():
			return envVars, ctx.Err()
		default:
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		value := parts[1]

		// Check include patterns (if any defined, only include matches)
		if len(includePatterns) > 0 {
			matched := false
			for _, re := range includePatterns {
				if re.MatchString(name) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Check exclude patterns
		excluded := false
		for _, re := range excludePatterns {
			if re.MatchString(name) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Mask secrets if enabled
		if c.config.EnvVars.MaskSecrets && isSecretVar(name) {
			value = maskValue(value)
		}

		envVars[name] = models.EnvVar{
			Name:   name,
			Value:  value,
			Exists: true,
		}
	}

	return envVars, nil
}

func isSecretVar(name string) bool {
	for _, re := range secretPatterns {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}