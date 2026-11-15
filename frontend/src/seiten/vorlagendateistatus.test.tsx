// **Der Stand unter dem Editor sagt, was ist** — gefunden an der Hüllenprobe
// `fmt02c` (FMT-02, #1691):
// unter einer frisch geöffneten, unveränderten Datei stand „Nicht gesichert.",
// als gäbe es etwas zu sichern.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

const texte = kataloge.de.seiten.vorlagen

describe('Der Stand unter dem Editor', () => {
  afterEach(raeumeApi)

  it('schweigt bei einer unveränderten Datei und nennt eine Änderung', async () => {
    belegeWerkstatt()
    render(
      <Sprachraum sprache="de">
        <Vorlagen />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByRole('button', { name: 'style' }))
    const feld = await screen.findByLabelText(texte.inhalt)

    expect(screen.queryByText(texte.ungespeichert)).not.toBeInTheDocument()

    await userEvent.type(feld, 'x')

    expect(screen.getByText(texte.ungespeichert)).toBeInTheDocument()
  })
})
