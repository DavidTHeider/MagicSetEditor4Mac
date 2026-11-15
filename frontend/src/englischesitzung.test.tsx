// Eine englische Sitzung zeigt englische Meldungen — auch die aus Go.
//
// # Der Befund, an dem diese Datei hängt
//
// #133 hat den
// Weg gebaut: eine Meldung aus Go wird über ihren **Schlüssel** übersetzt. Die
// letzte Zusage des Issues lautete „nachgemessen an der laufenden Anwendung in
// einer englischen Sitzung" — und genau dort fiel es auf.
//
// Im laufenden Fenster stand in einer englischen Sitzung:
//
//	This could not be loaded. Der gesuchte Eintrag wurde nicht gefunden.
//
// Halb englisch, halb deutsch. Und der deutsche Teil war **nicht** der Satz aus
// Go — der lautete „Es ist ein unerwarteter Fehler aufgetreten." —, sondern der
// Eintrag aus dem **deutschen Katalog**.
//
// **Die Ursache ist der Ort des Hakens.** `useSetSpeichern` ruft `useTexte()`,
// und `App` ruft den Haken **über** dem `Sprachraum`, den sie selbst herstellt.
// Der Haken las damit den Vorgabewert des Kontexts — Deutsch. Für den
// Meldefehler daneben stand der Fall bereits im Quelltext angemerkt und war
// dort umgangen; der Haken blieb übrig. **Ein Wächter, der die halbe Zusage
// hält, macht die andere unsichtbar** (§10).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { leisteneintrag } from './test/leistenweg'
import { waehleSethandlung } from './test/sethandlung'

afterEach(() => {
  raeumeApi()
})

// scheiterndesSichern belegt die Fassade mit einem Set, dessen Sichern
// fehlschlägt — mit demselben Fehler, der im Fenster auffiel.
function scheiterndesSichern() {
  belegeApi({
    ...belegeFenster('en'),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '7',
            name: 'Mein Probeset',
            spiel: 'probe',
            stil: 'probe-stil',
            herkunft: '/pfad/probe.json',
            karten: 1,
            herkunftFehlt: false,
          },
        ],
      }),
    SetSpeichern: () =>
      Promise.resolve({
        pfad: '',
        befunde: [],
        abgebrochen: false,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'nichtGefunden',
          // **Der deutsche Satz aus Go reist mit**, so wie im Betrieb. Er ist
          // der Rückfall für einen Schlüssel, den der Katalog nicht kennt —
          // und darf hier gerade **nicht** sichtbar werden.
          meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
          feld: '',
        },
      }),
  })
}

describe('Eine englische Sitzung', () => {
  it('zeigt eine Meldung aus Go auf Englisch', async () => {
    belegeApi({ ...belegeFenster('en') })
    scheiterndesSichern()

    render(<App />)

    // **Erst zur Setliste**: die Anwendung startet auf der Startseite.
    await userEvent.click(await leisteneintrag('sets', 'en'))

    // **Über das Menü der Setkarte** seit UXR-04 (#1912).
    await waehleSethandlung(
      kataloge.en,
      'Mein Probeset',
      kataloge.en.seiten.sets.speichern,
    )

    // **Zweimal ist richtig**: einmal im Band über dem Inhalt (#180), einmal
    // an der Setliste. Beide zeigen dieselbe Meldung, und beide müssen
    // englisch sein.
    expect(
      await screen.findAllByText(kataloge.en.fehler.nichtGefunden, {
        exact: false,
      }),
    ).toHaveLength(2)

    expect(
      screen.queryByText(kataloge.de.fehler.nichtGefunden, { exact: false }),
    ).not.toBeInTheDocument()
  })
})
