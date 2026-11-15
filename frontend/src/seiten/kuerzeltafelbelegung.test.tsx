// **Die Belegung von ⌘0 bis ⌘9 ändert sich an der Kürzeltafel** —
// UX-02.AK-03 (#1674).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Kuerzeltafel } from './Kuerzeltafel'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { alsWert, belegungAus, belegungVor06 } from '../rahmen/ziffernbelegung'

const eigene = kataloge.de.seiten.einstellungen

function zeige(wert = '', setze = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Kuerzeltafel belegung={belegungAus(wert)} setze={setze} />
    </Sprachraum>,
  )

  return setze
}

describe('die Belegung an der Kürzeltafel', () => {
  // **Die Tafel liest die Belegung**, nicht `navigation.ts`: wer ⌘3 auf die
  // Karten gelegt hat, sieht es dort.
  it('zeigt in der Tafel, was die Belegung sagt', () => {
    zeige('3=karten')

    const tafel = screen.getByRole('table')
    const zeile = within(tafel).getByText('⌘3').closest('tr')
    expect(zeile).toHaveTextContent('Karten')
  })

  it('legt eine Ziffer über ihre Auswahl auf eine Seite', async () => {
    const setze = zeige()

    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: 'Seite für ⌘3' }),
      'karten',
    )

    expect(setze).toHaveBeenCalledWith('3=karten')
  })

  it('räumt eine Ziffer über „—"', async () => {
    const setze = zeige()

    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: 'Seite für ⌘1' }),
      '',
    )

    expect(setze).toHaveBeenCalledWith('1=')
  })

  it('holt die Belegung von vor 0.6.0 und die Vorgabe zurück', async () => {
    const setze = zeige('3=karten')

    await userEvent.click(
      screen.getByRole('button', { name: eigene.belegungVor06 }),
    )
    expect(setze).toHaveBeenLastCalledWith(alsWert(belegungVor06))

    await userEvent.click(
      screen.getByRole('button', { name: eigene.belegungVorgabe }),
    )
    expect(setze).toHaveBeenLastCalledWith('')
  })

  // **Die Suche findet sie wie jede Einstellung** (UX-04.AK-03): sie springt
  // über `data-einstellung` auf das Feld.
  it('trägt ihren Schlüssel für die Suche', () => {
    zeige()

    expect(
      screen
        .getByRole('group', { name: eigene.belegungUeberschrift })
        .getAttribute('data-einstellung'),
    ).toBe('kuerzel.ziffern')
  })

  // **Ohne `setze` bleibt es eine Anzeige** — wie vor UX-02.AK-03.
  it('zeigt ohne Schreibweg keine Auswahl', () => {
    render(
      <Sprachraum sprache="de">
        <Kuerzeltafel />
      </Sprachraum>,
    )

    expect(screen.queryByRole('combobox')).toBeNull()
  })
})
