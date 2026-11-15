// **Feld und Set der Suche wachsen mit dem Bestand** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Auswahl ohne Suche, wächst mit dem Bestand
// über 15"*.
//
// Wer viele Sets hat oder ein Spiel mit vielen Feldern, wählt nicht mehr aus
// einem Menü, sondern tippt: ab fünfzehn Werten ist die Wahl eine Auswahl mit
// Suche (`Auswahlfeld`), darunter bleibt sie das native Feld. Die Verträge
// der Seite bleiben: ein Feld, das nur das vorige Set führt, wird beim
// Wechsel geräumt (§7, Regel 1).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, belegeFenster, raeumeApi } from '../test/api'
import { Suche } from './Suche'

const texte = kataloge.de.seiten.suche

const sets = (anzahl: number) =>
  Array.from({ length: anzahl }, (_, stelle) => ({
    id: String(stelle + 1),
    name: `Set ${String(stelle + 1).padStart(2, '0')}`,
    spiel: 'magic',
    stil: 'm15',
    herkunft: '',
    karten: 1,
    herkunftFehlt: false,
  }))

const felder = (anzahl: number) =>
  Array.from(
    { length: anzahl },
    (_, stelle) => `feld ${String(stelle + 1).padStart(2, '0')}`,
  )

function zeige(anzahlSets: number, anzahlFelder: number) {
  const Suchen = vi.fn(() => Promise.resolve({ treffer: [], gekuerzt: false }))
  const Suchfelder = vi.fn(() =>
    Promise.resolve({ felder: felder(anzahlFelder) }),
  )

  belegeApi({
    ...belegeFenster('de'),
    Sets: () => Promise.resolve({ sets: sets(anzahlSets) }),
    Suchfelder,
    Suchen,
  })

  render(
    <Sprachraum sprache="de">
      <Suche oeffnen={vi.fn()} />
    </Sprachraum>,
  )

  return { Suchen, Suchfelder }
}

const feldwahl = () => screen.findByRole('combobox', { name: texte.feld })
const setwahl = () => screen.findByRole('combobox', { name: texte.set })

afterEach(() => {
  raeumeApi()
})

describe('Die Filter der Suche', () => {
  it('bleiben unter fünfzehn Werten native Felder', async () => {
    zeige(3, 4)

    expect((await feldwahl()).tagName).toBe('SELECT')
    expect((await setwahl()).tagName).toBe('SELECT')
  })

  it('sind ab fünfzehn Werten Auswahlen mit Suche', async () => {
    zeige(16, 20)

    await waitFor(async () => {
      expect((await feldwahl()).tagName).toBe('INPUT')
    })
    expect((await setwahl()).tagName).toBe('INPUT')
    expect(await feldwahl()).toHaveValue(texte.alleFelder)
    expect(await setwahl()).toHaveValue(texte.alleSets)
  })

  it('wählen ein Feld über die Suche', async () => {
    const { Suchen } = zeige(2, 20)

    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.eingabe }),
      'feuer',
    )
    await waitFor(async () => {
      expect((await feldwahl()).tagName).toBe('INPUT')
    })
    await userEvent.click(await feldwahl())
    await userEvent.keyboard('feld 17{Enter}')

    await waitFor(() => {
      expect(Suchen).toHaveBeenLastCalledWith('feuer', 'feld 17', '')
    })
  })

  // **Das Feld wird mitgeräumt** (§7, Regel 1): und die Felder werden für das
  // neue Set geholt.
  it('wählen ein Set über die Suche, räumen das Feld und holen dessen Felder', async () => {
    const { Suchen, Suchfelder } = zeige(16, 20)

    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.eingabe }),
      'feuer',
    )
    await waitFor(async () => {
      expect((await feldwahl()).tagName).toBe('INPUT')
    })
    await userEvent.click(await feldwahl())
    await userEvent.keyboard('feld 03{Enter}')

    await waitFor(() => {
      expect(Suchen).toHaveBeenLastCalledWith('feuer', 'feld 03', '')
    })

    await userEvent.click(await setwahl())
    await userEvent.keyboard('set 16{Enter}')

    await waitFor(() => {
      expect(Suchen).toHaveBeenLastCalledWith('feuer', '', '16')
    })
    expect(Suchfelder).toHaveBeenLastCalledWith('16')
    expect(await feldwahl()).toHaveValue(texte.alleFelder)
  })
})
