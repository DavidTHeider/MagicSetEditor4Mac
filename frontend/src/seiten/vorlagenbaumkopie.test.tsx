// **Die Pfeile laufen über die Knoten, nicht über die Kopieaktionen** — WK02
// (#2103).
//
// Gefunden an der WebKit-Abnahme `ux07-baum`: Pfeile, Pos1 und Ende liefen
// über **jeden** Knopf im Baum, auch über »Eigene Kopie anlegen« unter einem
// mitgelieferten Paket. Ende landete dort statt auf dem letzten Knoten, und
// von dort führte kein Pfeil weiter — die Kopieaktion hat keinen
// Tastenhorcher, sie ist eine eigene Tabstelle.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Vorlagenbaum } from './Vorlagenbaum'

const texte = kataloge.de.seiten.vorlagen

function zeige() {
  belegeApi({
    Paketdateien: (paket: string) =>
      Promise.resolve(
        paket.startsWith('/saat/')
          ? {
              dateien: [],
              beiwerk: [],
              fehler: {
                code: 'validierung',
                schluessel: 'validierung',
                meldung: 'liegt nicht unter der Wurzel',
              },
            }
          : { dateien: ['style'], beiwerk: [], fehler: null },
      ),
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagenbaum
        geladen
        fehlschlag={false}
        pakete={['/pakete/eigen.mse-style', '/saat/mit.mse-style']}
        gewaehlt={null}
        waehlen={vi.fn()}
        geaendert={false}
        kopie={{
          kopieren: vi.fn(() => Promise.resolve()),
          laeuft: () => false,
          fehler: () => undefined,
        }}
      />
    </Sprachraum>,
  )
}

describe('Vorlagenbaum mit Kopieaktion', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('Ende führt zum letzten Knoten, nicht zur Kopieaktion', async () => {
    zeige()

    const kopieren = await screen.findByRole('button', { name: texte.kopieren })
    const eigen = screen.getByRole('button', { name: 'eigen.mse-style' })

    eigen.focus()
    await userEvent.keyboard('{End}')

    expect(screen.getByRole('button', { name: 'mit.mse-style' })).toHaveFocus()
    expect(kopieren).not.toHaveFocus()
  })

  it('der Pfeil ab läuft über die Kopieaktion hinweg nicht in sie hinein', async () => {
    zeige()

    await screen.findByRole('button', { name: texte.kopieren })
    screen.getByRole('button', { name: 'style' }).focus()

    await userEvent.keyboard('{ArrowDown}{ArrowDown}')

    expect(screen.getByRole('button', { name: 'mit.mse-style' })).toHaveFocus()
  })

  it('die Kopieaktion bleibt eine eigene Tabstelle', async () => {
    zeige()

    const kopieren = await screen.findByRole('button', { name: texte.kopieren })

    expect(kopieren.tabIndex).toBe(0)
    expect(kopieren).not.toHaveAttribute('data-baumknoten')
  })
})
