package main

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/intrinsec/protoc-gen-sanitize/sanitize"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
)

// SanitizeModule adds Sanitize methods on PB
type SanitizeModule struct {
	*pgs.ModuleBase
	ctx pgsgo.Context
	tpl *template.Template
}

// Sanitize returns an initialized SanitizePlugin
func Sanitize() *SanitizeModule { return &SanitizeModule{ModuleBase: &pgs.ModuleBase{}} }

// InitContext populates the module with needed context and fields
func (p *SanitizeModule) InitContext(c pgs.BuildContext) {
	c.Log("InitContext")
	p.ModuleBase.InitContext(c)
	p.ctx = pgsgo.InitContext(c.Parameters())

	tpl := template.New("Sanitize").Funcs(map[string]interface{}{
		"package":          p.ctx.PackageName,
		"name":             p.ctx.Name,
		"sanitizer":        p.sanitizer,
		"initializer":      p.initializer,
		"leadingCommenter": p.leadingCommenter,
	})

	p.tpl = template.Must(tpl.Parse(sanitizeTpl))
}

// Name satisfies the generator.Plugin interface.
func (p *SanitizeModule) Name() string { return "Sanitize" }

// Execute generates validation code for messages
func (p *SanitizeModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	p.Debug("Execute")

	for _, t := range targets {
		p.generateFile(t)
	}

	return p.Artifacts()
}

func (p *SanitizeModule) generateFile(f pgs.File) {
	if len(f.Messages()) == 0 {
		return
	}

	name := p.ctx.OutputPath(f).SetExt(".sanitize.go")

	p.Debug("generate:", name)

	p.AddGeneratorTemplateFile(name.String(), p.tpl, f)
}

func (p *SanitizeModule) leadingCommenter(f pgs.File) string {
	p.Debug("Comments:", f.SourceCodeInfo().LeadingDetachedComments())

	var comments []string
	re := regexp.MustCompile("(\r?\n)+")

	for _, comment := range f.SourceCodeInfo().LeadingDetachedComments() {
		tmpCmt := re.Split(comment, -1)
		comments = append(comments, tmpCmt[:len(tmpCmt)-1]...)
	}

	return "//" + strings.Join(comments, "\n//")

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

const sanitizeTpl = `{{ leadingCommenter . }}

package {{ package . }}
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
