// **Ein von außen geschriebener Wert steht im Feld** — gefunden an der
// Hüllenprobe zu AI-05 (#2042): die KI-Übernahme schrieb „Testkarte" in den
// Bestand, und das Feld zeigte weiter „Langtext". Die Eingabe steht auf
// `defaultValue` und nahm einen neuen gespeicherten Wert nicht an — dieselbe
// Klasse trifft jede Schreibart, die nicht getippt wurde.
import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Feldzeile } from './Feldzeile'

function zeile(wert: string) {
  return (
    <Sprachraum sprache="de">
      <Feldzeile
        feld="name"
        wert={wert}
        form={{
          name: 'name',
          aufschrift: 'Name',
          art: 'text',
          beschreibung: '',
          wahlen: [],
          mehrzeilig: false,
          vorgabe: '',
          teilfelder: [],
          reihe: '',
          stelle: 0,
          editierbar: true,
          gespeichert: true,
        }}
        fehler={undefined}
        laeuft={false}
        palette={undefined}
        wortlisten={[]}
        wortWaehlen={vi.fn()}
        hergeleitet={undefined}
        gewaehlt={vi.fn()}
        getippt={vi.fn()}
        schreiben={vi.fn()}
        einfuegen={vi.fn()}
        bildWaehlen={vi.fn()}
      />
    </Sprachraum>
  )
}

it('zeigt einen neu gespeicherten Wert', () => {
  const { rerender } = render(zeile('Langtext'))
  expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue('Langtext')

  rerender(zeile('Testkarte'))

  expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue('Testkarte')
})
