// **Kommandozeile und Schnittstellen sehen einander** — OPS03
// (#2349).
//
// Die Verknüpfungen in `~/.local/bin` stehen seither in **beiden** Blöcken:
// als *Kommandozeile einrichten* (OPS-03, #1720) und als Ziel `befehle` unter
// den Schnittstellen. Wer sie im einen ändert, soll im anderen nicht den
// Stand von vorhin sehen.
//
// # Wie
//
// **Über einen `key`, der beim anderen zählt** (§7, Regel 1): schreibt der
// eine Block, baut sich der andere neu auf und liest frisch. Der schreibende
// behält seinen Stand — samt den Ausgängen von „Alles zurücknehmen", die
// sonst mit dem Neuaufbau verschwänden.
import { useState } from 'react'

import { Befehlszeile } from './Befehlszeile'
import { Schnittstellen } from './Schnittstellen'

export function Werkzeugeinrichtung() {
  const [befehleGeschrieben, setzeBefehleGeschrieben] = useState(0)
  const [schnittstellenGeschrieben, setzeSchnittstellenGeschrieben] =
    useState(0)

  return (
    <>
      {/* **Negativ gezählt**, damit kein Schlüssel dem des Nachbarn gleicht. */}
      <Befehlszeile
        key={-1 - schnittstellenGeschrieben}
        nachSchreiben={() => {
          setzeBefehleGeschrieben((zahl) => zahl + 1)
        }}
      />
      <Schnittstellen
        key={befehleGeschrieben}
        nachSchreiben={() => {
          setzeSchnittstellenGeschrieben((zahl) => zahl + 1)
        }}
      />
    </>
  )
}
