// **Wie viele Stück jedes Drucks jemand behält** — UXR-F18
// (#2033).
//
// # Wofür
//
// Der Umfang „Dubletten" einer Sichtfreigabe sagt, **was über der
// Behaltemenge liegt** (`Sammlungsfreigabe.Zeigt`: `Menge - Behalten`, und nur
// wenn das mehr als null ist). Die Zahl steht an der Freigabe und nicht am
// Bestand (0057): einem Gegenüber sagt jemand, was er behält, einem anderen
// vielleicht etwas anderes. Das Fenster schickte bis dahin immer 0 — und mit 0
// ist jedes Stück eine Dublette, „Dubletten" gäbe also die ganze Sammlung
// frei.
//
// # Warum die Vorgabe 1 ist
//
// **Weil es dieselbe ist wie überall, wo das Programm nach Dubletten fragt**:
// `mse dubletten` (Vorgabe 1, `cmd/mse/sammlungssichten.go`) und die Ordnung
// der Sammlung im Fenster (`Sammlungsordnung`). Das eine Stück, das man
// behalten will, ist der Normalfall (0057: „meist 1"); wer mehr oder weniger
// meint, stellt es um.
//
// # Warum ein `Zahlenfeld`
//
// **Weil die Domäne zwei Regeln kennt**: `Behalten < 0` ist ein Befund, und
// 0 heißt die Vorgabe 1 (`Sammlungsfreigabe.MitVorgaben`); eine Obergrenze
// gibt es nicht. Das `Zahlenfeld` nimmt deshalb nur ganze Zahlen **ab eins**
// an — ein Feld, das 0 annähme, zeigte etwas anderes, als erteilt wird — und
// sagt am Feld, was nicht passt; ein Befund des Dienstes tritt als `befund`
// neben seine eigene Prüfung.
import { useId } from 'react'

import { Zahlenfeld } from './Zahlenfeld'
import { useTexte } from '../sprache/kontext'

interface Behaltemengeneigenschaften {
  wert: number
  geaendert: (wert: number) => void

  // gilt meldet bei jeder Eingabe, ob der getippte Text abgeschickt werden
  // darf — der Knopf daneben hält sich daran (siehe [Zahlenfeld]).
  gilt: (gilt: boolean) => void

  // befund ist der Satz des Dienstes zu diesem Feld.
  befund?: string | undefined
  angehalten?: boolean
}

export function Behaltemenge({
  wert,
  geaendert,
  gilt,
  befund,
  angehalten = false,
}: Behaltemengeneigenschaften) {
  const worte = useTexte().bausteine.druckwahl
  const hinweisID = useId()

  return (
    <div className="flex max-w-xs flex-col gap-1">
      <Zahlenfeld
        beschriftung={worte.behalten}
        wert={wert}
        kleinstens={1}
        ganzzahlig
        schritt={1}
        angehalten={angehalten}
        beschreibung={hinweisID}
        befund={befund}
        geaendert={geaendert}
        gilt={gilt}
      />

      <p id={hinweisID} className="text-sm text-leise">
        {worte.behaltenHinweis}
      </p>
    </div>
  )
}
