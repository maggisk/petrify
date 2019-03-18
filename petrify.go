package main

import "os"
import "fmt"
import "strings"
import "io/ioutil"
import "path/filepath"
import "runtime/debug"
import "github.com/pkg/browser"

const CONFIG_FILE = ".petrify"

func copyStatics(staticDirs []string, buildDir string) {
	for _, dir := range staticDirs {
		parts := strings.SplitN(dir, ":", 2)
		src, dst := parts[0], parts[0]
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

func Build(config *Config) bool {
	// initialize build directory
	isTempDir := config.BuildDir == ""
	if isTempDir {
		buildDir, err := ioutil.TempDir("", "petrify-build-")
		checkError(err)
		config.BuildDir = buildDir
	} else {
		checkError(os.RemoveAll(config.BuildDir))
		checkError(os.MkdirAll(config.BuildDir, DIR_BITMASK))
	}

	// copy static files
	copyStatics(config.StaticDirs, config.BuildDir)

	// initialize crawler
	extractLinks := strings.TrimSpace(config.DefaultExtractLinks + " " + config.ExtractLinks)
	crawler := NewCrawler(config.ServerURL, config.BuildDir, extractLinks)
	crawler.CrawlAll(config.EntryPoints)

	// fetch special 404 page if there is one
	if len(config.Path404) > 0 {
		crawler.Crawl(config.Path404)
	}

	info("Build finished in %s", config.BuildDir)
	return isTempDir
}

func Preview(config *Config) {
	port := ServeStatic(config.BuildDir)
	browser.OpenURL(fmt.Sprintf("http://localhost:%d", port))
	info("Build preview running at http://localhost:%d", port)
}

func Wizard(config *Config) {
	// start by building once
	if Build(config) {
		defer os.RemoveAll(config.BuildDir)
	}

	// start static server
	if config.PreviewBeforeDeploy {
		Preview(config)
	}

	// enter deploy build loop until deployed successfully
	for {
		if ReadYesNo(fmt.Sprintf("Deploy website to %s? (yes/no)", config.DeployToGithub)) {
			Deploy(config)
			break
		}
		if !ReadYesNo("Build static website? (yes/no)") {
			break
		}
		Build(config)
	}
}

func main() {
	command := "wizard"
	if len(os.Args) > 1 {
		command = strings.ToLower(os.Args[1])
	}

	if command == "wizard" {
		// TODO: We assume run by double-click when in wizard mode. can we do better?

		// set current working directory to the path of the petrify binary
		os.Chdir(filepath.Dir(os.Args[0]))

		// prevent window from immediately closing when run by double-clicking
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				fmt.Println(err)
			}
			fmt.Println("Press enter to exit...")
			ReadLine()
		}()
	}

	// create config
	config := DefaultConfig()
	config.ReadConfigFile(CONFIG_FILE)
	config.Normalize()
	config.ValidateForBuild()

	// set global VERBOSE var in util.go
	VERBOSE = config.Verbose || os.Getenv("VERBOSE") == "1"

	// set current working directory
	os.Chdir(config.CWD)

	// TODO: add option to launch dev server from this process?

	switch command {
	case "build":
		Build(config)
	case "preview":
		Preview(config)
		select {} // wait forever
	case "deploy":
		Deploy(config)
	case "wizard":
		Wizard(config)
	}
}
