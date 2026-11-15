// **Der Trenner zwischen zwei Spalten** — AP-12
// (#2369),
// Befund B-25.
//
// **adapted** aus MSE2: ein `wxSplitterWindow` trennt dort Liste und Tafel,
// mit einer Mindestbreite und nur mit der Maus (`src/gui/set/cards_panel.cpp`
// 52–55, `src/gui/set/keywords_panel.cpp` 88–91, Fassung f9d9356 vom
// 2021-07-06), und gemerkt wird nichts. Hier kommen die Tastatur und das
// Gedächtnis dazu; die Grenzen stehen in `domain.BreiteMindestens` und
// `BreiteHoechstens`, die obere wird vom Platz im Fenster weiter begrenzt.
//
// # Was er tut
//
//   - **Pfeiltasten** gehen 16 Punkte, mit ⇧ 64, Pos1 und Ende an die
//     Grenzen — jeder Schritt geht sofort als Wunsch hinaus, und das
//     Gedächtnis schreibt den letzten.
//   - **Beim Ziehen** wächst die Spalte als Vorschau mit; geschrieben wird
//     einmal, beim Loslassen. Ein abgebrochener Zug schreibt nichts.
//
// # Was er nicht tut
//
// Er kennt keine Spalte: wer breiter wird, sagt `richtung` (die linke Spalte
// wächst nach rechts, die rechte nach links), und was mit der Zahl geschieht,
// sagt der Aufrufer.
import { useRef, type KeyboardEvent, type PointerEvent } from 'react'

interface SpaltentrennerEigenschaften {
  bezeichnung: string
  // wertetext sagt die Breite mit Einheit — „240 Punkte“ statt „240“.
  wertetext: string
  steuert: string
  breite: number
  mindestens: number
  hoechstens: number
  richtung: 1 | -1
  aendern: (punkte: number) => void
  vorschau: (punkte: number | undefined) => void
}

const schritt = 16
const grosserSchritt = 64

export function Spaltentrenner({
  bezeichnung,
  wertetext,
  steuert,
  breite,
  mindestens,
  hoechstens,
  richtung,
  aendern,
  vorschau,
}: SpaltentrennerEigenschaften) {
  // zug hält Anfang und letzte Breite eines Zuges — kein Zustand, denn das
  // Zeichnen der Vorschau übernimmt der Aufrufer.
  const zug = useRef<{ x: number; breite: number; jetzt: number }>(undefined)

  const begrenzt = (punkte: number) =>
    Math.min(hoechstens, Math.max(mindestens, Math.round(punkte)))

  const wuensche = (punkte: number) => {
    const neu = begrenzt(punkte)
    if (neu !== breite) {
      aendern(neu)
    }
  }

  const aufTaste = (ereignis: KeyboardEvent<HTMLDivElement>) => {
    const weite = ereignis.shiftKey ? grosserSchritt : schritt
    const ziele: Record<string, number | undefined> = {
      ArrowRight: breite + richtung * weite,
      ArrowLeft: breite - richtung * weite,
      Home: mindestens,
      End: hoechstens,
    }
    const ziel = ziele[ereignis.key]
    if (ziel === undefined) {
      return
    }

    ereignis.preventDefault()
    wuensche(ziel)
  }

  const beginne = (ereignis: PointerEvent<HTMLDivElement>) => {
    if (ereignis.button !== 0) {
      return
    }

    ereignis.currentTarget.setPointerCapture(ereignis.pointerId)
    zug.current = { x: ereignis.clientX, breite, jetzt: breite }
  }

  const ziehe = (ereignis: PointerEvent<HTMLDivElement>) => {
    const laufend = zug.current
    if (laufend === undefined) {
      return
    }

    laufend.jetzt = begrenzt(
      laufend.breite + richtung * (ereignis.clientX - laufend.x),
    )
    vorschau(laufend.jetzt)
  }

  const ende = (schreiben: boolean) => {
    const laufend = zug.current
    if (laufend === undefined) {
      return
    }

    zug.current = undefined
    vorschau(undefined)
    if (schreiben) {
      wuensche(laufend.jetzt)
    }
  }

  return (
    // **Ein fokussierbarer Trenner ist ein Bedienelement** — WAI-ARIA 1.2,
    // Rolle `separator`, und das Muster „Window Splitter“ der WAI: mit
    // `aria-valuenow` und Fokus nimmt er Pfeiltasten. jsx-a11y kennt nur den
    // stummen Trenner und hält ihn für nicht bedienbar.
    // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions -- siehe oben
    <div
      role="separator"
      aria-orientation="vertical"
      aria-label={bezeichnung}
      aria-controls={steuert}
      aria-valuenow={breite}
      aria-valuemin={mindestens}
      aria-valuemax={hoechstens}
      aria-valuetext={wertetext}
      // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- ein Trenner mit Wert nimmt den Fokus
      tabIndex={0}
      onKeyDown={aufTaste}
      onPointerDown={beginne}
      onPointerMove={ziehe}
      onPointerUp={() => {
        ende(true)
      }}
      onPointerCancel={() => {
        ende(false)
      }}
      className={
        'group relative w-4 shrink-0 cursor-col-resize touch-none ' +
        'focus-visible:outline focus-visible:outline-2 ' +
        'focus-visible:outline-offset-[-2px] focus-visible:outline-ring'
      }
    >
      {/* **Die Linie ist die Kante**, wie zwischen den Spalten bisher;
          beim Zeigen und mit Fokus wird sie zur Marke. */}
      <span
        aria-hidden="true"
        className={
          'absolute inset-y-0 left-1/2 w-px -translate-x-1/2 bg-kante ' +
          'group-hover:w-0.5 group-hover:bg-marke ' +
          'group-focus-visible:w-0.5 group-focus-visible:bg-marke'
        }
      />
    </div>
  )
}
