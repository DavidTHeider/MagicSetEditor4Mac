// **Was man an einem offenen Deck tut** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Ein Haken und kein Bauteil**, weil dieselben vier Handlungen jetzt an
// **zwei** Stellen gebraucht werden: die Kartenwahl steht seit dem
// Arbeitsplatz im Navigator, die Plätze im Zentrum. Sie zweimal zu schreiben
// hieße, sie beim nächsten Mal einmal zu ändern.
//
// **Ein Vorgang je Deck** (§7, Regel 4): alle vier teilen sich den Schlüssel
// `deck:<id>` — wer legt, während ein Format wechselt, ginge gegen einen
// Stand, den der Bestand noch nicht kennt.
import { Bindungsfehler } from '../bindungsfehler'
import { api } from '../fassade'
import { auspacken } from '../umschlag'

import type { Deckstand } from './useDecks'
import type { Stueck } from '../anschluss'
import type { DeckDTO, DeckplatzDTO } from '../apidecks'

// Abweisung ist ein Stück, das der Bestand nicht nahm, mit seinem Grund.
export interface Abweisung {
  stueck: Stueck
  ursache: unknown
}

// Stapelergebnis sagt, wie viele Stücke eines Stapels liegen — und, wenn
// nicht alle, welche nicht und warum.
export interface Stapelergebnis {
  gelegt: number

  // abgewiesen sind die Stücke, die der Bestand **einzeln** abwies — die
  // übrigen wurden trotzdem gelegt (Befund G16, #1666).
  abgewiesen: readonly Abweisung[]

  // ursache hielt den Stapel an: ein Grund, der nicht an einem Stück liegt.
  // Was danach kam, wurde nicht versucht.
  ursache?: unknown

  // offen sind alle Stücke, die nicht liegen, in der Reihenfolge des Stapels
  // — abgewiesene wie nicht versuchte. Leer heißt: alles liegt.
  offen: readonly Stueck[]
}

export interface Deckaktionen {
  legen: (
    karteID: string,
    printing: string,
    menge: number,
    bereich: string,
  ) => Promise<void>
  nehmen: (platzID: string) => Promise<void>

  // legeAlle legt mehrere Drucke der Reihe nach ins Hauptdeck und sagt, wie
  // viele liegen (UX-03.AK-01, #1675).
  legeAlle: (stuecke: readonly Stueck[]) => Promise<Stapelergebnis>
  bereichWechseln: (platz: DeckplatzDTO, bereich: string) => Promise<void>
  formatWaehlen: (formatID: string) => Promise<void>
  laeuft: boolean
}

