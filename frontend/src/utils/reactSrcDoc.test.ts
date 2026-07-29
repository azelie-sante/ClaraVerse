import { describe, it, expect } from 'vitest';
import { buildReactSrcDoc } from './reactSrcDoc';

describe('buildReactSrcDoc', () => {
  const code = `export default function App() {
  const [n, setN] = React.useState(0);
  return <button onClick={() => setN(n + 1)}>{n}</button>;
}`;

  it('produces a full HTML document', () => {
    const html = buildReactSrcDoc(code);
    expect(html).toMatch(/^<!DOCTYPE html>/);
    expect(html).toContain('<div id="root">');
    expect(html).toContain('<script type="module">');
  });

  it('pins CDN versions (no floating @latest)', () => {
    const html = buildReactSrcDoc(code);
    expect(html).toContain('esm.sh/react@18.3.1');
    expect(html).toContain('@babel/standalone@7.26.4');
    expect(html).toContain('cdn.tailwindcss.com/3.4.16');
    expect(html).not.toContain('@latest');
  });

  it('provides an import map with the supported libraries', () => {
    const html = buildReactSrcDoc(code);
    expect(html).toContain('"react":');
    expect(html).toContain('"recharts":');
    expect(html).toContain('"lucide-react":');
    expect(html).toContain('"framer-motion":');
    // all libs share one React so hooks work / no "two Reacts"
    expect(html).toContain('deps=react@18.3.1');
  });

  it('includes the error-boundary + transpile error handling', () => {
    const html = buildReactSrcDoc(code);
    expect(html).toContain('class ErrorBoundary');
    expect(html).toContain('Failed to build component');
    expect(html).toContain('getDerivedStateFromError');
  });

  it('embeds the component source', () => {
    const html = buildReactSrcDoc(code);
    // JSON-stringified with < escaped as <
    expect(html).toContain('const SOURCE =');
    expect(html).toContain('setN');
  });

  it('neutralizes a literal </script> in user code so it cannot break out', () => {
    const evil = `export default function App(){ const s = "</script><img src=x onerror=alert(1)>"; return <div>{s}</div>; }`;
    const html = buildReactSrcDoc(evil);
    // The raw closing-script sequence must not appear inside the embedded source
    const moduleStart = html.indexOf('const SOURCE =');
    const embedded = html.slice(moduleStart, html.indexOf('\n', moduleStart));
    expect(embedded).not.toContain('</script>');
    expect(embedded).toContain('\\u003c/script');
  });
});
