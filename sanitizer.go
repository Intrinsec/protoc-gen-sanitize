package main

import (
	"fmt"
	"os"
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
	ctx              pgsgo.Context
	tpl              *template.Template
	importBluemonday map[pgs.File]bool
	enforce          bool
	hasMissing       bool
}

// Sanitize returns an initialized SanitizePlugin
func Sanitize() *SanitizeModule {
	return &SanitizeModule{
		ModuleBase:       &pgs.ModuleBase{},
		importBluemonday: make(map[pgs.File]bool),
		hasMissing:       false,
	}
}

// InitContext populates the module with needed context and fields
func (p *SanitizeModule) InitContext(c pgs.BuildContext) {
	c.Debug("InitContext")
	p.ModuleBase.InitContext(c)
	p.ctx = pgsgo.InitContext(c.Parameters())

	tpl := template.New("Sanitize").Funcs(map[string]interface{}{
		"package":            p.ctx.PackageName,
		"name":               p.ctx.Name,
		"sanitizer":          p.sanitizer,
		"initializer":        p.initializer,
		"leadingCommenter":   p.leadingCommenter,
		"doImportBluemonday": p.doImportBluemonday,
	})

	p.tpl = template.Must(tpl.Parse(sanitizeTpl))
}

// Name satisfies the generator.Plugin interface.
func (p *SanitizeModule) Name() string { return "Sanitize" }

// Execute generates sanitization code for files
func (p *SanitizeModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	p.Debug("Execute")

	if ok, _ := p.Parameters().Bool("enforce"); ok {
		p.enforce = true
	}

	for _, t := range targets {
		if !p.doSanitize(t) {
			continue
		}
		p.generateFile(t)
	}

	return p.Artifacts()
}

func (p *SanitizeModule) ExitCheck() {

	if p.hasMissing && p.enforce {
		p.Log("Sanitization is enforced. Add an option or explicitely disable the sanitization on the above fields")
		os.Exit(1)
	}
}

func (p *SanitizeModule) doSanitize(f pgs.File) bool {
	var disableFile bool

	p.importBluemonday[f] = false

	if ok, err := f.Extension(sanitize.E_DisableFile, &disableFile); ok && err == nil && disableFile {
		p.Debug("Skipping: ", f.InputPath())
		return false
	}

	for _, m := range f.AllMessages() {
		var disableMessage bool
		if ok, err := m.Extension(sanitize.E_DisableMessage, &disableMessage); ok && err == nil && disableMessage {
			p.Debug("Skipping: ", m.Name())
			continue
		}

		for _, field := range m.Fields() {
			var kind sanitize.Sanitization
			if ok, err := field.Extension(sanitize.E_Kind, &kind); ok && err == nil {
				// Only case where we will use bluemonday in the generated code
				p.importBluemonday[f] = true
				return true
			}
		}
	}

	p.Debug("No sanitization options encountered for:", f.InputPath())
	// Nonetheless we generate sanitization function to enable calling sanitization function of nested messages
	return true
}

func (p *SanitizeModule) doImportBluemonday(f pgs.File) bool {
	p.Debug("doImportBluemonday")
	if value, ok := p.importBluemonday[f]; ok {
		p.Debug(value)
		return value
	}
	return false
}

func (p *SanitizeModule) generateFile(f pgs.File) {
	if len(f.Messages()) == 0 {
		return
	}
	p.Push(f.Name().String())
	defer p.Pop()
	p.Debug("File:", f.InputPath())

	name := f.InputPath().BaseName() + ".pb.sanitize.go"

	p.Debug("generate:", name)

	p.AddGeneratorTemplateFile(name, p.tpl, f)
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

func (p *SanitizeModule) buildSanitizeCall(f pgs.Field, name string, sanitizeKind string) string {
	prefix := ""
	suffix := ""
	indent := ""
	sanitizeCall := ""
	iter := "i"
	elementName := name

	if f.Type().IsRepeated() {
		p.Debug("Repeated:", name)
		indent = "	"
		suffix = "\n}"
		elementName = strings.ToLower(string(name[0]))
		if sanitizeKind == "" {
			iter = "_"
		}
		prefix = fmt.Sprintf("for %s, %s := range m.%s {\n",
			iter,
			elementName,
			name,
		)
	}
	var format string
	if sanitizeKind == "" {
		// building call for message
		if f.Type().IsRepeated() {
			format = "%s.Sanitize()"
		} else {
			format = "m.%s.Sanitize()"
		}
		sanitizeCall = fmt.Sprintf(format, elementName)
	} else {
		// building call for string
		if f.Type().IsRepeated() {
			format = "m.%s[i] = %sSanitize.Sanitize(%s)"
		} else {
			format = "m.%s = %sSanitize.Sanitize(m.%s)"
		}
		sanitizeCall = fmt.Sprintf(format, name, strings.ToLower(sanitizeKind), elementName)
	}
	return fmt.Sprintf("%s%s%s%s", prefix, indent, sanitizeCall, suffix)
}

func (p *SanitizeModule) sanitizer(f pgs.Field) string {
	name := p.ctx.Name(f)

	switch f.Type().ProtoType() {
	case pgs.StringT:
		var kind sanitize.Sanitization

		ok, err := f.Extension(sanitize.E_Kind, &kind)
		if err == nil {
			if ok {
				switch kind {
				case sanitize.Sanitization_NONE:
					return ""
				case sanitize.Sanitization_HTML:
					return p.buildSanitizeCall(f, string(name), "html")
				case sanitize.Sanitization_TEXT:
					return p.buildSanitizeCall(f, string(name), "text")
				}
			} else {
				if f.Type().ProtoType() == pgs.StringT {
					fmt.Fprintf(
						os.Stderr,
						"%v:%d: no sanitize option on %v\n",
						f.File().Name(),
						f.SourceCodeInfo().Location().Span[0]+1,
						f.FullyQualifiedName())
					p.hasMissing = true
				}
			}
		}

	case pgs.MessageT:
		var disableField bool

		if ok, err := f.Extension(sanitize.E_DisableField, &disableField); ok && err == nil && disableField {
			p.Debug("Skipping field:", name)
			return ""
		}
		return p.buildSanitizeCall(f, string(name), "")
	}
	return ""
}

const sanitizeTpl = `{{ leadingCommenter . }}

package {{ package . }}
{{ if doImportBluemonday . }}
import (
	"github.com/microcosm-cc/bluemonday"
)
{{ end }}

{{ range .AllMessages }}
func (m *{{ name . }}) Sanitize() {
	if m == nil {
		return
	}

	{{ initializer . }}

{{ range .Fields }}
    {{ sanitizer . }}
{{ end }}
}
{{ end }}
`
