// **Die zwei kurzen Wege der Netzpartie** — P9c
// (#2377),
// Artboard `Netzpartie09`: S12 ≤ 3 und S13 ≤ 2 ab dem Bereich.
//
// **Für das erste Gegenüber mit dem Recht am Tisch** — wer mehrere hat, nimmt
// für die anderen die lange Wahl (Sitz „Einladen …“, Schublade „Bei jemandem
// mitspielen …“). **Mit dem eigenen Deck, mit dem zuletzt gearbeitet wurde**
// (`spiel/mitbringen.ts`); ohne eigenes Deck lädt niemand ein — am ersten
// Sitz läge keines —, mitspielen geht dann mit einem Deck des Wirts.
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { mitbringvorgabe } from '../spiel/mitbringen'
import { useAnschluss } from '../anschlussraum'
import { usePartiekontakte } from '../haken/usePartiekontakte'
import { useTexte } from '../sprache/kontext'
import { useVorbelegung } from '../haken/useVorbelegung'

export function Netzwege({ decks }: { decks: readonly { id: string }[] }) {
  const eigene = useTexte().seiten.spielen
  const anschluss = useAnschluss()
  const [erster] = usePartiekontakte()
  const quellen = useVorbelegung()

  if (erster === undefined || quellen === undefined) {
    return null
  }

  const deckID = mitbringvorgabe(decks, quellen.zuletzt)

  return (
    <>
      {deckID !== '' && (
        <Knopf
          aufKlick={() => {
            anschluss.uebergib({
              art: 'partie',
              name: erster.name,
              deckID,
              sofort: true,
              eingeladen: erster,
            })
          }}
        >
          {einsetzen(eigene.netzEinladen, { name: erster.name })}
        </Knopf>
      )}

      <Knopf
        aufKlick={() => {
          anschluss.uebergib({
            art: 'mitspielen',
            name: erster.name,
            gegenueber: erster.kennung,
            deckID,
          })
        }}
      >
        {einsetzen(eigene.netzMitspielen, { name: erster.name })}
      </Knopf>
    </>
  )
}
