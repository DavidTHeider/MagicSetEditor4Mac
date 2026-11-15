// **Die Liste der eigenen Schlüsselwörter** — UXR-F3
// (#2009).
//
// # Was MSE2 hier tut
//
// Eine Liste mit den Spalten *Keyword*, *Matches*, *Mode*, *Uses* und
// *Reminder text* (`data/en.mse-locale/locale`, Zeilen 435–439), links neben
// dem Editor
// (`KeywordList`, `src/gui/control/keyword_list.cpp`, Zeilen 34–38 und
// 193–222, Fassung f9d9356 vom 2021-07-06). Das Muster zeigt es **ohne
// Auszeichnung**, die Parameter in Winkelklammern (`match_string`, Zeilen
// 145–151) — und ein Wort ohne greifendes Muster rot.
//
// # Was hier geprüft wird
//
// Die Gestalt (Spalten, Muster ohne Tags, Wort ohne Namen, Wort ohne Wirkung)
// und die Tastatur **wie in der Kartenliste**: ↑ und ↓ wählen und nehmen den
// Fokus mit, Pos1 und Ende springen an die Enden, am Rand ist Schluss (kein
// Umlauf), und der Tabulator verlässt die Liste als Ganzes
// (`kartenlistetasten.ts`).
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { expect, describe, it, vi } from 'vitest'

import type { SchluesselwortDTO } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'
import { fuenf, texte, wort } from '../test/schluesselwoerter'
import { Schluesselwortliste } from './Schluesselwortliste'

// Beobachtet zeichnet die Liste mit einer Wahl, die sie sich selbst merkt —
// so wie die Seite es tut.
function Beobachtet({
  woerter,
  anfang,
  gewaehlt,
}: {
  woerter: SchluesselwortDTO[]
  anfang?: string | undefined
  gewaehlt?: ((kennung: string) => void) | undefined
}) {
  const [wahl, setzeWahl] = useState(anfang)

  return (
    <Sprachraum sprache="de">
      <Schluesselwortliste
        woerter={woerter}
        gewaehlt={wahl}
        waehlen={(kennung) => {
          gewaehlt?.(kennung)
          setzeWahl(kennung)
        }}
      />
    </Sprachraum>
  )
}

function zeige(
  woerter: SchluesselwortDTO[],
  anfang?: string,
  gewaehlt?: (kennung: string) => void,
) {
  render(<Beobachtet woerter={woerter} anfang={anfang} gewaehlt={gewaehlt} />)

  return screen.getByRole('grid', { name: texte.eigene })
}

const reihen = (gitter: HTMLElement) =>
  within(gitter).getAllByRole('row').slice(1)

// reihe gibt die Zeile an einer Stelle — **und wirft, wo keine ist**, statt
// `undefined` weiterzureichen: ein Test, der ins Leere greift, soll sagen,
// wo.
function reihe(gitter: HTMLElement, stelle: number): HTMLElement {
  const gefunden = reihen(gitter)[stelle]
  if (gefunden === undefined) {
    throw new Error(`Keine Zeile an Stelle ${String(stelle)}.`)
  }

  return gefunden
}

