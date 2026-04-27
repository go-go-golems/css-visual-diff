import { useState, useRef, useEffect, useCallback, useMemo } from "react";
import {
  Layers,
  SplitSquareHorizontal,
  Square,
  ArrowLeftRight,
  MessageCircle,
  Code2,
  FileText,
  Sparkles,
  Pin,
  Trash2,
  Pencil,
  X,
  Copy,
  AlertCircle,
  StickyNote,
  HelpCircle,
  Heart,
  Check,
  GripVertical,
  Zap,
  ChevronDown,
  ChevronRight,
  Camera,
  CornerDownRight,
} from "lucide-react";

/* ============================================================
   MOCK DATA — what the YAML/compare.json gives us
   ============================================================ */

const compareData = {
  page: "book",
  section: "content",
  classification: "review",
  changedPercent: 5.9817526469925655,
  bounds: {
    changed: true,
    delta: { height: 88.5, width: 0, x: 0, y: 0 },
    left: { height: 876.34375, width: 920, x: 0, y: 61 },
    right: { height: 964.84375, width: 920, x: 0, y: 61 },
  },
  pixel: {
    changedPixels: 53106,
    totalPixels: 887800,
    threshold: 30,
  },
  left: {
    name: "book-content",
    selector: "[data-page='book']",
    url: "http://localhost:7070/standalone/public/book.html",
  },
  right: {
    name: "book-content",
    selector: "[data-page='book']",
    url: "http://localhost:6007/iframe.html?id=public-site-pages-book--desktop&viewMode=story",
  },
  changedStyles: [
    { name: "background-color", left: "rgba(0, 0, 0, 0)", right: "rgb(255, 255, 255)" },
    { name: "color", left: "rgb(31, 30, 28)", right: "rgb(26, 26, 24)" },
    { name: "font-family", left: "Inter, sans-serif", right: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' },
    { name: "font-size", left: "16px", right: "14px" },
    { name: "line-height", left: "normal", right: "21px" },
    { name: "padding-bottom", left: "0px", right: "72px" },
    { name: "padding-left", left: "32px", right: "0px" },
    { name: "padding-right", left: "32px", right: "0px" },
    { name: "padding-top", left: "40px", right: "0px" },
  ],
  changedAttributes: [
    { name: "class", left: "", right: "pyxis-public-page pyxis-book-page" },
  ],
};

/* ============================================================
   COMMENT TYPES
   ============================================================ */

const COMMENT_TYPES = {
  issue: { label: "issue", color: "#B91C1C", icon: AlertCircle },
  note: { label: "note", color: "#B68545", icon: StickyNote },
  question: { label: "question", color: "#2D6F62", icon: HelpCircle },
  praise: { label: "praise", color: "#714464", icon: Heart },
};

/* ============================================================
   VIEW MODES
   ============================================================ */

const VIEW_MODES = [
  { id: "side-by-side", label: "Side-by-side", icon: SplitSquareHorizontal, hint: "Compare A and B next to each other" },
  { id: "overlay", label: "Overlay", icon: Layers, hint: "Stack with opacity — hold ␣ to flash B" },
  { id: "slider", label: "Slider", icon: ArrowLeftRight, hint: "Drag to reveal B over A" },
  { id: "diff", label: "Diff only", icon: Square, hint: "Just the pixels that changed" },
];

/* ============================================================
   THE MOCK FORM (stand-in for the real screenshot)
   Two variants render with the *actual* CSS deltas from the diff:
   font-size, padding, color, line-height — so overlay/slider modes
   actually demonstrate misalignment.
   ============================================================ */

const Pill = ({ children, active, style }) => (
  <button
    style={style}
    className={`px-2.5 py-1 rounded-[3px] border transition-colors ${
      active
        ? "bg-[#1A1815] text-white border-[#1A1815]"
        : "bg-white text-[#1A1815] border-[#D8D3C5]"
    }`}
  >
    {children}
  </button>
);

const Field = ({ label, placeholder, sizing }) => (
  <div>
    <div
      className="uppercase font-semibold text-[#6B6862] mb-1"
      style={{ fontSize: 10, letterSpacing: "0.14em" }}
    >
      {label}
    </div>
    <input
      type="text"
      placeholder={placeholder}
      className="w-full px-2.5 py-1.5 border border-[#D8D3C5] rounded-[3px] bg-white text-[#6B6862] placeholder:text-[#A8A39A] outline-none"
      style={{ fontSize: sizing.fontSize, lineHeight: sizing.lineHeight }}
    />
  </div>
);

const BookForm = ({ variant, tinted }) => {
  // Variant-specific styling, derived directly from compareData.changedStyles
  const isReact = variant === "react";

  const sizing = {
    fontSize: isReact ? "14px" : "16px",
    lineHeight: isReact ? "21px" : "normal",
    color: isReact ? "rgb(26, 26, 24)" : "rgb(31, 30, 28)",
    paddingTop: isReact ? "0px" : "26px", // scaled down from 40
    paddingRight: isReact ? "0px" : "20px", // scaled down from 32
    paddingBottom: isReact ? "46px" : "0px", // scaled down from 72
    paddingLeft: isReact ? "0px" : "20px",
    backgroundColor: isReact ? "#FFFFFF" : "transparent",
    fontFamily: isReact
      ? '"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
      : "Inter, sans-serif",
  };

  const tintFilter = tinted === "left" ? "hue-rotate(170deg) saturate(1.4)" : tinted === "right" ? "hue-rotate(-30deg) saturate(1.4)" : "none";

  return (
    <div
      style={{
        ...sizing,
        filter: tintFilter,
        width: 480,
      }}
    >
      <div className="flex gap-5 items-start">
        <div className="flex-1 min-w-0">
          <div
            className="uppercase font-semibold text-[#6B6862] mb-1"
            style={{ fontSize: 10, letterSpacing: "0.18em" }}
          >
            Inquiries
          </div>
          <h1
            className="font-bold tracking-tight mb-2"
            style={{
              fontFamily: '"Fraunces", "Tiempos Headline", Georgia, serif',
              fontSize: 28,
              lineHeight: 1.05,
              color: sizing.color,
            }}
          >
            Book the space
          </h1>
          <p className="text-[#6B6862] mb-4 max-w-md">
            tell us about your show. we read every submission. responses in 3–7 days. we book 6–10 weeks out; late requests get the unused-dates list.
          </p>
          <div className="space-y-2.5">
            <Field label="Your name" sizing={sizing} />
            <Field label="Email" placeholder="you@label.com" sizing={sizing} />
            <Field label="Project / artist name" sizing={sizing} />
            <div className="grid grid-cols-2 gap-2.5">
              <Field label="Preferred date" placeholder="e.g. late April" sizing={sizing} />
              <Field label="Expected draw" placeholder="Under 50" sizing={sizing} />
            </div>
            <div>
              <div
                className="uppercase font-semibold text-[#6B6862] mb-1"
                style={{ fontSize: 10, letterSpacing: "0.14em" }}
              >
                Show type
              </div>
              <div className="flex flex-wrap gap-1.5" style={{ fontSize: sizing.fontSize }}>
                <Pill style={{ fontSize: 12 }}>DJ night</Pill>
                <Pill active style={{ fontSize: 12 }}>Live music</Pill>
                <Pill style={{ fontSize: 12 }}>Listening party</Pill>
                <Pill style={{ fontSize: 12 }}>Workshop / meet-up</Pill>
                <Pill style={{ fontSize: 12 }}>Screening</Pill>
                <Pill style={{ fontSize: 12 }}>Other</Pill>
              </div>
            </div>
            <div>
              <div
                className="uppercase font-semibold text-[#6B6862] mb-1"
                style={{ fontSize: 10, letterSpacing: "0.14em" }}
              >
                Tell us about it
              </div>
              <textarea
                className="w-full h-14 px-2.5 py-1.5 border border-[#D8D3C5] rounded-[3px] bg-white text-[#6B6862] resize-none outline-none"
                style={{ fontSize: sizing.fontSize, lineHeight: sizing.lineHeight, fontFamily: "inherit" }}
                defaultValue="who's on the bill, what it sounds like, what you need from us"
              />
            </div>
            <div className="flex items-start gap-2 pt-1 text-[#1A1815]">
              <input type="checkbox" className="mt-0.5 accent-[#B91C1C]" defaultChecked />
              <span style={{ fontSize: sizing.fontSize, lineHeight: sizing.lineHeight }}>
                I've read the <a className="underline">safer-space policy</a> and agree to uphold it for my show.
              </span>
            </div>
            <div>
              <button className="bg-[#B91C1C] text-white px-3.5 py-1.5 rounded-[3px] font-medium" style={{ fontSize: 13 }}>
                Send inquiry →
              </button>
            </div>
          </div>
        </div>

        <aside
          className="bg-[#1A1815] text-white p-4 self-start"
          style={{ width: 168, fontSize: 11, lineHeight: 1.45 }}
        >
          <div
            className="italic mb-3"
            style={{ fontFamily: '"Fraunces", Georgia, serif', fontSize: 13 }}
          >
            the space
          </div>
          <dl className="space-y-2.5">
            {[
              ["Capacity", "150 standing · 80 seated"],
              ["PA", "Funktion-One F1201, 4-way · 2× Sub infra 108"],
              ["Backline", "CDJ-3000 ×2, DJM-900, Moog Sub37, house drum kit"],
              ["Tech", "projector, haze, moving heads ×4, basic video chain"],
              ["Hours", "close by 2 AM (3 on Sat)"],
            ].map(([k, v]) => (
              <div key={k}>
                <dt
                  className="uppercase font-semibold text-[#A39E92] mb-0.5"
                  style={{ fontSize: 9, letterSpacing: "0.16em" }}
                >
                  {k}
                </dt>
                <dd className="text-[#E8E4D9]">{v}</dd>
              </div>
            ))}
          </dl>
          <hr className="border-[#3A372F] my-3" />
          <div className="text-[#A39E92]" style={{ fontSize: 10 }}>25 Manton Ave · Providence RI 02909</div>
          <div className="text-[#A39E92]" style={{ fontSize: 10 }}>book@pyxis.space</div>
        </aside>
      </div>
    </div>
  );
};

/* ============================================================
   COMMENT PIN
   ============================================================ */

const CommentPin = ({ comment, index, active, onClick }) => {
  const meta = COMMENT_TYPES[comment.type];
  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
      className="absolute -translate-x-1/2 -translate-y-1/2 z-30"
      style={{ left: `${comment.x}%`, top: `${comment.y}%` }}
    >
      <span
        className={`flex items-center justify-center w-7 h-7 rounded-full text-white font-semibold text-xs shadow-[0_2px_8px_rgba(0,0,0,0.25)] ring-2 transition-all ${
          active ? "scale-110 ring-white" : "ring-white/80 hover:scale-105"
        }`}
        style={{
          background: meta.color,
          fontFamily: '"Geist", system-ui, sans-serif',
          outline: active ? `2px solid ${meta.color}` : "none",
          outlineOffset: "2px",
        }}
      >
        {index + 1}
      </span>
    </button>
  );
};

