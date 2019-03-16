package main

import "os"
import "log"
import "strings"
import "github.com/BurntSushi/toml"

type Config struct {
	CWD                 string
	ServerURL           string
	EntryPoints         []string
	BuildDir            string
	StaticDirs          []string
	PreviewBeforeDeploy bool
	DefaultExtractLinks string
	ExtractLinks        string
	Path404             string
	DeployToGithub      string
	CNAME               string
	Verbose             bool
}

func DefaultConfig() *Config {
	return &Config{
		CWD:                 ".",
		EntryPoints:         []string{"/"},
		PreviewBeforeDeploy: true,
		DefaultExtractLinks: "*.href *.src", // undocumented -  here in case someone needs to override it
	}
}

func (config *Config) ReadConfigFile(path string) {
	if _, err := os.Stat(path); err == nil {
		info("Loading config file: " + path)
		_, err := toml.DecodeFile(path, config)
		checkError(err)
	}
}

func (config *Config) Normalize() {
	config.ServerURL = normalizeURL(config.ServerURL)
	if len(config.Path404) > 0 {
		config.Path404 = normalizePath(config.Path404)
	}
	for i, path := range config.EntryPoints {
		config.EntryPoints[i] = normalizePath(path)
	}
}

func (config *Config) ValidateForBuild() {
	if config.ServerURL == "" {
		log.Fatal("ServerURL configuration missing. Create a .petrify configuration file containing: ServerURL = \"http://url-to-dev-server\"")
	}

	if !config.hasDeploymentConfig() {
		warn("No deployment configuration. You will not be able to deploy your website")
	}
}

func (config *Config) ValidateForDeploy() {
	if !config.hasDeploymentConfig() {
		log.Fatal("Missing deployment configuration. Unable to deploy website")
	}
}

func (config *Config) hasDeploymentConfig() bool {
	return len(config.DeployToGithub /* + more when added */) > 0
}

func normalizePath(path string) string {
	return "/" + strings.TrimLeft(path, "/")
}

func normalizeURL(url string) string {
	return strings.TrimRight(url, "/")
}
