// **Die Ansicht des Zeilen-Diffs** — was sich gegenüber dem gelesenen Stand
// geändert hat, bevor jemand speichert (Werkstück U7 von Phase 5.1,
// #1418).
//
// **Nur die Farben, die `docs/KONTRASTE.md` kennt**: Neues in
// `emerald-800`/`emerald-200`, Entferntes in `red-800`/`red-300` — beide auf
// dem Grund der Seite gemessen. Eine getönte Zeilenfläche wäre ein neues
// Paar, und das braucht eine Messung.
//
// **Die Marke steht als Text für die Sprachausgabe** — `+` und `−` sind für
// die Augen, „neu" und „entfernt" für die Ohren.
import { huenke, zeilendiff } from './diffzeilen'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Diffzeile } from './diffzeilen'

interface ZeilendiffEigenschaften {
  alt: string
  neu: string
}

const farben: Record<Diffzeile['art'], string> = {
  gleich: 'text-leise',
  weg: 'text-fehler',
  neu: 'text-erfolg',
}

const marken: Record<Diffzeile['art'], string> = {
  gleich: ' ',
  weg: '−',
  neu: '+',
}

export function Zeilendiff({ alt, neu }: ZeilendiffEigenschaften) {
  const texte = useTexte().seiten.vorlagen
  const diff = zeilendiff(alt, neu)
  const gruppen = huenke(diff, 1)
  const zaehlung = {
    neu: String(diff.filter((zeile) => zeile.art === 'neu').length),
    weg: String(diff.filter((zeile) => zeile.art === 'weg').length),
  }

  if (gruppen.length === 0) {
    return <p className="text-sm text-leise">{texte.aenderungenKeine}</p>
  }

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm text-leise">
        {einsetzen(texte.aenderungenZusammenfassung, zaehlung)}
      </p>

      {gruppen.map((gruppe, nummer) => (
        // **Eine Region mit Namen und `tabIndex`, weil sie quer rollt** —
        // das Muster aus `Diagnose.tsx` (Zweitblick über Phase 6.3, G10).
        <div
          // **Die Nummer als Schlüssel**: Hünke werden nicht umsortiert, und
          // ihre erste Zeile kann auf beiden Seiten fehlen.
          key={nummer}
          role="region"
          aria-label={einsetzen(texte.aenderungenAbschnitt, {
            nummer: String(nummer + 1),
          })}
          tabIndex={0}
          className="overflow-x-auto rounded-sm border border-kante font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          <ul>
            {gruppe.map((zeile) => (
              <li
                key={`${zeile.art}-${String(zeile.alt ?? '')}-${String(zeile.neu ?? '')}`}
                className={[
                  'flex gap-2 whitespace-pre px-2',
                  farben[zeile.art],
                ].join(' ')}
              >
                <span className="w-8 shrink-0 text-right text-leise">
                  {zeile.alt ?? ''}
                </span>
                <span className="w-8 shrink-0 text-right text-leise">
                  {zeile.neu ?? ''}
                </span>
                <span aria-hidden="true" className="w-3 shrink-0">
                  {marken[zeile.art]}
                </span>
                {zeile.art !== 'gleich' && (
                  <span className="sr-only">
                    {zeile.art === 'neu' ? texte.zeileNeu : texte.zeileWeg}
                  </span>
                )}
                <span>{zeile.text}</span>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  )
}
