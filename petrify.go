package main

import "os"
import "bufio"
import "fmt"
import "net"
import "net/http"
import "strings"
import "path/filepath"
import "github.com/pkg/browser"
import "github.com/BurntSushi/toml"

type Config struct {
	CWD               string
	ServerURL         string
	EntryPoints       []string
	BuildDir          string
	StaticDirs        []string
	Preview           bool
	PreviewServerPort int
	DefaultLinkRules  string
	LinkRules         string
	Path404           string
	DeployToGithub    string
	CNAME             string
}

func defaultConfig() *Config {
	return &Config{
		CWD:               ".",
		ServerURL:         "http://localhost:5000",
		EntryPoints:       []string{"/"},
		BuildDir:          "",
		StaticDirs:        []string{},
		Preview:           true,
		PreviewServerPort: 0,
		DefaultLinkRules:  "*.href *.src", // undocumented, should be no need to override
		LinkRules:         "",             // additional link rules e.g. for crawler to follow data-src
		Path404:           "",
	}
}

func prepBuildDir(buildDir string) {
	os.RemoveAll(buildDir)
	os.MkdirAll(buildDir, DIR_BITMASK)
}

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
		CopyDir(src, dst)
	}
}

func getListenerAndPort(port int) (net.Listener, int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	checkError(err)
	if port == 0 {
		// os picks a random port if we supply a 0, so we need to find out what port we were assigned
		port = listener.Addr().(*net.TCPAddr).Port
	}
	return listener, port
}

func main() {
	config := defaultConfig()
	if _, err := os.Stat(".petrify"); err == nil {
		fmt.Println("loading config file: .petrify")
		_, err := toml.DecodeFile(".petrify", config)
		checkError(err)
	}

	// TODO: validate config, assert/fix leading and trailing slashes and check that nothing critical is missing

	os.Chdir(config.CWD)

	prepBuildDir(config.BuildDir)
	copyStatics(config.StaticDirs, config.BuildDir)

	// TODO: add option to start dev server from this process?

	linkRules := strings.TrimSpace(config.DefaultLinkRules + " " + config.LinkRules)
	crawler := NewCrawler(config.ServerURL, config.BuildDir, linkRules)
	crawler.CrawlAll(config.EntryPoints)

	if len(config.Path404) > 0 {
		crawler.Crawl(config.Path404)
	}

	if config.Preview {
		go func() {
			listener, port := getListenerAndPort(config.PreviewServerPort)
			fs := http.FileServer(http.Dir(config.BuildDir))
			http.Handle("/", fs)
			browser.OpenURL(fmt.Sprintf("http://localhost:%d", port))
			http.Serve(listener, nil)
			fmt.Printf("Preview running at http://localhost:%d\n", port)
			fmt.Println("Press enter to continue")
		}()

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}

	if len(config.DeployToGithub) > 0 {
		deployToGithub(config)
	}
}
