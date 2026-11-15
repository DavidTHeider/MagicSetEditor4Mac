// **Ein Fehlschlag bietet den nächsten Versuch, ein leeres Ergebnis den
// nächsten Schritt** — UX-08 S15.NAT-02
// (#1680):
// *Ein erfolgreiches leeres Ergebnis bietet die passende Anlageaktion; eine
// fehlgeschlagene Abfrage bietet Wiederholen und keinen irreführenden
// Leerhinweis.*
//
// Das Gegenstück zu `ausstehendeantworten.test.tsx`: dort steht, dass eine
// ausstehende Antwort nichts behauptet — hier, dass die zwei anderen Ausgänge
// jeweils das Ihre sagen. **Der leere Fall ist zugleich der Beleg, dass der
// Wächter dort zuhört**: derselbe Satz, der beim Warten fehlen muss, steht
// hier da.
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Handeln } from './Handeln'
import { Spielen } from './Spielen'
import { Turniere } from './Turniere'
import { Vorlagen } from './Vorlagen'
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

const haengend = Object.fromEntries(
  Object.keys(vertrag.methoden).map((name) => [
    name,
    () => new Promise(() => undefined),
  ]),
) as unknown as Partial<Api>

async function zeige(
  seite: ReactElement,
  teile: Record<string, unknown>,
): Promise<void> {
  belegeApi({ ...haengend, ...(teile as Partial<Api>) })
  render(<Sprachraum sprache="de">{seite}</Sprachraum>)

  await act(async () => {
    await new Promise((weiter) => setTimeout(weiter, 20))
  })
}

const scheitert = () => Promise.reject(new Error('Die Leitung ist weg.'))

const faelle: {
  seite: string
  zeige: () => ReactElement
  liste: string[]
  leer: Record<string, unknown>
  satz: string

  // versuche ist die Zahl der Knöpfe „Erneut laden", wo eine Liste an mehr
  // als einer Stelle der Seite steht — ohne Angabe eine je Liste.
  versuche?: number
}[] = [
  {
    seite: 'Decks',
    zeige: () => <Decks wechseln={vi.fn()} />,
    liste: ['Decks'],
    leer: {
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () => Promise.resolve({ formate: [], fehler: null }),
    },
    satz: texte.seiten.decks.leer,
  },
  {
    seite: 'Turniere',
    zeige: () => <Turniere />,
    liste: ['Turniere'],
    leer: { Turniere: () => Promise.resolve({ turniere: [], fehler: null }) },
    satz: texte.seiten.turniere.leer,
  },
  {
    seite: 'Spielen',
    zeige: () => <Spielen wechseln={() => undefined} />,
    // **Beide Listen der Seite** — die Partien und die Decks des Anfangs.
    // **Die Leihe fragt ihre Decks erst, wenn ihre Schublade aufgeht** —
    // seit UXR-13 (#1921); bis dahin stand sie offen auf der Seite, und ihr
    // Versuch war der dritte. Ihren eigenen bietet sie dort weiter
    // (`deckleihe.test.tsx`).
    liste: ['Partien', 'Decks'],
    leer: { Partien: () => Promise.resolve({ partien: [], fehler: null }) },
    satz: texte.seiten.spielen.partienLeer,
  },
  {
    seite: 'Vorlagen',
    zeige: () => <Vorlagen />,
    liste: ['Pakete'],
    leer: {
      Pakete: () => Promise.resolve({ pakete: [], befunde: [], fehler: null }),
    },
    satz: texte.seiten.vorlagen.baumOhnePakete,
  },
  // **Seit UXR-15 lädt Handeln beim Aufschlagen die Liste der Handel** —
  // die Gegenüber erst in der Schublade hinter „Neuer Tausch …" (#1923).
  {
    seite: 'Handeln',
    zeige: () => <Handeln />,
    liste: ['HandelUebersicht'],
    leer: {
      HandelUebersicht: () => Promise.resolve({ zeilen: [], fehler: null }),
    },
    satz: texte.seiten.handeln.liste.leer,
  },
]

describe('ein Fehlschlag beim Laden', () => {
  for (const fall of faelle) {
    it(`bietet auf ${fall.seite} den nächsten Versuch und keinen Leerhinweis`, async () => {
      const gefragt = vi.fn(scheitert)
      await zeige(
        fall.zeige(),
        Object.fromEntries(fall.liste.map((name) => [name, gefragt])),
      )

      expect(screen.queryByText(fall.satz)).toBeNull()

      const vorher = gefragt.mock.calls.length
      const versuche = screen.getAllByRole('button', {
        name: texte.bausteine.zustaende.wiederholen,
      })
      expect(versuche).toHaveLength(fall.versuche ?? fall.liste.length)
      for (const versuch of versuche) await userEvent.click(versuch)

      await act(async () => {
        await new Promise((weiter) => setTimeout(weiter, 20))
      })
      expect(gefragt.mock.calls.length).toBeGreaterThan(vorher)
    })
  }
})

describe('ein leeres Ergebnis', () => {
  for (const fall of faelle) {
    it(`sagt auf ${fall.seite}, was als Nächstes dran ist`, async () => {
      await zeige(fall.zeige(), fall.leer)

      expect(await screen.findByText(fall.satz)).toBeInTheDocument()
      expect(
        screen.queryByRole('button', {
          name: texte.bausteine.zustaende.wiederholen,
        }),
      ).toBeNull()
    })
  }
})
