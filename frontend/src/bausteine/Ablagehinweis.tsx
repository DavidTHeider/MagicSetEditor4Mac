// **Was ein Ablegen tut, steht vor dem Loslassen da** — AP-4
// (#2361),
// Befund B-24, Entwurf `Zustaende09` („Ablageziel: die Wirkung steht da,
// bevor losgelassen wird“).
//
// Bis dahin griff kein Stil die Markierung von Wails auf, und ein Bild, das
// neben die Karte fiel, verschwand wortlos. **Die Live-Region steht immer**
// — eine, die erst mit ihrem Satz erscheint, sagt manches Vorleseprogramm
// nicht an —, und der sichtbare Kasten taucht mit dem Zug auf.
import { useAblageziel } from '../haken/useAblageziel'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

// **Über der Seite, ohne sie zu verdecken**: oben in der Mitte, und der
// Zeiger geht hindurch, damit der Zug darunter weiterläuft. Kante in der
// Marke, wie das Ziel selbst; kein Schatten (Festlegung 11 nennt ihn nur für
// Palette, Popover und Menü).
const hinweisklassen =
  'auftauchen pointer-events-none fixed left-1/2 top-3 z-50 max-w-md -translate-x-1/2 ' +
  'rounded-md border-2 border-marke bg-flaeche px-3 py-2 text-sm text-text'

export function Ablagehinweis() {
  const worte = useTexte().bausteine.ablage
  const ziel = useAblageziel()

  const satz =
    ziel === undefined
      ? undefined
      : ziel.bildfeld === undefined
        ? worte.fenster
        : einsetzen(worte.bildfeld, { feld: ziel.bildfeld })

  return (
    <div role="status" aria-live="polite">
      {satz !== undefined && <p className={hinweisklassen}>{satz}</p>}
    </div>
  )
}
