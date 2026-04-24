package verbcli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	glazerunner "github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestNewCommandIncludesBuiltinVerbs(t *testing.T) {
	cmd, err := NewCommand(Bootstrap{Repositories: []Repository{builtinRepository()}})
	require.NoError(t, err)

	found, _, err := cmd.Find([]string{"script", "compare", "region"})
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "region", found.Name())

	found, _, err = cmd.Find([]string{"catalog", "inspect-page"})
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "inspect-page", found.Name())
}

func TestDuplicateVerbPathsReturnError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "duplicate.js"), `
function region() { return { ok: true }; }
__verb__("region", {
  parents: ["script", "compare"],
  fields: {}
});
`)

	_, err := NewCommand(Bootstrap{Repositories: []Repository{
		builtinRepository(),
		{Name: "custom", Source: "test", RootDir: dir},
	}})
	require.Error(t, err)
	require.Contains(t, err.Error(), `duplicate jsverb path "script compare region"`)
}

func TestLoadRepositoriesFromConfigFile(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "verbs")
	disabledDir := filepath.Join(dir, "disabled")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	require.NoError(t, os.MkdirAll(disabledDir, 0o755))
	configPath := filepath.Join(dir, "config.yaml")
	writeFile(t, configPath, `
verbs:
  repositories:
    - name: local
      path: ./verbs
    - name: disabled
      path: ./disabled
      enabled: false
`)

	repos, err := loadRepositoriesFromConfigFile(configPath)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	require.Equal(t, "local", repos[0].Name)
	require.Equal(t, "config", repos[0].Source)
	require.Equal(t, repoDir, repos[0].RootDir)
}

func TestFilesystemRepositoryVerbExecutes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hello.js"), `
function hello(name) { return "hello " + name; }
__verb__("hello", {
  parents: ["custom"],
  output: "text",
  fields: {
    name: { argument: true, required: true }
  }
});
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{
		"default": {"name": "Manuel"},
	}))
	require.NoError(t, err)

	writerCommand, ok := commands[0].(cmds.WriterCommand)
	require.True(t, ok)
	var out bytes.Buffer
	require.NoError(t, writerCommand.RunIntoWriter(context.Background(), parsedValues, &out))
	require.Contains(t, out.String(), "hello Manuel")
}

func TestRepositoryVerbUsesPromiseFirstCVDModule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" style="color: rgb(255, 0, 0)">Book</button></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "inspect.js"), `
async function inspect(url, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  try {
    const page = await browser.page(url, { viewport: { width: 400, height: 300 } });
    const probes = [{ name: "cta", selector: "#cta", props: ["color"] }];
    const statuses = await page.preflight(probes);
    const result = await page.inspectAll(probes, { outDir, artifacts: "css-json" });
    await page.close();
    return { exists: statuses[0].exists, count: result.results.length, outputDir: result.outputDir };
  } finally {
    await browser.close();
  }
}
__verb__("inspect", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true },
    outDir: { argument: true, required: true }
  }
});
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	outDir := t.TempDir()
	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{
		"default": {"url": server.URL, "outDir": outDir},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, true, row["exists"])
	require.EqualValues(t, 1, row["count"])
	require.Equal(t, outDir, row["outputDir"])
	_, err = os.Stat(filepath.Join(outDir, "computed-css.json"))
	require.NoError(t, err)
}

func TestRepositoryVerbWritesCatalogManifestAndIndex(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "catalog.js"), `
async function catalogSmoke(outDir) {
  const cvd = require("css-visual-diff");
  const catalog = cvd.catalog({ title: "Verb Catalog Smoke", outDir, artifactRoot: "../artifacts" });
  const target = { slug: "../Demo Target!", name: "Demo Target", url: "http://example.test", selector: "#root", viewport: { width: 320, height: 240 } };
  catalog.addTarget(target);
  catalog.recordPreflight(target, [{ name: "root", selector: "#root", exists: true, visible: true, textStart: "Ready" }]);
  catalog.addResult(target, { outputDir: catalog.artifactDir(target.slug), results: [{ metadata: { name: "root", selector: "#root", createdAt: "2026-04-24T00:00:00Z" }, style: { exists: true, computed: { color: "rgb(0, 0, 0)" } } }] });
  const manifestPath = await catalog.writeManifest();
  const indexPath = await catalog.writeIndex();
  const summary = catalog.summary();
  return { manifestPath, indexPath, targetCount: summary.targetCount, resultCount: summary.resultCount, artifactDir: catalog.artifactDir(target.slug) };
}
__verb__("catalogSmoke", {
  parents: ["custom"],
  fields: {
    outDir: { argument: true, required: true }
  }
});
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	outDir := t.TempDir()
	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{
		"default": {"outDir": outDir},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, filepath.Join(outDir, "manifest.json"), row["manifestPath"])
	require.Equal(t, filepath.Join(outDir, "index.md"), row["indexPath"])
	require.EqualValues(t, 1, row["targetCount"])
	require.EqualValues(t, 1, row["resultCount"])
	require.Equal(t, filepath.Join(outDir, "artifacts", "demo-target"), row["artifactDir"])
	manifestBytes, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	require.NoError(t, err)
	require.Contains(t, string(manifestBytes), `"schema_version": "css-visual-diff.catalog.v1"`)
	require.Contains(t, string(manifestBytes), `"slug": "demo-target"`)
	indexBytes, err := os.ReadFile(filepath.Join(outDir, "index.md"))
	require.NoError(t, err)
	require.Contains(t, string(indexBytes), "# Verb Catalog Smoke")
}

func TestCVDModuleExposesLowerCamelGotoInspectAndTypedErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><main id="app"><p>Ready</p></main></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "missing.js"), `
async function missing(url, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.newPage();
    const gotoResult = await page.goto(url, { viewport: { width: 320, height: 240 }, name: "typed-error-test" });
    await page.inspect({ name: "missing", selector: "#missing" }, { outDir, artifacts: "html" });
    return { ok: true, gotoUrl: gotoResult.url };
  } catch (err) {
    return {
      ok: false,
      name: err.name,
      code: err.code,
      operation: err.operation,
      isSelector: err instanceof cvd.SelectorError,
      isCvd: err instanceof cvd.CvdError
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("missing", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true },
    outDir: { argument: true, required: true }
  }
});
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{
		"default": {"url": server.URL, "outDir": t.TempDir()},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, false, row["ok"])
	require.Equal(t, "SelectorError", row["name"])
	require.Equal(t, "SELECTOR_ERROR", row["code"])
	require.Equal(t, "css-visual-diff.page.inspect", row["operation"])
	require.Equal(t, true, row["isSelector"])
	require.Equal(t, true, row["isCvd"])
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func rowToMap(row types.Row) map[string]interface{} {
	ret := map[string]interface{}{}
	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		ret[pair.Key] = pair.Value
	}
	return ret
}

type captureProcessor struct {
	rows []types.Row
}

func (c *captureProcessor) AddRow(_ context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}

func (c *captureProcessor) Close(context.Context) error {
	return nil
}
