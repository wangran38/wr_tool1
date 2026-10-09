package exporter

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chromedp/chromedp"
)

// ToPDF renders the same HTML report and prints it through a local Chromium-based
// browser (Edge ships with Windows 10/11, so no extra install step is needed).
func (e Exporter) ToPDF(r Report, outPath string) error {
	bin, err := findChromium()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "contract-diff-report")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "report.html")
	if err := e.ToHTML(r, htmlPath); err != nil {
		return err
	}

	// The first action boots the browser and the browser lives as long as this context,
	// so it must stay timeout-free; only the print step gets a deadline.
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.ExecPath(bin),
		chromedp.Headless,
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("disable-extensions", true),
	)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Do(ctx,
		chromedp.Navigate(fileURL(htmlPath)),
		chromedp.WaitReady("body"),
	); err != nil {
		return fmt.Errorf("open report in browser: %w", err)
	}

	printCtx, printCancel := context.WithTimeout(ctx, 60*time.Second)
	defer printCancel()

	pdf, err := chromedp.Run(printCtx, chromedp.PrintToPDF(
		chromedp.PDFPaper(chromedp.PaperA4),
		chromedp.PDFMargin(0.4),
		chromedp.PDFPrintBackground(),
	))
	if err != nil {
		return fmt.Errorf("print report to pdf: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, pdf, 0o644)
}

var ErrNoBrowser = errors.New("未找到 Chrome 或 Edge，无法导出 PDF")

func findChromium() (string, error) {
	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && !info.IsDir() {
				return c, nil
			}
		}
	}

	for _, name := range []string{"msedge", "google-chrome", "chrome", "chromium"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", ErrNoBrowser
}

func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(abs)}
	return u.String()
}
