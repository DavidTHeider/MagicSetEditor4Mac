// **Marken an einem Paket** — UX-12 (#1684). Herausgelöst aus
// `Paketzeilen.tsx`, das sonst über die Zeilengrenze aus §2 ginge.
//
// Die Kacheln stehen in der Zeile, **der Editor in einer Zeile darunter** —
// erreicht über das Menü der Zeile, denn im Gitter ist die Zeile die
// Tabstelle und nicht jeder Knopf darin.
import { Objektmarken } from '../bausteine/Objektmarken'
import type { Markenfilterstand } from '../haken/useMarkenfilter'

export function Paketmarken({
  marken,
}: {
  marken: readonly string[] | undefined
}) {
  return (marken ?? []).map((marke) => (
    <span
      key={marke}
      className="ml-2 rounded-full border border-kante px-2 py-0.5 text-xs"
    >
      {marke}
    </span>
  ))
}

export function Paketmarkenzeile({
  kennung,
  marken,
}: {
  kennung: string
  marken: Markenfilterstand
}) {
  return (
    <tr>
      <td colSpan={6} className="py-2">
        <Objektmarken
          art="paket"
          kennung={kennung}
          marken={marken.je[kennung] ?? []}
          geaendert={marken.nachladen}
        />
      </td>
    </tr>
  )
}
