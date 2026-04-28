import type { AttributeChange, BoundsComparison, CompareData, StyleChange } from "../types";

type LooseCompareData = CompareData & Record<string, unknown>;

type SnakeStyleDiff = {
  changed?: boolean;
  property?: string;
  name?: string;
  left?: string;
  right?: string;
};

type SnakePixel = {
  changed_percent?: number;
  changed_pixels?: number;
  total_pixels?: number;
  threshold?: number;
};

type SnakeSide = {
  url?: string;
  selector?: string;
  computed?: {
    bounds?: { height: number; width: number; x: number; y: number };
    attributes?: Record<string, string | null | undefined>;
  };
};

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : {};
}

function snakeSide(data: LooseCompareData, key: "url1" | "url2"): SnakeSide {
  return asRecord(data[key]) as SnakeSide;
}

export function changedStyles(compareData: CompareData | null): StyleChange[] {
  if (!compareData) return [];
  const loose = compareData as LooseCompareData;
  if (Array.isArray(compareData.styles)) {
    return compareData.styles.filter((s) => s.changed);
  }
  const snakeDiffs = loose.computed_diffs;
  if (!Array.isArray(snakeDiffs)) return [];
  return (snakeDiffs as SnakeStyleDiff[])
    .filter((s) => s.changed)
    .map((s) => ({
      changed: true,
      name: s.name ?? s.property ?? "",
      left: s.left ?? "",
      right: s.right ?? "",
    }));
}

export function changedAttributes(compareData: CompareData | null): AttributeChange[] {
  if (!compareData) return [];
  const loose = compareData as LooseCompareData;
  if (Array.isArray(compareData.attributes)) {
    return compareData.attributes.filter((a) => a.changed);
  }

  const leftAttrs = snakeSide(loose, "url1").computed?.attributes ?? {};
  const rightAttrs = snakeSide(loose, "url2").computed?.attributes ?? {};
  return Object.keys({ ...leftAttrs, ...rightAttrs })
    .filter((name) => leftAttrs[name] !== rightAttrs[name])
    .map((name) => ({
      changed: true,
      name,
      left: leftAttrs[name] ?? null,
      right: rightAttrs[name] ?? null,
    }));
}

export function compareBounds(compareData: CompareData | null): BoundsComparison | null {
  if (!compareData) return null;
  if (compareData.bounds?.left && compareData.bounds?.right && compareData.bounds?.delta) {
    return compareData.bounds;
  }
  const loose = compareData as LooseCompareData;
  const left = snakeSide(loose, "url1").computed?.bounds;
  const right = snakeSide(loose, "url2").computed?.bounds;
  if (!left || !right) return null;
  return {
    changed: left.width !== right.width || left.height !== right.height || left.x !== right.x || left.y !== right.y,
    delta: {
      height: right.height - left.height,
      width: right.width - left.width,
      x: right.x - left.x,
      y: right.y - left.y,
    },
    left,
    right,
  };
}

export function compareSourceUrls(compareData: CompareData | null): { left: string; right: string } | null {
  if (!compareData) return null;
  const loose = compareData as LooseCompareData;
  return {
    left: compareData.left?.url ?? snakeSide(loose, "url1").url ?? "",
    right: compareData.right?.url ?? snakeSide(loose, "url2").url ?? "",
  };
}

export function comparePixel(compareData: CompareData | null): { changedPixels: number; totalPixels: number; threshold: number } | null {
  if (!compareData) return null;
  if (compareData.pixel) {
    return {
      changedPixels: compareData.pixel.changedPixels,
      totalPixels: compareData.pixel.totalPixels,
      threshold: compareData.pixel.threshold,
    };
  }
  const pixel = (compareData as LooseCompareData).pixel_diff as SnakePixel | undefined;
  if (!pixel) return null;
  return {
    changedPixels: pixel.changed_pixels ?? 0,
    totalPixels: pixel.total_pixels ?? 0,
    threshold: pixel.threshold ?? 0,
  };
}
