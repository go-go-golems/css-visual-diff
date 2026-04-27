import type { Middleware } from "@reduxjs/toolkit";
import type { RunReviewState, CardReview } from "../../types";
import { saveReviewState } from "../../utils/storage";

/**
 * Middleware that persists review + comments state to localStorage
 * after every review/comment action.
 */
export const localStorageSync: Middleware = (store) => (next) => (action) => {
  const result = next(action);

  const typedAction = action as { type: string };
  if (
    typedAction.type.startsWith("review/") ||
    typedAction.type.startsWith("comments/")
  ) {
    const state = store.getState() as {
      review: { runId: string; cards: Record<string, CardReview> };
      comments: { pins: Record<string, unknown[]> };
    };
    const reviewState: RunReviewState = {
      version: 1,
      runId: state.review.runId,
      cards: Object.fromEntries(
        Object.entries(state.review.cards).map(([key, card]) => {
          const pins = (state.comments.pins[key] ?? []) as CardReview["comments"];
          return [key, { ...card, comments: pins }];
        }),
      ),
      updatedAt: new Date().toISOString(),
    };
    saveReviewState(reviewState);
  }

  return result;
};
