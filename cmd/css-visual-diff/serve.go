package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/css-visual-diff/internal/cssvisualdiff/review"
	"github.com/spf13/cobra"
)

type serveSettings struct {
	dataDir string
	summary string
	port    int
	host    string
	open    bool
}

func newServeCommand() *cobra.Command {
	s := &serveSettings{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the interactive visual review site for css-visual-diff comparison results",
		Long: `Start an HTTP server that serves the React-based visual review site.
The review site loads comparison results from the specified data directory
and presents them as interactive review cards with local storage for feedback.

The data directory should contain the output of a css-visual-diff run
(with page/section subdirectories containing compare.json and PNG artifacts).`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd, s)
		},
	}
	cmd.Flags().StringVar(&s.dataDir, "data-dir", "", "Path to the css-visual-diff run output directory (required)")
	cmd.Flags().StringVar(&s.summary, "summary", "", "Explicit path to summary JSON (default: <data-dir>/summary.json)")
	cmd.Flags().IntVar(&s.port, "port", 8097, "HTTP server port")
	cmd.Flags().StringVar(&s.host, "host", "127.0.0.1", "Bind address")
	cmd.Flags().BoolVar(&s.open, "open", false, "Open browser automatically")
	_ = cmd.MarkFlagRequired("data-dir")
	return cmd
}

func runServe(cmd *cobra.Command, s *serveSettings) error {
	// Verify data directory exists
	info, err := os.Stat(s.dataDir)
	if err != nil {
		return fmt.Errorf("data-dir %s: %w", s.dataDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("data-dir %s is not a directory", s.dataDir)
	}

	// Resolve summary path
	summaryPath := s.summary
	if summaryPath == "" {
		summaryPath = filepath.Join(s.dataDir, "summary.json")
	}

	mux := http.NewServeMux()

	// API: manifest (returns summary JSON)
	mux.HandleFunc("GET /api/manifest", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(summaryPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("summary not found: %v", err), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(data)
	})

	// API: compare.json for a specific page/section
	mux.HandleFunc("GET /api/compare", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		section := r.URL.Query().Get("section")
		if page == "" || section == "" {
			http.Error(w, "page and section query parameters required", http.StatusBadRequest)
			return
		}
		comparePath := filepath.Join(s.dataDir, page, "artifacts", section, "compare.json")
		data, err := os.ReadFile(comparePath)
		if err != nil {
			http.Error(w, "compare.json not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(data)
	})

	// Artifact serving: /artifacts/<page>/<section>/<file>
	// React app converts: /tmp/.../shows/artifacts/content/diff_only.png -> /artifacts/shows/content/diff_only.png
	// Go handler maps back: shows/content/diff_only.png -> shows/artifacts/content/diff_only.png
	mux.HandleFunc("GET /artifacts/{path...}", func(w http.ResponseWriter, r *http.Request) {
		artifactPath := r.PathValue("path")
		// artifactPath = "shows/content/diff_only.png"
		// Full path on disk = data-dir/shows/artifacts/content/diff_only.png
		parts := strings.SplitN(artifactPath, "/", 3)
		if len(parts) < 3 {
			http.Error(w, "invalid artifact path", http.StatusBadRequest)
			return
		}
		page, section, file := parts[0], parts[1], parts[2]
		fullPath := filepath.Join(s.dataDir, page, "artifacts", section, file)
		if _, err := os.Stat(fullPath); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, fullPath)
	})

	// SPA: serve embedded React app
	spaHandler, err := review.NewSPAHandler(&review.SPAOptions{
		APIPrefix: "/api",
	})
	if err != nil {
		// If no embedded SPA, serve a simple message
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<h1>css-visual-diff review</h1><p>No embedded SPA found. Run <code>go generate ./internal/cssvisualdiff/review</code> first.</p>")
		})
	} else {
		mux.Handle("/", spaHandler)
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	fmt.Fprintf(cmd.OutOrStdout(), "Serving review site at http://%s\n", addr)
	fmt.Fprintf(cmd.OutOrStdout(), "Data dir: %s\n", s.dataDir)
	fmt.Fprintf(cmd.OutOrStdout(), "Summary:  %s\n", summaryPath)

	if s.open {
		fmt.Fprintf(cmd.OutOrStdout(), "Opening browser...\n")
		_ = openBrowser(fmt.Sprintf("http://%s", addr))
	}

	return http.ListenAndServe(addr, mux)
}

func openBrowser(url string) error {
	return exec.Command("xdg-open", url).Start()
}
