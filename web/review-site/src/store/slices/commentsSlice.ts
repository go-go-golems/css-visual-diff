import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { CommentPin } from "../../types";
import { makeCardKey } from "../../types";

interface CommentsState {
  pins: Record<string, CommentPin[]>; // key = "page/section"
}

const initialState: CommentsState = {
  pins: {},
};

const commentsSlice = createSlice({
  name: "comments",
  initialState,
  reducers: {
    addPin(
      state,
      action: PayloadAction<{
        page: string;
        section: string;
        pin: CommentPin;
      }>,
    ) {
      const key = makeCardKey(action.payload.page, action.payload.section);
      if (!state.pins[key]) state.pins[key] = [];
      state.pins[key].push(action.payload.pin);
    },
    updatePin(
      state,
      action: PayloadAction<{
        page: string;
        section: string;
        id: string;
        patch: Partial<CommentPin>;
      }>,
    ) {
      const key = makeCardKey(action.payload.page, action.payload.section);
      const pins = state.pins[key] ?? [];
      const idx = pins.findIndex((p) => p.id === action.payload.id);
      if (idx >= 0) {
        pins[idx] = { ...pins[idx], ...action.payload.patch };
      }
    },
    deletePin(
      state,
      action: PayloadAction<{
        page: string;
        section: string;
        id: string;
      }>,
    ) {
      const key = makeCardKey(action.payload.page, action.payload.section);
      state.pins[key] = (state.pins[key] ?? []).filter(
        (p) => p.id !== action.payload.id,
      );
    },
    setPinsForCard(
      state,
      action: PayloadAction<{
        page: string;
        section: string;
        pins: CommentPin[];
      }>,
    ) {
      const key = makeCardKey(action.payload.page, action.payload.section);
      state.pins[key] = action.payload.pins;
    },
  },
});

export const { addPin, updatePin, deletePin, setPinsForCard } =
  commentsSlice.actions;
export default commentsSlice.reducer;

/* Selector */
export const selectPins = (page: string, section: string) =>
  (state: { comments: CommentsState }): CommentPin[] =>
    state.comments.pins[makeCardKey(page, section)] ?? [];
