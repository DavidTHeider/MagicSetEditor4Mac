// **Ein Deck als Umschlag hinaus und wieder herein** — Werkstück T7 von
// Phase 5.6
// (#1472).
//
// # Warum es neben der Übernahme steht
//
// **Weil beide ein Deck in dieses Profil bringen.** Wer eines sucht, sucht
// sie an derselben Stelle — derselbe Satz, der `Deckuebernahme` dorthin
// gestellt hat.
//
// # Und warum das angenommene Deck ein neues ist
//
// **Weil ein Umschlag von außen nichts überschreiben darf.** Wer zweimal
// annimmt, hat zwei Decks — und sieht es am Namen: trägt dieses Profil den
// Namen schon, bekommt das angenommene einen sichtbaren Zusatz.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { DeckDTO } from '../apidecks'

interface Deckumschlageigenschaften {
  // deck ist das offene Deck — ohne eines gibt es nichts zu schnüren.
  deck: DeckDTO | null

  // angenommen läuft, nachdem ein Deck hereingekommen ist.
  angenommen: () => void
}

export function Deckumschlag({ deck, angenommen }: Deckumschlageigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const lauf = useVorgang()

  const [fehler, setzeFehler] = useState<string>()
  const [satz, setzeSatz] = useState<string>()

  const kennzeichen = 'deckumschlag'

  const schnueren = (): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeSatz(undefined)

      if (deck === null) {
        setzeFehler(eigene.umschlagOhneDeck)

        return
      }

      try {
        const antwort = await api().UmschlagSchnueren('deck', deck.id)

        auspacken(antwort, () => undefined)

        // **Ein abgebrochener Dialog ist keine Meldung wert.**
        if (!antwort.abgebrochen) {
          setzeSatz(eigene.umschlagGeschnuert)
        }
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const annehmen = (): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeSatz(undefined)

      try {
        const antwort = await api().UmschlagDeck()

        const gelesen = auspacken(antwort, () => antwort.deck)

        // **Ein abgebrochener Dialog gibt kein Deck** — und keine Meldung.
        if (gelesen === null) {
          return
        }

        setzeSatz(`${eigene.umschlagAngenommen}: ${gelesen.name}`)
        angenommen()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <div className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.umschlagErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void schnueren()
          }}
        >
          {eigene.umschlagSchnueren}
        </Knopf>

        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void annehmen()
          }}
        >
          {eigene.umschlagAnnehmen}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {satz !== undefined && (
        <p role="status" className="text-sm text-leise">
          {satz}
        </p>
      )}
    </div>
  )
}
