// **Der Fokus bleibt im selben Feld, wenn die Karte wechselt** —
// #1290.
//
// # Warum das ein eigener Haken ist
//
// Der Karteneditor wird beim Kartenwechsel **absichtlich** neu gebaut — sein
// `key` trägt die Kartenkennung und den Verlaufsstand (`Kartenarbeitsflaeche`).
// Damit ist das alte Eingabefeld weg, und der Fokus liegt danach am `body`.
//
// Gemessen am 2026-08-30 an der laufenden Brücke, vor diesem Haken: Fokus im
// Feld `name`, ⇟ gedrückt — die Karte wechselte von „1 von 107" auf „2 von
// 107", und `document.activeElement` war `null`. Der Weg „anlegen, füllen,
// weiter" endete damit nach dem ersten *weiter* an der Maus.
//
// # Warum nicht über `feldwahl`
//
// `feldwahl` (#1075) hält, **welches Feld gewählt ist** — für die
// Symbolpalette und die Wortlisten. Das überlebt den Neuaufbau, weil es eine
// Ebene höher liegt. Der **Fokus** ist etwas anderes: er ist ein Zustand des
// Browsers, kein Zustand von React, und ein neu gebautes `input` hat ihn nicht.
//
// # Warum ein MutationObserver und kein Zeitgeber
//
// Weil die Felder der neuen Karte **später** kommen: `geoeffnete` lädt, und
// bis dahin steht dort ein leeres Formular. Ein `setTimeout` müsste raten, wie
// lange — und ein Test darüber prüfte den Rechner. Der Beobachter meldet sich,
// wenn das Feld da ist, und legt sich danach hin.

import { render, screen } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { useFeldfokus } from './useFeldfokus'
import { effekteNachholen } from '../test/effekte'

function Probe({ spaeter = false }: { spaeter?: boolean }) {
  const [wunsch, wuenschen] = useState<string>()
  const [da, setzeDa] = useState(!spaeter)
  const [andres, setzeAndres] = useState(false)

  useFeldfokus(wunsch, wuenschen)

  return (
    <>
      <button
        type="button"
        onClick={() => {
          wuenschen(
            document.activeElement?.getAttribute('aria-label') ?? undefined,
          )
        }}
      >
        merken
      </button>
      <button
        type="button"
        onClick={() => {
          setzeDa(true)
        }}
      >
        nachladen
      </button>
      <button
        type="button"
        onClick={() => {
          wuenschen('name')
        }}
      >
        wünschen
      </button>
      <button
        type="button"
        onClick={() => {
          setzeAndres(true)
        }}
      >
        anderes
      </button>
      {andres ? <p>etwas anderes</p> : null}
      {da ? <input aria-label="name" defaultValue="" /> : null}
    </>
  )
}

describe('useFeldfokus', () => {
  it('holt den Fokus in das Feld mit demselben Namen zurück', async () => {
    render(<Probe />)
    await effekteNachholen()

    screen.getByLabelText('name').focus()
    screen.getByText('merken').click()
    await effekteNachholen()

    expect(document.activeElement).toBe(screen.getByLabelText('name'))
  })

  it('wartet, bis das Feld da ist', async () => {
    render(<Probe spaeter />)
    await effekteNachholen()

    // **Der Wunsch kommt hier von Hand**, denn es gibt noch kein Feld, aus
    // dem er sich lesen ließe — genau die Lage beim Kartenwechsel: der Fokus
    // stand im **alten** Feld, und das ist schon weg.
    screen.getByText('wünschen').click()
    await effekteNachholen()

    // **Noch gibt es das Feld nicht** — der Wunsch steht und wartet.
    expect(screen.queryByLabelText('name')).not.toBeInTheDocument()

    screen.getByText('nachladen').click()
    await effekteNachholen()
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(document.activeElement).toBe(screen.getByLabelText('name'))
  })

  it('bleibt liegen, solange sich etwas anderes ändert', async () => {
    // **Der Beobachter meldet jede Änderung am Baum**, nicht nur die
    // gesuchte. Ohne diesen Fall bliebe der Zweig „gemeldet, und das Feld ist
    // immer noch nicht da" unbetreten — und der ist im Betrieb der häufigste:
    // die Vorschau kommt zuerst, das Formular danach.
    render(<Probe spaeter />)
    await effekteNachholen()

    screen.getByText('wünschen').click()
    await effekteNachholen()

    screen.getByText('anderes').click()
    await effekteNachholen()
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(screen.getByText('etwas anderes')).toBeInTheDocument()
    expect(document.activeElement).toBe(document.body)

    screen.getByText('nachladen').click()
    await effekteNachholen()
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(document.activeElement).toBe(screen.getByLabelText('name'))
  })

  it('nimmt einen Fokus nicht weg, den jemand inzwischen gesetzt hat', async () => {
    // **Befund 6 des Zweitblicks über Phase 4.3.** Wer während des Ladens
    // weiterklickt, bekam den Fokus rückwirkend weggenommen — eine alte
    // Absicht schlug eine neue.
    render(<Probe spaeter />)
    await effekteNachholen()

    screen.getByText('wünschen').click()
    await effekteNachholen()

    // Jemand setzt den Fokus selbst — und danach kommt das Feld.
    screen.getByText('nachladen').focus()
    screen.getByText('nachladen').click()
    await effekteNachholen()
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(screen.getByLabelText('name')).toBeInTheDocument()
    expect(document.activeElement).toBe(screen.getByText('nachladen'))
  })

  it('tut ohne Wunsch nichts', async () => {
    render(<Probe />)
    await effekteNachholen()

    expect(document.activeElement).toBe(document.body)
  })
})
