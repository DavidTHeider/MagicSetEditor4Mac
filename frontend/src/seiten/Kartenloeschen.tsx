// Das Löschen einer Karte, samt Rückfrage
// (#621).
//
// **Eine eigene Komponente**, weil `Kartenhandlungen.tsx` sonst über die
// Zeilengrenze ginge (§2) — und weil der Schnitt der Frage folgt: hier steht,
// *wie* gefragt wird, dort, *was* danach mit dem offenen Kontext geschieht.
//
// # Warum überhaupt gefragt wird
//
// **Weil ⌘Z eine Zusage ist und keine Ausrede.** Der Verlauf trägt das Löschen
// seit #621 zurück; trotzdem ist eine gelöschte Karte ein Verlust, den man
// nicht aus Versehen machen will. Die Rückfrage nennt deshalb beides: was
// geschieht, und dass es sich zurücknehmen lässt.
//
// **Kein Modal.** §6: keine verschachtelten Dialoge — die Frage steht an der
// Stelle, an der geklickt wurde. Seit Werkstück U2 von Phase 5.1
// (#1413) ist
// sie eine `Entscheidung`: der Fokus springt hinein, Escape ist Abbrechen,
// und danach steht er wieder auf dem Menü, aus dem sie kam. **Der Auslöser
// ist ein Menüeintrag** — deshalb sagt der Aufrufer, wann die Frage steht.
//
// # Und eine zweite Frage, die von unten kommt
//
// Liegt die Karte in einem Deck, weist der Dienst sie ab und nennt die Decks
// (#1369).
// Diese Antwort trägt den Schlüssel `karteInDeck` — daran, und **nur** daran,
// erkennt diese Stelle, dass ein „Trotzdem" angeboten werden darf. Ein
// gewöhnlicher Fehler bekommt keines.
import { useState } from 'react'

import { api, auspacken, Bindungsfehler } from '../api'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { meldungAus } from '../meldung'
import { type Vorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'

interface KartenloeschenEigenschaften {
  // setID ist das Set, gegen das der Dienst die Karte hält (#808).
  setID: string

  karteID: string

  // vorgang hält Regel 4 aus §7: ein zweiter Klick, während der erste läuft,
  // ginge gegen einen Stand, den der Bestand noch nicht kennt.
  vorgang: Vorgang

  // schliessen nimmt die Frage weg — nach dem Abbrechen wie nach dem Löschen.
  schliessen: () => void

  // aufGeloescht läuft **nach** dem geglückten Löschen. Was dann mit der
  // Auswahl und den Nebenlisten geschieht, weiß nur die Seite (§7, Regel 1).
  //
  // **Mit der Kennung, nicht ohne.** Die Seite braucht sie, um zu prüfen, ob
  // die gelöschte noch die aufgeschlagene ist (§7, Regel 2) — inzwischen kann
  // jemand eine andere aufgeschlagen haben.
  aufGeloescht: (karteID: string) => void
}

export function Kartenloeschen({
  setID,
  karteID,
  vorgang,
  schliessen,
  aufGeloescht,
}: KartenloeschenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten

  const [fehler, setzeFehler] = useState<string>()

  // hindernis ist gesetzt, wenn der Dienst wegen eines Deckplatzes abgewiesen
  // hat — dann und nur dann steht das Trotzdem da.
  const [hindernis, setzeHindernis] = useState(false)

  const loeschen = async (trotzdem: boolean) => {
    await vorgang.starte(`loeschen:${karteID}`, async () => {
      try {
        const antwort = await api().KarteLoeschen(setID, karteID, trotzdem)
        auspacken(antwort, () => undefined)

        setzeFehler(undefined)
        setzeHindernis(false)
        schliessen()
        aufGeloescht(karteID)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))

        // **Am Schlüssel und nicht am Satz.** Ein Vergleich auf den Text wäre
        // in einer englischen Sitzung falsch — und beim nächsten Umformulieren
        // still falsch.
        setzeHindernis(
          ursache instanceof Bindungsfehler &&
            ursache.schluessel === 'karteInDeck',
        )
      }
    })
  }

  const laeuft = vorgang.laeuft(`loeschen:${karteID}`)

  return (
    <Entscheidung
      frage={eigene.loeschenFrage}
      bestaetigen={eigene.loeschenJa}
      aktion="loeschen"
      abbrechen={eigene.loeschenNein}
      laeuft={laeuft}
      aufBestaetigen={() => {
        void loeschen(false)
      }}
      aufAbbrechen={() => {
        setzeFehler(undefined)
        schliessen()
      }}
    >
      {hindernis && (
        <div>
          <Aktionsknopf
            aktion="loeschen"
            gewicht="haupt"
            angehalten={laeuft}
            aufKlick={() => {
              void loeschen(true)
            }}
          >
            {eigene.loeschenTrotzdem}
          </Aktionsknopf>
        </div>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </Entscheidung>
  )
}
