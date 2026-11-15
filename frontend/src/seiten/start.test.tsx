// Die Startseite.
//
// **Sie ist das erste, was jemand sieht** — und stand bis
// #288 als
// Überschrift mit einem Satz darunter, der den Entwicklungsstand meldete. Ein
// Satz über den Entwicklungsstand ist eine Notiz an uns, keine Auskunft an die
// Nutzenden.
//
// Geprüft wird, was sie **zusagt**: sie nennt die nächsten Schritte, und jeder
// führt dorthin, wo er hinführt.
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

describe('Start', () => {
  it('nennt die Anwendung und was sie ist', () => {
    render(
      <Sprachraum sprache="de">
        <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={0} />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('heading', { name: texte.ueberschrift, level: 1 }),
    ).toBeInTheDocument()
    expect(screen.getByText(texte.was)).toBeInTheDocument()
  })

  // **Der Überblick steht auch bei leerem Bestand** —
  // #1325.
  //
  // # Was vorher war
  //
  // Die Seite zeigte *Erste Schritte* **oder** *Dein Überblick*, je nachdem,
  // ob der Bestand ein Set führt. Weil „noch nicht geladen" wie „kein Set"
  // aussah, **sprang** sie um, sobald die Antwort eintraf — während der
  // Kommentar an genau dieser Stelle sagte, das dürfe nicht geschehen.
  //
  // **Die ersten Schritte sind jetzt ein Fenster** und stehen in
  // `bausteine/einfuehrung.test.tsx`; hier bleibt die Zusage, dass die
  // Startseite **eine** Gestalt hat.
  it('zeigt den Überblick, auch wenn noch nichts da ist', async () => {
    render(
      <Sprachraum sprache="de">
        <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={0} />
      </Sprachraum>,
    )

    // **Über die Liste und ihren Namen** (UXR-03): „Dein Überblick" ist
    // seither der Name der Kennzahlenzeile und keine Überschrift mehr.
    expect(
      await screen.findByRole('list', { name: texte.ueberblick.titel }),
    ).toBeInTheDocument()

    // **Und die Schritte stehen nicht mehr hier.**
    //
    // **Über die Überschrift und nicht über einen Knopf**: „Zu den Sets"
    // stand bis UXR-03 auch auf einer Kachel des Überblicks, und ein Test,
    // der sie traf, prüfte etwas anderes, als er sagt.
    expect(screen.queryByText(texte.ersteSchritte)).toBeNull()
  })
})
