// **Ein Turnierbestand, wie ihn ein Verein hat** — UXR-12
// (#1920).
//
// # Warum es das braucht
//
// Die Ausgangslage in `docs/UIREVIEW.md` ist am Probebestand gemessen —
// **ohne ein einziges Turnier**. Dort zeigt die Seite drei Bedienelemente und
// fünf Felder. Wie die Werkbank während eines Abends aussieht, sagt erst ein
// Bestand mit echten Mengen: **sechzehn Teilnehmende**, die empfohlenen
// **vier Runden** nach Appendix E der MTR (Fassung 2026-02-27) mit je acht
// Tischen, die vierte läuft noch — fünf von acht Ergebnissen stehen. Dazu
// sieben weitere Turniere in der Liste, wie sie sich über die Monate
// sammeln.
//
// **Die Namen sind ausgedacht**; kein Deck und keine Karte ist eine echte
// (§11).
import {
  paarungsprobe,
  rangprobe,
  rundenprobe,
  teilnehmerprobe,
  turnierprobe,
} from '../seiten/turnierprobe'

import type {
  PaarungDTO,
  ProtokolleintragDTO,
  RangDTO,
  TeilnehmerDTO,
  TurnierDTO,
} from '../apiturniere'

export const namen = [
  'Ada',
  'Bert',
  'Cleo',
  'Dora',
  'Emil',
  'Fenja',
  'Gero',
  'Hanna',
  'Ida',
  'Jonas',
  'Kira',
  'Lars',
  'Mia',
  'Nils',
  'Olga',
  'Paul',
] as const

// sechzehn sind die Teilnehmenden — die Kennung ist `t1` bis `t16`, und
// jede zweite hat ein Deck eingereicht.
export const sechzehn: TeilnehmerDTO[] = namen.map((name, stelle) =>
  teilnehmerprobe({
    id: `t${String(stelle + 1)}`,
    name,
    stelle: stelle + 1,
    deckID: stelle % 2 === 0 ? `d${String(stelle + 1)}` : '',
    deckabzug:
      stelle % 2 === 0
        ? [
            {
              id: `a${String(stelle + 1)}`,
              printing: '',
              druckID: '',
              bezeichnung: `Glutfalke ${String(stelle + 1)}`,
              menge: 4,
              bereich: 'hauptdeck',
            },
          ]
        : [],
  }),
)

// paarungen einer Runde: Tisch für Tisch, der Erste gewinnt 2:1 — bis auf
// die Tische ab `offenAb`, die noch spielen.
function paarungen(runde: number, offenAb = 8): PaarungDTO[] {
  return Array.from({ length: 8 }, (_, tisch) => {
    // **Die erste Hälfte gegen die zweite, je Runde um einen Tisch
    // verschoben** — so spielt niemand zweimal gegen denselben, wie das
    // Schweizer System es verlangt.
    const a = tisch
    const b = 8 + ((tisch + runde - 1) % 8)

    return paarungsprobe({
      id: `p${String(runde)}-${String(tisch + 1)}`,
      a: `t${String(a + 1)}`,
      b: `t${String(b + 1)}`,
      ergebnis:
        tisch < offenAb ? { spieleA: 2, spieleB: 1, unentschieden: 0 } : null,
    })
  })
}

// abend ist das Turnier mitten in der vierten Runde.
export const abend: TurnierDTO = turnierprobe({
  id: '7',
  name: 'Dämmerpokal',
  formatID: '4',
  teilnehmer: sechzehn,
  empfohleneRunden: 4,
  runden: [1, 2, 3, 4].map((nummer) =>
    rundenprobe({
      id: `r${String(nummer)}`,
      nummer,
      paarungen: paarungen(nummer, nummer === 4 ? 5 : 8),
    }),
  ),
})

// rangliste ist die Rangliste dazu — sechzehn Zeilen.
export const rangliste: RangDTO[] = sechzehn.map((teilnehmer, stelle) =>
  rangprobe({
    teilnehmerID: teilnehmer.id,
    name: teilnehmer.name,
    platz: stelle + 1,
    siege: 3 - (stelle % 4),
    niederlagen: stelle % 4,
    matchpunkte: 3 * (3 - (stelle % 4)),
  }),
)

// protokoll ist, was an diesem Abend geschah: Anlegen, sechzehn Anmeldungen,
// vier Runden, 29 Ergebnisse und eine Berichtigung.
export const protokoll: ProtokolleintragDTO[] = [
  {
    zeit: '2026-10-01T18:00:00Z',
    vorgang: 'angelegt',
    angaben: 'Dämmerpokal',
    paarungID: '',
  },
  ...namen.map((name, stelle) => ({
    zeit: `2026-10-01T18:${String(stelle + 1).padStart(2, '0')}:00Z`,
    vorgang: 'angemeldet',
    angaben: name,
    paarungID: '',
  })),
  ...abend.runden.flatMap((runde) => [
    {
      zeit: `2026-10-01T${String(18 + runde.nummer)}:30:00Z`,
      vorgang: 'runde',
      angaben: String(runde.nummer),
      paarungID: '',
    },
    ...runde.paarungen
      .filter((paarung) => paarung.ergebnis !== null)
      .map((paarung) => ({
        zeit: `2026-10-01T${String(19 + runde.nummer)}:15:00Z`,
        vorgang: 'ergebnis',
        angaben: '2:1:0',
        paarungID: paarung.id,
      })),
  ]),
  {
    zeit: '2026-10-01T22:20:00Z',
    vorgang: 'ergebnis',
    angaben: '1:2:0',
    paarungID: 'p2-3',
  },
]

// weitere sind die Turniere früherer Abende — sieben, ohne Runden.
export const weitere: TurnierDTO[] = [
  'Aprilturnier',
  'Februarturnier',
  'Herbstcup',
  'Januarturnier',
  'Märzturnier',
  'Pauper-Abend',
  'Sommercup',
].map((name, stelle) =>
  turnierprobe({
    id: String(20 + stelle),
    name,
    notiz: stelle % 2 === 0 ? 'Vereinsheim' : '',
  }),
)

// liste ist die Turnierliste mit dem laufenden Abend.
export const liste: TurnierDTO[] = [abend, ...weitere]
