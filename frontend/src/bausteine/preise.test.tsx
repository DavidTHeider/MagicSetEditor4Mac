// **Die Preise und die Übersicht** — Werkstück 3 von Phase 3.4
// (#1165).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Portfolio } from './Portfolio'
import { Preise } from './Preise'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.sammlung

// stand baut einen Preisstand.
function stand(teil: Record<string, unknown> = {}) {
  return {
    printing: 'probe-druck',
    druckID: '',
    quelle: 'scryfall',
    markt: 'paper',
    art: 'retail',
    zustand: '',
    sprache: '',
    lage: 'genannt',
    beobachtet: '',
    waehrung: 'EUR',
    finish: 'nonfoil',
    wert: '4.17',
    erhoben: '2026-08-27T12:00:00Z',
    alterStunden: 3,
    ...teil,
  }
}

// zeigePreise zeichnet die Preise eines Drucks.
//
// **Der Rückruf kommt zurück**, damit ein Lauf prüfen kann, dass die Meldung
// nach oben geht — die Übersicht hängt daran.
function zeigePreise() {
  const geholt = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Preise printing="probe-druck" aufGeholt={geholt} />
    </Sprachraum>,
  )

  return geholt
}

describe('Die Preise', () => {
  afterEach(raeumeApi)

  // **Beim Aufschlagen wird gelesen, nicht gefragt.**
  //
  // Ein Programm, das für dreihundert Zeilen dreihundert Abrufe auslöst,
  // überrascht seine Nutzenden (§16).
  it('fragt beim Aufschlagen nicht im Netz', async () => {
    const geholt = vi.fn(() =>
      Promise.resolve({ staende: [], befunde: [], fehler: null }),
    )

    belegeApi({
      Preise: () =>
        Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
      PreiseHolen: geholt,
    })

    zeigePreise()

    expect(await screen.findByText(/4,17\s€/u)).toBeInTheDocument()
    expect(geholt).not.toHaveBeenCalled()
  })

  it('zeigt Alter und Quelle neben dem Preis', async () => {
    belegeApi({
      Preise: () =>
        Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
    })

    zeigePreise()

    expect(await screen.findByText(/vor 3 h · scryfall/)).toBeInTheDocument()
  })

  // **Unter einer Stunde heißt „gerade eben".** Eine Null in einer
  // Altersangabe liest sich wie ein Fehler.
  it('sagt bei einem frischen Preis »gerade eben«', async () => {
    belegeApi({
      Preise: () =>
        Promise.resolve({
          staende: [stand({ alterStunden: 0 })],
          befunde: [],
          fehler: null,
        }),
    })

    zeigePreise()

    expect(await screen.findByText(/gerade eben/)).toBeInTheDocument()
  })

  // **Veraltet ist nicht null** (U8, #1419): ein Preis, der älter als eine
  // Woche ist, sagt es in Tagen — und unterscheidet sich von „kein Preis".
  it('nennt einen Preis älter als eine Woche veraltet', async () => {
    belegeApi({
      Preise: () =>
        Promise.resolve({
          staende: [stand({ alterStunden: 200 })],
          befunde: [],
          fehler: null,
        }),
    })
    zeigePreise()
    expect(
      await screen.findByText(/veraltet, 8 Tage alt · scryfall/),
    ).toBeInTheDocument()
  })

  it('sagt es, wenn kein Preis vorliegt', async () => {
    belegeApi({
      Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    })

    zeigePreise()

    expect(await screen.findByText(texte.keinePreise)).toBeInTheDocument()
  })

  it('holt Preise auf Klick und lädt danach neu', async () => {
    const geholt = vi.fn(() =>
      Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
    )
    const gelesen = vi.fn(() =>
      Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
    )

    belegeApi({ Preise: gelesen, PreiseHolen: geholt })

    const gemeldet = zeigePreise()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.preiseHolen} probe-druck`,
      }),
    )

    await waitFor(() => {
      expect(geholt).toHaveBeenCalledWith('probe-druck')
    })

    await waitFor(() => {
      expect(gelesen).toHaveBeenCalledTimes(2)
    })

    // **Und die Übersicht erfährt davon.** Ohne diese Meldung stünde sie nach
    // einem Abruf weiter auf „noch keine Preise geholt" — im laufenden Fenster
    // gesehen, von keinem Lauf gezeigt.
    expect(gemeldet).toHaveBeenCalled()
  })

  it('meldet einen Fehler beim Holen', async () => {
    belegeApi({
      Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
      PreiseHolen: () =>
        Promise.resolve({
          staende: [],
          befunde: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'weg',
          },
        }),
    })

    zeigePreise()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.preiseHolen} probe-druck`,
      }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  // **Der Verlauf klappt auf und wieder zu.**
  it('klappt den Verlauf auf und zu', async () => {
    belegeApi({
      Preise: () =>
        Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
      Preisverlauf: () =>
        Promise.resolve({
          staende: [stand(), stand({ erhoben: '2026-08-20T12:00:00Z' })],
          befunde: [],
          fehler: null,
        }),
    })

    zeigePreise()

    const reihe = await screen.findByRole('button', {
      name: `${texte.verlauf} EUR nonfoil`,
    })

    await userEvent.click(reihe)

    expect(await screen.findByText('2026-08-20T12:00:00Z')).toBeInTheDocument()

    await userEvent.click(reihe)

    await waitFor(() => {
      expect(screen.queryByText('2026-08-20T12:00:00Z')).not.toBeInTheDocument()
    })
  })

  it('sagt es, wenn ein Verlauf noch keiner ist', async () => {
    belegeApi({
      Preise: () =>
        Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
      Preisverlauf: () =>
        Promise.resolve({ staende: [], befunde: [], fehler: null }),
    })

    zeigePreise()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.verlauf} EUR nonfoil`,
      }),
    )

    expect(await screen.findByText(texte.keinVerlauf)).toBeInTheDocument()
  })

  it('meldet einen Fehler des Verlaufs', async () => {
    belegeApi({
      Preise: () =>
        Promise.resolve({ staende: [stand()], befunde: [], fehler: null }),
      Preisverlauf: () =>
        Promise.resolve({
          staende: [],
          befunde: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'weg',
          },
        }),
    })

    zeigePreise()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.verlauf} EUR nonfoil`,
      }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})

