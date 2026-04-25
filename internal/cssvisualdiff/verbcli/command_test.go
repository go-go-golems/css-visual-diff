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

	found, _, err = cmd.Find([]string{"catalog", "inspect-config"})
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "inspect-config", found.Name())
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
  const catalog = cvd.catalog.create({ title: "Verb Catalog Smoke", outDir, artifactRoot: "../artifacts" });
  const target = { slug: "../Demo Target!", name: "Demo Target", url: "http://example.test", selector: "#root", viewport: { width: 320, height: 240 } };
  catalog.addTarget(target);
  catalog.recordPreflight(target, [{ name: "root", selector: "#root", exists: true, visible: true, textStart: "Ready" }]);
  catalog.addResult(target, { outputDir: catalog.artifactDir(target.slug), results: [{ metadata: { name: "root", selector: "#root", createdAt: "2026-04-24T00:00:00Z" }, style: { exists: true, computed: { color: "rgb(0, 0, 0)" } } }] });
  const inMemoryManifest = catalog.manifest();
  const manifestPath = await catalog.writeManifest();
  const indexPath = await catalog.writeIndex();
  const summary = catalog.summary();
  return { manifestPath, indexPath, targetCount: summary.targetCount, resultCount: summary.resultCount, preflightRecordCount: inMemoryManifest.preflights.length, manifestResultCount: inMemoryManifest.results.length, failureRecordCount: inMemoryManifest.failures.length, artifactDir: catalog.artifactDir(target.slug) };
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
	require.EqualValues(t, 1, row["preflightRecordCount"])
	require.EqualValues(t, 1, row["manifestResultCount"])
	require.EqualValues(t, 0, row["failureRecordCount"])
	require.Equal(t, filepath.Join(outDir, "artifacts", "demo-target"), row["artifactDir"])
	manifestBytes, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	require.NoError(t, err)
	require.Contains(t, string(manifestBytes), `"schema_version": "css-visual-diff.catalog.v1"`)
	require.Contains(t, string(manifestBytes), `"slug": "demo-target"`)
	indexBytes, err := os.ReadFile(filepath.Join(outDir, "index.md"))
	require.NoError(t, err)
	require.Contains(t, string(indexBytes), "# Verb Catalog Smoke")
}

