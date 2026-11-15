// **Der Kartentext wird geprüft, aber nicht ersetzt** — G03
// (#2250).
//
// MSE2 unterringelt, was nicht im Wörterbuch steht, und ersetzt nie von
// selbst (`src/script/functions/spelling.cpp`, Fassung f9d9356 vom
// 2021-07-06: `check_spelling` setzt nur `<error-spelling>`-Marken). Ein
// `textarea` in WebKit korrigiert dagegen nach der Systemeinstellung — und die
// ist auf macOS ab Werk an. Ein Kartenname oder ein Regelwort, das das System
// nicht kennt, würde beim Tippen still „berichtigt".
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Feldeingabe } from './Feldeingabe'

import type { FeldbeschreibungDTO } from '../api'

function zeige(mehrzeilig: boolean) {
  const form: FeldbeschreibungDTO = {
    name: 'rule text',
    aufschrift: 'Rule Text',
    art: 'text',
    mehrzeilig,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }

  render(
    <Sprachraum sprache="de">
      <Feldeingabe
        feld="rule text"
        wert="Flickerwisp"
        form={form}
        angehalten={false}
        fehlerhaft={false}
        beschriebenVon="hinweis"
        hergeleitet={undefined}
        aendern={vi.fn()}
        verlassen={vi.fn()}
      />
    </Sprachraum>,
  )

  return screen.getByRole('textbox', { name: 'rule text' })
}

describe('Der Kartentext', () => {
  it.each([
    ['mehrzeilig', true],
    ['einzeilig', false],
  ])(
    'wird %s nicht von selbst berichtigt und bleibt geprüft',
    (_name, mehrzeilig) => {
      const feld = zeige(mehrzeilig)

      expect(feld).toHaveAttribute('autocorrect', 'off')
      expect(feld).toHaveAttribute('autocapitalize', 'off')

      // **Die Unterringelung bleibt** — sie ist MSE2s Hilfe, und sie ändert
      // nichts am Text.
      expect(feld).not.toHaveAttribute('spellcheck', 'false')
    },
  )
})
