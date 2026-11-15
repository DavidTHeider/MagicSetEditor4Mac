// **Keine Formularwand aus Stiloptionen** — UXR-07
// (#1915).
//
// R3 aus `docs/UIREVIEW.md`: höchstens zwölf gezeichnete Felder, sonst
// gruppiert und eingeklappt bis auf eine Gruppe. **82 der 376 Stile** des
// Full-Magic-Pack führen mehr als zwölf Optionen, `magic-m15-dungeons` 66
// (gezählt am 2026-10-02 über `styling field:`). Das Artboard
// *Eigenschaften* zeigt „4 von 11“ und legt den Rest unter *Mehr* —
// „eingeklappt, nicht entfernt“.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { StiloptionDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { Stiloptionen } from './Stiloptionen'

const worte = kataloge.de.seiten.karten

afterEach(() => {
  raeumeApi()
})

// schalter ist eine Option, die als Kästchen erscheint.
const schalter = (name: string): StiloptionDTO => ({
  name,
  art: 'boolean',
  beschreibung: '',
  wert: 'no',
  moeglichkeiten: [],
  eigeneFarbe: false,
  gesetzt: false,
  hatVorgabe: false,
  hatWahlbilder: false,
  vorgabe: '',
})

function zeige(optionen: StiloptionDTO[], StiloptionSetzen = vi.fn()) {
  belegeApi({
    Stiloptionen: () => Promise.resolve({ optionen }),
    StiloptionSetzen,
  })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  return StiloptionSetzen
}

const sieben = ['eins', 'zwei', 'drei', 'vier', 'fünf', 'sechs', 'sieben']

describe('Weitere Stiloptionen', () => {
  it('zeichnet vier Optionen und legt den Rest hinter einen Knopf', async () => {
    zeige(sieben.map(schalter))

    expect(await screen.findByLabelText('vier')).toBeInTheDocument()
    expect(screen.getAllByRole('checkbox')).toHaveLength(4)
    expect(screen.queryByLabelText('fünf')).not.toBeInTheDocument()

    // **Neben der Überschrift steht, wie viele es sind** — wie im Artboard.
    expect(
      screen.getByRole('heading', { name: worte.stiloptionen }),
    ).toBeInTheDocument()
    expect(screen.getByText('4 von 7')).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Weitere Optionen (3)' }),
    ).toHaveAttribute('aria-expanded', 'false')
  })

  // **Eingeklappt, nicht entfernt**: jede Option bleibt erreichbar, und sie
  // schreibt wie die vorn.
  it('zeigt und schreibt die weiteren nach dem Aufklappen', async () => {
    const StiloptionSetzen = zeige(
      sieben.map(schalter),
      vi.fn(() => Promise.resolve({})),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: 'Weitere Optionen (3)' }),
    )
    await userEvent.click(screen.getByLabelText('sieben'))

    expect(screen.getAllByRole('checkbox')).toHaveLength(7)
    expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'sieben', 'yes')
  })

  it('faltet nichts, solange es höchstens vier sind', async () => {
    zeige(sieben.slice(0, 4).map(schalter))

    expect(await screen.findByLabelText('vier')).toBeInTheDocument()
    expect(
      screen.getByRole('heading', { name: worte.stiloptionen }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })

  // **WCAG 2.2 AA**, zu und offen — der Knopf sagt über `aria-expanded`, ob
  // dahinter etwas steht.
  it('hält die Regeln der Barrierefreiheit', async () => {
    zeige(sieben.map(schalter))
    const knopf = await screen.findByRole('button', {
      name: 'Weitere Optionen (3)',
    })

    expect(await verstoesse(document.body)).toEqual([])

    await userEvent.click(knopf)
    await screen.findByLabelText('sieben')

    expect(await verstoesse(document.body)).toEqual([])
  })
})
