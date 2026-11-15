// **Unsichtbare Marken der Zonenfelder** — A04
// (#2218).
//
// Warum es sie gibt und was sie nicht dürfen, steht in
// `tischzonenmarken.test.tsx`: sie nennen einer Probe den Ort einer Zone in
// Punkten der Seite — **ohne Zeiger, ohne Stimme**.
import { nachPunkten } from '../tisch/kamera'

import type { Kamera } from '../tisch/kamera'
import type { Zonenfeld } from '../tisch/zonenfeld'

// rahmen rechnet ein Feld auf seinen Platz über der Leinwand — dieselbe
// Rechnung wie `stelle` für eine Karte.
function rahmen(
  feld: Zonenfeld,
  kamera: Kamera,
  masse: { breite: number; hoehe: number },
): React.CSSProperties {
  const mitte = nachPunkten(
    kamera,
    { x: feld.x + feld.breite / 2, y: feld.y + feld.hoehe / 2 },
    masse,
  )
  const breite = feld.breite * kamera.massstab
  const hoehe = feld.hoehe * kamera.massstab

  return {
    left: `${String(mitte.x - breite / 2)}px`,
    top: `${String(mitte.y - hoehe / 2)}px`,
    width: `${String(breite)}px`,
    height: `${String(hoehe)}px`,
    transform: `rotate(${String(kamera.drehung)}deg)`,
  }
}

export function Tischzonenmarken({
  felder,
  kamera,
  masse,
}: {
  felder: readonly Zonenfeld[]
  kamera: Kamera
  masse: { breite: number; hoehe: number }
}) {
  return (
    <div aria-hidden="true" className="pointer-events-none absolute inset-0">
      {felder.map((feld) => (
        <div
          key={`${feld.sitz}/${feld.zone}`}
          data-zonenfeld={feld.zone}
          data-sitz={feld.sitz}
          aria-hidden="true"
          className="pointer-events-none absolute"
          style={rahmen(feld, kamera, masse)}
        />
      ))}
    </div>
  )
}
