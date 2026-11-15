// **Laschen für eine Seite mit mehreren Gebieten** — Werkstück I von Phase
// 4.10 (#1325).
//
// # Der Befund
//
// Die Deckseite legte vier Blöcke untereinander — Anlegen, Liste, das offene
// Deck, Formate —, und im offenen Deck noch einmal fünf. Am 2026-09-03 im
// Fenster nachgesehen: 1.614 Punkte Höhe, und wer ein Format anlegen will,
// scrollt an allem vorbei. Der Setarbeitsplatz löst dasselbe Problem seit
// Phase 3 mit einer Laschenleiste; sie stand nur nicht zur Verfügung.
//
// # Was hier geprüft wird
//
// Die Leiste selbst: dass sie das ARIA-Muster für Laschen einhält, dass die
// Pfeiltasten wandern und dass immer genau ein Gebiet sichtbar ist.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Laschen } from './Laschen'
import { Sprachraum } from '../sprache/kontext'

// zeige zeichnet drei Gebiete.
function zeige(oeffne?: string) {
  return render(
    <Sprachraum sprache="de">
      <Laschen
        bezeichnung="Gebiete"
        oeffne={oeffne}
        gebiete={[
          { schluessel: 'a', name: 'Erstes', inhalt: <p>Inhalt A</p> },
          { schluessel: 'b', name: 'Zweites', inhalt: <p>Inhalt B</p> },
          { schluessel: 'c', name: 'Drittes', inhalt: <p>Inhalt C</p> },
        ]}
      />
    </Sprachraum>,
  )
}

describe('die Laschen einer Seite', () => {
  it('zeigt das erste Gebiet und keines der anderen', () => {
    zeige()

    expect(screen.getByText('Inhalt A')).toBeInTheDocument()
    expect(screen.queryByText('Inhalt B')).not.toBeInTheDocument()
  })

  it('wechselt auf Klick', async () => {
    zeige()

    await userEvent.click(screen.getByRole('tab', { name: 'Zweites' }))

    expect(screen.getByText('Inhalt B')).toBeInTheDocument()
    expect(screen.queryByText('Inhalt A')).not.toBeInTheDocument()
  })

  // **Nur die gewählte Lasche ist eine Tabstelle** — das ARIA-Muster: eine
  // Leiste ist **eine** Station, zwischen den Laschen wandern die Pfeile.
  it('hält genau eine Tabstelle in der Leiste', () => {
    zeige()

    const laschen = screen.getAllByRole('tab')

    expect(laschen.filter((l) => l.tabIndex === 0)).toHaveLength(1)
  })

  it('wandert mit den Pfeiltasten und nimmt den Fokus mit', async () => {
    zeige()

    screen.getByRole('tab', { name: 'Erstes' }).focus()

    await userEvent.keyboard('{ArrowRight}')

    expect(screen.getByRole('tab', { name: 'Zweites' })).toHaveFocus()
    expect(screen.getByText('Inhalt B')).toBeInTheDocument()
  })

  // **Am Ende geht es von vorn los.** So verhält sich jede Laschenleiste, und
  // ein Pfeil, der ins Leere zeigt, wäre eine tote Taste.
  it('läuft am Ende um', async () => {
    zeige()

    screen.getByRole('tab', { name: 'Erstes' }).focus()

    await userEvent.keyboard('{ArrowLeft}')

    expect(screen.getByRole('tab', { name: 'Drittes' })).toHaveFocus()
  })

  // **Das Feld nennt seine Lasche** (`aria-labelledby`), damit eine
  // Vorlesehilfe beim Betreten sagt, wozu der Inhalt gehört.
  it('verknüpft das Feld mit seiner Lasche', () => {
    zeige()

    const feld = screen.getByRole('tabpanel')
    const lasche = screen.getByRole('tab', { name: 'Erstes' })

    expect(feld).toHaveAttribute('aria-labelledby', lasche.id)
  })
})

