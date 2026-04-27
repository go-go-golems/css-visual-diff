import type { RunReviewState } from "../types";

export function loadReviewState(runId: string): RunReviewState | null {
  try {
    const raw = localStorage.getItem(`cssvd-review-${runId}`);
    if (!raw) return null;
    return JSON.parse(raw) as RunReviewState;
  } catch {
    return null;
  }
}

export function saveReviewState(state: RunReviewState): void {
  try {
    localStorage.setItem(
      `cssvd-review-${state.runId}`,
      JSON.stringify(state),
    );
  } catch {
    // localStorage may be full or unavailable
  }
}

export function clearReviewState(runId: string): void {
  localStorage.removeItem(`cssvd-review-${runId}`);
}
