// **Die Hinweise des Betriebsbands** — OPS02
// (#2348),
// Artboard `Start09`: in „Was Aufmerksamkeit braucht", je Lage ein Satz, ein
// Rat und ein Ziel.
//
// # Wohin ein Hinweis führt
//
// **Zur Gruppe der Einstellungen, die das Ziel nennt** — über denselben
// Anschluss wie „Diagnose öffnen" an einem Fehler. Ein Ziel, das dieses
// Fenster nicht kennt, bekommt keinen Knopf: ein Knopf, der nirgendwohin
// führt, wäre eine Sackgasse mit Beschriftung.
//
// # Gigabyte wie im Finder
//
// **Dezimal**, wie macOS den Platz seit 10.6 nennt — sonst stünde hier eine
// andere Zahl als im Finder daneben.
import { useAnschluss } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import type { BetriebshinweisDTO } from '../apibetriebsband'

const gruppen = new Map([
  ['einstellungen/daten', 'daten'],
  ['einstellungen/diagnose', 'diagnose'],
])

// abschnitte schlägt auf, wo die Antwort eines Hinweises steht. **Der
// Speicher der Webprozesse** steht seit #2465 eingeklappt unter „Gewirkt seit
// dem Start“ — ohne den Abschnitt führte der Hinweis zu einer Gruppe, in der
// er nicht zu sehen ist.
const abschnitte: Partial<Record<BetriebshinweisDTO['art'], string>> = {
  webprozessNeu: 'wirkung',
}

// alsGigabyte schreibt Bytes als Gigabyte mit einer Nachkommastelle.
function alsGigabyte(bytes: string | undefined, darstellung: string): string {
  return `${(Number(bytes ?? 0) / 1e9).toLocaleString(darstellung, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  })} GB`
}

export function Betriebshinweise({
  hinweise,
}: {
  hinweise: BetriebshinweisDTO[]
}) {
  const eigene = useTexte().seiten.start.aufmerksamkeit.betrieb
  const darstellung = useDarstellung()
  const anschluss = useAnschluss()

  if (hinweise.length === 0) {
    return null
  }

  return (
    <ul className="mb-2 flex flex-col gap-1">
      {hinweise.map((hinweis) => {
        const satz = einsetzen(eigene.arten[hinweis.art], {
          ...hinweis.werte,
          frei: alsGigabyte(hinweis.werte.frei, darstellung),
        })
        const gruppe = gruppen.get(hinweis.ziel)
        const abschnitt = abschnitte[hinweis.art]
        const inhalt = (
          <>
            <span className="font-medium">{satz}</span>
            <span className="block text-xs text-leise">
              {eigene.rat[hinweis.art]}
            </span>
          </>
        )

        return (
          <li key={hinweis.art}>
            {gruppe === undefined ? (
              <p className="px-2 py-1 text-sm">{inhalt}</p>
            ) : (
              <button
                type="button"
                className="w-full rounded-sm px-2 py-1 text-left text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                onClick={() => {
                  anschluss.uebergib({
                    art: 'einstellungsgruppe',
                    name: satz,
                    gruppe,
                    ...(abschnitt === undefined ? {} : { abschnitt }),
                  })
                }}
              >
                {inhalt}
              </button>
            )}
          </li>
        )
      })}
    </ul>
  )
}
