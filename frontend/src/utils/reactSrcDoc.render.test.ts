/**
 * Live render test: builds the srcDoc, loads it inside the SAME sandboxed iframe
 * the app uses (allow-scripts, no allow-same-origin), in headless Chromium, and
 * verifies a real component — useState + recharts + lucide-react — actually
 * mounts with no error panel. This exercises the whole chain: Babel transpile →
 * blob module → esm.sh import map → React mount, under the real sandbox.
 *
 * Requires network (esm.sh) and Playwright's chromium. Skipped automatically if
 * chromium isn't installed.
 */
import { describe, it, expect } from 'vitest';
import { buildReactSrcDoc } from './reactSrcDoc';

const COMPONENT = `
import { useState } from 'react';
import { BarChart, Bar } from 'recharts';
import { Heart } from 'lucide-react';

export default function App() {
  const [n, setN] = useState(7);
  return (
    <div className="p-4">
      <h1 id="title">Count: {n}</h1>
      <button onClick={() => setN(n + 1)}>inc</button>
      <Heart id="icon" />
      <BarChart width={200} height={100} data={[{ v: 3 }, { v: 5 }]}>
        <Bar dataKey="v" />
      </BarChart>
    </div>
  );
}
`;

describe('React artifact live render (sandboxed iframe)', () => {
  it('mounts a component using hooks + recharts + lucide with no error', async () => {
    let chromium: typeof import('playwright').chromium;
    try {
      ({ chromium } = await import('playwright'));
    } catch {
      console.warn('playwright not installed — skipping live render test');
      return;
    }

    const browser = await chromium.launch();
    try {
      const page = await browser.newPage();
      const srcDoc = buildReactSrcDoc(COMPONENT);

      // Load the srcDoc inside the exact sandboxed iframe the app uses.
      await page.setContent('<body></body>');
      await page.evaluate((html) => {
        const f = document.createElement('iframe');
        f.id = 'frame';
        f.setAttribute('sandbox', 'allow-scripts allow-popups allow-popups-to-escape-sandbox');
        f.style.width = '400px';
        f.style.height = '400px';
        f.srcdoc = html as string;
        document.body.appendChild(f);
      }, srcDoc);

      const frame = page.frameLocator('#frame');

      // Component mounted (network fetch + transpile + render can take a moment).
      await frame.locator('#title').waitFor({ timeout: 45000 });
      expect(await frame.locator('#title').textContent()).toBe('Count: 7');

      // recharts rendered an SVG chart, lucide rendered an icon.
      expect(await frame.locator('svg.recharts-surface').count()).toBeGreaterThan(0);
      expect(await frame.locator('svg#icon, #icon').count()).toBeGreaterThan(0);

      // Interactivity: hooks work.
      await frame.locator('button').click();
      expect(await frame.locator('#title').textContent()).toBe('Count: 8');

      // No error panel.
      const errVisible = await frame.locator('#err').isVisible();
      expect(errVisible).toBe(false);
    } finally {
      await browser.close();
    }
  }, 60000);
});
