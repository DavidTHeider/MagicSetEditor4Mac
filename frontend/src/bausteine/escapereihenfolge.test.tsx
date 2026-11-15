// **Escape schließt das Innerste zuerst** — G12 aus dem Zweitblick über
// Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// WCAG 1.4.13: was bei Zeiger oder Fokus erscheint, lässt sich ohne Bewegung
// wieder schließen. Ein Hinweis (`Tooltip`, die Feldhilfe) in einem Behälter
// war es nicht: `Entscheidung` und `Popover` hielten Escape an ihrer Gruppe
// an, bevor der Hinweis am Fenster es sah, und der `Dialog` schloss über
// Mantine schon in der Capture-Phase. Ein Druck schloss den ganzen Behälter
// samt Eingaben. Jetzt: das erste Escape nimmt der Hinweis, das zweite der
// Behälter.
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { Dialog } from './Dialog'
import { Entscheidung } from './Entscheidung'
import { Popover } from './Popover'
import { Tooltip } from './Tooltip'
import { Sprachraum } from '../sprache/kontext'
import { effekteNachholen } from '../test/effekte'

const hinweis = 'Was dieser Knopf tut.'

// Hinweisknopf ist ein Knopf mit einem Hinweis, der beim Fokus erscheint.
function Hinweisknopf() {
  return (
    <Tooltip text={hinweis} verzoegerung={0}>
      {(griffe) => (
        <button type="button" {...griffe}>
          Hilfe
        </button>
      )}
    </Tooltip>
  )
}

// hinweisZuerst fokussiert den Hinweisknopf, drückt Escape und gibt zurück,
// ob der Behälter danach noch offen ist; dann drückt es noch einmal.
//
// **Fokussiert ohne Zeiger** — seit #2068 räumt ein Druck auf den Auslöser
// den Hinweis, und ein Klick zeigte ihn hier deshalb nicht mehr. Gemeint war
// immer der Fokus.
async function hinweisZuerst(offen: () => boolean) {
  act(() => {
    screen.getByRole('button', { name: 'Hilfe' }).focus()
  })
  expect(await screen.findByText(hinweis)).toBeVisible()

  await userEvent.keyboard('{Escape}')
  await effekteNachholen()

  expect(screen.queryByText(hinweis)).not.toBeVisible()
  expect(offen()).toBe(true)

  await userEvent.keyboard('{Escape}')
  await effekteNachholen()

  expect(offen()).toBe(false)
}

describe('Escape mit einem Hinweis darin', () => {
  it('schließt im Popover erst den Hinweis', async () => {
    function Probe() {
      const [offen, setzeOffen] = useState(true)

      return (
        <Popover
          ausloeser={<span>Anker</span>}
          offen={offen}
          titel="Gruppe"
          schliessen={() => {
            setzeOffen(false)
          }}
        >
          <Hinweisknopf />
        </Popover>
      )
    }

    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    await hinweisZuerst(
      () => screen.queryByRole('group', { name: 'Gruppe' }) !== null,
    )
  })

  it('bricht eine Entscheidung erst mit dem zweiten Escape ab', async () => {
    const abbrechen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Entscheidung
          frage="Wirklich?"
          bestaetigen="Ja"
          abbrechen="Nein"
          aufBestaetigen={vi.fn()}
          aufAbbrechen={abbrechen}
        >
          <Hinweisknopf />
        </Entscheidung>
      </Sprachraum>,
    )

    await hinweisZuerst(() => abbrechen.mock.calls.length === 0)
  })

  it('schließt einen Dialog erst mit dem zweiten Escape', async () => {
    function Probe() {
      const [offen, setzeOffen] = useState(true)

      return offen ? (
        <Dialog
          bezeichnung="Kasten"
          schliessen={() => {
            setzeOffen(false)
          }}
        >
          <Hinweisknopf />
        </Dialog>
      ) : null
    }

    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    await hinweisZuerst(
      () => screen.queryByRole('dialog', { name: 'Kasten' }) !== null,
    )
  })
})
