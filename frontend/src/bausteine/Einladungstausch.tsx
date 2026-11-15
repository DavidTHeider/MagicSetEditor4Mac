// **Jemanden einladen** — NET-01
// (#1705),
// nach dem Artboard *Gegenüber*: „Weiter weg? Dann tauscht ihr einen
// Einladungscode aus." Zwei Knöpfe, und jeder öffnet seinen Schritt.
//
// # Warum die zwei Knöpfe aufklappen und nicht eingedrückt werden
//
// **UXR-16** (#1924):
// ein gedrückter Knopf trägt das Hauptgewicht (`Knopf`, `gedrueckt`), und
// darunter stand „Code erzeugen" im Hauptgewicht — zwei gefüllte Knöpfe in
// einer Region (R1 in `docs/UIREVIEW.md`), und welcher führt, sagte keiner.
// Die beiden öffnen je einen Schritt; das ist ein Aufklappen
// (`aria-expanded`), und der Schritt darunter trägt die eine Handlung.
//
// # Warum hier kein Name vorgegeben wird
//
// **Weil er hinausgeht.** Der Name in einer Einladung liest jeder, dem sie
// weitergereicht wird; welcher das sein soll, entscheidet ein Mensch — und
// ein vorausgefüllter wäre einer, den niemand gewählt hat.
import { useId, useState } from 'react'

import { Einladungsannahme } from './Einladungsannahme'
import { Einladungsausgabe } from './Einladungsausgabe'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Knopf } from './Knopf'
import { useEinladungstausch } from '../haken/useEinladungstausch'
import { useTexte } from '../sprache/kontext'

type Schritt = 'zu' | 'zeigen' | 'eingeben'

interface Einladungstauscheigenschaften {
  // zwecke kommen aus Go, wie in der Liste daneben.
  zwecke: string[]

  // nachAufnahme lädt die Liste daneben nach (§7, Regel 3).
  nachAufnahme: () => void
}

export function Einladungstausch({
  zwecke,
  nachAufnahme,
}: Einladungstauscheigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik.einladung
  const tausch = useEinladungstausch(nachAufnahme)
  const [schritt, setzeSchritt] = useState<Schritt>('zu')
  const schrittID = useId()

  const umschalten = (ziel: Schritt) => {
    setzeSchritt(schritt === ziel ? 'zu' : ziel)
  }

  return (
    <section
      aria-label={eigene.titel}
      data-einladungstausch=""
      className="flex flex-col gap-3 rounded-sm border border-kante p-3"
    >
      <Erklaerungsabsatz
        text={eigene.erklaerung}
        klasse="max-w-prose text-sm"
      />

      <div className="flex flex-wrap gap-2">
        <Knopf
          ausgeklappt={schritt === 'zeigen'}
          steuert={schrittID}
          aufKlick={() => {
            umschalten('zeigen')
          }}
        >
          {eigene.zeigen}
        </Knopf>
        <Knopf
          ausgeklappt={schritt === 'eingeben'}
          steuert={schrittID}
          aufKlick={() => {
            umschalten('eingeben')
          }}
        >
          {eigene.eingeben}
        </Knopf>
      </div>

      {schritt !== 'zu' && (
        <div id={schrittID}>
          {schritt === 'zeigen' ? (
            <Einladungsausgabe zwecke={zwecke} tausch={tausch} />
          ) : (
            <Einladungsannahme tausch={tausch} />
          )}
        </div>
      )}

      {tausch.meldung !== undefined && (
        <p role="status" className="text-sm">
          {tausch.meldung}
        </p>
      )}
    </section>
  )
}
