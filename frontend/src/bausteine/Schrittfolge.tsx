// **Die Schritte eines Assistenten** — UX-08.AK-01
// (#1680):
// *„Wizard für zusammenhängende Mehrschrittaufgaben"*. Nach dem Artboard
// *Sets* („Ein neues Set — in drei Schritten").
//
// # Was sie zeigt, und was nicht
//
// **Wo man steht, und was davor und danach kommt** — eine geordnete Liste,
// der aktuelle Schritt mit `aria-current="step"`. **Sie führt nicht**: die
// Schritte sind keine Knöpfe, denn wer zurückspringt, soll es mit „Zurück"
// tun, das dieselbe Eingabe behält, und nicht über eine Abkürzung, die einen
// Schritt überspränge, ohne dass jemand es merkt.
interface SchrittfolgeEigenschaften {
  // bezeichnung nennt die Folge — ein Vorleser sagt sie vor den Schritten.
  bezeichnung: string

  schritte: readonly string[]

  // aktuell ist die Stelle des Schritts, an dem gearbeitet wird.
  aktuell: number
}

// Die Klassen eines Schritts — **fertige Zeichenketten und keine
// zusammengesetzten**: was welcher Zustand bekommt, steht so auf einen Blick.
const zeile = {
  aktuell: 'flex items-center gap-2 text-sm font-medium text-text',
  sonst: 'flex items-center gap-2 text-sm text-leise',
}

const kreis = {
  erreicht:
    'inline-flex h-5 w-5 items-center justify-center rounded-full border border-marke bg-marke text-xs text-marketext',
  offen:
    'inline-flex h-5 w-5 items-center justify-center rounded-full border border-kante text-xs',
}

export function Schrittfolge({
  bezeichnung,
  schritte,
  aktuell,
}: SchrittfolgeEigenschaften) {
  return (
    <ol aria-label={bezeichnung} className="flex flex-wrap items-center gap-2">
      {schritte.map((schritt, stelle) => (
        <li
          key={schritt}
          aria-current={stelle === aktuell ? 'step' : undefined}
          className={stelle === aktuell ? zeile.aktuell : zeile.sonst}
        >
          <span
            aria-hidden="true"
            className={stelle <= aktuell ? kreis.erreicht : kreis.offen}
          >
            {stelle + 1}
          </span>
          {schritt}
        </li>
      ))}
    </ol>
  )
}
