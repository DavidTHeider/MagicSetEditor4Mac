// **Ein Stilwechsel lädt die Optionen des neuen Stils** — aus dem Zweitblick
// über Phase L (2026-08-22).
//
// # Der Befund
//
// `wechseln` lud nach erfolgreichem `StilWechseln` nur die **Stilangebote**
// nach und rief den äußeren Rückruf. `Stiloptionen` hat einen eigenen Lader,
// der allein an `setID` hängt, und `stilstand` stieg nur nach einer
// **Options**änderung.
//
// Ein Wechsel **innerhalb desselben Sets** ließ damit die Optionsliste des
// alten Stils stehen — und die Vorschau daneben ebenso.
//
// # Warum das §7 Regel 1 ist
//
// Das Bezugsobjekt dieser Seite ist nicht nur das Set, sondern **das Set in
// seinem Stil**. Wechselt er, gehört der offene Kontext geräumt: andere
// Optionen, andere Vorschau.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { waehleZeile } from '../test/suchwahl'
import { Stilwahl } from './Stilwahl'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

// stile gibt zwei Angebote — dieselbe Gestalt wie in `stilwahl.test.tsx`.
function stile() {
  return {
    stile: [
      {
        paket: 'probe-schlicht',
        anzeigename: 'Probe-schlicht',
        spiel: 'probe',
        version: '2026-08-15',
        eigenesSpiel: true,
        gewaehlt: true,
      },
      {
        paket: 'fremd',
        anzeigename: 'Fremd',
        spiel: 'probe',
        version: '2026-01-01',
        eigenesSpiel: true,
        gewaehlt: false,
      },
    ],
  }
}

it('lädt nach einem Stilwechsel die Optionen des neuen Stils', async () => {
  const optionen = vi.fn(() => Promise.resolve({ optionen: [] }))

  belegeApi({
    Stile: () => Promise.resolve(stile()),
    Stilfolgen: () =>
      Promise.resolve({
        folgen: {
          ungedeuteteFelder: [],
          verloreneOptionen: [],
          spielwechsel: false,
        },
      }),
    StilWechseln: () => Promise.resolve({}),
    Stiloptionen: optionen,
    Karten: () => Promise.resolve({ karten: [] }),
  })

  render(
    <Sprachraum sprache="de">
      <Stilwahl setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  // **Über die Auswahl** — seit UXR-07 (#1915) kein Optionsfeld je Stil, seit
  // UXR-F1 (#2007) eine mit Suche.
  await waehleZeile(
    await screen.findByRole('combobox', {
      name: texte.seiten.karten.stilWaehlen,
    }),
    'Fremd fremd',
  )

  // **Erst der Beleg, dass die Liste schon geladen hat** (§3): sonst zählte
  // die Verneinung darunter einen Lauf, den es nie gab.
  const vorher = optionen.mock.calls.length
  expect(vorher).toBeGreaterThan(0)

  await userEvent.click(
    screen.getByRole('button', { name: texte.seiten.karten.stilUebernehmen }),
  )

  await waitFor(() => {
    expect(optionen.mock.calls.length).toBeGreaterThan(vorher)
  })
})
