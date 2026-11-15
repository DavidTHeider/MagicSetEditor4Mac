// **Ein Verweis auf eine Ansicht nennt eine, die es gibt** — OPS-07.NAT-01
// (#1724):
// „Alle Diagnoseverweise öffnen die tatsächlich vorhandene Ansicht;
// insbesondere Datei-/Identitätsfehler nennen keinen erfundenen
// Hilfe-Unterbereich."
//
// # Was er hält
//
// **Jedes „Einstellungen › X" und „Hilfe › X"** in beiden Katalogen — auch in
// Fehlern und Hinweisen — nennt nach dem Pfeil eine Gruppe der Einstellungen
// oder einen Bereich der Hilfe, wie das Fenster sie beschriftet.
//
// # Was er nicht hält
//
// **Einen Verweis in anderer Gestalt** — „in den Einstellungen", „unter
// Technik". Die Gestalt mit Pfeil ist die, die eine Stelle behauptet; die
// anderen sagen „irgendwo dort", und das ist keine Zusage.
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'

// texte sammelt jede Zeichenkette eines Katalogs samt ihrem Pfad.
function texte(wert: unknown, pfad: string, aus: [string, string][]) {
  if (typeof wert === 'string') {
    aus.push([pfad, wert])
  } else if (typeof wert === 'object' && wert !== null) {
    for (const [schluessel, kind] of Object.entries(wert)) {
      texte(kind, `${pfad}.${schluessel}`, aus)
    }
  }

  return aus
}

const pfeil =
  /(?<!System )\b(Einstellungen|Settings|Hilfe|Help) ?[›→] ?([^.,;:!?)“”»«"'—\n*_]+)/g

describe.each(['de', 'en'] as const)('Die Verweise in %s', (sprache) => {
  const katalog = kataloge[sprache]
  const gruppen = Object.values(katalog.seiten.einstellungen.gruppen)
  const hilfe = [
    ...Object.values(katalog.hilfe).flatMap((teil) =>
      typeof teil === 'object' && 'ueberschrift' in teil
        ? [teil.ueberschrift]
        : [],
    ),
    katalog.seiten.hilfe.glossar,
    katalog.seiten.hilfe.seitenuebersicht,
  ]

  it('nennen nur Ansichten, die es gibt', () => {
    const erfunden: string[] = []

    for (const [pfad, text] of texte(katalog, sprache, [])) {
      for (const treffer of text.matchAll(pfeil)) {
        const ziel = treffer[2]?.trim() ?? ''
        const namen = /^(Einstellungen|Settings)$/.test(treffer[1] ?? '')
          ? gruppen
          : hilfe

        if (!namen.some((name) => ziel.startsWith(name))) {
          erfunden.push(`${pfad}: »${treffer[0]}«`)
        }
      }
    }

    expect(erfunden).toEqual([])
  })

  // **Und die Regel greift wirklich** — ohne diesen Satz wäre ein Muster,
  // das nichts trifft, von einem, das alles trifft, nicht zu unterscheiden.
  it('findet überhaupt Verweise', () => {
    const alle = texte(katalog, sprache, []).filter(
      ([, text]) => [...text.matchAll(pfeil)].length > 0,
    )

    expect(alle.length).toBeGreaterThan(3)
  })
})
