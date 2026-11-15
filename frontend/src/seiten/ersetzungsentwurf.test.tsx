// **Was getippt ist, ist ein Entwurf — auch während die Ersetzung läuft** —
// U01 und U02 (#2119),
// Zweitblick über C und D, Befund 3.
//
// `auto replace:` ist ein asynchroner Ruf, und die Eingabe meldete sich erst
// mit seiner Antwort als Entwurf. Kam dazwischen ein fremder Wert für dasselbe
// Feld, sah der Entwurfshaken keinen Entwurf und baute die Eingabe neu auf:
// das Getippte und der Fokus waren fort, bevor ein Konflikt greifen konnte.
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { effekteNachholen } from '../test/effekte'
import { Feldzeile } from './Feldzeile'

afterEach(cleanup)

interface Ersetzung {
  text: string
  stelle: number
}

describe('ein Entwurf während der Ersetzung', () => {
  it.each([
    ['beantwortet, bevor der fremde Wert kommt', true],
    ['noch unterwegs, wenn der fremde Wert kommt', false],
  ])('bleibt mit Text und Fokus stehen — Ersetzung %s', async (_, vorher) => {
    let antworten: (ersetzt: Ersetzung) => void = () => undefined
    const autoErsetzen = vi.fn(
      () =>
        new Promise<Ersetzung>((erfuellen) => {
          antworten = erfuellen
        }),
    )

    const zeile = (wert: string) => (
      <Sprachraum sprache="de">
        <Feldzeile
          feld="name"
          wert={wert}
          form={undefined}
          fehler={undefined}
          laeuft={false}
          palette={undefined}
          wortlisten={[]}
          hergeleitet={undefined}
          gewaehlt={vi.fn()}
          getippt={vi.fn()}
          schreiben={vi.fn()}
          einfuegen={vi.fn()}
          bildWaehlen={vi.fn()}
          wortWaehlen={vi.fn()}
          autoErsetzen={autoErsetzen}
        />
      </Sprachraum>
    )

    const { rerender } = render(zeile('Saat'))
    const eingabe = screen.getByRole<HTMLInputElement>('textbox', {
      name: 'name',
    })
    eingabe.focus()

    fireEvent.input(eingabe, {
      target: { value: 'Saat X', selectionStart: 6 },
      inputType: 'insertText',
    })
    expect(autoErsetzen).toHaveBeenCalledTimes(1)

    if (vorher) {
      act(() => {
        antworten({ text: 'Saat X', stelle: 6 })
      })
      await effekteNachholen()
    }

    rerender(zeile('Fremd'))

    if (!vorher) {
      act(() => {
        antworten({ text: 'Saat X', stelle: 6 })
      })
      await effekteNachholen()
    }

    const danach = screen.getByRole<HTMLInputElement>('textbox', {
      name: 'name',
    })
    expect(danach).toBe(eingabe)
    expect(danach.value).toBe('Saat X')
    expect(document.activeElement).toBe(danach)
  })
})
