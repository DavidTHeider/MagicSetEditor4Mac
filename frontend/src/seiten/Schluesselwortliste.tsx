// Die **Liste der eigenen Schlüsselwörter** — UXR-F3
// (#2009).
//
// # Was MSE2 hier tut
//
// `KeywordList` (`src/gui/control/keyword_list.cpp`, Fassung f9d9356 vom
// 2021-07-06) ist eine Liste mit fünf Spalten — *Keyword*, *Matches*, *Mode*,
// *Uses* und *Reminder text* (Zeilen 34–38) —, und neben ihr steht der Editor
// des gewählten Wortes. Das Muster zeigt sie ohne Auszeichnung, die
// Parameter in Winkelklammern (`match_string`, Zeilen 145–151); ein Wort, das
// nichts erkennt, steht rot da (`OnGetItemAttr`, Zeilen 225–235).
//
// # Was hier übernommen ist — und was nicht
//
// **Übernommen** sind *Keyword*, *Matches* und *Mode*: Name, Muster,
// Verwendung. **Nicht** der Erinnerungstext — der steht im Editor, und in
// einer Spalte von 460 Punkten gäbe er den anderen keinen Platz — und nicht
// *Uses*: die Zählung braucht die Auswertung jeder Karte
// (`KeywordList::updateUsageStatistics`), und die steht hier an keiner
// Bindung. Sortieren nach einer Spalte und die Suche (`ID_KEYWORD_FILTER`)
// fehlen ebenfalls; ein Set bringt selten mehr als eine Handvoll eigene
// mit. Alle drei stehen in `docs/PORTIERUNGSKARTE.md` als offen.
//
// # Die Tastatur ist die der Kartenliste
//
// ↑ und ↓ wählen und nehmen den Fokus mit, Pos1 und Ende springen an die
// Enden, am Rand ist Schluss (kein Umlauf), → führt in die Zeile und ← zurück,
// und der Tabulator verlässt die Liste als Ganzes — mit einer wandernden
// Tabstelle auf dem gewählten Wort. **Dieselbe Funktion** (`blaettertaste`),
// keine zweite Fassung; warum ein `grid` und keine `listbox`, steht in
// `kartenlistetasten.ts`.
//
// **Ohne Sichtfenster**: die Karten liegen zu Tausenden vor, die eigenen
// Schlüsselwörter eines Sets zu Dutzenden — gezeichnet wird alles, und das
// Rollen besorgt der Behälter.
import { useRef } from 'react'

import type { SchluesselwortDTO } from '../api'
import { useTexte } from '../sprache/kontext'
import { blaettertaste } from './kartenlistetasten'

interface ListeEigenschaften {
  woerter: readonly SchluesselwortDTO[]

  // gewaehlt ist die Kennung des gewählten Wortes — oder keine: beim Öffnen
  // ist, wie bei MSE, nichts gewählt.
  gewaehlt: string | undefined
  waehlen: (wortID: string) => void
}

// musterAnzeige macht aus einem Muster, was die Liste zeigt.
//
// **Wie `match_string` bei MSE**: `<atom-param>` und `</atom-param>` werden zu
// den einfachen Winkelklammern ‹ und › (`LEFT_ANGLE_BRACKET`,
// `src/util/string.hpp`, Zeilen 77–78), jede andere Auszeichnung fällt weg
// (`untag`, `src/util/tagged_string.cpp`, Zeilen 33–47).
export function musterAnzeige(muster: string): string {
  return muster
    .replaceAll('<atom-param>', '‹')
    .replaceAll('</atom-param>', '›')
    .replace(/<[^>]*>/g, '')
}

// Die Spalten stehen in **einer** Klasse, damit Kopf und Zeilen nicht
// auseinanderlaufen. **Benannt und nicht als Ausdruck im JSX**, wie in der
// Seitenleiste: `keintextimjsx.test.ts` weist eine Zeichenkette mit
// Platzhalter dort ab.
const spalten =
  'grid grid-cols-[minmax(0,2fr)_minmax(0,3fr)_minmax(0,1fr)] gap-x-3 '

const kopfklassen =
  spalten + 'sticky top-0 bg-flaeche px-2 py-1 text-xs text-leise'

const zeilenklassen =
  spalten +
  'relative items-baseline px-2 py-1.5 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-ring aria-selected:bg-betont'

