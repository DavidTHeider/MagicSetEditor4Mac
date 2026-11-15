// **Ein gedrückter Umschalter sieht gedrückt aus** — gefunden bei der
// Sichtprüfung des Übersetzungseditors am 2026-09-30: von „Fehlend",
// „Veraltet", „Übersetzt" und „Alle" war nicht zu sehen, welcher gilt. Nur
// `aria-pressed` wechselte; dieselbe Lücke hatten die Zoomleiste, der
// Paketfilter und jede andere Gruppe aus `gedrueckt`.
//
// **Und nicht wie der Hauptknopf** — AP-3
// (#2360),
// Befund B-08: damals bekam er dessen Füllung, und bis zu vier gefüllte
// Flächen standen in einer Sicht. „Kacheln“ gedrückt sah aus wie „+ Neues
// Set“. Gewählt ist jetzt ein Segment mit Rand, wie in `Segmente`
// (Entwurf `Zustaende09`).
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Knopf } from './Knopf'

it('zeichnet den gedrückten Umschalter als Segment und nicht wie den Hauptknopf', () => {
  render(
    <>
      <Knopf gedrueckt>an</Knopf>
      <Knopf gedrueckt={false}>aus</Knopf>
      <Knopf gewicht="haupt">Anlegen</Knopf>
    </>,
  )

  const an = screen.getByRole('button', { name: 'an', pressed: true })
  const aus = screen.getByRole('button', { name: 'aus', pressed: false })
  const haupt = screen.getByRole('button', { name: 'Anlegen' })

  expect(an).toHaveClass('border-marke', 'bg-betont', 'text-text')
  expect(an).not.toHaveClass('bg-marke')
  expect(haupt).toHaveClass('bg-marke', 'text-marketext')
  expect(haupt).not.toHaveClass('bg-betont')
  expect(aus).not.toHaveClass('bg-marke')
  expect(aus).not.toHaveClass('border-marke')
})

it('gibt einem Umschalter nie das Hauptgewicht, auch wenn der Aufrufer es verlangt', () => {
  // **Ein ruhender Umschalter ist kein Hauptknopf**: wer `gewicht="haupt"`
  // und `gedrueckt` zugleich setzt, bekäme sonst im ungedrückten Stand die
  // Füllung — und der gedrückte sähe schwächer aus als der ungedrückte.
  render(
    <>
      <Knopf gewicht="haupt" gedrueckt={false}>
        Tiefe 1
      </Knopf>
      <Knopf gewicht="haupt" gedrueckt>
        Tiefe 2
      </Knopf>
    </>,
  )

  expect(screen.getByRole('button', { name: 'Tiefe 1' })).not.toHaveClass(
    'bg-marke',
  )
  expect(screen.getByRole('button', { name: 'Tiefe 2' })).not.toHaveClass(
    'bg-marke',
  )
  expect(screen.getByRole('button', { name: 'Tiefe 2' })).toHaveClass(
    'border-marke',
  )
})

// **Titel und Beschreibung ohne Tooltip** — UXR-02
// (#1910):
// der Knopf über einer Spalte des Arbeitsplatzes nennt sein Kürzel im Titel,
// wie der Knopf der Seitenleiste, und bei Enge beschreibt ihn ein Satz, der
// nicht in der Zeile steht. Ein `Tooltip` verbrauchte Escape, solange er
// unter dem ruhenden Zeiger stand.
it('trägt einen Titel und eine Beschreibung, wenn man sie ihm gibt', () => {
  render(
    <>
      <Knopf titel="Navigator ausblenden (⌥⌘[)" beschriebenVon="satz">
        Navigator ausblenden
      </Knopf>
      <p id="satz" hidden>
        Zu schmal
      </p>
      <Knopf>ohne</Knopf>
    </>,
  )

  const mit = screen.getByRole('button', { name: 'Navigator ausblenden' })
  expect(mit).toHaveAttribute('title', 'Navigator ausblenden (⌥⌘[)')
  expect(mit).toHaveAccessibleDescription('Zu schmal')

  const ohne = screen.getByRole('button', { name: 'ohne' })
  expect(ohne).not.toHaveAttribute('title')
  expect(ohne).not.toHaveAttribute('aria-describedby')
})

// **Rückmeldung bei Zeiger und Druck** — AP-4
// (#2361),
// Befund B-11, Entwurf `Zustaende09`: Hover hebt den Grund, Drücken senkt den
// Knopf um einen Punkt, in der kurzen Dauer aus Vermerk 0065 (Festlegung 10).
// **Nicht an einem gesperrten Knopf** — er antwortet nur mit seiner
// Deckkraft. Ob es am Fenster so aussieht, misst `ap4zustaende`; hier steht,
// dass die Klassen dastehen.
describe('Rückmeldung', () => {
  it('trägt Übergang, Hover und Druck in jedem Gewicht', () => {
    render(
      <>
        <Knopf>Neben</Knopf>
        <Knopf gewicht="haupt">Haupt</Knopf>
        <Knopf gewicht="leise">Leise</Knopf>
        <Knopf gedrueckt>Gewählt</Knopf>
      </>,
    )

    for (const name of ['Neben', 'Haupt', 'Leise', 'Gewählt']) {
      const knopf = screen.getByRole('button', { name })
      expect(knopf).toHaveClass(
        'transition-[color,background-color,border-color,opacity,translate]',
        'duration-(--dauer-kurz)',
        'enabled:not-aria-disabled:active:translate-y-px',
      )
    }

    expect(screen.getByRole('button', { name: 'Neben' })).toHaveClass(
      'enabled:not-aria-disabled:hover:bg-betont',
    )
    expect(screen.getByRole('button', { name: 'Leise' })).toHaveClass(
      'enabled:not-aria-disabled:hover:bg-betont',
    )
    expect(screen.getByRole('button', { name: 'Haupt' })).toHaveClass(
      'enabled:not-aria-disabled:hover:bg-marke/90',
    )
  })
})
