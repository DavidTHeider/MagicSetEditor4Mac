// **Ein Auswahlfeld des Spiels mit vielen Wahlen** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: *„`watermark` mit 229 Wahlen ohne Suche"*.
//
// Ein `choice:`-Feld hat so viele Wahlen, wie das Spielpaket aufzählt — vier
// bei der Seltenheit, 229 beim Wasserzeichen von `magic`. Bis vierzehn bleibt
// es das native Feld; ab fünfzehn ist es eine Auswahl mit Suche, die ihren Wert
// **sofort** schreibt, wie das native Feld es tat.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'
import { suchschwelle } from '../bausteine/suchwahl'
import { Feldeingabe } from './Feldeingabe'

import type { FeldbeschreibungDTO } from '../api'

// wahlen zählt so viele Wahlen auf, wie verlangt sind: „Marke 001", …
const wahlen = (anzahl: number) =>
  Array.from(
    { length: anzahl },
    (_, stelle) => `Marke ${String(stelle + 1).padStart(3, '0')}`,
  )

function form(anzahl: number): FeldbeschreibungDTO {
  return {
    name: 'watermark',
    aufschrift: 'Watermark',
    art: 'choice',
    mehrzeilig: false,
    wahlen: wahlen(anzahl),
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }
}

function zeige(
  anzahl: number,
  wert = 'Marke 002',
  teile: {
    angehalten?: boolean
    fehlerhaft?: boolean
    hergeleitet?: string
  } = {},
) {
  const aendern = vi.fn()
  const verlassen = vi.fn()

  const stand = render(
    <Sprachraum sprache="de">
      <Feldeingabe
        feld="watermark"
        wert={wert}
        form={form(anzahl)}
        angehalten={teile.angehalten ?? false}
        fehlerhaft={teile.fehlerhaft ?? false}
        beschriebenVon="hinweis"
        hergeleitet={teile.hergeleitet}
        aendern={aendern}
        verlassen={verlassen}
      />
    </Sprachraum>,
  )

  return { aendern, verlassen, stand }
}

const feld = () => screen.getByRole('combobox', { name: 'watermark' })

describe('Ein Auswahlfeld des Spiels', () => {
  it('bleibt unter der Schwelle das native Feld', () => {
    zeige(suchschwelle - 1)

    expect(feld().tagName).toBe('SELECT')
  })

  it('ist ab der Schwelle eine Auswahl mit Suche, benannt nach dem Feld', () => {
    zeige(229)

    expect(feld().tagName).toBe('INPUT')
    expect(feld()).toHaveValue('Marke 002')
    expect(feld()).toHaveAttribute('aria-describedby', 'hinweis')
  })

  it('findet eine Wahl über die Suche und schreibt sie sofort', async () => {
    const { aendern, verlassen } = zeige(229)

    await userEvent.click(feld())
    await userEvent.keyboard('marke 217{Enter}')

    expect(aendern).toHaveBeenCalledExactlyOnceWith('Marke 217')
    expect(verlassen).toHaveBeenCalledExactlyOnceWith('Marke 217')
  })

  it('schreibt nichts, wenn die Wahl dieselbe bleibt', async () => {
    const { aendern, verlassen } = zeige(229)

    await userEvent.click(feld())
    await userEvent.keyboard('{Enter}')

    expect(aendern).not.toHaveBeenCalled()
    expect(verlassen).not.toHaveBeenCalled()
  })

  // **Ein Wert, der nicht in der Liste steht, steht trotzdem da** — MSE
  // zeichnet den Wert selbst, gleich ob die Liste ihn führt.
  it('zeigt einen Wert, den die Liste nicht führt, wie er ist', () => {
    zeige(229, 'Fremdmarke')

    expect(feld()).toHaveValue('Fremdmarke')
  })

  // **Der gerechnete Wert steht im Steuerelement**, solange das Feld leer ist
  // — wie beim nativen Feld.
  it('zeigt den gerechneten Wert eines leeren Feldes', () => {
    zeige(229, '', { hergeleitet: 'Marke 007' })

    expect(feld()).toHaveValue('Marke 007')
  })

  it('ist gesperrt und fehlerhaft, wo es das Feld ist', () => {
    zeige(229, 'Marke 002', { angehalten: true, fehlerhaft: true })

    expect(feld()).toBeDisabled()
    expect(feld()).toHaveAttribute('aria-invalid', 'true')
  })

  it('hält die Regeln der Barrierefreiheit, auch geöffnet', async () => {
    const { stand } = zeige(229)

    expect(await verstoesse(stand.container)).toEqual([])

    await userEvent.click(feld())

    expect(await verstoesse(document.body)).toEqual([])
  })
})
