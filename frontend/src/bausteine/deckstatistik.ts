// **Typen, Länder, Durchschnitt und Quellen** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Eine eigene Datei neben `deckrechnung.ts`**, damit beide unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort steht, was
// **Manakosten** hergeben — Manawert, Farben, Kurve —, hier, was die
// **Typzeile** hergibt.
//
// # Der Manawert kommt von Scryfall, wo es einen gibt
//
// Eine zweiseitige Karte hat dort **leere** Kosten und trotzdem einen Wert
// (`cmc`). Wer ihn aus den Kosten rechnete, bekäme null und eine Kurve, die
// lügt. Für eine eigene Karte gibt es keinen — dort rechnet [manawert] aus
// den Kosten, und der Rückfall steht in [wertVon].
import { wertBekannt, wertVon } from './deckrechnung'

import type { Farbe, Kartenposten } from './deckrechnung'

export interface Statistikposten extends Kartenposten {
  typzeile: string
}

// Quellenposten ist ein Posten samt den Farben, die er geben kann.
export interface Quellenposten extends Statistikposten {
  farben: readonly Farbe[]
}

// istLand sagt, ob eine Typzeile ein Land nennt.
//
// **Über den Obertyp und nicht über den Namen** — dieselbe Regel wie in
// `domain.IstLand`, und aus demselben Grund: „Forest" ist ein Land, „Forest
// Bear" eine Kreatur.
export function istLand(typzeile: string): boolean {
  return obertypen(typzeile).some((wort) => wort.toLowerCase() === 'land')
}

// auszeichnungen sind Wörter vor dem Kartentyp, die keiner sind.
//
// **Nachgelesen** in den Comprehensive Rules 205.4a (Supertypen): `Basic`,
// `Legendary`, `Ongoing`, `Snow`, `World`. `Artifact` steht **nicht** dabei —
// es ist ein Kartentyp, und „Artifact Creature" trägt zwei. Gezählt wird der
// **letzte**, weil er die Karte benennt: eine Artefaktkreatur ist eine
// Kreatur, die zufällig aus Metall ist.
const auszeichnungen = new Set([
  'basic',
  'legendary',
  'ongoing',
  'snow',
  'world',
])

// vorderseite gibt die Typzeile der **Vorderseite**.
//
// **Nachgelesen in den Comprehensive Rules 712.8a**: außerhalb von
// Schlachtfeld und Stapel hat eine doppelseitige Karte die Merkmale ihrer
// Vorderseite. Scryfall verbindet die Seiten in `type_line` mit ` // ` —
// `Sorcery // Land` ist eine Zauberei, deren Rückseite ein Land ist, und
// **kein Land**. Wer sie als eines zählte, versprächen einem Deck Mana, das
// es nicht ziehen kann.
//
// **Was das nicht tut**: eine **geteilte** Karte trägt nach CR 709.4 die
// Merkmale **beider** Hälften; hier zählt auch dort die vordere. Für „wie
// viele Länder" macht das keinen Unterschied — geteilte Karten sind nie
// Länder —, für „wie viele Instants" schon. Wer es genauer will, braucht die
// Seiten einzeln (`card_faces`), und das ist ein eigenes Werkstück.
function vorderseite(typzeile: string): string {
  const [vorn = ''] = typzeile.split('//')

  return vorn
}

// obertypen gibt die Wörter vor dem Gedankenstrich.
function obertypen(typzeile: string): string[] {
  const [ober = ''] = vorderseite(typzeile).split('—')

  return ober.trim().split(/\s+/u).filter(Boolean)
}

// kartentyp gibt den Typ, unter dem eine Karte gezählt wird — '' ohne Zeile.
function kartentyp(typzeile: string): string {
  const worte = obertypen(typzeile).filter(
    (wort) => !auszeichnungen.has(wort.toLowerCase()),
  )

  return worte.at(-1) ?? ''
}

// typverteilung zählt die Stücke je Kartentyp — in der Reihenfolge des Decks.
//
// **In der Reihenfolge und nicht nach Zahl**: dieselbe Liste soll zweimal
// gleich aussehen, und eine Sortierung nach Häufigkeit springt bei jedem
// Hinzulegen.
export function typverteilung(
  posten: readonly Statistikposten[],
): [string, number][] {
  const nach = new Map<string, number>()

  for (const eintrag of posten) {
    const typ = kartentyp(eintrag.typzeile)

    if (typ === '') {
      continue
    }

    nach.set(typ, (nach.get(typ) ?? 0) + eintrag.menge)
  }

  return [...nach.entries()]
}

// laenderzahl zählt die Länder — Stücke, nicht Zeilen.
export function laenderzahl(posten: readonly Statistikposten[]): number {
  return posten
    .filter((eintrag) => istLand(eintrag.typzeile))
    .reduce((summe, eintrag) => summe + eintrag.menge, 0)
}

// durchschnitt ist der mittlere Manawert der **Nichtländer**.
//
// **Ohne Länder**, weil sie keine kosten und den Schnitt sonst nach unten
// zögen: ein Deck mit vierundzwanzig Wäldern sähe billiger aus als es ist.
//
// **`undefined` und nicht null, wenn es keine gibt**: „im Schnitt null" wäre
// eine Aussage über ein Deck, das keine Nichtländer hat — und die gibt es
// nicht.
export function durchschnitt(
  posten: readonly Statistikposten[],
): number | undefined {
  // **Und ohne die, deren Wert niemand kennt** — [wertBekannt] sagt, welche
  // das sind. Eine Karte ohne Angaben als Nullkostenkarte zu mitteln macht
  // das Deck billiger, als es ist.
  const ohneLand = posten.filter(
    (eintrag) => !istLand(eintrag.typzeile) && wertBekannt(eintrag),
  )

  const stuecke = ohneLand.reduce((summe, eintrag) => summe + eintrag.menge, 0)

  if (stuecke === 0) {
    return undefined
  }

  const summe = ohneLand.reduce(
    (bisher, eintrag) => bisher + wertVon(eintrag) * eintrag.menge,
    0,
  )

  return Math.round((summe / stuecke) * 100) / 100
}

// quellen zählt je Farbe die **Länder**, die sie geben.
//
// **Nur Länder**, und das ist die Frage: „woraus ziehe ich mein Mana". Eine
// grüne Kreatur ist keine Quelle, auch wenn sie grün ist.
//
// **Ein Zweifarbenland zählt in beiden Farben** — wie bei der
// Farbverteilung: die Summe ist größer als die Zahl der Länder, und genau
// das will die Frage wissen.
export function quellen(posten: readonly Quellenposten[]): [Farbe, number][] {
  const reihenfolge: Farbe[] = ['weiss', 'blau', 'schwarz', 'rot', 'gruen']
  const nach = new Map<Farbe, number>()

  for (const eintrag of posten) {
    if (!istLand(eintrag.typzeile)) {
      continue
    }

    for (const farbe of eintrag.farben) {
      nach.set(farbe, (nach.get(farbe) ?? 0) + eintrag.menge)
    }
  }

  return reihenfolge.flatMap((farbe): [Farbe, number][] => {
    const zahl = nach.get(farbe)

    return zahl === undefined ? [] : [[farbe, zahl]]
  })
}
