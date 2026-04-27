# Tasks

## TODO

- [x] Add tasks here

- [x] Write diary with initial analysis and step-by-step plan
- [x] Write detailed design/implementation guide document
- [ ] Scaffold React+RTK+Vite app in web/review-site
- [ ] Port diff-review.jsx mock into real React components with RTK state
- [ ] Add Go embed support and 'serve' glazed verb to css-visual-diff
- [ ] Wire Vite build into Go binary via go:embed
- [ ] Implement data loading from css-visual-diff summary JSON + compare.json
- [x] Implement local storage persistence for review notes and status
- [x] Upload design doc to reMarkable
- [ ] Scaffold Vite+React+TS+Tailwind+RTK in web/review-site with package.json, vite.config, tailwind
- [ ] Create TypeScript types: summary.ts, compare.ts, review.ts
- [ ] Create RTK store: cardsSlice, viewSlice, reviewSlice, commentsSlice, localStorage middleware
- [ ] Build utility functions: paths.ts (artifact URL rewriting), export.ts (markdown+yaml), storage.ts
- [ ] Build layout components: App.tsx, Header.tsx, FilterToolbar.tsx, CardList.tsx
- [ ] Build view mode components: ViewModeSideBySide, ViewModeOverlay, ViewModeSlider, ViewModeDiff
- [ ] Build ReviewCard + ImageCanvas + CommentPin components
- [ ] Build Sidebar with CommentsTab, StylesTab, MetaTab
- [ ] Build ExportModal component
- [ ] Create Go internal/web package: embed.go, embed_none.go, static.go, generate.go
- [ ] Create cmd/build-web Dagger pipeline (with local fallback)
- [ ] Create serve subcommand in cmd/css-visual-diff/serve.go
- [ ] Wire serve into main.go, add Makefile targets, test end-to-end
- [ ] Commit, update diary, upload to reMarkable
