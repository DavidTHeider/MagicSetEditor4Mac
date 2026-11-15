// **Der Weg von Scryfall bis in die Karte** —
// #1005.
//
// # Was hier geprüft wird und was nicht
//
// Geprüft wird der **Weg**: Name eingeben, Art wählen, holen, zuschneiden,
// einsetzen — und dass dabei der Künstler mitgeht. Dass er wirklich in die
// Karte geschrieben wird, prüft die Go-Seite
// (`internal/app/scryfallbild_test.go`); hier endet die Zusage am Aufruf.
//
// # Die Nennungspflicht ist eine Zusicherung dieses Tests
//
// Scryfall verlangt Künstler und Copyright „in the same interface presenting
// the art crop" — dass der Satz **dasteht**, solange das Bild offen ist, kann
// nur ein Test im Fenster sehen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { zuschnittantwort } from '../test/zuschnitt'
import { Karteneditor } from './Karteneditor'
import { kurzform } from '../sprache/kurzform'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const scryfalltexte = texte.seiten.karten.scryfall

function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [{ feld: 'kunst', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
}

const bildform = {
  name: 'kunst',
  aufschrift: 'Kunst',
  art: 'image',
  mehrzeilig: false,
  wahlen: [],
  beschreibung: '',
  vorgabe: '',
  teilfelder: [],
  reihe: '',
  stelle: 0,
  editierbar: true,
  gespeichert: true,
}

function karte(): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'name', wert: 'Lightning Bolt' }],
  }
}

// bildantwort ist, was `ScryfallBild` zurückgibt.
function bildantwort(teile = {}) {
  return {
    ...zuschnittantwort(),
    kuenstler: 'Christopher Rush',
    nennung: 'Illustration: Christopher Rush. ™ & © Wizards of the Coast',
    adresse: 'https://scryfall.com/card/lea/161',
    alterStunden: 0,
    befunde: [],
    fehler: null,
    ...teile,
  }
}

function zeige(teile = {}) {
  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    Zuschnittgrenze: () =>
      Promise.resolve({ x: 0, y: 0, breite: 125, hoehe: 100 }),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte()}
        formen={[bildform]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

// waehleBildfeld klickt das Bildfeld an, damit der Kasten erscheint.
async function waehleBildfeld() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: `Kunst — ${texte.seiten.karten.bildWaehlen}`,
    }),
  )
}

