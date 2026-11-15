import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { Sprachraum } from './sprache/kontext'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { ausKarten } from './test/kartenproben'
import { leisteneintrag } from './test/leistenweg'

// Die erwarteten Texte kommen aus dem Katalog, nicht aus dieser Datei. Ein Test,
// der den Satz noch einmal hinschreibt, prüft, dass zwei Kopien gleich sind —
// und geht mit der ersten Umformulierung kaputt, ohne dass etwas falsch wäre.
// **Die Anwendung wählt ihre Sprache jetzt selbst** — aus dem **Bestand**
// (#245),
// sonst aus der Sprache des Systems. Unter jsdom ist die `en-US`; ein Test, der
// deutsche Sätze erwartet, muss also sagen, dass er sie erwartet.
//
// **Und er muss warten.** Gezeichnet wird erst, wenn die Sprache feststeht;
// bis dahin gibt `App` nichts zurück. Bis #245 stand die Sprache im
// `localStorage` und war da, bevor React lief.
describe('App', () => {
  beforeEach(() => {
    belegeApi({ ...belegeFenster('de') })
  })

  afterEach(() => {
    raeumeApi()
  })

  it('zeigt den Namen der Anwendung als Überschrift', async () => {
    render(<App />)

    expect(
      await screen.findByRole('heading', {
        name: kataloge.de.seiten.start.ueberschrift,
      }),
    ).toBeInTheDocument()
  })

  // **Sie sagt, was die Anwendung ist, und bietet den ersten Schritt an.**
  //
  // Hier stand bis #288
  // ein `role="status"` mit dem **Entwicklungsstand** darin — eine Notiz an
  // uns, angekündigt als Auskunft an die Nutzenden.
  it('sagt, was die Anwendung ist', async () => {
    render(<App />)

    expect(
      await screen.findByText(kataloge.de.seiten.start.was),
    ).toBeInTheDocument()
    // **`getAllBy`, weil der Knopf zweimal dasteht.** Die Startseite bietet
    // den ersten Schritt an, und das Einführungsfenster darüber tut es auch —
    // wer den Text sucht, findet ihn zweimal, sobald die Einführung geladen
    // ist. Ein `getBy` war damit vom Zeitpunkt abhängig, und das ist genau
    // das Flackern, vor dem §3 warnt.
    expect(
      screen.getAllByRole('button', {
        name: kataloge.de.seiten.start.schritte.setOeffnen.handlung,
      }).length,
    ).toBeGreaterThan(0)
  })

  it('folgt der gewählten Sprache', async () => {
    belegeApi({ ...belegeFenster('en') })

    render(<App />)

    expect(
      await screen.findByText(kataloge.en.seiten.start.was),
    ).toBeInTheDocument()
  })

  // Ohne Wahl gilt die Sprache des Systems — unter jsdom Englisch. **Der
  // Bestand gibt für die Sprache keinen Vorgabewert her**: was ohne Wahl gilt,
  // weiß nur das Fenster.
  it('folgt sonst der Sprache des Systems', async () => {
    belegeApi({ ...belegeFenster('') })

    render(<App />)

    expect(
      await screen.findByText(kataloge.en.seiten.start.was),
    ).toBeInTheDocument()
  })

  // **Und wenn der Bestand schweigt, auch.** Ein Fehler beim Laden zählt als
  // „geladen": dann gelten die Vorgabewerte und die Systemsprache — genau das,
  // was ein frischer Rechner zeigt. Warten hilft dann nicht mehr, und eine
  // Anwendung, die auf eine Antwort wartet, die nicht kommt, ist stumm.
  it('zeichnet auch, wenn der Bestand die Einstellungen nicht hergibt', async () => {
    belegeApi({})

    render(<App />)

    expect(
      await screen.findByText(kataloge.en.seiten.start.was),
    ).toBeInTheDocument()
  })
})

// **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7, Regel 1).
//
// Wer in der Setliste ein Set wählt, landet bei dessen Karten — und die
// Kartenseite lädt beides neu, weil `setID` sich geändert hat. Ohne diesen Weg
// bliebe die Kartenliste des vorigen Sets stehen.
describe('Setwahl', () => {
  afterEach(() => {
    raeumeApi()
  })

  it('führt zu den Karten des gewählten Sets', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '42',
              name: 'Gewähltes Set',
              spiel: 'magic',
              stil: 'm15',
              herkunft: '',
              karten: 1,
              herkunftFehlt: false,
            },
          ],
        }),
      ...ausKarten((setID) =>
        Promise.resolve({
          karten: [
            {
              id: '1',
              nummer: 0,
              werte: [{ feld: 'name', wert: `Karte aus ${setID}` }],
            },
          ],
        }),
      ),
      Karten: (setID) =>
        Promise.resolve({
          karten: [
            {
              id: '1',
              nummer: 0,
              werte: [{ feld: 'name', wert: `Karte aus ${setID}` }],
            },
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    render(
      <Sprachraum sprache="de">
        <App />
      </Sprachraum>,
    )

    // **`findBy` und nicht `getBy`** — seit #245 zeichnet `App` erst, wenn die
    // Einstellungen da sind, und die kommen über die Fassade.
    await userEvent.click(await leisteneintrag('sets'))

    await userEvent.click(
      await screen.findByRole('button', { name: /^Gewähltes Set —/ }),
    )

    // **Gesucht im Inhalt**: seit UX-03.AK-02 steht die aufgeschlagene Karte
    // auch in der Brotkrume darüber.
    await waitFor(() => {
      expect(
        within(screen.getByRole('main')).getByText('Karte aus 42'),
      ).toBeInTheDocument()
    })
  })
})
