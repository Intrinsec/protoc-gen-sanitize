package main

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/intrinsec/protoc-gen-sanitize/sanitize"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

// SanitizePlugin adds Sanitize methods on PB
type SanitizeModule struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
	tpl *template.Template
}

// Sanitize returns an initialized SanitizePlugin
func Sanitize() *SanitizeModule { return &SanitizeModule{ModuleBase: &pgs.ModuleBase{}} }

func (p *SanitizeModule) InitContext(c pgs.BuildContext) {

	log.Println("InitContext")
	p.ModuleBase.InitContext(c)
	p.ctx = pgsgo.InitContext(c.Parameters())

	tpl := template.New("Sanitize").Funcs(map[string]interface{}{
		"package":     p.ctx.PackageName,
		"name":        p.ctx.Name,
		"sanitizer":   p.sanitizer,
		"initializer": p.initializer,
	})

	p.tpl = template.Must(tpl.Parse(SanitizeTpl))
}

// Name satisfies the generator.Plugin interface.
func (p *SanitizeModule) Name() string { return "Sanitize" }

func (p *SanitizeModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {

	log.Println("Execute")

	for _, t := range targets {
		p.generate(t)
	}

	return p.Artifacts()
}

func (p *SanitizeModule) generate(f pgs.File) {

	log.Println("generate")

	if len(f.Messages()) == 0 {
		return
	}

	name := p.ctx.OutputPath(f).SetExt(".sanitize.go")

	log.Println(name)

	p.AddGeneratorTemplateFile(name.String(), p.tpl, f)
}

func (p *SanitizeModule) initializer(m pgs.Message) string {
	html := false
	text := false

	for _, f := range m.Fields() {
		var kind sanitize.Sanitization

		if ok, err := f.Extension(sanitize.E_Kind, &kind); ok && err == nil {
			switch kind {
			case sanitize.Sanitization_NONE:
				break
			case sanitize.Sanitization_HTML:
				html = true
				break
			case sanitize.Sanitization_TEXT:
				text = true
				break
			}
		}
	}
	out := make([]string, 0)
	if html {
		out = append(out, "htmlSanitize := bluemonday.UGCPolicy()")
	}
	if text {
		out = append(out, "textSanitize := bluemonday.StrictPolicy()")
	}
	return strings.Join(out, "\n	")
}

func (p *SanitizeModule) sanitizer(f pgs.Field) string {

	if f.Type().ProtoType() == pgs.StringT {
		var kind sanitize.Sanitization

		name := p.ctx.Name(f)

		if ok, err := f.Extension(sanitize.E_Kind, &kind); ok && err == nil {
			switch kind {
			case sanitize.Sanitization_NONE:
				return ""
			case sanitize.Sanitization_HTML:
				return fmt.Sprintf("m.%s = htmlSanitize.Sanitize(m.%s)", name, name)
			case sanitize.Sanitization_TEXT:
				return fmt.Sprintf("m.%s = textSanitize.Sanitize(m.%s)", name, name)
			}
		}
	}

	return ""
}

const SanitizeTpl = `package {{ package . }}
import (
	"github.com/microcosm-cc/bluemonday"
)

{{ range .AllMessages }}
func (m *{{ name . }}) Sanitize() error {
	{{ initializer . }}

{{ range .Fields }}
    {{ sanitizer . }}
{{ end }}
    return nil
}

{{ end }}
`
