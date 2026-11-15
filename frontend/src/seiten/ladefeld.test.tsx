// **Der gemeinsame Rahmen für Laden, Fehler und Leere.**
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Ladefeld } from './Ladefeld'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

// **Ein Laden, das dauert, sagt warum** — Werkstück C von Phase 4.10
// (#1325).
//
// # Was gemessen ist
//
// Die Lasche „Zufallspack" stand im Abzug vom 2026-09-03 bei „Wird geladen
// …", und der Befund lautete zunächst, sie hänge. **Sie hängt nicht, sie
// rechnet**: gemessen am selben Tag braucht `Packarten` beim **ersten**
// Aufruf 12,7 Sekunden, danach 29 bis 43 Millisekunden — der Speicher im
// Packbereich greift. An der Kommandozeile, ohne die Nebenlast des
// aufgehenden Fensters, sind es 1,58 Sekunden. **Welches Set, stand nicht
// dabei**; mit *Blue* (3.693 Karten) waren es am 2026-10-17 rund 100 s am
// Fenster und 80 s an der Kommandozeile, seit #2466 16,6 und 15,5 s.
//
// **Der Mangel war also nicht die Dauer, sondern das Schweigen darüber.**
// Zwölf Sekunden „Wird geladen …" ohne Grund sehen aus wie ein Fehler; §8
// verlangt, dass ein Zustand sagt, was los ist.
it('nennt beim Laden den Grund, wenn einer angegeben ist', () => {
  render(
    <Sprachraum sprache="de">
      <Ladefeld
        laedt
        fehler={undefined}
        leer={false}
        grund="Das dauert, weil …"
      >
        <p>leer</p>
      </Ladefeld>
    </Sprachraum>,
  )

  const stand = screen.getByRole('status')

  expect(stand.textContent).toContain(kataloge.de.bausteine.zustaende.laedt)
  expect(stand.textContent).toContain('Das dauert, weil …')
})

// **Ohne Grund bleibt es beim Satz.** Die meisten Läufe sind schnell, und
// eine Erklärung für etwas, das man nicht bemerkt, ist Lärm.
it('bleibt ohne Grund bei der Lademeldung', () => {
  render(
    <Sprachraum sprache="de">
      <Ladefeld laedt fehler={undefined} leer={false}>
        <p>leer</p>
      </Ladefeld>
    </Sprachraum>,
  )

  expect(screen.getByRole('status').textContent).toBe(
    kataloge.de.bausteine.zustaende.laedt,
  )
})

// **Beim Nachlesen steht der Inhalt schon da** — U11 (#2125): die Zeile
// „Wird geladen" schob ihn nach unten und gleich wieder zurück. Gesagt wird
// es jetzt nur noch dem Vorleser, an derselben Stelle.
it('schiebt beim Nachlesen keine Zeile über vorhandenen Inhalt', () => {
  render(
    <Sprachraum sprache="de">
      <Ladefeld laedt fehler={undefined} leer={false} vorhanden>
        <p>leer</p>
      </Ladefeld>
    </Sprachraum>,
  )

  const stand = screen.getByRole('status')

  expect(stand.textContent).toContain(kataloge.de.bausteine.zustaende.laedt)
  expect(stand.className).toContain('sr-only')
})
