import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type ViewMode = "side-by-side" | "overlay" | "slider" | "diff";
export type SidebarTab = "comments" | "styles" | "meta";
export type CommentDraftType = "issue" | "note" | "question" | "praise";

interface ViewState {
  mode: ViewMode;
  sidebarTab: SidebarTab;
  commentMode: boolean;
  commentDraftType: CommentDraftType;
  overlayOpacity: number;
  overlayBlend: "normal" | "difference";
  sliderPosition: number;
}

const initialState: ViewState = {
  mode: "side-by-side",
  sidebarTab: "comments",
  commentMode: false,
  commentDraftType: "note",
  overlayOpacity: 50,
  overlayBlend: "normal",
  sliderPosition: 50,
};

const viewSlice = createSlice({
  name: "view",
  initialState,
  reducers: {
    setViewMode(state, action: PayloadAction<ViewMode>) {
      state.mode = action.payload;
    },
    setSidebarTab(state, action: PayloadAction<SidebarTab>) {
      state.sidebarTab = action.payload;
    },
    setCommentMode(state, action: PayloadAction<boolean>) {
      state.commentMode = action.payload;
    },
    setCommentDraftType(state, action: PayloadAction<CommentDraftType>) {
      state.commentDraftType = action.payload;
    },
    setOverlayOpacity(state, action: PayloadAction<number>) {
      state.overlayOpacity = action.payload;
    },
    setOverlayBlend(
      state,
      action: PayloadAction<"normal" | "difference">,
    ) {
      state.overlayBlend = action.payload;
    },
    setSliderPosition(state, action: PayloadAction<number>) {
      state.sliderPosition = action.payload;
    },
  },
});

export const {
  setViewMode,
  setSidebarTab,
  setCommentMode,
  setCommentDraftType,
  setOverlayOpacity,
  setOverlayBlend,
  setSliderPosition,
} = viewSlice.actions;

export default viewSlice.reducer;
