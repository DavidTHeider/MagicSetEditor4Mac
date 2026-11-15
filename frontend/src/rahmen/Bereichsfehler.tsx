// **Die Fläche, wenn ein Bereich nicht zeichnet** — F01
// (#2340),
// nach dem Artboard `Betrieb09` (`docs/PROTOTYP-0.9.0.md`).
//
// # Was sie sagt und anbietet
//
// **Ein Satz, eine Kennung, zwei Rückwege**: den Bereich neu zeichnen, und die
// Diagnose — **hier an Ort und Stelle** und nicht als Sprung in die
// Einstellungen: ist dort der Fehler, führte der Sprung in denselben.
// Die Kennung steht mit derselben Zeichenfolge im Protokoll, das das
// Diagnosepaket einsammelt.
//
// **Der Fokus geht auf den ersten Rückweg** — wer mit der Tastatur arbeitet,
// stünde sonst auf einem Knoten, den es nicht mehr gibt.
import { useEffect, useId, useRef, useState } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { Diagnosepaket } from '../seiten/Diagnosepaket'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Fehlerstand } from './Fehlergrenze'

export function Bereichsfehler({ kennung, neu }: Fehlerstand) {
  const eigene = useTexte().allgemein.fehlergrenze
  const titel = useId()
  const rueckweg = useRef<HTMLButtonElement>(null)
  const [diagnose, setzeDiagnose] = useState(false)

  useEffect(() => {
    rueckweg.current?.focus()
  }, [])

  return (
    <section
      role="alert"
      aria-labelledby={titel}
      className="flex max-w-prose flex-col gap-3 rounded-sm border border-kante bg-flaeche p-4"
    >
      <h2 id={titel} className="text-base font-semibold">
        {eigene.bereichTitel}
      </h2>

      <p className="text-sm">{einsetzen(eigene.bereichSatz, { kennung })}</p>

      <div className="flex flex-wrap gap-3">
        <Knopf gewicht="haupt" griff={rueckweg} aufKlick={neu}>
          {eigene.neuLaden}
        </Knopf>

        <Knopf
          gewicht="leise"
          ausgeklappt={diagnose}
          aufKlick={() => {
            setzeDiagnose((offen) => !offen)
          }}
        >
          {eigene.diagnoseOeffnen}
        </Knopf>
      </div>

      {diagnose && <Diagnosepaket />}
    </section>
  )
}
