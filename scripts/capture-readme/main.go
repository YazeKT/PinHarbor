// Command capture-readme creates reproducible PinHarbor repository screenshots.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	baseURL := flag.String("url", "http://127.0.0.1:8080", "running PinHarbor URL")
	outputDir := flag.String("out", "docs/assets", "screenshot output directory")
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("start Playwright: %v", err)
	}
	defer func() {
		if stopErr := pw.Stop(); stopErr != nil {
			log.Printf("stop Playwright: %v", stopErr)
		}
	}()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("launch Chromium: %v", err)
	}
	defer func() {
		if closeErr := browser.Close(); closeErr != nil {
			log.Printf("close Chromium: %v", closeErr)
		}
	}()

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1440, Height: 900},
	})
	if err != nil {
		log.Fatalf("create page: %v", err)
	}

	if _, err = page.Goto(*baseURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("open PinHarbor: %v", err)
	}

	if err = page.Fill("#name", "Cape Town creative businesses"); err != nil {
		log.Fatalf("fill example job name: %v", err)
	}
	if err = page.Fill("#keywords", "graphic design studios in Cape Town\nsign makers in Cape Town"); err != nil {
		log.Fatalf("fill example searches: %v", err)
	}

	if err = capture(page, filepath.Join(*outputDir, "pinharbor-dark.png")); err != nil {
		log.Fatal(err)
	}
	if err = page.Click("#theme-toggle"); err != nil {
		log.Fatalf("switch to light theme: %v", err)
	}
	if err = capture(page, filepath.Join(*outputDir, "pinharbor-light.png")); err != nil {
		log.Fatal(err)
	}
	if err = captureSocialPreview(browser, *outputDir); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("captured PinHarbor screenshots in %s\n", *outputDir)
}

func captureSocialPreview(browser playwright.Browser, outputDir string) error {
	sourcePath, err := filepath.Abs(filepath.Join(outputDir, "social-preview.svg"))
	if err != nil {
		return fmt.Errorf("resolve social preview source: %w", err)
	}

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 640},
	})
	if err != nil {
		return fmt.Errorf("create social preview page: %w", err)
	}

	sourceURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(sourcePath)}).String()
	if _, err = page.Goto(sourceURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	}); err != nil {
		return fmt.Errorf("open social preview source: %w", err)
	}

	_, err = page.Screenshot(playwright.PageScreenshotOptions{
		Animations: playwright.ScreenshotAnimationsDisabled,
		Path:       playwright.String(filepath.Join(outputDir, "social-preview.png")),
		Scale:      playwright.ScreenshotScaleCss,
	})
	if err != nil {
		return fmt.Errorf("capture social preview: %w", err)
	}
	return nil
}

func capture(page playwright.Page, path string) error {
	_, err := page.Screenshot(playwright.PageScreenshotOptions{
		Animations: playwright.ScreenshotAnimationsDisabled,
		Path:       playwright.String(path),
		Scale:      playwright.ScreenshotScaleCss,
	})
	if err != nil {
		return fmt.Errorf("capture %s: %w", path, err)
	}
	return nil
}
