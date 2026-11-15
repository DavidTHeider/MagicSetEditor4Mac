// **Was man selten braucht, in zwei Schubladen** — UXR-13
// (#1921):
// *An einem fremden Tisch mitspielen* und *Ein Deck verleihen oder leihen*.
//
// # Was vorher war
//
// **Beide standen offen unter dem Anfang und der Partienliste**, dazwischen
// der Kasten für ein Gegenüber. Mit echten Mengen (`test/spielbestand.ts`)
// zeichnete die Seite 19 Felder — R3 in `docs/UIREVIEW.md` lässt zwölf zu —,
// und keiner ihrer Knöpfe führte (R1). Das Artboard *Eine Partie anfangen*
// zeichnet keinen der beiden Abschnitte; sie gehören auf diese Seite, aber
// hinter einen Knopf im Kopf, wie *Einlesen und ausgeben* bei der Sammlung.
//
// # Der Kasten für ein Gegenüber steht in beiden
//
// **Weil die Seite unter einer Schublade gesperrt ist** (`inert`): wer beim
// Beitritt oder bei der Leihe niemanden zur Wahl hat, nimmt ihn dort auf, wo
// er gerade ist (UX-03.AK-03). Er steht unter dem Formular — die leeren Wahlen
// sagen „darunter".
//
// # Gezeichnet beim ersten Öffnen, danach nur verborgen
//
// **Wie die Schubladen der Sammlung** (`Sammlungswerkzeuge.tsx`): eine halb
// getroffene Wahl, ein getipptes Ziel überstehen das Schließen. Was eine
// Schublade fragt — die Kontakte, die Decks —, fragt sie erst, wenn jemand sie
// aufzieht.
import { Activity, useState } from 'react'

import { Deckleihe } from './Deckleihe'
import { Gastbeitritt } from './Partiegast'
import { Gegenueberkasten } from './Gegenueberkasten'
import { Schublade } from '../bausteine/Schublade'
import { useTexte } from '../sprache/kontext'

import type { Gastlauf } from '../haken/usePartiegast'

// Spielschublade ist eine der zwei.
export type Spielschublade = 'gast' | 'leihe'

interface SchubladenEigenschaften {
  offen: Spielschublade | undefined
  schliessen: () => void

  // gast ist der Lauf des Beitritts — **einer für die Seite**: der Platz am
  // fremden Tisch steht auf ihr, der Beitritt in der Schublade.
  gast: Gastlauf

  // kontaktstand zählt, wie oft ein Gegenüber aufgenommen oder bestätigt
  // wurde; die Wahlen laden dann nach.
  kontaktstand: number

  // leiheGeaendert meldet ein geglücktes Schreiben an einer Leihe — die
  // Deckwahl am Sitz lädt dann neu (P1b, #2370).
  leiheGeaendert?: (() => void) | undefined
  kontaktGeaendert: () => void
}

// sichtbar macht aus einem Wahrheitswert die Art einer `<Activity>`.
const sichtbar = (ja: boolean) => (ja ? 'visible' : 'hidden')

export function Spielschubladen({
  offen,
  schliessen,
  gast,
  kontaktstand,
  leiheGeaendert,
  kontaktGeaendert,
}: SchubladenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  // **Beim Zeichnen gesetzt und nicht in einem Effekt**, wie im
  // `Aufklappkasten`: sonst stünde die Schublade erst einen Zug später da.
  const [gezeichnet, setzeGezeichnet] = useState<readonly Spielschublade[]>([])
  if (offen !== undefined && !gezeichnet.includes(offen)) {
    setzeGezeichnet([...gezeichnet, offen])
  }

  const titel: Record<Spielschublade, string> = {
    gast: eigene.gast,
    leihe: eigene.leihe,
  }

  return gezeichnet.map((schublade) => (
    <Activity key={schublade} mode={sichtbar(schublade === offen)}>
      <Schublade
        titel={titel[schublade]}
        schliessenBezeichnung={
          texte.allgemein.arbeitsplatz.schubladen.schliessen
        }
        schliessen={schliessen}
      >
        <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
          {schublade === 'gast' ? (
            <Gastbeitritt lauf={gast} kontaktstand={kontaktstand} />
          ) : (
            <Deckleihe kontaktstand={kontaktstand} geaendert={leiheGeaendert} />
          )}

          <Gegenueberkasten aufGeaendert={kontaktGeaendert} />
        </div>
      </Schublade>
    </Activity>
  ))
}
