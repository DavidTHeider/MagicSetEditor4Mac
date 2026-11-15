// **„Deck anlegen" führt den Navigator** — UXR-11
// (#1919).
//
// Das Artboard *Decks* setzt an den Kopf des Navigators einen Hauptknopf
// „+ Deck anlegen" — die eine Handlung, die diese Spalte führt. Bis UXR-11
// stand dort ein leiser Aufklapper wie vor jedem anderen Werkzeug, und der
// Navigator hatte gar keine führende Handlung.
//
// # Warum kein `Aufklappkasten`
//
// **Weil der Knopf sein Gewicht abgibt, sobald er offen ist.** Dann führt
// „Anlegen" im Formular darunter, und zwei Hauptknöpfe übereinander sagten
// nicht mehr, welcher dran ist (R1 in `docs/UIREVIEW.md`). Der
// [Aufklappkasten] hat dafür keinen Schalter, und viele Seiten benutzen ihn;
// das Verhalten darüber hinaus ist seines: einmal gezeichnet, bleibt der
// Inhalt stehen und wird nur verborgen — sonst verlöre, wer zuklappt, was er
// eingegeben hat.
//
// # Und solange kein Deck offen ist
//
// **Im offenen Deck führt „Ins Deck“** — AP-3 (#2360), Entwurf `Bereiche09`;
// je Blick führt einer (`Zustaende09`). Dann steht „Deck anlegen“ im
// Nebengewicht, wo er bleibt.
import { useEffect, useId, useRef, useState } from 'react'

import { Deckanlage } from '../bausteine/Deckanlage'
import { Deckausset } from '../bausteine/Deckausset'
import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'

import type { Deckstand } from '../haken/useDecks'

export function Deckanlagekasten({ stand }: { stand: Deckstand }) {
  const eigene = useTexte().seiten.decks
  const [offen, setzeOffen] = useState(false)
  const [gezeichnet, setzeGezeichnet] = useState(false)
  const kastenID = useId()
  const kasten = useRef<HTMLDivElement | null>(null)

  // **Wer aufklappt, will einen Namen tippen** — AP-6 (#2363): der Klick
  // ins Namensfeld war eine der Handlungen über dem Ziel des Kernwegs
  // `deck`. Nur beim Aufgehen, nicht beim ersten Zeichnen: der Kasten
  // beginnt zu.
  useEffect(() => {
    if (offen) {
      kasten.current?.querySelector('input')?.focus()
    }
  }, [offen])

  return (
    <section className="flex flex-col gap-2">
      <Knopf
        gewicht={offen ? 'leise' : stand.offen === null ? 'haupt' : 'neben'}
        ausgeklappt={offen}
        steuert={kastenID}
        aufKlick={() => {
          setzeOffen(!offen)
          setzeGezeichnet(true)
        }}
      >
        {eigene.neu}
      </Knopf>

      {gezeichnet && (
        <div
          ref={kasten}
          id={kastenID}
          hidden={!offen}
          className="rounded-sm border border-kante p-3"
        >
          <Deckanlage stand={stand} />
          <Deckausset stand={stand} />
        </div>
      )}
    </section>
  )
}
