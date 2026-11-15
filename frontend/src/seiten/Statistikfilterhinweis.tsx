// **Die Zeile über der Kartenliste, wenn die Statistik sie filtert** — UXR-F6
// (#2013).
//
// # Wofür
//
// In MSE2 steht die gefilterte Liste unter dem Diagramm, und was sie filtert,
// ist an der Wahl im Diagramm zu sehen. Hier steht sie auf den Karten, weit
// weg vom Diagramm — und ohne diese Zeile wüsste niemand, warum nur 41 von 312
// Karten dastehen. Sie nennt den Filter, zählt, was er übrig lässt, und zeigt
// die zwei Wege: **zurück zur Statistik** (die Wahl bleibt, dort steht der
// Balken noch markiert) und **aufheben**.
//
// **Die Zahl kommt aus der Liste, die den Filter anwendet** und nicht aus der
// Wahl: was Marken und Schnellfilter zusätzlich wegnehmen, zählt mit — dieselbe
// Zusage wie beim `Schnellfilter`.
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { Statistikfilter } from '../statistikfilter'
import { ohneAuszeichnung } from './ohneauszeichnung'

interface HinweisEigenschaften {
  filter: Statistikfilter

  // gesamt und gezeigt füttern die Trefferzahl.
  gesamt: number
  gezeigt: number
}

export function Statistikfilterhinweis({
  filter,
  gesamt,
  gezeigt,
}: HinweisEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.statistik

  // Ohne Wahl steht keine Zeile da.
  if (filter.wahl === undefined) {
    return null
  }

  return (
    <div
      role="group"
      aria-label={eigene.filterBezeichnung}
      className="mb-2 flex flex-col gap-1 rounded-md border border-kante bg-betont px-2 py-1.5 text-sm"
    >
      <p>
        {einsetzen(eigene.filterKopf, {
          quelle: filter.wahl.quelle,
          gruppe: ohneAuszeichnung(filter.wahl.gruppe),
        })}
      </p>

      <p role="status" className="text-leise">
        {einsetzen(texte.seiten.karten.schnellfilter.treffer, {
          gezeigt: String(gezeigt),
          gesamt: String(gesamt),
        })}
      </p>

      <div className="flex flex-wrap gap-2">
        <Knopf gewicht="neben" klein aufKlick={filter.zurStatistik}>
          {eigene.filterZurStatistik}
        </Knopf>

        <Knopf gewicht="leise" klein aufKlick={filter.aufheben}>
          {eigene.filterAufheben}
        </Knopf>
      </div>
    </div>
  )
}
