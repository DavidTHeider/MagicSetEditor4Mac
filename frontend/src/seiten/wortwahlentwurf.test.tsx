// **Eine Wahl aus der Auswahlliste trägt die Werte des Entwurfs mit** —
// Befund 1 des Nachlaufs über Phase 0.7.0-I zu U02b
// (#2179).
//
// Die Wahl rechnet in `service` aus dem gespeicherten Wert und schrieb ohne
// Bedingung: stand das Feld sichtbar im Konflikt, änderte sie den fremden
// Wert, bevor jemand entschieden hatte. Jetzt schickt sie die Werte, die der
// Entwurf kennt — im Konflikt **ohne** den fremden, und dann weist der Bestand
// die Wahl ab (`app.WortWaehlenBedingt`).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  FeldbeschreibungDTO,
  KarteDTO,
  VorschauAntwort,
  WortlistenbereichDTO,
} from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const entwurf = kataloge.de.bausteine.entwurf

const vorschau: VorschauAntwort = {
  vorschau: {
    stil: '',
    gezeichnet: '',
    bild: '/vorschau/AAAA.png',
    breite: 10,
    hoehe: 14,
    dpi: 150,
    befunde: [],
    flaechen: [{ feld: 'type', links: 0, oben: 0, breite: 10, hoehe: 5 }],
    nennungen: [],
    vorgaben: [],
  },
}

const form: FeldbeschreibungDTO = {
  name: 'type',
  aufschrift: 'type',
  art: 'text',
  mehrzeilig: false,
  wahlen: [],
  beschreibung: '',
  vorgabe: '',
  teilfelder: [],
  reihe: '',
  stelle: 0,
  editierbar: true,
  gespeichert: true,
}

const bereich: WortlistenbereichDTO = {
  liste: 'type-en',
  text: 'Creature',
  kaestchen: true,
  gewaehlt: 1,
  eintraege: [
    {
      name: 'Token ',
      istPraefix: true,
      zeileDarunter: false,
      aktiv: false,
      kinder: null,
    },
    {
      name: 'Creature',
      istPraefix: false,
      zeileDarunter: false,
      aktiv: true,
      kinder: null,
    },
  ],
}

const karte = (type: string): KarteDTO => ({
  id: '5',
  nummer: 0,
  werte: [{ feld: 'type', wert: type }],
})

function zeige() {
  const WortWaehlen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau),
    Probevorschau: () => Promise.resolve(vorschau),
    Symbolpalette: () =>
      Promise.resolve({ eintraege: [], roh: false, fehler: null }),
    Wortlisten: () => Promise.resolve({ bereiche: [bereich], fehler: null }),
    WortWaehlen,
  })

  const seite = (type: string) => (
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte(type)}
        formen={[form]}
        geaendert={() => undefined}
      />
    </Sprachraum>
  )

  const { rerender } = render(seite('Creature'))

  return {
    WortWaehlen,
    fremd: (type: string) => {
      rerender(seite(type))
    },
  }
}

describe('Eine Wahl aus der Auswahlliste', () => {
  it('schickt den gezeigten Wert als bekannt', async () => {
    const { WortWaehlen } = zeige()

    await userEvent.click(await screen.findByLabelText('type'))
    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Token' }),
    )

    await waitFor(() => {
      expect(WortWaehlen).toHaveBeenCalledWith(
        '1',
        '5',
        'type',
        0,
        [0],
        ['Creature'],
      )
    })
  })

  it('schickt im Konflikt die Werte des Entwurfs, ohne den fremden', async () => {
    const { WortWaehlen, fremd } = zeige()

    const feld = await screen.findByLabelText('type')
    await userEvent.click(feld)
    await userEvent.type(feld, ' eigener Entwurf')
    fremd('Sorcery')

    expect(
      await screen.findByRole('button', { name: entwurf.behalten }),
    ).toBeInTheDocument()

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Token' }),
    )

    await waitFor(() => {
      expect(WortWaehlen).toHaveBeenCalledWith(
        '1',
        '5',
        'type',
        0,
        [0],
        ['Creature'],
      )
    })
  })
})
