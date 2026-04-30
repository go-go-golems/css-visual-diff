__package__({
  name: "catalog",
  short: "Script-backed visual catalog workflows"
});

function _splitList(value) {
  if (!value) return [];
  if (Array.isArray(value)) return value.filter(Boolean).map(String);
  return String(value)
    .split(",")
    .map((part) => part.trim())
    .filter(Boolean);
}

function _catalogFailureRow(target, preflight, manifestPath, indexPath, message, code = "SELECTOR_ERROR") {
  return {
    ok: false,
    slug: target.slug,
    url: target.url,
    selector: target.selector,
    exists: preflight ? preflight.exists : false,
    visible: preflight ? preflight.visible : false,
    manifestPath,
    indexPath,
    code,
    message,
  };
}

async function inspectPage(url, selector, outDir, values) {
  values = values || {};
  values.url = url;
  values.selector = selector;
  values.outDir = outDir;
  const cvd = require("css-visual-diff");
  const width = values.width || 1280;
  const height = values.height || 720;
  const slug = values.slug || values.name || values.selector;
  const name = values.name || slug || "page";
  const artifacts = values.artifacts || "bundle";
  const failOnMissing = values.failOnMissing === true;
  const target = {
    slug,
    name,
    url: values.url,
    selector: values.selector,
    viewport: { width, height },
    metadata: {
      source: "builtin:catalog.inspectPage",
    },
  };
  const probe = {
    name: slug || name,
    selector: values.selector,
    props: _splitList(values.props) || ["display", "color", "background-color"],
    attributes: _splitList(values.attributes),
    source: "catalog.inspectPage",
    required: failOnMissing,
  };

  const catalog = cvd.catalog({
    title: values.title || "css-visual-diff Page Catalog",
    outDir,
    artifactRoot: values.artifactRoot || "artifacts",
  });
  catalog.addTarget(target);

  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(values.url, {
      viewport: target.viewport,
      waitMs: values.waitMs || 0,
      name,
    });

    const statuses = await page.preflight([probe]);
    catalog.recordPreflight(target, statuses);
    const status = statuses[0];
    if (!status || !status.exists || status.error) {
      const message = status && status.error ? status.error : `selector did not match: ${values.selector}`;
      const err = new cvd.SelectorError(message, "SELECTOR_ERROR", { selector: values.selector, target });
      err.operation = "css-visual-diff.verbs.catalog.inspectPage";
      catalog.addFailure(target, err);
      const manifestPath = await catalog.writeManifest();
      const indexPath = await catalog.writeIndex();
      if (failOnMissing) throw err;
      return _catalogFailureRow(target, status, manifestPath, indexPath, message);
    }

    const artifactDir = catalog.artifactDir(target.slug);
    const artifact = await page.inspect(probe, { outDir: artifactDir, artifacts });
    const result = { outputDir: artifactDir, results: [artifact] };
    catalog.addResult(target, result);
    const manifestPath = await catalog.writeManifest();
    const indexPath = await catalog.writeIndex();
    await page.close();
    page = undefined;
    const summary = catalog.summary();
    return {
      ok: true,
      slug: target.slug,
      url: target.url,
      selector: target.selector,
      exists: status.exists,
      visible: status.visible,
      artifactDir,
      manifestPath,
      indexPath,
      resultCount: summary.resultCount,
      failureCount: summary.failureCount,
    };
  } catch (err) {
    catalog.addFailure(target, err);
    const manifestPath = await catalog.writeManifest();
    const indexPath = await catalog.writeIndex();
    if (failOnMissing) throw err;
    return _catalogFailureRow(target, undefined, manifestPath, indexPath, err && err.message ? err.message : String(err), err && err.code ? err.code : "CVD_ERROR");
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}

__verb__("inspectPage", {
  short: "Inspect one URL/selector and write a visual catalog manifest",
  fields: {
    values: { bind: "all" },
    url: {
      argument: true,
      required: true,
      help: "URL to inspect"
    },
    selector: {
      argument: true,
      required: true,
      help: "CSS selector to inspect"
    },
    outDir: {
      argument: true,
      required: true,
      help: "Output directory for the catalog"
    },
    slug: {
      type: "string",
      help: "Catalog target slug (defaults to name or selector)"
    },
    name: {
      type: "string",
      help: "Human-readable target name"
    },
    title: {
      type: "string",
      default: "css-visual-diff Page Catalog",
      help: "Catalog title"
    },
    artifactRoot: {
      type: "string",
      default: "artifacts",
      help: "Relative directory under outDir for inspect artifacts"
    },
    artifacts: {
      type: "choice",
      choices: ["bundle", "png", "html", "css-json", "css-md", "inspect-json", "metadata-json"],
      default: "bundle",
      help: "Artifact format to write"
    },
    waitMs: {
      type: "int",
      default: 0,
      help: "Wait after navigation in milliseconds"
    },
    width: {
      type: "int",
      default: 1280,
      help: "Viewport width"
    },
    height: {
      type: "int",
      default: 720,
      help: "Viewport height"
    },
    props: {
      type: "string",
      help: "Comma-separated computed CSS properties for CSS artifacts"
    },
    attributes: {
      type: "string",
      help: "Comma-separated attributes for CSS artifacts"
    },
    failOnMissing: {
      type: "bool",
      default: false,
      help: "Return a non-zero error when the selector is missing (CI mode)"
    }
  }
});
