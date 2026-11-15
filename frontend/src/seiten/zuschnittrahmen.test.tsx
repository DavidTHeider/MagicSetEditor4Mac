// **Der Rahmen rechnet den Maßstab, und sonst nichts** —
// #576.
//
// Wohin ein Zug führt, entscheidet Go (`ImageSliceSelector::onMotion`,
// `src/gui/image_slice_window.cpp`, Zeilen 609–646, Fassung f9d9356 vom
// 2021-07-06) — geprüft wird dort. Hier zählt die eine Umrechnung, die das
// Fenster beisteuert: **Bildschirmpunkte in Quellpunkte**.
//
// Sie ist der Grund, warum der Rahmen bei jeder Fensterbreite stimmt: das Bild
// ist über `max-w-full` schmaler als seine Punktzahl, sobald das Fenster es
// ist, und ein fester Maßstab wäre bei jeder anderen Breite falsch.
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { zuschnittantwort } from '../test/zuschnitt'
import { Zuschnittrahmen } from './Zuschnittrahmen'

beforeAll(() => {
  // **jsdom kennt die Zeigererfassung nicht.** Sie hält den Zug auch dann
  // beim Rahmen, wenn der Zeiger ihn verlässt — im Browser unverzichtbar,
  // hier ohne Wirkung.
  Element.prototype.setPointerCapture = vi.fn()
})

afterEach(() => {
  vi.restoreAllMocks()
})

function zeige(breite: number, frei = false) {
  const ziehen = vi.fn()

  // **Die gezeichnete Breite kommt aus dem Element**, und unter jsdom hat
  // keines eine. Hier wird sie gesetzt, damit die Umrechnung etwas zu rechnen
  // hat.
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    width: breite,
    height: breite / 4,
    x: 0,
    y: 0,
    top: 0,
    left: 0,
    right: breite,
    bottom: breite / 4,
    toJSON: () => ({}),
  })

  render(
    <Sprachraum sprache="de">
      <Zuschnittrahmen
        zuschnitt={zuschnittantwort()}
        auswahl={{ x: 0, y: 0, breite: 125, hoehe: 100 }}
        frei={frei}
        ziehen={ziehen}
      />
    </Sprachraum>,
  )

  return ziehen
}

// griffe gibt die Griffe einer Art — **über ihre Marke und nicht über eine
// Rolle**: sie sind Zeigerhilfen und vor Hilfsmitteln verborgen (G17).
function griffe(art: 'rahmen' | 'ecke'): HTMLElement[] {
  return [...document.querySelectorAll<HTMLElement>(`[data-griff="${art}"]`)]
}

// ziehe führt einen Zug von (0,0) nach (x,y) auf einem Griff aus.
function ziehe(art: 'rahmen' | 'ecke', stelle: number, x: number, y: number) {
  const griff = griffe(art)[stelle]

  if (griff === undefined) {
    throw new Error('kein Griff')
  }

  fireEvent.pointerDown(griff, { clientX: 0, clientY: 0, pointerId: 1 })
  fireEvent.pointerMove(griff, { clientX: x, clientY: y, pointerId: 1 })
  fireEvent.pointerUp(griff, { pointerId: 1 })
}

describe('Der Zuschnittrahmen', () => {
  // **Kein Knopf, der nichts tut** — G17 aus dem Zweitblick über Phase 6.3:
  // die Griffe trugen `role="button"` und taten auf ein Aktivieren nichts;
  // ein Vorleser bot sie als Knöpfe an. Der Weg ohne Zeiger sind die vier
  // Zahlenfelder des Zuschnittfensters.
  it('bietet Hilfsmitteln keine Griffe als Knöpfe an', () => {
    zeige(400, true)

    expect(screen.queryAllByRole('button')).toHaveLength(0)
    for (const griff of [...griffe('rahmen'), ...griffe('ecke')]) {
      expect(griff.closest('[aria-hidden="true"]')).not.toBeNull()
    }
  })

  // **Das Bild ist 400 Punkte breit und wird auf 200 gezeigt** — ein
  // Bildschirmpunkt sind zwei Quellpunkte.
  it('rechnet Bildschirmpunkte in Quellpunkte um', () => {
    const ziehen = zeige(200)

    ziehe('rahmen', 0, 30, 10)

    expect(ziehen).toHaveBeenCalledWith(60, 20, 0, 0)
  })

  // **Ein Griff nennt seine Ecke** — MSEs `dragX`/`dragY`.
  it('nennt die gezogene Ecke', () => {
    const ziehen = zeige(400)

    ziehe('ecke', 0, 10, 10)

    expect(ziehen).toHaveBeenCalledWith(10, 10, -1, -1)
  })

  // **Vier Ecken, solange das Seitenverhältnis steht** — MSE gibt eine
  // Kantenmitte dann als Griff zurück (Zeile 596). Ohne Bindung kommen sie
  // dazu.
  it('zeigt die Kantengriffe nur ohne festes Seitenverhältnis', () => {
    zeige(400)
    expect(griffe('ecke')).toHaveLength(4)
  })

  it('zeigt sechs Griffe ohne Bindung', () => {
    const ziehen = zeige(400, true)

    expect(griffe('ecke')).toHaveLength(6)

    ziehe('ecke', 4, 20, 0)

    expect(ziehen).toHaveBeenCalledWith(20, 0, 1, 0)
  })

  // **Ein Zug, bevor das Bild gezeichnet ist**, rechnet mit eins — und die
  // nächste Bewegung rechnet richtig. Ein Sturz wäre die Alternative.
  it('kommt ohne gezeichnete Breite aus', () => {
    const ziehen = zeige(0)

    ziehe('rahmen', 0, 30, 10)

    expect(ziehen).toHaveBeenCalledWith(30, 10, 0, 0)
  })
})
