/** Full comparison result from per-section compare.json. */
export interface CompareData {
  schemaVersion: string;
  name: string;
  bounds: import("./summary").BoundsComparison;
  left: CompareSide;
  right: CompareSide;
  pixel: PixelData;
  styles: StyleChange[];
  attributes: AttributeChange[];
  text?: TextComparison;
  artifacts: ArtifactRef[];
}

export interface CompareSide {
  name: string;
  selector: string;
  url: string;
  bounds: { height: number; width: number; x: number; y: number };
  exists: boolean;
  visible: boolean;
}

export interface PixelData {
  changedPercent: number;
  changedPixels: number;
  totalPixels: number;
  normalizedWidth: number;
  normalizedHeight: number;
  threshold: number;
  diffOnlyPath: string;
  diffComparisonPath: string;
}

export interface StyleChange {
  changed: boolean;
  name: string;
  left: string;
  right: string;
}

export interface AttributeChange {
  changed: boolean;
  name: string;
  left?: string | null;
  right?: string | null;
}

export interface TextComparison {
  changed: boolean;
  left: string;
  right: string;
}

export interface ArtifactRef {
  kind: string;
  name: string;
  path: string;
}
