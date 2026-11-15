// **Rechtsklick und ⇧F10 an einer Setkarte** — AP-5
// (#2362),
// Befund B-01: dieselben Handlungen wie im Menü der Karte, an der Karte
// selbst. Löschen behält seine Rückfrage, und ein Leser sieht es gesperrt.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { stillerUmschlaglauf } from '../test/umschlaglauf'
import { Setkarte } from './Setkarte'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

const einSet: SetDTO = {
  id: '7',
  name: 'Mein Set',
  spiel: 'magic',
  stil: 'm15',
  karten: 3,
  herkunft: '/tmp/mein.mse-set',
  herkunftFehlt: false,
}

function zeige(nurLesend = false) {
  const loeschen = vi.fn(() => Promise.resolve())
  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend={nurLesend}>
        <Setkarte
          satz={einSet}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
          loeschung={{
            loeschen,
            laeuft: () => false,
            befunde: () => [],
            fehler: () => undefined,
          }}
          umschlag={stillerUmschlaglauf()}
        />
      </Schreibrechtraum>
    </Sprachraum>,
  )

  return { loeschen }
}

const menue = () =>
  screen.getByRole('menu', {
    name: einsetzen(worte.handlungenAn, { name: 'Mein Set' }),
  })

describe('der Rechtsklick an einer Setkarte', () => {
  it('zeigt ihre Handlungen, und Löschen fragt zuerst', async () => {
    const { loeschen } = zeige()

    const ereignis = fireEvent.contextMenu(screen.getByText(einSet.herkunft), {
      clientX: 5,
      clientY: 5,
    })

    expect(ereignis).toBe(false)
    for (const name of [
      worte.speichern,
      worte.speichernUnter,
      worte.umschlagAblegen,
      worte.vorlageAusfuhr,
      worte.loeschen,
    ]) {
      expect(screen.getByRole('menuitem', { name })).toBeInTheDocument()
    }

    await userEvent.click(
      screen.getByRole('menuitem', { name: worte.loeschen }),
    )

    expect(loeschen).not.toHaveBeenCalled()
    expect(
      screen.getByRole('button', { name: worte.loeschenJa }),
    ).toBeInTheDocument()
  })

  it('geht auch mit ⇧F10 am Namen auf', () => {
    zeige()

    fireEvent.keyDown(
      screen.getByRole('button', {
        name: `${einSet.name} — ${worte.oeffnen}`,
      }),
      { key: 'F10', shiftKey: true },
    )

    expect(menue()).toBeInTheDocument()
  })

  it('zeigt einem Leser Löschen gesperrt, mit Grund', () => {
    zeige(true)

    fireEvent.contextMenu(screen.getByText(einSet.herkunft))

    const eintrag = screen.getByRole('menuitem', { name: worte.loeschen })
    expect(eintrag).toHaveAttribute('aria-disabled', 'true')
    expect(eintrag).toHaveAccessibleDescription()
  })
})
