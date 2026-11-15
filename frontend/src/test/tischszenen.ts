// **Die zehn Szenen des Spieltischs** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Warum es sie gibt
//
// **#1460 verlangt sie namentlich**: „Deterministische Szenen empty-board,
// two-player-simple, creature-combat, large-land-piles, stack-heavy,
// tokens-and-counters, four-player-commander, extreme-board-state,
// double-faced-cards und many-attachments existieren als selbst erstellte
// Fixtures."
//
// **Selbst erstellt, und das ist keine Formsache** (§11): kein Kartenname
// stammt aus einem fremden Paket, keine Zahl aus einem echten Set. Was hier
// steht, ist erfunden — und muss es sein.
//
// # Wozu sie taugen und wozu nicht
//
// **Sie sind die Regressionsprobe des Tischplans.** Jede Szene stellt eine
// Frage, die eine echte Partie stellt: passt ein volles Schlachtfeld noch in
// seine Reihen? Bleibt ein Stapel ein Stapel? Findet ein Anhängsel sein Ziel?
//
// **Sie messen nichts über die Grafikkarte.** Was hier herauskommt, ist ein
// Plan; ob er in sechzehn Millisekunden gezeichnet wird, sagt er nicht.
import { einObjekt, einePartie } from '../seiten/partieprobe'

import type { Partie, Sitz, Spielobjekt } from '../apipartie'

// Szenenname ist eine der zehn.
export type Szenenname =
  | 'empty-board'
  | 'two-player-simple'
  | 'creature-combat'
  | 'large-land-piles'
  | 'stack-heavy'
  | 'tokens-and-counters'
  | 'four-player-commander'
  | 'extreme-board-state'
  | 'double-faced-cards'
  | 'many-attachments'

// szenennamen ist die Liste, gegen die ein Lauf sich hält.
//
// **Aus dem Issue abgeschrieben und in seiner Reihenfolge** — eine kürzere
// Liste wäre grün, und niemand sähe es.
export const szenennamen: Szenenname[] = [
  'empty-board',
  'two-player-simple',
  'creature-combat',
  'large-land-piles',
  'stack-heavy',
  'tokens-and-counters',
  'four-player-commander',
  'extreme-board-state',
  'double-faced-cards',
  'many-attachments',
]

const sitz = (kennung: string, name: string, karten = 60): Sitz => ({
  kennung,
  name,
  deckabdruck: `abdruck-${kennung}`,
  profilID: '',
  karten,
  wartet: false,
  eingeladen: '',

  // **Zwanzig, weil CR 119.1 es sagt** (G9, #1592) — dieselbe Zahl wie im
  // Grundprofil, und die Szenen sollen aussehen wie eine echte Partie.
  leben: 20,
  vergiftung: 0,
  manavorrat: '',
  strahlung: 0,
  commander: '',
  commandersteuer: 0,
  commanderschaden: {},
  leergezogen: false,
  eroeffnung: '',
  mulligans: 0,
  unterlegen: 0,
})

// reihe baut gleichartige Objekte durch — mit fortlaufender Kennung und
// **fortlaufender Ordnung**, damit ein Stapel eine Reihenfolge hat.
const reihe = (
  zahl: number,
  vorsatz: string,
  teile: (stelle: number) => Partial<Spielobjekt>,
): Spielobjekt[] =>
  Array.from({ length: zahl }, (_, stelle) =>
    einObjekt({
      kennung: `${vorsatz}-${String(stelle)}`,
      ordnung: stelle,
      geheim: false,
      ...teile(stelle),
    }),
  )

// bibliothek gibt eine verdeckte Bibliothek — sie gehört in jede Szene, in
// der jemand sitzt (CR 401.1, Fassung 2026-08-19).
const bibliothek = (besitzer: string, zahl: number): Spielobjekt[] =>
  reihe(zahl, `bib-${besitzer}`, () => ({
    zone: 'bibliothek',
    eigner: besitzer,
    beherrscher: besitzer,
    verdeckt: true,
    geheim: true,
    name: '',
    karte: '',
  }))

