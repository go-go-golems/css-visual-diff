import {
  createSlice,
  createAsyncThunk,
  type PayloadAction,
} from "@reduxjs/toolkit";
import type {
  SummaryRow,
  SuiteSummary,
  SummaryPayload,
  Classification,
} from "../../types";
import { toArtifactUrl } from "../../utils/paths";

interface CardsState {
  rows: SummaryRow[];
  /** Rows after rewriting absolute paths to relative artifact URLs. */
  resolvedRows: SummaryRow[];
  loading: boolean;
  error: string | null;
  filter: {
    classification: Classification | null;
    status: string | null;
    search: string;
  };
  selectedCardIndex: number;
  classificationCounts: Record<string, number>;
  worstClassification: string;
  pageCount: number;
}

const initialState: CardsState = {
  rows: [],
  resolvedRows: [],
  loading: false,
  error: null,
  filter: { classification: null, status: null, search: "" },
  selectedCardIndex: 0,
  classificationCounts: {},
  worstClassification: "",
  pageCount: 0,
};

export const fetchManifest = createAsyncThunk(
  "cards/fetchManifest",
  async () => {
    const res = await fetch("/api/manifest");
    if (!res.ok) throw new Error(`Failed to load manifest: ${res.status}`);
    const payload: SummaryPayload = await res.json();
    const summary: SuiteSummary = Array.isArray(payload)
      ? payload[0]
      : payload;
    return summary;
  },
);

function rewriteRowPaths(row: SummaryRow): SummaryRow {
  return {
    ...row,
    diffOnlyPath: toArtifactUrl(row.diffOnlyPath),
    diffComparisonPath: toArtifactUrl(row.diffComparisonPath),
    leftRegionPath: toArtifactUrl(row.leftRegionPath),
    rightRegionPath: toArtifactUrl(row.rightRegionPath),
    artifactJson: toArtifactUrl(row.artifactJson),
  };
}

const cardsSlice = createSlice({
  name: "cards",
  initialState,
  reducers: {
    setClassificationFilter(
      state,
      action: PayloadAction<Classification | null>,
    ) {
      state.filter.classification = action.payload;
    },
    setStatusFilter(state, action: PayloadAction<string | null>) {
      state.filter.status = action.payload;
    },
    setSearch(state, action: PayloadAction<string>) {
      state.filter.search = action.payload;
    },
    selectCard(state, action: PayloadAction<number>) {
      state.selectedCardIndex = action.payload;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchManifest.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchManifest.fulfilled, (state, action) => {
        state.loading = false;
        state.rows = action.payload.rows;
        state.resolvedRows = action.payload.rows.map(rewriteRowPaths);
        state.classificationCounts = action.payload.classificationCounts;
        state.worstClassification =
          action.payload.policy.worstClassification;
        state.pageCount = action.payload.pageCount;
      })
      .addCase(fetchManifest.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || "Failed to load";
      });
  },
});

export const {
  setClassificationFilter,
  setStatusFilter,
  setSearch,
  selectCard,
} = cardsSlice.actions;

export default cardsSlice.reducer;

/* Selectors */

export const selectAllRows = (state: { cards: CardsState }) =>
  state.cards.resolvedRows;

export const selectFilteredRows = (state: { cards: CardsState }) => {
  const { resolvedRows, filter } = state.cards;
  return resolvedRows.filter((row) => {
    if (
      filter.classification &&
      row.classification !== filter.classification
    )
      return false;
    if (
      filter.search &&
      !`${row.page}/${row.section}`
        .toLowerCase()
        .includes(filter.search.toLowerCase())
    )
      return false;
    return true;
  });
};

export const selectSelectedRow = (state: { cards: CardsState }) => {
  const rows = selectFilteredRows(state);
  return rows[state.cards.selectedCardIndex] ?? null;
};
