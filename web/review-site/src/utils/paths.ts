/**
 * Convert an absolute css-visual-diff artifact path to a relative URL
 * that the Go serve command can handle.
 *
 * /tmp/.../shows/artifacts/content/diff_only.png
 *   → /artifacts/shows/content/diff_only.png
 */
export function toArtifactUrl(absolutePath: string): string {
  const match = absolutePath.match(
    /\/([^/]+)\/artifacts\/([^/]+)\/([^/]+(?:\.[a-z]+)?)$/,
  );
  if (match) {
    return `/artifacts/${match[1]}/${match[2]}/${match[3]}`;
  }
  return absolutePath;
}

/**
 * Build the URL for a compare.json given page and section.
 */
export function compareJsonUrl(page: string, section: string): string {
  return `/api/compare?page=${encodeURIComponent(page)}&section=${encodeURIComponent(section)}`;
}
