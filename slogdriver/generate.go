package slogdriver

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	slogexpr "github.com/kitagry/goaplugin/slogdriver/expr"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type fileToModify struct {
	file        *codegen.File
	path        string
	serviceName string
	isMain      bool
}

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginFirst("slogdriver", "gen", nil, Generate)
	codegen.RegisterPluginLast("slogdriver-updater", "example", nil, UpdateExample)
}

// Generate generates slog logger specific files.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {
			files = append(files, GenerateFiles(genpkg, r)...)
		}
	}
	return files, nil
}

// UpdateExample modifies the example generated files by replacing
// the log import reference when needed
// It also modify the initially generated main and service files
func UpdateExample(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	filesToModify := []*fileToModify{}

	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {

			// Add the generated main files
			for _, svr := range r.API.Servers {
				pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
				filesToModify = append(filesToModify,
					&fileToModify{path: filepath.Join("cmd", pkg, "main.go"), serviceName: svr.Name, isMain: true})
				filesToModify = append(filesToModify,
					&fileToModify{path: filepath.Join("cmd", pkg, "http.go"), serviceName: svr.Name, isMain: true})
				filesToModify = append(filesToModify,
					&fileToModify{path: filepath.Join("cmd", pkg, "grpc.go"), serviceName: svr.Name, isMain: true})
			}

			// Add the generated service files
			for _, svc := range r.API.HTTP.Services {
				servicePath := codegen.SnakeCase(svc.Name()) + ".go"
				filesToModify = append(filesToModify, &fileToModify{path: servicePath, serviceName: svc.Name(), isMain: false})
			}

			// Update the added files
			for _, fileToModify := range filesToModify {
				for _, file := range files {
					if file.Path == fileToModify.path {
						fileToModify.file = file
						updateExampleFile(genpkg, r, fileToModify)
						break
					}
				}
			}
		}
	}

	return files, nil
}

// GenerateFiles create log specific files
func GenerateFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	fw := make([]*codegen.File, 1)
	fw[0] = GenerateLoggerFile(genpkg)
	return fw
}

// GenerateLoggerFile returns the generated slogdriver logger file.
func GenerateLoggerFile(genpkg string) *codegen.File {
	path := filepath.Join(codegen.Gendir, "log", "logger.go")
	title := fmt.Sprint("slogdriver logger implementation")
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "log", []*codegen.ImportSpec{
			{Path: "github.com/kitagry/slogdriver"},
			{Path: "log/slog"},
			{Path: "goa.design/goa/v3/http/middleware", Name: "httpmdlwr"},
			{Path: "goa.design/goa/v3/middleware"},
			{Path: "os"},
			{Path: "io"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "time"},
			{Path: "context"},
		}),
	}

	sections = append(sections, &codegen.SectionTemplate{
		Name:   "slogdriver",
		Source: loggerT,
	})

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func updateExampleFile(genpkg string, root *expr.RootExpr, f *fileToModify) {
	header := f.file.SectionTemplates[0]
	logPath := path.Join(genpkg, "log")

	data := header.Data.(map[string]any)
	specs := data["Imports"].([]*codegen.ImportSpec)

	for _, spec := range specs {
		if spec.Path == "log" {
			spec.Name = "log"
			spec.Path = logPath
		}
	}

	if f.isMain {

		codegen.AddImport(header, &codegen.ImportSpec{Path: "github.com/kitagry/slogdriver"})
		healthPaths := buildHealthCheckPaths()

		for _, s := range f.file.SectionTemplates {
			switch s.Name {
			case "server-main-logger":
				s.Source = strings.Replace(s.Source, `logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)`,
					`logger = log.New(os.Stderr, slogdriver.HandlerOptions{})`, 1)
			case "server-http-logger":
				s.Source = strings.Replace(s.Source, "adapter = middleware.NewLogger(logger)", "adapter = logger", 1)
			case "server-http-middleware":
				s.Source = strings.Replace(s.Source, "handler = httpmdlwr.Log(adapter)(handler)", fmt.Sprintf("handler = log.SlogdriverHttpMiddleware(adapter, []string{%s})(handler)", strings.Join(healthPaths, ", ")), 1)
				// RequestID is deprecated. And, slogdriver can set openTelemetry id.
				s.Source = strings.Replace(s.Source, "handler = httpmdlwr.RequestID()(handler)\n", ``, 1)
			case "server-http-errorhandler":
				s.Source = `// errorHandler returns a function that writes and logs the given error.
func errorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
        return func(ctx context.Context, w http.ResponseWriter, err error) {
                logger.ErrorContext(ctx, err.Error())
        }
}
`
			}
		}
	}
}

