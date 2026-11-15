// **Die Paketliste lädt nach, wenn ein anderer Prozess schreibt** — #2011,
// Folge aus UXR-17 (#1925),
// Befund 3 der Hüllenprobe `fmt02e`.
//
// Ein Paket, das die Kommandozeile, MCP oder der Dienst bezieht oder
// herausnimmt, stand bis dahin weiter in der offenen Liste — erst ein Wechsel
// der Seite lud neu. Dasselbe Zeichen, das die übrigen Seiten bekommen
// (`useBestandsstand`, #156), gilt jetzt auch hier.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { belegeSeite } from '../test/paketseite'
import { paketDTO } from '../test/vorlagenprobe'
import { Pakete } from './Pakete'

afterEach(() => {
  raeumeApi()
})

function seite(bestandsstand: number) {
  return (
    <Sprachraum sprache="de">
      <Pakete wechseln={vi.fn()} bestandsstand={bestandsstand} />
    </Sprachraum>
  )
}

it('lädt die Liste nach einer fremden Änderung neu', async () => {
  let namen = ['erstes']
  belegeSeite({
    Pakete: () => Promise.resolve({ pakete: namen.map(paketDTO), befunde: [] }),
  })

  const { rerender } = render(seite(0))
  expect(await screen.findByText('erstes')).toBeInTheDocument()

  // Ein anderer Prozess legt ein zweites dazu — und der Stand steigt.
  namen = ['erstes', 'zweites']
  rerender(seite(1))

  expect(await screen.findByText('zweites')).toBeInTheDocument()
})
