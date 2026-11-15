// **Die Kennzahlen der Sammlung** — AP-9
// (#2366),
// Befund B-15, Entwurf `Bereiche09`: Exemplare, Wert, ohne Preis.
//
// # Ehrliche Preise
//
// **Je Währung ihre Summe, das Alter des ältesten Preises darunter** — und
// kein Trend: dafür bräuchte es eine gespeicherte Reihe über die Sammlung,
// und die gibt es nicht. **Ohne Preis steht „kein Preis“**, keine Null in
// einer Währung, die niemand gewählt hat. Dieselbe Rechnung wie die
// Übersicht im Navigator (`bausteine/Portfolio.tsx`), dieselbe Antwort
// (`usePortfolio`).
//
// # Die Ziele
//
// **Exemplare** nimmt Suche, Ausführung und Marken zurück — danach steht die
// Liste mit allen Einträgen da, die die Zahl darunter nennt. **Wert und ohne
// Preis** führen zur Übersicht im Navigator, über der die Preise einer
// ganzen Quelle geholt werden.
import { useState } from 'react'

import type { Portfolioantwort } from '../api'
import { Kennzahlleiste } from '../bausteine/Kennzahlleiste'
import { geldAusText } from '../geld'
import type { Arbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'
import type { Ladezustand } from '../haken/useLaden'
import { usePortfolio } from '../haken/usePortfolio'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { alterSatz } from './altersatz'

// Sammlungsleiste ist, was Leiste und Navigator teilen: die Übersicht und
// den Anlass, zu ihr zu führen.
export interface Sammlungsleiste {
  portfolio: Ladezustand<Portfolioantwort>
  anlass: number
  zeigePortfolio: () => void
}

// useSammlungsleiste fragt die Übersicht einmal für beide. **Das Ziel liegt
// im Navigator** — ist er eingeklappt, klappt der Klick ihn auf und merkt
// es wie „Navigator einblenden“: der Klick ist der Wunsch, dorthin zu sehen.
export function useSammlungsleiste(
  runde: number,
  gedaechtnis: Arbeitsplatzgedaechtnis,
): Sammlungsleiste {
  const portfolio = usePortfolio(runde)
  const [anlass, setzeAnlass] = useState(0)

  return {
    portfolio,
    anlass,
    zeigePortfolio: () => {
      if (!gedaechtnis.navigator) {
        gedaechtnis.merke('navigator', true)
      }
      setzeAnlass(anlass + 1)
    },
  }
}

export function Sammlungskennzahlen({
  stueck,
  insgesamt,
  leiste: { portfolio, zeigePortfolio },
  zeigeAlle,
}: {
  // stueck ist die Summe der Mengen, insgesamt die Zahl der Einträge — beide
  // ohne Sicht.
  stueck: number
  insgesamt: number
  leiste: Pick<Sammlungsleiste, 'portfolio' | 'zeigePortfolio'>
  zeigeAlle: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung.leiste
  const start = texte.seiten.start.ueberblick
  const darstellung = useDarstellung()

  // **Gescheitert ist kein Wert**: die Meldung steht an der Übersicht, zu
  // der der Knopf führt — die Leiste behauptet nichts.
  const gescheitert = portfolio.fehler !== undefined
  const daten = gescheitert ? undefined : portfolio.daten
  const zeilen = daten?.zeilen ?? []
  const stunden = Math.max(0, ...zeilen.map((zeile) => zeile.alterStunden))

  const wert =
    gescheitert || (daten !== undefined && zeilen.length === 0)
      ? eigene.keinPreis
      : daten === undefined
        ? undefined
        : zeilen
            .map((zeile) =>
              geldAusText(zeile.wert, zeile.waehrung, darstellung),
            )
            .join(' · ')

  return (
    <Kennzahlleiste
      bezeichnung={eigene.bezeichnung}
      kennzahlen={[
        {
          schluessel: 'exemplare',
          was: eigene.exemplare,
          wert: String(stueck),
          unter: einsetzen(eigene.eintraege, { zahl: String(insgesamt) }),
          hin: zeigeAlle,
        },
        {
          schluessel: 'wert',
          was: eigene.wert,
          wert,
          unter:
            zeilen.length === 0
              ? undefined
              : einsetzen(start.aeltesterPreis, {
                  wann: alterSatz(stunden * 3600, start.alter),
                }),
          hin: zeigePortfolio,
        },
        {
          schluessel: 'ohnePreis',
          was: eigene.ohnePreis,
          wert: gescheitert ? eigene.nichtGelesen : daten?.ohnePreis.toString(),
          hin: zeigePortfolio,
        },
      ]}
    />
  )
}
