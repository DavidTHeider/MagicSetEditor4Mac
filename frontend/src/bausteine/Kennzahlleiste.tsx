// **Eine Leiste für die Kennzahlen eines Bereichs** — AP-9
// (#2366),
// Befund B-15, Entwurf `Bereiche09`: „Hauptdeck 60 · Sideboard 15“,
// „Fehlt in der Sammlung 4“, „Wert 38 € · ältester Preis vor 3 Tagen“.
//
// # Ein Baustein, jede Zahl ein Ziel
//
// **Über Set, Sammlung, Deck und Paketen dieselbe Leiste**, damit eine Zahl
// überall gleich aussieht und gleich bedient wird. **Eine Zahl mit Ziel ist
// ein Knopf** und führt mit einem Klick dorthin, wo sie zerfällt; eine ohne
// Ziel ist Text — eine Fläche, die wie ein Knopf aussieht und nichts tut,
// wäre eine lose Bindung.
//
// # Eine Tabstelle, wie eine Symbolleiste
//
// **Die Leiste ist eine `toolbar` mit wandernder Tabstelle**: ⇥ betritt sie
// einmal, ← → Pos1 Ende gehen in ihr weiter, und wer wiederkommt, landet,
// wo er war. Mit einer Tabstelle je Zahl rückten drei Zahlen zwischen Kopf
// und Suchfeld der Sammlung, und der Kernweg `sammlung` verlöre sein Ziel
// „≤ 13. Tabstelle“ (`docs/UX-ARCHITEKTUR.md`) — dieselbe Überlegung wie bei
// der Seitenleiste (#1488).
//
// # Laden in Endhöhe
//
// **Solange ein Wert aussteht, steht die Zeile unsichtbar gefüllt da** und
// sagt `aria-busy` (AP-4): sonst rückte, was darunter steht, beim Eintreffen.
// Ein Ziel bietet sie dann nicht an — es zeigte auf eine Zahl, die noch
// niemand kennt.

import { useState, type KeyboardEvent } from 'react'

// Kennzahl ist eine Zahl der Leiste.
export interface Kennzahl {
  schluessel: string
  // was sagt, worauf die Zahl sich bezieht — „Hauptdeck“.
  was: string
  // wert ist die Zahl, wie man sie liest — `undefined`, solange sie aussteht.
  wert: string | undefined
  // unter ist die Zeile darunter — „Sideboard 15“, „ältester Preis …“.
  unter?: string | undefined
  // hin führt zum Ziel. Ohne ist die Zahl Text.
  hin?: (() => void) | undefined
}

// platzhalter füllt einen ausstehenden Wert unsichtbar — so hoch wie einer.
const platzhalter = '0'

const knopf =
  'flex h-full w-full flex-col items-start px-4 py-2 text-left ' +
  'hover:bg-betont focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:-outline-offset-2 focus-visible:outline-ring'

export function Kennzahlleiste({
  bezeichnung,
  kennzahlen,
}: {
  bezeichnung: string
  kennzahlen: readonly Kennzahl[]
}) {
  const laedt = kennzahlen.some((kennzahl) => kennzahl.wert === undefined)
  const ziele = kennzahlen.filter(
    (kennzahl) => kennzahl.hin !== undefined && kennzahl.wert !== undefined,
  )

  // **Die wandernde Tabstelle**: die zuletzt erreichte Zahl, sonst die erste
  // mit Ziel.
  const [zuletzt, setzeZuletzt] = useState<string>()
  const tabstelle = ziele.some((kennzahl) => kennzahl.schluessel === zuletzt)
    ? zuletzt
    : ziele[0]?.schluessel

  // **Tasten kommen nur von ihren Knöpfen** — sonst ist in der Leiste
  // nichts fokussierbar.
  const aufTaste = (ereignis: KeyboardEvent<HTMLDivElement>) => {
    const knoepfe = [...ereignis.currentTarget.querySelectorAll('button')]
    const stelle = knoepfe.indexOf(ereignis.target as HTMLButtonElement)
    const ziel = {
      ArrowRight: (stelle + 1) % knoepfe.length,
      ArrowLeft: (stelle - 1 + knoepfe.length) % knoepfe.length,
      Home: 0,
      End: knoepfe.length - 1,
    }[ereignis.key]

    if (ziel !== undefined) {
      ereignis.preventDefault()
      for (const knopf of knoepfe.slice(ziel, ziel + 1)) {
        knopf.focus()
      }
    }
  }

  return (
    <div
      role="toolbar"
      aria-label={bezeichnung}
      aria-busy={laedt ? true : undefined}
      onKeyDown={aufTaste}
      className="flex flex-wrap divide-x divide-kante overflow-hidden rounded-lg border border-kante bg-flaeche"
    >
      {kennzahlen.map((kennzahl) => {
        const inhalt = (
          <>
            {/* **Die Leerzeichen sind für den Namen** — „Hauptdeck 60“ und
                nicht „Hauptdeck60“; im Bild stehen die drei untereinander. */}
            <span className="text-xs text-leise">{kennzahl.was}</span>{' '}
            <span
              className={
                kennzahl.wert === undefined
                  ? 'invisible text-lg font-semibold tabular-nums'
                  : 'text-lg font-semibold tabular-nums'
              }
            >
              {kennzahl.wert ?? platzhalter}
            </span>
            {kennzahl.unter !== undefined && ' '}
            <span className="min-h-4 text-xs text-leise">{kennzahl.unter}</span>
          </>
        )

        return (
          <div key={kennzahl.schluessel} className="min-w-32 flex-1">
            {kennzahl.hin !== undefined && kennzahl.wert !== undefined ? (
              <button
                type="button"
                tabIndex={kennzahl.schluessel === tabstelle ? 0 : -1}
                className={knopf}
                onClick={kennzahl.hin}
                onFocus={() => {
                  setzeZuletzt(kennzahl.schluessel)
                }}
              >
                {inhalt}
              </button>
            ) : (
              <div className="flex flex-col items-start px-4 py-2">
                {inhalt}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
