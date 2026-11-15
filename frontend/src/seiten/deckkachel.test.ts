// **Was die Kachel „Decks“ unter ihre Zahl schreibt** — AP-8, Teil 2b (#2365).
import { describe, expect, it } from 'vitest'

import type { AufmerksamkeitAntwort, HinweisDTO } from '../apiaufmerksamkeit'
import { kataloge } from '../sprache/kataloge'
import { deckkachel } from './deckkachel'

const texte = kataloge.de

function hinweis(
  art: HinweisDTO['art'],
  kennung: string,
  name: string,
): HinweisDTO {
  return {
    art,
    objekt: {
      art: art === 'deckFehlend' ? 'deck' : 'set',
      kennung,
      name,
      angeheftet: false,
    },
    anzahl: 2,
    grund: '',
    ort: '',
    schluessel: '',
    werte: [],
  }
}

function antwort(hinweise: HinweisDTO[]): AufmerksamkeitAntwort {
  return {
    hinweise,
    durchgesehen: { sets: 1, decks: 2, pakete: 0 },
    befunde: [],
    fehler: null,
  }
}

describe('die Kachel der Decks', () => {
  it('schweigt ohne Antwort und ohne Deck mit fehlenden Karten', () => {
    expect(deckkachel(undefined, texte)).toEqual({})
    expect(
      deckkachel(antwort([hinweis('setVorbehalte', '1', 'Set')]), texte),
    ).toEqual({})
  })

  // **Keine Null**: ein Deck, dessen Abgleich scheiterte, stünde nicht in
  // den Hinweisen — „0 mit fehlenden Karten“ wäre dann eine Behauptung.
  it('zählt nur die Decks mit fehlenden Karten', () => {
    expect(
      deckkachel(
        antwort([
          hinweis('deckFehlend', '1', 'Waldschritt'),
          hinweis('setVorbehalte', '9', 'Set'),
          hinweis('deckFehlend', '3', 'Moorwacht'),
        ]),
        texte,
      ),
    ).toEqual({
      unterzeile: '2 mit fehlenden Karten',
      ziel: {
        art: 'deckauswahl',
        name: '2 Decks mit fehlenden Karten',
        deckIDs: ['1', '3'],
      },
    })
  })

  it('öffnet ein einzelnes Deck mit seiner Prüfung', () => {
    expect(
      deckkachel(antwort([hinweis('deckFehlend', '3', 'Moorwacht')]), texte),
    ).toEqual({
      unterzeile: '1 mit fehlenden Karten',
      ziel: {
        art: 'deckansicht',
        name: 'Moorwacht',
        deckID: '3',
        lasche: 'pruefung',
      },
    })
  })
})
