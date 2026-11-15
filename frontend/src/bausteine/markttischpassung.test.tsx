// **Was zusammenpasst, steht am Markttisch beim Aufschlagen da** —
// #2494,
// H3 im Audit („Sehen, was die Gegenseite hat, und was zusammenpasst“,
// Schwelle 0 Handlungen). Bis hierher sagte es erst „Freigegebene Sicht
// anfragen“ — ein Klick, und dieselbe Sicht ging ein zweites Mal über die
// Leitung.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Markttisch } from './Markttisch'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { einAbgleich } from '../test/tauschabgleich'

import type { Handelsposition } from '../apihandelsleitung'
import type {
  Tauschabgleichantwort,
  Tauschkandidat,
} from '../apitauschabgleich'

const texte = kataloge.de.seiten.handeln.sicht
const tisch = kataloge.de.seiten.handeln.markttisch

afterEach(() => {
  raeumeApi()
})

const sturmrufer: Tauschkandidat = {
  wunsch: 'Sturmrufer',
  angebot: 'Sturmrufer',
  angebotDruck: 'p-sturm',
  angebotZustand: 'excellent',
  angebotSprache: 'en',
  angebotFinish: 'foil',
  menge: 1,
  fehlend: 0,
  rang: 'normal',
  art: 'genau',
}

function Buehne() {
  const [positionen, setzePositionen] = useState<Handelsposition[]>([])

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Markttisch
          ich="ich-1"
          gegenueber="du-1"
          name="Bert"
          positionen={positionen}
          setzePositionen={setzePositionen}
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

function zeige(
  abgleichen: () => Promise<Tauschabgleichantwort> = () =>
    // **Was Bert zeigt, steht auf deiner Wunschliste.**
    Promise.resolve(einAbgleich([sturmrufer])),
) {
  const sicht = vi.fn(() =>
    Promise.resolve({
      zeilen: [
        {
          referenz: 'r1',
          printing: 'p-sturm',
          menge: 2,
          zustand: 'excellent',
          sprache: 'en',
          finish: 'foil',
          bezeichnung: 'Sturmrufer',
        },
      ],
      fehler: null,
    }),
  )
  const abgleich = vi.fn(abgleichen)
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare: [], fehler: null }),
    HandelSichtAnfragen: sicht,
    Tauschabgleich: abgleich,
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  render(<Buehne />)

  return { sicht, abgleich }
}

describe('die Passung am Markttisch', () => {
  it('steht beim Aufschlagen da — ohne Klick und ohne zweite Anfrage', async () => {
    const { sicht, abgleich } = zeige()

    const gruppe = await screen.findByRole('group', { name: texte.titel })
    expect(within(gruppe).getByText('1 × Sturmrufer')).toBeVisible()
    expect(
      within(gruppe).getByRole('button', { name: texte.uebernehmen }),
    ).toBeEnabled()

    // **Aus der Sicht, die schon geholt ist** — derselbe Abgleich wie nach
    // dem Klick, ohne dass sie noch einmal über die Leitung geht.
    expect(sicht).toHaveBeenCalledTimes(1)
    expect(abgleich).toHaveBeenCalledWith(
      'du-1',
      {
        angebot: [expect.objectContaining({ printing: 'p-sturm' })],
        wuensche: [],
      },
      1,
      0,
    )
    expect(
      screen.queryByRole('button', { name: texte.anfragen }),
    ).not.toBeInTheDocument()
  })

  it('übernimmt den Vorschlag — und eine geänderte Sicht steht wieder offen', async () => {
    const { sicht } = zeige()

    const gruppe = await screen.findByRole('group', { name: texte.titel })
    await userEvent.click(
      within(gruppe).getByRole('button', { name: texte.uebernehmen }),
    )
    expect(
      await within(gruppe).findByText(texte.uebernommen.replace('{zahl}', '1')),
    ).toBeVisible()
    expect(
      within(gruppe).getByRole('button', { name: texte.uebernehmen }),
    ).toBeDisabled()

    // **Übernommen gilt für ein Ergebnis**: zeigt Bert danach etwas anderes,
    // steht der neue Vorschlag wieder offen. Dieselbe Sicht noch einmal
    // gefragt bleibt übernommen — sie ist derselbe Vorschlag.
    sicht.mockResolvedValueOnce({
      zeilen: [
        {
          referenz: 'r1',
          printing: 'p-sturm',
          menge: 3,
          zustand: 'excellent',
          sprache: 'en',
          finish: 'foil',
          bezeichnung: 'Sturmrufer',
        },
      ],
      fehler: null,
    })
    await userEvent.click(screen.getByRole('button', { name: tisch.neuFragen }))
    await vi.waitFor(async () => {
      const neu = await screen.findByRole('group', { name: texte.titel })
      expect(
        within(neu).getByRole('button', { name: texte.uebernehmen }),
      ).toBeEnabled()
    })
  })

  it('sagt am Ort, wenn der Abgleich scheitert', async () => {
    zeige(() => Promise.reject(new Error('der Abgleich klemmt')))

    expect(await screen.findByRole('alert')).toBeVisible()
    expect(
      screen.queryByRole('group', { name: texte.titel }),
    ).not.toBeInTheDocument()
  })
})
