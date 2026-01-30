package collector

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)

// common secret patterns to match
var secretPatterns = []*regexp.Regexp{
	// Generic password/secret patterns
	regexp.MustCompile(`(?i)(password|passwd|pwd|passphrase)`),
	regexp.MustCompile(`(?i)(secret|token|key|api_key|apikey|auth_token|private_key)`),
	regexp.MustCompile(`(?i)(auth|credential|cred|login|userpass)`),
	regexp.MustCompile(`(?i)(private|priv|secret_key|secretToken)`),
	regexp.MustCompile(`(?i)(access_token|refresh_token|id_token)`),
	regexp.MustCompile(`(?i)(connection_string|conn_str|db_uri|database_url)`),

	// Cloud provider credentials
	regexp.MustCompile(`(?i)(aws_access_key_id|aws_secret_access_key|aws_session_token)`),
	regexp.MustCompile(`(?i)(azure_client_secret|azure_client_id|azure_tenant_id)`),
	regexp.MustCompile(`(?i)(gcp_service_account|gcp_key|google_api_key|google_credentials)`),
	regexp.MustCompile(`(?i)(digitalocean_token|do_api_key)`),
	regexp.MustCompile(`(?i)(heroku_api_key|heroku_oauth_token)`),

	// Common AI API keys
	regexp.MustCompile(`(?i)(openai_api_key|openai_secret)`),
	regexp.MustCompile(`(?i)(huggingface_api_key|hf_token|hf_api_token)`),
	regexp.MustCompile(`(?i)(anthropic_api_key|claude_api_key)`),
	regexp.MustCompile(`(?i)(cohere_api_key|cohere_token)`),
	regexp.MustCompile(`(?i)(stability_ai_key|stability_api_key)`),
	regexp.MustCompile(`(?i)(replicate_api_token|replicate_key)`),

	// Database credentials
	regexp.MustCompile(`(?i)(db_password|db_user|db_secret|db_token)`),
	regexp.MustCompile(`(?i)(mysql_pass|postgres_password|mongo_uri)`),

	// JWT, OAuth, and SSO tokens
	regexp.MustCompile(`(?i)(jwt_secret|jwt_token|oauth_token|sso_token)`),

	// Misc API keys / secrets
	regexp.MustCompile(`(?i)(stripe_secret|stripe_key|paypal_client_secret|paypal_key)`),
	regexp.MustCompile(`(?i)(twilio_auth_token|twilio_sid|twilio_key)`),
	regexp.MustCompile(`(?i)(slack_token|slack_webhook|slack_signing_secret)`),
	regexp.MustCompile(`(?i)(discord_token|discord_webhook|discord_secret)`),
	regexp.MustCompile(`(?i)(telegram_bot_token|telegram_api_key)`),
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
