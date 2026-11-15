// **Was die Kachel „Sammlung“ unter ihre Zahl schreibt** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`: „486 € · ältester Preis vor 3 Tagen“.
//
// # Aus dem Portfolio, und ohne Netz
//
// **Gelesen wird, was im Bestand liegt** — dieselbe Rechnung wie die
// Übersicht in der Sammlung (`bausteine/Portfolio.tsx`). Kein Preis wird
// dafür geholt: Preise kommen auf Anweisung (§16, Regel 5), und eine
// Startseite, die beim Öffnen fragte, überraschte.
//
// # Drei Angaben, und keine Umrechnung
//
// **Je Währung ihre Summe**, nebeneinander — umgerechnet wird nirgends, und
// eine Summe über zwei Währungen wäre eine erfundene Zahl. **Das Alter ist
// das des ältesten Preises**: es sagt, wie weit man der Zahl trauen kann.
// **Und die Exemplare ohne Preis stehen dabei, sobald es welche gibt** —
// sonst sähe die Summe aus wie eine über alles; sind es keine, ist sie es.
import type { Portfolioantwort } from '../api'
import { geldAusText } from '../geld'
import { einsetzen } from '../sprache/einsetzen'
import type { Katalog } from '../sprache/kataloge'
import { alterSatz } from './altersatz'

export function sammlungsunterzeile(
  portfolio: Pick<Portfolioantwort, 'zeilen' | 'ohnePreis'> | undefined,
  texte: Katalog,
  sprache: string,
): string | undefined {
  if (portfolio === undefined) {
    return undefined
  }

  const eigene = texte.seiten.start.ueberblick
  const teile = portfolio.zeilen.map((zeile) =>
    geldAusText(zeile.wert, zeile.waehrung, sprache),
  )

  if (portfolio.zeilen.length > 0) {
    const stunden = Math.max(
      ...portfolio.zeilen.map((zeile) => zeile.alterStunden),
    )
    teile.push(
      einsetzen(eigene.aeltesterPreis, {
        wann: alterSatz(stunden * 3600, eigene.alter),
      }),
    )
  }

  if (portfolio.ohnePreis > 0) {
    teile.push(
      einsetzen(eigene.ohnePreis, { anzahl: String(portfolio.ohnePreis) }),
    )
  }

  return teile.length === 0 ? undefined : teile.join(' · ')
}