export function useDeckaktionen(stand: Deckstand, deck: DeckDTO): Deckaktionen {
  const schluessel = `deck:${deck.id}`

  // schreibe führt eine Handlung am Deck aus und liest es **danach** neu —
  // nicht darin. Ein Lesen in `fuehre` zählt als neuer Lauf, und `fuehre`
  // hielt die Handlung dann für überholt: die Liste daneben wurde nicht
  // nachgeladen und zeigte die Zahlen von vorher (gefunden von der
  // WebKit-Probe für UX-03, #1675). Die Anlage eines Decks liest aus
  // demselben Grund danach (`Deckanlage.tsx`).
  //
  // **Nachgelesen wird nur nach einem Schreiben** (§7, Regel 3) — und nur,
  // was noch offen ist (`liesNach`).
  const schreibe = async (arbeit: () => Promise<boolean>) => {
    const ausgang = { geschrieben: false }

    await stand.fuehre(schluessel, async () => {
      ausgang.geschrieben = await arbeit()
    })

    if (ausgang.geschrieben) {
      await stand.liesNach(deck.id)
    }
  }

  const legen = async (
    karteID: string,
    printing: string,
    menge: number,
    bereich: string,
  ) => {
    await schreibe(async () => {
      const antwort = await api().DeckplatzLegen(
        deck.id,
        karteID,
        printing,
        menge,
        bereich,
      )

      auspacken(antwort, () => antwort.platz)

      return true
    })
  }

  const nehmen = async (platzID: string) => {
    await schreibe(async () => {
      auspacken(await api().DeckplatzNehmen(deck.id, platzID), () => undefined)

      return true
    })
  }

  // **Der Bereichswechsel ist ein Ruf** — Befund 7 des Zweitblicks über Phase
  // 5.1 (#1405).
  //
  // Hier standen **zwei**: erst legen, dann den alten nehmen. Scheiterte der
  // zweite, war der Zielplatz geschrieben und der Quellplatz stand weiter da —
  // dieselbe Karte lag in **beiden** Bereichen, und §7 Regel 3 nimmt nach
  // einem gescheiterten Schreiben nichts zurück. Über eine Naht mit zwei
  // Rufen ist das nicht zu heilen; ein Vorgang muss einer sein, und er ist es
  // seit `DeckplatzVerschieben` (eine Transaktion im Bestand).
  const bereichWechseln = async (platz: DeckplatzDTO, bereich: string) => {
    await schreibe(async () => {
      const antwort = await api().DeckplatzVerschieben(
        deck.id,
        platz.id,
        bereich,
      )

      auspacken(antwort, () => antwort.platz)

      return true
    })
  }

  const formatWaehlen = async (formatID: string) => {
    await schreibe(async () => {
      const antwort = await api().DeckFormatWaehlen(deck.id, formatID)

      auspacken(antwort, () => antwort.deck)

      return true
    })
  }

  // **Ein Vorgang für den ganzen Stapel und nicht einer je Druck** (§7,
  // Regel 4): der zweite Ruf käme sonst gegen einen Stand, den das Backend
  // noch nicht kennt.
  //
  // **Scheitert einer, bleibt liegen, was liegt** (Regel 3), und das Deck
  // wird nachgelesen, sobald etwas lag. **Der Fehler geht nicht über die
  // Seite**: er kommt als Ergebnis zurück, und der Stapel zeigt ihn dort, wo
  // gelegt wurde.
  //
  // # Ein abgewiesenes Stück hält die übrigen nicht auf
  //
  // **Befund G16 des Zweitblicks über Phase 6.3** (#1666): hier brach der
  // Stapel beim ersten Fehler ab, und was danach kam, wurde nicht versucht.
  // Dieselbe Gestalt hat #1995 beim Umschlag geschlossen, und Vermerk 0127
  // verwirft sie für die Sammlung: **eine Zeile scheitert, nicht die Liste**.
  //
  // **Der Schnitt ist der der Listeneinfuhr in Go**: eine Abweisung der
  // Eingabe (`validierung`) betrifft das Stück, und es geht weiter; alles
  // andere — ein fehlendes Deck, ein Profil nur zum Lesen, ein Systemfehler —
  // beträfe jedes, und dann hält der Stapel an.
  const legeAlle = async (
    stuecke: readonly Stueck[],
  ): Promise<Stapelergebnis> => {
    let gelegt = 0
    let ursache: unknown
    const abgewiesen: Abweisung[] = []
    const offen: Stueck[] = []

    await schreibe(async () => {
      for (const [stelle, stueck] of stuecke.entries()) {
        try {
          const antwort = await api().DeckplatzLegen(
            deck.id,
            '',
            stueck.printing,
            stueck.menge,
            'hauptdeck',
          )

          auspacken(antwort, () => antwort.platz)
          gelegt += 1
        } catch (fehler: unknown) {
          if (
            !(fehler instanceof Bindungsfehler) ||
            fehler.code !== 'validierung'
          ) {
            ursache = fehler
            offen.push(...stuecke.slice(stelle))

            break
          }

          abgewiesen.push({ stueck, ursache: fehler })
          offen.push(stueck)
        }
      }

      return gelegt > 0
    })

    return { gelegt, abgewiesen, ursache, offen }
  }

  return {
    legen,
    legeAlle,
    nehmen,
    bereichWechseln,
    formatWaehlen,
    laeuft: stand.vorgang.laeuft(schluessel),
  }
}
