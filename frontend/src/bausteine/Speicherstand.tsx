// **Ein Satz, ob das Eingetippte im Bestand steht** — UX-05.AK-03
// (#1677).
//
// **`status` und nicht `alert`**: ein Vorleser sagt ihn, sobald er Luft
// holt. Der Fehler selbst steht am Feld (§6) und ist dort ein `alert`; hier
// steht nur, **dass** etwas nicht geschrieben ist und wo.
//
// **Nichts, solange nichts geschrieben wurde** — ein „alles gespeichert"
// über einer Karte, die niemand angefasst hat, wäre eine Behauptung ohne
// Anlass.
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useSpeicherlage } from '../speicherstand'

export function Speicherstand({ setID }: { setID: string }) {
  const eigene = useTexte().bausteine.speicherstand
  const lage = useSpeicherlage(setID)

  if (lage.art === 'still') {
    return null
  }

  return (
    <p role="status" className="text-sm text-leise">
      {lage.art === 'schreibt'
        ? eigene.schreibt
        : lage.art === 'geschrieben'
          ? eigene.geschrieben
          : einsetzen(eigene.fehler, { feld: lage.feld })}
    </p>
  )
}
