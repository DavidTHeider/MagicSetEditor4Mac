// **Was das Sichtfenster rechnet**
// (#891).
//
// Unter jsdom hat kein Element eine Höhe — deshalb setzt dieser Test sie
// ausdrücklich. Ohne das prüfte er den einen Zweig, den jede andere Datei
// ohnehin nimmt: „ohne Maß wird alles gezeichnet".
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useSichtfenster } from './useSichtfenster'

// hoeheSetzen gibt einem Element eine gemessene Höhe.
//
// **`clientHeight` und `getBoundingClientRect` getrennt**, weil jsdom beide
// getrennt führt und der Haken beide liest: die Sichthöhe aus dem einen, die
// Zeilenhöhe aus dem anderen.
function hoeheSetzen(knoten: HTMLElement, hoehe: number, rollen = 0) {
  Object.defineProperty(knoten, 'clientHeight', {
    configurable: true,
    value: hoehe,
  })
  Object.defineProperty(knoten, 'scrollTop', {
    configurable: true,
    writable: true,
    value: rollen,
  })
  knoten.getBoundingClientRect = () => ({ height: hoehe }) as DOMRect
}

interface Probe {
  anzahl: number
  sicht: number
  zeile: number
  rollen: number
}

function Liste({ anzahl, sicht, zeile, rollen }: Probe) {
  const fenster = useSichtfenster(anzahl)

  return (
    <ul
      ref={(knoten) => {
        if (knoten !== null) {
          hoeheSetzen(knoten, sicht, rollen)
        }

        fenster.behaelter(knoten)
      }}
    >
      <li data-oben={fenster.obenHoch} data-unten={fenster.untenHoch}>
        {`${String(fenster.erste)}–${String(fenster.letzte)}`}
      </li>
      <li
        ref={(knoten) => {
          if (knoten !== null) {
            hoeheSetzen(knoten, zeile)
          }

          fenster.zeilenmass(knoten)
        }}
      />
    </ul>
  )
}

describe('Das Sichtfenster', () => {
  // **Ohne Maß wird eine Probezeile gezeichnet** — und nicht die ganze Liste.
  //
  // # Was hier bis zum 2026-08-27 stand
  //
  // „Ohne Maß wird alles gezeichnet — der Weg zum ersten Maß führt darüber:
  // die erste Zeile muss dastehen, damit ihre Höhe messbar ist." Der zweite
  // Halbsatz stimmt, der erste folgte nie daraus: **eine** Zeile genügt zum
  // Messen, und 3.788 zu zeichnen, um eine zu messen, ist genau der Aufwand,
  // den dieser Haken vermeiden soll. Angestrichen als Z7-04 im Audit
  // (#1166).
  //
  // **Der Weg zum Maß bleibt derselbe**: die Probezeile trägt `zeilenmass`,
  // React setzt Kindreferenzen vor der des Behälters, und der misst beide —
  // noch vor dem Zeichnen. Ein Aufblitzen der einzelnen Zeile gibt es deshalb
  // nicht.
  it('zeichnet nur eine Probezeile, solange nichts gemessen ist', () => {
    render(<Liste anzahl={3788} sicht={0} zeile={0} rollen={0} />)

    expect(screen.getByText('0–1')).toBeInTheDocument()
  })

  // **Und bei einer leeren Liste bleibt es bei null.**
  //
  // Ohne diese Zeile stünde `Math.min` ungeprüft da, und eine Probezeile in
  // einer Liste ohne Zeilen wäre ein `slice(0, 1)` auf nichts — grün, aber die
  // Rechnung wäre falsch, sobald jemand den Wert anfasst.
  it('zeichnet nichts, wenn es nichts gibt', () => {
    render(<Liste anzahl={0} sicht={0} zeile={0} rollen={0} />)

    expect(screen.getByText('0–0')).toBeInTheDocument()
  })

  // 486 Punkte Sicht, 41 je Zeile: zwölf sichtbar, plus acht Überhang je
  // Seite — gemessen an der laufenden Anwendung am 2026-08-25.
  it('zeichnet am Anfang die sichtbaren Zeilen und den Überhang', () => {
    render(<Liste anzahl={3788} sicht={486} zeile={41} rollen={0} />)

    // ceil(486/41) = 12, erste = max(0, floor(0/41) - 8) = 0, letzte = 0+12+16
    expect(screen.getByText('0–28')).toBeInTheDocument()

    const zeile = screen.getByText('0–28')

    // **Oben nichts, unten der ganze Rest** — der Abstandhalter trägt den
    // Rollbalken: ohne ihn wäre die Liste so hoch wie ihr Fenster.
    expect(zeile).toHaveAttribute('data-oben', '0')
    expect(zeile).toHaveAttribute('data-unten', String((3788 - 28) * 41))
  })

  it('wandert mit dem Rollbalken und lässt den Überhang stehen', () => {
    render(<Liste anzahl={3788} sicht={486} zeile={41} rollen={41 * 100} />)

    // erste = 100 - 8 = 92, letzte = 92 + 12 + 16 = 120
    expect(screen.getByText('92–120')).toBeInTheDocument()
    expect(screen.getByText('92–120')).toHaveAttribute(
      'data-oben',
      String(92 * 41),
    )
  })

  // **Am Ende steht kein Abstandhalter mehr**, und `letzte` läuft nicht über
  // die Liste hinaus — ein `slice` darüber gäbe sonst leere Zeilen.
  //
  // **Und das Fenster bleibt voll.** Bei 30 Karten und einem Fenster von 28
  // beginnt es bei 2, nicht bei 12: `erste` wird an `anzahl - fenster`
  // gekappt. Vorher zeigte es am Ende weniger Zeilen, als hineinpassen.
  it('endet an der letzten Karte', () => {
    render(<Liste anzahl={30} sicht={486} zeile={41} rollen={41 * 20} />)

    expect(screen.getByText('2–30')).toBeInTheDocument()
    expect(screen.getByText('2–30')).toHaveAttribute('data-unten', '0')
    expect(screen.getByText('2–30')).toHaveAttribute(
      'data-oben',
      String(2 * 41),
    )
  })

  // **Eine Liste, die schrumpft, bleibt sichtbar** — Befund 1 des Zweitblicks
  // über Phase Y.
  //
  // `erste` kommt aus dem Rollstand, und der gehört zu der Liste von vorhin.
  // Ohne die Kappung wäre `erste = 992` bei `letzte = 107`, `slice` gäbe
  // nichts, und der Abstandhalter oben hielte die Höhe von 992 Zeilen — der
  // Rollbalken käme nie zurück.
  it('zeigt Zeilen, wenn die Liste unter den Rollstand schrumpft', () => {
    render(<Liste anzahl={107} sicht={486} zeile={41} rollen={41 * 1000} />)

    // fenster = 12 + 16 = 28, erste = min(992, 107 - 28) = 79
    expect(screen.getByText('79–107')).toBeInTheDocument()
    expect(screen.getByText('79–107')).toHaveAttribute('data-unten', '0')
  })

  // **Und eine Liste, die kürzer ist als das Fenster, beginnt bei null.**
  //
  // Ohne das `Math.max` wäre `anzahl - fenster` negativ, und ein negativer
  // Abstandhalter ist keine Höhe.
  it('beginnt bei null, wenn die Liste kürzer ist als das Fenster', () => {
    render(<Liste anzahl={6} sicht={486} zeile={41} rollen={41 * 1000} />)

    expect(screen.getByText('0–6')).toBeInTheDocument()
    expect(screen.getByText('0–6')).toHaveAttribute('data-oben', '0')
  })
})
