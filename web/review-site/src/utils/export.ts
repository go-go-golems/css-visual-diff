import type {
  SummaryRow,
  CompareData,
  CardReview,
} from "../types";
import { changedStyles, compareBounds, comparePixel, compareSourceUrls } from "./compareData";

/**
 * Build a markdown + YAML export block for one card.
 */
export function buildExportMarkdown(
  row: SummaryRow,
  compareData: CompareData | null,
  review: CardReview,
): string {
  const parts: string[] = [];

  parts.push(`# Visual review — \`${row.page} / ${row.section}\``);
  parts.push("");
  parts.push(
    `**Classification:** ${row.classification} · ${row.changedPercent.toFixed(2)}% changed`,
  );
  parts.push(`**Status:** ${review.status}`);
  const bounds = compareBounds(compareData);
  const sources = compareSourceUrls(compareData);
  const pixel = comparePixel(compareData);
  const styles = changedStyles(compareData);

  if (bounds) {
    parts.push(
      `**Bounds delta:** Δheight ${bounds.delta.height}px · Δwidth ${bounds.delta.width}px`,
    );
  }
  parts.push("");

  if (review.note) {
    parts.push("## General observation");
    parts.push("");
    parts.push(review.note);
    parts.push("");
  }

  if (review.comments.length > 0) {
    parts.push("## Pin-drop comments");
    parts.push("");
    review.comments.forEach((c, i) => {
      const sideLabel =
        c.side === "left"
          ? "prototype"
          : c.side === "right"
            ? "react"
            : "merged";
      parts.push(
        `${i + 1}. **[${c.type}]** _${sideLabel} @ ${c.x.toFixed(0)}%, ${c.y.toFixed(0)}%_ — ${c.text || "(no text)"}`,
      );
    });
    parts.push("");
  }

  if (row.styleDiffs && row.styleDiffs.length > 0) {
    parts.push("## Computed style diffs");
    parts.push("");
    row.styleDiffs.forEach((s) => {
      parts.push(`- \`${s.property}\`: \`${s.left}\` → \`${s.right}\``);
    });
    parts.push("");
  }

  parts.push("## Source");
  parts.push("");
  if (sources) {
    parts.push(`- prototype: ${sources.left}`);
    parts.push(`- react: ${sources.right}`);
  }
  parts.push("");

  parts.push("---");
  parts.push("");
  parts.push("```yaml");
  const yamlObj: Record<string, unknown> = {
    page: row.page,
    section: row.section,
    classification: row.classification,
    changedPercent: row.changedPercent,
  };
  if (bounds) {
    yamlObj.bounds = {
      delta: bounds.delta,
      left: bounds.left,
      right: bounds.right,
    };
  }
  if (pixel) {
    yamlObj.pixel = pixel;
  }
  if (styles.length > 0) {
    yamlObj.changedStyles = styles
      .slice(0, 20)
      .map((s) => ({ name: s.name, left: s.left, right: s.right }));
  }
  yamlObj.review = {
    status: review.status,
    generalNote: review.note,
    comments: review.comments.map((c, i) => ({
      id: i + 1,
      type: c.type,
      side: c.side,
      position: { x: +c.x.toFixed(2), y: +c.y.toFixed(2) },
      text: c.text,
    })),
  };
  parts.push(toYaml(yamlObj));
  parts.push("```");

  return parts.join("\n");
}

/**
 * Very minimal YAML serializer — enough for our export format.
 */
function toYaml(obj: unknown, indent = 0): string {
  const pad = "  ".repeat(indent);
  if (obj === null || obj === undefined) return "null";
  if (typeof obj === "string") {
    if (obj.includes("\n") || obj.includes(":") || obj.includes('"')) {
      return JSON.stringify(obj);
    }
    return obj;
  }
  if (typeof obj === "number" || typeof obj === "boolean") return String(obj);
  if (Array.isArray(obj)) {
    if (obj.length === 0) return "[]";
    return obj
      .map((item) => {
        const val = toYaml(item, indent + 1);
        if (typeof item === "object" && item !== null && !Array.isArray(item)) {
          const inner = Object.entries(item as Record<string, unknown>)
            .map(
              ([k, v]) =>
                `${pad}  ${k}: ${toYaml(v, indent + 2)}`,
            )
            .join("\n");
          return `${pad}- ${inner}`;
        }
        return `${pad}- ${val}`;
      })
      .join("\n");
  }
  if (typeof obj === "object") {
    return Object.entries(obj as Record<string, unknown>)
      .map(([k, v]) => {
        const val = toYaml(v, indent + 1);
        if (typeof v === "object" && v !== null) {
          return `${pad}${k}:\n${val}`;
        }
        return `${pad}${k}: ${val}`;
      })
      .join("\n");
  }
  return String(obj);
}

/**
 * Build export for all cards, separated by `---`.
 */
export function buildAllExportMarkdown(
  rows: SummaryRow[],
  compareDataMap: Record<string, CompareData | null>,
  reviewMap: Record<string, CardReview>,
): string {
  return rows
    .map((row) => {
      const key = `${row.page}/${row.section}`;
      return buildExportMarkdown(
        row,
        compareDataMap[key] ?? null,
        reviewMap[key] ?? {
          page: row.page,
          section: row.section,
          status: "unreviewed" as const,
          note: "",
          comments: [],
          updatedAt: "",
        },
      );
    })
    .join("\n\n---\n\n");
}