describe('Die Übersicht', () => {
  afterEach(raeumeApi)

  // **Drei Angaben stehen neben jeder Summe** — und die Zahl ohne Preis
  // dazu. Ohne sie sähe die Summe aus wie eine über alles.
  it('nennt Summe, Zahl, Alter und die ohne Preis', async () => {
    belegeApi({
      Portfolio: () =>
        Promise.resolve({
          zeilen: [
            {
              waehrung: 'EUR',
              wert: '12.51',
              exemplare: 3,
              aeltesterPreis: '2026-08-27T12:00:00Z',
              alterStunden: 5,
              quelle: 'scryfall',
            },
          ],
          ohnePreis: 7,
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Portfolio runde={0} />
      </Sprachraum>,
    )

    expect(await screen.findByText(/12,51\s€/u)).toBeInTheDocument()
    expect(screen.getByText(/3 Exemplare/)).toBeInTheDocument()

    // **Und das Alter steht in denselben Worten wie an der Zeile.** Zwei
    // Fassungen desselben Satzes standen nebeneinander im Bild: „gerade eben"
    // und „vor 0 h".
    expect(screen.getByText(/vor 5 h · scryfall/)).toBeInTheDocument()
    expect(screen.getByText(/ohne Preis: 7/)).toBeInTheDocument()

    // **Und der Grund, warum nicht umgerechnet wird.**
    expect(screen.getByText(texte.keineUmrechnung)).toBeInTheDocument()
  })

  // **Ein Exemplar heißt „1 Exemplar" und nicht „1 Exemplare".**
  //
  // Eine Zahl mit falscher Form liest sich wie ein Fehler im Programm — im
  // laufenden Fenster gesehen.
  it('nennt ein einzelnes Exemplar in der Einzahl', async () => {
    belegeApi({
      Portfolio: () =>
        Promise.resolve({
          zeilen: [
            {
              waehrung: 'EUR',
              wert: '4.17',
              exemplare: 1,
              aeltesterPreis: '2026-08-27T12:00:00Z',
              alterStunden: 0,
              quelle: 'scryfall',
            },
          ],
          ohnePreis: 0,
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Portfolio runde={0} />
      </Sprachraum>,
    )

    expect(await screen.findByText(/1 Exemplar/)).toBeInTheDocument()

    // **Und ein frischer Stand sagt auch hier „gerade eben".**
    expect(screen.getByText(/gerade eben/)).toBeInTheDocument()
  })

  it('sagt es, wenn noch nichts gerechnet ist', async () => {
    belegeApi({
      Portfolio: () =>
        Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Portfolio runde={0} />
      </Sprachraum>,
    )

    expect(await screen.findByText(texte.uebersichtLeer)).toBeInTheDocument()
  })

  it('meldet einen Fehler', async () => {
    belegeApi({
      Portfolio: () =>
        Promise.resolve({
          zeilen: [],
          ohnePreis: 0,
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'weg',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Portfolio runde={0} />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