func TestCVDModuleDiffReportAndWritePrimitives(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "diff.js"), `
async function diffSmoke(outDir) {
  const cvd = require("css-visual-diff");
  const before = { results: [{ name: "cta", snapshot: { text: "Book", computed: { color: "red" } } }] };
  const after = { results: [{ name: "cta", snapshot: { text: "Book now", computed: { color: "red" } } }] };
  const diff = cvd.diff.structural(before, after);
  const markdown = cvd.report(diff).markdown();
  const jsonPath = outDir + "/diff.json";
  const markdownPath = outDir + "/diff.md";
  await cvd.write.json(jsonPath, diff);
  await cvd.report(diff).writeMarkdown(markdownPath);
  const ignored = cvd.diff.structural(before, after, { ignorePaths: ["results[0].snapshot.text"] });
  return {
    equal: diff.equal,
    changeCount: diff.changeCount,
    firstPath: diff.changes[0].path,
    markdownHasPath: markdown.includes("results[0].snapshot.text"),
    ignoredEqual: ignored.equal,
    jsonPath,
    markdownPath
  };
}
__verb__("diffSmoke", {
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
	require.Equal(t, false, row["equal"])
	require.EqualValues(t, 1, row["changeCount"])
	require.Equal(t, "results[0].snapshot.text", row["firstPath"])
	require.Equal(t, true, row["markdownHasPath"])
	require.Equal(t, true, row["ignoredEqual"])
	require.Equal(t, filepath.Join(outDir, "diff.json"), row["jsonPath"])
	require.Equal(t, filepath.Join(outDir, "diff.md"), row["markdownPath"])
	_, err = os.Stat(filepath.Join(outDir, "diff.json"))
	require.NoError(t, err)
	markdownBytes, err := os.ReadFile(filepath.Join(outDir, "diff.md"))
	require.NoError(t, err)
	require.Contains(t, string(markdownBytes), "# Snapshot Diff")
}

func TestCVDModuleSnapshotsPageWithProbeBuilders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" style="color: rgb(255, 0, 0)">Book now</button><p id="copy">Hello</p></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "snapshot.js"), `
async function snapshotSmoke(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    const snapshot = await cvd.snapshot.page(page, [
      cvd.probe("cta").selector("#cta").required().text().styles(["color"]),
      cvd.probe("copy").selector("#copy").text()
    ]);
    return {
      count: snapshot.results.length,
      firstName: snapshot.results[0].name,
      firstText: snapshot.results[0].snapshot.text,
      firstColor: snapshot.results[0].snapshot.computed.color,
      secondText: snapshot.results[1].snapshot.text
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("snapshotSmoke", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.EqualValues(t, 2, row["count"])
	require.Equal(t, "cta", row["firstName"])
	require.Equal(t, "Book now", row["firstText"])
	require.Equal(t, "rgb(255, 0, 0)", row["firstColor"])
	require.Equal(t, "Hello", row["secondText"])
}

func TestCVDModuleSnapshotRejectsRawObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta">Book</button></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "snapshot-error.js"), `
async function snapshotError(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    try {
      await cvd.snapshot.page(page, [{ name: "cta", selector: "#cta" }]);
      return { ok: true };
    } catch (err) {
      return { ok: false, name: err.name, message: err.message };
    }
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("snapshotError", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, false, row["ok"])
	require.Equal(t, "TypeError", row["name"])
	require.Contains(t, row["message"], "css-visual-diff.snapshot: expected array of cvd.probe() builders")
}

func TestCVDModuleExtractsFromLocatorWithExtractorHandles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" style="color: rgb(255, 0, 0)">  Book
now  </button></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "extract.js"), `
async function extractSmoke(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    const snapshot = await cvd.extract(page.locator("#cta"), [
      cvd.extractors.exists(),
      cvd.extractors.visible(),
      cvd.extractors.text(),
      cvd.extractors.bounds(),
      cvd.extractors.computedStyle(["color"]),
      cvd.extractors.attributes(["id", "class"])
    ]);
    return {
      selector: snapshot.selector,
      exists: snapshot.exists,
      visible: snapshot.visible,
      text: snapshot.text,
      widthPositive: snapshot.bounds.width > 0,
      color: snapshot.computed.color,
      attrClass: snapshot.attributes.class
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("extractSmoke", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, "#cta", row["selector"])
	require.Equal(t, true, row["exists"])
	require.Equal(t, true, row["visible"])
	require.Equal(t, "Book now", row["text"])
	require.Equal(t, true, row["widthPositive"])
	require.Equal(t, "rgb(255, 0, 0)", row["color"])
	require.Equal(t, "primary", row["attrClass"])
}

func TestCVDModuleExtractRejectsRawObjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta">Book</button></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "extract-error.js"), `
async function extractError(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    try {
      await cvd.extract({ selector: "#cta" }, [cvd.extractors.text()]);
      return { ok: true };
    } catch (err) {
      return { ok: false, name: err.name, message: err.message };
    }
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("extractError", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, false, row["ok"])
	require.Equal(t, "TypeError", row["name"])
	require.Contains(t, row["message"], "css-visual-diff.extract: expected cvd.locator")
}

func TestCVDModuleExposesTargetProbeAndExtractorBuilders(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "builders.js"), `
async function buildersSmoke() {
  const cvd = require("css-visual-diff");
  const target = cvd.target("booking")
    .url("http://example.test/booking")
    .viewport(cvd.viewport.mobile())
    .waitMs(25)
    .root("#app")
    .build();
  const probe = cvd.probe("cta")
    .selector("#cta")
    .required()
    .text()
    .bounds()
    .styles(["color"])
    .attributes(["id"])
    .build();
  const extractor = cvd.extractors.attributes(["id", "class"]).build();
  return {
    targetName: target.name,
    targetWidth: target.viewport.width,
    waitMs: target.waitMs,
    probeSelector: probe.selector,
    propCount: probe.props.length,
    extractorKind: extractor.kind,
    extractorAttrCount: extractor.attributes.length
  };
}
__verb__("buildersSmoke", {
  parents: ["custom"],
  fields: {}
});
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	parsedValues, err := glazerunner.ParseCommandValues(commands[0])
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, "booking", row["targetName"])
	require.EqualValues(t, 390, row["targetWidth"])
	require.EqualValues(t, 25, row["waitMs"])
	require.Equal(t, "#cta", row["probeSelector"])
	require.EqualValues(t, 1, row["propCount"])
	require.Equal(t, "attributes", row["extractorKind"])
	require.EqualValues(t, 2, row["extractorAttrCount"])
}

func TestCVDModuleExposesLocatorMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" data-kind="booking" style="color: rgb(255, 0, 0)">  Book
now  </button><div id="hidden" style="display:none">Hidden</div></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "locator.js"), `
async function locatorSmoke(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    const cta = page.locator("#cta");
    const [status, exists, visible, waited, pageWaited, text, bounds, styles, attrs, missingExists, hiddenVisible] = await Promise.all([
      cta.status(),
      cta.exists(),
      cta.visible(),
      cta.waitFor({ timeoutMs: 1000 }),
      page.waitForSelector("#cta", { timeoutMs: 1000 }),
      cta.text({ normalizeWhitespace: true, trim: true }),
      cta.bounds(),
      cta.computedStyle(["color", "display"]),
      cta.attributes(["id", "class", "data-kind", "missing"]),
      page.locator("#missing").exists(),
      page.locator("#hidden").visible()
    ]);
    return {
      statusExists: status.exists,
      exists,
      visible,
      waitedExists: waited.exists,
      waitedSelector: waited.selector,
      pageWaitedExists: pageWaited.exists,
      text,
      widthPositive: bounds.width > 0,
      color: styles.color,
      attrClass: attrs.class,
      attrMissing: attrs.missing,
      missingExists,
      hiddenVisible
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("locatorSmoke", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, true, row["statusExists"])
	require.Equal(t, true, row["exists"])
	require.Equal(t, true, row["visible"])
	require.Equal(t, true, row["waitedExists"])
	require.Equal(t, "#cta", row["waitedSelector"])
	require.Equal(t, true, row["pageWaitedExists"])
	require.Equal(t, "Book now", row["text"])
	require.Equal(t, true, row["widthPositive"])
	require.Equal(t, "rgb(255, 0, 0)", row["color"])
	require.Equal(t, "primary", row["attrClass"])
	require.Equal(t, "", row["attrMissing"])
	require.Equal(t, false, row["missingExists"])
	require.Equal(t, false, row["hiddenVisible"])
}

func TestCVDModuleLocatorWrongParentError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta">Book</button></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "locator-error.js"), `
async function locatorError(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    try {
      page.locator("#cta").styles(["color"]);
      return { ok: true };
    } catch (err) {
      return { ok: false, name: err.name, message: err.message };
    }
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("locatorError", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, false, row["ok"])
	require.Equal(t, "TypeError", row["name"])
	require.Contains(t, row["message"], ".styles() is not available here")
	require.Contains(t, row["message"], "belongs to cvd.probe")
	require.Contains(t, row["message"], "computedStyle")
}

func TestCVDModuleSerializesSamePagePromiseAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><main id="app"><p id="a">A</p><p id="b">B</p></main></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "same-page-concurrent.js"), `
async function samePageConcurrent(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    const [first, second, third] = await Promise.all([
      page.preflight([{ name: "a", selector: "#a" }]),
      page.preflight([{ name: "b", selector: "#b" }]),
      page.preflight([{ name: "app", selector: "#app" }])
    ]);
    return {
      firstExists: first[0].exists,
      secondExists: second[0].exists,
      thirdExists: third[0].exists
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("samePageConcurrent", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, true, row["firstExists"])
	require.Equal(t, true, row["secondExists"])
	require.Equal(t, true, row["thirdExists"])
}

func TestCVDModuleAllowsConcurrentOperationsOnSeparatePages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><main id="app"><p id="a">A</p><p id="b">B</p></main></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "separate-pages-concurrent.js"), `
async function separatePagesConcurrent(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let left;
  let right;
  try {
    [left, right] = await Promise.all([
      browser.page(url, { viewport: { width: 320, height: 240 }, name: "left" }),
      browser.page(url, { viewport: { width: 480, height: 320 }, name: "right" })
    ]);
    const [leftStatus, rightStatus] = await Promise.all([
      left.preflight([{ name: "a", selector: "#a" }]),
      right.preflight([{ name: "b", selector: "#b" }])
    ]);
    return {
      leftExists: leftStatus[0].exists,
      rightExists: rightStatus[0].exists
    };
  } finally {
    if (left) await left.close();
    if (right) await right.close();
    await browser.close();
  }
}
__verb__("separatePagesConcurrent", {
  parents: ["custom"],
  fields: {
    url: { argument: true, required: true }
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
		"default": {"url": server.URL},
	}))
	require.NoError(t, err)

	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, true, row["leftExists"])
	require.Equal(t, true, row["rightExists"])
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

