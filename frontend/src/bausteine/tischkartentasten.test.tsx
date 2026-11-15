// **Die Einzeltasten am gezeichneten Tisch** — P4
// (#2372),
// Befund C-09: der Knopf über einer Karte nennt seine Tasten in
// `aria-keyshortcuts` und reicht einen Druck an die Bedienung weiter. Was sie
// annimmt, geht nicht weiter an die Fläche — sonst rollte ⌫ den Tisch oder
// die Seite. **⌘, ⌥ und ⌃ gehören dem Fenster.**
import { fireEvent, render, screen } from '@testing-library/react'
import type { KeyboardEvent } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { tischplan } from '../tisch/tischplan'
import type { Tischmalerei } from '../tisch/maler'
import { einzeltaste } from './Tischkartenknopf'
import { Tischflaeche } from './Tischflaeche'

const plan = tischplan(
  einePartie({
    objekte: [
      einObjekt({
        kennung: 'a',
        zone: 'tisch',
        name: 'Wald',
        eigner: 'sitz1',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ],
  }),
)

const ohneGrafik = () =>
  Promise.resolve<Tischmalerei>({ art: 'ohne', grund: 'kein-webgl' })

function zeige(taste?: (kennung: string, wort: string) => boolean) {
  render(
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={plan}
        bild={() => undefined}
        malerei={ohneGrafik}
        {...(taste === undefined
          ? {}
          : { taste, tastenFuer: () => 'T Backspace E' })}
      />
    </Sprachraum>,
  )

  return screen.findByRole('button', { name: /Wald/ })
}

describe('die Einzeltasten am gezeichneten Tisch', () => {
  it('reicht eine angenommene Taste weiter und hält sie an', async () => {
    const taste = vi.fn(() => true)
    const knopf = await zeige(taste)

    expect(knopf).toHaveAttribute('aria-keyshortcuts', 'T Backspace E')
    // **fireEvent gibt `false`, wenn das Ereignis angehalten wurde.**
    expect(fireEvent.keyDown(knopf, { key: 't' })).toBe(false)
    expect(taste).toHaveBeenCalledWith('a', 't')
  })

  it('lässt eine Taste durch, die die Bedienung nicht annimmt', async () => {
    const taste = vi.fn(() => false)
    const knopf = await zeige(taste)

    expect(fireEvent.keyDown(knopf, { key: 'x' })).toBe(true)
    expect(taste).toHaveBeenCalledWith('a', 'x')
  })

  it('lässt ⌘T, ⌥E und ⌃⌫ dem Fenster', async () => {
    const taste = vi.fn(() => true)
    const knopf = await zeige(taste)

    fireEvent.keyDown(knopf, { key: 't', metaKey: true })
    fireEvent.keyDown(knopf, { key: 'e', altKey: true })
    fireEvent.keyDown(knopf, { key: 'Backspace', ctrlKey: true })

    expect(taste).not.toHaveBeenCalled()
  })

  it('nennt ohne Bedienung keine Tasten und hält nichts an', async () => {
    const knopf = await zeige()

    expect(knopf).not.toHaveAttribute('aria-keyshortcuts')
    expect(fireEvent.keyDown(knopf, { key: 't' })).toBe(true)
  })

  it('nimmt ohne Empfänger nichts an', () => {
    const ereignis = {
      key: 't',
      metaKey: false,
      ctrlKey: false,
      altKey: false,
    } as KeyboardEvent<HTMLElement>

    expect(einzeltaste(ereignis, undefined)).toBe(false)
  })
})
