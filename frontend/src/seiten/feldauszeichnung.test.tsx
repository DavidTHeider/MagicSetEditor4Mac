// **Das Eingabefeld zeigt den Klartext und schreibt den ausgezeichneten Wert**
// (#687, Stufe 2).
//
// Gesehen an der gebauten Anwendung: das Feld **Type** einer M15-Karte zeigte
//
//	<word-list-type-en></word-list-type-en><sep-soft> — </sep-soft>
//
// als das Erste, was jemand an diesem Feld sieht. Wer die Marken löscht,
// verliert die Wortliste, und er sieht nirgends, warum.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO } from '../api'
import { Feldeingabe } from './Feldeingabe'

function form(teile: Partial<FeldbeschreibungDTO> = {}): FeldbeschreibungDTO {
  return {
    name: 'rule text',
    aufschrift: 'Rule text',
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    ...teile,
  }
}

function zeige(
  wert: string,
  teile: Partial<FeldbeschreibungDTO> = {},
  hergeleitet?: string,
) {
  const aendern = vi.fn()
  const verlassen = vi.fn()

  render(
    <Feldeingabe
      feld="rule text"
      wert={wert}
      form={form(teile)}
      angehalten={false}
      fehlerhaft={false}
      beschriebenVon={undefined}
      hergeleitet={hergeleitet}
      aendern={aendern}
      verlassen={verlassen}
    />,
  )

  return { aendern, verlassen }
}

describe('Ein Feld mit Auszeichnungen', () => {
  it('zeigt den Klartext statt der Marken', () => {
    zeige('<i-flavor>Es brennt.</i-flavor>')

    expect(screen.getByRole('textbox')).toHaveValue('Es brennt.')
  })

  it('zeigt auch in einem mehrzeiligen Kasten den Klartext', () => {
    zeige('<i-flavor>Es brennt.</i-flavor>', { mehrzeilig: true })

    expect(screen.getByRole('textbox')).toHaveValue('Es brennt.')
  })

  it('schreibt die Marken wieder mit', async () => {
    const nutzer = userEvent.setup()
    const { aendern, verlassen } = zeige('<i-flavor>Es</i-flavor>')

    const kasten = screen.getByRole('textbox')
    await nutzer.type(kasten, '!')
    await nutzer.tab()

    expect(aendern).toHaveBeenLastCalledWith('<i-flavor>Es!</i-flavor>')
    expect(verlassen).toHaveBeenLastCalledWith('<i-flavor>Es!</i-flavor>')
  })

  it('behält die Marken, wenn jemand das Feld leert', async () => {
    const nutzer = userEvent.setup()
    const { verlassen } = zeige('<i-flavor>Es</i-flavor>')

    await nutzer.clear(screen.getByRole('textbox'))
    await nutzer.tab()

    expect(verlassen).toHaveBeenLastCalledWith('<i-flavor></i-flavor>')
  })

  it('lässt ein Feld ohne Auszeichnung unangetastet', async () => {
    const nutzer = userEvent.setup()
    const { verlassen } = zeige('Blitz')

    await nutzer.type(screen.getByRole('textbox'), 'schlag')
    await nutzer.tab()

    expect(verlassen).toHaveBeenLastCalledWith('Blitzschlag')
  })

  // **Ein hergeleiteter Wert steht als Platzhalter da, nicht als Wert** —
  // und auch er ohne seine Marken.
  //
  // Das war der gemeldete Fall: `type` einer frischen Karte trägt **keinen**
  // Wert, und der gerechnete `<word-list-type-en>…</word-list-type-en>` stand
  // roh im Feld. Wer ihn als Wert einsetzte, machte aus einer Herleitung eine
  // Eingabe — beim nächsten Speichern stünde sie in der Datei.
  it('zeigt eine Herleitung als Platzhalter ohne Marken', () => {
    zeige('', {}, '<word-list-type-en>Creature</word-list-type-en>')

    const kasten = screen.getByRole('textbox')

    expect(kasten).toHaveValue('')
    expect(kasten).toHaveAttribute('placeholder', 'Creature')
  })

  it('schreibt in ein leeres Feld nur, was jemand tippt', async () => {
    const nutzer = userEvent.setup()
    const { verlassen } = zeige(
      '',
      {},
      '<word-list-type-en>Creature</word-list-type-en>',
    )

    await nutzer.type(screen.getByRole('textbox'), 'Goblin')
    await nutzer.tab()

    expect(verlassen).toHaveBeenLastCalledWith('Goblin')
  })

  // **Eine Auskunft ist auch Text** — `info` ist bei MSE nur eine Aufschrift
  // (`src/data/field/information.hpp`), und die rohen Marken darin wären
  // dasselbe Rätsel wie überall sonst.
  it('zeigt auch eine Auskunft ohne Marken', () => {
    zeige('<i>Hinweis</i>', { art: 'info' })

    expect(screen.getByRole('textbox')).toHaveValue('Hinweis')
  })
})