func TestCVDModuleCollectsLocatorSelection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" style="color: rgb(255, 0, 0)">  Book
now  </button></body></html>`)
	}))
	defer server.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "collect.js"), `
async function collectSmoke(url) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(url, { viewport: { width: 320, height: 240 } });
    const selected = await page.locator("#cta").collect({ inspect: "rich" });
    const viaNamespace = await cvd.collect.selection(page.locator("#cta"), { inspect: "minimal" });
    const json = selected.toJSON();
    return {
      schemaVersion: json.schemaVersion,
      summaryExists: selected.summary().exists,
      text: selected.text(),
      color: selected.styles(["color"]).color,
      attrClass: selected.attributes(["class"]).class,
      minimalExists: viaNamespace.summary().exists,
      minimalTextEmpty: viaNamespace.text() === ""
    };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}
__verb__("collectSmoke", { parents: ["custom"], fields: { url: { argument: true, required: true } } });
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{"default": {"url": server.URL}}))
	require.NoError(t, err)
	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, "cssvd.collectedSelection.v1", row["schemaVersion"])
	require.Equal(t, true, row["summaryExists"])
	require.Equal(t, "Book now", row["text"])
	require.Equal(t, "rgb(255, 0, 0)", row["color"])
	require.Equal(t, "primary", row["attrClass"])
	require.Equal(t, true, row["minimalExists"])
	require.Equal(t, true, row["minimalTextEmpty"])
}

