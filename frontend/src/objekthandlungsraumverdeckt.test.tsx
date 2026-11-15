// **Ein Dialog verdeckt das Objekt** — Rest R1 des Nachlaufs über Phase 6
// von 0.9.0 (#2364): ⌘D aus einem Feld im Dialog ging am Raum vorbei an die
// Leiste, und die Leiste verdoppelte die Karte dahinter. **Wie ein Blatt
// unter macOS**: solange der Fokus in einem Dialog steht, meldet der Raum
// die Einträge gesperrt, und ein Klick, der trotzdem ankommt, bewirkt
// nichts. Ein offenes Menü sperrt nicht — warum, steht am Fall.
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Menueeintrag } from './bausteine/Menue'
import { ereignisObjekthandlung } from './fenster'
import { useHandlungsliste } from './haken/useHandlungsliste'
import { Objekthandlungsraum } from './objekthandlungsraum'
import { belegeApi, raeumeApi } from './test/api'
import { laufzeitprobe, type Laufzeitprobe } from './test/laufzeitprobe'

let laufzeit: Laufzeitprobe

beforeEach(() => {
  laufzeit = laufzeitprobe()
})

afterEach(() => {
  cleanup()
  laufzeit.raeume()
  raeumeApi()
})

function Objekt({ eintraege }: { eintraege: readonly Menueeintrag[] }) {
  useHandlungsliste({ titel: 'Karte', art: 'karte', kennung: '7', eintraege })

  return null
}

function zeige(dialog: boolean, rolle: 'dialog' | 'menu' = 'dialog') {
  const ObjekthandlungenMelden = vi.fn(() => Promise.resolve({ fehler: null }))
  belegeApi({ ObjekthandlungenMelden })
  const duplizieren = vi.fn()

  const baum = (offen: boolean) => (
    <Objekthandlungsraum>
      <Objekt
        eintraege={[
          {
            schluessel: 'duplizieren',
            name: 'Duplizieren',
            aufWahl: duplizieren,
          },
        ]}
      />
      <button type="button">Daneben</button>
      {offen && (
        <div role={rolle} aria-label="Überlagerung">
          <input aria-label="Name" />
        </div>
      )}
    </Objekthandlungsraum>
  )
  const { rerender } = render(baum(dialog))

  return {
    ObjekthandlungenMelden,
    duplizieren,
    schliessen: () => {
      rerender(baum(false))
    },
  }
}

const gesperrt = (wert: boolean): unknown =>
  expect.objectContaining({
    eintraege: [expect.objectContaining({ gesperrt: wert })],
  })

const klick = () => {
  act(() => {
    laufzeit.melde(ereignisObjekthandlung, {
      art: 'karte',
      kennung: '7',
      epoche: 1,
      schluessel: 'duplizieren',
    })
  })
}

describe('Ein Dialog verdeckt das Objekt', () => {
  it('meldet die Leiste gesperrt, solange der Fokus in einem Dialog steht', async () => {
    const { ObjekthandlungenMelden } = zeige(true)
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(false))
    })

    act(() => {
      screen.getByLabelText('Name').focus()
    })
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(true))
    })

    act(() => {
      screen.getByRole('button', { name: 'Daneben' }).focus()
    })
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(false))
    })
  })

  // **Ein offenes Menü behält die Leiste** — es zeigt die Handlungen
  // desselben Objekts, und ein Wechsel über die Leiste schließt es (AP-5,
  // `ap5kontext09`). Gesperrt hätte er „Nächste Karte“ grau gemacht, während
  // das Menü offen ist.
  it('lässt die Leiste offen, solange der Fokus in einem Menü steht', async () => {
    const { ObjekthandlungenMelden, duplizieren } = zeige(true, 'menu')
    await waitFor(() => {
      expect(laufzeit.horchende(ereignisObjekthandlung)).toBe(1)
    })

    act(() => {
      screen.getByLabelText('Name').focus()
    })
    klick()

    expect(duplizieren).toHaveBeenCalledTimes(1)
    expect(ObjekthandlungenMelden).not.toHaveBeenCalledWith(gesperrt(true))
  })

  it('gibt die Leiste frei, wenn der Fokus ins Leere geht', async () => {
    const { ObjekthandlungenMelden } = zeige(true)
    act(() => {
      screen.getByLabelText('Name').focus()
    })
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(true))
    })

    act(() => {
      screen.getByLabelText('Name').blur()
    })

    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(false))
    })
  })

  // **Ein Dialog, der mit dem Fokus darin verschwindet**, meldet kein
  // `focusout` — der Raum sieht es an der Entfernung.
  it('gibt die Leiste frei, wenn der Dialog samt Fokus verschwindet', async () => {
    const { ObjekthandlungenMelden, schliessen } = zeige(true)
    act(() => {
      screen.getByLabelText('Name').focus()
    })
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(true))
    })

    schliessen()

    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(false))
    })
  })

  // **Eine Änderung im Dialog ist keine Freigabe** — wer tippt, ändert den
  // Baum, und der Fokus bleibt, wo er war.
  it('bleibt gesperrt, solange sich im Dialog etwas ändert', async () => {
    const { ObjekthandlungenMelden, schliessen } = zeige(true)
    act(() => {
      screen.getByLabelText('Name').focus()
    })
    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(true))
    })
    const bisher = ObjekthandlungenMelden.mock.calls.length

    // **Der Beobachter meldet eine Mikroaufgabe später** — erst danach geht
    // der Dialog, sonst sähe er beides auf einmal.
    await act(async () => {
      screen.getByRole('dialog').append(document.createElement('span'))
      await Promise.resolve()
    })
    schliessen()

    await waitFor(() => {
      expect(ObjekthandlungenMelden).toHaveBeenLastCalledWith(gesperrt(false))
    })
    expect(ObjekthandlungenMelden).toHaveBeenCalledTimes(bisher + 1)
  })

  it('überhört einen Klick aus der Leiste, solange ein Dialog den Fokus hat', async () => {
    const { duplizieren } = zeige(true)
    await waitFor(() => {
      expect(laufzeit.horchende(ereignisObjekthandlung)).toBe(1)
    })

    act(() => {
      screen.getByLabelText('Name').focus()
    })
    klick()
    expect(duplizieren).not.toHaveBeenCalled()

    // **Ein Beleg, dass jemand zuhört** (§3): derselbe Klick außerhalb wirkt.
    act(() => {
      screen.getByRole('button', { name: 'Daneben' }).focus()
    })
    klick()
    expect(duplizieren).toHaveBeenCalledTimes(1)
  })
})
