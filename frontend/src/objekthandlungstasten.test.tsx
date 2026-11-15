// **⇞ und ⇟ blättern, auch wo die Leiste sie nicht zustellt** — AP-7
// (#2364),
// Befund B-02.
//
// Die Leiste zeigt die beiden Kürzel, wie MSE2 sie führt
// (`data/en.mse-locale/locale`, Zeilen 48–49, Fassung f9d9356 vom
// 2021-07-06). Auslösen kann sie sie nicht: Wails setzt für „page up" die
// Glyphe U+21DE als Äquivalent, eine Taste liefert aber `NSPageUpFunctionKey`
// — und ohne Befehlstaste fragt macOS die Leiste ohnehin nicht. **Der Raum
// hört deshalb selbst**, löst über denselben Weg aus wie ein Klick in der
// Leiste — mit denselben Sperren — und lässt ein Feld, einen Dialog und eine
// Liste, die die Taste schon verbraucht hat, in Ruhe.
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Menueeintrag } from './bausteine/Menue'
import { useHandlungsliste } from './haken/useHandlungsliste'
import { Objekthandlungsraum } from './objekthandlungsraum'
import { belegeApi, raeumeApi } from './test/api'

afterEach(() => {
  cleanup()
  raeumeApi()
})

function Liste({ eintraege }: { eintraege: readonly Menueeintrag[] }) {
  useHandlungsliste({ titel: 'Karte', art: 'karte', kennung: '1', eintraege })

  return null
}

async function zeige(eintraege: readonly Menueeintrag[]) {
  belegeApi({ ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }) })
  render(
    <Objekthandlungsraum>
      <Liste eintraege={eintraege} />
      <input aria-label="Name" />
      <div role="dialog" aria-label="Frage">
        <button type="button">Im Dialog</button>
      </div>
    </Objekthandlungsraum>,
  )
  await act(async () => {
    await Promise.resolve()
  })
}

function blaettern() {
  const vorige = vi.fn()
  const naechste = vi.fn()

  return {
    vorige,
    naechste,
    eintraege: [
      {
        schluessel: 'vorige',
        name: 'Vorige',
        kuerzel: 'bild auf',
        aufWahl: vorige,
      },
      {
        schluessel: 'naechste',
        name: 'Nächste',
        kuerzel: 'bild ab',
        aufWahl: naechste,
      },
    ] satisfies Menueeintrag[],
  }
}

// **⇞ und ⇟ hört der Raum nicht** — Befund 1 des Zweitblicks über Phase 6
// (erster Teil): sie blättern seit #1290 über `useKartenblaettern`, und das
// mit der Regel aus MSE2 — auch aus einem Feld, mit dem Fokus, der
// mitreist. Ein zweiter Horcher mit anderer Regel wählte dieselbe Karte ein
// zweites Mal.
describe('⇞ und ⇟ am Raum', () => {
  it('bleiben bei der Seite', async () => {
    const { naechste, eintraege } = blaettern()
    await zeige(eintraege)

    expect(fireEvent.keyDown(document.body, { key: 'PageDown' })).toBe(true)
    expect(naechste).not.toHaveBeenCalled()
  })
})

// **Auch die Kürzel mit ⌘** — gemessen am 2026-10-16 mit `ap5kontext09`:
// ⌘⌫ außerhalb eines Felds löste am nativen Fenster nichts aus. Wails setzt
// für „backspace" U+0008 als Äquivalent, die Taste liefert DEL (U+007F),
// und die Leiste fand keinen Eintrag. Der Raum hört deshalb alle Kürzel der
// Objekthandlungen; verbraucht er eine Taste, fragt WebKit die Leiste nicht
// mehr — ausgelöst wird einmal.
describe('die Kürzel mit ⌘ am Raum', () => {
  function handlungen() {
    const loeschen = vi.fn()
    const verdoppeln = vi.fn()

    return {
      loeschen,
      verdoppeln,
      eintraege: [
        {
          schluessel: 'loeschen',
          name: 'Löschen',
          kuerzel: 'befehl+backspace',
          aufWahl: loeschen,
        },
        {
          schluessel: 'duplizieren',
          name: 'Verdoppeln',
          kuerzel: 'befehl+d',
          aufWahl: verdoppeln,
        },
      ] satisfies Menueeintrag[],
    }
  }

  it('lösen außerhalb eines Felds aus und verbrauchen die Taste', async () => {
    const { loeschen, verdoppeln, eintraege } = handlungen()
    await zeige(eintraege)

    expect(
      fireEvent.keyDown(document.body, { key: 'Backspace', metaKey: true }),
    ).toBe(false)
    fireEvent.keyDown(document.body, { key: 'd', metaKey: true })

    expect(loeschen).toHaveBeenCalledTimes(1)
    expect(verdoppeln).toHaveBeenCalledTimes(1)
  })

  it('lassen ein Feld und einen weiteren Umschalter in Ruhe', async () => {
    const { loeschen, verdoppeln, eintraege } = handlungen()
    await zeige(eintraege)

    expect(
      fireEvent.keyDown(screen.getByRole('textbox', { name: 'Name' }), {
        key: 'Backspace',
        metaKey: true,
      }),
    ).toBe(true)
    fireEvent.keyDown(document.body, {
      key: 'd',
      metaKey: true,
      shiftKey: true,
    })
    fireEvent.keyDown(document.body, { key: 'd' })

    expect(loeschen).not.toHaveBeenCalled()
    expect(verdoppeln).not.toHaveBeenCalled()
  })

  it('lassen ein ⌘-Kürzel ohne Eintrag dem Fenster', async () => {
    const { loeschen, verdoppeln, eintraege } = handlungen()
    await zeige(eintraege)

    expect(fireEvent.keyDown(document.body, { key: 'x', metaKey: true })).toBe(
      true,
    )
    expect(loeschen).not.toHaveBeenCalled()
    expect(verdoppeln).not.toHaveBeenCalled()
  })
})
