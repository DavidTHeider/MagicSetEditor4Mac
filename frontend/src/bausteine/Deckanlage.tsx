// **Ein Deck anlegen** — Werkstück 7 von Phase 4.6
// (#1293).
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from './Aktionsknopf'
import { useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import { useVorgabe } from '../haken/useVorgabe'
import { useSetImFenster } from '../haken/offenesset'

import type { DeckDTO } from '../apidecks'
import type { Deckstand } from '../haken/useDecks'
import { Pflichtfeld } from './Pflichtfeld'
import { feldbefund } from '../meldung'

export function Deckanlage({ stand }: { stand: Deckstand }) {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  const [name, setzeName] = useState('')
  const [spiel, setzeSpiel] = useState('')
  const [notiz, setzeNotiz] = useState('')
  const [namensbefund, setzeNamensbefund] = useState<string>()

  // **Die Spiele des Profils** — aus seinen Sets, ohne zweiten Weg.
  //
  // **`magic` steht immer dabei**, auch ohne Set: ein Profil ohne Sets soll
  // ein Deck anlegen können, und die Vorgabe der Anwendung ist Magic.
  const sets = useLaden(
    useCallback(async () => {
      const antwort = await api().Sets()

      return auspacken(antwort, () => antwort.sets)
    }, []),
    [],
  )

  const ausSets = [
    ...new Set((sets.daten ?? []).map((satz) => satz.spiel)),
  ].filter((eines) => eines !== '')

  // **`magic` steht immer in der Liste** — ein Profil ohne Sets soll ein Deck
  // anlegen können, und die Vorgabe der Anwendung ist Magic.
  const spiele = ausSets.includes('magic') ? ausSets : [...ausSets, 'magic']

  // **Vorbelegt ist das Spiel des offenen Sets**, sonst das des ersten —
  // wer Decks baut, baut sie meist zu dem, woran er gerade arbeitet (AP-6,
  // #2363: die Wahl war eine der vier Handlungen über dem Ziel des Kernwegs
  // `deck`). Ohne Sets bleibt `magic`.
  const offenesSet = useSetImFenster()
  const gewaehlt =
    spiel === '' ? (offenesSet?.spiel ?? ausSets[0] ?? 'magic') : spiel

  // **Das Vorgabeformat des Profils** — Werkstück 15 (#1351), Punkt 4.
  const vorgabeformat = useVorgabe('format')[0] ?? ''

  const anlegen = async () => {
    const angelegt: { deck: DeckDTO | null } = { deck: null }

    setzeNamensbefund(undefined)

    await stand.fuehre(
      'deck-anlegen',
      async () => {
        const antwort = await api().DeckAnlegen(name, gewaehlt, notiz)

        const deck = auspacken(antwort, () => antwort.deck)

        setzeName('')
        setzeNotiz('')

        // **Der Merker steht vor dem zweiten Ruf** — Befund 4 des Nachlaufs zum
        // Zweitblick über Phase 5.1. Er stand darunter, und damit hing das
        // Aufschlagen am **Format**: scheiterte `DeckFormatWaehlen`, blieb er
        // leer, das Deck ging nicht auf, und weil `fuehre` nach einem Fehler
        // nicht nachlädt (§7, Regel 3), stand es auch nicht in der Liste. Das
        // Deck war da und **nirgends zu sehen** — genau der halbe Stand, den
        // dieselbe Regel verhindern soll.
        angelegt.deck = deck

        // **Das Format kommt in einem zweiten Ruf** — `DeckAnlegen` nimmt
        // keines entgegen, und ihm eines beizubringen hieße, eine Fläche zu
        // ändern, die seit Phase 4.6 steht (§16).
        //
        // **Und geschrieben ist geschrieben** (§7, Regel 3): das Deck ist da,
        // auch wenn das Format nicht gilt. Ein Fehlschlag hier nimmt es nicht
        // zurück — er kommt als Meldung, und das Deck steht ohne Format da, wie
        // jedes, dem niemand eines gab.
        if (vorgabeformat !== '' && deck !== null) {
          auspacken(
            await api().DeckFormatWaehlen(deck.id, vorgabeformat),
            () => undefined,
          )
        }
      },
      // **Ein doppelter Name steht am Namen** — ENG-05.AK-03 (#1730).
      (ursache) => {
        const befund = feldbefund(ursache, 'name', texte)
        setzeNamensbefund(befund)

        return befund !== undefined
      },
    )

    // **Das neue Deck geht auf** — Befund 3 des Zweitblicks über Phase 5.1.
    // Wer eines anlegt, will darin bauen; der Weg „anlegen, suchen, öffnen"
    // war eine Handlung mehr, und der Kernweg `deck` rechnet ohne sie
    // (`docs/UX-ARCHITEKTUR.md`, Ziel ≤ 6).
    //
    // **Danach und nicht darin** (§7, Regel 3): geschrieben ist geschrieben.
    // Innerhalb von `fuehre` machte ein Fehlschlag beim Aufschlagen die
    // Anlage rückwirkend zu einem Fehler — samt ausbleibendem Nachladen der
    // Liste, in der das neue Deck längst steht.
    if (angelegt.deck !== null) {
      await stand.waehle(angelegt.deck.id)
    }
  }

  return (
    <section aria-label={eigene.neu} className="flex flex-col gap-3">
      <h2 className="text-lg font-medium">{eigene.neu}</h2>

      <div className="flex flex-wrap items-end gap-3">
        {/* **Der Name ist Pflicht, und das Feld sagt es** — Werkstück F von
            Phase 4.10. Ohne ihn ruhte der Knopf, und wer ihn ausgegraut sah,
            wusste nicht, welches der drei Felder fehlt (§6). */}
        <Pflichtfeld
          beschriftung={eigene.name}
          wert={name}
          befund={namensbefund}
          geaendert={(wert) => {
            setzeName(wert)
            setzeNamensbefund(undefined)
          }}
        />

        {/* **Das Spiel wird gewählt, nicht getippt** — Befund 6 des
            Zweitblicks über Phase 5.1: ein Textfeld kostete **drei**
            Handlungen für eine Angabe, die im Bestand schon steht. Der
            Kernweg `deck` hat genau das gemessen (`docs/UX-ARCHITEKTUR.md`).

            **Die Liste kommt aus den Sets des Profils**: ein Deck verweist
            auf Karten, und Karten liegen in Sets. Was dort kein Spiel hat,
            gibt es hier nicht — und die Vorgabe des ersten Sets ist die, die
            am ehesten stimmt. */}
        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.spiel}</span>

          <select
            value={gewaehlt}
            onChange={(ereignis) => {
              setzeSpiel(ereignis.target.value)
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {spiele.map((eines) => (
              <option key={eines} value={eines}>
                {eines}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.notiz}</span>

          <input
            value={notiz}
            onChange={(ereignis) => {
              setzeNotiz(ereignis.target.value)
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </label>

        <Aktionsknopf
          aktion="anlegen"
          gewicht="haupt"
          schreibt
          angehalten={stand.vorgang.laeuft('deck-anlegen')}
          aufKlick={() => {
            void anlegen()
          }}
        >
          {eigene.anlegen}
        </Aktionsknopf>
      </div>
    </section>
  )
}