describe('Die Liste der eigenen Schlüsselwörter', () => {
  it('zeigt Name, Muster und Verwendung in Spalten mit Kopf', () => {
    const gitter = zeige([
      wort({ id: '7', wort: 'Beben', muster: 'Beben ', modus: 'custom' }),
    ])

    const kopf = within(gitter).getAllByRole('columnheader')
    expect(kopf.map((zelle) => zelle.textContent)).toEqual([
      texte.wort,
      texte.muster,
      texte.modus,
    ])

    const zellen = within(reihe(gitter, 0)).getAllByRole('gridcell')
    expect(zellen.map((zelle) => zelle.textContent.trim())).toEqual([
      'Beben',
      'Beben',
      'custom',
    ])
  })

  // **Wie MSE es zeigt** (`match_string`): die Parameter in Winkelklammern,
  // jede andere Auszeichnung weg. Wer das rohe Muster zeigte, zeigte
  // `<atom-param>` in der Liste.
  it('zeigt das Muster ohne Auszeichnung und die Parameter in Klammern', () => {
    const gitter = zeige([
      wort({
        id: '7',
        wort: 'Kicker',
        muster: 'Kicker <atom-param>zahl</atom-param> <i>x</i>',
      }),
    ])

    expect(
      within(reihe(gitter, 0)).getAllByRole('gridcell')[1],
    ).toHaveTextContent('Kicker ‹zahl› x')
  })

  // **Ein Wort ohne Namen ist eine Zeile, die etwas sagt** — angelegt ist es
  // leer, und eine Zeile ohne Text hätte keinen Namen (§6).
  it('nennt ein Wort ohne Namen beim Platzhalter', () => {
    const gitter = zeige([wort({ id: '7' })])

    const zeile = reihe(gitter, 0)

    expect(zeile).toHaveAccessibleName(texte.ohneName)
    expect(within(zeile).getByRole('button')).toHaveTextContent(texte.ohneName)
  })

  // **Rot, wie bei MSE — und als Text**, damit es nicht an der Farbe hängt.
  it('sagt bei einem Wort ohne greifendes Muster, dass es nichts erkennt', () => {
    const gitter = zeige([
      wort({ id: '1', wort: 'Gut' }),
      wort({ id: '2', wort: 'Kaputt', gueltig: false }),
    ])

    expect(
      within(reihe(gitter, 1)).getByText(texte.ungueltig),
    ).toBeInTheDocument()
    expect(within(reihe(gitter, 0)).queryByText(texte.ungueltig)).toBeNull()
  })

  it('wählt mit einem Klick und sagt, welches gewählt ist', async () => {
    const gewaehlt = vi.fn()
    const gitter = zeige(fuenf(), undefined, gewaehlt)

    await userEvent.click(within(reihe(gitter, 2)).getByRole('button'))

    expect(gewaehlt).toHaveBeenCalledWith('3')
    expect(reihen(gitter).map((zeile) => zeile.ariaSelected)).toEqual([
      'false',
      'false',
      'true',
      'false',
      'false',
    ])
    expect(within(reihe(gitter, 2)).getByRole('button')).toHaveAttribute(
      'aria-current',
      'true',
    )
  })

  it('zählt die Zeilen samt Kopf für eine Vorlesehilfe', () => {
    const gitter = zeige(fuenf())

    expect(gitter).toHaveAttribute('aria-rowcount', '6')
    expect(reihen(gitter).map((zeile) => zeile.ariaRowIndex)).toEqual([
      '2',
      '3',
      '4',
      '5',
      '6',
    ])
  })

  it('hält hier die Regeln der Barrierefreiheit', async () => {
    zeige(fuenf(), '2')

    expect(await verstoesse(document.body)).toEqual([])
  })
})

describe('Die Tabstelle der Liste', () => {
  const tabstellen = (gitter: HTMLElement) =>
    reihen(gitter)
      .filter((zeile) => zeile.tabIndex === 0)
      .map((zeile) => zeile.getAttribute('aria-label'))

  it('liegt auf dem gewählten Wort — und nur dort', () => {
    const gitter = zeige(fuenf(), '4')

    expect(tabstellen(gitter)).toEqual(['Sog'])
  })

  // **Ohne Wahl führt der Tabulator in die Liste**, wohin genau, entscheidet
  // sie: auf die erste Zeile. Sonst wäre die Liste nur mit der Maus zu
  // erreichen.
  it('liegt ohne Wahl auf der ersten Zeile', () => {
    const gitter = zeige(fuenf())

    expect(tabstellen(gitter)).toEqual(['Beben'])
  })

  it('liegt auf der ersten Zeile, wenn das gewählte Wort fehlt', () => {
    const gitter = zeige(fuenf(), '99')

    expect(tabstellen(gitter)).toEqual(['Beben'])
  })

  it('verlässt die Liste mit einem Tabulator als Ganzes', async () => {
    render(
      <>
        <button type="button">davor</button>
        <Beobachtet woerter={fuenf()} anfang="2" />
        <button type="button">danach</button>
      </>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'davor' }))
    await userEvent.tab()
    expect(screen.getByRole('row', { name: 'Kicker' })).toHaveFocus()

    await userEvent.tab()
    expect(screen.getByRole('button', { name: 'danach' })).toHaveFocus()
  })
})

