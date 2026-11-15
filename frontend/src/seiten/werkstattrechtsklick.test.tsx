// **Rechtsklick und ⇧F10 an einer Vorlagendatei** — AP-5
// (#2362),
// Befund B-01: dieselben Handgriffe wie im Menü »Datei« über dem Editor, an
// der Datei im Baum. Eine andere Datei wird zuerst gewählt — mit der
// Rückfrage bei Ungesichertem —, und das Menü geht erst auf, wenn **sie**
// offen ist. Ein Rechtsklick auf ein Paket im Baum zeigt kein Dateimenü.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

const texte = kataloge.de.seiten.vorlagen
const pfad = '/pakete/probe.mse-style'

function zeige() {
  const lesen = vi.fn((_paket: string, datei: string) =>
    Promise.resolve({ inhalt: `# ${datei}\n`, stand: 'alt', fehler: null }),
  )
  belegeWerkstatt({
    Paketdateien: () =>
      Promise.resolve({
        dateien: ['style', 'card_fields'],
        beiwerk: ['rahmen.png'],
        fehler: null,
      }),
    PaketdateiLesen: lesen,
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  return { lesen }
}

describe('der Rechtsklick an einer Vorlagendatei', () => {
  afterEach(raeumeApi)

  it('wählt die Datei, wartet, bis sie offen ist, und zeigt dann ihre Handgriffe', async () => {
    const { lesen } = zeige()
    const datei = await screen.findByRole('button', { name: 'card_fields' })

    expect(fireEvent.contextMenu(datei, { clientX: 5, clientY: 5 })).toBe(false)
    await waitFor(() => {
      expect(lesen).toHaveBeenCalledWith(pfad, 'card_fields')
    })

    const menue = await screen.findByRole('menu', { name: texte.datei })
    for (const name of [
      texte.dateiNeu,
      texte.dateiHinzufuegen,
      texte.dateiUmbenennen,
      texte.dateiLoeschen,
    ]) {
      expect(screen.getByRole('menuitem', { name })).toBeInTheDocument()
    }

    // **Umbenennen führt in dasselbe Feld wie über das Menü »Datei«.**
    await userEvent.click(
      screen.getByRole('menuitem', { name: texte.dateiUmbenennen }),
    )
    expect(menue).not.toBeInTheDocument()
    expect(await screen.findByLabelText(texte.dateiNeuerName)).toHaveValue(
      'card_fields',
    )
  })

  it('zeigt an einem Paket im Baum kein Dateimenü', async () => {
    zeige()
    await userEvent.click(await screen.findByRole('button', { name: 'style' }))
    await screen.findByLabelText(texte.inhalt)

    const paket = screen.getByRole('button', { name: /probe\.mse-style/ })
    expect(fireEvent.contextMenu(paket)).toBe(true)

    expect(screen.queryByRole('menu')).toBeNull()
  })

  it('öffnet mit ⇧F10 an einer Datei ebenso ihre Handgriffe', async () => {
    zeige()
    const datei = await screen.findByRole('button', { name: 'card_fields' })

    fireEvent.keyDown(datei, { key: 'F10', shiftKey: true })

    expect(
      await screen.findByRole('menu', { name: texte.datei }),
    ).toBeInTheDocument()
  })
})
