import { useState, useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import type { AppDispatch, RootState } from "./store";
import { fetchManifest, selectFilteredRows, selectCard } from "./store/slices/cardsSlice";
import { loadReviewState, setRunId, setStatus } from "./store/slices/reviewSlice";
import { setViewMode, setCommentMode } from "./store/slices/viewSlice";
import { loadReviewState as loadReviewFromStorage } from "./utils/storage";
import { Header } from "./components/Header";
import { CardList } from "./components/CardList";
import { Sidebar } from "./components/Sidebar";
import { ExportModal } from "./components/ExportModal";

export default function App() {
  const dispatch = useDispatch<AppDispatch>();
  const loading = useSelector((s: RootState) => s.cards.loading);
  const error = useSelector((s: RootState) => s.cards.error);
  const rows = useSelector((s: RootState) => s.cards.resolvedRows);
  const filteredRows = useSelector(selectFilteredRows);
  const selectedIdx = useSelector((s: RootState) => s.cards.selectedCardIndex);
  const runId = useSelector((s: RootState) => s.review.runId);
  const [exportOpen, setExportOpen] = useState(false);

  // Load manifest on mount
  useEffect(() => {
    dispatch(fetchManifest());
  }, [dispatch]);

  // Load persisted review state
  useEffect(() => {
    if (!runId) return;
    const saved = loadReviewFromStorage(runId);
    if (saved) {
      dispatch(loadReviewState({ runId: saved.runId, cards: saved.cards }));
    }
  }, [runId, dispatch]);

  // Derive runId
  useEffect(() => {
    if (rows.length > 0 && !runId) {
      const id = rows.map((row) => `${row.page}/${row.section}`).join(",");
      let hash = 0;
      for (let i = 0; i < id.length; i++) {
        hash = ((hash << 5) - hash + id.charCodeAt(i)) | 0;
      }
      dispatch(setRunId(`run-${Math.abs(hash).toString(36)}`));
    }
  }, [rows, runId, dispatch]);

  // Global keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // Skip when typing in inputs
      const tag = (e.target as HTMLElement).tagName;
      if (tag === "TEXTAREA" || tag === "INPUT" || tag === "SELECT") return;

      const currentRow = filteredRows[selectedIdx];
      if (!currentRow) return;

      switch (e.key) {
        // Navigation
        case "j":
          e.preventDefault();
          if (selectedIdx < filteredRows.length - 1)
            dispatch(selectCard(selectedIdx + 1));
          break;
        case "k":
          e.preventDefault();
          if (selectedIdx > 0) dispatch(selectCard(selectedIdx - 1));
          break;

        // Status shortcuts
        case "a":
          e.preventDefault();
          dispatch(setStatus({ page: currentRow.page, section: currentRow.section, status: "accepted" }));
          break;
        case "n":
          e.preventDefault();
          dispatch(setStatus({ page: currentRow.page, section: currentRow.section, status: "needs-work" }));
          break;
        case "w":
          e.preventDefault();
          dispatch(setStatus({ page: currentRow.page, section: currentRow.section, status: "wont-fix" }));
          break;
        case "x":
          e.preventDefault();
          dispatch(setStatus({ page: currentRow.page, section: currentRow.section, status: "fixed" }));
          break;

        // View modes
        case "1":
          e.preventDefault();
          dispatch(setViewMode("side-by-side"));
          break;
        case "2":
          e.preventDefault();
          dispatch(setViewMode("overlay"));
          break;
        case "3":
          e.preventDefault();
          dispatch(setViewMode("slider"));
          break;
        case "4":
          e.preventDefault();
          dispatch(setViewMode("diff"));
          break;

        // Export
        case "e":
          e.preventDefault();
          setExportOpen(true);
          break;

        // Comment mode toggle
        case "p":
          e.preventDefault();
          dispatch(setCommentMode(true));
          break;
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [dispatch, filteredRows, selectedIdx]);

  // Callback for Header to open export modal
  // (wired via onExport prop)

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream text-text-muted">
        Loading manifest…
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream text-red">
        Error: {error}
      </div>
    );
  }

  if (rows.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-cream text-text-muted">
        No comparison results found.
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-cream text-text font-sans">
      <Header onExport={() => setExportOpen(true)} />
      <main className="max-w-[1400px] mx-auto px-6 py-5 grid grid-cols-[1fr_360px] gap-6">
        <CardList />
        <Sidebar />
      </main>
      <ExportModal open={exportOpen} onClose={() => setExportOpen(false)} />
    </div>
  );
}