describe('Die Tasten in der Liste', () => {
  function fokussiere(name: string) {
    const zeile = screen.getByRole('row', { name })
    zeile.focus()
    expect(zeile).toHaveFocus()
  }

  it('wählen mit ↓ und ↑ und nehmen den Fokus mit', async () => {
    const gewaehlt = vi.fn()
    zeige(fuenf(), '1', gewaehlt)
    fokussiere('Beben')

    await userEvent.keyboard('{ArrowDown}')
    expect(gewaehlt).toHaveBeenLastCalledWith('2')
    expect(screen.getByRole('row', { name: 'Kicker' })).toHaveFocus()

    await userEvent.keyboard('{ArrowDown}')
    expect(screen.getByRole('row', { name: 'Rausch' })).toHaveFocus()

    await userEvent.keyboard('{ArrowUp}')
    expect(gewaehlt).toHaveBeenLastCalledWith('2')
    expect(screen.getByRole('row', { name: 'Kicker' })).toHaveFocus()
  })

  it('springen mit Pos1 und Ende an die Enden', async () => {
    zeige(fuenf(), '3')
    fokussiere('Rausch')

    await userEvent.keyboard('{End}')
    expect(screen.getByRole('row', { name: 'Wacht' })).toHaveFocus()
    expect(screen.getByRole('row', { name: 'Wacht' })).toHaveAttribute(
      'aria-selected',
      'true',
    )

    await userEvent.keyboard('{Home}')
    expect(screen.getByRole('row', { name: 'Beben' })).toHaveFocus()
  })

  // **Kein Umlauf** — MSEs `canSelectPrevious` und `canSelectNext` sagen am
  // Rand „nein", und dabei bleibt es. Mit Beleg, dass jemand zuhört: dieselbe
  // Taste in der Mitte wählt.
  it('laufen am Rand nicht um', async () => {
    const gewaehlt = vi.fn()
    zeige(fuenf(), '5', gewaehlt)
    fokussiere('Wacht')

    await userEvent.keyboard('{ArrowDown}')
    expect(gewaehlt).not.toHaveBeenCalled()
    expect(screen.getByRole('row', { name: 'Wacht' })).toHaveFocus()

    await userEvent.keyboard('{ArrowUp}')
    expect(gewaehlt).toHaveBeenCalledWith('4')
  })

  it('lassen ⌘ und ⇧ in Ruhe', async () => {
    const gewaehlt = vi.fn()
    zeige(fuenf(), '2', gewaehlt)
    fokussiere('Kicker')

    await userEvent.keyboard('{Meta>}{ArrowDown}{/Meta}')
    await userEvent.keyboard('{Shift>}{ArrowDown}{/Shift}')
    expect(gewaehlt).not.toHaveBeenCalled()

    await userEvent.keyboard('{ArrowDown}')
    expect(gewaehlt).toHaveBeenCalledWith('3')
  })

  it('führen mit → in die Zeile und mit ← zurück', async () => {
    zeige(fuenf(), '2')
    fokussiere('Kicker')

    await userEvent.keyboard('{ArrowRight}')
    expect(
      within(screen.getByRole('row', { name: 'Kicker' })).getByRole('button'),
    ).toHaveFocus()

    await userEvent.keyboard('{ArrowLeft}')
    expect(screen.getByRole('row', { name: 'Kicker' })).toHaveFocus()
  })

  // **Ohne Wahl wählt keine Pfeiltaste die erste Zeile**: ↓ führt zur zweiten,
  // Pos1 bleibt, wo es ist. Wer mit dem Tabulator hereinkommt, wählt deshalb
  // mit ⏎ oder der Leertaste — wie in jeder Liste.
  it('wählen das fokussierte Wort mit ⏎ und mit der Leertaste', async () => {
    const gewaehlt = vi.fn()
    zeige(fuenf(), undefined, gewaehlt)

    fokussiere('Beben')
    await userEvent.keyboard('{Enter}')
    expect(gewaehlt).toHaveBeenLastCalledWith('1')
    expect(screen.getByRole('row', { name: 'Beben' })).toHaveAttribute(
      'aria-selected',
      'true',
    )

    fokussiere('Sog')
    await userEvent.keyboard(' ')
    expect(gewaehlt).toHaveBeenLastCalledWith('4')
    expect(gewaehlt).toHaveBeenCalledTimes(2)
  })

  // **Auf dem Knopf in der Zeile wählt ⏎ einmal** — der Knopf tut es selbst,
  // und die Zeile darf es nicht ein zweites Mal tun.
  it('wählen mit ⏎ auf dem Knopf der Zeile genau einmal', async () => {
    const gewaehlt = vi.fn()
    zeige(fuenf(), undefined, gewaehlt)

    fokussiere('Kicker')
    await userEvent.keyboard('{ArrowRight}')
    expect(
      within(screen.getByRole('row', { name: 'Kicker' })).getByRole('button'),
    ).toHaveFocus()
    await userEvent.keyboard('{Enter}')

    expect(gewaehlt).toHaveBeenCalledTimes(1)
    expect(gewaehlt).toHaveBeenCalledWith('2')
  })

  // **Die Taste ist verbraucht, auch am Rand** — sonst rollte die Seite.
  it('verbrauchen ↓ auch am Ende', () => {
    zeige(fuenf(), '5')

    const verbraucht = fireEvent.keyDown(
      screen.getByRole('row', { name: 'Wacht' }),
      { key: 'ArrowDown' },
    )

    expect(verbraucht).toBe(false)
  })
})
