/**
 * buildReactSrcDoc — turns a single React component's source into a full HTML
 * document string for a sandboxed iframe (`srcDoc`). Pure function (string in,
 * string out) so it can be unit-tested without a DOM.
 *
 * How it works inside the iframe:
 *  1. Tailwind (Play CDN) + Babel-standalone load as globals.
 *  2. An import map maps `react`, `react-dom/client`, `recharts`, `lucide-react`,
 *     `framer-motion` to pinned esm.sh builds — all sharing ONE React (via
 *     `?deps=react@…`) so hooks work and there's no "two Reacts" error.
 *  3. Babel transpiles the component (JSX/TS) to JS, keeping its `import`s and
 *     using the automatic JSX runtime.
 *  4. The transpiled code becomes a Blob module and is dynamically imported, so
 *     the browser's native ESM + import map resolve the libraries.
 *  5. The default export is mounted inside an error boundary; transpile and
 *     runtime errors render a visible panel instead of a blank frame.
 *
 * Security: the caller renders this with `sandbox="allow-scripts"` and NO
 * `allow-same-origin`, so the frame is origin-null and can't touch the parent.
 */

// Pinned versions — never floating (@latest) so a CDN change can't alter behavior.
const REACT = '18.3.1';
const CDN = {
  tailwind: 'https://cdn.tailwindcss.com/3.4.16',
  babel: 'https://unpkg.com/@babel/standalone@7.26.4/babel.min.js',
  react: `https://esm.sh/react@${REACT}`,
  reactJsx: `https://esm.sh/react@${REACT}/jsx-runtime`,
  reactDomClient: `https://esm.sh/react-dom@${REACT}/client`,
  recharts: `https://esm.sh/recharts@2.13.3?deps=react@${REACT},react-dom@${REACT}`,
  lucide: `https://esm.sh/lucide-react@0.468.0?deps=react@${REACT}`,
  framer: `https://esm.sh/framer-motion@11.15.0?deps=react@${REACT},react-dom@${REACT}`,
};

/**
 * Embed arbitrary source as a JS string literal that is also safe inside an HTML
 * `<script>` — escaping every `<` (so a literal `</script>` in the code can't end
 * the tag) and line separators the JSON spec leaves raw.
 */
function toJsStringLiteral(code: string): string {
  return JSON.stringify(code)
    .replace(/</g, '\\u003c')
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029');
}

export function buildReactSrcDoc(code: string): string {
  const src = toJsStringLiteral(code);
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<script src="${CDN.tailwind}"></script>
<script src="${CDN.babel}"></script>
<script type="importmap">
{
  "imports": {
    "react": "${CDN.react}",
    "react/jsx-runtime": "${CDN.reactJsx}",
    "react-dom": "https://esm.sh/react-dom@${REACT}",
    "react-dom/client": "${CDN.reactDomClient}",
    "recharts": "${CDN.recharts}",
    "lucide-react": "${CDN.lucide}",
    "framer-motion": "${CDN.framer}"
  }
}
</script>
<style>
  html, body { margin: 0; background: #fff; }
  #root { min-height: 100vh; }
  #err {
    display: none; margin: 12px; padding: 12px 14px; border-radius: 8px;
    border: 1px solid #fecaca; background: #fef2f2; color: #991b1b;
    font: 13px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; white-space: pre-wrap;
  }
  #err b { color: #7f1d1d; }
</style>
</head>
<body>
<div id="root"></div>
<pre id="err"></pre>
<script type="module">
  const errEl = document.getElementById('err');
  function showError(title, message) {
    errEl.style.display = 'block';
    errEl.textContent = '';
    const b = document.createElement('b');
    b.textContent = title + '\\n';
    errEl.appendChild(b);
    errEl.appendChild(document.createTextNode(String(message || '')));
  }
  window.addEventListener('error', (e) => showError('Runtime error', e.message));
  window.addEventListener('unhandledrejection', (e) => showError('Runtime error', e.reason && e.reason.message ? e.reason.message : e.reason));

  const SOURCE = ${src};
  try {
    // Transpile JSX/TS, keep imports, use the automatic JSX runtime.
    const transpiled = Babel.transform(SOURCE, {
      presets: [['react', { runtime: 'automatic' }], 'typescript'],
      filename: 'artifact.tsx',
    }).code;

    const blobUrl = URL.createObjectURL(new Blob([transpiled], { type: 'text/javascript' }));
    const mod = await import(blobUrl);
    const App = mod.default;
    if (typeof App !== 'function') {
      throw new Error('Expected a default-exported React component, e.g. \`export default function App() { ... }\`');
    }

    const React = await import('react');
    const { createRoot } = await import('react-dom/client');

    // Error boundary so a render-time throw shows a message, not a blank frame.
    class ErrorBoundary extends React.Component {
      constructor(props) { super(props); this.state = { error: null }; }
      static getDerivedStateFromError(error) { return { error }; }
      componentDidCatch(error) { showError('Component error', error && error.message ? error.message : error); }
      render() {
        if (this.state.error) return null;
        return this.props.children;
      }
    }

    createRoot(document.getElementById('root')).render(
      React.createElement(ErrorBoundary, null, React.createElement(App))
    );
  } catch (err) {
    showError('Failed to build component', err && err.message ? err.message : err);
  }
</script>
</body>
</html>`;
}
