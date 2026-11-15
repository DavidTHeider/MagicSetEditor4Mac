// **Die Wahl des Erhaltungszustands** — Werkstück 2 von Phase 3.4
// (#1165).
//
// # Warum die Namen englisch sind und die Beschriftung übersetzt
//
// Die Skala ist die von **Cardmarket** und damit ein fremder Begriff (§9): der
// Wert, der in den Bestand geht, heißt `nearMint`. Was daneben steht, ist ein
// Text aus dem Katalog — und der darf deutsch sein.
import { useTexte } from '../sprache/kontext'

// zustaende ist die Aufzählung in ihrer fachlichen Reihenfolge.
//
// **Beste zuerst**, wie in `domain.Zustaende()`. Eine alphabetische Liste
// stellte `excellent` vor `mint`, und niemand fände das Naheliegende zuerst.
// zustaende sind die sieben aus `domain.Zustand` — **ausgeführt**, weil die
// Handelsposition dieselben wählt (UX-06.AK-01, #1678).
export const zustaende = [
  'mint',
  'nearMint',
  'excellent',
  'good',
  'lightPlayed',
  'played',
  'poor',
] as const

interface ZustandswahlEigenschaften {
  wert: string
  setzeWert: (wert: string) => void
}

export function Zustandswahl({ wert, setzeWert }: ZustandswahlEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm font-medium">{eigene.zustand}</span>
      <select
        value={wert}
        onChange={(ereignis) => {
          setzeWert(ereignis.target.value)
        }}
        className="rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        {/* **Der leere Eintrag steht oben und heißt nichts.** Ohne Angabe
            setzt die Domäne `nearMint`; ihn hier vorwegzunehmen hieße, die
            Vorgabe an zwei Stellen zu führen. */}
        <option value="">—</option>
        {zustaende.map((zustand) => (
          <option key={zustand} value={zustand}>
            {eigene.zustaende[zustand]}
          </option>
        ))}
      </select>
    </label>
  )
}
