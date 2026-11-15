// **Ein Spielbestand, wie ihn ein Mensch hat** — UXR-13
// (#1921).
//
// # Warum es das braucht
//
// Die Ausgangslage in `docs/UIREVIEW.md` ist am Probebestand gemessen —
// **ohne ein Deck, ohne eine Partie, ohne ein Gegenüber**. Dort zeigt
// *Spielen* einen Leerzustand und zwei Formulare darunter. Wie die Seite mit
// echten Mengen aussieht, sagt erst ein Bestand mit echten Mengen: die acht
// Decks aus `deckbestand.ts`, neunzehn Partien — laufende, angehaltene und
// beendete, zu zweit, zu dritt und zu viert —, drei Gegenüber, bei denen
// mitgespielt werden darf, und zwei Profile.
//
// **Die Namen sind ausgedacht** und folgen dem Artboard *Eine Partie
// anfangen*; keine Partie ist eine echte (§11).
import { deckuebersicht } from './deckbestand'
import { einKontakt, eineListe } from '../seiten/peerkontaktprobe'
import { einNutzer } from './anmeldebelegung'
import { probeprofil } from './profil'

import { eineFreigabe } from '../seiten/spielenprobe'

import type { Api } from '../api'
import type { Partiezeile } from '../apipartie'

// sitzeAus baut die Sitze einer Partie aus Namen.
function sitzeAus(namen: readonly string[]): Partiezeile['sitze'] {
  return namen.map((name, stelle) => ({
    kennung: `sitz${String(stelle + 1)}`,
    name,
  }))
}

// **Neunzehn Partien** — so viele, wie der Leerzustand der Liste in
// `Partienliste.tsx` als Beispiel nennt. Neun laufen, vier sind angehalten,
// sechs beendet; die meisten zu zweit, einige zu dritt und zu viert, eine
// allein (Goldfish).
export const partienbestand: Partiezeile[] = (
  [
    [['Ada', 'Bo'], 'laeuft', 7],
    [['Ada', 'Cleo', 'Dario', 'Bo'], 'laeuft', 3],
    [['Ada'], 'laeuft', 0],
    [['Ada', 'Bo'], 'pausiert', 12],
    [['Bo', 'Cleo'], 'laeuft', 2],
    [['Ada', 'Dario'], 'beendet', 9],
    [['Ada', 'Bo', 'Cleo'], 'laeuft', 5],
    [['Cleo', 'Ada'], 'beendet', 14],
    [['Ada', 'Bo'], 'pausiert', 4],
    [['Dario', 'Ada'], 'laeuft', 1],
    [['Ada', 'Bo'], 'beendet', 11],
    [['Ada', 'Cleo', 'Bo', 'Dario'], 'beendet', 8],
    [['Ada', 'Bo'], 'laeuft', 6],
    [['Bo', 'Ada'], 'pausiert', 3],
    [['Ada', 'Cleo'], 'laeuft', 10],
    [['Ada', 'Bo'], 'beendet', 13],
    [['Ada', 'Dario'], 'pausiert', 2],
    [['Cleo', 'Bo'], 'laeuft', 4],
    [['Ada', 'Bo'], 'beendet', 16],
  ] as const
).map(([namen, stand, zug], stelle) => ({
  id: `p${String(stelle + 1)}`,
  fassung: String(10 + stelle),
  spiel: stelle % 4 === 3 ? 'pokemon' : 'magic',
  stand,
  sitze: sitzeAus(namen),
  objekte: 40 + stelle * 3,
  zug,
  eigenerSitz: '',
  dran: false,
}))

// **Drei Gegenüber, bei denen mitgespielt werden darf** — bestätigt und mit
// „Eine Partie spielen" (UX-08.AK-04, #1680).
export const spielgegenueber = ['Bo', 'Cleo', 'Dario'].map((name) =>
  einKontakt({
    kennung: `b${name.toLowerCase()}`,
    eigenerName: name,
    stand: 'bestaetigt',
    rechte: ['partie'],
  }),
)

// **Der Regelbaum, wie der Kern ihn gibt** — dreizehn Bereiche in zwei
// Gruppen (`domain.Regelbaum`). Die Vorrichtung der Spielenseite trägt zwei;
// wer „Wir wählen selbst" misst, braucht alle.
const baum = [
  {
    kennung: 'ablauf',
    bereiche: [
      'zonen',
      'zugstruktur',
      'prioritaet',
      'stapel',
      'combat',
      'zustandsbasiert',
    ],
  },
  {
    kennung: 'mechanik',
    bereiche: [
      'typen',
      'marken',
      'anhaengsel',
      'spielsteine',
      'mana',
      'wechselwirkungen',
      'commander',
    ],
  },
]

// **Je Bereich ein Grad**, verteilt wie im Grundprofil: das meiste hilft
// mit, einiges ist von Hand.
const grade = ['begleitet', 'begleitet', 'manuell', 'automatisch']

// spielbestand belegt die Fassade für die Spielenseite mit echten Mengen.
//
// **Zwei Profile**: dann steht am Sitz die Profilwahl, wie bei jemandem, der
// ein Turnier- und ein Küchentischprofil führt (G8, #1465).
export function spielbestand(): Partial<Api> {
  const nutzer = {
    ...einNutzer(),
    profile: [
      probeprofil({ id: '10', name: 'Küchentisch' }),
      probeprofil({ id: '11', name: 'Turnier' }),
    ],
  }

  return {
    Decks: () => Promise.resolve({ decks: deckuebersicht, fehler: null }),
    Partien: () => Promise.resolve({ partien: partienbestand, fehler: null }),
    Peerkontakte: () => Promise.resolve(eineListe(spielgegenueber)),
    Nutzer: () => Promise.resolve({ nutzer: [nutzer], fehler: null }),
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),
    Regelfreigabe: () => {
      const vorgabe = eineFreigabe()

      return Promise.resolve({
        ...vorgabe,
        freigabe: {
          ...vorgabe.freigabe,
          baum,
          matrix: baum
            .flatMap((gruppe) => gruppe.bereiche)
            .map((bereich, stelle) => ({
              bereich,
              grad: grade[stelle % grade.length] ?? 'manuell',
              quelle: '',
              grenzen: [],
            })),
        },
      })
    },
  }
}
