// **⇞ und ⇟ blättern die Karten** —
// #1290.
//
// # Auch aus einem Feld heraus, und das ist portiert
//
// Bei MSE2 hängen *Select Previous/Next Card* als Menükürzel am Fenster. Der
// Karteneditor gibt jede Taste weiter, die er nicht selbst braucht:
//
//	} else if (current_editor) {
//	  if (!current_editor->onChar(ev)) {
//	    ev.Skip();
//	  }
//
// `src/gui/control/card_editor.cpp`, Zeilen 458–461, und der Texteditor
// beansprucht ←, →, ↑, ↓, Pos1, Ende, Rück, Entf und Eingabe — **nicht** ⇞ und
// ⇟ (`src/gui/value/text.cpp`, Zeilen 412–475). Beide Fassungen f9d9356 vom
// 2021-07-06.
//
// Damit blättert MSE **während des Tippens**, und genau das ist der Weg, für
// den es dieses Werkstück gibt: anlegen, füllen, weiter — ohne Maus.
//
// # Was hier zugesichert wird
//
//   - ⇞ und ⇟ wählen die vorige und die nächste Karte.
//   - **Auch, wenn der Fokus in einem Textfeld steht.**
//   - Am Rand geschieht nichts — kein Umlauf.
//   - Mit einer Zusatztaste geschieht nichts: ⌘⇟ ist etwas anderes.
//   - **Nach dem Abräumen hört der Haken auf zu hören.**

import { render, screen } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { useKartenblaettern } from './useKartenblaettern'
import { effekteNachholen } from '../test/effekte'
import { Verdeckung } from '../verdeckung'
import type { KartenzeileDTO } from '../api'

const zeile = (id: string): KartenzeileDTO =>
  ({ id, nummer: Number(id), aufschriften: [] }) as unknown as KartenzeileDTO

const drei = [zeile('1'), zeile('2'), zeile('3')]

function Probe({
  gewaehlt,
  waehlen,
}: {
  gewaehlt: string
  waehlen: (id: string) => void
}) {
  useKartenblaettern(drei, gewaehlt, waehlen)

  return <input aria-label="feld" defaultValue="" />
}

// Wechselprobe baut ihr Feld beim Kartenwechsel neu — wie der Editor.
function Wechselprobe() {
  const [gewaehlt, setzeGewaehlt] = useState('2')

  useKartenblaettern(drei, gewaehlt, setzeGewaehlt)

  return <input key={gewaehlt} aria-label="feld" defaultValue="" />
}

async function gezeichnet(gewaehlt: string) {
  const waehlen = vi.fn()

  render(<Probe gewaehlt={gewaehlt} waehlen={waehlen} />)

  // **Ohne das hängt der Horcher noch nicht** (§3): ein `useEffect` läuft nach
  // dem Zeichnen, und ein Tastendruck davor geht verloren.
  await effekteNachholen()

  return waehlen
}

function taste(schluessel: string, zusatz: Partial<KeyboardEventInit> = {}) {
  window.dispatchEvent(
    new KeyboardEvent('keydown', { key: schluessel, bubbles: true, ...zusatz }),
  )
}

