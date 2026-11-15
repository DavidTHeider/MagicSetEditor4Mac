// **Liste und Abschnitt lesen die Abhängigkeiten einmal** — U14
// (#2159).
//
// Die Zahl an jeder Paketzeile und der Abschnitt „Abhängigkeiten" darunter
// fragten je für sich — sobald der Kasten offen war, zwei Rufe für dieselbe
// Frage, und mit verschiedenen Ständen: die Liste räumte bei jedem neuen
// Stand, der Abschnitt las still nach.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AbhaengigkeitDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { belegeSeite, oeffneKasten, texte } from '../test/paketseite'
import { Pakete } from './Pakete'

afterEach(() => {
  raeumeApi()
})

const befund: AbhaengigkeitDTO = {
  paket: 'probe-schlicht.mse-style',
  braucht: 'probe.mse-game',
  art: 'fehlt',
  verlangt: '',
  da: '',
}

describe('die Abhängigkeiten der Paketseite', () => {
  it('werden für Liste und Abschnitt einmal gelesen, auch bei neuem Stand', async () => {
    const gefragt = vi.fn(() => Promise.resolve({ befunde: [befund] }))
    belegeSeite({ Abhaengigkeiten: gefragt })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Pakete wechseln={vi.fn()} bestandsstand={0} />
      </Sprachraum>,
    )

    await oeffneKasten(texte.seiten.pakete.abhaengigUeberschrift)
    expect(await screen.findByText(/probe\.mse-game/)).toBeInTheDocument()
    expect(gefragt).toHaveBeenCalledTimes(1)

    rerender(
      <Sprachraum sprache="de">
        <Pakete wechseln={vi.fn()} bestandsstand={1} />
      </Sprachraum>,
    )
    await effekteNachholen()

    expect(gefragt).toHaveBeenCalledTimes(2)
    expect(screen.getByText(/probe\.mse-game/)).toBeInTheDocument()
  })
})
