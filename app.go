package main

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"wr_tool/internal/cleaner"
	"wr_tool/internal/differ"
	"wr_tool/internal/exporter"
	"wr_tool/internal/license"
	"wr_tool/internal/parser"
)

type App struct {
	ctx      context.Context
	exporter exporter.Exporter
}

type ComparisonResult struct {
	OldFile   string             `json:"oldFile"`
	NewFile   string             `json:"newFile"`
	Text      differ.Result      `json:"text"`
	Tables    []differ.TableDiff `json:"tables"`
	Truncated bool               `json:"truncated"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ErrTrialExpired is what the UI shows once the trial window is over.
var ErrTrialExpired = errors.New("试用期已结束，加微信 wangran38 获取激活码后继续使用")

type LicenseStatus struct {
	MachineCode string `json:"machineCode"`
	Activated   bool   `json:"activated"`
	TrialDays   int    `json:"trialDays"`
	Expired     bool   `json:"expired"`
}

func (a *App) PickDocument() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择合同文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "合同文件 (*.docx *.pdf *.txt)", Pattern: "*.docx;*.pdf;*.txt"},
		},
	})
}

func (a *App) CompareFiles(fileA, fileB string, opts cleaner.Options) (*ComparisonResult, error) {
	st, err := a.LicenseStatus()
	if err != nil {
		return nil, err
	}
	if st.Expired {
		return nil, ErrTrialExpired
	}

	oldDoc, err := parser.Open(fileA)
	if err != nil {
		return nil, err
	}
	newDoc, err := parser.Open(fileB)
	if err != nil {
		return nil, err
	}

	return &ComparisonResult{
		OldFile: filepath.Base(fileA),
		NewFile: filepath.Base(fileB),
		Text:    differ.Diff(cleaner.Prepare(oldDoc, opts), cleaner.Prepare(newDoc, opts)),
		Tables:  differ.DiffTables(oldDoc.Tables, newDoc.Tables, opts),
	}, nil
}

func (a *App) ChooseExportPath(baseName, format string) (string, error) {
	pattern := "*.html"
	displayName := "HTML 报告 (*.html)"
	if format == "pdf" {
		pattern = "*.pdf"
		displayName = "PDF 报告 (*.pdf)"
	}

	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存比对报告",
		DefaultFilename: exporter.SuggestOutputPath("", baseName, format),
		Filters: []runtime.FileFilter{
			{DisplayName: displayName, Pattern: pattern},
		},
	})
}

func (a *App) ExportReport(result ComparisonResult, outPath, format string) error {
	report := exporter.Report{
		OldFile:   result.OldFile,
		NewFile:   result.NewFile,
		Text:      result.Text,
		Tables:    result.Tables,
		Generated: time.Now().Format("2006-01-02 15:04:05"),
	}
	if format == "pdf" {
		return a.exporter.ToPDF(report, outPath)
	}
	return a.exporter.ToHTML(report, outPath)
}

func (a *App) GetMachineCode() (string, error) {
	return license.MachineCode()
}

func (a *App) LicenseStatus() (*LicenseStatus, error) {
	code, err := license.MachineCode()
	if err != nil {
		return nil, err
	}
	activated, err := license.Status()
	if err != nil {
		return nil, err
	}

	st := &LicenseStatus{MachineCode: code, Activated: activated}
	if !activated {
		days, err := license.TrialDaysLeft()
		if err != nil {
			return nil, err
		}
		st.TrialDays = days
		st.Expired = days <= 0
	}
	return st, nil
}

// VerifyLicense checks the activation code against this machine and stores it on success.
func (a *App) VerifyLicense(key string) error {
	code, err := license.MachineCode()
	if err != nil {
		return err
	}
	return license.Activate(code, key)
}
