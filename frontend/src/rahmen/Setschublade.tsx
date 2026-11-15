// **Eine Schublade des Sets** — *Eigenschaften* oder *Werkzeuge*, mit ihrer
// eigenen Laschenleiste: UX-08.AK-01
// (#1680),
// nach den gleichnamigen Artboards des Entwurfs.
//
// # Warum eine eigene Laschenleiste
//
// **Weil die Seiten Seiten bleiben** (`schubladen.ts`): *Stil* hat weiter
// sein Kürzel und seinen Eintrag in der Palette. Die Laschen hier sind
// dieselben Seitenwechsel wie in der Leiste des Arbeitsplatzes — sie bleiben
// nur in der Schublade, und die Karten darunter bleiben, wo sie sind.
//
// **Dasselbe Tastenmuster wie die Hauptleiste**: die Pfeile wandern, der Fokus
// wandert mit, und nur die gewählte Lasche steht im Tabweg.
import { useId, useRef, useState, type ReactNode } from 'react'

import { lasche } from '../bausteine/laschenstil'
import { Schublade } from '../bausteine/Schublade'
import { SchubladenweiteKontext } from '../haken/useSchubladenweite'
import {
  beschriftung,
  nachbarlasche,
  type Seitenschluessel,
  type Setschluessel,
} from '../navigation'
import { useTexte } from '../sprache/kontext'

import type { Schublade as Schubladenordnung } from '../schubladen'

interface SetschubladenEigenschaften {
  schublade: Schubladenordnung

  // aktuell ist die Seite, die in der Schublade steht.
  aktuell: Setschluessel

  wechseln: (seite: Seitenschluessel) => void

  // schliessen führt auf die Fläche darunter zurück.
  schliessen: () => void

  children: ReactNode
}

// **Die Klassen stehen benannt und nicht als Ausdruck im JSX** — dieselbe
// Regel wie in der Seitenleiste. Die gewählte Lasche trägt eine Unterlinie in
// der Markenfarbe, wie im Entwurf; seit AP-3 (#2360) ist das die Gestalt jeder
// Lasche (`bausteine/laschenstil.ts`), und die Farbe steht in `KONTRASTE.md`.
const leistenklassen = 'flex gap-1 border-b border-kante px-3 pt-1'

const hinweisklassen = 'px-4 pt-3 text-sm'

// **`@container`**: die Seiten darin fragen mit `@3xl:` und `@2xl:` ihren
// Platz und nicht das Fenster — sonst stünden sie in 460 Punkten zweispaltig.
// **Weit für das Symbol** — AP-12 (#2369), B-26: 56rem lassen der
// Zeichenfläche 32rem und der Teilliste ihre Spalte daneben. MSE2 öffnet
// den Symboleditor in einem Fenster von 650 × 600 (`window.cpp` 68).
const weit = '56rem'

const feldklassen =
  '@container px-4 pb-5 pt-3 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Setschublade({
  schublade,
  aktuell,
  wechseln,
  schliessen,
  children,
}: SetschubladenEigenschaften) {
  const texte = useTexte()
  const worte = texte.allgemein.arbeitsplatz.schubladen

  const leiste = useRef<HTMLDivElement>(null)
  const laschen = useId()
  const feld = useId()

  // **Der Wunsch gilt, solange die Schublade offen ist** — wie bei MSE2, das
  // die Größe des Symbolfensters nicht merkt. Eine andere Lasche bleibt
  // schmal; zurück beim Symbol ist er wieder da.
  const [weitGewuenscht, setzeWeitGewuenscht] = useState(false)

  // aufTaste bewegt den Fokus zwischen den Laschen der Schublade — das Muster,
  // das ein Vorleseprogramm bei `role="tablist"` erwartet.
  const aufTaste = (ereignis: React.KeyboardEvent) => {
    const richtung =
      ereignis.key === 'ArrowRight' ? 1 : ereignis.key === 'ArrowLeft' ? -1 : 0

    if (richtung === 0) {
      return
    }

    ereignis.preventDefault()

    const naechste = nachbarlasche(schublade.laschen, aktuell, richtung)

    /* v8 ignore next 3 -- der Aufrufer öffnet die Schublade, in der die Seite
       steht (`schubladeVon`); einen Nachbarn gibt es deshalb immer. TypeScript
       sieht es nicht. */
    if (naechste === undefined) {
      return
    }

    wechseln(naechste)

    leiste.current
      ?.querySelector<HTMLElement>(`[data-lasche="${naechste}"]`)
      ?.focus()
  }

  return (
    <Schublade
      titel={worte[schublade.schluessel].titel}
      schliessenBezeichnung={worte.schliessen}
      schliessen={schliessen}
      breite={weitGewuenscht && aktuell === 'symbol' ? weit : undefined}
    >
      <div
        ref={leiste}
        role="tablist"
        aria-label={worte[schublade.schluessel].laschen}
        className={leistenklassen}
      >
        {schublade.laschen.map((schluessel) => {
          const gewaehlt = schluessel === aktuell

          return (
            <button
              key={schluessel}
              type="button"
              role="tab"
              id={`${laschen}-${schluessel}`}
              data-lasche={schluessel}
              aria-selected={gewaehlt}
              aria-controls={feld}
              tabIndex={gewaehlt ? 0 : -1}
              // **Hier landet der Fokus beim Öffnen** — Mantine sucht
              // `data-autofocus` vor allem anderen (`use-focus-trap.mjs`,
              // Zeile 9). Sonst nähme es das ×, das erste Bedienelement.
              data-autofocus={gewaehlt ? true : undefined}
              onClick={() => {
                wechseln(schluessel)
              }}
              onKeyDown={aufTaste}
              className={lasche}
            >
              {beschriftung(texte, schluessel)}
            </button>
          )
        })}
      </div>

      {/* **Der Satz sagt, warum es hier steht** — wie im Entwurf, unter den
          Laschen und über jeder von ihnen gleich. */}
      <p className={hinweisklassen}>{worte[schublade.schluessel].hinweis}</p>

      <div
        id={feld}
        role="tabpanel"
        aria-labelledby={`${laschen}-${aktuell}`}
        tabIndex={0}
        className={feldklassen}
      >
        <SchubladenweiteKontext
          value={{
            weit: weitGewuenscht,
            umschalten: () => {
              setzeWeitGewuenscht((war) => !war)
            },
          }}
        >
          {children}
        </SchubladenweiteKontext>
      </div>
    </Schublade>
  )
}
