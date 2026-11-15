// **Ein Deckbestand, wie ihn ein Mensch hat** — UXR-11
// (#1919).
//
// # Warum es das braucht
//
// Die Ausgangslage in `docs/UIREVIEW.md` ist am Probebestand gemessen —
// **ohne ein einziges Deck**. Dort misst die Seite zwölf Bedienelemente und
// sieht aufgeräumt aus. Wie sie mit echten Mengen aussieht, sagt erst ein
// Bestand mit echten Mengen: acht Decks, ein Standarddeck mit 60 + 15 Karten
// und ein Commanderdeck mit 100 — so viele Plätze, wie solche Decks
// tatsächlich haben, und eine Sammlung, in der sieben Karten fehlen.
//
// **Die Namen sind ausgedacht** und folgen dem Artboard *Decks*; keine Karte
// ist eine echte (§11).
import { keineHerkunft } from '../apiherkunft'
import { deckprobe, formatprobe, platzprobe } from '../seiten/deckprobe'

import type { ExemplarDTO } from '../api'
import type {
  DeckDTO,
  DeckplatzDTO,
  DeckurteilDTO,
  FormatDTO,
  ZuordnungszeileDTO,
} from '../apidecks'

// druck gibt die Scryfall-Kennung eines ausgedachten Drucks.
function druck(stelle: number): string {
  return `00000000-0000-4000-8000-${String(stelle).padStart(12, '0')}`
}

// plaetze baut die Plätze eines Decks aus Mengen — je Menge ein Platz.
function plaetze(
  vorsilbe: string,
  mengen: readonly number[],
  bereich: string,
  ab: number,
): DeckplatzDTO[] {
  return mengen.map((menge, stelle) => {
    const nummer = ab + stelle
    const land = stelle >= mengen.length - 4

    return platzprobe({
      id: `${vorsilbe}-${String(nummer)}`,
      printing: druck(nummer),
      bezeichnung: `${vorsilbe} ${String(nummer)}`,
      menge,
      bereich,
      typzeile: land ? 'Land' : 'Kreatur — Elementar',
      manakosten: land ? '' : `{${String(1 + (stelle % 5))}}{R}`,
    })
  })
}

// **Ein Standarddeck: 60 + 15.** Elf Zauber, fünf Länder, sieben im
// Sideboard — 23 Plätze, wie bei einem gespielten Deck.
const standardhaupt = plaetze(
  'Aschen',
  [4, 4, 4, 4, 4, 4, 3, 3, 3, 2, 1, 4, 4, 4, 4, 8],
  'hauptdeck',
  1,
)
const standardneben = plaetze('Glut', [3, 3, 3, 2, 2, 1, 1], 'sideboard', 100)

export const standarddeck: DeckDTO = deckprobe({
  id: 'aschenrat',
  name: 'Aschenrat',
  formatID: 'standard',
  plaetze: [...standardhaupt, ...standardneben],
  hauptdeck: 60,
  sideboard: 15,
})

// **Ein Commanderdeck: hundert Karten, fast jede einmal.** Ein Commander,
// 63 Zauber, zehn Länder einzeln und zwei Grundländer zu je 13 — 76 Plätze.
export const commanderdeck: DeckDTO = deckprobe({
  id: 'bernsteinflut',
  name: 'Bernsteinflut',
  formatID: 'commander',
  plaetze: plaetze(
    'Bernstein',
    [...Array<number>(74).fill(1), 13, 13],
    'hauptdeck',
    200,
  ),
  hauptdeck: 100,
  sideboard: 0,
})

// **Acht Decks in der Übersicht** — die Übersicht trägt keine Plätze.
export const deckuebersicht: DeckDTO[] = [
  { ...standarddeck, plaetze: [] },
  { ...commanderdeck, plaetze: [] },
  ...[
    ['irrlichter', 'Irrlichter (Entwurf)', 23, 0],
    ['glutsturm', 'Glutsturm', 60, 15],
    ['moorgeister', 'Moorgeister', 60, 15],
    ['elfenpfad', 'Elfenpfad', 60, 0],
    ['freitag', 'Draft vom Freitag', 40, 0],
    ['kuechentisch', 'Küchentisch', 75, 0],
  ].map(([id, name, haupt, neben]) =>
    deckprobe({
      id: String(id),
      name: String(name),
      hauptdeck: Number(haupt),
      sideboard: Number(neben),
    }),
  ),
]

// **Zehn Formate**, wie sie aus den Bauarten entstehen.
export const deckformate: FormatDTO[] = [
  'standard',
  'pioneer',
  'modern',
  'legacy',
  'vintage',
  'pauper',
  'commander',
  'brawl',
  'historic',
  'draft',
].map((id) => formatprobe({ id, name: id, scryfallname: id }))

// urteil gibt eine Prüfung mit **einem** Verstoß, wie im Artboard: fünf von
// einer Karte, wo vier erlaubt sind.
export function deckurteil(deck: DeckDTO): DeckurteilDTO {
  return {
    formatname: deck.formatID,
    hauptdeck: deck.hauptdeck,
    sideboard: deck.sideboard,
    urteile: deck.plaetze.slice(0, 1).map((platz) => ({
      platzID: platz.id,
      bezeichnung: platz.bezeichnung,
      menge: 5,
      bereich: platz.bereich,
      zulaessigkeit: 'zuviele',
      grund: 'Standard erlaubt vier.',
      quelle: 'format',
    })),
    groessen: [],
    zulaessig: false,
    ungeprueft: 0,
  }
}

// zuordnungszeilen gibt je Platz eine Zeile — **die letzten sieben Karten
// fehlen**, alles andere liegt in der Sammlung.
export function zuordnungszeilen(deck: DeckDTO): ZuordnungszeileDTO[] {
  return deck.plaetze.map((platz, stelle) => {
    const fehlt = stelle >= deck.plaetze.length - 7 ? 1 : 0

    return {
      printing: platz.printing,
      druckID: '',
      vorlaeufig: false,
      bezeichnung: platz.bezeichnung,
      gebraucht: platz.menge,
      vorhanden: platz.menge - fehlt,
      fehlend: fehlt,
      grund: '',
      reserviert: 0,
      zugeordnet: 0,
      frei: platz.menge - fehlt,
      zuordnungen: [],
    }
  })
}

// sammlung gibt zu jedem Platz die Exemplare, die es davon gibt.
export function deckexemplare(deck: DeckDTO): ExemplarDTO[] {
  return deck.plaetze.map((platz, stelle) => ({
    id: 1000 + stelle,
    fassung: 1,
    printing: platz.printing,
    druckID: '',
    vorlaeufig: false,
    menge: platz.menge,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: platz.bezeichnung,
    setcode: 'xyz',
    nummer: String(stelle + 1),
    lagerort: '',
    wunsch: false,
    bezeichnung: platz.bezeichnung,
    herkunft: keineHerkunft,
  }))
}
