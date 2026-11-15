// **Die Gruppenliste** — UX-04.AK-01
// (#1676).
//
// # Warum es sie gibt
//
// **Das Artboard `Einstellungen` zeichnet keine Laschen.** Links eine Spalte
// von 232 Punkten — das Suchfeld oben, darunter die neun Gruppen
// untereinander mit Name und Untertitel —, rechts die Fläche mit dem
// Gruppennamen im Kopf. Gebaut waren waagerechte Laschen, und sie brachen
// bei 960 Punkten Inhaltsbreite **auf zwei Reihen** um (Sichtprüfung vom
// 2026-09-21).
//
// # Was sie mit den Laschen teilt
//
// **Welches Gebiet offen ist und wie die Tastatur es wechselt** — das steht
// in [useGebietswahl] und ist beiden gemeinsam. Verschieden ist nur, **wie**
// es aussieht: dort eine Leiste und Pfeil links/rechts, hier eine Liste und
// Pfeil hoch/runter.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Gruppenliste } from './Gruppenliste'

const gebiete = [
  {
    schluessel: 'eins',
    name: 'Erstes',
    unter: 'was im ersten steht',
    inhalt: <p>Inhalt eins</p>,
  },
  {
    schluessel: 'zwei',
    name: 'Zweites',
    unter: 'was im zweiten steht',
    inhalt: <p>Inhalt zwei</p>,
  },
  {
    schluessel: 'drei',
    name: 'Drittes',
    unter: 'was im dritten steht',
    inhalt: <p>Inhalt drei</p>,
  },
]

function zeige(teile: Partial<Parameters<typeof Gruppenliste>[0]> = {}) {
  render(<Gruppenliste bezeichnung="Gruppen" gebiete={gebiete} {...teile} />)
}

describe('die Gruppenliste', () => {
  // **Senkrecht, und das steht am Bauteil.** Ein Vorleseprogramm sagt die
  // Richtung an und erwartet danach die Pfeiltasten; eine Liste, die sich
  // als waagerecht ausgibt, verspricht die falschen.
  it('steht senkrecht', () => {
    zeige()

    expect(screen.getByRole('tablist', { name: 'Gruppen' })).toHaveAttribute(
      'aria-orientation',
      'vertical',
    )
  })

  // **Jede Zeile trägt ihren Untertitel.** Er stand bisher im Feld, über den
  // Einstellungen — dort sieht ihn nur, wer die Gruppe schon aufgeschlagen
  // hat, und die Frage „bin ich hier richtig?" stellt sich davor.
  it('nennt zu jeder Gruppe, was darin steht', () => {
    zeige()

    for (const gebiet of gebiete) {
      const zeile = screen.getByRole('tab', { name: new RegExp(gebiet.name) })

      expect(zeile).toHaveTextContent(gebiet.unter)
    }
  })

  it('zeigt das erste Gebiet, solange niemand wählt', () => {
    zeige()

    expect(screen.getByRole('tabpanel')).toHaveTextContent('Inhalt eins')
  })

  it('öffnet die angeklickte Gruppe', async () => {
    zeige()

    await userEvent.click(screen.getByRole('tab', { name: /Zweites/ }))

    expect(screen.getByRole('tabpanel')).toHaveTextContent('Inhalt zwei')
  })

  // **Pfeil hoch und runter und nicht links und rechts** — die Richtung, die
  // das Bauteil ansagt.
  it('wandert mit den Pfeiltasten', async () => {
    zeige()

    await userEvent.click(screen.getByRole('tab', { name: /Erstes/ }))
    await userEvent.keyboard('{ArrowDown}')

    expect(screen.getByRole('tabpanel')).toHaveTextContent('Inhalt zwei')
    expect(screen.getByRole('tab', { name: /Zweites/ })).toHaveFocus()

    await userEvent.keyboard('{ArrowUp}')

    expect(screen.getByRole('tabpanel')).toHaveTextContent('Inhalt eins')
  })

  // **Am Ende geht es von vorn los.** Ein Pfeil, der ins Leere zeigt, wäre
  // eine tote Taste.
  it('läuft am Ende um', async () => {
    zeige()

    await userEvent.click(screen.getByRole('tab', { name: /Erstes/ }))
    await userEvent.keyboard('{ArrowUp}')

    expect(screen.getByRole('tabpanel')).toHaveTextContent('Inhalt drei')
  })

  // **Ein Wunsch von außen öffnet eine Gruppe** — daran hängt der Sprung aus
  // der Suche.
  it('öffnet, was von außen gewünscht wird', () => {
    zeige({ oeffne: 'drei#1' })

    expect(screen.getByRole('tabpanel')).toHaveTextContent('Inhalt drei')
  })

  // **Über der Liste ist Platz, und im Entwurf steht dort die Suche.**
  //
  // **„In derselben Spalte" und „darüber" sind zwei Aussagen**, und beide
  // stehen hier: das Feld neben der Liste statt über der Seite war der
  // Befund, die Reihenfolge ist der Entwurf.
  it('zeigt über der Liste, was ihr mitgegeben wird', () => {
    zeige({ darueber: <input aria-label="Einstellung suchen" /> })

    const liste = screen.getByRole('tablist', { name: 'Gruppen' })
    const suchfeld = screen.getByLabelText('Einstellung suchen')

    expect(liste.parentElement?.contains(suchfeld)).toBe(true)
    expect(
      suchfeld.compareDocumentPosition(liste) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeGreaterThan(0)
  })

  // **Ohne Gebiete steht nichts da** — dieselbe Zusage wie bei den Laschen:
  // eine leere Liste wäre ein Bedienelement, das nichts öffnet.
  it('zeigt ohne Gebiete nichts', () => {
    zeige({ gebiete: [] })

    expect(screen.queryByRole('tablist')).not.toBeInTheDocument()
  })

  // **Die Spalte bleibt stehen, wenn die Fläche rollt** — UXR-19
  // (#1927).
  // Am nativen Fenster beobachtet (`ops09b`, #1961): unter einer langen
  // Gruppe lag links ein leerer Streifen, und die nächste Gruppe stand ganz
  // oben. **Was der Lauf nicht hält**: dass sie wirklich haftet — jsdom
  // rechnet weder Behälterabfrage noch Rollen; das sieht die Hüllenprobe.
  // Er hält, dass die Spalte darum gebeten hat, und dass sie, höher als das
  // Fenster, selbst rollt statt unter den Rand zu laufen.
  it('haftet mit Suchfeld und Liste, solange die Fläche rollt', () => {
    zeige({ darueber: <input aria-label="Suche" /> })

    const spalte = screen.getByRole('textbox', { name: 'Suche' }).parentElement

    expect(spalte).toContainElement(
      screen.getByRole('tablist', { name: 'Gruppen' }),
    )
    expect(spalte).toHaveClass('@2xl:sticky', '@2xl:overflow-y-auto')
  })
})
