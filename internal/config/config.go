// Package config loads and saves the global shipquiz config and per-repo overrides.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
)

// Quiz is the grading and timing policy.
type Quiz struct {
	MinQuestions   int     `toml:"min_questions"`
	MaxQuestions   int     `toml:"max_questions"`
	PassingAverage float64 `toml:"passing_average"`
	TimeoutSeconds int     `toml:"timeout_seconds"`
}

// Hooks selects which git hooks shipquiz runs.
type Hooks struct {
	PrePush bool `toml:"pre_push"`
}

// Config is the resolved shipquiz configuration.
type Config struct {
	Provider      string   `toml:"provider"`
	Model         string   `toml:"model"`
	BaseURL       string   `toml:"base_url,omitempty"`
	Quiz          Quiz     `toml:"quiz"`
	Hooks         Hooks    `toml:"hooks"`
	SkipPaths     []string `toml:"skip_paths,omitempty"`
	QuestionStyle string   `toml:"question_style,omitempty"`
}

// Default returns the v1 defaults from the implementation plan.
func Default() Config {
	return Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5",
		Quiz: Quiz{
			MinQuestions:   1,
			MaxQuestions:   5,
			PassingAverage: 1.5,
			TimeoutSeconds: 60,
		},
		Hooks:         Hooks{PrePush: true},
		QuestionStyle: "default",
	}
}

// DefaultModel is the fast default for a provider when the user has not chosen one.
func DefaultModel(provider string) string {
	switch provider {
	case "openai":
		return "gpt-4.1-mini"
	case "gemini":
		return "gemini-2.5-flash"
	case "xai":
		return "grok-4.3"
	default:
		return "claude-sonnet-4-5"
	}
}

// Dir is the global config directory (~/.config/shipquiz, or SHIPQUIZ_CONFIG_DIR).
func Dir() (string, error) {
	if dir := os.Getenv("SHIPQUIZ_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "shipquiz"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "shipquiz"), nil
}

// Path is the global config.toml location.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the global config. A missing file returns Default.
func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("read %s: %w", path, err)
	}
	fillMissing(&cfg, meta)
	return cfg, nil
}

// LoadForRepo loads the global config and applies .shipquiz.toml when present.
func LoadForRepo(repoRoot string) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	if repoRoot == "" {
		return cfg, nil
	}
	path := filepath.Join(repoRoot, ".shipquiz.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := applyRepo(&cfg, data); err != nil {
		return cfg, fmt.Errorf("read %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes the global config with mode 0600.
func Save(cfg Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.toml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

// Set updates one dotted key on cfg.
func Set(cfg *Config, key, value string) error {
	switch key {
	case "provider":
		cfg.Provider = value
	case "model":
		cfg.Model = value
	case "base_url":
		cfg.BaseURL = value
	case "question_style":
		cfg.QuestionStyle = value
	case "quiz.min_questions":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("quiz.min_questions: %w", err)
		}
		cfg.Quiz.MinQuestions = n
	case "quiz.max_questions":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("quiz.max_questions: %w", err)
		}
		cfg.Quiz.MaxQuestions = n
	case "quiz.passing_average":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("quiz.passing_average: %w", err)
		}
		cfg.Quiz.PassingAverage = n
	case "quiz.timeout_seconds":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("quiz.timeout_seconds: %w", err)
		}
		cfg.Quiz.TimeoutSeconds = n
	case "hooks.pre_push":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("hooks.pre_push: %w", err)
		}
		cfg.Hooks.PrePush = b
	default:
		return fmt.Errorf("unknown key %q", key)
	}
	return nil
}

func fillMissing(cfg *Config, meta toml.MetaData) {
	d := Default()
	if cfg.Provider == "" {
		cfg.Provider = d.Provider
	}
	if cfg.Model == "" {
		cfg.Model = d.Model
	}
	if !meta.IsDefined("quiz", "min_questions") || cfg.Quiz.MinQuestions == 0 {
		cfg.Quiz.MinQuestions = d.Quiz.MinQuestions
	}
	if !meta.IsDefined("quiz", "max_questions") || cfg.Quiz.MaxQuestions == 0 {
		cfg.Quiz.MaxQuestions = d.Quiz.MaxQuestions
	}
	if !meta.IsDefined("quiz", "passing_average") || cfg.Quiz.PassingAverage == 0 {
		cfg.Quiz.PassingAverage = d.Quiz.PassingAverage
	}
	if !meta.IsDefined("quiz", "timeout_seconds") || cfg.Quiz.TimeoutSeconds == 0 {
		cfg.Quiz.TimeoutSeconds = d.Quiz.TimeoutSeconds
	}
	if !meta.IsDefined("hooks", "pre_push") {
		cfg.Hooks.PrePush = true
	}
	if cfg.QuestionStyle == "" {
		cfg.QuestionStyle = d.QuestionStyle
	}
	if cfg.Quiz.MaxQuestions > 5 {
		cfg.Quiz.MaxQuestions = 5
	}
	if cfg.Quiz.MinQuestions < 1 {
		cfg.Quiz.MinQuestions = 1
	}
	if cfg.Quiz.MinQuestions > cfg.Quiz.MaxQuestions {
		cfg.Quiz.MinQuestions = cfg.Quiz.MaxQuestions
	}
}

type repoFile struct {
	Provider      *string  `toml:"provider"`
	Model         *string  `toml:"model"`
	BaseURL       *string  `toml:"base_url"`
	SkipPaths     []string `toml:"skip_paths"`
	QuestionStyle *string  `toml:"question_style"`
	Quiz          *struct {
		MinQuestions   *int     `toml:"min_questions"`
		MaxQuestions   *int     `toml:"max_questions"`
		PassingAverage *float64 `toml:"passing_average"`
		TimeoutSeconds *int     `toml:"timeout_seconds"`
	} `toml:"quiz"`
	Hooks *struct {
		PrePush *bool `toml:"pre_push"`
	} `toml:"hooks"`
}

func applyRepo(cfg *Config, data []byte) error {
	var over repoFile
	if _, err := toml.Decode(string(data), &over); err != nil {
		return err
	}
	if over.Provider != nil {
		cfg.Provider = *over.Provider
	}
	if over.Model != nil {
		cfg.Model = *over.Model
	}
	if over.BaseURL != nil {
		cfg.BaseURL = *over.BaseURL
	}
	if over.SkipPaths != nil {
		cfg.SkipPaths = over.SkipPaths
	}
	if over.QuestionStyle != nil {
		cfg.QuestionStyle = *over.QuestionStyle
	}
	if over.Quiz != nil {
		if over.Quiz.MinQuestions != nil {
			cfg.Quiz.MinQuestions = *over.Quiz.MinQuestions
		}
		if over.Quiz.MaxQuestions != nil {
			cfg.Quiz.MaxQuestions = *over.Quiz.MaxQuestions
		}
		if over.Quiz.PassingAverage != nil {
			cfg.Quiz.PassingAverage = *over.Quiz.PassingAverage
		}
		if over.Quiz.TimeoutSeconds != nil {
			cfg.Quiz.TimeoutSeconds = *over.Quiz.TimeoutSeconds
		}
	}
	if over.Hooks != nil && over.Hooks.PrePush != nil {
		cfg.Hooks.PrePush = *over.Hooks.PrePush
	}
	return nil
}