func TestCVDModuleComparesCollectedSelections(t *testing.T) {
	left := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" style="color: rgb(0, 0, 0); font-size: 16px">Book now</button></body></html>`)
	}))
	defer left.Close()
	right := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="secondary" style="color: rgb(255, 0, 0); font-size: 18px">Book now</button></body></html>`)
	}))
	defer right.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "compare-selections.js"), `
async function compareSelectionsSmoke(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let leftPage, rightPage;
  try {
    leftPage = await browser.page(leftUrl, { viewport: { width: 320, height: 240 } });
    rightPage = await browser.page(rightUrl, { viewport: { width: 320, height: 240 } });
    const left = await leftPage.locator("#cta").collect({ inspect: "rich" });
    const right = await rightPage.locator("#cta").collect({ inspect: "rich" });
    const comparison = await cvd.compare.selections(left, right, {
      styleProps: ["color", "font-size"],
      attributes: ["class"]
    });
    const writeResult = await comparison.artifacts.write(outDir, ["json", "markdown"]);
    return {
      schemaVersion: comparison.toJSON().schemaVersion,
      boundsChanged: comparison.bounds.diff().changed,
      styleCount: comparison.styles.diff().length,
      firstStyle: comparison.styles.diff(["color"])[0].name,
      attrName: comparison.attributes.diff(["class"])[0].name,
      markdownHasTitle: comparison.report.markdown().includes("Selection Comparison"),
      writtenCount: writeResult.written.length
    };
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}
__verb__("compareSelectionsSmoke", { parents: ["custom"], fields: { leftUrl: { argument: true, required: true }, rightUrl: { argument: true, required: true }, outDir: { argument: true, required: true } } });
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	outDir := t.TempDir()
	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{"default": {"leftUrl": left.URL, "rightUrl": right.URL, "outDir": outDir}}))
	require.NoError(t, err)
	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, "cssvd.selectionComparison.v1", row["schemaVersion"])
	require.Equal(t, true, row["boundsChanged"])
	require.EqualValues(t, 2, row["styleCount"])
	require.Equal(t, "color", row["firstStyle"])
	require.Equal(t, "class", row["attrName"])
	require.Equal(t, true, row["markdownHasTitle"])
	require.EqualValues(t, 2, row["writtenCount"])
	_, err = os.Stat(filepath.Join(outDir, "compare.json"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(outDir, "compare.md"))
	require.NoError(t, err)
}

