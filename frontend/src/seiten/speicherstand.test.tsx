// **Ob das Eingetippte im Bestand steht** — UX-05.AK-03
// (#1677).
//
// Geschrieben wird beim Verlassen eines Feldes. **Gemessen wird hier der
// Weg über den Karteneditor**: ein Feld verlassen, und der Satz darüber sagt,
// wie es ausging — und dass ein zweiter, geglückter Versuch den Fehler räumt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Karteneditor } from './Karteneditor'
import { Speicherstand } from '../bausteine/Speicherstand'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  lageAus,
  schreibenBeginnt,
  schreibenEndet,
  vergessen,
} from '../speicherstand'

import type { SchreibAntwort } from '../api'

const eigene = kataloge.de.bausteine.speicherstand

afterEach(() => {
  raeumeApi()
  vergessen()
})

// zeige stellt Stand und Editor nebeneinander, wie die Arbeitsfläche.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Speicherstand setID="1" />
      <Karteneditor
        setID="1"
        karte={{
          id: '5',
          nummer: 0,
          werte: [{ feld: 'name', wert: 'Feuerball' }],
        }}
        formen={[]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Der Speicherstand am Kartenformular', () => {
  it('sagt nichts, solange nichts geschrieben wurde', () => {
    belegeApi({})
    zeige()

    expect(screen.queryByText(eigene.geschrieben)).toBeNull()
  })

  it('sagt nach dem Verlassen eines Feldes, dass es im Bestand steht', async () => {
    belegeApi({
      FeldSetzen: () => Promise.resolve({} satisfies SchreibAntwort),
    })
    zeige()

    await userEvent.type(
      await screen.findByRole('textbox', { name: 'name' }),
      'X',
    )
    await userEvent.tab()

    expect(await screen.findByText(eigene.geschrieben)).toBeInTheDocument()
  })

  // **Solange die Antwort aussteht, sagt er es** — und danach, wie es ausging.
  it('sagt, dass gespeichert wird, solange die Antwort aussteht', async () => {
    let antworten: (antwort: SchreibAntwort) => void = () => undefined
    belegeApi({
      FeldSetzen: () =>
        new Promise<SchreibAntwort>((fertig) => {
          antworten = fertig
        }),
    })
    zeige()

    await userEvent.type(
      await screen.findByRole('textbox', { name: 'name' }),
      'X',
    )
    await userEvent.tab()

    expect(await screen.findByText(eigene.schreibt)).toBeInTheDocument()

    antworten({})

    expect(await screen.findByText(eigene.geschrieben)).toBeInTheDocument()
  })

  // **Ein Fehlschlag nennt sein Feld — und ein zweiter, geglückter Versuch
  // räumt ihn.**
  it('nennt ein nicht geschriebenes Feld, bis es geschrieben ist', async () => {
    const FeldSetzen = vi
      .fn<() => Promise<SchreibAntwort>>()
      .mockResolvedValueOnce({
        fehler: { code: 'belegt', schluessel: '', meldung: 'Gesperrt.' },
      })
      .mockResolvedValue({})

    belegeApi({ FeldSetzen })
    zeige()

    const feld = await screen.findByRole('textbox', { name: 'name' })

    await userEvent.type(feld, 'X')
    await userEvent.tab()

    expect(
      await screen.findByText(eigene.fehler.replace('{feld}', 'name')),
    ).toBeInTheDocument()

    await userEvent.type(feld, 'Y')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByText(eigene.geschrieben)).toBeInTheDocument()
    })
  })
})

describe('Der Speicher des Standes', () => {
  // **Während geschrieben wird, sagt er es** — und ein Fehler in einem
  // anderen Feld bleibt stehen, wenn dieses glückt.
  it('führt Schreiben, Fehler und Erfolg je Karte und Feld', () => {
    schreibenBeginnt('1')
    expect(lageAus('1')).toEqual({ art: 'schreibt' })

    schreibenEndet('1', '5', 'name', false)
    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'typ', true)

    expect(lageAus('1')).toEqual({ art: 'fehler', feld: 'name' })
  })

  // **Jeder Fehler bleibt, bis sein Feld geschrieben ist** — angestrichen
  // vom Zweitblick über Phase 6.3 (M8): es gab einen Fehlerplatz je Set, und
  // ein geglücktes Feld räumte den Fehler eines anderen. „Jede Änderung
  // steht im Bestand." stand dann da, obwohl ein Wert fehlte.
  it('räumt einen Fehler nur mit seinem eigenen Feld auf seiner Karte', () => {
    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'name', false)
    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'typ', false)
    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'typ', true)

    expect(lageAus('1')).toEqual({ art: 'fehler', feld: 'name' })

    schreibenBeginnt('1')
    schreibenEndet('1', '6', 'name', true)
    expect(lageAus('1')).toEqual({ art: 'fehler', feld: 'name' })

    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'name', true)
    expect(lageAus('1')).toEqual({ art: 'geschrieben' })
  })

  // **Jedes Set zählt für sich** — ein Schreiben, das in A endet, nachdem in
  // B eines begann, ersetzte B's Zähler.
  it('zählt offene Schreibvorgänge je Set', () => {
    schreibenBeginnt('2')
    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'name', true)

    expect(lageAus('2')).toEqual({ art: 'schreibt' })
    expect(lageAus('1')).toEqual({ art: 'geschrieben' })
  })

  it('zeigt einem anderen Set nichts vom ersten', () => {
    schreibenBeginnt('1')
    schreibenEndet('1', '5', 'name', false)

    expect(lageAus('2')).toEqual({ art: 'still' })
  })
})