const szenen: Record<Szenenname, () => Partie> = {
  // **Ein leeres Schlachtfeld** — der Anfang einer Partie und der Fall, an
  // dem eine Rechnung über Listen als Erstes bricht.
  'empty-board': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: [],
    }),

  // **Zwei Spieler, je eine Hand und ein paar Permanente** — der Normalfall.
  'two-player-simple': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: [
        ...bibliothek('a', 53),
        ...bibliothek('b', 53),
        ...reihe(7, 'hand-a', (stelle) => ({
          zone: 'hand',
          eigner: 'a',
          beherrscher: 'a',
          name: `Handkarte ${String(stelle)}`,
        })),
        ...reihe(7, 'hand-b', () => ({
          zone: 'hand',
          eigner: 'b',
          beherrscher: 'b',
          geheim: true,
          name: '',
        })),
        ...reihe(4, 'land-a', (stelle) => ({
          zone: 'tisch',
          eigner: 'a',
          beherrscher: 'a',
          name: `Quelle ${String(stelle)}`,
        })),
      ],
    }),

  // **Ein Kampf** — getappte Angreifer, ungetappte Verteidiger, und beide
  // Seiten tragen Schadensmarken.
  'creature-combat': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: [
        ...reihe(5, 'angriff', (stelle) => ({
          zone: 'tisch',
          eigner: 'a',
          beherrscher: 'a',
          getappt: true,
          name: `Angreifer ${String(stelle)}`,
          zaehler: { schaden: stelle },
        })),
        ...reihe(4, 'abwehr', (stelle) => ({
          zone: 'tisch',
          eigner: 'b',
          beherrscher: 'b',
          name: `Verteidiger ${String(stelle)}`,
          zaehler: { schaden: 1 },
        })),
      ],
    }),

  // **Viele Länder** — die Probe darauf, dass ein volles Schlachtfeld enger
  // gesetzt wird statt über den Rand zu laufen.
  'large-land-piles': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: [
        ...reihe(40, 'land-a', (stelle) => ({
          zone: 'tisch',
          eigner: 'a',
          beherrscher: 'a',
          name: `Quelle ${String(stelle)}`,
          getappt: stelle % 3 === 0,
        })),
        ...reihe(38, 'land-b', (stelle) => ({
          zone: 'tisch',
          eigner: 'b',
          beherrscher: 'b',
          name: `Feld ${String(stelle)}`,
        })),
      ],
    }),

  // **Ein hoher Stapel** — CR 405.2: Neues kommt oben drauf, und die
  // Reihenfolge ist zu sehen.
  'stack-heavy': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: reihe(12, 'stapel', (stelle) => ({
        zone: 'stapel',
        eigner: stelle % 2 === 0 ? 'a' : 'b',
        beherrscher: stelle % 2 === 0 ? 'a' : 'b',
        name: `Spruch ${String(stelle)}`,
      })),
    }),

  // **Spielsteine und Marken** — CR 111 und CR 122.
  'tokens-and-counters': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: [
        ...reihe(20, 'stein', (stelle) => ({
          zone: 'tisch',
          art: 'spielstein',
          eigner: 'a',
          beherrscher: 'a',
          karteID: '',
          name: `Stein ${String(stelle)}`,
          zaehler: { plus: 1 + (stelle % 4) },
        })),
        ...reihe(10, 'marke', (stelle) => ({
          zone: 'tisch',
          eigner: 'b',
          beherrscher: 'b',
          name: `Träger ${String(stelle)}`,
          zaehler: { plus: stelle, aufladung: 10 - stelle },
        })),
      ],
    }),

  // **Vier Sitze mit Kommandozone** — CR 903.
  'four-player-commander': () =>
    einePartie({
      fuer: 'a',
      sitze: [
        sitz('a', 'Ada', 99),
        sitz('b', 'Bo', 99),
        sitz('c', 'Cem', 99),
        sitz('d', 'Dil', 99),
      ],
      objekte: ['a', 'b', 'c', 'd'].flatMap((wer) => [
        ...bibliothek(wer, 20),
        einObjekt({
          kennung: `kommandant-${wer}`,
          zone: 'kommando',
          eigner: wer,
          beherrscher: wer,
          geheim: false,
          name: `Anführer ${wer.toUpperCase()}`,
        }),
        ...reihe(6, `feld-${wer}`, (stelle) => ({
          zone: 'tisch',
          eigner: wer,
          beherrscher: wer,
          name: `Gefolge ${wer}${String(stelle)}`,
        })),
      ]),
    }),

  // **Der Ausnahmezustand** — und die Zahlen sind die aus der
  // Leistungsvorgabe von #1460, nicht ungefähr, sondern genau: „4 Spielern,
  // 150 Karten, 100 Countern, 50 Tokens, 30 Beziehungsgrafiken".
  //
  // **Sie werden gezählt und nicht geschätzt** — `szenen.test.ts` hält jede
  // der vier Zahlen nach. Eine Leistungsszene, die um ein Drittel danebenläge,
  // machte jede Messung darauf unvergleichbar (§10).
  'extreme-board-state': () => {
    const sitze = ['a', 'b', 'c', 'd']

    // 150 Karten auf dem Schlachtfeld: 38 + 38 + 37 + 37.
    const karten = sitze.flatMap((wer, welcher) =>
      reihe(welcher < 2 ? 38 : 37, `feld-${wer}`, (stelle) => ({
        zone: 'tisch',
        eigner: wer,
        beherrscher: wer,
        name: `Ding ${wer}${String(stelle)}`,
        getappt: (stelle + welcher) % 4 === 0,
      })),
    )

    // 50 Spielsteine: 13 + 13 + 12 + 12.
    const steine = sitze.flatMap((wer, welcher) =>
      reihe(welcher < 2 ? 13 : 12, `stein-${wer}`, (stelle) => ({
        zone: 'tisch',
        art: 'spielstein',
        eigner: wer,
        beherrscher: wer,
        karteID: '',
        name: `Stein ${wer}${String(stelle)}`,
      })),
    )

    // **100 Marken, an den ersten hundert Karten.** Eine Marke ist ein Eintrag
    // in `zaehler` (CR 122, Fassung 2026-08-19); gezählt wird, wie viele
    // Objekte welche tragen.
    const mitMarken = karten.map((eines, stelle) =>
      stelle < 100
        ? {
            ...eines,
            zaehler: { plus: 1 + (stelle % 5), aufladung: stelle % 3 },
          }
        : eines,
    )

    // **30 Beziehungen**: dreißig Anhängsel, verteilt auf die ersten Karten
    // jedes Sitzes.
    const anhaengsel = reihe(30, 'hang', (stelle) => ({
      zone: 'tisch',
      eigner: sitze[stelle % 4] ?? 'a',
      beherrscher: sitze[stelle % 4] ?? 'a',
      name: `Anhang ${String(stelle)}`,
      haengtAn: mitMarken[stelle]?.kennung ?? '',
    }))

    return einePartie({
      fuer: 'a',
      sitze: sitze.map((wer, welcher) =>
        sitz(wer, `Spieler ${String(welcher + 1)}`),
      ),
      objekte: [
        ...sitze.flatMap((wer) => bibliothek(wer, 25)),
        ...mitMarken,
        ...steine,
        ...anhaengsel,
      ],
    })
  },

  // **Zweiseitige Karten** — die zweite öffentliche Seite aus G2
  // (#1459).
  'double-faced-cards': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: reihe(8, 'zwei', (stelle) => ({
        zone: 'tisch',
        eigner: 'a',
        beherrscher: 'a',
        karteID: `dfc-${String(stelle)}`,
        name: `Wandler ${String(stelle)}`,
      })),
    }),

  // **Viele Anhängsel** — CR 301 und CR 303: was woran hängt, muss für alle
  // klar bleiben (CR 400.5).
  'many-attachments': () =>
    einePartie({
      fuer: 'a',
      sitze: [sitz('a', 'Ada'), sitz('b', 'Bo')],
      objekte: [
        ...reihe(6, 'traeger', (stelle) => ({
          zone: 'tisch',
          eigner: 'a',
          beherrscher: 'a',
          name: `Träger ${String(stelle)}`,
        })),
        ...reihe(18, 'hang', (stelle) => ({
          zone: 'tisch',
          eigner: stelle % 2 === 0 ? 'a' : 'b',
          beherrscher: stelle % 2 === 0 ? 'a' : 'b',
          name: `Anhang ${String(stelle)}`,
          haengtAn: `traeger-${String(stelle % 6)}`,
        })),
      ],
    }),
}

// szene baut eine der zehn — **jedes Mal neu und jedes Mal gleich**.
export function szene(name: Szenenname): Partie {
  return szenen[name]()
}
