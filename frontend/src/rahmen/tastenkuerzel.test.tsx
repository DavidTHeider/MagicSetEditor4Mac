// Die Tastenkürzel für Rückgängig — ⌘Z und ⇧⌘Z (#52).
//
// **Eigene Datei**, weil `Rahmen.test.tsx` über 400 Zeilen wuchs (AGENTS.md
// §2). Der Schnitt folgt der Frage: dort Navigation und Palette, hier die
// Rücknahme und die eine Stelle, an der sie **nicht** gilt.
import { fireEvent, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Ruecknahmeraum, useRuecknahmeweg } from '../ruecknahme'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'

// zeige stellt den Rahmen um einen Inhalt.
function zeige(
  eigenschaften: Partial<Parameters<typeof Rahmen>[0]> = {},
  inhalt: ReactNode = <p>{kataloge.de.seiten.start.was}</p>,
) {
  render(
    <Sprachraum sprache="de">
      <Rahmen aktuell="start" wechseln={vi.fn()} {...eigenschaften}>
        {inhalt}
      </Rahmen>
    </Sprachraum>,
  )
}

describe('Rückgängig über Tastenkürzel', () => {
  // ⌘Z und ⇧⌘Z (#52).
  it('nimmt mit ⌘Z zurück und stellt mit ⇧⌘Z wieder her', () => {
    const zuruecknehmen = vi.fn()
    const wiederherstellen = vi.fn()

    zeige({ zuruecknehmen, wiederherstellen })

    fireEvent.keyDown(window, { key: 'z', metaKey: true })
    expect(zuruecknehmen).toHaveBeenCalledTimes(1)
    expect(wiederherstellen).not.toHaveBeenCalled()

    fireEvent.keyDown(window, { key: 'z', metaKey: true, shiftKey: true })
    expect(wiederherstellen).toHaveBeenCalledTimes(1)
    expect(zuruecknehmen).toHaveBeenCalledTimes(1)
  })

  // **Die Taste nimmt denselben Weg wie der Menüeintrag** — P8 (#2376). Bis
  // dahin rief ⌘Z den Set-Verlauf auch dann, wenn ein Tisch oder eine offene
  // Zeichnung sich als Ziel gemeldet hatte; nur der Menüeintrag fragte nach.
  it('trifft mit ⌘Z und ⇧⌘Z den gemeldeten Weg vor dem Set', () => {
    const zuruecknehmen = vi.fn()
    const wiederherstellen = vi.fn()
    const zurueck = vi.fn()
    const vor = vi.fn()

    function Tisch() {
      useRuecknahmeweg({ zuruecknehmen: zurueck, wiederherstellen: vor })

      return <p>tisch</p>
    }

    render(
      <Sprachraum sprache="de">
        <Ruecknahmeraum>
          <Rahmen
            aktuell="spielen"
            wechseln={vi.fn()}
            zuruecknehmen={zuruecknehmen}
            wiederherstellen={wiederherstellen}
          >
            <Tisch />
          </Rahmen>
        </Ruecknahmeraum>
      </Sprachraum>,
    )

    fireEvent.keyDown(window, { key: 'z', metaKey: true })
    fireEvent.keyDown(window, { key: 'z', metaKey: true, shiftKey: true })

    expect(zurueck).toHaveBeenCalledTimes(1)
    expect(vor).toHaveBeenCalledTimes(1)
    expect(zuruecknehmen).not.toHaveBeenCalled()
    expect(wiederherstellen).not.toHaveBeenCalled()
  })

  // **Ohne Angebot geschieht nichts** — und nichts bricht.
  //
  // Der Rahmen steht auch dort, wo es kein Set gibt; ein Aufruf ins Leere wäre
  // ein Absturz für eine Taste, die jemand aus Gewohnheit drückt.
  it('erträgt ⌘Z ohne Rückrufe', () => {
    zeige()

    expect(() => {
      fireEvent.keyDown(window, { key: 'z', metaKey: true })
      fireEvent.keyDown(window, { key: 'z', metaKey: true, shiftKey: true })
    }).not.toThrow()
  })

  // **Im Textfeld gehört ⌘Z dem Textfeld.**
  //
  // Geschrieben wird beim Verlassen des Feldes; was jemand gerade tippt, steht
  // in keinem Verlauf. Eine Rücknahme träfe hier etwas anderes als das, was die
  // tippende Person vor Augen hat.
  it('überlässt ⌘Z im Eingabefeld dem Feld', () => {
    const zuruecknehmen = vi.fn()

    zeige({ zuruecknehmen }, <input aria-label="feld" />)

    fireEvent.keyDown(screen.getByLabelText('feld'), {
      key: 'z',
      metaKey: true,
    })

    expect(zuruecknehmen).not.toHaveBeenCalled()
  })

  // **Und in jedem Feld, das selbst tippt.** Der erste Anlauf fragte nur nach
  // `INPUT`; ein mehrzeiliges Feld und ein bearbeitbarer Bereich sind dieselbe
  // Lage und wären durchgefallen.
  it.each([
    ['textarea', <textarea key="t" aria-label="feld" />],
    ['contentEditable', <div key="c" contentEditable aria-label="feld" />],
  ])('überlässt ⌘Z auch %s dem Feld', (_name, feld) => {
    const zuruecknehmen = vi.fn()

    zeige({ zuruecknehmen }, feld)

    fireEvent.keyDown(screen.getByLabelText('feld'), {
      key: 'z',
      metaKey: true,
    })

    expect(zuruecknehmen).not.toHaveBeenCalled()
  })
})

