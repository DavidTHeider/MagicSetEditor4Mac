// **Wie man eine Karte am Tisch bedient** — P4
// (#2372):
// Klick und Doppelklick (`useKartengriff`) und das Kontextmenü
// (`menueZurKarte`), für den gezeichneten Tisch und die Liste dasselbe —
// beides aus der Auskunft des Kerns (`handlungen`).
import type { Objekthandlung, Partie } from '../apipartie'
import type { Menueeintrag } from '../bausteine/Menueliste'
import { handlungZu, type Klick } from '../spiel/kartengriff'
import { menueZurKarte } from '../spiel/kartenmenue'
import { useTexte } from '../sprache/kontext'
import type { Partielauf } from './partielauf'
import { useKartengriff } from './useKartengriff'
import { useNurLesend } from './schreibrecht'

export interface Kartenbedienung {
  greife: (kennung: string, klick: Klick) => void
  menueFuer: (
    kennung: string,
  ) => { bezeichnung: string; eintraege: Menueeintrag[] } | undefined

  // tastenFuer nennt die Einzeltasten einer Karte für `aria-keyshortcuts`;
  // taste stellt, was eine von ihnen meint, und sagt, ob sie galt (C-09).
  tastenFuer: (kennung: string) => string
  taste: (kennung: string, taste: string) => boolean
}

// Taste ist eine Einzeltaste an einer Karte: wie sie heißt, wann sie gilt und
// was sie stellt.
interface Einzeltaste {
  name: string
  tasten: readonly string[]
  gilt: (handlung: Objekthandlung) => boolean
  befehl: (
    handlung: Objekthandlung,
    eigner: string,
  ) => [string, Record<string, unknown>]
}

// **T, ⌫ und E** — der Vorschlag aus C-09. ⌫ geht über die gedeutete Geste,
// denn was im Friedhof landet, entscheidet das Profil (P3).
// keineHandlungen ist, was ein Leser am Tisch tun kann.
const keineHandlungen: readonly Objekthandlung[] = []

const einzeltasten: readonly Einzeltaste[] = [
  {
    name: 'T',
    tasten: ['t', 'T'],
    gilt: (handlung) => handlung.primaer !== '',
    befehl: (handlung) => [handlung.primaer, { objekt: handlung.objekt }],
  },
  {
    name: 'Backspace',
    tasten: ['Backspace', 'Delete'],
    gilt: (handlung) => handlung.handlungen.includes('objekt.ziehen'),
    befehl: (handlung, eigner) => [
      'objekt.ziehen',
      { objekt: handlung.objekt, zone: 'friedhof', sitz: eigner },
    ],
  },
  {
    name: 'E',
    tasten: ['e', 'E'],
    gilt: (handlung) => handlung.handlungen.includes('objekt.exilieren'),
    befehl: (handlung) => ['objekt.exilieren', { objekt: handlung.objekt }],
  },
]

export function useKartenbedienung({
  partie,
  lauf,
  waehle,
  bedienbar = true,
}: {
  partie: Partie
  lauf: Pick<Partielauf, 'handlungen' | 'befehlen'>
  waehle: (kennung: string) => void

  // bedienbar ist falsch beim Sichtwechsel und hinter dem Vorhang.
  bedienbar?: boolean
}): Kartenbedienung {
  const texte = useTexte()

  // **Ein Leser stellt nichts** (#2482): ohne Handlungen wählt ein Klick
  // nur, und keine Einzeltaste gilt — der Kern wiese ohnehin ab, aber erst
  // danach. Das Menü sperrt seine schreibenden Zeilen selbst (`schreibt`).
  const nurLesend = useNurLesend()
  const handlungen = nurLesend ? keineHandlungen : lauf.handlungen
  const greife = useKartengriff({
    handlungen,
    befehlen: lauf.befehlen,
    waehle,
    stand: [partie.id, partie.fuer, String(bedienbar)].join('\u0000'),
  })

  const eignerVon = (kennung: string) =>
    partie.objekte.find((eines) => eines.kennung === kennung)?.eigner ?? ''

  return {
    greife,
    tastenFuer: (kennung) => {
      const handlung = handlungZu(handlungen, kennung)

      return handlung === undefined
        ? ''
        : einzeltasten
            .filter((eine) => eine.gilt(handlung))
            .map((eine) => eine.name)
            .join(' ')
    },
    taste: (kennung, taste) => {
      const handlung = handlungZu(handlungen, kennung)
      const eine = einzeltasten.find((kandidat) =>
        kandidat.tasten.includes(taste),
      )

      if (handlung === undefined || eine?.gilt(handlung) !== true) {
        return false
      }

      lauf.befehlen(...eine.befehl(handlung, eignerVon(kennung)))

      return true
    },
    menueFuer: (kennung) => {
      const objekt = partie.objekte.find((eines) => eines.kennung === kennung)

      return objekt === undefined
        ? undefined
        : menueZurKarte(
            texte.seiten.spielen,
            texte.seiten.handlungen,
            objekt,
            handlungZu(lauf.handlungen, kennung),
            lauf.befehlen,
            waehle,
          )
    },
  }
}
