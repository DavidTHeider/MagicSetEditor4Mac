// **Ein Deck kopieren** — DOM-03.AK-02
// (#1698).
//
// **Ein Haken und kein Knopf** seit AP-6
// (#2363):
// das Kopieren beginnt im Menü „…“ am offenen Deck oder im Rechtsklick, und
// beide brauchen denselben Lauf. Das Formular dazu steht in `Deckkopie.tsx`.
import { useState } from 'react'

import type { DeckDTO } from '../apidecks'
import { api } from '../fassade'
import { feldbefund } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import type { Deckstand } from './useDecks'

export interface Deckkopierlauf {
  // name ist der Name der Kopie — `undefined`, solange niemand kopiert.
  name: string | undefined
  befund: string | undefined
  laeuft: boolean
  beginnen: () => void
  setzeName: (name: string) => void
  abbrechen: () => void
  kopieren: (name: string) => Promise<void>
}

// **Ohne offenes Deck gibt es nichts zu kopieren** (#2469): der Lauf steht
// seitdem über dem Navigator, und dort ist nicht immer ein Deck offen. Nur
// dann zeigt keine Zeile das Formular, das ihn ruft.
export function useDeckkopie(
  deck: DeckDTO | null,
  stand: Deckstand,
): Deckkopierlauf {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  // **Der Vorschlag steht in der Sprache des Fensters** — deshalb hier und
  // nicht im Dienst: dort stünde er in einer Sprache im Bestand.
  const [name, setzeName] = useState<string>()
  const [befund, setzeBefund] = useState<string>()

  const deckID = deck?.id ?? ''
  const deckname = deck?.name ?? ''
  const schluessel = `deck:${deckID}`

  const kopieren = async (gewaehlt: string) => {
    const kopie: { deck: DeckDTO | null } = { deck: null }
    const wahl = stand.wahl()

    await stand.fuehre(
      schluessel,
      async () => {
        const antwort = await api().DeckKopieren(deckID, gewaehlt)
        kopie.deck = auspacken(antwort, () => antwort.deck)
        setzeName(undefined)
      },
      // **Ein vergebener Name steht am Namen** (§6) — und das Popover
      // bleibt offen, damit der nächste Versuch einer ist.
      (ursache) => {
        const amNamen = feldbefund(ursache, 'name', texte)
        setzeBefund(amNamen)

        return amNamen !== undefined
      },
    )

    // **Die Kopie geht auf** — wer kopiert, will an ihr weiterbauen. Danach
    // und nicht darin (§7, Regel 3): geschrieben ist geschrieben, wie beim
    // Anlegen in `Deckanlage`. **Nur, wenn seit dem Kopierbeginn niemand
    // gewechselt hat** (Regel 2): wer inzwischen ein anderes Deck aufschlug
    // — oder das Original neu —, behält seine Wahl. Befund 9 des Zweitblicks
    // über Phase 6 von 0.9.0 und Rest R3 seines Nachlaufs.
    if (kopie.deck !== null) {
      await stand.wechsleVon(wahl, kopie.deck.id)
    }
  }

  return {
    name,
    befund,
    laeuft: stand.vorgang.laeuft(schluessel),
    beginnen: () => {
      setzeBefund(undefined)
      setzeName(einsetzen(eigene.kopieVorschlag, { name: deckname }))
    },
    setzeName: (wert) => {
      setzeName(wert)
      setzeBefund(undefined)
    },
    abbrechen: () => {
      setzeName(undefined)
    },
    kopieren,
  }
}
