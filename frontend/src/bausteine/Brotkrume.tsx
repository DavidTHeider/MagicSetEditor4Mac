// **Der Weg über dem Kopf der Seite** — UX-03
// (#1675).
//
// # Warum eine Liste und kein Fließtext
//
// **Weil ein Vorleseprogramm sie zählen können soll.** Ein `<nav>` mit einer
// geordneten Liste sagt „drei Stationen, du bist bei der dritten"; drei
// `<span>` mit Pfeilen dazwischen sagen gar nichts. Das Trennzeichen ist
// deshalb `aria-hidden` — es ist Gestaltung und kein Wort.
import { useTexte } from '../sprache/kontext'

import type { Krume } from '../brotkrumen'
import type { Seitenschluessel } from '../navigation'

// **Die Klassen stehen benannt und nicht als Ausdruck im JSX** — dieselbe
// Regel wie in der Seitenleiste.
const reihe = 'flex flex-wrap items-center gap-1.5 px-8 pt-4 text-xs text-leise'

const verweis =
  'rounded-sm underline-offset-2 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

const hier = 'font-semibold text-text'

export interface Brotkrumeneigenschaften {
  krumen: readonly Krume[]
  wechseln: (seite: Seitenschluessel) => void
}

export function Brotkrume({ krumen, wechseln }: Brotkrumeneigenschaften) {
  const texte = useTexte()

  // **Kein leeres `<nav>`.** Auf der Startseite gibt es keinen Weg, und ein
  // Orientierungspunkt ohne Inhalt ist einer, den ein Vorleseprogramm
  // aufzählt und in dem nichts steht.
  if (krumen.length === 0) {
    return null
  }

  return (
    <nav aria-label={texte.allgemein.navigation.weg}>
      <ol className={reihe}>
        {krumen.map((krume, stelle) => {
          // **Das Ziel wird gebunden, nicht im Rückruf gelesen.** TypeScript
          // verengt eine Eigenschaft nicht über eine Funktionsgrenze hinweg
          // — `krume.ziel` wäre dort wieder `string | undefined`.
          const ziel = krume.ziel

          return (
            // **Die Stelle gehört zum Schlüssel**: auf *Handeln* heißen
            // Bereich und Seite gleich, und zwei gleiche Schlüssel ließen beim
            // Wechsel eine Station stehen (WebKit, 2026-09-24).
            <li
              key={`${String(stelle)}:${krume.beschriftung}`}
              className="flex items-center gap-1.5"
            >
              {stelle > 0 && <span aria-hidden="true">›</span>}

              {/* **Nur die letzte Station ist, wo man steht.** Eine ohne
                  Ziel mitten im Weg — die Partie über dem Tisch — führte auf
                  dieselbe Seite und ist deshalb kein Knopf; die Seite ist sie
                  trotzdem nicht. */}
              {ziel === undefined ? (
                <span
                  aria-current={
                    stelle === krumen.length - 1 ? 'page' : undefined
                  }
                  className={stelle === krumen.length - 1 ? hier : undefined}
                >
                  {krume.beschriftung}
                </span>
              ) : (
                <button
                  type="button"
                  className={verweis}
                  onClick={() => {
                    wechseln(ziel)
                  }}
                >
                  {krume.beschriftung}
                </button>
              )}
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
