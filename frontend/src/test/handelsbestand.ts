// **Ein Handelsbestand, wie ihn ein Mensch nach einem Jahr hat** — UXR-15
// (#1923).
//
// # Warum es das braucht
//
// Die Ausgangslage in `docs/UIREVIEW.md` ist am Probebestand gemessen —
// **ohne einen einzigen Handel und ohne bestätigtes Gegenüber**. Dort misst
// *Handeln* dreizehn Bedienelemente und sieht aufgeräumt aus. Mit echten
// Mengen wachsen drei Listen mit dem Bestand: die Handel in „Was gerade
// läuft", die bestätigten Gegenüber und das Journal eines Tauschs.
//
// **Die Mengen**: zehn bestätigte Gegenüber, zwölf offene Handel (sechs, an
// denen du dran bist, sechs, die warten), fünfundzwanzig abgeschlossene und
// acht abgebrochene — dazu ein Journal mit achtzig Bewegungen, wie es ein
// Tausch gegen eine ganze Sammelmappe ergibt. **Die Namen sind ausgedacht**;
// keine Karte ist eine echte (§11).
import { einKontakt, eineListe } from '../seiten/peerkontaktprobe'
import { eineZeile } from '../seiten/handelslisteprobe'

import type { Journalzeile } from '../apihandelsverbuchung'
import type { Peerkontakteantwort } from '../apipeerkontakt'
import type {
  Handelsfilter,
  Handelslage,
  Handelsuebersichtantwort,
  Handelsvorgangzeile,
} from '../apihandelsverzeichnis'

// dieNamen sind die zehn Gegenüber — der erste ist der der Kontaktprobe.
export const dieNamen = [
  'Fremder Rechner',
  'Anna',
  'Bex',
  'Cori',
  'Dario',
  'Elif',
  'Finn',
  'Greta',
  'Hiro',
  'Ines',
] as const

// kennungVon gibt die Kennung des Gegenübers an einer Stelle.
export function kennungVon(stelle: number): string {
  return stelle === 0 ? 'bgegenueber' : `bpartner${String(stelle)}`
}

// zehnGegenueber sind die bestätigten Gegenüber dieses Bestands.
export function zehnGegenueber(): Peerkontakteantwort {
  return eineListe(
    dieNamen.map((name, stelle) =>
      einKontakt({
        kennung: kennungVon(stelle),
        selbstgewaehlterName: name,
        fingerabdruck: `${String(1000 + stelle)}-2222-3333-4444-5555-6666-7777-8888`,
        stand: 'bestaetigt',
        rechte: ['handel'],
      }),
    ),
  )
}

// zeilenMit baut `anzahl` Handel einer Lage, reihum mit den zehn Gegenübern.
function zeilenMit(
  lage: Handelslage,
  anzahl: number,
  zustand: string,
  ab: number,
): Handelsvorgangzeile[] {
  return Array.from({ length: anzahl }, (_, stelle) => {
    const nummer = ab + stelle
    const partner = nummer % dieNamen.length

    return eineZeile(
      {
        lage,
        naechster: lage === 'dran' ? 'zustimmen' : '',
        gegenueber: [
          { kennung: kennungVon(partner), name: dieNamen[partner] ?? '' },
        ],
        karten: [`Glutschmiede ${String(nummer)}`, 'Nebelhüterin'],
        geaendert: `2026-09-${String(10 + (nummer % 18)).padStart(2, '0')}T10:00:00Z`,
      },
      {
        vorgang: `h-${String(nummer)}`,
        zustand,
        ende: lage === 'abgeschlossen' || lage === 'abgebrochen',
        parteien: ['bich', kennungVon(partner)],
      },
    )
  })
}

// derHandelsbestand sind alle Handel, nach Lage.
export const derHandelsbestand: Record<Handelslage, Handelsvorgangzeile[]> = {
  dran: zeilenMit('dran', 6, 'angeboten', 1),
  wartet: zeilenMit('wartet', 6, 'verhandlung', 7),
  abgeschlossen: zeilenMit('abgeschlossen', 25, 'beidseitigVerbucht', 13),
  abgebrochen: zeilenMit('abgebrochen', 8, 'abgelehnt', 38),
}

// handelsuebersicht antwortet wie der Dienst: nach den gewünschten Lagen.
export function handelsuebersicht(
  filter: Handelsfilter,
): Promise<Handelsuebersichtantwort> {
  return Promise.resolve({
    zeilen: filter.lagen.flatMap((lage) => derHandelsbestand[lage]),
    fehler: null,
  })
}

// einLangesJournal sind achtzig Bewegungen zu einem Tausch.
export function einLangesJournal(vorgang = 'h-1'): Journalzeile[] {
  return Array.from({ length: 80 }, (_, stelle) => ({
    id: stelle + 1,
    art: stelle % 2 === 0 ? 'abgang' : 'zugang',
    anlass: 'handel',
    vorgang,
    bezug: 0,
    grund: '',
    printing: `druck-${String(stelle)}`,
    bezeichnung: `Mappenkarte ${String(stelle + 1)}`,
    menge: 1,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    exemplarId: stelle + 1,
    lagerort: 'Mappe 3',
    notiz: '',
    gebucht: '2026-09-29T10:00:00Z',
  }))
}
