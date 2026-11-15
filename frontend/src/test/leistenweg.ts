// **Der Weg über die Seitenleiste, an einer Stelle** — UX-02
// (#1674).
//
// # Wofür es ihn gibt
//
// **Weil vierzehn Läufe den Namen des Eintrags abgeschrieben hatten.** Sie
// klickten `kataloge.de.seiten.sets.titel` — „Sets" —, und das war richtig,
// solange jede Seite ihren eigenen Eintrag hatte. Seit die Leiste **Bereiche**
// führt, heißt derselbe Eintrag *Gestalten*, und vierzehn Läufe suchten einen
// Knopf, den es nicht mehr gibt.
//
// **Der Name steht jetzt dort, wo die Leiste ihn bildet.** Wer die Ordnung
// wieder ändert, ändert sie einmal.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { beschriftung } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { leisteneintraege } from '../leistenordnung'

import type { Seitenschluessel } from '../navigation'
import type { Sprache } from '../sprache/kataloge'

// leistenname gibt den Namen, unter dem eine Seite in der Leiste steht.
//
// **Und wirft, wenn sie dort nicht steht.** Seit UX-02 führt die Leiste neun
// Einträge statt vierzehn; *Turniere*, *Vorlagen* und jede Lasche des Sets
// erreicht man anders. Ein Lauf, der sie hier sucht, soll das erfahren —
// nicht als „Knopf nicht gefunden" eine Ebene später.
export function leistenname(
  ziel: Seitenschluessel,
  sprache: Sprache = 'de',
): string {
  const eintrag = leisteneintraege.find((einer) => einer.ziel === ziel)

  if (eintrag === undefined) {
    throw new Error(`${ziel} steht nicht in der Seitenleiste.`)
  }

  return eintrag.alsBereich
    ? kataloge[sprache].allgemein.bereiche[eintrag.bereich.schluessel]
    : beschriftung(kataloge[sprache], eintrag.ziel)
}

// geheZu klickt den Eintrag, der auf eine Seite führt.
export async function geheZu(
  ziel: Seitenschluessel,
  sprache: Sprache = 'de',
): Promise<void> {
  await userEvent.click(await leisteneintrag(ziel, sprache))
}

// leisteneintrag gibt den Knopf der Leiste, der auf eine Seite führt.
//
// **Und sucht dafür in der Leiste, nicht auf der Seite** — UX-03: der Weg
// über dem Kopf trägt dieselben Namen. Auf der Setliste steht *Gestalten*
// zweimal: einmal als Eintrag der Leiste, einmal als erste Station des Weges.
// Eine Abfrage über den ganzen Baum fände beide und schlüge fehl.
// **Und wartet dabei auf sie.** Das Fenster zeichnet erst, wenn die
// Einstellungen da sind (#245); ein `getByRole` unmittelbar nach `render`
// fände die Leiste nicht.
export async function leisteneintrag(
  ziel: Seitenschluessel,
  sprache: Sprache = 'de',
): Promise<HTMLElement> {
  const leiste = await screen.findByRole('navigation', {
    name: kataloge[sprache].allgemein.navigation.bezeichnung,
  })

  return within(leiste).getByRole('button', {
    name: leistenname(ziel, sprache),
  })
}
