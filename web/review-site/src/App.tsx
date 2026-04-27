import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import type { AppDispatch, RootState } from "./store";
import { fetchManifest } from "./store/slices/cardsSlice";
import { loadReviewState, setRunId } from "./store/slices/reviewSlice";
import { loadReviewState as loadReviewFromStorage } from "./utils/storage";
import { Header } from "./components/Header";
import { CardList } from "./components/CardList";
import { Sidebar } from "./components/Sidebar";

export default function App() {
  const dispatch = useDispatch<AppDispatch>();
  const loading = useSelector((s: RootState) => s.cards.loading);
  const error = useSelector((s: RootState) => s.cards.error);
  const rows = useSelector((s: RootState) => s.cards.resolvedRows);
  const runId = useSelector((s: RootState) => s.review.runId);

  useEffect(() => {
    dispatch(fetchManifest());
  }, [dispatch]);

  useEffect(() => {
    if (!runId) return;
    const saved = loadReviewFromStorage(runId);
    if (saved) {
      dispatch(loadReviewState({ runId: saved.runId, cards: saved.cards }));
    }
  }, [runId, dispatch]);

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
      <Header />
      <main className="max-w-[1400px] mx-auto px-6 py-5 grid grid-cols-[1fr_360px] gap-6">
        <CardList />
        <Sidebar />
      </main>
    </div>
  );
}