/* ============================================================
   CANVAS — handles click-to-comment for any view mode
   ============================================================ */

const Canvas = ({ children, commentMode, onAddComment, comments, activeId, setActiveId, side, className = "" }) => {
  const ref = useRef(null);

  const handleClick = (e) => {
    if (!commentMode) return;
    const rect = ref.current.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * 100;
    const y = ((e.clientY - rect.top) / rect.height) * 100;
    onAddComment({ x, y, side });
  };

  return (
    <div
      ref={ref}
      onClick={handleClick}
      className={`relative bg-white border border-[#E0DCD0] rounded-[2px] overflow-hidden ${
        commentMode ? "cursor-crosshair" : ""
      } ${className}`}
    >
      <div className="relative">{children}</div>
      {comments
        .filter((c) => c.side === side || (side === "merged" && (c.side === "both" || c.side === "left" || c.side === "right")))
        .map((c, i) => (
          <CommentPin
            key={c.id}
            comment={c}
            index={comments.findIndex((x) => x.id === c.id)}
            active={activeId === c.id}
            onClick={() => setActiveId(c.id)}
          />
        ))}
    </div>
  );
};

/* ============================================================
   MAIN APP
   ============================================================ */

export default function App() {
  /* ---------- state ---------- */
  const [viewMode, setViewMode] = useState("side-by-side");
  const [commentMode, setCommentMode] = useState(false);
  const [draftType, setDraftType] = useState("note");

  const [comments, setComments] = useState([
    {
      id: 1,
      x: 18,
      y: 13,
      side: "right",
      type: "issue",
      text: "Body copy under the heading shrank 16→14px. Looks anemic next to the display serif.",
    },
    {
      id: 2,
      x: 62,
      y: 49,
      side: "right",
      type: "note",
      text: "Pills now read at 12px regardless. Worth checking the 'Workshop / meet-up' label doesn't break to two lines on narrower viewports.",
    },
    {
      id: 3,
      x: 30,
      y: 87,
      side: "merged",
      type: "question",
      text: "Bottom padding 0→72px — is this from a page-level wrapper (.pyxis-public-page) or genuinely the section?",
    },
  ]);
  const [activeId, setActiveId] = useState(null);

  // Overlay
  const [overlayOpacity, setOverlayOpacity] = useState(50);
  const [overlayBlend, setOverlayBlend] = useState("normal");
  const [swapPressed, setSwapPressed] = useState(false);

  // Slider
  const [sliderPos, setSliderPos] = useState(50);
  const [draggingSlider, setDraggingSlider] = useState(false);

  // Sidebar / export
  const [sidebarTab, setSidebarTab] = useState("comments");
  const [generalNote, setGeneralNote] = useState(
    "vertical whitespace seems to always be a little bit too much below separators or saw in the react version, and it adds up as the page goes further down."
  );
  const [exportOpen, setExportOpen] = useState(false);
  const [includeImage, setIncludeImage] = useState(true);
  const [copied, setCopied] = useState(false);

  /* ---------- spacebar swap for overlay ---------- */
  useEffect(() => {
    const down = (e) => {
      if (e.code === "Space" && !e.repeat && viewMode === "overlay" && document.activeElement.tagName !== "TEXTAREA" && document.activeElement.tagName !== "INPUT") {
        e.preventDefault();
        setSwapPressed(true);
      }
    };
    const up = (e) => {
      if (e.code === "Space") setSwapPressed(false);
    };
    window.addEventListener("keydown", down);
    window.addEventListener("keyup", up);
    return () => {
      window.removeEventListener("keydown", down);
      window.removeEventListener("keyup", up);
    };
  }, [viewMode]);

  /* ---------- slider drag ---------- */
  const sliderRef = useRef(null);
  useEffect(() => {
    if (!draggingSlider) return;
    const move = (e) => {
      if (!sliderRef.current) return;
      const rect = sliderRef.current.getBoundingClientRect();
      const pos = ((e.clientX - rect.left) / rect.width) * 100;
      setSliderPos(Math.max(0, Math.min(100, pos)));
    };
    const up = () => setDraggingSlider(false);
    window.addEventListener("mousemove", move);
    window.addEventListener("mouseup", up);
    return () => {
      window.removeEventListener("mousemove", move);
      window.removeEventListener("mouseup", up);
    };
  }, [draggingSlider]);

  /* ---------- comment ops ---------- */
  const addComment = ({ x, y, side }) => {
    const newComment = {
      id: Date.now(),
      x,
      y,
      side,
      type: draftType,
      text: "",
    };
    setComments([...comments, newComment]);
    setActiveId(newComment.id);
    setCommentMode(false);
  };

  const updateComment = (id, patch) => {
    setComments(comments.map((c) => (c.id === id ? { ...c, ...patch } : c)));
  };

  const deleteComment = (id) => {
    setComments(comments.filter((c) => c.id !== id));
    if (activeId === id) setActiveId(null);
  };

  /* ---------- effective overlay opacity (with swap) ---------- */
  const effectiveOpacity = swapPressed ? (overlayOpacity > 50 ? 0 : 100) : overlayOpacity;

  /* ---------- export markdown ---------- */
  const exportMarkdown = useMemo(() => {
    const cd = compareData;
    const fmtPct = (n) => `${n.toFixed(2)}%`;

    const styleRows = cd.changedStyles
      .map((s) => `  - { name: ${s.name.padEnd(18)}, left: ${JSON.stringify(s.left)}, right: ${JSON.stringify(s.right)} }`)
      .join("\n");

    const commentRows = comments
      .map((c, i) => {
        const sideLabel = c.side === "left" ? "prototype" : c.side === "right" ? "react" : "merged";
        const text = (c.text || "(no text)").replace(/\n/g, " ");
        return `${i + 1}. **[${COMMENT_TYPES[c.type].label}]** _${sideLabel} @ ${c.x.toFixed(0)}%, ${c.y.toFixed(0)}%_ — ${text}`;
      })
      .join("\n");

    const yamlComments = comments
      .map(
        (c, i) => `  - id: ${i + 1}
    type: ${c.type}
    side: ${c.side}
    position: { x: ${c.x.toFixed(2)}, y: ${c.y.toFixed(2)} }
    text: ${JSON.stringify(c.text || "")}`
      )
      .join("\n");

    return `# Visual review — \`${cd.page} / ${cd.section}\`

**Classification:** ${cd.classification} · ${fmtPct(cd.changedPercent)} changed (${cd.pixel.changedPixels.toLocaleString()} / ${cd.pixel.totalPixels.toLocaleString()} px)
**Bounds delta:** Δheight ${cd.bounds.delta.height}px · Δwidth ${cd.bounds.delta.width}px${includeImage ? "\n**Annotated screenshot:** _attached_" : ""}

## General observation

${generalNote}

## Pin-drop comments

${commentRows || "_(none yet)_"}

## Computed style diffs

${cd.changedStyles
  .map((s) => `- \`${s.name}\`: \`${s.left}\` → \`${s.right}\``)
  .join("\n")}

## Source

- prototype: ${cd.left.url}
- react: ${cd.right.url}

---

\`\`\`yaml
page: ${cd.page}
section: ${cd.section}
classification: ${cd.classification}
changedPercent: ${cd.changedPercent}
bounds:
  delta: { height: ${cd.bounds.delta.height}, width: ${cd.bounds.delta.width} }
  left:  { width: ${cd.bounds.left.width}, height: ${cd.bounds.left.height} }
  right: { width: ${cd.bounds.right.width}, height: ${cd.bounds.right.height} }
pixel:
  changedPixels: ${cd.pixel.changedPixels}
  totalPixels: ${cd.pixel.totalPixels}
  threshold: ${cd.pixel.threshold}
changedStyles:
${styleRows}
review:
  generalNote: ${JSON.stringify(generalNote)}
  comments:
${yamlComments || "    []"}
\`\`\`
`;
  }, [comments, generalNote, includeImage]);

  /* ---------- rendering view modes ---------- */

  const renderCanvas = () => {
    if (viewMode === "side-by-side") {
      return (
        <div className="flex gap-6">
          <div className="flex-1">
            <CanvasLabel side="left" url={compareData.left.url} />
            <Canvas
              commentMode={commentMode}
              onAddComment={addComment}
              comments={comments}
              activeId={activeId}
              setActiveId={setActiveId}
              side="left"
              className="p-5"
            >
              <BookForm variant="prototype" />
            </Canvas>
          </div>
          <div className="flex-1">
            <CanvasLabel side="right" url={compareData.right.url} />
            <Canvas
              commentMode={commentMode}
              onAddComment={addComment}
              comments={comments}
              activeId={activeId}
              setActiveId={setActiveId}
              side="right"
              className="p-5"
            >
              <BookForm variant="react" />
            </Canvas>
          </div>
        </div>
      );
    }

    if (viewMode === "overlay") {
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-3 px-3 py-2 bg-[#FFFDF7] border border-[#E0DCD0] rounded-[2px]">
            <span className="text-xs text-[#6B6862] uppercase tracking-wider">A</span>
            <input
              type="range"
              min="0"
              max="100"
              value={effectiveOpacity}
              onChange={(e) => setOverlayOpacity(parseInt(e.target.value))}
              className="flex-1 accent-[#1A1815]"
            />
            <span className="text-xs text-[#6B6862] uppercase tracking-wider">B</span>
            <div className="w-px h-5 bg-[#E0DCD0]" />
            <button
              onClick={() => setOverlayBlend(overlayBlend === "normal" ? "difference" : "normal")}
              className={`px-2 py-1 text-xs rounded-[3px] flex items-center gap-1 ${
                overlayBlend === "difference"
                  ? "bg-[#1A1815] text-white"
                  : "bg-white border border-[#D8D3C5] text-[#1A1815]"
              }`}
            >
              <Zap size={12} /> diff blend
            </button>
            <span className="text-[10px] text-[#A8A39A] font-mono ml-auto">hold ␣ to flash B</span>
          </div>
          <Canvas
            commentMode={commentMode}
            onAddComment={addComment}
            comments={comments}
            activeId={activeId}
            setActiveId={setActiveId}
            side="merged"
            className="p-5 mx-auto"
          >
            <div className="relative">
              <BookForm variant="prototype" />
              <div
                className="absolute inset-0"
                style={{
                  opacity: effectiveOpacity / 100,
                  mixBlendMode: overlayBlend === "difference" ? "difference" : "normal",
                  pointerEvents: "none",
                }}
              >
                <BookForm variant="react" />
              </div>
            </div>
          </Canvas>
        </div>
      );
    }

    if (viewMode === "slider") {
      return (
        <div>
          <div className="flex items-center gap-3 px-3 py-2 bg-[#FFFDF7] border border-[#E0DCD0] rounded-[2px] mb-3 text-xs text-[#6B6862]">
            <CornerDownRight size={12} /> drag the handle to sweep
          </div>
          <Canvas
            commentMode={commentMode}
            onAddComment={addComment}
            comments={comments}
            activeId={activeId}
            setActiveId={setActiveId}
            side="merged"
            className="p-5"
          >
            <div ref={sliderRef} className="relative select-none">
              <div style={{ clipPath: `inset(0 ${100 - sliderPos}% 0 0)` }}>
                <BookForm variant="prototype" />
              </div>
              <div
                className="absolute inset-0"
                style={{ clipPath: `inset(0 0 0 ${sliderPos}%)`, pointerEvents: "none" }}
              >
                <BookForm variant="react" />
              </div>
              {/* Slider handle */}
              <div
                className="absolute top-0 bottom-0 z-20"
                style={{ left: `${sliderPos}%`, transform: "translateX(-50%)" }}
              >
                <div className="w-px h-full bg-[#1A1815]/40" />
                <button
                  onMouseDown={(e) => {
                    e.stopPropagation();
                    setDraggingSlider(true);
                  }}
                  className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 left-1/2 w-7 h-7 rounded-full bg-[#1A1815] text-white flex items-center justify-center shadow-[0_4px_12px_rgba(0,0,0,0.3)] cursor-ew-resize"
                >
                  <GripVertical size={14} />
                </button>
              </div>
            </div>
          </Canvas>
        </div>
      );
    }

    if (viewMode === "diff") {
      return (
        <div>
          <div className="flex items-center gap-3 px-3 py-2 bg-[#FFFDF7] border border-[#E0DCD0] rounded-[2px] mb-3 text-xs text-[#6B6862]">
            <Square size={12} /> only the pixels that differ — derived from <code className="bg-[#F0EDE3] px-1 rounded font-mono">diff_only.png</code>
          </div>
          <Canvas
            commentMode={commentMode}
            onAddComment={addComment}
            comments={comments}
            activeId={activeId}
            setActiveId={setActiveId}
            side="merged"
            className="p-5"
          >
            <div
              className="relative"
              style={{
                filter: "url(#diff-tint)",
              }}
            >
              <svg width="0" height="0" style={{ position: "absolute" }}>
                <filter id="diff-tint">
                  <feColorMatrix
                    type="matrix"
                    values="0 0 0 0 0.73
                            0 0 0 0 0.11
                            0 0 0 0 0.11
                            0 0 0 1 0"
                  />
                </filter>
              </svg>
              <BookForm variant="react" />
            </div>
          </Canvas>
        </div>
      );
    }
  };

  /* ---------- render ---------- */
  return (
    <div className="min-h-screen bg-[#F5F2EA] text-[#1A1815]" style={{ fontFamily: '"Geist", system-ui, sans-serif' }}>
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,400;9..144,500;9..144,600;9..144,700;9..144,800&family=Geist:wght@400;500;600;700&family=Geist+Mono:wght@400;500&family=Inter:wght@400;500;600;700&display=swap');
        body { font-family: "Geist", system-ui, sans-serif; }
        .font-display { font-family: "Fraunces", "Tiempos Headline", Georgia, serif; }
        .font-mono-x { font-family: "Geist Mono", ui-monospace, monospace; }
      `}</style>

      {/* ============== HEADER ============== */}
      <header className="border-b border-[#E0DCD0] bg-[#F5F2EA]/90 backdrop-blur-sm sticky top-0 z-40">
        <div className="max-w-[1400px] mx-auto px-6 py-3 flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-[2px] bg-[#1A1815] flex items-center justify-center">
              <span className="font-display text-white italic" style={{ fontSize: 16 }}>p</span>
            </div>
            <span className="font-display text-[18px] tracking-tight">pyxis review</span>
          </div>

          <div className="h-5 w-px bg-[#E0DCD0]" />

          <div className="flex items-center gap-1.5 text-sm">
            <span className="text-[#6B6862]">page</span>
            <span className="font-mono-x font-medium">{compareData.page}</span>
            <span className="text-[#A8A39A]">/</span>
            <span className="font-mono-x font-medium">{compareData.section}</span>
          </div>

          <div className="flex items-center gap-2 ml-2">
            <span className="px-2 py-0.5 text-[11px] uppercase tracking-wider rounded-[2px] bg-[#FBE9E9] text-[#B91C1C] font-semibold">
              {compareData.classification}
            </span>
            <span className="font-mono-x text-xs text-[#6B6862]">
              {compareData.changedPercent.toFixed(2)}% changed
            </span>
          </div>

          <div className="ml-auto flex items-center gap-2">
            <button
              onClick={() => setExportOpen(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-[3px] bg-[#1A1815] text-white text-sm font-medium hover:bg-black"
            >
              <Sparkles size={14} /> Send to LLM
            </button>
          </div>
        </div>

        {/* ============== TOOLBAR ============== */}
        <div className="border-t border-[#E0DCD0]">
          <div className="max-w-[1400px] mx-auto px-6 py-2 flex items-center gap-2">
            {/* View modes */}
            <div className="flex items-center bg-white border border-[#E0DCD0] rounded-[3px] p-0.5">
              {VIEW_MODES.map((m) => {
                const Icon = m.icon;
                const active = viewMode === m.id;
                return (
                  <button
                    key={m.id}
                    onClick={() => setViewMode(m.id)}
                    title={m.hint}
                    className={`flex items-center gap-1.5 px-2.5 py-1 rounded-[2px] text-xs font-medium transition-colors ${
                      active ? "bg-[#1A1815] text-white" : "text-[#1A1815] hover:bg-[#F0EDE3]"
                    }`}
                  >
                    <Icon size={13} /> {m.label}
                  </button>
                );
              })}
            </div>

            <div className="h-5 w-px bg-[#E0DCD0]" />

            {/* Comment mode */}
            <button
              onClick={() => setCommentMode(!commentMode)}
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-[3px] text-xs font-medium transition-colors ${
                commentMode
                  ? "bg-[#B91C1C] text-white"
                  : "bg-white border border-[#E0DCD0] text-[#1A1815] hover:bg-[#F0EDE3]"
              }`}
            >
              <Pin size={13} /> {commentMode ? "click to drop pin…" : "Add comment"}
            </button>

            {commentMode && (
              <div className="flex items-center gap-1 ml-1 animate-[fadeIn_120ms_ease-out]">
                {Object.entries(COMMENT_TYPES).map(([key, t]) => {
                  const Icon = t.icon;
                  const sel = draftType === key;
                  return (
                    <button
                      key={key}
                      onClick={() => setDraftType(key)}
                      title={t.label}
                      className={`w-6 h-6 rounded-full flex items-center justify-center transition-transform ${
                        sel ? "scale-110 ring-2 ring-offset-1 ring-[#1A1815]" : ""
                      }`}
                      style={{ background: t.color, color: "#fff" }}
                    >
                      <Icon size={12} />
                    </button>
                  );
                })}
              </div>
            )}

            <div className="ml-auto flex items-center gap-3 text-xs text-[#6B6862]">
              <span className="font-mono-x">
                Δ {compareData.bounds.delta.height}×{compareData.bounds.delta.width}
              </span>
              <span className="font-mono-x">
                {compareData.pixel.changedPixels.toLocaleString()} / {compareData.pixel.totalPixels.toLocaleString()} px
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* ============== MAIN ============== */}
      <main className="max-w-[1400px] mx-auto px-6 py-5 grid grid-cols-[1fr_360px] gap-6">
        {/* CANVAS COLUMN */}
        <div>
          {renderCanvas()}

          {/* General note */}
          <div className="mt-5 bg-white border border-[#E0DCD0] rounded-[2px] p-4">
            <div className="flex items-center gap-2 mb-2">
              <FileText size={13} className="text-[#6B6862]" />
              <span className="text-xs uppercase tracking-wider font-semibold text-[#6B6862]">
                General observation
              </span>
            </div>
            <textarea
              value={generalNote}
              onChange={(e) => setGeneralNote(e.target.value)}
              className="w-full text-sm bg-transparent outline-none resize-none text-[#1A1815] placeholder:text-[#A8A39A]"
              rows={2}
              placeholder="Anything that applies to the whole page…"
            />
          </div>
        </div>

        {/* SIDEBAR */}
        <aside className="bg-white border border-[#E0DCD0] rounded-[2px] flex flex-col" style={{ height: "calc(100vh - 180px)", position: "sticky", top: 130 }}>
          {/* tabs */}
          <div className="flex border-b border-[#E0DCD0]">
            {[
              { id: "comments", icon: MessageCircle, label: "Comments", count: comments.length },
              { id: "styles", icon: Code2, label: "CSS diff", count: compareData.changedStyles.length },
              { id: "meta", icon: FileText, label: "Meta" },
            ].map((t) => {
              const Icon = t.icon;
              const active = sidebarTab === t.id;
              return (
                <button
                  key={t.id}
                  onClick={() => setSidebarTab(t.id)}
                  className={`flex-1 px-3 py-2.5 text-xs font-medium flex items-center justify-center gap-1.5 transition-colors border-b-2 ${
                    active
                      ? "border-[#1A1815] text-[#1A1815]"
                      : "border-transparent text-[#6B6862] hover:text-[#1A1815]"
                  }`}
                >
                  <Icon size={13} /> {t.label}
                  {t.count != null && (
                    <span className={`text-[10px] font-mono-x px-1.5 rounded-full ${active ? "bg-[#1A1815] text-white" : "bg-[#F0EDE3] text-[#6B6862]"}`}>
                      {t.count}
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {/* tab content */}
          <div className="flex-1 overflow-y-auto">
            {sidebarTab === "comments" && (
              <CommentsTab
                comments={comments}
                activeId={activeId}
                setActiveId={setActiveId}
                updateComment={updateComment}
                deleteComment={deleteComment}
              />
            )}
            {sidebarTab === "styles" && <StylesTab />}
            {sidebarTab === "meta" && <MetaTab />}
          </div>
        </aside>
      </main>

      {/* ============== EXPORT MODAL ============== */}
      {exportOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-6">
          <div className="bg-[#F5F2EA] w-full max-w-3xl max-h-[85vh] rounded-[2px] shadow-2xl border border-[#E0DCD0] flex flex-col">
            <div className="flex items-center justify-between px-5 py-3 border-b border-[#E0DCD0]">
              <div className="flex items-center gap-2">
                <Sparkles size={16} />
                <span className="font-display text-lg">LLM handoff</span>
                <span className="text-xs text-[#6B6862] font-mono-x ml-2">markdown + yaml</span>
              </div>
              <button onClick={() => setExportOpen(false)} className="text-[#6B6862] hover:text-[#1A1815]">
                <X size={18} />
              </button>
            </div>

            <div className="px-5 py-3 border-b border-[#E0DCD0] flex items-center gap-4">
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input
                  type="checkbox"
                  checked={includeImage}
                  onChange={(e) => setIncludeImage(e.target.checked)}
                  className="accent-[#1A1815]"
                />
                <Camera size={13} /> attach annotated screenshot
              </label>
              <span className="text-xs text-[#A8A39A] font-mono-x ml-auto">
                {comments.length} pin{comments.length === 1 ? "" : "s"} · {compareData.changedStyles.length} style change{compareData.changedStyles.length === 1 ? "" : "s"}
              </span>
            </div>

            <div className="flex-1 overflow-auto p-5">
              <pre className="font-mono-x text-[12px] leading-relaxed text-[#1A1815] whitespace-pre-wrap bg-white border border-[#E0DCD0] rounded-[2px] p-4">
                {exportMarkdown}
              </pre>
            </div>

            <div className="border-t border-[#E0DCD0] px-5 py-3 flex items-center gap-2 justify-end">
              <button
                onClick={() => setExportOpen(false)}
                className="px-3 py-1.5 text-sm rounded-[3px] border border-[#E0DCD0] hover:bg-white"
              >
                Cancel
              </button>
              <button
                onClick={() => {
                  navigator.clipboard.writeText(exportMarkdown);
                  setCopied(true);
                  setTimeout(() => setCopied(false), 1600);
                }}
                className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-[3px] bg-[#1A1815] text-white font-medium hover:bg-black"
              >
                {copied ? <Check size={14} /> : <Copy size={14} />}
                {copied ? "Copied" : "Copy markdown"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

/* ============================================================
   SIDEBAR TABS
   ============================================================ */

const CommentsTab = ({ comments, activeId, setActiveId, updateComment, deleteComment }) => {
  if (comments.length === 0) {
    return (
      <div className="p-6 text-center text-[#6B6862] text-sm">
        <Pin size={18} className="mx-auto mb-2 opacity-50" />
        No pins yet. Click <span className="font-medium">Add comment</span> then click on the canvas.
      </div>
    );
  }

  return (
    <div className="divide-y divide-[#E0DCD0]">
      {comments.map((c, i) => {
        const meta = COMMENT_TYPES[c.type];
        const Icon = meta.icon;
        const active = activeId === c.id;
        return (
          <div
            key={c.id}
            className={`p-3 cursor-pointer transition-colors ${active ? "bg-[#FFFDF7]" : "hover:bg-[#FBF8F0]"}`}
            onClick={() => setActiveId(c.id)}
          >
            <div className="flex items-start gap-2">
              <span
                className="flex items-center justify-center w-6 h-6 rounded-full text-white text-xs font-semibold shrink-0 font-mono-x"
                style={{ background: meta.color }}
              >
                {i + 1}
              </span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <select
                    value={c.type}
                    onClick={(e) => e.stopPropagation()}
                    onChange={(e) => updateComment(c.id, { type: e.target.value })}
                    className="text-[11px] uppercase tracking-wider font-semibold bg-transparent outline-none cursor-pointer"
                    style={{ color: meta.color }}
                  >
                    {Object.entries(COMMENT_TYPES).map(([k, t]) => (
                      <option key={k} value={k}>{t.label}</option>
                    ))}
                  </select>
                  <span className="text-[10px] text-[#A8A39A] font-mono-x">
                    {c.side === "left" ? "prototype" : c.side === "right" ? "react" : "merged"} · {c.x.toFixed(0)},{c.y.toFixed(0)}
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteComment(c.id);
                    }}
                    className="ml-auto text-[#A8A39A] hover:text-[#B91C1C]"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
                <textarea
                  value={c.text}
                  onClick={(e) => e.stopPropagation()}
                  onChange={(e) => updateComment(c.id, { text: e.target.value })}
                  placeholder="What's wrong here?"
                  className="w-full text-sm bg-transparent outline-none resize-none text-[#1A1815] placeholder:text-[#A8A39A]"
                  rows={c.text.length > 80 ? 3 : 2}
                />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};

const StylesTab = () => {
  const [expanded, setExpanded] = useState(true);
  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-4 py-2.5 text-left text-xs uppercase tracking-wider font-semibold text-[#6B6862] flex items-center gap-1.5 border-b border-[#E0DCD0] hover:bg-[#FBF8F0]"
      >
        {expanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        computed style diffs ({compareData.changedStyles.length})
      </button>
      {expanded && (
        <div className="divide-y divide-[#F0EDE3]">
          {compareData.changedStyles.map((s) => (
            <div key={s.name} className="px-4 py-2.5">
              <div className="font-mono-x text-[11px] font-semibold text-[#1A1815] mb-1">
                {s.name}
              </div>
              <div className="space-y-0.5">
                <div className="flex items-start gap-2 font-mono-x text-[11px]">
                  <span className="text-[#A8A39A] shrink-0">L</span>
                  <span className="text-[#6B6862] break-all">{s.left}</span>
                </div>
                <div className="flex items-start gap-2 font-mono-x text-[11px]">
                  <span className="text-[#B91C1C] shrink-0">R</span>
                  <span className="text-[#1A1815] break-all">{s.right}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
      <div className="px-4 py-2.5 text-xs uppercase tracking-wider font-semibold text-[#6B6862] border-t border-b border-[#E0DCD0]">
        attribute diffs ({compareData.changedAttributes.length})
      </div>
      {compareData.changedAttributes.map((a) => (
        <div key={a.name} className="px-4 py-2.5">
          <div className="font-mono-x text-[11px] font-semibold mb-1">{a.name}</div>
          <div className="font-mono-x text-[11px] text-[#1A1815] break-all">
            <span className="text-[#A8A39A]">+</span> {a.right}
          </div>
        </div>
      ))}
    </div>
  );
};

const MetaTab = () => {
  return (
    <div className="p-4 space-y-4 text-xs">
      <Section title="Bounds">
        <KV k="prototype" v={`${compareData.bounds.left.width}×${compareData.bounds.left.height}`} />
        <KV k="react" v={`${compareData.bounds.right.width}×${compareData.bounds.right.height}`} />
        <KV k="delta" v={`+${compareData.bounds.delta.height}px tall`} highlight />
      </Section>
      <Section title="Pixel diff">
        <KV k="changed" v={compareData.pixel.changedPixels.toLocaleString()} />
        <KV k="total" v={compareData.pixel.totalPixels.toLocaleString()} />
        <KV k="threshold" v={compareData.pixel.threshold} />
      </Section>
      <Section title="Selectors">
        <div className="font-mono-x text-[11px] text-[#1A1815] break-all">{compareData.left.selector}</div>
      </Section>
      <Section title="Sources">
        <div className="space-y-1">
          <div>
            <div className="text-[10px] uppercase tracking-wider text-[#6B6862] mb-0.5">prototype</div>
            <a href={compareData.left.url} className="font-mono-x text-[11px] text-[#1A1815] break-all hover:underline">
              {compareData.left.url}
            </a>
          </div>
          <div className="pt-2">
            <div className="text-[10px] uppercase tracking-wider text-[#6B6862] mb-0.5">react</div>
            <a href={compareData.right.url} className="font-mono-x text-[11px] text-[#1A1815] break-all hover:underline">
              {compareData.right.url}
            </a>
          </div>
        </div>
      </Section>
      <Section title="Suggested YAML additions" muted>
        <ul className="space-y-1.5 text-[11px] text-[#6B6862] leading-relaxed">
          <li>· per-element bounds & selectors (so pins can attach to DOM nodes)</li>
          <li>· a11y deltas (aria attrs, tab order)</li>
          <li>· detected component regions (header, form-field-name…)</li>
          <li>· transition / animation diffs</li>
          <li>· responsive breakpoint rendering</li>
        </ul>
      </Section>
    </div>
  );
};

const Section = ({ title, children, muted }) => (
  <div className={muted ? "opacity-80" : ""}>
    <div className="text-[10px] uppercase tracking-wider font-semibold text-[#6B6862] mb-1.5">
      {title}
    </div>
    <div className="space-y-1">{children}</div>
  </div>
);

const KV = ({ k, v, highlight }) => (
  <div className="flex items-baseline justify-between gap-3">
    <span className="text-[#6B6862] font-mono-x text-[11px]">{k}</span>
    <span className={`font-mono-x text-[11px] ${highlight ? "text-[#B91C1C] font-semibold" : "text-[#1A1815]"}`}>
      {v}
    </span>
  </div>
);

/* ============================================================
   CANVAS LABEL (URL strip above each canvas in side-by-side)
   ============================================================ */

const CanvasLabel = ({ side, url }) => {
  const labelText = side === "left" ? "prototype" : "react";
  const tone =
    side === "left"
      ? "bg-[#FFFDF7] text-[#6B6862] border-[#E0DCD0]"
      : "bg-[#1A1815] text-[#FFFDF7] border-[#1A1815]";
  return (
    <div className={`flex items-center gap-2 px-3 py-1.5 text-xs border rounded-t-[2px] -mb-px ${tone}`}>
      <span className="uppercase tracking-wider font-semibold">{labelText}</span>
      <span className="font-mono-x text-[10px] truncate opacity-80">{url}</span>
    </div>
  );
};