export function Schluesselwortliste({
  woerter,
  gewaehlt,
  waehlen,
}: ListeEigenschaften) {
  const texte = useTexte().seiten.karten.schluesselwoerter
  const gitter = useRef<HTMLDivElement>(null)

  // **Die Tabstelle liegt auf dem gewählten Wort — sonst auf dem ersten.**
  // Der Tabulator führt in die Liste; wohin genau, entscheidet sie. Ohne
  // Wahl (so beginnt sie) wäre die Liste sonst nur mit der Maus zu
  // erreichen.
  const tabwort = woerter.some((wort) => wort.id === gewaehlt)
    ? gewaehlt
    : woerter[0]?.id

  // fokussiere setzt den Fokus auf die Zeile eines Wortes. **Die Zeilen
  // stehen alle im Baum**, also gibt es nichts zu rollen, bevor der Fokus
  // folgen kann; der Browser holt die Zeile selbst ins Bild.
  const fokussiere = (wortID: string) => {
    gitter.current
      ?.querySelector<HTMLElement>(`[data-wort="${wortID}"]`)
      ?.focus()
  }

  return (
    <div
      ref={gitter}
      role="grid"
      aria-label={texte.eigene}
      // **Der Kopf zählt als Zeile** — sonst sagte eine Vorlesehilfe „Zeile 1
      // von 5" zum zweiten Wort.
      aria-rowcount={woerter.length + 1}
      className="max-h-56 divide-y divide-kante overflow-y-auto rounded-sm border border-kante"
    >
      <div role="row" aria-rowindex={1} className={kopfklassen}>
        <span role="columnheader">{texte.wort}</span>
        <span role="columnheader">{texte.muster}</span>
        <span role="columnheader">{texte.modus}</span>
      </div>

      {woerter.map((wort, stelle) => {
        const ist = wort.id === gewaehlt
        const name = wort.wort === '' ? texte.ohneName : wort.wort

        return (
          <div
            key={wort.id}
            role="row"
            aria-rowindex={stelle + 2}
            aria-selected={ist}
            aria-label={name}
            tabIndex={wort.id === tabwort ? 0 : -1}
            data-wort={wort.id}
            onKeyDown={(ereignis) => {
              // **⏎ und die Leertaste wählen die fokussierte Zeile.** Ohne
              // Wahl (so beginnt die Liste) wählt keine Pfeiltaste die erste
              // — ↓ führt zur zweiten, Pos1 bleibt, wo es ist —, und wer mit
              // dem Tabulator hereinkommt, käme sonst nicht an sie heran.
              // **Auch auf dem Knopf darin**: die Taste ist verbraucht, der
              // Knopf klickt nicht noch einmal, und gewählt wird genau einmal.
              if (ereignis.key === 'Enter' || ereignis.key === ' ') {
                ereignis.preventDefault()
                waehlen(wort.id)

                return
              }

              blaettertaste(ereignis, woerter, stelle, (wortID) => {
                waehlen(wortID)
                fokussiere(wortID)
              })
            }}
            // **Die ganze Zeile ist das Ziel**: der Knopf in der ersten Zelle
            // spannt sich über sie (`after:inset-0`). Ein Ziel von 12 × 60
            // Punkten trifft man, eines von 12 × 12 nicht.
            className={zeilenklassen}
          >
            <span role="gridcell" className="min-w-0">
              <button
                type="button"
                // **Innerhalb des Gitters erreichbar, nicht über den
                // Tabulator** — mit → und ←, wie in der Kartenliste. **Und
                // deshalb mit Ring** (#2249): wer mit → hineinkommt, sah
                // bis dahin nicht, wo der Fokus steht.
                tabIndex={-1}
                aria-current={ist ? 'true' : undefined}
                onClick={() => {
                  waehlen(wort.id)
                }}
                className="block max-w-full truncate rounded-sm text-left after:absolute after:inset-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[current=true]:font-medium"
              >
                {name}
              </button>
            </span>

            <span role="gridcell" className="min-w-0 text-leise">
              <span className="block truncate">
                {musterAnzeige(wort.muster)}
              </span>

              {/* **Rot wie bei MSE, und als Text** — eine Farbe allein sagt
                  es nur dem, der sie sieht. */}
              {!wort.gueltig && (
                <span className="block text-fehler">{texte.ungueltig}</span>
              )}
            </span>

            <span role="gridcell" className="min-w-0 truncate text-leise">
              {wort.modus}
            </span>
          </div>
        )
      })}
    </div>
  )
}
