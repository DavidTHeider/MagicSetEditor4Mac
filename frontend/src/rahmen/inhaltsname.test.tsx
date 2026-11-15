// **Der Inhalt trägt den Namen der Seite** — H01
// (#2255).
//
// Seit UX-10.AK-02 (#1682) setzen Sprunglink und Seitenwechsel den Fokus auf
// `<main>`. Der Kernweg je Rolle am nativen Fenster (`h01rollen`) fand ihn
// dort **ohne Namen**: ein Vorleser sagte nach ⌘2 nur »Hauptbereich« und
// nicht, auf welcher Seite man steht. Der Name ist der Titel der Seite —
// derselbe wie in der Leiste und der Brotkrume.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { beschriftung, type Seitenschluessel } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Rahmen } from './Rahmen'

afterEach(raeumeApi)

function zeige(aktuell: Seitenschluessel, sprache: 'de' | 'en' = 'de') {
  belegeApi({})
  render(
    <Sprachraum sprache={sprache}>
      <Rahmen aktuell={aktuell} wechseln={vi.fn()}>
        <p>{kataloge[sprache].seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )
}

describe('Der Inhalt', () => {
  it.each(['sets', 'karten', 'einstellungen', 'start'] as const)(
    'heißt auf »%s« wie die Seite',
    (seite) => {
      zeige(seite)

      expect(
        screen.getByRole('main', { name: beschriftung(kataloge.de, seite) }),
      ).toHaveAttribute('id', 'inhalt')
    },
  )

  it('folgt der Sprache des Fensters', () => {
    zeige('sets', 'en')

    expect(
      screen.getByRole('main', { name: beschriftung(kataloge.en, 'sets') }),
    ).toBeInTheDocument()
  })
})
