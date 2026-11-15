// **Die Belegung der Ziffern steht als Tafel und nicht als Textfeld** —
// UX-02.AK-03 (#1674).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { eintraege, zeigeEinstellungen } from '../test/einstellungsseite'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

const ziffern: EinstellungDTO = {
  schluessel: 'kuerzel.ziffern',
  wert: '3=karten',
  wirksam: '3=karten',
  werte: [],
  bereich: 'anwendung',
  ebene: 'nutzer',
  herkunft: 'nutzer',
  vorgabe: '',
  wirkung: 'sofort',
}

describe('die Belegung auf der Einstellungsseite', () => {
  it('zeigt sie an der Kürzeltafel und schreibt, was gewählt wird', async () => {
    const setze = zeigeEinstellungen({
      alleEinstellungen: [...eintraege, ziffern],
    })

    // **Kein zweites Bedienelement**: ein Textfeld mit „3=karten" darin
    // stünde neben der Tafel.
    expect(
      screen.queryByRole('textbox', { name: eigene.namen['kuerzel.ziffern'] }),
    ).toBeNull()

    // **Seit UXR-19 hinter dem Aufklapper „Tastenkürzel"** (#1927): zehn
    // Wahlen, die man einmal stellt, standen vor jeder anderen Einstellung.
    await userEvent.click(
      screen.getByRole('button', { name: eigene.kuerzelUeberschrift }),
    )

    const auswahl = screen.getByRole('combobox', { name: 'Seite für ⌘3' })
    expect(auswahl).toHaveValue('karten')

    await userEvent.selectOptions(auswahl, 'sammlung')

    expect(setze).toHaveBeenCalledWith('kuerzel.ziffern', '')
  })
})
