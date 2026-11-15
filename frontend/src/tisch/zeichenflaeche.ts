// **Wie groß die Zeichenfläche ist** — gefunden am 2026-09-24 in der
// WebKit-Probe
// (#1828).
//
// # Warum eine eigene Datei
//
// **Weil es eine Entscheidung ist, und `pixinaht.ts` wird nicht gemessen**
// (`.covignore`: ohne WebGL gibt es dort keinen Renderer). Die Fläche geht in
// alles ein, was über der Leinwand liegt — die Kamera passt den Tisch in sie
// ein, die Kartenknöpfe stehen nach ihr, der Zeiger wird nach ihr
// zurückgerechnet.
//
// # Warum `screen` und nicht `width / resolution`
//
// **Weil PixiJS 8 die Dichte schon herausgerechnet hat.** `resize` setzt die
// Quelle der Sichttextur auf Punkte — Pixel geteilt durch die Dichte — und
// `screen` auf deren Rahmen; `width` des Renderers ist derselbe Rahmen
// (pixi.js 8.20.1, `ViewSystem.mjs`, `TextureSource.mjs` und
// `AbstractRenderer.mjs` unter `lib/rendering/renderers/shared/`). Der
// Kommentar über `width` spricht zwar von Pixeln der Leinwand; der Rumpf
// sagt es anders, und er gilt.
//
// **Mit `width / resolution`** wäre die Fläche auf einem Retina-Bildschirm
// halb so breit und halb so hoch wie die Leinwand: der Tisch läge im linken
// oberen Viertel, seine Knöpfe mit ihm. Kein Lauf sieht das — jsdom und
// headless Chrome haben die Dichte eins, und dort teilt die Rechnung durch
// eins. Sehen kann es die WebKit-Probe mit doppelter Dichte
// (`scripts/webkitprobe/abnahme/play06-tisch.probe`).
export function zeichenflaeche(renderer: {
  screen: { width: number; height: number }
}): { breite: number; hoehe: number } {
  return { breite: renderer.screen.width, hoehe: renderer.screen.height }
}
