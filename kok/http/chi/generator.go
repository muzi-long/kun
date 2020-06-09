package chi

import (
	"fmt"
	"strings"

	"github.com/RussellLuo/kok/kok/endpoint"
	"github.com/RussellLuo/kok/pkg/gen"
	"github.com/RussellLuo/kok/pkg/openapi"
	"github.com/RussellLuo/kok/pkg/reflector"
)

var (
	template = `// Code generated by kok; DO NOT EDIT.
// github.com/RussellLuo/kok

{{- $pkgName := .Result.PkgName}}
{{- $enableTracing := .Opts.EnableTracing}}

package {{$pkgName}}

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/go-chi/chi"
	{{- if $enableTracing}}
	"github.com/RussellLuo/kok/pkg/tracing/xnet"
	{{- end}}
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/kit/endpoint"

	{{- range .Result.Imports}}
	"{{.}}"
	{{- end}}
)


func NewHTTPHandler(svc {{.Result.SrcPkgPrefix}}{{.Result.Interface.Name}}
                    {{- if $enableTracing}}, newTracer xnet.NewFunc{{end}}) http.Handler {
	r := chi.NewRouter()

	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(errorEncoder),
	}
	{{range .Spec.Operations}}
	r.Method(
		"{{.Method}}", "{{.Pattern}}",
		kithttp.NewServer(
			MakeEndpointOf{{.Name}}(svc),
			decode{{.Name}}Request,
			encodeGenericResponse,

			{{- if $enableTracing}}
			append(options, kithttp.ServerBefore(xnet.HTTPToContext(newTracer, "{{$pkgName}}", "{{.Name}}")))...,
			{{- else}}
			options...,
			{{- end}}
		),
	)
	{{- end}}

	return r
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// err2code (signature: func(error) int) must be provided in the same
	// package, to transform any business error to an HTTP code!
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

type errorWrapper struct {
	Error string ` + "`" + `json:"error"` + "`" + `
}

{{- range .Spec.Operations}}

{{- $nonCtxParams := nonCtxParams .Request.Params}}

func decode{{.Name}}Request(_ context.Context, r *http.Request) (interface{}, error) {
	{{$nonBodyParams := nonBodyParams $nonCtxParams -}}
	{{range $nonBodyParams -}}

	{{- if eq .Type "string" -}}
	{{.Name}} := {{extractParam .}}
	{{- else -}}
	{{.Name}}Value := {{extractParam .}}
	{{.Name}}, err := {{parseExpr .Name .Type}}
	if err != nil {
		return nil, err
	}
	{{end}}

	{{end -}}

	{{- $bodyParams := bodyParams $nonCtxParams}}
	{{- if $bodyParams -}}
	var body struct {
		{{- range $bodyParams}}
		{{title .Name}} {{.Type}} {{addTag .Name .Type}}
		{{- end}}
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	{{- end}}

	return {{addAmpersand .Name}}Request{
		{{- range $nonCtxParams}}

		{{- if eq .In "body"}}
		{{title .Name}}: body.{{title .Name}},
		{{- else}}
		{{title .Name}}: {{castIfInt .Name .Type}},
		{{- end}}

		{{- end}}
	}, nil
}

{{- end}}

func encodeGenericResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		return f.Failed()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
`
)

type RequestField struct {
	Name  string
	Value string
}

type Server struct {
	Service     interface{}
	NewEndpoint interface{}
	Request     interface{}
	Response    interface{}
}

type Options struct {
	SchemaPtr         bool
	SchemaTag         string
	TagKeyToSnakeCase bool
	Formatted         bool
	EnableTracing     bool
}

type Generator struct {
	opts *Options
}

func New(opts *Options) *Generator {
	return &Generator{opts: opts}
}

func (g *Generator) Generate(result *reflector.Result, spec *openapi.Specification) ([]byte, error) {
	data := struct {
		Result *reflector.Result
		Spec   *openapi.Specification
		Opts   *Options
	}{
		Result: result,
		Spec:   spec,
		Opts:   g.opts,
	}

	return gen.Generate(template, data, gen.Options{
		Funcs: map[string]interface{}{
			"title": strings.Title,
			"addTag": func(name, typ string) string {
				if g.opts.SchemaTag == "" {
					return ""
				}

				if typ == "error" {
					name = "-"
				} else if g.opts.TagKeyToSnakeCase {
					name = endpoint.ToSnakeCase(name)
				}

				return fmt.Sprintf("`%s:\"%s\"`", g.opts.SchemaTag, name)
			},
			"addAmpersand": func(name string) string {
				if g.opts.SchemaPtr {
					return "&" + name
				}
				return name
			},
			"extractParam": func(param *openapi.Param) string {
				switch param.In {
				case openapi.InPath:
					return fmt.Sprintf(`chi.URLParam(r, "%s")`, param.Name)
				case openapi.InQuery:
					return fmt.Sprintf(`r.URL.Query().Get("%s")`, param.Name)
				default:
					panic(fmt.Errorf("param.In `%s` not supported", param.In))
				}
			},
			"nonBodyParams": func(in []*openapi.Param) (out []*openapi.Param) {
				for _, p := range in {
					if p.In != openapi.InBody {
						out = append(out, p)
					}
				}
				return
			},
			"bodyParams": func(in []*openapi.Param) (out []*openapi.Param) {
				for _, p := range in {
					if p.In == openapi.InBody {
						out = append(out, p)
					}
				}
				return
			},
			"nonCtxParams": func(params []*openapi.Param) (out []*openapi.Param) {
				for _, p := range params {
					if p.Type != "context.Context" {
						out = append(out, p)
					}
				}
				return
			},
			"parseExpr": func(name, typ string) string {
				switch typ {
				case "int", "int8", "int16", "int32", "int64":
					return fmt.Sprintf("strconv.ParseInt(%sValue, 10, 64)", name)
				case "uint", "uint8", "uint16", "uint32", "uint64":
					return fmt.Sprintf("strconv.ParseUint(%sValue, 10, 64)", name)
				default:
					panic(fmt.Errorf("unrecognized integer type %s", typ))
				}
			},
			"castIfInt": func(name, typ string) string {
				switch typ {
				case "int", "int8", "int16", "int32",
					"uint", "uint8", "uint16", "uint32":
					return fmt.Sprintf("%s(%s)", typ, name)
				default:
					return name
				}
			},
		},
		Formatted: g.opts.Formatted,
	})
}