func buildHealthCheckPaths() []string {
	result := make([]string, 0)
	for _, hc := range slogexpr.Root.HealthChecks {
		result = append(result, hc.Paths...)
	}

	for i, r := range result {
		result[i] = fmt.Sprintf(`"%s"`, r)
	}
	return result
}

const loggerT = `
// Logger is an adapted slogdriver logger
type Logger struct {
	*slog.Logger
}
// New creates a new slogdriver logger
func New(w io.Writer, opts slogdriver.HandlerOptions) *Logger {
	logger := slogdriver.New(w, opts)
	return &Logger{logger}
}
// Log is called by the log middleware to log HTTP requests key values
func (logger *Logger) Log(keyvals ...any) error {
	logger.Logger.Info("HTTP Request", keyvals...)
	return nil
}
// Print is called by the log middleware
func (logger *Logger) Print(v ...any) {
	logger.Logger.Info(fmt.Sprint(v...))
}
// Printf is called by the log middleware
func (logger *Logger) Printf(msg string, args ...any) {
	logger.Logger.Info(fmt.Sprintf(msg, args...))
}
// Println is called by the log middleware
func (logger *Logger) Println(v ...any) {
	logger.Logger.Info(fmt.Sprint(v...))
}
// Fatal is called by the log middleware
func (logger *Logger) Fatal(v ...any) {
	logger.Logger.Log(context.Background(), slogdriver.LevelCritical, fmt.Sprint(v...))
	os.Exit(1)
}
// Fatalf is called by the log middleware
func (logger *Logger) Fatalf(msg string, args ...any) {
	logger.Logger.Log(context.Background(), slogdriver.LevelCritical, fmt.Sprintf(msg, args...))
	os.Exit(1)
}
// Fatalln is called by the log middleware
func (logger *Logger) Fataln(v ...any) {
	logger.Logger.Log(context.Background(), slogdriver.LevelCritical, fmt.Sprint(v...))
	os.Exit(1)
}

// responseCapture is a http.ResponseWriter which captures the response status
// code and content length.
type responseCapture struct {
	http.ResponseWriter
	StatusCode    int
	ContentLength int
}

// captureResponse creates a ResponseCapture that wraps the given ResponseWriter.
func captureResponse(w http.ResponseWriter) *responseCapture {
	return &responseCapture{ResponseWriter: w}
}

// WriteHeader records the value of the status code before writing it.
func (w *responseCapture) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write computes the written len and stores it in ContentLength.
func (w *responseCapture) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.ContentLength += n
	return n, err
}

// Flush implements the http.Flusher interface if the underlying response
// writer supports it.
func (w *responseCapture) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// FormatFields formats input keyvals
// ref: https://github.com/goadesign/goa/blob/v1/logging/logrus/adapter.go#L64
func FormatFields(keyvals []any) map[string]any {
	n := (len(keyvals) + 1) / 2
	res := make(map[string]any, n)
	for i := 0; i < len(keyvals); i += 2 {
		k := keyvals[i]
		var v any
		if i+1 < len(keyvals) {
			v = keyvals[i+1]
		}
		res[fmt.Sprintf("%v", k)] = v
	}
	return res
}
// SlogdriverHttpMiddleware extracts and formats http request and response information into
// GCP Cloud Logging optimized format.
// If logger is not *Logger, it returns goa default middleware.
// healthCheckPaths is used to skip log when the request is correct.
func SlogdriverHttpMiddleware(logger middleware.Logger, healthCheckPaths []string) func(h http.Handler) http.Handler {
	switch logr := logger.(type) {
	case *Logger:
		return slogdriverHttpMiddleware(logr, healthCheckPaths)
	default:
		return httpmdlwr.Log(logger)
	}
}

func slogdriverHttpMiddleware(logger *Logger, healthCheckPaths []string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := captureResponse(w)
			h.ServeHTTP(rw, r)

			var res http.Response
			res.StatusCode = rw.StatusCode
			res.ContentLength = int64(rw.ContentLength)

			p := slogdriver.MakeHTTPPayload(r, &res)
			p.Latency = time.Since(start).String()

			var level slog.Level
			switch {
			case rw.StatusCode < 400:
				level = slogdriver.LevelInfo
			case rw.StatusCode < 500:
				level = slogdriver.LevelWarning
			default:
				level = slogdriver.LevelError
			}

			if isHealthCheckPath(r.URL.Path, healthCheckPaths) && rw.StatusCode < 400 {
				return
			}

			logger.Logger.Log(r.Context(), level, "request finished", slogdriver.HTTPKey, p)
		})
	}
}

func isHealthCheckPath(path string, healthCheckPaths []string) bool {
	for _, hp := range healthCheckPaths {
		if path == hp {
			return true
		}
	}
	return false
}
`
