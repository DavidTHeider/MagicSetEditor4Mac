// **Ein Paket lässt sich im Fenster anlegen** —
// #1350.
//
// Die Abnahme sagt „ein **im Fenster** angelegtes `.mse-game`". Ohne diese
// Prüfung stünde die Bindung da und niemand riefe sie — eine Zusage ohne
// Einlösung, und genau das hält `bindungsaufrufer_test.go` fest.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { waehleZeile } from '../test/suchwahl'
import { Paketanlage } from './Paketanlage'

import type { PaketDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

// zeige stellt das Formular in einen Sprachraum.
function zeige(spiele: readonly PaketDTO[], angelegt: () => void) {
  render(
    <Sprachraum sprache="de">
      <Paketanlage spiele={spiele} angelegt={angelegt} />
    </Sprachraum>,
  )
}

const spiel: PaketDTO = {
  art: 'mse-game',
  name: 'eigenes',
  anzeigename: 'eigenes',
  version: '2026-09-04',
  pfad: '/p/eigenes.mse-game',
  herkunft: '',
  pruefsumme: '',
  gruppe: '',
  rang: 0,
  braucht: [],
  artrang: 30,
  herkunftsart: 'mitgeliefert' as const,
  hauptdatei: 'style',
  geaendert: false,
  gebrauchtVon: [],
  benutzt: 0,
}

describe('Paketanlage', () => {
  it('legt ein Spiel an und meldet es nach oben', async () => {
    const anlegen = vi.fn().mockResolvedValue({
      paket: { ...spiel, name: 'neues' },
      abgebrochen: false,
      fehler: null,
    })

    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
        ]),
      PaketAnlegen: anlegen,
    })

    const angelegt = vi.fn()

    zeige([spiel], angelegt)
    await effekteNachholen()

    // **Der Kurzname wird zum Verzeichnis** — die Korrektur von macOS machte
    // am nativen Fenster aus „namen“ ein „Namen“ (Hüllenprobe `fmt02d`).
    expect(screen.getByLabelText(/Kurzname/)).toHaveAttribute(
      'spellcheck',
      'false',
    )
    await userEvent.type(screen.getByLabelText(/Kurzname/), 'neues')
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    await screen.findByText(/Angelegt: neues/)

    expect(anlegen).toHaveBeenCalledWith('mse-game', 'neues', '', '')
    expect(angelegt).toHaveBeenCalled()
  })

  // **Ein Stil zeigt die Spielwahl, ein Spiel nicht.**
  //
  // Und die Antwort darauf kommt aus `PaketArten`, nicht aus einem Vergleich
  // im Fenster — sonst stünde dieselbe Regel an zwei Stellen.
  it('zeigt die Spielwahl nur, wo die Anwendung sie verlangt', async () => {
    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
        ]),
    })

    zeige([spiel], vi.fn())
    await effekteNachholen()

    // **Dieselbe Wirkung wie oben** — die Einträge der Auswahl.
    await screen.findByRole('option', { name: /Stil/ })

    expect(
      screen.queryByLabelText(/Spiel des neuen Pakets/),
    ).not.toBeInTheDocument()

    await userEvent.selectOptions(
      screen.getByLabelText(/Art des neuen Pakets/),
      'mse-style',
    )

    expect(
      await screen.findByLabelText(/Spiel des neuen Pakets/),
    ).toBeInTheDocument()
  })

  // **Ein Stil wird mit Vollnamen und Spiel angelegt.**
  //
  // Der Lauf darüber legt ein Spiel an und lässt zwei Felder unberührt — den
  // Vollnamen und die Spielwahl. Die beiden zusammen sind der Weg, den die
  // Abnahme meint: „ein Set lässt sich darauf anlegen" braucht **beide**
  // Pakete.
  it('legt einen Stil mit Vollnamen und Spiel an', async () => {
    const anlegen = vi.fn().mockResolvedValue({
      paket: { ...spiel, art: 'mse-style', name: 'eigenes-schlicht' },
      abgebrochen: false,
      fehler: null,
    })

    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
        ]),
      PaketAnlegen: anlegen,
    })

    zeige([spiel], vi.fn())
    await effekteNachholen()

    // **Gewartet wird auf die Wirkung, nicht auf die Beschriftung** (§3).
    // Die Auswahl steht von Anfang an da; was der Effekt bringt, sind ihre
    // Einträge. Der erste Anlauf wartete auf das Feld und griff in eine leere
    // Auswahl.
    await screen.findByRole('option', { name: /Stil/ })

    await userEvent.selectOptions(
      screen.getByLabelText(/Art des neuen Pakets/),
      'mse-style',
    )

    await userEvent.type(screen.getByLabelText(/Kurzname/), 'schlicht')
    await userEvent.type(
      screen.getByLabelText(/Vollständiger Name/),
      'Schlichter Stil',
    )
    await waehleZeile(
      await screen.findByLabelText(/Spiel des neuen Pakets/),
      'eigenes',
    )

    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    await screen.findByText(/Angelegt: eigenes-schlicht/)

    expect(anlegen).toHaveBeenCalledWith(
      'mse-style',
      'schlicht',
      'Schlichter Stil',
      'eigenes',
    )
  })

  // **Auch ein Fehler beim Lesen der Arten steht am Formular** (§8).
  //
  // Ohne diesen Fall ging die Ablehnung ins Leere — der Lauf meldete sie als
  // unbehandelt, und die Seite zeigte ein Formular ohne Arten, dessen Knopf
  // nichts tut. Gefunden hat es die Vorrichtung: sie scheitert mit Nennung der
  // Methode, statt `undefined` zu liefern (§3).
  it('zeigt einen Fehler beim Lesen der Arten', async () => {
    belegeApi({
      PaketArten: () => Promise.reject(new Error('die Arten kamen nicht')),
    })

    zeige([], vi.fn())

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Arten kamen nicht',
    )
  })

  // **Ein Formular, das man verlässt, schreibt nicht mehr in seinen Zustand.**
  //
  // §7, Regel 2 in ihrer kleinsten Gestalt: die Antwort gehört zu dem Stand,
  // aus dem sie stammt — und wenn es den nicht mehr gibt, gehört sie nirgends
  // hin. Belegt über die Konsole: ohne den Wächter meldet React einen
  // Zustandswechsel an einem abgeräumten Bauteil.
  it('schreibt nach dem Abräumen nicht mehr', async () => {
    let aufloesen: (arten: unknown[]) => void = () => undefined

    belegeApi({
      PaketArten: () =>
        new Promise((gelingt) => {
          aufloesen = gelingt as (arten: unknown[]) => void
        }),
    })

    const { unmount } = render(
      <Sprachraum sprache="de">
        <Paketanlage spiele={[]} angelegt={vi.fn()} />
      </Sprachraum>,
    )

    unmount()
    aufloesen([{ art: 'mse-game', brauchtSpiel: false }])

    // **Ein Durchlauf der Warteschlange**, damit das `then` wirklich läuft.
    await effekteNachholen()
  })

  // **Eine Antwort ohne Paket nennt trotzdem einen Namen.**
  //
  // „Angelegt: undefined" wäre die Meldung, die niemandem hilft.
  it('nennt den eingegebenen Namen, wenn die Antwort keinen trägt', async () => {
    belegeApi({
      PaketArten: () =>
        Promise.resolve([{ art: 'mse-game', brauchtSpiel: false }]),
      PaketAnlegen: () =>
        Promise.resolve({ paket: null, abgebrochen: false, fehler: null }),
    })

    zeige([], vi.fn())
    await effekteNachholen()

    await userEvent.type(screen.getByLabelText(/Kurzname/), 'namenlos')
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    expect(await screen.findByText(/Angelegt: namenlos/)).toBeInTheDocument()
  })

  // **Ein Feldfehler steht am Feld, nicht über dem Formular** (§6).
  //
  // # Was hier vorher stand
  //
  // Dieser Lauf hielt fest, dass die Meldung als anonymer Hinweis erscheint
  // und dabei zum allgemeinen Katalogsatz „Die Eingabe ist nicht gültig" wird.
  // **Er hat das falsche Verhalten festgeschrieben** — angestrichen vom
  // Nachlauf über Phase 4.10.
  //
  // Zwei Dinge waren falsch: die Anwendung meldet den Fehler **mit** dem Feld,
  // und das Fenster warf den Bezug weg; und der Schlüssel `validierung` gab
  // einen Satz, der die Abhilfe nicht nennt. Die Namensregel hat jetzt einen
  // eigenen Schlüssel, und die Regel **ist** die Abhilfe.
  it('zeigt einen Feldfehler am verursachenden Feld', async () => {
    belegeApi({
      PaketArten: () =>
        Promise.resolve([{ art: 'mse-game', brauchtSpiel: false }]),
      PaketAnlegen: () =>
        Promise.resolve({
          paket: null,
          abgebrochen: false,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'paketnameUngueltig',
            meldung: 'Ein Paketname besteht aus …',
            feld: 'name',
          },
        }),
    })

    zeige([], vi.fn())
    await effekteNachholen()

    const feld = screen.getByLabelText(/Kurzname/)

    await userEvent.type(feld, 'stil!')
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    // **Die Regel steht da, nicht „nicht gültig".**
    const meldung = await screen.findByRole('alert')

    expect(meldung).toHaveTextContent(/Bindestrichen und Unterstrichen/)

    // **Und sie hängt am Feld** — eine Vorlesehilfe findet sie von dort.
    expect(feld).toHaveAttribute('aria-invalid', 'true')
    // **Enthält, nicht gleich** — seit `Feld` (#1413) nennt die Beschreibung
    // Befund und Hinweis zusammen.
    expect(feld.getAttribute('aria-describedby')).toContain(meldung.id)
  })

  // **Ein Fehler ohne Feldbezug steht weiter über dem Formular.**
  //
  // Die Gegenprobe zum Lauf darüber: ohne sie wäre eine Fassung, die **jeden**
  // Fehler ans Namensfeld hängt, ebenso grün — und ein Fehler ohne Feld
  // verschwände.
  it('zeigt einen Fehler ohne Feldbezug über dem Formular', async () => {
    belegeApi({
      PaketArten: () =>
        Promise.resolve([{ art: 'mse-game', brauchtSpiel: false }]),
      PaketAnlegen: () =>
        Promise.resolve({
          paket: null,
          abgebrochen: false,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Etwas ging schief.',
          },
        }),
    })

    zeige([], vi.fn())
    await effekteNachholen()

    const feld = screen.getByLabelText(/Kurzname/)

    await userEvent.type(feld, 'eigenes')
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    await screen.findByRole('alert')

    expect(feld).not.toHaveAttribute('aria-invalid', 'true')
  })
})

// **Jede Art trägt ihren Namen** (FMT-02, #1691) — hier stand „ist es kein
// Spiel, ist es ein Stil“, und Include wie Exportvorlage hießen im Menü
// „Stil (.mse-style)“.
describe('Die Arten im Menü', () => {
  it('nennt jede der vier Arten bei ihrem Namen', async () => {
    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
          { art: 'mse-include', brauchtSpiel: false },
          { art: 'mse-export-template', brauchtSpiel: true },
          // **Eine Art, die hier noch keinen Namen hat**, steht mit ihrer
          // Endung da — nicht unter einem falschen.
          { art: 'mse-neu', brauchtSpiel: false },
        ]),
    })

    zeige([spiel], vi.fn())
    await effekteNachholen()

    expect(
      (await screen.findAllByRole('option')).map(
        (option) => option.textContent,
      ),
    ).toEqual([
      'Spiel (.mse-game)',
      'Stil (.mse-style)',
      'Include (.mse-include)',
      'Exportvorlage (.mse-export-template)',
      'mse-neu',
    ])
  })
})
