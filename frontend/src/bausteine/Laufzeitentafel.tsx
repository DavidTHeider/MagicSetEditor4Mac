// **Die Laufzeiten je Fläche und Art** — OPS-05.AK-01
// (#1722):
// Rufe, Fehlerrate, Mittelwert und die Verteilung über die Eimer.
//
// **Die Fehlerrate zählt Abgewiesenes und Gescheitertes zusammen** — wer
// die beiden trennen will, findet sie in `GET /metriken`; hier soll eine
// Zahl sagen, ob ein Weg klemmt.
import { useDarstellung, useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { alsProzent } from '../lokal'
import { alsDauer } from './dauer'

import type { MetrikenAntwort } from '../apimetriken'

export function Laufzeitentafel({ zahlen }: { zahlen: MetrikenAntwort }) {
  const technik = useTexte().seiten.technik
  const eigene = technik.betrieb
  const darstellung = useDarstellung()

  if (zahlen.laufzeiten.length === 0) {
    return <p className="text-sm">{eigene.laufzeitenLeer}</p>
  }

  const flaechen = technik.flaechen as Record<string, string>
  const arten = eigene.arten as Record<string, string>
  const eimer = [
    ...zahlen.eimerMs.map((grenze) =>
      einsetzen(eigene.eimer, { grenze: String(grenze) }),
    ),
    eigene.eimerOffen,
  ]

  return (
    <table className="w-full max-w-3xl text-sm">
      <caption className="sr-only">{eigene.laufzeiten}</caption>

      <thead>
        <tr className="text-left">
          <th scope="col">{technik.spalteFlaeche}</th>
          <th scope="col">{eigene.spalteArt}</th>
          <th scope="col" className="text-right">
            {eigene.spalteAnzahl}
          </th>
          <th scope="col" className="text-right">
            {eigene.spalteFehler}
          </th>
          <th scope="col" className="text-right">
            {eigene.spalteMittel}
          </th>
          <th scope="col">{eigene.spalteVerteilung}</th>
        </tr>
      </thead>

      <tbody>
        {zahlen.laufzeiten.map((reihe) => (
          <tr key={`${reihe.flaeche}/${reihe.art}`}>
            <td>{flaechen[reihe.flaeche] ?? reihe.flaeche}</td>
            <td>{arten[reihe.art] ?? reihe.art}</td>
            <td className="text-right tabular-nums">
              {reihe.anzahl.toLocaleString(darstellung)}
            </td>
            <td className="text-right tabular-nums">
              {alsProzent(
                (reihe.abgewiesen + reihe.gescheitert) / reihe.anzahl,
                darstellung,
              )}
            </td>
            <td className="text-right tabular-nums">
              {alsDauer(reihe.summeMs / reihe.anzahl, darstellung)}
            </td>
            <td className="text-leise">
              {eimer
                .map((name, stelle) => ({
                  name,
                  anzahl: reihe.eimer[stelle] ?? 0,
                }))
                .filter(({ anzahl }) => anzahl > 0)
                .map(({ anzahl, name }) =>
                  einsetzen(eigene.verteilung, {
                    eimer: name,
                    anzahl: anzahl.toLocaleString(darstellung),
                  }),
                )
                .join(', ')}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