describe('Sichern über Tastenkürzel', () => {
  // ⌘S und ⇧⌘S (#162).
  it('sichert mit ⌘S und sichert mit ⇧⌘S unter', () => {
    const sichern = vi.fn()
    const sichernUnter = vi.fn()

    zeige({ sichern, sichernUnter })

    fireEvent.keyDown(window, { key: 's', metaKey: true })
    expect(sichern).toHaveBeenCalledTimes(1)
    expect(sichernUnter).not.toHaveBeenCalled()

    fireEvent.keyDown(window, { key: 's', metaKey: true, shiftKey: true })
    expect(sichernUnter).toHaveBeenCalledTimes(1)
    expect(sichern).toHaveBeenCalledTimes(1)
  })

  // **Ohne Angebot geschieht nichts** — und nichts bricht. Die Taste bleibt
  // trotzdem abgefangen: ⌘S gehört ab hier der Anwendung, und der Browser böte
  // „Seite sichern" an.
  it('kommt ohne Angebot aus', () => {
    zeige()

    expect(() => {
      fireEvent.keyDown(window, { key: 's', metaKey: true })
    }).not.toThrow()
  })

  // **Das Meldungsband bleibt bedienbar, auch ohne Rückruf** (#180): der
  // Schließknopf tut dann nichts, statt einen Fehler zu werfen.
  it('lässt das Band auch ohne Schließrückruf schließen', () => {
    zeige({ meldung: { text: 'Etwas ging schief', art: 'schlecht' } })

    const knopf = screen.getByRole('button', {
      name: kataloge.de.bausteine.aktionen.schliessen,
    })

    expect(() => {
      fireEvent.click(knopf)
    }).not.toThrow()
  })
})

// **⌘[ und ⌘] gehen den Seitengang** — UX-03.AK-02 (#1675).
describe('Zurück und Vorwärts über Tastenkürzel', () => {
  it('geht mit ⌘[ zurück und mit ⌘] vor', () => {
    const seiteZurueck = vi.fn()
    const seiteVor = vi.fn()

    zeige({ seiteZurueck, seiteVor })

    fireEvent.keyDown(window, { key: '[', code: 'BracketLeft', metaKey: true })
    expect(seiteZurueck).toHaveBeenCalledTimes(1)
    expect(seiteVor).not.toHaveBeenCalled()

    fireEvent.keyDown(window, { key: ']', code: 'BracketRight', metaKey: true })
    expect(seiteVor).toHaveBeenCalledTimes(1)
  })

  it('erträgt ⌘[ ohne Rückrufe', () => {
    zeige()

    expect(() => {
      fireEvent.keyDown(window, {
        key: '[',
        code: 'BracketLeft',
        metaKey: true,
      })
    }).not.toThrow()
  })
})
