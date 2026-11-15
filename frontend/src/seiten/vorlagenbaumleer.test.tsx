// **Ein leerer Vorlagenbaum führt weiter** — G02
// (#2249).
//
// „Lege unter Pakete eines an" verlangte einen Weg, den man sich merken muss,
// und „Kein Paket passt zur Suche" ließ die Suche stehen, die es verursacht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'
import { Vorlagenbaum } from './Vorlagenbaum'

const texte = kataloge.de.seiten.vorlagen
const sucheLeeren = kataloge.de.bausteine.listenansicht.sucheLeeren

afterEach(() => {
  raeumeApi()
})

function zeige(pakete: string[], zuDenPaketen = vi.fn()) {
  belegeApi({
    Paketdateien: () =>
      Promise.resolve({ dateien: ['style'], beiwerk: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagenbaum
        geladen
        fehlschlag={false}
        pakete={pakete}
        gewaehlt={null}
        waehlen={vi.fn()}
        geaendert={false}
        zuDenPaketen={zuDenPaketen}
      />
    </Sprachraum>,
  )

  return zuDenPaketen
}

describe('Leerzustände des Vorlagenbaums', () => {
  it('führt ohne Paket zu den Paketen', async () => {
    const zuDenPaketen = zeige([])

    expect(screen.getByText(texte.baumOhnePakete)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.allgemein.arbeitsplatz.zuDenPaketen,
      }),
    )

    expect(zuDenPaketen).toHaveBeenCalledOnce()
  })

  it('leert eine Suche ohne Treffer', async () => {
    zeige(['/pakete/probe.mse-style'])

    const nutzer = userEvent.setup()

    await screen.findByText('probe.mse-style')
    await nutzer.type(
      screen.getByRole('searchbox', { name: texte.baumSuche }),
      'gibtsnicht',
    )

    expect(await screen.findByText(texte.baumLeer)).toBeInTheDocument()

    await nutzer.click(screen.getByRole('button', { name: sucheLeeren }))

    expect(await screen.findByText('probe.mse-style')).toBeInTheDocument()
    expect(
      screen.getByRole('searchbox', { name: texte.baumSuche }),
    ).toHaveValue('')
  })

  it('wechselt von der Seite aus zu den Paketen', async () => {
    belegeWerkstatt({
      Pakete: () => Promise.resolve({ pakete: [], befunde: [], fehler: null }),
    })

    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Vorlagen wechseln={wechseln} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.allgemein.arbeitsplatz.zuDenPaketen,
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('pakete')
  })
})