func TestCVDModuleCompareRegionLowEffortAPI(t *testing.T) {
	left := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" style="color: rgb(0, 0, 0); font-size: 16px; padding: 8px">Book now</button></body></html>`)
	}))
	defer left.Close()
	right := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="secondary" style="color: rgb(255, 0, 0); font-size: 18px; padding: 12px">Book now</button></body></html>`)
	}))
	defer right.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "compare-region.js"), `
async function compareRegionSmoke(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let leftPage, rightPage;
  try {
    leftPage = await browser.page(leftUrl, { viewport: { width: 320, height: 240 } });
    rightPage = await browser.page(rightUrl, { viewport: { width: 320, height: 240 } });
    const comparison = await cvd.compare.region({
      name: "cta",
      left: leftPage.locator("#cta"),
      right: rightPage.locator("#cta"),
      outDir,
      styleProps: ["color", "font-size"],
      attributes: ["class"]
    });
    const written = await comparison.artifacts.write(outDir, ["json", "markdown"]);
    return {
      schemaVersion: comparison.toJSON().schemaVersion,
      changedPercent: comparison.pixel.summary().changedPercent,
      styleCount: comparison.styles.diff().length,
      boundsChanged: comparison.bounds.diff().changed,
      attrName: comparison.attributes.diff()[0].name,
      jsonPath: written.json,
      markdownPath: written.markdown,
      leftRegionPath: written.leftRegion,
      rightRegionPath: written.rightRegion,
      diffOnlyPath: written.diffOnly,
      diffComparisonPath: written.diffComparison,
      writtenCount: written.written.length
    };
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}
__verb__("compareRegionSmoke", { parents: ["custom"], fields: { leftUrl: { argument: true, required: true }, rightUrl: { argument: true, required: true }, outDir: { argument: true, required: true } } });
`)

	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	require.Len(t, commands, 1)

	outDir := t.TempDir()
	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{"default": {"leftUrl": left.URL, "rightUrl": right.URL, "outDir": outDir}}))
	require.NoError(t, err)
	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.Equal(t, "cssvd.selectionComparison.v1", row["schemaVersion"])
	require.Greater(t, row["changedPercent"].(float64), 0.0)
	require.EqualValues(t, 2, row["styleCount"])
	require.Equal(t, true, row["boundsChanged"])
	require.Equal(t, "class", row["attrName"])
	require.Equal(t, filepath.Join(outDir, "compare.json"), row["jsonPath"])
	require.Equal(t, filepath.Join(outDir, "compare.md"), row["markdownPath"])
	require.Equal(t, filepath.Join(outDir, "left_region.png"), row["leftRegionPath"])
	require.Equal(t, filepath.Join(outDir, "right_region.png"), row["rightRegionPath"])
	require.Equal(t, filepath.Join(outDir, "diff_only.png"), row["diffOnlyPath"])
	require.Equal(t, filepath.Join(outDir, "diff_comparison.png"), row["diffComparisonPath"])
	require.EqualValues(t, 2, row["writtenCount"])
	for _, name := range []string{"left_region.png", "right_region.png", "diff_only.png", "diff_comparison.png", "compare.json", "compare.md"} {
		_, err = os.Stat(filepath.Join(outDir, name))
		require.NoError(t, err, name)
	}
}

func TestCVDModuleCompareRegionRejectsRawObjects(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "compare-region-error.js"), `
async function compareRegionError() {
  const cvd = require("css-visual-diff");
  try {
    await cvd.compare.region({ left: { selector: "#cta" }, right: { selector: "#cta" } });
    return { ok: true };
  } catch (err) {
    return { ok: false, name: err.name, message: err.message };
  }
}
__verb__("compareRegionError", { parents: ["custom"], fields: {} });
`)
	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	parsedValues, err := glazerunner.ParseCommandValues(commands[0])
	require.NoError(t, err)
	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	row := rowToMap(processor.rows[0])
	require.Equal(t, false, row["ok"])
	require.Equal(t, "TypeError", row["name"])
	require.Contains(t, row["message"], "css-visual-diff.compare.region: expected cvd.locator")
}

func TestCVDModuleRecordsComparisonInCatalog(t *testing.T) {
	left := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="primary" style="color: rgb(0, 0, 0); font-size: 16px">Book now</button></body></html>`)
	}))
	defer left.Close()
	right := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" class="secondary" style="color: rgb(255, 0, 0); font-size: 18px">Book now</button></body></html>`)
	}))
	defer right.Close()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "comparison-catalog.js"), `
async function comparisonCatalogSmoke(leftUrl, rightUrl, outDir) {
  const cvd = require("css-visual-diff");
  const browser = await cvd.browser();
  let leftPage, rightPage;
  try {
    leftPage = await browser.page(leftUrl, { viewport: { width: 320, height: 240 } });
    rightPage = await browser.page(rightUrl, { viewport: { width: 320, height: 240 } });
    const comparison = await cvd.compare.region({
      name: "cta-comparison",
      left: leftPage.locator("#cta"),
      right: rightPage.locator("#cta"),
      outDir: outDir + "/artifacts/cta-comparison",
      styleProps: ["color", "font-size"],
      attributes: ["class"]
    });
    await comparison.artifacts.write(outDir + "/artifacts/cta-comparison", ["json", "markdown"]);
    const catalog = cvd.catalog.create({ title: "Comparison Catalog", outDir, artifactRoot: "artifacts" });
    catalog.record(comparison, { slug: "cta-comparison", name: "CTA comparison", url: leftUrl, selector: "#cta" });
    const manifestPath = await catalog.writeManifest();
    const indexPath = await catalog.writeIndex();
    const manifest = catalog.manifest();
    return {
      manifestPath,
      indexPath,
      comparisonCount: manifest.summary.comparisonCount,
      artifactCount: manifest.summary.artifactCount,
      firstComparisonName: manifest.comparisons[0].comparison.name,
      changedPercent: manifest.comparisons[0].comparison.pixel.changedPercent
    };
  } finally {
    if (leftPage) await leftPage.close();
    if (rightPage) await rightPage.close();
    await browser.close();
  }
}
__verb__("comparisonCatalogSmoke", { parents: ["custom"], fields: { leftUrl: { argument: true, required: true }, rightUrl: { argument: true, required: true }, outDir: { argument: true, required: true } } });
`)
	repositories, err := ScanRepositories(Bootstrap{Repositories: []Repository{{Name: "custom", Source: "test", RootDir: dir}}})
	require.NoError(t, err)
	discovered, err := CollectDiscoveredVerbs(repositories)
	require.NoError(t, err)
	commands, err := buildCommands(discovered, runtimeInvokerFactory)
	require.NoError(t, err)
	outDir := t.TempDir()
	parsedValues, err := glazerunner.ParseCommandValues(commands[0], glazerunner.WithValuesForSections(map[string]map[string]interface{}{"default": {"leftUrl": left.URL, "rightUrl": right.URL, "outDir": outDir}}))
	require.NoError(t, err)
	glazeCommand, ok := commands[0].(cmds.GlazeCommand)
	require.True(t, ok)
	processor := &captureProcessor{}
	require.NoError(t, glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, processor))
	require.Len(t, processor.rows, 1)
	row := rowToMap(processor.rows[0])
	require.EqualValues(t, 1, row["comparisonCount"])
	require.EqualValues(t, 2, row["artifactCount"])
	require.Equal(t, "cta-comparison", row["firstComparisonName"])
	require.Greater(t, row["changedPercent"].(float64), 0.0)
	manifestBytes, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	require.NoError(t, err)
	require.Contains(t, string(manifestBytes), `"comparisons"`)
	indexBytes, err := os.ReadFile(filepath.Join(outDir, "index.md"))
	require.NoError(t, err)
	require.Contains(t, string(indexBytes), "## Comparisons")
}
