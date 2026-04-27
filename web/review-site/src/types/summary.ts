/** Computed classification from pixel-change percentage. */
export type Classification =
  | "accepted"
  | "review"
  | "tune-required"
  | "major-mismatch";

/** Human reviewer status (independent of classification). */
export type ReviewStatus =
  | "unreviewed"
  | "accepted"
  | "needs-work"
  | "fixed"
  | "wont-fix";

/** Bounds comparison between prototype and React. */
export interface BoundsComparison {
  changed: boolean;
  delta: { height: number; width: number; x: number; y: number };
  left: { height: number; width: number; x: number; y: number };
  right: { height: number; width: number; x: number; y: number };
  normalizedWidth?: number;
  normalizedHeight?: number;
}

/** A single changed CSS property. */
export interface StyleDiff {
  property: string;
  left: string;
  right: string;
}

/** A single changed HTML attribute. */
export interface AttributeDiff {
  attribute: string;
  left: string | null;
  right: string;
}

/** One row from the summary JSON — one page/section to review. */
export interface SummaryRow {
  page: string;
  section: string;
  classification: Classification;
  changedPercent: number;
  changedPixels: number;
  totalPixels?: number;
  diffOnlyPath: string;
  diffComparisonPath: string;
  leftRegionPath: string;
  rightRegionPath: string;
  artifactJson: string;
  artifactMarkdown?: string;
  leftSelector: string;
  rightSelector: string;
  styleChangeCount: number;
  attributeChangeCount: number;
  styleDiffs: StyleDiff[];
  attributeDiffs: AttributeDiff[];
  bounds: BoundsComparison;
}

/** Top-level suite summary shape. */
export interface SuiteSummary {
  classificationCounts: Record<string, number>;
  pageCount: number;
  maxChangedPercent: number;
  policy: {
    ok: boolean;
    worstClassification: string;
    failureCount: number;
  };
  rows: SummaryRow[];
}

/** The manifest can be either a bare object or a single-element array. */
export type SummaryPayload = SuiteSummary | [SuiteSummary];
