// **Was in der Seitenleiste steht, und in welcher Ordnung** — UX-02
// (#1674),
// gezeichnet im Artboard `Main` des Entwurfs.
//
// **Ein Eintrag je Bereich und nicht je Seite.** Eine Seite je Eintrag stünde
// ohne Ordnung nebeneinander, und wer „ein Deck bauen" will, müsste wissen,
// dass das unter *Decks* steht und nicht unter *Sammlung*.
//
// Der Entwurf zeichnet neun: sechs Bereiche, ein Trenner, drei Dienste —
// nachgelesen in `prototyp/bau.py` (Entwicklungsarchiv),
// Zeilen 107–120. Wohin es innerhalb eines Bereichs weitergeht, sagt der Kopf
// der Seite; das ist UX-03.
//
// **Die Dienstzone bleibt seitenweise.** Ihre drei Einträge sind keine
// Bereiche: sie tragen die macOS-Kürzel ⌘, ⌘' ⌘?, die §6 ohnehin verlangt,
// und jeder führt auf genau eine Seite.

import { bereiche, dienstseiten, ersteSeiteVon } from './bereiche'
import { istSetseite, seiten } from './navigation'

import type { Bereich, Bereichsschluessel, Dienstseite } from './bereiche'
import type { Seite, Seitenschluessel } from './navigation'
import { ziffern, type Ziffernbelegung } from './rahmen/ziffernbelegung'

// reservierteKuerzel sind Tasten, die das System für sich beansprucht.
export const reservierteKuerzel: Readonly<Record<string, string>> = {}

// anwendungsseiten sind alle Seiten, die nicht zu einem Set gehören.
export const anwendungsseiten = seiten.filter(
  (seite) => !istSetseite(seite.schluessel),
)

// Leisteneintrag ist eine Zeile der Seitenleiste, fertig zum Zeichnen.
export interface Leisteneintrag {
  // bereich ist der Bereich, den dieser Eintrag hervorhebt.
  bereich: Bereich

  // ziel ist die Seite, auf die er führt.
  //
  // **Die erste des Bereichs und nicht eine gemerkte** — der Grund steht bei
  // `ersteSeiteVon`.
  ziel: Seitenschluessel

  // alsBereich sagt, ob der Eintrag den Namen des Bereichs trägt oder den
  // seiner Seite.
  //
  // **Als Feld und nicht als Vergleich mit `bereich.dienst`**: was der
  // Eintrag anschreibt, ist eine Frage an ihn und nicht an seinen Bereich.
  alsBereich: boolean

  // kuerzel ist die Taste, die zusammen mit ⌘ hierher führt.
  kuerzel: string

  // sinnbild steht links vom Namen — im Entwurf eine Spalte für sich.
  //
  // **Ein Zeichen und keine Grafik.** Es steht `aria-hidden` neben dem Namen
  // und sagt nichts, was der Name nicht sagt; wer es lädt, lädt eine
  // Abhängigkeit für eine Verzierung.
  sinnbild: string
}

// bereichsbilder sind die Zeichen aus dem Entwurf.
//
// **Nachgelesen und nicht gewählt**: `docs/prototyp/bau.py`, Zeilen 107–120.
//
// **Zwei vollständige Zuordnungen und nicht eine offene.** Der Schlüsseltyp
// zählt die Fälle auf; damit gibt es keinen Rückfall, den kein Aufruf betritt
// — und ein neuer Bereich ohne Sinnbild wird nicht still leer, sondern rot
// beim Übersetzen.
const bereichsbilder: Readonly<Record<Bereichsschluessel, string>> = {
  start: '⌂',
  gestalten: '✎',
  sammeln: '▤',
  spielen: '☗',
  handeln: '⇄',
  ressourcen: '☰',

  // **Die Dienstzone trägt keines**: ihre drei Einträge tragen eigene.
  programm: '',
}

const dienstbilder: Readonly<Record<Dienstseite, string>> = {
  einstellungen: '⚙',
  nutzer: '☺',
  hilfe: '?',
}

// leisteneintraege ist die Leiste, wie sie dasteht.
//
// **Aus `bereiche` und nicht aus einer zweiten Aufzählung.** Zwei Listen über
// dieselbe Ordnung laufen auseinander; die Frage „welcher Bereich?" hat eine
// Antwort, und sie steht in `bereiche.ts`.
export const leisteneintraege: readonly Leisteneintrag[] = bereiche.flatMap(
  (bereich): Leisteneintrag[] =>
    bereich.dienst
      ? dienstseiten.map((schluessel) => ({
          bereich,
          ziel: schluessel,
          alsBereich: false,
          kuerzel: kuerzelVon(schluessel),
          sinnbild: dienstbilder[schluessel],
        }))
      : [
          {
            bereich,
            ziel: ersteSeiteVon(bereich),
            alsBereich: true,
            kuerzel: bereich.kuerzel,
            sinnbild: bereichsbilder[bereich.schluessel],
          },
        ],
)

// fehlendeLeistenseiten sind die Bereichsseiten, die es nicht gibt.
//
// **Sie steht hier, damit sie sich zählen lässt.** Ein Bereich, dessen Seite
// es nicht gibt, führte ins Leere — und ein stiller Verlust ist genau das,
// was UX-01.AK-04 verhindern soll. Der Lauf daneben verlangt, dass diese
// Liste leer ist.
export const fehlendeLeistenseiten: readonly string[] = bereiche
  .flatMap((bereich) => bereich.seiten)
  .filter(
    (schluessel) => !seiten.some((seite) => seite.schluessel === schluessel),
  )

// kuerzelVon gibt die Taste, die auf eine Seite führt.
function kuerzelVon(schluessel: Seitenschluessel): string {
  const gefunden: Seite | undefined = seiten.find(
    (seite) => seite.schluessel === schluessel,
  )

  /* v8 ignore next -- jede Bereichsseite ist registriert, und
     `bereichsordnung.test.ts` hält das; TypeScript sieht es nicht. */
  return gefunden?.kuerzel ?? ''
}

// mitBelegung gibt jedem Bereich die Ziffer, die die Belegung seiner ersten
// Seite gibt — UX-02.AK-03
// (#1674).
//
// **Die Leiste zeigt, was die Taste tut.** Wer ⌘3 auf die Karten gelegt hat,
// soll bei *Sammeln* keine ⌘3 mehr lesen. **Die Dienstzone bleibt**: ⌘, ⌘'
// und ⌘? sind keine Ziffern und nicht einstellbar.
export function mitBelegung(
  eintraege: readonly Leisteneintrag[],
  belegung: Ziffernbelegung,
): Leisteneintrag[] {
  return eintraege.map((eintrag) =>
    eintrag.alsBereich
      ? {
          ...eintrag,
          kuerzel:
            ziffern.find((ziffer) => belegung[ziffer] === eintrag.ziel) ?? '',
        }
      : eintrag,
  )
}