describe('useKartenblaettern', () => {
  it('führt mit ⇟ zur nächsten Karte', async () => {
    const waehlen = await gezeichnet('2')

    taste('PageDown')

    expect(waehlen).toHaveBeenCalledWith('3')
  })

  it('führt mit ⇞ zur vorigen Karte', async () => {
    const waehlen = await gezeichnet('2')

    taste('PageUp')

    expect(waehlen).toHaveBeenCalledWith('1')
  })

  it('blättert auch, während der Fokus in einem Feld steht', async () => {
    const waehlen = await gezeichnet('2')

    screen.getByLabelText('feld').focus()
    taste('PageDown')

    expect(waehlen).toHaveBeenCalledWith('3')
  })

  it('tut am Ende nichts — es gibt keinen Umlauf', async () => {
    const waehlen = await gezeichnet('3')

    taste('PageDown')

    // **Erst der Beleg, dass jemand zuhört** (§3): ohne ihn träfe diese
    // Verneinung auch dann zu, wenn der Horcher gar nicht hinge.
    taste('PageUp')
    expect(waehlen).toHaveBeenCalledWith('2')
    expect(waehlen).toHaveBeenCalledTimes(1)
  })

  // **Unter einer Schublade blättert nichts** (UX-08, #1680). Die
  // Kartenseite bleibt darunter stehen; ohne diesen Fall wechselte ⇟ in der
  // Konsole die Karte, die niemand sieht, verbrauchte die Taste und zöge den
  // Fokus aus der Schublade.
  it('blättert nicht, solange die Seite verdeckt ist', async () => {
    const waehlen = vi.fn()

    const { rerender } = render(
      <Verdeckung value>
        <Probe gewaehlt="2" waehlen={waehlen} />
      </Verdeckung>,
    )
    await effekteNachholen()

    taste('PageDown')
    expect(waehlen).not.toHaveBeenCalled()

    // **Der Beleg, dass jemand zuhört** (§3): aufgedeckt blättert dieselbe
    // Taste.
    rerender(
      <Verdeckung value={false}>
        <Probe gewaehlt="2" waehlen={waehlen} />
      </Verdeckung>,
    )
    await effekteNachholen()

    taste('PageDown')
    expect(waehlen).toHaveBeenCalledWith('3')
  })

  // **Aus einem Dialog und einem offenen Menü blättert er nicht** — Befund 1
  // des Zweitblicks über Phase 6 (erster Teil): die Palette ist ein Dialog,
  // und die Verdeckung kennt nur Schubladen. Ein Feld der Karte bleibt
  // ausgenommen, wie MSE2 es hält (oben).
  it('blättert nicht aus einem Dialog oder einem offenen Menü', async () => {
    const waehlen = await gezeichnet('2')

    for (const rolle of ['dialog', 'menu']) {
      const kasten = document.createElement('div')
      kasten.setAttribute('role', rolle)
      const knopf = document.createElement('button')
      kasten.append(knopf)
      document.body.append(kasten)
      knopf.dispatchEvent(
        new KeyboardEvent('keydown', { key: 'PageDown', bubbles: true }),
      )
      kasten.remove()
    }
    expect(waehlen).not.toHaveBeenCalled()

    // **Der Beleg, dass jemand zuhört** (§3).
    taste('PageDown')
    expect(waehlen).toHaveBeenCalledWith('3')
  })

  it('lässt eine schon verbrauchte Taste liegen', async () => {
    const waehlen = await gezeichnet('2')
    const verbraucht = new KeyboardEvent('keydown', {
      key: 'PageDown',
      bubbles: true,
      cancelable: true,
    })
    verbraucht.preventDefault()

    window.dispatchEvent(verbraucht)

    expect(waehlen).not.toHaveBeenCalled()
  })

  it('lässt eine Zusatztaste in Ruhe', async () => {
    const waehlen = await gezeichnet('2')

    taste('PageDown', { metaKey: true })
    taste('PageDown', { altKey: true })
    taste('PageDown', { ctrlKey: true })

    taste('PageDown')
    expect(waehlen).toHaveBeenCalledTimes(1)
  })

  it('nimmt den Fokus in dasselbe Feld der nächsten Karte mit', async () => {
    // **Der Weg, für den es dieses Werkstück gibt**: im Feld stehen, ⇟, im
    // selben Feld der nächsten Karte weiterschreiben. Gemessen an der
    // laufenden Brücke ging der Fokus vorher an den `body` verloren.
    //
    // **Die Probe baut das Feld dabei neu**, mit einem `key` an der
    // Kartenkennung — wie `Kartenarbeitsflaeche` es tut. Ohne diesen Neuaufbau
    // prüfte der Lauf gar nichts: der Fokus stünde ohnehin noch da.
    render(<Wechselprobe />)
    await effekteNachholen()

    const erstes = screen.getByLabelText('feld')
    erstes.focus()

    taste('PageDown')
    await effekteNachholen()

    const zweites = screen.getByLabelText('feld')

    expect(zweites).not.toBe(erstes)
    expect(document.activeElement).toBe(zweites)
  })

  it('wünscht sich nichts, wenn das Feld keinen Namen trägt', async () => {
    // **Ein Feld ohne `aria-label` gibt es**, und dann gibt es nichts
    // wiederzufinden. Der Wechsel muss trotzdem stattfinden — er hängt nicht
    // am Fokus.
    function OhneNamen() {
      const [gewaehlt, setzeGewaehlt] = useState('2')

      useKartenblaettern(drei, gewaehlt, setzeGewaehlt)

      return (
        <>
          <input key={gewaehlt} data-testid="namenlos" defaultValue="" />
          <output>{gewaehlt}</output>
        </>
      )
    }

    render(<OhneNamen />)
    await effekteNachholen()

    screen.getByTestId('namenlos').focus()
    taste('PageDown')
    await effekteNachholen()

    expect(screen.getByRole('status').textContent).toBe('3')
    expect(document.activeElement).toBe(document.body)
  })

  it('schreibt das Feld weg, bevor es die Karte wechselt', async () => {
    // **Sonst geht das Getippte verloren** — gemessen an der laufenden Brücke
    // am 2026-08-30:
    //
    //	A · tippen, dann blur:   ["MitBlur",""]
    //	B · tippen, dann ⇟:      ["MitBlur",""]
    //
    // Im Fall B stand »OhneBlur« im Feld und **nichts** im Bestand. Der Wert
    // wird beim Verlassen geschrieben (`Feldeingabe.tsx`, `onBlur`), und ⇟
    // verließ das Feld nie: der Editor wurde mitsamt dem Eingabefeld
    // abgeräumt, und mit ihm die Eingabe.
    //
    // **Die Reihenfolge ist die Zusage**: erst weg mit dem Wert, dann die
    // neue Karte.
    const reihenfolge: string[] = []
    const waehlen = vi.fn(() => {
      reihenfolge.push('gewaehlt')
    })

    render(<Probe gewaehlt="2" waehlen={waehlen} />)
    await effekteNachholen()

    const feld = screen.getByLabelText('feld')
    feld.addEventListener('blur', () => {
      reihenfolge.push('verlassen')
    })
    feld.focus()

    taste('PageDown')

    expect(reihenfolge).toEqual(['verlassen', 'gewaehlt'])
  })

  it('hört nach dem Abräumen auf zu hören', async () => {
    const waehlen = vi.fn()
    const { unmount } = render(<Probe gewaehlt="2" waehlen={waehlen} />)

    await effekteNachholen()
    unmount()

    taste('PageDown')

    expect(waehlen).not.toHaveBeenCalled()
  })
})
