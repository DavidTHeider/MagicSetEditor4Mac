// **Eine kurze Eingabe dort, wo sie hingehört** — UX-08.AK-01
// (#1680).
import { MantineProvider } from '@mantine/core'
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { mantinethema } from '../mantine'
import { Sprachraum } from '../sprache/kontext'
import { Popover } from './Popover'

// Probe stellt ein Popover an einen Auslöser, wie eine Zeile es tut — **in
// einer schmalen Spalte, die abschneidet**, wie der Navigator der Decks.
function Probe() {
  const [offen, setzeOffen] = useState(false)

  return (
    <>
      <div className="w-60 overflow-y-auto">
        <Popover
          ausloeser={
            <button
              type="button"
              onClick={() => {
                setzeOffen(true)
              }}
            >
              Umbenennen
            </button>
          }
          offen={offen}
          titel="Neuer Name"
          schliessen={() => {
            setzeOffen(false)
          }}
        >
          <input aria-label="Name" defaultValue="Aggro" />
        </Popover>
      </div>

      <button type="button">Daneben</button>

      {/* **Wie ein Sichern, das erst später fertig wird** — es schließt das
          Popover, während der Fokus längst woanders steht. */}
      <button
        type="button"
        onClick={() => {
          setzeOffen(false)
        }}
      >
        Von außen
      </button>
    </>
  )
}

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Probe />
    </Sprachraum>,
  )
}

