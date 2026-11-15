// **Die Angebote eines Marktplatzes** — Werkstück A4 von Phase 5.3
// (#1443).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Marktangebote } from './Marktangebote'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Marktantwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.sammlung

// zeige zeichnet das Bauteil.
function zeige(printing = 'probe-druck') {
  render(
    <Sprachraum sprache="de">
      <Marktangebote printing={printing} />
    </Sprachraum>,
  )
}

// angebot baut einen Posten.
function angebot(teil = {}) {
  return {
    quelle: 'cardtrader',
    kennung: '101862104',
    wert: '2.17',
    waehrung: 'EUR',
    menge: 3,
    buendel: 1,
    finish: 'nonfoil',
    zustand: 'Moderately Played',
    sprache: 'en',
    signiert: false,
    veraendert: false,
    gradiert: false,
    imUrlaub: false,
    verkaeufer: 'einer',
    land: 'FI',
    ...teil,
  }
}

// antwort baut eine Marktantwort.
function antwort(teil = {}) {
  return {
    quelle: 'cardtrader',
    fassung: '89',
    zugeordnet: false,
    angebote: [angebot()],
    befunde: [],
    fehler: null,
    ...teil,
  }
}

// klickeHolen drückt den Abrufknopf.
//
// **Den Abrufknopf und nicht „den einen"** — seit UX-15
// (#1771)
// steht neben der Erläuterung ein Hilfeknopf, und `getByRole('button')` ohne
// Namen war damit eine Aussage über die Fläche statt über den Abruf (§10).
async function klickeHolen() {
  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.marktHolenBei, { quelle: 'cardtrader' }),
    }),
  )
}

it('holt die Angebote für den gezeigten Druck', async () => {
  const gerufen = vi.fn(() => Promise.resolve(antwort()))

  belegeApi({ Marktangebote: gerufen })
  zeige()

  await klickeHolen()

  expect(gerufen).toHaveBeenCalledWith('cardtrader', 'probe-druck', '', '', [])

  expect(await screen.findByText(/2,17\s€/u)).toBeInTheDocument()
  expect(screen.getByText(/Moderately Played/)).toBeInTheDocument()
  expect(screen.getByText(/einer \(FI\)/)).toBeInTheDocument()
})

it('sagt es, wenn nichts angeboten wird', async () => {
  belegeApi({
    Marktangebote: () => Promise.resolve(antwort({ angebote: [] })),
  })
  zeige()

  await klickeHolen()

  expect(await screen.findByText(eigene.marktLeer)).toBeInTheDocument()
})

it('nennt die gefundene Zuordnung', async () => {
  belegeApi({
    Marktangebote: () =>
      Promise.resolve(antwort({ zugeordnet: true, fassung: '89' })),
  })
  zeige()

  await klickeHolen()

  expect(
    await screen.findByText(
      einsetzen(eigene.marktZugeordnet, { fassung: '89' }),
    ),
  ).toBeInTheDocument()
})

// **Die Merkmale stehen dabei und werden nicht weggelassen.**
//
// Ein Posten, der da ist und nicht kommt, ist eine andere Auskunft als keiner
// — und dasselbe gilt für signiert, verändert und bewertet.
it('zeigt Bündel, Urlaub und die drei Merkmale', async () => {
  belegeApi({
    Marktangebote: () =>
      Promise.resolve(
        antwort({
          angebote: [
            angebot({
              buendel: 4,
              imUrlaub: true,
              signiert: true,
              veraendert: true,
              gradiert: true,
              land: '',
            }),
          ],
        }),
      ),
  })
  zeige()

  await klickeHolen()

  expect(await screen.findByText(eigene.marktImUrlaub)).toBeInTheDocument()
  expect(screen.getByText(eigene.marktSigniert)).toBeInTheDocument()
  expect(screen.getByText(eigene.marktVeraendert)).toBeInTheDocument()
  expect(screen.getByText(eigene.marktGradiert)).toBeInTheDocument()
  expect(
    screen.getByText(einsetzen(eigene.marktBuendel, { buendel: '4' })),
  ).toBeInTheDocument()

  // **Ohne Land steht keine leere Klammer da.**
  expect(screen.getByText('einer')).toBeInTheDocument()
})

it('zeigt die Befunde der Quelle', async () => {
  belegeApi({
    Marktangebote: () =>
      Promise.resolve(antwort({ befunde: ['ein Posten ohne Währung'] })),
  })
  zeige()

  await klickeHolen()

  expect(await screen.findByText(eigene.marktBefunde)).toBeInTheDocument()
  expect(screen.getByText('ein Posten ohne Währung')).toBeInTheDocument()
})

it('zeigt einen Fehler am Ort und lässt die Liste leer', async () => {
  belegeApi({
    Marktangebote: () =>
      Promise.resolve(
        antwort({
          angebote: [],
          fehler: { code: 'validierung', meldung: 'cardtrader braucht marke' },
        }),
      ),
  })
  zeige()

  await klickeHolen()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'cardtrader braucht marke',
  )
})

// **Ein zweiter Ruf während des ersten geht nicht hinaus** (§7, Regel 4).
it('lässt keinen zweiten Abruf zu, solange der erste läuft', async () => {
  let loese: (wert: Marktantwort) => void = () => undefined

  const gerufen = vi.fn(
    () =>
      new Promise<Marktantwort>((aufloesen) => {
        loese = aufloesen
      }),
  )

  belegeApi({ Marktangebote: gerufen })
  zeige()

  await klickeHolen()

  expect(await screen.findByText(eigene.marktLaeuft)).toBeInTheDocument()

  // **Der Abrufknopf und nicht irgendeiner** — seit UX-15
  // (#1771)
  // steht neben der Erläuterung ein Hilfeknopf. `getByRole('button')` fand
  // vorher genau einen und war damit eine Aussage über die Fläche, nicht
  // über den Abruf (§10).
  await klickeHolen()

  expect(gerufen).toHaveBeenCalledTimes(1)

  loese(antwort())

  expect(await screen.findByText(/2,17\s€/u)).toBeInTheDocument()
})
