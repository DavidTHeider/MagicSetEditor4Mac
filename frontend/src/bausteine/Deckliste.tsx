// **Die Deckübersicht** — Werkstück 7 von Phase 4.6
// (#1293).
import { useEffect, useState } from 'react'

import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { api } from '../fassade'
import { Auswahlfilter } from './Auswahlfilter'
import { auspacken } from '../umschlag'
import { Deckzeile } from './Deckzeile'
import { Wirdgeladen } from './Ladehinweise'
import { useSprache, useTexte } from '../sprache/kontext'
import { Listenansicht } from './Listenansicht'
import {
  deckansichtsschluessel,
  deckformen,
  decksortierungen,
} from './listenordnung'
import { useAnsicht } from '../haken/useAnsicht'
import { useAnstoss } from '../haken/useAnstoss'
import { useVorgabe } from '../haken/useVorgabe'

import type { Deckhandlung } from '../haken/useDeckhandlung'
import type { Deckstand } from '../haken/useDecks'

export function Deckliste({
  stand,
  handlung,
}: {
  stand: Deckstand
  handlung: Deckhandlung
}) {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  // **Umbenennen und die Rückfrage beim Löschen** stehen seit #2469 eine
  // Ebene höher (`useDeckhandlung`), samt dem Räumen beim Wechsel des Decks.
  const offenID = stand.offen?.id

  // **Rechtsklick und ⇧F10 auf einem anderen Deck öffnen es zuerst** —
  // AP-6 (#2363); das Menü gehört dann ihm (`useAnstoss`).
  const { anstoss, verbraucht, anstossen } = useAnstoss<string>(
    offenID,
    (deckID) => {
      void stand.waehle(deckID)
    },
  )

  // **Die Präferenzdecks stehen zuerst** — Werkstück 15 (#1351), Punkt 6.
  //
  // **In ihrer eigenen Reihenfolge**, denn genau die ist die Präferenz: was
  // das Profil zuerst nennt, steht zuerst. Alles andere folgt in der
  // Reihenfolge, in der es kam.
  //
  // **Sortiert und nicht gefiltert.** Wer ein Deck bevorzugt, will die
  // anderen nicht verlieren — eine Liste, die sie verschwinden ließe, wäre
  // eine Falle für den, der einmal etwas anderes sucht.
  const bevorzugt = useVorgabe('deck')

  // **Die Decks mit fehlenden Karten aus dem Überblick** — AP-8, Teil 2b
  // (#2365): die Liste zeigt genau sie; der Filter nennt sich und lässt sich
  // lösen, wie in Set- und Paketliste.
  const anschluss = useAnschluss()
  const mitgereist =
    anschluss.offen?.art === 'deckauswahl' ? anschluss.offen : undefined
  const [auswahl, setzeAuswahl] = useState<{
    name: string
    deckIDs: ReadonlySet<string>
  }>()
  useAufgenommen(mitgereist, ({ name, deckIDs }) => {
    setzeAuswahl({ name, deckIDs: new Set(deckIDs) })
  })

  // **Verbraucht, sobald aufgenommen** — sonst nähme die Liste, die beim
  // Wechsel zwischen den Laschen „Decks“ und „Formate“ neu entsteht, die
  // Auswahl ein zweites Mal auf, auch nachdem jemand sie gelöst hat.
  useEffect(() => {
    if (mitgereist !== undefined) {
      anschluss.verbrauche()
    }
  }, [mitgereist, anschluss])

  const decks =
    auswahl === undefined
      ? stand.decks
      : stand.decks.filter((deck) => auswahl.deckIDs.has(deck.id))

  // **Liste oder Kacheln, mit Suche und Ordnung** — UX-11 (#1683). Die
  // erste Ordnung ist die bisherige: zuletzt benutzt zuerst.
  const sprache = useSprache()
  const ansicht = useAnsicht(deckansichtsschluessel, deckformen)

  // **Leer heißt: angekommen und leer** (UX-08.NAT-01). Solange die Antwort
  // aussteht, steht das da; nach einem Fehlschlag steht er über der Seite,
  // samt dem nächsten Versuch — die Liste behauptet dann nichts.
  if (!stand.geladen) {
    return stand.ladefehler === undefined ? <Wirdgeladen /> : null
  }

  const umbenennen = async (
    deckID: string,
    fassung: string,
    name: string,
    notiz: string,
  ) => {
    await stand.fuehre(`deck:${deckID}`, async () => {
      const antwort = await api().DeckUmbenennen(deckID, fassung, name, notiz)

      auspacken(antwort, () => antwort.deck)
      handlung.setzeUmbenannt(undefined)
    })
  }

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-medium">{eigene.titel}</h2>

      {auswahl !== undefined && (
        <Auswahlfilter
          name={auswahl.name}
          loesen={eigene.auswahlLoesen}
          aufLoesen={() => {
            setzeAuswahl(undefined)
          }}
        />
      )}

      <Listenansicht
        bezeichnung={eigene.titel}
        eintraege={decks}
        schluessel={(deck) => deck.id}
        suchtext={(deck) => `${deck.name} ${deck.spiel} ${deck.notiz}`}
        sortierungen={decksortierungen(texte, sprache, (deckID) =>
          rang(bevorzugt, deckID),
        )}
        form={ansicht.form}
        setzeForm={ansicht.setzeForm}
        leer={<p>{eigene.leer}</p>}
        marken={{ art: 'deck', name: (deck) => deck.name }}
        // **Das offene Deck bleibt stehen** (#2469) — auch wenn Suche oder
        // Marke es nicht treffen: an ihm hängen seine Handlungen.
        festgehalten={(deck) => deck.id === offenID}
        eintrag={(deck, gestalt) => (
          <Deckzeile
            deck={deck}
            offen={deck.id === offenID}
            gestalt={gestalt}
            stand={stand}
            handlung={handlung}
            umbenennen={umbenennen}
            anstoss={anstoss}
            anstossVerbraucht={verbraucht}
            anstossen={(ereignis) => {
              anstossen(ereignis, deck.id, deck.id)
            }}
          />
        )}
      />
    </section>
  )
}

// rang gibt die Stelle eines Decks unter den bevorzugten — oder eine dahinter.
//
// **`Array.prototype.sort` ist seit ES2019 stabil**, und das ist hier die
// Zusage: alles, was denselben Rang hat, behält seine Reihenfolge.
function rang(bevorzugt: readonly string[], deckID: string): number {
  const stelle = bevorzugt.indexOf(deckID)

  return stelle === -1 ? bevorzugt.length : stelle
}
