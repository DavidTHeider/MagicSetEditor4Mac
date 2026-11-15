// **Eine Zeile der Liste „Was gerade läuft"** — TRADE-01
// (#1715).
//
// **Ein Knopf je Handel** und kein Radio: ein Klick schlägt ihn auf, wie im
// Artboard *Handeln*. Wer dran ist, sieht es in der ersten Zeile — farbig
// und in Worten, denn Farbe allein sagt niemandem etwas, der sie nicht sieht.
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { lagesatz, partnername } from './handelslage'

import type { Handelsvorgangzeile } from '../apihandelsverzeichnis'

interface Handelslistenzeileneigenschaften {
  zeile: Handelsvorgangzeile
  gewaehlt: boolean
  // oeffnen fehlt, wo es nichts aufzuschlagen gibt — dann ruht der Knopf.
  oeffnen: (() => void) | undefined
}

const marke =
  'rounded-sm border border-kante bg-flaeche px-1.5 text-xs text-leise'

export function Handelslistenzeile({
  zeile,
  gewaehlt,
  oeffnen,
}: Handelslistenzeileneigenschaften) {
  const handeln = useTexte().seiten.handeln
  const eigene = handeln.liste
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()

  return (
    <button
      type="button"
      aria-current={gewaehlt ? 'true' : undefined}
      disabled={oeffnen === undefined}
      onClick={oeffnen}
      className="flex w-full flex-col gap-0.5 rounded-md border border-kante px-2 py-1.5 text-left aria-[current=true]:border-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
    >
      <span className="flex flex-wrap items-center gap-1.5 font-medium">
        {partnername(zeile)}
        {zeile.vorgang.zustand === 'entwurf' && (
          <span className={marke}>{eigene.entwurf}</span>
        )}
        {zeile.unbestaetigt && (
          <span className={marke} title={eigene.unbestaetigtErklaerung}>
            {eigene.unbestaetigt}
          </span>
        )}
        {zeile.archiviert && <span className={marke}>{eigene.archiviert}</span>}
      </span>

      <span
        className={
          zeile.lage === 'dran' ? 'text-sm text-marke' : 'text-sm text-leise'
        }
      >
        {lagesatz(zeile, handeln)}
      </span>

      <span className="truncate text-xs text-leise">
        {zeile.karten.length === 0
          ? eigene.ohneKarten
          : zeile.karten.join(', ')}
        {zeile.geaendert !== '' &&
          ` · ${einsetzen(eigene.geaendert, { wann: alsZeitpunkt(zeile.geaendert, darstellung, zeitzone) })}`}
      </span>
    </button>
  )
}
