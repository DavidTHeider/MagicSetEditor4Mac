// **Die Befunde des Preisholens erreichen das Fenster** —
// #1202.
//
// # Der Befund
//
// Go reicht unlesbare Preisfelder ausdrücklich als `befunde` bis zur
// Wails-Antwort, und Kommandozeile, MCP und Dienst geben sie aus. `Preise.tsx`
// las sie **nirgends**: bei einem teilweise verständlichen Fremddatensatz sah
// die Person einen erfolgreichen Abruf und die übrigen Preise — nicht, dass
// eine Reihe ausgelassen wurde.
//
// **Und im Fehlerfall gingen sie ganz verloren**: `auspacken` warf nur den
// Fehler. Das betrifft jede Bindung mit Befunden, nicht nur diese — deshalb
// trägt `Bindungsfehler` sie jetzt mit.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { PreisstandDTO } from '../antwortensammlung'
import { Preise } from './Preise'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.sammlung

const einStand: PreisstandDTO = {
  printing: 'probe',
  druckID: '',
  quelle: 'scryfall',
  markt: 'paper',
  art: 'retail',
  zustand: '',
  sprache: '',
  lage: 'genannt',
  waehrung: 'EUR',
  finish: 'nonfoil',
  wert: '1.50',
  beobachtet: '',
  erhoben: '2026-08-28T00:00:00Z',
  alterStunden: 0,
}

function zeige(teile = {}) {
  belegeApi({
    Preise: () =>
      Promise.resolve({ staende: [einStand], befunde: [], fehler: null }),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Preise printing="probe" aufGeholt={vi.fn()} />
    </Sprachraum>,
  )
}

// **Ein geglückter Abruf mit Befunden zeigt sie.**
it('zeigt die Befunde eines geglückten Holens', async () => {
  zeige({
    PreiseHolen: () =>
      Promise.resolve({
        staende: [einStand],
        befunde: ['die Reihe USD foil war nicht zu deuten'],
        fehler: null,
      }),
  })

  await userEvent.click(
    await screen.findByRole('button', { name: `${eigene.preiseHolen} probe` }),
  )

  expect(await screen.findByText(eigene.preisbefunde)).toBeInTheDocument()
  expect(
    await screen.findByText('die Reihe USD foil war nicht zu deuten'),
  ).toBeInTheDocument()
})

// **Und ein gescheiterter zeigt sie auch.**
//
// Das ist die Hälfte, die zählt: ein Abruf, der an einer Reihe scheitert, hat
// die anderen gelesen. Wer nur den Fehler zeigte, verschwiege sie.
it('zeigt die Befunde neben einem Fehler', async () => {
  zeige({
    PreiseHolen: () =>
      Promise.resolve({
        staende: [],
        befunde: ['die Reihe TIX war nicht zu deuten'],
        fehler: {
          code: 'unbekannt' as const,
          schluessel: 'unbekannt',
          meldung: 'Scryfall antwortet nicht',
        },
      }),
  })

  await userEvent.click(
    await screen.findByRole('button', { name: `${eigene.preiseHolen} probe` }),
  )

  expect(await screen.findByRole('alert')).toBeInTheDocument()
  expect(
    await screen.findByText('die Reihe TIX war nicht zu deuten'),
  ).toBeInTheDocument()
})

// **Die Befunde des Lesens stehen ebenso da.**
//
// Sie kommen aus derselben Quelle wie die des Holens — und wer nur die des
// Holens zeigte, ließe jemanden, der die Seite bloß aufschlägt, im Dunkeln.
it('zeigt die Befunde des Lesens', async () => {
  zeige({
    Preise: () =>
      Promise.resolve({
        staende: [einStand],
        befunde: ['die Reihe EUR foil war nicht zu deuten'],
        fehler: null,
      }),
  })

  expect(
    await screen.findByText('die Reihe EUR foil war nicht zu deuten'),
  ).toBeInTheDocument()
})

// **Ein Fehler, der keiner der Bindung ist, bringt keine Befunde mit.**
//
// # Warum dieser Fall danebensteht
//
// Weil er der einzige ist, in dem die Liste **leer** bleiben muss. Ein
// abgestürzter Aufruf — die Fassade unbelegt, die Leitung weg — hat nichts
// gelesen, und alte Befunde daneben stehen zu lassen, hieße etwas über einen
// Abruf zu sagen, der gar nicht stattfand.
it('zeigt keine Befunde, wenn schon der Aufruf scheitert', async () => {
  zeige({
    PreiseHolen: () => Promise.reject(new Error('die Leitung ist weg')),
  })

  await userEvent.click(
    await screen.findByRole('button', { name: `${eigene.preiseHolen} probe` }),
  )

  expect(await screen.findByRole('alert')).toBeInTheDocument()
  expect(screen.queryByText(eigene.preisbefunde)).not.toBeInTheDocument()
})