// **Ein Gebiet lässt sich von außen aufschlagen** — Werkstück I von Phase
// 4.10 (#1325).
//
// # Der Anlass ist eine Regression, die derselbe Umbau erzeugt hat
//
// Der Menüeintrag „Über MagicSetEditor4Mac" rollt auf der Hilfeseite zum
// Abschnitt „Über". Seit die Hilfe Laschen führt, steht der Abschnitt gar
// nicht mehr im Baum, solange sein Gebiet zu ist — der Menüeintrag führte
// ins Leere. Wer eine Seite gliedert, muss jeden Weg dorthin mitnehmen.
describe('ein Gebiet von außen', () => {
  it('schlägt das genannte Gebiet auf', () => {
    zeige('c')

    expect(screen.getByText('Inhalt C')).toBeInTheDocument()
  })

  // **Und ein Wechsel von außen schlägt erneut auf**, auch wenn jemand
  // zwischendurch von Hand woanders war: der Menüeintrag soll jedes Mal
  // wirken und nicht nur beim ersten Mal.
  it('schlägt erneut auf, wenn der Wunsch sich ändert', async () => {
    const { rerender } = zeige('c')

    await userEvent.click(screen.getByRole('tab', { name: 'Erstes' }))

    expect(screen.getByText('Inhalt A')).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Laschen
          bezeichnung="Gebiete"
          oeffne="c#2"
          gebiete={[
            { schluessel: 'a', name: 'Erstes', inhalt: <p>Inhalt A</p> },
            { schluessel: 'b', name: 'Zweites', inhalt: <p>Inhalt B</p> },
            { schluessel: 'c', name: 'Drittes', inhalt: <p>Inhalt C</p> },
          ]}
        />
      </Sprachraum>,
    )

    expect(screen.getByText('Inhalt C')).toBeInTheDocument()
  })

  // **Ein Wunsch, den es nicht gibt, ändert nichts.** Er kann nur aus einem
  // Aufrufer kommen, der einen Schlüssel falsch schreibt; das erste Gebiet
  // stehen zu lassen ist besser als eine leere Seite.
  it('lässt einen unbekannten Wunsch liegen', () => {
    zeige('zzz')

    expect(screen.getByText('Inhalt A')).toBeInTheDocument()
  })
})

// **Die Ränder der Leiste** — Werkstück I von Phase 4.10
// (#1325).
describe('die Ränder der Laschenleiste', () => {
  // **Ohne Gebiete zeichnet sie nichts.** Das kommt vor: die Einstellungsseite
  // lässt das Gebiet »Dienste« weg, wo keine Dienste sind — und eine Liste,
  // die dabei leer wird, darf keine leere Leiste hinterlassen.
  it('zeichnet nichts, wo es kein Gebiet gibt', () => {
    render(
      <Sprachraum sprache="de">
        <Laschen bezeichnung="Gebiete" gebiete={[]} />
      </Sprachraum>,
    )

    expect(screen.queryByRole('tablist')).toBeNull()
  })

  // **Eine andere Taste tut nichts.** Ohne diesen Zweig verschluckte die
  // Leiste jede Taste — auch Tab, mit dem man sie verlässt.
  it('lässt eine andere Taste durch', async () => {
    zeige()

    screen.getByRole('tab', { name: 'Erstes' }).focus()

    await userEvent.keyboard('{ArrowDown}')

    expect(screen.getByText('Inhalt A')).toBeInTheDocument()
  })

  // **Verschwindet das offene Gebiet, steht das erste da.** Auf der
  // Einstellungsseite kann das Gebiet »Dienste« zwischen zwei Zeichnungen
  // wegfallen; ohne diesen Rückfall stünde die Seite leer.
  it('fällt auf das erste Gebiet zurück, wenn das offene wegfällt', async () => {
    const { rerender } = zeige()

    await userEvent.click(screen.getByRole('tab', { name: 'Drittes' }))

    expect(screen.getByText('Inhalt C')).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Laschen
          bezeichnung="Gebiete"
          gebiete={[
            { schluessel: 'a', name: 'Erstes', inhalt: <p>Inhalt A</p> },
            { schluessel: 'b', name: 'Zweites', inhalt: <p>Inhalt B</p> },
          ]}
        />
      </Sprachraum>,
    )

    expect(screen.getByText('Inhalt A')).toBeInTheDocument()
  })
})
