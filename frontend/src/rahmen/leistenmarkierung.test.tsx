// **Was die Seitenleiste zeigt und was sie hervorhebt** — UX-02
// (#1674).
//
// Aus `Rahmen.test.tsx` herausgelöst, als die Datei über 400 Zeilen wuchs
// (§2) — derselbe Schnitt wie bei `leistenkuerzel.test.tsx`. Der Schnitt
// folgt der Frage: dort steht, **was der Rahmen tut** — Palette, Suche,
// Tastatur, Sprung zum Inhalt; hier, **welchen Eintrag die Leiste führt und
// welcher leuchtet**.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { beschriftung, setschluessel } from '../navigation'
import { leisteneintraege } from '../leistenordnung'
import { leisteneintrag } from '../test/leistenweg'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'

function zeige(eigenschaften: Partial<Parameters<typeof Rahmen>[0]> = {}) {
  belegeApi({})

  render(
    <Sprachraum sprache="de">
      <Rahmen
        aktuell="start"
        wechseln={vi.fn()}
        suchen={vi.fn()}
        {...eigenschaften}
      >
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

describe('Die Seitenleiste', () => {
  it('zeigt jeden Eintrag der Leiste', async () => {
    zeige()

    const navigation = screen.getByRole('navigation')

    for (const eintrag of leisteneintraege) {
      expect(await leisteneintrag(eintrag.ziel)).toBeInTheDocument()
    }

    expect(navigation).toBeInTheDocument()
  })

  // **Und keine Setseite.** Sie stehen seit
  // #288 als
  // Laschen im Arbeitsplatz; zweimal derselbe Weg hieße, dass einer von beiden
  // nicht sagt, worüber er spricht.
  // **Und seit UX-02 auch die Karten nicht mehr.** Sie standen dort seit
  // Werkstück I von Phase 4.10
  // (#1325),
  // weil sie die Fläche sind, auf der gearbeitet wird. Jetzt führt der
  // Eintrag *Gestalten* dorthin — über die Setliste, so wie der Entwurf es
  // zeichnet.
  it('führt keine Setseite in der Leiste', () => {
    zeige()

    // **Über alle Setseiten, nicht über die Laschenleiste** — seit UX-08
    // stehen sieben davon in einer Schublade (#1680), und in der
    // Seitenleiste hat keine von ihnen etwas zu suchen.
    for (const schluessel of setschluessel) {
      expect(
        screen.queryByRole('button', {
          name: beschriftung(kataloge.de, schluessel),
        }),
      ).not.toBeInTheDocument()
    }
  })

  it('kennzeichnet den aufgeschlagenen Bereich', async () => {
    zeige({ aktuell: 'sets' })

    expect(await leisteneintrag('sets')).toHaveAttribute('aria-current', 'page')
  })

  // **Auch eine Lasche des Sets lässt *Gestalten* leuchten.** Sie steht nicht
  // in der Leiste — und ohne diesen Rückfall sagte die Leiste auf der
  // meistbenutzten Fläche der Anwendung gar nichts.
  it('kennzeichnet Gestalten, solange ein Set offen ist', async () => {
    zeige({ aktuell: 'stil' })

    expect(await leisteneintrag('sets')).toHaveAttribute('aria-current', 'page')
  })

  // **Und ein Dienst leuchtet allein.** Alle drei teilen sich einen Bereich;
  // wer dort nach dem Bereich fragte, ließe alle drei zugleich leuchten.
  it('kennzeichnet in der Dienstzone nur die eine Seite', async () => {
    zeige({ aktuell: 'hilfe' })

    expect(await leisteneintrag('hilfe')).toHaveAttribute(
      'aria-current',
      'page',
    )

    expect(await leisteneintrag('nutzer')).not.toHaveAttribute('aria-current')
  })
})
