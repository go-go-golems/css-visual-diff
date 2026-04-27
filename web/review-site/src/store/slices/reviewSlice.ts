import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { ReviewStatus, CardReview } from "../../types";
import { makeCardKey, defaultCardReview } from "../../types";

interface ReviewState {
  runId: string;
  cards: Record<string, CardReview>;
}

const initialState: ReviewState = {
  runId: "",
  cards: {},
};

const reviewSlice = createSlice({
  name: "review",
  initialState,
  reducers: {
    setRunId(state, action: PayloadAction<string>) {
      state.runId = action.payload;
    },
    setStatus(
      state,
      action: PayloadAction<{ page: string; section: string; status: ReviewStatus }>,
    ) {
      const key = makeCardKey(action.payload.page, action.payload.section);
      if (!state.cards[key]) {
        state.cards[key] = defaultCardReview(
          action.payload.page,
          action.payload.section,
        );
      }
      state.cards[key].status = action.payload.status;
      state.cards[key].updatedAt = new Date().toISOString();
    },
    setNote(
      state,
      action: PayloadAction<{ page: string; section: string; note: string }>,
    ) {
      const key = makeCardKey(action.payload.page, action.payload.section);
      if (!state.cards[key]) {
        state.cards[key] = defaultCardReview(
          action.payload.page,
          action.payload.section,
        );
      }
      state.cards[key].note = action.payload.note;
      state.cards[key].updatedAt = new Date().toISOString();
    },
    loadReviewState(
      state,
      action: PayloadAction<{ runId: string; cards: Record<string, CardReview> }>,
    ) {
      state.runId = action.payload.runId;
      state.cards = action.payload.cards;
    },
  },
});

export const { setRunId, setStatus, setNote, loadReviewState } =
  reviewSlice.actions;
export default reviewSlice.reducer;

/* Selector */
export const selectReview = (page: string, section: string) =>
  (state: { review: ReviewState }): CardReview =>
    state.review.cards[makeCardKey(page, section)] ??
    defaultCardReview(page, section);
