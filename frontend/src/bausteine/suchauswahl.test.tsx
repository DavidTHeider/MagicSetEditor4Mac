// **Eine Auswahl mit Suche, Teil 1: Gestalt und Rollen** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: ab fünfzehn Werten gehört zur Auswahl eine
// Suche.
//
// Geprüft wird das ARIA-Muster des Kombinationsfelds, wie die Praktiken zur
// Barrierefreiheit es beschreiben: ein Eingabefeld mit `role="combobox"`, das
// eine Liste (`listbox`) mit `option` öffnet, der Fokus bleibt im Feld, und
// die hervorgehobene Zeile steht in `aria-activedescendant`. **Die Tasten
// stehen in `suchauswahltasten.test.tsx`.**
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'
import { Probe, angeboten, feld, stile, zonen } from '../test/suchauswahlprobe'
import { Suchauswahl } from './Suchauswahl'
import { passtZu } from './suchwahl'
import type { Suchwahl } from './suchwahl'

describe('passtZu', () => {
  const eintrag: Suchwahl = {
    wert: 'x',
    name: 'America/New_York',
    zusatz: 'probe-schlicht',
  }

  it.each([
    ['über den Namen', 'york', true],
    ['ohne Rücksicht auf Groß und klein', 'NEW_YORK', true],
    ['über ein Teilwort', 'merica', true],
    ['über den Zusatz', 'schlicht', true],
    ['mit Leerzeichen statt Unterstrich', 'new york', true],
    ['mit Leerzeichen statt Schrägstrich', 'america new', true],
    ['über die Schreibweise der Kennung', 'america/new', true],
    ['über Name und Zusatz zusammen', 'york probe', true],
    ['mit Rand aus Leerzeichen', '  york  ', true],
    ['mit leerer Suche', '', true],
    ['mit einer Suche aus Leerzeichen', '   ', true],
    ['nicht, wenn nichts übereinstimmt', 'tokio', false],
  ])('prüft %s', (_, suche, soll) => {
    expect(passtZu(eintrag, suche)).toBe(soll)
  })

  it('prüft einen Eintrag ohne Zusatz', () => {
    expect(passtZu({ wert: 'x', name: 'Africa/Cairo' }, 'cairo')).toBe(true)

    // **Kein „undefined" im Heuhaufen**: ohne Zusatz steht dort nichts.
    expect(passtZu({ wert: 'x', name: 'Africa/Cairo' }, 'undefined')).toBe(
      false,
    )
  })
})

