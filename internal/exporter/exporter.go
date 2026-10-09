package exporter

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"wr_tool/internal/differ"
)

type Report struct {
	OldFile   string
	NewFile   string
	Text      differ.Result
	Tables    []differ.TableDiff
	Generated string
}

type Exporter struct{}

const htmlTemplate = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>合同比对报告</title>
<style>
body{font-family:"Microsoft YaHei",sans-serif;margin:24px;color:#111}
h1{font-size:20px}h2{font-size:16px;margin-top:28px}
.meta{color:#555;font-size:13px}
table{border-collapse:collapse;width:100%;margin-top:12px}
th,td{border:1px solid #ddd;padding:6px 8px;font-size:13px;vertical-align:top;text-align:left}
.equal td{background:#fff}.insert td{background:#e6ffed}.delete td{background:#ffebe9}
.replace td{background:#fff8c5}
td.miss{background:#f6f8fa;color:#999}
.seg-del{background:#ffc1c1;text-decoration:line-through}
.seg-ins{background:#9ff0b6}
@media print{
 body{margin:0}
 *{-webkit-print-color-adjust:exact;print-color-adjust:exact}
 tr{break-inside:avoid}
 h2,h3{break-after:avoid}
}
</style></head><body>
<h1>合同比对报告</h1>
<p class="meta">旧文件：{{.OldFile}}<br>新文件：{{.NewFile}}<br>生成时间：{{.Generated}}</p>
<p class="meta">一致 {{.Text.Stats.Equal}} ／ 新增 {{.Text.Stats.Insert}} ／ 删除 {{.Text.Stats.Delete}} ／ 修改 {{.Text.Stats.Replace}} ／ 相似度 {{printf "%.1f%%" (mul .Text.Stats.Similarity 100.0)}}</p>
<h2>正文差异</h2>
<table><thead><tr><th style="width:50%">原文</th><th style="width:50%">对比文</th></tr></thead><tbody>
{{range .Text.Pairs}}<tr class="{{.Kind}}">
<td>{{if .OldSegs}}{{range .OldSegs}}<span class="{{segClass .Kind}}">{{.Text}}</span>{{end}}{{else if .Old}}{{.Old}}{{else}}（无）{{end}}</td>
<td>{{if .NewSegs}}{{range .NewSegs}}<span class="{{segClass .Kind}}">{{.Text}}</span>{{end}}{{else if .New}}{{.New}}{{else}}（无）{{end}}</td></tr>{{end}}
</tbody></table>
{{if .Tables}}<h2>表格差异</h2>
{{range $t := .Tables}}<h3>表格 {{add $t.Index 1}}（{{$t.Kind}}）</h3>
{{if $t.Rows}}<table><tbody>
{{range $r := $t.Rows}}<tr class="{{$r.Kind}}"><td>{{join $r.Old}}</td><td>{{join $r.New}}</td></tr>{{end}}
</tbody></table>{{else}}<p class="meta">无差异</p>{{end}}
{{end}}{{end}}
</body></html>`

func (e Exporter) ToHTML(r Report, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return e.render(f, r)
}

func (Exporter) render(w io.Writer, r Report) error {
	funcs := template.FuncMap{
		"mul":  func(a, b float64) float64 { return a * b },
		"add":  func(a, b int) int { return a + b },
		"join": func(cells []string) string { return strings.Join(cells, " | ") },
		"segClass": func(k differ.Kind) string {
			switch k {
			case differ.Insert:
				return "seg-ins"
			case differ.Delete:
				return "seg-del"
			}
			return ""
		},
	}

	tpl, err := template.New("report").Funcs(funcs).Parse(htmlTemplate)
	if err != nil {
		return err
	}
	if err := tpl.Execute(w, r); err != nil {
		return fmt.Errorf("render report: %w", err)
	}
	return nil
}

func SuggestOutputPath(dir, base, format string) string {
	ext := "." + strings.ToLower(format)
	return filepath.Join(dir, base+"_比对报告"+ext)
}
