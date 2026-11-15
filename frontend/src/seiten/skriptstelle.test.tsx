// **Ein Syntaxfehler wird an seiner Stelle angestrichen** — Werkstück 1 von
// Phase 3.3 (#1164).
//
// # Der Stand davor
//
// Der Zerteiler kennt Zeile und Spalte seit je; die Konsole warf beides weg
// und zeigte einen Satz. Wer einen Ausdruck über drei Zeilen schreibt, suchte
// die Stelle von Hand — und die Angabe stand die ganze Zeit da.
//
// # Warum ein Zeiger und keine Farbe
//
// Eine Farbe im Textfeld hieße, den Text nachzubauen: ein zweites Feld über
// dem ersten, das bei jeder Schriftgröße und jedem Umbruch danebenliegt. Ein
// Zeiger darunter ist, was ein Übersetzer seit je tut — und er ist vorlesbar.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Skriptkonsole } from './Skriptkonsole'

const texte = kataloge.de.seiten.karten.konsole

// gescheitert belegt die Fassade mit einem Fehler samt Stelle.
//
// **Mit dem Schlüssel, den Go wirklich setzt** (§3: eine Vorrichtung, die
// großzügiger ist als der Betrieb, prüft den falschen Weg). Hier stand
// `'probe'` — ein Schlüssel, den kein Katalog kennt, und deshalb fiel die
// Anzeige auf den Go-Satz zurück. Im Betrieb kommt `skriptFehler`, das Fenster
// **bevorzugt den Katalog**, und über dem angestrichenen Ausdruck stand
// deshalb „Es ist ein unerwarteter Fehler aufgetreten".
function gescheitert(zeile: number, spalte: number) {
  belegeApi({
    Skript: () =>
      Promise.resolve({
        quelltext: '',
        text: '',
        art: '',
        befunde: [],
        dauerMs: 0,
        abgebrochen: false,
        eintraege: [],
        stelle: { zeile, spalte },
        fehler: {
          code: 'validierung' as const,
          schluessel: 'skriptFehler',
          meldung: 'Zeile 2, Spalte 3: etwas fehlt',
          werte: { meldung: 'etwas fehlt' },
        },
      }),
  })
}

describe('Die Stelle eines Syntaxfehlers', () => {
  afterEach(raeumeApi)

  it('steht in der Meldung', async () => {
    gescheitert(2, 3)

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.type(
      screen.getByLabelText(texte.ausdruck),
      '1 +{Enter}  * 2',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    const meldung = await screen.findByRole('alert')

    expect(meldung).toHaveTextContent('Zeile 2, Spalte 3')

    // **Und sagt, was fehlt.** Eine Stelle ohne Aussage schickt nur an eine
    // andere Zeile — der Zerteiler weiß mehr, und er sagt es seit je.
    expect(meldung).toHaveTextContent('etwas fehlt')
  })

  // **Die Stelle wird im Feld selbst markiert, nicht daneben.**
  //
  // Hier stand ein Lauf über einen `<pre>`-Block mit einem `^` darunter. Den
  // gibt es nicht mehr: seit das Quelltextfeld die Spalte anstreicht, wäre er
  // die „Liste daneben", die das Issue ausdrücklich nicht will
  // (#1164).
  it('markiert die Spalte im Feld', async () => {
    gescheitert(2, 3)

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.type(
      screen.getByLabelText(texte.ausdruck),
      '1 +{Enter}  * 2',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.auswerten }))

    const marke = await screen.findByTestId('fehlermarke')

    expect(marke.textContent).toBe('*')
  })
})