describe('Popover', () => {
  // **Eine Ebene über der Seite** — AP-4 (#2361), Entwurf `Zustaende09`,
  // Vermerk 0065 Festlegung 11: die Kante bleibt, eine Schattenstufe kommt
  // hinzu, mehr Kontrast nimmt sie weg; und es erscheint in `--dauer-mittel`.
  it('liegt mit Kante und einer Schattenstufe über der Seite', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))

    expect(screen.getByRole('group', { name: 'Neuer Name' })).toHaveClass(
      'border-2',
      'shadow-ueberlage',
      'contrast-more:shadow-none',
    )
  })

  it('erscheint in der mittleren Dauer des Stilblatts', async () => {
    document.documentElement.style.setProperty('--dauer-mittel', '160ms')
    try {
      render(
        <Sprachraum sprache="de">
          <MantineProvider theme={mantinethema} env="default">
            <Probe />
          </MantineProvider>
        </Sprachraum>,
      )

      await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))

      // **Gewartet wird auf das Dropdown**: unter der Betriebsumgebung hängt
      // Mantine es erst ein, wenn sein Übergang beginnt.
      const gruppe = (await screen.findByText('Neuer Name')).closest(
        '.mantine-Popover-dropdown',
      )
      expect(gruppe).toBeInstanceOf(HTMLElement)
      expect((gruppe as HTMLElement).style.transitionDuration).toBe('160ms')
    } finally {
      document.documentElement.style.removeProperty('--dauer-mittel')
    }
  })

  it('geht mit dem Fokus im ersten Feld auf, in einer benannten Gruppe', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))

    expect(
      screen.getByRole('group', { name: 'Neuer Name' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'Name' })).toHaveFocus()
  })

  // **Es steht nicht in der Spalte seines Auslösers** — sonst schneidet sie
  // es ab. Gesehen in der WebKit-Probe `ux08-popover` am 2026-09-25: das
  // Namensfeld endete an der Kante des Navigators.
  //
  // **Unter der Umgebung des Betriebs**: unter `test` stellt Mantine jedes
  // Portal an Ort und Stelle (`OptionalPortal.mjs`, Mantine 9.6.1) — und
  // genau das Portal ist hier die Frage.
  it('steht außerhalb des Behälters, in dem sein Auslöser steht', async () => {
    const { container } = render(
      <Sprachraum sprache="de">
        <MantineProvider theme={mantinethema} env="default">
          <Probe />
        </MantineProvider>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))

    // **Gefragt wird nach dem Ort, nicht nach der Sichtbarkeit**: jsdom
    // rechnet kein Layout, und Mantine hält einen Auslöser ohne Maße für
    // verdeckt und blendet aus (`hideDetached`) — dann hat die Gruppe keinen
    // zugänglichen Namen mehr. Gefunden wird sie an ihrem Titel; dass sie
    // sichtbar dasteht, zeigt die WebKit-Probe.
    const titel = screen.getByText('Neuer Name')
    expect(container).toContainElement(
      screen.getByRole('button', { name: 'Umbenennen' }),
    )
    expect(document.body).toContainElement(titel)
    expect(container).not.toContainElement(titel)
  })

  // **Escape schließt, und der Fokus kehrt zurück** — auch nach einem Klick,
  // der wie in WebKit keinen Fokus gab.
  // **Tab bleibt drin** — angestrichen vom Zweitblick über Phase 6.3 (M4):
  // das Popover steht im Portal hinter `#root`, und Tab vom letzten Feld
  // ging ans Ende des Dokuments, während es offen blieb (WCAG 2.4.3).
  it('hält den Fokus beim Tabben in sich', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))
    await userEvent.tab()

    expect(screen.getByRole('textbox', { name: 'Name' })).toHaveFocus()
    await userEvent.tab({ shift: true })
    expect(screen.getByRole('textbox', { name: 'Name' })).toHaveFocus()
  })

  it('schließt mit Escape und gibt den Fokus dem Auslöser zurück', () => {
    zeige()

    const knopf = screen.getByRole('button', { name: 'Umbenennen' })
    fireEvent.pointerDown(knopf)
    fireEvent.click(knopf)

    fireEvent.keyDown(screen.getByRole('textbox', { name: 'Name' }), {
      key: 'Escape',
    })

    expect(screen.queryByRole('group', { name: 'Neuer Name' })).toBeNull()
    expect(knopf).toHaveFocus()
  })

  // **Zum Auslöser, auch wenn der Fokus beim Aufgehen nirgends stand** —
  // QA01 (#2382): gemerkt war `body`, und Escape ließ ihn dort. Gefunden beim
  // Nachgehen der WebKit-Abnahme `ux08-popover`; deren Rot war am Ende ein
  // Wettlauf im Drehbuch, dieser Fehler ist trotzdem einer.
  it('gibt den Fokus dem Auslöser zurück, auch wenn er beim Aufgehen nirgends stand', () => {
    zeige()

    const knopf = screen.getByRole('button', { name: 'Umbenennen' })
    expect(document.activeElement).toBe(document.body)
    fireEvent.click(knopf)

    fireEvent.keyDown(screen.getByRole('textbox', { name: 'Name' }), {
      key: 'Escape',
    })

    expect(screen.queryByRole('group', { name: 'Neuer Name' })).toBeNull()
    expect(knopf).toHaveFocus()
  })

  it('schließt mit einem Klick daneben, nicht mit einem hinein', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))
    await userEvent.click(screen.getByRole('textbox', { name: 'Name' }))
    expect(
      screen.getByRole('group', { name: 'Neuer Name' }),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Daneben' }))
    expect(screen.queryByRole('group', { name: 'Neuer Name' })).toBeNull()

    // **Und der Fokus bleibt, wo geklickt wurde** — wer daneben klickt,
    // meint das, worauf er klickt, und nicht den Auslöser von vorhin.
    expect(screen.getByRole('button', { name: 'Daneben' })).toHaveFocus()
  })

  // **Der Fokus kehrt nur zurück, wenn er sonst verloren wäre** — wer
  // inzwischen woanders arbeitet, dem nimmt ein spätes Schließen ihn nicht.
  it('nimmt beim Schließen keinen Fokus, der woanders steht', async () => {
    zeige()

    await userEvent.click(screen.getByRole('button', { name: 'Umbenennen' }))
    const aussen = screen.getByRole('button', { name: 'Von außen' })
    aussen.focus()
    await userEvent.keyboard('{Enter}')

    expect(screen.queryByRole('group', { name: 'Neuer Name' })).toBeNull()
    expect(aussen).toHaveFocus()
  })
})
