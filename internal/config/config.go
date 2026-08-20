package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Notebook is one Linny notebook served by the process. Each has its own corpus
// working tree (with its own git-safety guard) and its own disposable index.
type Notebook struct {
	Name       string `json:"name"`
	CorpusPath string `json:"corpusPath"`
	StateDir   string `json:"stateDir"`
}

// Config is the full server configuration. PublicHostname is a plain
// configuration value — no hostname is ever compiled in as a constant.
type Config struct {
	ListenAddress  string     `json:"listenAddress"`
	Port           int        `json:"port"`
	TokensFile     string     `json:"tokensFile"`
	LogLevel       string     `json:"logLevel"`
	ReadOnly       bool       `json:"readOnly"`
	PublicHostname string     `json:"publicHostname"`
	Notebooks      []Notebook `json:"notebooks"`
}

// Defaults returns a config with the non-notebook defaults filled in.
func Defaults() Config {
	return Config{
		ListenAddress: "127.0.0.1",
		Port:          8765,
		LogLevel:      "info",
	}
}

// Load reads and validates a JSON config file.
func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Defaults()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// FromFlags builds a single-notebook configuration named "default" from the
// individual serve flags. It is the no-`--config` convenience path.
func FromFlags(listen string, port int, tokensFile, logLevel string, readOnly bool, corpusPath, stateDir string) (Config, error) {
	cfg := Defaults()
	cfg.ListenAddress = listen
	cfg.Port = port
	cfg.TokensFile = tokensFile
	cfg.LogLevel = logLevel
	cfg.ReadOnly = readOnly
	cfg.Notebooks = []Notebook{{Name: "default", CorpusPath: corpusPath, StateDir: stateDir}}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks structural invariants. PublicHostname is intentionally not
// required — an unset hostname is valid.
func (c Config) Validate() error {
	if len(c.Notebooks) == 0 {
		return fmt.Errorf("config: at least one notebook is required")
	}
	seen := map[string]bool{}
	for i, nb := range c.Notebooks {
		if nb.Name == "" {
			return fmt.Errorf("config: notebook %d has an empty name", i)
		}
		if seen[nb.Name] {
			return fmt.Errorf("config: duplicate notebook name %q", nb.Name)
		}
		seen[nb.Name] = true
		if nb.CorpusPath == "" {
			return fmt.Errorf("config: notebook %q has an empty corpusPath", nb.Name)
		}
	}
	return nil
}

// Notebook returns the notebook with the given name, or false.
func (c Config) Notebook(name string) (Notebook, bool) {
	for _, nb := range c.Notebooks {
		if nb.Name == name {
			return nb, true
		}
	}
	return Notebook{}, false
}

// DefaultNotebook returns the first configured notebook. Validation guarantees
// at least one exists.
func (c Config) DefaultNotebook() Notebook { return c.Notebooks[0] }

// Resolve returns the notebook selected by name, or the default when name is
// empty. It errors if a non-empty name does not match.
func (c Config) Resolve(name string) (Notebook, error) {
	if name == "" {
		return c.DefaultNotebook(), nil
	}
	nb, ok := c.Notebook(name)
	if !ok {
		return Notebook{}, fmt.Errorf("config: no notebook named %q", name)
	}
	return nb, nil
}
