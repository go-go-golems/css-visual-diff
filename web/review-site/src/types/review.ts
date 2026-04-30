import type { ReviewStatus } from "./summary";

/** Comment pin type. */
export type CommentType = "issue" | "note" | "question" | "praise";

/** Comment pin side. */
export type CommentSide = "left" | "right" | "merged";

/** A single comment pin dropped on a screenshot. */
export interface CommentPin {
  id: string;
  x: number; // percentage 0-100
  y: number; // percentage 0-100
  side: CommentSide;
  type: CommentType;
  text: string;
}

/** The review state for one card (page/section). */
export interface CardReview {
  page: string;
  section: string;
  status: ReviewStatus;
  note: string;
  comments: CommentPin[];
  updatedAt: string; // ISO timestamp
}

/** Everything stored in localStorage for one run. */
export interface RunReviewState {
  version: 1;
  runId: string;
  cards: Record<string, CardReview>; // key = "page/section"
  updatedAt: string;
}

export const DEFAULT_STATUS: ReviewStatus = "unreviewed";

export function makeCardKey(page: string, section: string): string {
  return `${page}/${section}`;
}

export function defaultCardReview(
  page: string,
  section: string,
): CardReview {
  return {
    page,
    section,
    status: "unreviewed",
    note: "",
    comments: [],
    updatedAt: new Date().toISOString(),
  };
}
