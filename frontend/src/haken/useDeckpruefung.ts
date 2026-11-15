// **Die laufende Regelprüfung** — Werkstück U10 von Phase 5.1
// (#1421).
//
// # Warum ein Haken und nicht ein Bauteil
//
// Weil das Urteil an **zwei** Stellen steht: am einzelnen Posten im Zentrum
// und im Kopf des Inspectors. Das ist die Abnahme aus dem Issue — „illegaler
// Posten erscheint direkt am Posten und im Kopf" —, und zweimal zu prüfen
// hieße, zwei Antworten auf dieselbe Frage zu haben.
//
// # Was der Schlüssel ist
//
// Der **Abzug der Änderung**: Format und Plätze. Ein Urteil zu einem Deck,
// das inzwischen anders aussieht, ist keine Auskunft, sondern eine Falle.
// `useLaden` verwirft die Antwort eines überholten Laufs (§7, Regel 2) — wer
// dreimal tippt, sieht das Urteil zum dritten Stand.
//
// **Was nicht gebaut ist**: ein harter Abbruch der laufenden Anfrage. Die
// Prüfung ist **eine** Sammelabfrage an Scryfall; ihre Antwort wird beim
// nächsten Wechsel ohnehin verworfen, und ein Abbruch spart eine Anfrage,
// die niemand liest. Wer ihn will, braucht eine Bindung wie
// `SkriptAbbrechen` — ein eigenes Werkstück, kein Nebeneffekt.
import { useCallback } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { DeckDTO, DeckurteilDTO } from '../apidecks'

export interface Deckpruefung {
  urteil: DeckurteilDTO | null
  befunde: readonly string[]
  laedt: boolean
  fehler: string | undefined
  nachladen: () => void

  // ohneFormat sagt, dass gar nicht gefragt wurde — **nicht**, dass die
  // Antwort leer war. Das Fenster zeigt dafür einen Satz statt eines Urteils,
  // und der Knopf ruht.
  ohneFormat: boolean

  // unzulaessig sagt zu einem Platz, ob sein Posten abgewiesen ist.
  unzulaessig: (platzID: string) => boolean
}

export function useDeckpruefung(deck: DeckDTO): Deckpruefung {
  const abzug = JSON.stringify([
    deck.formatID,
    deck.plaetze.map((platz) => [platz.id, platz.menge, platz.bereich]),
  ])

  const holen = useCallback(async () => {
    // **Ohne Format wird nicht gefragt**, und das ist keine Bequemlichkeit:
    // `service.Formatpruefung.Pruefe` weist ein Deck ohne Format mit
    // `ErrValidation` ab — „gegen was soll geprüft werden?". Solange jemand
    // den Knopf drückte, war das die richtige Antwort auf eine gestellte
    // Frage. Seit die Prüfung von selbst läuft, stellt sie **niemand**, und
    // im Inspector stand bei jedem frisch angelegten Deck ein Fehler.
    //
    // **Gesehen am 2026-09-09 in der Sichtprüfung**, nicht überlegt: dort
    // stand „Es ist ein unerwarteter Fehler aufgetreten" — der Rückfall des
    // Katalogs, denn eine Validierung ohne Feldbezug trägt den Schlüssel
    // `unbekannt` (`internal/app/fehler.go`). Ein Deck ohne Format ist
    // ungeprüft, nicht falsch ([0061](../../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md)).
    if (deck.formatID === '') {
      return { urteil: null, befunde: [] }
    }

    const antwort = await api().DeckPruefen(deck.id)

    return {
      urteil: auspacken(antwort, () => antwort.urteil),
      befunde: antwort.befunde,
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- der Abzug ist der Schlüssel
  }, [deck.id, deck.formatID, abzug])

  // **`behalten` steht hier nicht**: wer das Deck ändert, soll kein Urteil
  // zum vorigen Stand sehen — es verschwindet, bis das neue da ist.
  const geladen = useLaden(holen, [deck.id, abzug])

  const urteil = geladen.daten?.urteil ?? null

  return {
    urteil,
    befunde: geladen.daten?.befunde ?? [],
    laedt: geladen.laedt,
    fehler: geladen.fehler,
    nachladen: geladen.nachladen,
    ohneFormat: deck.formatID === '',

    // **`legal` und `restricted` sind zulässig, `banned` und `not_legal`
    // nicht** — und `ungeprüft` **auch nicht unzulässig** (Vermerk 0061):
    // eine eigene Karte steht in keiner Liste der Welt, und sie rot zu
    // färben hieße, jedem eigenen Deck zu sagen, es sei kaputt.
    unzulaessig: (platzID: string) =>
      (urteil?.urteile ?? []).some(
        (eines) =>
          eines.platzID === platzID &&
          (eines.zulaessigkeit === 'banned' ||
            eines.zulaessigkeit === 'not_legal'),
      ),
  }
}
