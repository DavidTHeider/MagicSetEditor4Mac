// **Die Rangliste** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Eine Tabelle mit Kopfzellen**, weil sie eine ist: acht Spalten, die nur
// zusammen etwas sagen. Eine Liste aus Sätzen wäre für eine Vorlesehilfe
// schlechter, nicht besser.
//
// **Bei einem Einzelspielturnier fehlen zwei Spalten** — MTR §3.1 hält fest, dass nicht
// jeder dieser Tiebreaker in Formaten mit einspieligen Matches gilt.
// Gerechnet werden sie trotzdem; hier entscheidet nur die Anzeige.
import { useTexte } from '../sprache/kontext'
import type { Turnierstand } from '../haken/useTurniere'

export function Turnierrangliste({
  stand,
  einzelspiel,
}: {
  stand: Turnierstand
  einzelspiel: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  if (stand.rangliste.length === 0) {
    return (
      <div className="flex flex-col gap-2">
        <h3 className="font-semibold">{eigene.rangliste}</h3>

        <p>{eigene.ranglisteLeer}</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <h3 className="font-semibold">{eigene.rangliste}</h3>

      <table className="w-full text-left">
        <thead>
          <tr>
            <th scope="col">{eigene.platz}</th>
            <th scope="col">{eigene.name}</th>
            <th scope="col">{eigene.bilanz}</th>
            <th scope="col">{eigene.matchpunkte}</th>
            <th scope="col">{eigene.gegnermatchquote}</th>

            {!einzelspiel && (
              <>
                <th scope="col">{eigene.spielquote}</th>
                <th scope="col">{eigene.gegnerspielquote}</th>
              </>
            )}
          </tr>
        </thead>

        <tbody>
          {stand.rangliste.map((rang) => (
            <tr key={rang.teilnehmerID}>
              <td>{rang.platz}</td>

              <td>
                {rang.name}
                {rang.ausgestiegen && ` — ${eigene.ausgestiegen}`}
              </td>

              <td>
                {rang.siege + rang.byes}–{rang.niederlagen}–{rang.unentschieden}
              </td>

              <td>{rang.matchpunkte}</td>
              <td>{rang.gegnermatchquote.toFixed(4)}</td>

              {!einzelspiel && (
                <>
                  <td>{rang.spielquote.toFixed(4)}</td>
                  <td>{rang.gegnerspielquote.toFixed(4)}</td>
                </>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
