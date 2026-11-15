// **Die gegliederte Laschenleiste** — Werkstück 1 von Phase 3.5
// (#1166).
//
// **Was hier steht und was in `laschengruppen.test.ts`**: dort die Einteilung,
// hier ihre Wirkung im Fenster. Eine Gruppierung, die nur in einer Konstante
// steht, ist keine.
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { Setarbeitsplatz } from './Setarbeitsplatz'

const texte = kataloge.de

function zeige() {
  belegeApi({
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: undefined }),
  })

  return render(
    <Sprachraum sprache="de">
      <Setarbeitsplatz
        setID="7"
        aktuell="karten"
        wechseln={vi.fn()}
        verlaufsstand={0}
        oeffneKarte={vi.fn()}
      >
        {() => <p>{texte.seiten.karten.titel}</p>}
      </Setarbeitsplatz>
    </Sprachraum>,
  )
}

describe('die Laschenleiste ist gegliedert', () => {
  // **Die Gruppe erreicht das Vorleseprogramm an ihrer ersten Lasche** —
  // #1991.
  //
  // Bis hierher stand: die Aufschriften sind lesbar (Befund 4 des Zweitblicks
  // über Phase 3.5). **Gemessen am 2026-10-01 in der Abnahme-VM** las
  // VoiceOver damit jede Lasche der ersten Gruppe mit der Aufschrift der
  // **zweiten** — „Cards OUTPUT AND PRINT" —, mit dem gebauten CSS und nur
  // mit ihm (ohne `text-xs` las es richtig: VoiceOver hängt Text nach der
  // Lage auf dem Bildschirm an). Jetzt ist die Aufschrift für die
  // Vorlesehilfe verborgen, und die **erste Lasche jeder Gruppe** nennt sie
  // als Beschreibung: „Cards BUILD CARDS, selected, tab, 1 of 6" — gemessen
  // mit derselben Probe. Die Gliederung bleibt hörbar, an genau einer Stelle
  // je Gruppe.
  it('nennt jede Gruppe an ihrer ersten Lasche und nur dort', () => {
    zeige()

    const laschen = screen.getAllByRole('tab')
    const gruppen = Object.values(texte.allgemein.arbeitsplatz.gruppen)

    // **Eine Gruppe seit UXR-09** (#1917): die Gruppe „Ausgabe und Druck"
    // ist in den Ausgabeassistenten gewandert.
    expect(gruppen).toHaveLength(1)
    expect(laschen[0]).toHaveAccessibleDescription(gruppen[0])

    for (const stelle of [1, 2]) {
      expect(laschen[stelle]).not.toHaveAttribute('aria-describedby')
    }

    for (const name of gruppen) {
      expect(
        screen.getByText(name).closest('[aria-hidden="true"]'),
      ).not.toBeNull()
    }
  })

  // **Die Aufschrift steht neben ihren Laschen, nicht darüber** — UXR-02
  // (#1910).
  // Über den Laschen nahm sie eine eigene Zeile auf jeder Seite des Sets,
  // gerechnet 20 Punkte; im Mindestfenster stand das Zentrum der Kartenseite
  // damit bei 265 Punkten Höhe. Das Artboard *Karten* zeichnet über dem
  // Arbeitsplatz gar keine Laschenzeile. **Was bleibt, ist die Gliederung**:
  // Aufschrift, verborgen für die Vorlesehilfe, und die Beschreibung an der
  // ersten Lasche (#1991) — nur die Lage ändert sich.
  it('stellt die Aufschrift der Gruppe neben ihre Laschen', () => {
    zeige()

    const name = texte.allgemein.arbeitsplatz.gruppen.bauen
    const gruppe = screen.getByText(name).parentElement

    expect(gruppe).toHaveAttribute('role', 'presentation')
    expect(gruppe).not.toHaveClass('flex-col')
    expect(gruppe).toHaveClass('items-center')
    expect(screen.getByText(name)).not.toHaveClass('pt-1')
  })

  // **Und die Leiste bleibt eine Laschenleiste.**
  //
  // Die Gliederung ist eine Sehhilfe; das Bedienmuster darf sie nicht kosten.
  // Alle Laschen in **einer** `tablist`, mit den Pfeiltasten erreichbar — das
  // ist es, was ein Vorleseprogramm ankündigt.
  //
  // **Drei seit dem Ausgabeassistenten** (UXR-09, #1917), sechs seit den
  // Schubladen (UX-08, #1680), dreizehn seit dem Ersetzen (#1168). Die Zahl
  // steht hier und nicht als Länge einer Liste: eine Erwartung, die sich mit
  // der Quelle mitbewegt, prüft nichts.
  it('hält alle drei Laschen in einer Leiste', () => {
    zeige()

    const leiste = screen.getByRole('tablist')

    expect(screen.getAllByRole('tab')).toHaveLength(3)
    expect(leiste).toContainElement(
      screen.getByRole('tab', { name: texte.seiten.ersetzen.titel }),
    )
  })

  // **Und die Ausgabe steht im Kopf, nicht in der Leiste.**
  //
  // Hier stand „setzt die Ausgabe hinter das Bauen" — die Abnahme von #1166,
  // an der Reihenfolge der Laschen. Seit UXR-09 (#1917) öffnet *Ausgeben* im
  // Kopf den Assistenten, und *Bilder*, *Ausgeben* und *Druckbogen* sind
  // keine Laschen mehr: ein Knopf neben einer Lasche, die dasselbe öffnet,
  // wäre derselbe Weg zweimal.
  it('stellt die Ausgabe in den Kopf und nicht in die Leiste', async () => {
    zeige()

    const laschen = screen.getAllByRole('tab').map((knopf) => knopf.textContent)

    expect(laschen.indexOf(texte.seiten.karten.titel)).toBe(0)
    for (const name of [
      texte.seiten.bilder.titel,
      texte.seiten.ausgeben.titel,
      texte.seiten.druckbogen.titel,
      texte.seiten.konsole.titel,
    ]) {
      expect(laschen).not.toContain(name)
    }
    expect(
      await screen.findByRole('button', { name: texte.seiten.ausgeben.titel }),
    ).toBeInTheDocument()
  })
})
