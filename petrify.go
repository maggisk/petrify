package main

import "os"
import "fmt"
import "strings"
import "io/ioutil"
import "path/filepath"
import "github.com/pkg/browser"

const CONFIG_FILE = ".petrify"

func copyStatics(staticDirs []string, buildDir string) {
	for _, dir := range staticDirs {
		parts := strings.SplitN(dir, ":", 2)
		src := parts[0]
		dst := parts[0]
		if len(parts) > 1 {
			dst = parts[1]
		}
		if !filepath.IsAbs(dst) {
			dst = filepath.Join(buildDir, dst)
		}
		verbose("Copying static '%s' -> '%s'", src, dst)
		CopyDir(src, dst)
	}
}

func main() {
	// create config
	config := DefaultConfig()
	config.ReadConfigFile(CONFIG_FILE)
	config.Normalize()
	config.ValidateForBuild()

	// set global VERBOSE var in util.go
	VERBOSE = config.Verbose || os.Getenv("VERBOSE") == "1"

	// set current working directory
	os.Chdir(config.CWD)

	// initialize temporary build directory
	buildDir, err := ioutil.TempDir("", "petrify-build-")
	checkError(err)
	config.BuildDir = buildDir
	defer os.RemoveAll(buildDir)

	// copy static files
	copyStatics(config.StaticDirs, config.BuildDir)

	// TODO: add option to start dev server from this process?

	// initialize crawler
	extractLinks := strings.TrimSpace(config.DefaultExtractLinks + " " + config.ExtractLinks)
	crawler := NewCrawler(config.ServerURL, config.BuildDir, extractLinks)
	crawler.CrawlAll(config.EntryPoints)

	// fetch special 404 page if there is one
	if len(config.Path404) > 0 {
		crawler.Crawl(config.Path404)
	}

	// preview build before deploy if configured to do so
	if config.PreviewBeforeDeploy {
		port := ServeStatic(config.BuildDir)
		browser.OpenURL(fmt.Sprintf("http://localhost:%d", port))
		info("Build preview running at http://localhost:%d", port)
		info("Press enter to continue")
		ReadLine()
	}

	// deploy
	config.ValidateForDeploy()
	if len(config.DeployToGithub) > 0 {
		deployToGithub(config)
	}

	// TODO: more deployment options
}
