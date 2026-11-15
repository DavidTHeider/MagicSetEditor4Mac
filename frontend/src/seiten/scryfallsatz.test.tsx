// **Ein ganzer Satz von Scryfall** —
// #1033.
//
// Bis dahin kamen ganze Sets **nur** über MTG JSON. Damit fehlte der Einstieg,
// der für die meisten der naheliegende ist: „gib mir diesen Satz als Set."
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sets } from './Sets'
import { stillerSpeicherlauf } from '../test/speicherlauf'

// einlesenAufklappen öffnet den Kasten mit den Einfuhrwegen.
//
// **Seit dem 2026-09-03 sind sie zugeklappt** — Werkstück C von Phase 4.10
// (#1325):
// vier Wege standen über der Setliste, und das erste Set begann bei 915
// Bildpunkten, also unter der Falz. Was diese Tests prüfen, steht darin und
// ändert sich nicht; erreicht wird es über einen Klick mehr.
async function einlesenAufklappen() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.sets.einlesenTitel,
    }),
  )
}

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.sets

function zeige(teile = {}) {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [] }),
    Beispielsets: () => Promise.resolve({ sets: [] }),
    MTGJSONSpiele: () => Promise.resolve(['magic']),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        speichern={stillerSpeicherlauf()}
        waehlen={() => undefined}
      />
    </Sprachraum>,
  )
}

it('liest einen Satz über sein Kürzel ein', async () => {
  const ScryfallSatzEinlesen = vi.fn(() =>
    Promise.resolve({
      setID: '9',
      karten: 103,
      befunde: [],
      angelegt: ['9'],
      schonOffen: false,
      unvollstaendig: false,
      abgebrochen: false,
      fehler: null,
    }),
  )

  zeige({ ScryfallSatzEinlesen })
  await einlesenAufklappen()

  const feld = await screen.findByLabelText(texte.scryfallsatzCode)
  await userEvent.type(feld, 'lea')
  await userEvent.click(
    screen.getByRole('button', { name: texte.scryfallsatzEinlesen }),
  )

  await waitFor(() => {
    expect(ScryfallSatzEinlesen).toHaveBeenCalledWith('lea', 'magic')
  })
})

// **Ohne Kürzel ist der Knopf gesperrt.**
//
// Ein leerer Satzcode ergäbe eine Suche über alles — und die Bedingungen sagen,
// dass für viele Karten die Sammeldatei zu nehmen ist.
it('sperrt den Knopf ohne Kürzel', async () => {
  zeige()
  await einlesenAufklappen()

  expect(
    await screen.findByRole('button', { name: texte.scryfallsatzEinlesen }),
  ).toBeDisabled()
})

// **Ohne Spiel gibt es nichts einzulesen** — dieselbe Auskunft wie bei
// MTG JSON: ohne Zuordnung lässt sich keine Karte füllen.
it('sagt es, wenn es kein Spiel mit Zuordnung gibt', async () => {
  zeige({ MTGJSONSpiele: () => Promise.resolve([]) })
  await einlesenAufklappen()

  expect(
    await screen.findByText(texte.scryfallsatzKeinSpiel),
  ).toBeInTheDocument()

  expect(
    screen.queryByRole('button', { name: texte.scryfallsatzEinlesen }),
  ).not.toBeInTheDocument()
})

// **Die Spielwahl wirkt** — und der zweite Eintrag geht wirklich ans Backend.
//
// Ohne diesen Lauf hinge die Zuordnung immer am ersten Spiel, und niemand
// sähe es: mit nur einem Spiel in der Liste sieht beides gleich aus.
it('reicht das gewählte Spiel durch', async () => {
  const ScryfallSatzEinlesen = vi.fn(() =>
    Promise.resolve({
      setID: '9',
      karten: 1,
      befunde: [],
      angelegt: ['9'],
      schonOffen: false,
      unvollstaendig: false,
      abgebrochen: false,
      fehler: null,
    }),
  )

  zeige({
    ScryfallSatzEinlesen,
    MTGJSONSpiele: () => Promise.resolve(['magic', 'vs']),
  })
  await einlesenAufklappen()

  const feld = await screen.findByLabelText(texte.scryfallsatzCode)
  await userEvent.type(feld, 'lea')

  await userEvent.selectOptions(
    screen.getByLabelText(texte.scryfallsatzSpiel),
    'vs',
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.scryfallsatzEinlesen }),
  )

  await waitFor(() => {
    expect(ScryfallSatzEinlesen).toHaveBeenCalledWith('lea', 'vs')
  })
})
