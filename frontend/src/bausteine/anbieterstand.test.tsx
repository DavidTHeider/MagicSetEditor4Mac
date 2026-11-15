// **Das Kontingent an einem Anbieterkonto** — Werkstück A5 von Phase 5.3
// (#1444).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Anbieterstand } from './Anbieterstand'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { Anbieterstandantwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// stand baut eine Antwort.
function stand(teil = {}): Anbieterstandantwort {
  return {
    dienst: 'scrydex',
    gesamt: 100000,
    rest: 84213,
    verbraucht: 15787,
    ueberzugssatz: '0.0004',
    hoechstensAltMinuten: 30,
    fehler: null,
    ...teil,
  }
}

// klickeHolen drückt den Abrufknopf.
async function klickeHolen() {
  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.kontingentHolenBei, { dienst: 'scrydex' }),
    }),
  )
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Anbieterstand />
    </Sprachraum>,
  )
}

it('fragt den Stand und zeigt ihn samt Frist', async () => {
  const gerufen = vi.fn(() => Promise.resolve(stand()))

  belegeApi({ Anbieterstand: gerufen })
  zeige()

  await klickeHolen()

  expect(gerufen).toHaveBeenCalledWith('scrydex')

  expect(
    await screen.findByText(
      einsetzen(eigene.kontingentZahlen, {
        rest: '84213',
        gesamt: '100000',
      }),
    ),
  ).toBeInTheDocument()

  // **Die Frist steht dabei** — ohne sie sähe der Rest aus wie eine Sperre.
  expect(
    screen.getByText(einsetzen(eigene.kontingentFrist, { minuten: '30' })),
  ).toBeInTheDocument()

  expect(
    screen.getByText(einsetzen(eigene.kontingentUeberzug, { satz: '0.0004' })),
  ).toBeInTheDocument()
})

it('zeigt einen Fehler am Ort', async () => {
  belegeApi({
    Anbieterstand: () =>
      Promise.resolve(
        stand({
          fehler: {
            code: 'validierung',
            meldung: 'scrydex braucht schluessel',
          },
        }),
      ),
  })
  zeige()

  await klickeHolen()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'scrydex braucht schluessel',
  )
})

// **Ein zweiter Ruf während des ersten geht nicht hinaus** (§7, Regel 4) —
// und hier zählt das doppelt: jede Abfrage kostet einen Kredit.
it('lässt keine zweite Abfrage zu, solange die erste läuft', async () => {
  let loese: (wert: Anbieterstandantwort) => void = () => undefined

  const gerufen = vi.fn(
    () =>
      new Promise<Anbieterstandantwort>((aufloesen) => {
        loese = aufloesen
      }),
  )

  belegeApi({ Anbieterstand: gerufen })
  zeige()

  await klickeHolen()

  expect(await screen.findByText(eigene.kontingentLaeuft)).toBeInTheDocument()

  // **Den Abrufknopf und nicht „den einen"** — seit UX-15
  // (#1771)
  // steht neben der Erläuterung ein Hilfeknopf, und `getByRole('button')`
  // ohne Namen war damit eine Aussage über die Fläche statt über den Abruf
  // (§10).
  await klickeHolen()

  expect(gerufen).toHaveBeenCalledTimes(1)

  loese(stand())

  expect(
    await screen.findByText(
      einsetzen(eigene.kontingentZahlen, {
        rest: '84213',
        gesamt: '100000',
      }),
    ),
  ).toBeInTheDocument()
})
