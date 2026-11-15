// **Leben am Sitz** — P4
// (#2372),
// Befund C-20: Lebenspunkte gingen nur über ein Formular in der Spalte.
//
// # Ändern und nicht setzen
//
// **Jede Wahl geht als Änderung hinaus** — `spieler.leben` mit `um`, wie der
// Kern es verlangt: „drei weniger“ übersteht eine Antwort, die sich mit einer
// anderen kreuzt, „siebzehn“ überschriebe sie (§7, Regel 2). Auch „Wert
// setzen“ rechnet die Differenz aus dem Stand, den es zeigt.
//
// **⇧ ändert um fünf** — der Vorschlag aus C-20. **Nur am eigenen Sitz**:
// der Kern ändert das Leben des Absenders, und die Regel bleibt dort.
import { useState, type MouseEvent } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'
import { nebenknopf } from './knopfklassen'
import { Popover } from './Popover'
import { Zahlenfeld } from './Zahlenfeld'

const schritt = (ereignis: MouseEvent) => (ereignis.shiftKey ? 5 : 1)

// wertklasse ist der Knopf mit dem Stand — Ziffern gleicher Breite, damit
// er beim Zählen nicht springt.
const wertklasse = `${nebenknopf} tabular-nums`

export function Lebensregler({
  wert,
  aendere,
}: {
  wert: number
  aendere: (um: number) => void
}) {
  const eigene = useTexte().seiten.spielen
  const [offen, setzeOffen] = useState(false)
  const [neu, setzeNeu] = useState(wert)

  const schliessen = () => {
    setzeOffen(false)
  }

  return (
    <span className="inline-flex items-center gap-1">
      <button
        type="button"
        aria-label={eigene.lebenWeniger}
        className={nebenknopf}
        onClick={(ereignis) => {
          aendere(-schritt(ereignis))
        }}
      >
        {eigene.lebenMinus}
      </button>

      <Popover
        offen={offen}
        titel={eigene.lebenSetzenTitel}
        schliessen={schliessen}
        ausloeser={
          <button
            type="button"
            aria-label={einsetzen(eigene.lebenSetzen, { stand: String(wert) })}
            className={wertklasse}
            onClick={() => {
              setzeNeu(wert)
              setzeOffen(true)
            }}
          >
            {wert}
          </button>
        }
      >
        <Zahlenfeld
          beschriftung={eigene.lebenNeu}
          wert={neu}
          kleinstens={-999}
          ganzzahlig
          geaendert={setzeNeu}
        />
        <Knopf
          gewicht="haupt"
          aufKlick={() => {
            if (neu !== wert) {
              aendere(neu - wert)
            }
            schliessen()
          }}
        >
          {eigene.lebenUebernehmen}
        </Knopf>
      </Popover>

      <button
        type="button"
        aria-label={eigene.lebenMehr}
        className={nebenknopf}
        onClick={(ereignis) => {
          aendere(schritt(ereignis))
        }}
      >
        {eigene.lebenPlus}
      </button>
    </span>
  )
}
