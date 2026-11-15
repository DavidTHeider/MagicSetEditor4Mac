// **Welche Seite ⌘0 bis ⌘9 öffnen** — UX-02.AK-03
// (#1674):
// *„bestehende Kürzel werden erhalten oder mit dokumentierter Migration und
// konfigurierbarer Belegung ersetzt."*
//
// # Die Vorgabe und die Abweichung
//
// **Die Vorgabe steht in `navigation.ts`** — die Ziffern der ersten Seiten
// der Bereiche. Die Einstellung `kuerzel.ziffern` hält nur, **was davon
// abweicht**: „3=karten" übersteuert eine Ziffer, „2=" räumt eine, und was
// nicht genannt ist, bleibt. So verliert niemand eine Ziffer, die er nie
// angefasst hat, wenn die Vorgabe sich ändert.
//
// # Wer die Gestalt prüft und wer die Seite
//
// **Beides prüft die Domäne** (`domain.PruefeZiffernbelegung`) — seit Befund
// G13 des Zweitblicks über Phase 6.3 auch die Seite, denn MCP, Kommandozeile
// und REST haben kein Fenster, das `3=kartne` hätte abweisen können. Dass
// ihre Liste die von `navigation.ts` ist, hält
// `internal/werkzeuge/fensterseiten_test.go`. **Eine unbekannte fällt hier
// trotzdem weg** — ein Bestand von gestern kann sie noch tragen, und eine
// Ziffer ins Leere wäre schlimmer als die Vorgabe.
import { seiten, type Seitenschluessel } from '../navigation'

// Ziffernbelegung ordnet einer Ziffer eine Seite zu; eine fehlende Ziffer ist
// frei.
export type Ziffernbelegung = Readonly<Record<string, Seitenschluessel>>

// ziffern sind die zehn, in der Reihenfolge der Tastatur.
export const ziffern = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '0']

const bekannt = new Set<string>(seiten.map((seite) => seite.schluessel))

// vorgabebelegung ist, was `navigation.ts` den Ziffern gibt.
export const vorgabebelegung: Ziffernbelegung = Object.fromEntries(
  seiten
    .filter((seite) => ziffern.includes(seite.kuerzel))
    .map((seite) => [seite.kuerzel, seite.schluessel]),
)

// belegungVor06 ist die Belegung bis 0.5.0 — die Spalte „führte auf" der
// Migrationstabelle in `CHANGELOG.md` (#1753). Wer sie gelernt hat, holt sie
// mit einem Knopf zurück.
export const belegungVor06: Ziffernbelegung = {
  '1': 'start',
  '2': 'sets',
  '3': 'karten',
  '4': 'pakete',
  '5': 'vorlagen',
  '6': 'sammlung',
  '7': 'decks',
  '8': 'turniere',
  '9': 'spielen',
  '0': 'handeln',
}

// eintrag erkennt „ziffer=seite" — die Gestalt aus der Domäne.
const eintrag = /^([0-9])=([a-z]*)$/u

// belegungAus liest den Wert der Einstellung über die Vorgabe.
export function belegungAus(wert: string): Ziffernbelegung {
  const belegung = new Map(Object.entries(vorgabebelegung))

  for (const teil of wert === '' ? [] : wert.split(',')) {
    const treffer = eintrag.exec(teil)
    if (treffer === null) continue

    const [, ziffer = '', seite = ''] = treffer

    if (seite === '') {
      belegung.delete(ziffer)
    } else if (bekannt.has(seite)) {
      belegung.set(ziffer, seite as Seitenschluessel)
    }
  }

  return Object.fromEntries(belegung)
}

// mitZiffer gibt eine Belegung, in der eine Ziffer auf eine Seite führt —
// **oder auf keine**, wenn die Seite leer ist.
export function mitZiffer(
  belegung: Ziffernbelegung,
  ziffer: string,
  seite: string,
): Ziffernbelegung {
  const neu = new Map(Object.entries(belegung))

  if (seite === '') {
    neu.delete(ziffer)
  } else {
    neu.set(ziffer, seite as Seitenschluessel)
  }

  return Object.fromEntries(neu)
}

// alsWert schreibt, was von der Vorgabe abweicht — in der Reihenfolge der
// Tastatur, und leer, wenn nichts.
export function alsWert(belegung: Ziffernbelegung): string {
  return ziffern
    .filter((ziffer) => belegung[ziffer] !== vorgabebelegung[ziffer])
    .map((ziffer) => `${ziffer}=${belegung[ziffer] ?? ''}`)
    .join(',')
}
