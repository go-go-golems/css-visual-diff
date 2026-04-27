import { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import type { AppDispatch, RootState } from "../store";
import { setSidebarTab } from "../store/slices/viewSlice";
import type { SidebarTab } from "../store/slices/viewSlice";
import type { CompareData } from "../types";
import { selectFilteredRows } from "../store/slices/cardsSlice";
import { MessageCircle, Code2, FileText } from "lucide-react";
import { CommentsTab } from "./CommentsTab";
import { StylesTab } from "./StylesTab";
import { MetaTab } from "./MetaTab";
import { compareJsonUrl } from "../utils/paths";

const TABS: { id: SidebarTab; icon: React.ElementType; label: string }[] = [
  { id: "comments", icon: MessageCircle, label: "Comments" },
  { id: "styles", icon: Code2, label: "CSS diff" },
  { id: "meta", icon: FileText, label: "Meta" },
];

export function Sidebar() {
  const dispatch = useDispatch<AppDispatch>();
  const activeTab = useSelector((s: RootState) => s.view.sidebarTab);
  const rows = useSelector(selectFilteredRows);
  const selectedIdx = useSelector((s: RootState) => s.cards.selectedCardIndex);
  const row = rows[selectedIdx] ?? rows[0] ?? null;
  const [compareData, setCompareData] = useState<CompareData | null>(null);
  const pins = useSelector((s: RootState) =>
    row ? s.comments.pins[`${row.page}/${row.section}`] ?? [] : [],
  );

  useEffect(() => {
    if (!row) { setCompareData(null); return; }
    fetch(compareJsonUrl(row.page, row.section))
      .then((r) => (r.ok ? r.json() : Promise.reject(r.status)))
      .then(setCompareData)
      .catch(() => setCompareData(null));
  }, [row]);

  if (!row) return null;

  return (
    <aside className="bg-white border border-border rounded-[2px] flex flex-col sticky top-[130px] max-h-[calc(100vh-180px)]">
      {/* Tabs */}
      <div className="flex border-b border-border">
        {TABS.map((t) => {
          const Icon = t.icon;
          const active = activeTab === t.id;
          return (
            <button
              key={t.id}
              onClick={() => dispatch(setSidebarTab(t.id))}
              className={`flex-1 px-3 py-2.5 text-xs font-medium flex items-center justify-center gap-1.5 transition-colors border-b-2 ${
                active
                  ? "border-text text-text"
                  : "border-transparent text-text-muted hover:text-text"
              }`}
            >
              <Icon size={13} /> {t.label}
              {t.id === "comments" && pins.length > 0 && (
                <span
                  className={`text-[10px] font-mono px-1.5 rounded-full ${
                    active ? "bg-text text-white" : "bg-cream-mid text-text-muted"
                  }`}
                >
                  {pins.length}
                </span>
              )}
              {t.id === "styles" && row.styleChangeCount > 0 && (
                <span
                  className={`text-[10px] font-mono px-1.5 rounded-full ${
                    active ? "bg-text text-white" : "bg-cream-mid text-text-muted"
                  }`}
                >
                  {row.styleChangeCount}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === "comments" && (
          <CommentsTab pins={pins} row={row} />
        )}
        {activeTab === "styles" && (
          <StylesTab row={row} compareData={compareData} />
        )}
        {activeTab === "meta" && (
          <MetaTab row={row} compareData={compareData} />
        )}
      </div>
    </aside>
  );
}
