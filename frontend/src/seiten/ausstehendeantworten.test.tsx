// **Wer noch wartet, hat keinen leeren Bestand** — UX-08.NAT-01 und
// S15.NAT-01 (#1680):
// *Bei gezielt verzögerter Antwort erscheint auf keiner betroffenen Seite eine
// Behauptung, der Bestand sei leer.*
//
// # Der Befund, gemessen am 2026-09-25
//
// Mit **jeder** Bindung unbeantwortet sagten fünf Seiten, es gebe nichts:
// „Noch kein Deck", „Noch kein Turnier", „Zum Spielen braucht es ein Deck",
// „Du hast keine Partie", „Noch kein Paket zum Bearbeiten", „Noch nichts
// gezeichnet". Wer ein langsames Profil öffnete, las, sein Bestand sei weg.
//
// # Wie hier geprüft wird
//
// **Alle Bindungen hängen**, nicht nur die eine, die man für die Liste hält —
// eine Seite fragt oft mehrere, und die, die man vergisst, ist die, die den
// Satz auslöst.
import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Handeln } from './Handeln'
import { Spielen } from './Spielen'
import { Turniere } from './Turniere'
import { Vorlagen } from './Vorlagen'
import { Symbol as Symbolseite } from './Symbol'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import vertrag from '../bindungsvertrag.json'

import type { Api } from '../api'
import type { ReactElement } from 'react'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

// haengend beantwortet keine einzige Bindung.
const haengend = Object.fromEntries(
  Object.keys(vertrag.methoden).map((name) => [
    name,
    () => new Promise(() => undefined),
  ]),
) as unknown as Partial<Api>

const faelle: { seite: string; zeige: () => ReactElement; leer: string[] }[] = [
  {
    seite: 'Decks',
    zeige: () => <Decks wechseln={vi.fn()} />,
    leer: [texte.seiten.decks.leer],
  },
  {
    seite: 'Turniere',
    zeige: () => <Turniere />,
    leer: [texte.seiten.turniere.leer],
  },
  {
    seite: 'Spielen',
    zeige: () => <Spielen wechseln={() => undefined} />,
    leer: [texte.seiten.spielen.leerDecks, texte.seiten.spielen.partienLeer],
  },
  {
    seite: 'Vorlagen',
    zeige: () => <Vorlagen />,
    leer: [texte.seiten.vorlagen.baumOhnePakete],
  },
  {
    seite: 'Handeln',
    zeige: () => <Handeln />,
    // **Die Liste der Handel** — die Gegenüber stehen seit UXR-15 in der
    // Schublade hinter „Neuer Tausch …" (#1923).
    leer: [texte.seiten.handeln.liste.leer],
  },
  {
    seite: 'Symbol',
    zeige: () => <Symbolseite setID="s1" />,
    leer: [texte.seiten.karten.symbol.keineTeile],
  },
]

describe('eine ausstehende Antwort', () => {
  for (const fall of faelle) {
    it(`gilt auf ${fall.seite} nicht als leerer Bestand`, async () => {
      belegeApi(haengend)
      render(<Sprachraum sprache="de">{fall.zeige()}</Sprachraum>)

      await act(async () => {
        await new Promise((weiter) => setTimeout(weiter, 20))
      })

      for (const satz of fall.leer) {
        expect(screen.queryByText(satz)).toBeNull()
      }

      expect(
        screen.getAllByText(texte.bausteine.zustaende.laedt).length,
      ).toBeGreaterThan(0)
    })
  }
})
