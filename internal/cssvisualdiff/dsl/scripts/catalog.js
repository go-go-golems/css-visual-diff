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

function _selectorForSide(spec, side) {
  if (side === "react") return spec.selectorReact || spec.selector || spec.selectorOriginal;
  return spec.selectorOriginal || spec.selector || spec.selectorReact;
}

function _probesFromConfig(cfg, side) {
  const probes = [];
  for (const style of cfg.styles || []) {
    const selector = _selectorForSide(style, side);
    if (!selector) continue;
    probes.push({
      name: style.name || selector,
      selector,
      props: style.props || [],
      attributes: style.attributes || [],
      source: `config.styles.${side}`,
    });
  }
  if (probes.length === 0) {
    for (const section of cfg.sections || []) {
      const selector = _selectorForSide(section, side);
      if (!selector) continue;
      probes.push({
        name: section.name || selector,
        selector,
        source: `config.sections.${side}`,
      });
    }
  }
  return probes;
}

async function inspectConfig(configPath, side, outDir, values) {
  values = values || {};
  const cvd = require("css-visual-diff");
  const cfg = await cvd.loadConfig(configPath);
  const targetConfig = side === "react" ? cfg.react : cfg.original;
  const probes = _probesFromConfig(cfg, side);
  const slug = values.slug || `${cfg.metadata.slug || "config"}-${side}`;
  const target = {
    slug,
    name: values.name || targetConfig.name || side,
    url: targetConfig.url,
    selector: targetConfig.rootSelector || "",
    viewport: targetConfig.viewport,
    metadata: {
      source: "builtin:catalog.inspectConfig",
      configPath,
      side,
    },
  };
  const catalog = cvd.catalog({
    title: values.title || `${cfg.metadata.title || cfg.metadata.slug || "css-visual-diff Config"} (${side})`,
    outDir,
    artifactRoot: values.artifactRoot || "artifacts",
  });
  catalog.addTarget(target);
  const browser = await cvd.browser();
  let page;
  try {
    page = await browser.page(target.url, {
      viewport: target.viewport,
      waitMs: targetConfig.waitMs || 0,
      name: target.name,
    });
    if (targetConfig.prepare) {
      await page.prepare(targetConfig.prepare);
    }
    if (probes.length === 0) {
      const err = new cvd.SelectorError("config did not contain styles or sections to inspect", "SELECTOR_ERROR", { configPath, side });
      err.operation = "css-visual-diff.verbs.catalog.inspectConfig";
      catalog.addFailure(target, err);
      const manifestPath = await catalog.writeManifest();
      const indexPath = await catalog.writeIndex();
      if (values.failOnMissing) throw err;
      return { ok: false, slug, missingCount: 0, inspectedCount: 0, manifestPath, indexPath, code: err.code, message: err.message };
    }
    const statuses = await page.preflight(probes);
    catalog.recordPreflight(target, statuses);
    const readyProbes = [];
    const missing = [];
    for (let i = 0; i < probes.length; i++) {
      const status = statuses[i];
      if (status && status.exists && !status.error) readyProbes.push(probes[i]);
      else missing.push({ probe: probes[i], status });
    }
    if (missing.length > 0) {
      const err = new cvd.SelectorError(`${missing.length} selector(s) missing`, "SELECTOR_ERROR", { missing });
      err.operation = "css-visual-diff.verbs.catalog.inspectConfig";
      catalog.addFailure(target, err);
      if (values.failOnMissing) {
        await catalog.writeManifest();
        await catalog.writeIndex();
        throw err;
      }
    }
    let result = { outputDir: catalog.artifactDir(slug), results: [] };
    if (readyProbes.length > 0) {
      result = await page.inspectAll(readyProbes, { outDir: catalog.artifactDir(slug), artifacts: values.artifacts || "css-json" });
      catalog.addResult(target, result);
    }
    const manifestPath = await catalog.writeManifest();
    const indexPath = await catalog.writeIndex();
    const summary = catalog.summary();
    return {
      ok: missing.length === 0,
      slug,
      side,
      manifestPath,
      indexPath,
      inspectedCount: readyProbes.length,
      missingCount: missing.length,
      resultCount: summary.resultCount,
      failureCount: summary.failureCount,
    };
  } catch (err) {
    catalog.addFailure(target, err);
    await catalog.writeManifest();
    await catalog.writeIndex();
    if (values.failOnMissing) throw err;
    return { ok: false, slug, side, inspectedCount: 0, missingCount: probes.length, code: err.code || "CVD_ERROR", message: err.message || String(err) };
  } finally {
    if (page) await page.close();
    await browser.close();
  }
}

__verb__("inspectConfig", {
  short: "Inspect styles or sections from a css-visual-diff YAML config into a catalog",
  fields: {
    configPath: {
      argument: true,
      required: true,
      help: "Path to a css-visual-diff YAML config"
    },
    side: {
      argument: true,
      type: "choice",
      choices: ["original", "react"],
      required: true,
      help: "Config side to inspect"
    },
    outDir: {
      argument: true,
      required: true,
      help: "Output directory for the catalog"
    },
    values: { bind: "all" },
    slug: { type: "string", help: "Catalog target slug" },
    name: { type: "string", help: "Catalog target name" },
    title: { type: "string", help: "Catalog title" },
    artifactRoot: { type: "string", default: "artifacts", help: "Relative artifact root" },
    artifacts: {
      type: "choice",
      choices: ["bundle", "png", "html", "css-json", "css-md", "inspect-json", "metadata-json"],
      default: "css-json",
      help: "Artifact format to write"
    },
    failOnMissing: {
      type: "bool",
      default: false,
      help: "Return a non-zero error when any config selector is missing"
    }
  }
});

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
