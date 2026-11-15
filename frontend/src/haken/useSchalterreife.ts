// **Die Reife der Schalter** — OPS-10.AK-04
// (#1908).
//
// **Aus Go und nicht von hier**: das Verzeichnis steht in
// `internal/domain/schalter.go`, und eine Funktion, die als Vorschau
// ausgeliefert wird, soll ihre Marke bekommen, ohne dass hier eine Zeile
// dazukommt. **Ein Fehler lässt sie weg** — eine fehlende Marke ist besser
// als eine Seitenleiste, die nicht steht.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

async function holen(): Promise<Readonly<Record<string, string>>> {
  const antwort = await api().Schalter()

  return auspacken(antwort, () =>
    Object.fromEntries(
      antwort.schalter.map((einer) => [einer.kennung, einer.reife]),
    ),
  )
}

// useSchalterreife gibt je Kennung die Reife — leer, solange nichts da ist.
export function useSchalterreife(): Readonly<Record<string, string>> {
  return useLaden(holen, []).daten ?? {}
}
