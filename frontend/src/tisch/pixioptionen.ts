// **Was der Tisch von PixiJS verlangt** — die Optionen für `Application.init`,
// herausgezogen aus `pixinaht.ts`, das ohne echten Renderer nicht läuft und
// deshalb von der Abdeckung ausgenommen ist. Was dort entscheidet, steht hier
// und wird gemessen.
import type { ApplicationOptions } from 'pixi.js'

export function tischoptionen(
  leinwand: HTMLCanvasElement,
  grund: number,
): Partial<ApplicationOptions> {
  return {
    canvas: leinwand,
    preference: 'webgl',
    antialias: true,
    background: grund,

    // **Kein Ticker.** Ein Kartentisch ist keine Animation: er ändert sich,
    // wenn jemand etwas tut, und steht sonst still. Ein laufender Ticker
    // kostet dann sechzig Bilder je Sekunde für nichts — und wenn der
    // Grafikkontext wegfällt, scheitert er sechzigmal je Sekunde laut.
    //
    // Gemessen in der Sichtprüfung am 2026-09-16: **4.700 Konsolenmeldungen**
    // („Could not retrieve shader source"), bis der Puffer überlief und die
    // Seite nicht mehr antwortete. Gezeichnet wird jetzt auf Zuruf.
    autoStart: false,

    // **autoDensity mit der Gerätedichte** — sonst ist eine Karte auf einem
    // Retina-Bildschirm weich. #1460 verlangt die Probe daran ausdrücklich.
    autoDensity: true,
    resolution: leinwand.ownerDocument.defaultView?.devicePixelRatio ?? 1,

    resizeTo: leinwand.parentElement ?? leinwand,

    // **Tab bleibt beim Tisch.** PixiJS 8.22 aktiviert seine
    // Barrierefreiheitsschicht wieder beim ersten Tab (Vorgabe `true`, von
    // 8.14 bis 8.21 ausgefallen); dieser Tisch bedient sich über eigene
    // Knöpfe, und mit 8.20.1 lag über der Leinwand keine solche Schicht.
    accessibilityOptions: { activateOnTab: false },
  }
}
