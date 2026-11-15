// **Ein Deck aus einem fremden Profil holen** — Werkstück 12 von Phase 4.12
// (#1351),
// nach Vermerk [0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md).
//
// # Warum hier seit UX-08.AK-04 eine Auswahl steht
//
// **Hier standen zwei Kennungen**, und der Grund war: die Decks des anderen
// Profils zu laden hieße, über die Profilgrenze zu lesen, bevor jemand gesagt
// hat, was er will. **Der Grund trägt nicht mehr**, seit die Liste hinter
// derselben Erlaubnis steht wie das Holen (`Deckuebernahmewahl`,
// #1680):
// wer keine Freigabe hat, sieht das Nein und keine Decks.
//
// # Warum die Befunde stehen bleiben
//
// **Weil sie das Ergebnis sind und keine Meldung.** „Zwölf Karten kamen
// nicht mit" verschwindet nicht nach drei Sekunden; wer das Deck ansieht,
// muss es daneben lesen können (§8).
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Aufklappkasten } from './Aufklappkasten'
import { Deckuebernahmewahl } from './Deckuebernahmewahl'
import { Knopf } from './Knopf'
import { Teilerfolg } from './Teilerfolg'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { useVorgang } from '../haken/useVorgang'

import type { DeckDTO } from '../apidecks'
import type { Deckstand } from '../haken/useDecks'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Deckuebernahme({ stand }: { stand: Deckstand }) {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  const vorgang = useVorgang()

  const [profil, setzeProfil] = useState('')
  const [deck, setzeDeck] = useState('')
  // geholt ist, was die letzte Übernahme hinterließ — das Deck **und** was
  // nicht mitkam (S08.NAT-03).
  const [geholt, setzeGeholt] = useState<{
    deck: DeckDTO | null
    befunde: readonly string[]
  }>()
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  const hole = () =>
    vorgang.starte('uebernahme', async () => {
      setzeFehler(undefined)
      setzeGeholt(undefined)

      try {
        const antwort = await api().DeckUebernehmen(profil, deck)
        const ergebnis = auspacken(antwort, () => antwort)

        setzeGeholt({ deck: ergebnis.deck, befunde: ergebnis.befunde })
        setzeProfil('')
        setzeDeck('')

        // **Nachladen ist Bestenfalls-Arbeit** (§7, Regel 3): das Deck ist
        // da, auch wenn die Liste sich nicht erneuern lässt.
        stand.neuLaden()
      } catch (grund) {
        setzeFehler(meldungAus(grund, texte))
      }
    })

  return (
    <section
      aria-label={eigene.uebernahme}
      className="flex flex-col gap-3 rounded-md border border-kante p-4"
    >
      <h2 className="text-lg font-medium">{eigene.uebernahme}</h2>

      <Erklaerungsabsatz
        text={eigene.uebernahmeErklaerung}
        klasse="max-w-prose text-leise"
      />

      {/* **Gewählt und nicht getippt** (UX-08.AK-04, #1680): das Profil aus
          der Nutzerliste, das Deck aus seiner Liste — hinter derselben
          Erlaubnis wie das Holen. **Die Kennungen bleiben nennbar**, für
          wen sie ohnehin von der Kommandozeile kennt. */}
      <Deckuebernahmewahl
        profil={profil}
        deck={deck}
        setzeProfil={setzeProfil}
        setzeDeck={setzeDeck}
      />

      <Aufklappkasten titel={eigene.uebernahmeFachleute}>
        <div className="flex flex-col gap-2">
          <Erklaerungsabsatz
            text={eigene.uebernahmeFachleuteErklaerung}
            klasse="max-w-prose text-sm text-leise"
          />

          <div className="flex flex-wrap items-end gap-3">
            <label className="flex flex-col gap-1">
              <span className="text-sm font-medium">
                {eigene.uebernahmeProfil}
              </span>
              <input
                value={profil}
                onChange={(ereignis) => {
                  setzeProfil(ereignis.target.value)
                }}
                className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              />
            </label>

            <label className="flex flex-col gap-1">
              <span className="text-sm font-medium">
                {eigene.uebernahmeDeck}
              </span>
              <input
                value={deck}
                onChange={(ereignis) => {
                  setzeDeck(ereignis.target.value)
                }}
                className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              />
            </label>
          </div>
        </div>
      </Aufklappkasten>

      <div className="flex flex-wrap items-end gap-3">
        <Knopf
          gewicht="haupt"
          angehalten={
            vorgang.laeuft('uebernahme') || profil === '' || deck === ''
          }
          aufKlick={() => {
            void hole()
          }}
        >
          {eigene.uebernahmeHolen}
        </Knopf>
      </div>

      {/* **Was geschrieben ist, und was nicht — getrennt** (S08.NAT-03):
          bis hierher stand nur, was nicht mitkam, und ob das Deck da ist,
          sagte erst die Liste darüber. */}
      {geholt?.deck != null && (
        <Teilerfolg
          geschrieben={einsetzen(eigene.uebernahmeGeschrieben, {
            name: geholt.deck.name,
            karten: zaehlwort(
              geholt.deck.hauptdeck + geholt.deck.sideboard,
              texte.allgemein.zaehlwoerter.karten,
            ),
          })}
          offen={geholt.befunde}
          offenTitel={eigene.uebernahmeNichtMit}
        />
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}
