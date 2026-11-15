// **Die Bedienfläche der Massenänderung** — Werkstück 5 von Phase 3.5
// (#1166).
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2):
// `massenaenderung.test.tsx` prüft, was die Änderung **tut**; hier steht, ob
// ein Mensch sie überhaupt auslösen kann, ohne die Vorlage auswendig zu
// kennen.
import { render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenzeileDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Massenaenderung } from './Massenaenderung'

const karten: KartenzeileDTO[] = [
  { id: '1', nummer: 0, name: 'Feuerball', felder: 1 },
]

const feldwerte = {
  Kartenfeldwerte: () => Promise.resolve({ werte: [] }),
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

describe('die Feldreihen der Massenänderung', () => {
  it('stellt die Glieder einer Reihe unter ihr Grundfeld', async () => {
    belegeApi({
      ...feldwerte,
      Kartenfelder: vi.fn().mockResolvedValue({
        felder: [
          {
            name: 'name',
            aufschrift: 'Name',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: 'name',
            stelle: 1,
          },
          {
            name: 'name 2',
            aufschrift: 'Name 2',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: 'name',
            stelle: 2,
          },
          {
            name: 'rarity',
            aufschrift: 'Rarity',
            art: 'text',
            mehrzeilig: false,
            wahlen: [],
            beschreibung: '',
            vorgabe: '',
            editierbar: true,
            gespeichert: true,
            teilfelder: [],
            reihe: '',
            stelle: 0,
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    // **Die Gruppe trägt die Aufschrift des Grundfeldes** — `optgroup` hat die
    // Rolle `group`, und ihr `label` ist ihr zugänglicher Name.
    const gruppe = await screen.findByRole('group', { name: 'Name' })

    expect(
      within(gruppe).getByRole('option', { name: 'Name 2' }),
    ).toBeInTheDocument()

    // **Und ein Feld ohne Reihe steht nicht darin.**
    expect(
      within(gruppe).queryByRole('option', { name: 'Rarity' }),
    ).not.toBeInTheDocument()

    expect(screen.getByRole('option', { name: 'Rarity' })).toBeInTheDocument()
  })
})
