// **Die Überlagerungen der Vorschau** — Werkstück U5 von Phase 5.1
// (#1416).
//
// Drei Schalter, drei Linien über dem Bild: die **Bildgrenzen** zeichnen die
// Flächen der Bildfelder nach, die **Diagnose** die Flächen, an denen der
// Rasterer einen Befund hat, und der **Beschnitt** eine Linie 3 mm innerhalb
// der Kante. Die Flächen selbst liegen als Knöpfe über dem Bild (#246); die
// Linien sind `pointer-events: none` und nehmen ihnen nichts.
//
// **Neu, nicht aus MSE2**: dort gibt es weder Überlagerungen noch Beschnitt
// (`card_viewer.cpp` zeichnet die Karte und sonst nichts; Karten werden auf
// Maß gedruckt). Die Beschnittlinie ist eine Hilfe für den, der zu Hause
// schneidet — und sie steht nur da, wo die Vorlage eine Auflösung nennt.
import type { VorschauDTO } from '../api'
import type { FeldflaecheDTO } from '../formen'
import { useTexte } from '../sprache/kontext'
import { beschnittProzent } from './zoom'

export interface Overlaywahl {
  bildgrenzen: boolean
  diagnose: boolean
  beschnitt: boolean
}

export const keineOverlays: Overlaywahl = {
  bildgrenzen: false,
  diagnose: false,
  beschnitt: false,
}

interface SchalterEigenschaften {
  wahl: Overlaywahl
  umschalten: (art: keyof Overlaywahl) => void
  beschnittMoeglich: boolean
}

export function Overlayschalter({
  wahl,
  umschalten,
  beschnittMoeglich,
}: SchalterEigenschaften) {
  const worte = useTexte().seiten.karten.overlays

  const kasten = (art: keyof Overlaywahl, text: string, moeglich = true) => (
    <label
      key={art}
      className="flex items-center gap-1 text-xs has-disabled:opacity-50"
    >
      <input
        type="checkbox"
        checked={wahl[art]}
        disabled={!moeglich}
        onChange={() => {
          umschalten(art)
        }}
        className="size-3.5 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      />
      {text}
    </label>
  )

  return (
    <fieldset className="flex flex-wrap items-center gap-3">
      <legend className="sr-only">{worte.bezeichnung}</legend>
      {kasten('bildgrenzen', worte.bildgrenzen)}
      {kasten('diagnose', worte.diagnose)}
      {kasten('beschnitt', worte.beschnitt, beschnittMoeglich)}
    </fieldset>
  )
}

interface LinienEigenschaften {
  wahl: Overlaywahl
  gezeigt: VorschauDTO
  istBildfeld: (feld: string) => boolean

  // befundfelder sind die Felder, an denen der Rasterer etwas fand.
  befundfelder: ReadonlySet<string>
}

// prozent legt eine Fläche in Prozent des Bildes — dieselbe Rechnung wie bei
// den Knöpfen, damit Linie und Knopf deckungsgleich sind.
function prozent(flaeche: FeldflaecheDTO, gezeigt: VorschauDTO) {
  return {
    left: `${String((flaeche.links / gezeigt.breite) * 100)}%`,
    top: `${String((flaeche.oben / gezeigt.hoehe) * 100)}%`,
    width: `${String((flaeche.breite / gezeigt.breite) * 100)}%`,
    height: `${String((flaeche.hoehe / gezeigt.hoehe) * 100)}%`,
  }
}

export function Overlaylinien({
  wahl,
  gezeigt,
  istBildfeld,
  befundfelder,
}: LinienEigenschaften) {
  const worte = useTexte().seiten.karten.overlays
  const beschnitt = wahl.beschnitt ? beschnittProzent(gezeigt) : undefined

  return (
    <>
      {gezeigt.flaechen.map((flaeche, stelle) => {
        const bild = wahl.bildgrenzen && istBildfeld(flaeche.feld)
        const befund = wahl.diagnose && befundfelder.has(flaeche.feld)

        if (!bild && !befund) {
          return null
        }

        return (
          <div
            key={`linie/${flaeche.feld}/${String(stelle)}`}
            data-overlay={befund ? 'diagnose' : 'bildgrenzen'}
            title={befund ? `${worte.diagnose}: ${flaeche.feld}` : flaeche.feld}
            style={prozent(flaeche, gezeigt)}
            className={
              'pointer-events-none absolute border-2 ' +
              (befund ? 'border-fehler' : 'border-dashed border-auswahlkante')
            }
          />
        )
      })}

      {beschnitt !== undefined && (
        <div
          data-overlay="beschnitt"
          title={worte.beschnittHinweis}
          style={{
            left: `${String(beschnitt.x)}%`,
            top: `${String(beschnitt.y)}%`,
            right: `${String(beschnitt.x)}%`,
            bottom: `${String(beschnitt.y)}%`,
          }}
          className="pointer-events-none absolute border border-dotted border-markierungskante"
        />
      )}
    </>
  )
}
