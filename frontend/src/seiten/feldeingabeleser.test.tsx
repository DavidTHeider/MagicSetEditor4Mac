// **Ein Leser tippt nicht in ein Kartenfeld** — #2290
// (#2290).
//
// Die Felder des Karteneditors und der Setangaben schreiben beim Verlassen
// oder beim Wählen, ohne Knopf; für einen Leser standen sie offen, und erst
// die Bindung wies ab. Jetzt ist ein Textfeld **schreibgeschützt** — es
// bleibt fokussierbar, sein Text lässt sich lesen und kopieren —, und was
// sofort schreibt, Auswahl und Ankreuzfeld, ist gesperrt. Den Grund nennt
// der Streifen oben (`Nurlesehinweis`) — **und das Feld selbst** (H01,
// #2255):
// ein Vorleser hört am Feld »schreibgeschützt«, aber den Streifen nicht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO } from '../api'
import { suchschwelle } from '../bausteine/suchwahl'
import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Feldeingabe } from './Feldeingabe'

const satz = kataloge.de.bausteine.nurlesen.knopf

function form(teil: Partial<FeldbeschreibungDTO>): FeldbeschreibungDTO {
  return {
    name: 'name',
    aufschrift: 'Name',
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
    ...teil,
  }
}

function zeige(
  beschreibung: FeldbeschreibungDTO,
  wert: string,
  nurLesend = true,
) {
  const aendern = vi.fn()
  const verlassen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend={nurLesend}>
        <Feldeingabe
          feld="probe"
          wert={wert}
          form={beschreibung}
          angehalten={false}
          fehlerhaft={false}
          beschriebenVon="hinweis"
          aendern={aendern}
          verlassen={verlassen}
        />
      </Schreibrechtraum>
    </Sprachraum>,
  )

  return { aendern, verlassen }
}

describe('Ein Kartenfeld für einen Leser', () => {
  it('ist als Zeile schreibgeschützt und nimmt nichts an', async () => {
    const { aendern } = zeige(form({}), 'Feuerball')

    const feld = screen.getByRole('textbox', { name: 'probe' })
    expect(feld).toHaveAttribute('readonly')

    await userEvent.type(feld, 'x')
    expect(feld).toHaveValue('Feuerball')
    expect(aendern).not.toHaveBeenCalled()
  })

  it('ist als Kasten schreibgeschützt', () => {
    zeige(form({ mehrzeilig: true }), 'Fliegend')

    expect(screen.getByRole('textbox', { name: 'probe' })).toHaveAttribute(
      'readonly',
    )
  })

  it('sperrt eine Auswahl und ein Ankreuzfeld', () => {
    zeige(form({ art: 'choice', wahlen: ['rot', 'blau'] }), 'rot')
    expect(screen.getByRole('combobox', { name: 'probe' })).toBeDisabled()
  })

  it('sperrt ein Ankreuzfeld', () => {
    zeige(form({ art: 'boolean' }), 'yes')
    expect(screen.getByRole('checkbox', { name: 'probe' })).toBeDisabled()
  })

  it('sperrt eine Auswahl mit Suche', () => {
    const wahlen = Array.from(
      { length: suchschwelle },
      (_, i) => `W${String(i)}`,
    )
    zeige(form({ art: 'choice', wahlen }), 'W1')
    expect(screen.getByRole('combobox', { name: 'probe' })).toBeDisabled()
  })

  // **Und jedes sagt, warum** — derselbe Satz wie an einem gesperrten Knopf
  // (`useSchreibsperre`), als Beschreibung neben der eigenen.
  it.each([
    ['eine Zeile', form({}), 'Feuerball', 'textbox'],
    ['ein Kasten', form({ mehrzeilig: true }), 'Fliegend', 'textbox'],
    [
      'eine Auswahl',
      form({ art: 'choice', wahlen: ['rot', 'blau'] }),
      'rot',
      'combobox',
    ],
    ['ein Ankreuzfeld', form({ art: 'boolean' }), 'yes', 'checkbox'],
    [
      'eine Auswahl mit Suche',
      form({
        art: 'choice',
        wahlen: Array.from({ length: suchschwelle }, (_, i) => `W${String(i)}`),
      }),
      'W1',
      'combobox',
    ],
  ] as const)(
    '%s nennt den Grund als Beschreibung',
    (_, beschreibung, wert, rolle) => {
      zeige(beschreibung, wert)

      expect(
        screen.getByRole(rolle, { name: 'probe' }),
      ).toHaveAccessibleDescription(satz)
    },
  )

  // **Einem Mitarbeiter bleibt alles offen** — Beleg, dass der Raum zählt.
  it('lässt einem Mitarbeiter das Feld offen', () => {
    zeige(form({}), 'Feuerball', false)

    const feld = screen.getByRole('textbox', { name: 'probe' })
    expect(feld).not.toHaveAttribute('readonly')
    expect(feld).toHaveAccessibleDescription('')
  })
})