describe('Der Bildweg über Scryfall', () => {
  it('belegt den Kartennamen vor und holt das Bild', async () => {
    const ScryfallBild = vi.fn(() => Promise.resolve(bildantwort()))

    zeige({ ScryfallBild, BildSetzen: () => Promise.resolve({ pfad: '' }) })

    await waehleBildfeld()

    const namensfeld = await screen.findByLabelText(scryfalltexte.name)
    expect(namensfeld).toHaveValue('Lightning Bolt')

    await userEvent.click(
      screen.getByRole('button', { name: scryfalltexte.holen }),
    )

    await waitFor(() => {
      expect(ScryfallBild).toHaveBeenCalledWith(
        '5',
        'kunst',
        'Lightning Bolt',
        'normal',
      )
    })
  })

  it('zeigt die Nennung, solange das Bild offen ist', async () => {
    zeige({
      ScryfallBild: () => Promise.resolve(bildantwort()),
      BildSetzen: () => Promise.resolve({ pfad: '' }),
    })

    await waehleBildfeld()
    await userEvent.click(
      await screen.findByRole('button', { name: scryfalltexte.holen }),
    )

    // **Der Kern.** Ohne diesen Satz verletzt die Anwendung die Bedingung, zu
    // der sie sich mit jedem Abruf verpflichtet.
    expect(await screen.findByText(/Christopher Rush/)).toHaveTextContent(
      'Wizards of the Coast',
    )

    expect(
      screen.getByRole('link', { name: scryfalltexte.quelle }),
    ).toHaveAttribute('href', 'https://scryfall.com/card/lea/161')
  })

  // **Und was die Auskunft tolerant hingenommen hat, steht dabei.**
  //
  // §16 sagt: nichts wird still übernommen, und ein unbekannter
  // Aufzählungswert wird als Befund geführt. `Bildweg.Holen` warf die Befunde
  // von `KarteNachName` mit `_` weg — wer über den Bildweg ging, erfuhr von
  // einer Drift beim Anbieter nichts, während der Feldabgleich sie zeigte.
  //
  // Nachlauf über Phase 4.5, Befund 10.
  it('zeigt die Befunde der Kartenauskunft neben dem Bild', async () => {
    zeige({
      ScryfallBild: () =>
        Promise.resolve(
          bildantwort({ befunde: ['layout: unbekannter Wert »seltsam«'] }),
        ),
      BildSetzen: () => Promise.resolve({ pfad: '' }),
    })

    await waehleBildfeld()
    await userEvent.click(
      await screen.findByRole('button', { name: scryfalltexte.holen }),
    )

    expect(
      await screen.findByText('layout: unbekannter Wert »seltsam«'),
    ).toBeInTheDocument()
  })

  it('setzt das Bild samt Künstler ein', async () => {
    const ScryfallBildEinsetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({
      ScryfallBild: () => Promise.resolve(bildantwort()),
      BildSetzen: () => Promise.resolve({ pfad: '' }),
      ScryfallBildEinsetzen,
    })

    await waehleBildfeld()
    await userEvent.click(
      await screen.findByRole('button', { name: scryfalltexte.holen }),
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    )

    await waitFor(() => {
      expect(ScryfallBildEinsetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'kunst',
        '/pfad/urlaub.jpg',
        'Christopher Rush',
        // **Die Herkunft geht mit** (#888): sie landet im Verlauf, an beiden
        // Zeilen — dem Bildfeld und dem Künstler.
        'https://scryfall.com/card/lea/161',
        { x: 0, y: 0, breite: 125, hoehe: 100 },
      )
    })
  })

  it('reicht die gewählte Bildart durch und warnt beim Ausschnitt', async () => {
    const ScryfallBild = vi.fn(() => Promise.resolve(bildantwort()))

    zeige({ ScryfallBild, BildSetzen: () => Promise.resolve({ pfad: '' }) })

    await waehleBildfeld()

    await userEvent.selectOptions(
      await screen.findByLabelText(scryfalltexte.art),
      'ausschnitt',
    )

    // **Der Hinweis erscheint erst bei der Wahl** — er gilt nur für sie.
    expect(
      screen.getByText(kurzform(scryfalltexte.ausschnittHinweis)),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: scryfalltexte.holen }),
    )

    await waitFor(() => {
      expect(ScryfallBild).toHaveBeenCalledWith(
        '5',
        'kunst',
        'Lightning Bolt',
        'ausschnitt',
      )
    })
  })

  it('meldet einen Fehler am Feld und öffnet keinen Zuschnitt', async () => {
    zeige({
      ScryfallBild: () =>
        Promise.resolve({
          ...bildantwort(),
          fehler: {
            code: 'nichtGefunden',
            schluessel: '',
            meldung: 'Scryfall kennt diese Karte nicht.',
          },
        }),
      BildSetzen: () => Promise.resolve({ pfad: '' }),
    })

    await waehleBildfeld()
    await userEvent.click(
      await screen.findByRole('button', { name: scryfalltexte.holen }),
    )

    expect(
      // **Zweimal**: am Feld und in der Problemliste des Inspectors (U5, #1416).
      (await screen.findAllByText('Scryfall kennt diese Karte nicht.'))[0],
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    ).not.toBeInTheDocument()
  })

  it('hält den Knopf an, solange kein Name dasteht', async () => {
    zeige({
      ScryfallBild: () => Promise.resolve(bildantwort()),
      BildSetzen: () => Promise.resolve({ pfad: '' }),
    })

    await waehleBildfeld()

    await userEvent.clear(await screen.findByLabelText(scryfalltexte.name))

    expect(
      screen.getByRole('button', { name: scryfalltexte.holen }),
    ).toBeDisabled()
  })
})
