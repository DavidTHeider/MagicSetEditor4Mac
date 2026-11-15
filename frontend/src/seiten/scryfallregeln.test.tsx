// **Die Auskünfte von Scryfall im Fenster** —
// #888.
//
// # Was hier geprüft wird
//
// Dass die Auskünfte **erreichbar** sind und dass die Unterscheidung
// offiziell/Anmerkung ankommt. Sie ist der Punkt: eine Anmerkung des
// Scryfall-Teams erklärt oft ein inoffizielles Format, und sie als Regel
// darzustellen wäre falsch.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.regeln

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
      flaechen: [{ feld: 'name', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
}

const namensform = {
  name: 'name',
  aufschrift: 'Name',
  art: 'text',
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
    werte: [
      { feld: 'name', wert: 'Lightning Bolt' },
      { feld: 'casting cost', wert: '{2}{R}' },
    ],
  }
}

function zeige(teile = {}) {
  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte()}
        formen={[namensform]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

const auskunft = {
  kartenname: 'Lightning Bolt',
  hinweise: [
    {
      offiziell: true,
      quelle: 'wotc',
      datum: '2004-10-04',
      text: 'Der Schaden lässt sich nicht aufteilen.',
    },
    {
      offiziell: false,
      quelle: 'scryfall',
      datum: '2021-07-23',
      text: 'In Duel Commander gilt dasselbe.',
    },
  ],
  adresse: 'https://scryfall.invalid/lea/161',
  befunde: [],
  alterStunden: 0,
  fehler: null,
}

describe('Die Auskünfte von Scryfall', () => {
  it('holt die Hinweise mit dem Kartennamen', async () => {
    const ScryfallRegeln = vi.fn(() => Promise.resolve(auskunft))

    zeige({ ScryfallRegeln })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.holen }),
    )

    expect(ScryfallRegeln).toHaveBeenCalledWith('Lightning Bolt')

    // **Beide Hinweise stehen da, und sie sind unterscheidbar.**
    expect(
      await screen.findByText(/Der Schaden lässt sich nicht aufteilen/),
    ).toBeInTheDocument()

    // **Die Überschriften der beiden Hinweise, nicht irgendein Vorkommen.**
    //
    // Ein `new RegExp('offiziell')` traf auch den Erklärsatz darunter — dort
    // steht „inoffizielles Format". Ein Muster, das mehr trifft als gemeint,
    // prüft nicht mehr, was es sagt.
    expect(
      screen.getByText(`${texte.offiziell} · 2004-10-04`),
    ).toBeInTheDocument()

    expect(
      screen.getByText(`${texte.anmerkung} · 2021-07-23`),
    ).toBeInTheDocument()

    expect(screen.getByRole('link', { name: texte.quelle })).toHaveAttribute(
      'href',
      'https://scryfall.invalid/lea/161',
    )
  })

  it('sagt es, wenn es keine Hinweise gibt', async () => {
    zeige({
      ScryfallRegeln: () => Promise.resolve({ ...auskunft, hinweise: [] }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.holen }),
    )

    expect(await screen.findByText(texte.keine)).toBeInTheDocument()
  })

  it('nennt das Alter eines gespeicherten Standes', async () => {
    zeige({
      ScryfallRegeln: () => Promise.resolve({ ...auskunft, alterStunden: 5.7 }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.holen }),
    )

    // **Abgerundet**: „5 Stunden alt" ist wahr, „5,7" wäre eine Genauigkeit,
    // die der Zwischenspeicher nicht zusagt.
    expect(await screen.findByText(/5 Stunden alt/)).toBeInTheDocument()
  })

  it('meldet einen Fehler der Auskunft', async () => {
    zeige({
      ScryfallRegeln: () =>
        Promise.resolve({
          ...auskunft,
          fehler: {
            code: 'nichtGefunden',
            schluessel: '',
            meldung: 'Scryfall kennt diese Karte nicht.',
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.holen }),
    )

    expect(
      await screen.findByText('Scryfall kennt diese Karte nicht.'),
    ).toBeInTheDocument()
  })

  it('prüft den Manakosten der Karte', async () => {
    const ScryfallManakosten = vi.fn(() =>
      Promise.resolve({ unbekannt: [], fehler: null }),
    )

    zeige({ ScryfallManakosten })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.manaPruefen }),
    )

    expect(ScryfallManakosten).toHaveBeenCalledWith('{2}{R}')
    expect(await screen.findByText(texte.manaGut)).toBeInTheDocument()
  })

  it('nennt die unbekannten Symbole', async () => {
    zeige({
      ScryfallManakosten: () =>
        Promise.resolve({ unbekannt: ['{r}', '{q}'], fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.manaPruefen }),
    )

    expect(await screen.findByText(/\{r\}, \{q\}/)).toBeInTheDocument()
  })

  it('meldet einen Fehler der Manakostenprüfung', async () => {
    zeige({
      ScryfallManakosten: () =>
        Promise.resolve({
          unbekannt: [],
          fehler: { code: 'unbekannt', schluessel: '', meldung: 'Kein Netz.' },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.manaPruefen }),
    )

    expect(await screen.findByText('Kein Netz.')).toBeInTheDocument()
  })
})

// **Eine Karte ohne Manakosten wird nicht geprüft.**
//
// Der Aufruf ginge sonst mit einer leeren Zeichenkette hinaus und käme mit
// „in Ordnung" zurück — eine Auskunft über nichts.
describe('Eine Karte ohne Manakosten', () => {
  it('sagt es, statt zu fragen', async () => {
    const ScryfallManakosten = vi.fn(() =>
      Promise.resolve({ unbekannt: [], fehler: null }),
    )

    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      Probevorschau: () => Promise.resolve(vorschau()),
      ScryfallManakosten,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={{
            id: '5',
            nummer: 0,
            werte: [{ feld: 'name', wert: 'Ohne' }],
          }}
          formen={[namensform]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.manaPruefen }),
    )

    expect(await screen.findByText(texte.manaLeer)).toBeInTheDocument()
    expect(ScryfallManakosten).not.toHaveBeenCalled()
  })
})