describe('Suchauswahl — die Rollen', () => {
  // **Das Muster des Kombinationsfelds**, mit dem Namen aus der Beschriftung.
  it('ist ein Kombinationsfeld mit Namen, zu und ohne Liste', () => {
    render(<Probe />)

    expect(feld()).toHaveAttribute('aria-expanded', 'false')
    expect(feld()).toHaveAttribute('aria-autocomplete', 'list')
    expect(feld()).toHaveAttribute('aria-haspopup', 'listbox')
    expect(feld()).toHaveAttribute('autocomplete', 'off')
    expect(feld()).not.toHaveAttribute('aria-controls')
    expect(feld()).not.toHaveAttribute('aria-activedescendant')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  // **Die Liste steht nur da, solange sie offen ist**: Mantine hielte sie
  // sonst verborgen im Baum, und bei vierhundert Zeilen wären das vierhundert
  // Knoten je Feld. `queryByRole` sähe sie nicht — sie wäre `display: none`.
  it('hält die Liste nicht im Baum, solange sie zu ist', async () => {
    render(<Probe />)

    expect(document.querySelector('[role="listbox"]')).toBeNull()
    expect(document.querySelector('[role="option"]')).toBeNull()

    await userEvent.click(feld())
    await userEvent.keyboard('{Escape}')

    expect(document.querySelector('[role="listbox"]')).toBeNull()
  })

  it('zeigt die Wahl mit ihrem Namen', () => {
    render(<Probe />)

    expect(feld()).toHaveValue('Europe/Berlin')
  })

  it('öffnet auf einen Klick eine benannte Liste, die das Feld steuert', async () => {
    render(<Probe />)

    await userEvent.click(feld())

    const liste = screen.getByRole('listbox', { name: 'Zeitzone' })

    expect(feld()).toHaveAttribute('aria-expanded', 'true')
    expect(feld()).toHaveAttribute('aria-controls', liste.id)
    expect(angeboten()).toEqual([
      'Wie das System',
      'Africa/Cairo',
      'America/New_York',
      'Europe/Berlin',
      'Europe/Paris',
    ])
  })

  // **Der Fokus bleibt im Feld**: die hervorgehobene Zeile steht in
  // `aria-activedescendant`, und die Liste nimmt ihn nie.
  it('hebt beim Öffnen die Wahl hervor und lässt den Fokus im Feld', async () => {
    render(<Probe />)

    await userEvent.click(feld())

    const gewaehlt = screen.getByRole('option', { name: 'Europe/Berlin' })

    expect(feld()).toHaveFocus()
    expect(feld()).toHaveAttribute('aria-activedescendant', gewaehlt.id)
    expect(gewaehlt).toHaveAttribute('aria-selected', 'true')
    expect(
      screen.getByRole('option', { name: 'Europe/Paris' }),
    ).toHaveAttribute('aria-selected', 'false')
  })

  // **Markiert wird, was gilt**, nicht nur, was die Pfeile berühren.
  it('kennzeichnet die gewählte Zeile, auch wenn die Pfeile woanders stehen', async () => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard('{ArrowDown}')

    expect(
      screen.getByRole('option', { name: 'Europe/Berlin' }),
    ).toHaveAttribute('data-combobox-active')
    expect(
      screen.getByRole('option', { name: 'Europe/Paris' }),
    ).not.toHaveAttribute('data-combobox-active')
  })

  it('öffnet nicht, wenn der Fokus nur ankommt', async () => {
    render(<Probe />)

    await userEvent.tab()

    expect(feld()).toHaveFocus()
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  // **Offen ist das Feld leer und nennt die Wahl als Platzhalter**, damit
  // Tippen neu beginnt — gemessen am Fenster (`fmt02d`, #2007): eine
  // Markierung per `select()` hält WKWebView nach dem Klick nicht, und
  // „Bitte wählen" + „probe" ergab „Bitte wählenprobe".
  it('leert das Feld beim Öffnen und nennt die Wahl als Platzhalter', async () => {
    render(<Probe />)

    await userEvent.click(feld())

    expect(feld()).toHaveValue('')
    expect(feld()).toHaveAttribute('placeholder', 'Europe/Berlin')

    await userEvent.keyboard('cairo')
    expect(feld()).toHaveValue('cairo')

    await userEvent.keyboard('{Escape}')
    expect(feld()).toHaveValue('Europe/Berlin')
  })

  it('lässt einen zweiten Klick bei offener Liste den Cursor setzen', async () => {
    render(<Probe />)

    await userEvent.click(feld())
    await userEvent.keyboard('{ArrowDown}')
    await userEvent.click(feld())

    expect(screen.getByRole('listbox')).toBeInTheDocument()
    expect(feld()).toHaveAttribute(
      'aria-activedescendant',
      screen.getByRole('option', { name: 'Europe/Paris' }).id,
    )
  })

  it('führt Name und Zusatz in jeder Zeile', async () => {
    render(<Probe eintraege={stile} anfang="magic-m15" />)

    await userEvent.click(feld())

    expect(angeboten()).toEqual([
      'M15 Normal magic-m15',
      'Altered Art magic-m15-altered',
      'Schlicht probe-schlicht · probe',
    ])
  })
})

describe('Suchauswahl — der Wert', () => {
  // **Was nicht in der Liste steht, bleibt sichtbar gewählt**, statt still zu
  // verschwinden oder auf die erste Zeile zu fallen.
  it('zeigt einen Wert, den die Liste nicht kennt, wie er ist', () => {
    render(<Probe anfang="US/Pacific" />)

    expect(feld()).toHaveValue('US/Pacific')
  })

  it('zeigt ohne Wahl den Platzhalter', () => {
    render(
      <Probe anfang="" eintraege={zonen.slice(1)} platzhalter="— wählen —" />,
    )

    expect(feld()).toHaveValue('')
    expect(feld()).toHaveAttribute('placeholder', '— wählen —')
  })

  it('zeigt ohne Platzhalter keinen', () => {
    render(<Probe />)

    expect(feld()).not.toHaveAttribute('placeholder')
  })

  it('zeigt eine Wahl mit leerem Wert unter ihrem Namen', () => {
    render(<Probe anfang="" />)

    expect(feld()).toHaveValue('Wie das System')
  })

  it('ist angehalten, solange ein Vorgang läuft', async () => {
    render(<Probe angehalten />)

    expect(feld()).toBeDisabled()

    await userEvent.click(feld())

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('reicht Beschreibung, Befund und Kennzeichen an das Feld', () => {
    render(
      <Sprachraum sprache="de">
        <Suchauswahl
          id="a"
          bezeichnung="Zeitzone"
          eintraege={zonen}
          wert=""
          aendern={vi.fn()}
          beschriebenVon="hilfe"
          ungueltig
          einstellung="zeitzone"
        />
      </Sprachraum>,
    )

    const eingabe = screen.getByRole('combobox')

    expect(eingabe).toHaveAttribute('aria-describedby', 'hilfe')
    expect(eingabe).toHaveAttribute('aria-invalid', 'true')
    expect(eingabe).toHaveAttribute('data-einstellung', 'zeitzone')
  })

  // **Ohne `<label>` trägt das Feld seinen Namen selbst** — die Kartenfelder
  // haben nur ein `aria-label`.
  it('benennt sich selbst, wenn es gebeten wird', () => {
    render(
      <Sprachraum sprache="de">
        <Suchauswahl
          id="a"
          bezeichnung="Wasserzeichen"
          eintraege={zonen}
          wert=""
          aendern={vi.fn()}
          benennen
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('combobox')).toHaveAttribute(
      'aria-label',
      'Wasserzeichen',
    )
  })

  it('lässt Beschreibung, Befund und Kennzeichen weg, wo es keine gibt', () => {
    render(<Probe />)

    expect(feld()).not.toHaveAttribute('aria-describedby')
    expect(feld()).not.toHaveAttribute('aria-invalid')
    expect(feld()).not.toHaveAttribute('data-einstellung')
    expect(feld()).not.toHaveAttribute('aria-label')
  })

  it('folgt einem neuen Wert von außen', () => {
    const { rerender } = render(
      <Sprachraum sprache="de">
        <Suchauswahl
          id="a"
          bezeichnung="Zeitzone"
          eintraege={zonen}
          wert="Europe/Paris"
          aendern={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('combobox')).toHaveValue('Europe/Paris')

    rerender(
      <Sprachraum sprache="de">
        <Suchauswahl
          id="a"
          bezeichnung="Zeitzone"
          eintraege={zonen}
          wert="Africa/Cairo"
          aendern={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('combobox')).toHaveValue('Africa/Cairo')
  })
})

describe('Suchauswahl — Barrierefreiheit', () => {
  // **WCAG 2.2 AA, im offenen Zustand**: die Liste steht in einem Portal am
  // Ende von `body`, nicht im Behälter der Probe — geprüft wird deshalb
  // `body`.
  it('hält die Regeln, zu, offen, mit Treffern und ohne', async () => {
    const { container } = render(<Probe />)

    expect(await verstoesse(container)).toEqual([])

    await userEvent.click(feld())

    expect(await verstoesse(document.body)).toEqual([])

    await userEvent.keyboard('europe')

    expect(await verstoesse(document.body)).toEqual([])

    await userEvent.keyboard('{ArrowDown}')

    expect(await verstoesse(document.body)).toEqual([])

    await userEvent.keyboard('xyz')

    expect(await verstoesse(document.body)).toEqual([])
  })
})
