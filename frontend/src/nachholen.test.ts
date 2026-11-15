// **Was ein verborgener Horcher verpasst hat, bekommt er nach** — M9 (b) aus
// dem Zweitblick über Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// `<Activity mode="hidden">` baut die Effekte einer verborgenen Seite ab, und
// Wails meldet nicht nach: wer am fremden Tisch saß und kurz die Seite
// wechselte, verpasste Tischbild, Züge und Gespräch.
import { afterEach, describe, expect, it } from 'vitest'

import { horcheNachholend, Nachholgrenze } from './nachholen'
import { laufzeitprobe, type Laufzeitprobe } from './test/laufzeitprobe'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
})

function mitLaufzeit(): Laufzeitprobe {
  probe = laufzeitprobe()

  return probe
}

describe('horcheNachholend', () => {
  it('reicht weiter, was kommt, solange jemand horcht', () => {
    const laufzeit = mitLaufzeit()
    const bekommen: unknown[] = []

    horcheNachholend('probe:a', 'stelle', (daten) => bekommen.push(daten[0]))
    laufzeit.melde('probe:a', 1)

    expect(bekommen).toEqual([1])
  })

  it('holt nach, was kam, während niemand an dieser Stelle horchte', () => {
    const laufzeit = mitLaufzeit()
    const zuerst: unknown[] = []

    const abmelden = horcheNachholend('probe:b', 'stelle', (daten) =>
      zuerst.push(daten[0]),
    )
    laufzeit.melde('probe:b', 1)
    abmelden()

    laufzeit.melde('probe:b', 2)
    laufzeit.melde('probe:b', 3)

    const danach: unknown[] = []
    horcheNachholend('probe:b', 'stelle', (daten) => danach.push(daten[0]))
    laufzeit.melde('probe:b', 4)

    expect(zuerst).toEqual([1])
    expect(danach).toEqual([2, 3, 4])
  })

  // **Wer zum ersten Mal horcht, fängt jetzt an** — sonst bekäme eine Seite
  // beim ersten Öffnen die Züge einer Partie, an der sie nie saß.
  it('spielt einer neuen Stelle nichts nach', () => {
    const laufzeit = mitLaufzeit()

    horcheNachholend('probe:c', 'erste', () => undefined)
    laufzeit.melde('probe:c', 1)

    const neu: unknown[] = []
    horcheNachholend('probe:c', 'zweite', (daten) => neu.push(daten[0]))

    expect(neu).toEqual([])
  })

  // **Der Puffer hat eine Grenze** — eine lange Partie ist sonst ein
  // Speicherleck mit Zügen darin.
  it('hält höchstens so viele, wie die Grenze sagt', () => {
    const laufzeit = mitLaufzeit()

    horcheNachholend('probe:d', 'stelle', () => undefined)()

    for (let nr = 0; nr < Nachholgrenze + 5; nr++) {
      laufzeit.melde('probe:d', nr)
    }

    const danach: unknown[] = []
    horcheNachholend('probe:d', 'stelle', (daten) => danach.push(daten[0]))

    expect(danach).toHaveLength(Nachholgrenze)
    expect(danach[0]).toBe(5)
  })

  // **Und sagt, wenn etwas fehlt** (Z3 aus dem Zweitblick über Phase 6.11):
  // ein Verlauf, dem vorn Züge fehlen, sähe sonst vollständig aus.
  it('meldet eine Lücke, wenn über der Grenze etwas verworfen wurde', () => {
    const laufzeit = mitLaufzeit()
    const luecken: string[] = []

    horcheNachholend('probe:h', 'stelle', () => undefined)()
    horcheNachholend('probe:h', 'knapp', () => undefined)()
    laufzeit.melde('probe:h', 0)
    horcheNachholend(
      'probe:h',
      'knapp',
      () => undefined,
      () => luecken.push('knapp'),
    )()

    for (let nr = 1; nr <= Nachholgrenze + 1; nr++) {
      laufzeit.melde('probe:h', nr)
    }

    horcheNachholend(
      'probe:h',
      'stelle',
      () => undefined,
      () => luecken.push('stelle'),
    )
    horcheNachholend(
      'probe:h',
      'neu',
      () => undefined,
      () => luecken.push('neu'),
    )

    // **„knapp“ sah bis 0, „stelle“ nichts davon** — verworfen sind 0 und 1.
    expect(luecken).toEqual(['stelle'])

    horcheNachholend(
      'probe:h',
      'knapp',
      () => undefined,
      () => luecken.push('knapp'),
    )
    expect(luecken).toEqual(['stelle', 'knapp'])
  })

  // **Eine neue Laufzeit ist ein neuer Anfang** — im Betrieb gibt es eine,
  // in den Läufen je Test eine eigene.
  it('fängt mit einer neuen Laufzeit neu an', () => {
    mitLaufzeit()
    horcheNachholend('probe:e', 'stelle', () => undefined)()
    probe?.raeume()

    const zweite = mitLaufzeit()
    const bekommen: unknown[] = []
    horcheNachholend('probe:e', 'stelle', (daten) => bekommen.push(daten[0]))
    zweite.melde('probe:e', 7)

    expect(bekommen).toEqual([7])
  })

  // **Und lässt die alte los** — sonst hielte jede abgelöste Laufzeit ihren
  // Sammler samt Puffer und Horchern fest.
  it('meldet sich bei einer abgelösten Laufzeit ab', () => {
    const erste = laufzeitprobe()
    horcheNachholend('probe:g', 'stelle', () => undefined)
    expect(erste.horchende('probe:g')).toBe(1)

    const zweite = mitLaufzeit()
    horcheNachholend('probe:g', 'stelle', () => undefined)

    expect(erste.horchende('probe:g')).toBe(0)
    expect(zweite.horchende('probe:g')).toBe(1)
  })

  it('tut ohne Laufzeit nichts', () => {
    const bekommen: unknown[] = []

    horcheNachholend('probe:f', 'stelle', (daten) => bekommen.push(daten))()

    expect(bekommen).toEqual([])
  })
})
