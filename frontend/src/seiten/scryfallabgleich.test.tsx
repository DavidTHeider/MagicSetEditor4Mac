// **Der Feldabgleich mit Scryfall im Fenster** —
// #1013.
//
// `ScryfallVorschlag` und `ScryfallUebernehmen` waren als Bindungen
// angemeldet, standen im Fähigkeitskatalog und waren über MCP, CLI und REST
// erreichbar — **kein Bauteil unter `frontend/src` rief sie**. Dieselbe
// Gestalt wie #852: ein Weg mit voller Abdeckung, den niemand erreichen kann.
//
// # Die zwei Schritte sind der Punkt
//
// §16, Regel 6: „Nichts wird still übernommen. Was von außen in ein Kartenfeld
// will, zeigt vorher, was sich ändert." Der Kasten zeigt **je Feld alt und
// neu** — ein Vorschlag, der nur den neuen Wert zeigte, ließe niemanden
// entscheiden.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import type { KarteDTO, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.abgleich

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
      nennungen: [],
      flaechen: [],
      vorgaben: [],
    },
  }
}

function karte(): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'name', wert: 'Lightning Bolt' }],
  }
}

function einVorschlag(teile = {}) {
  return {
    vorschlag: {
      aenderungen: [
        { feld: 'rule text', alt: '', neu: 'Fügt 3 Schaden zu.' },
        { feld: 'rarity', alt: 'common', neu: 'uncommon' },
      ],
      befunde: ['unbekannte Seltenheit'],
      herkunft: 'https://scryfall.com/card/lea/161',
      alterStunden: 2,
      stand: 'stand-des-vorschlags',
    },
    fehler: null,
    ...teile,
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
        formen={[]}
        geaendert={() => undefined}
        karte={karte()}
        setID="1"
      />
    </Sprachraum>,
  )
}

async function holeVorschlag() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.holen }),
  )
}

it('zeigt je Feld den alten und den neuen Wert', async () => {
  zeige({ ScryfallVorschlag: () => Promise.resolve(einVorschlag()) })

  await holeVorschlag()

  // **Beide Seiten** — sonst kann niemand entscheiden.
  expect(await screen.findByText(/Fügt 3 Schaden zu\./)).toBeInTheDocument()
  expect(screen.getByText(/bisher: common/)).toBeInTheDocument()
  expect(screen.getByText(/neu: uncommon/)).toBeInTheDocument()

  // **Das Alter steht dabei** (§16, Regel 5): ein Vorschlag aus einem drei
  // Tage alten Zwischenspeicher sieht sonst aus wie einer von eben.
  expect(screen.getByText(/2 Stunden alt/)).toBeInTheDocument()

  // **Und die Befunde** (§8).
  expect(screen.getByText(/unbekannte Seltenheit/)).toBeInTheDocument()
})

it('übernimmt nur die angehakten Felder', async () => {
  const ScryfallUebernehmen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({
    ScryfallVorschlag: () => Promise.resolve(einVorschlag()),
    ScryfallUebernehmen,
  })

  await holeVorschlag()

  // **Vorbelegt ist nichts** — §16, Regel 6: nichts wird still übernommen.
  await userEvent.click(await screen.findByLabelText(/rule text/))
  await userEvent.click(screen.getByRole('button', { name: texte.uebernehmen }))

  await waitFor(() => {
    expect(ScryfallUebernehmen).toHaveBeenCalledWith(
      '1',
      '5',
      'Lightning Bolt',
      ['rule text'],
      'stand-des-vorschlags',
    )
  })
})

it('sperrt das Übernehmen, solange nichts angehakt ist', async () => {
  zeige({ ScryfallVorschlag: () => Promise.resolve(einVorschlag()) })

  await holeVorschlag()

  expect(
    await screen.findByRole('button', { name: texte.uebernehmen }),
  ).toBeDisabled()
})

it('sagt es, wenn nichts zu ändern ist', async () => {
  zeige({
    ScryfallVorschlag: () =>
      Promise.resolve(
        einVorschlag({
          vorschlag: {
            aenderungen: [],
            befunde: [],
            herkunft: 'https://scryfall.com/card/lea/161',
            alterStunden: 0,
            stand: '',
          },
        }),
      ),
  })

  await holeVorschlag()

  expect(await screen.findByText(texte.nichtsZuTun)).toBeInTheDocument()
})

// **Die Meldung bleibt in diesem Kasten** (§6, und die Lehre aus #888): der
// Karteneditor zeigt eine Meldung nur am Feld, das sie verursacht hat — ein
// Abgleich verursacht keines, und die Meldung fiele still aus.
it('zeigt einen Fehler im Kasten', async () => {
  zeige({
    ScryfallVorschlag: () =>
      Promise.resolve({
        vorschlag: null,
        fehler: {
          code: 'nichtGefunden',
          meldung: 'Scryfall kennt die Karte nicht',
        },
      }),
  })

  await holeVorschlag()

  expect(
    await screen.findByText(/Scryfall kennt die Karte nicht/),
  ).toBeInTheDocument()
})

// **Ein Haken lässt sich auch wieder entfernen.**
//
// Ohne diesen Lauf ginge das Anhaken nur in eine Richtung, und niemand sähe
// es: ein einmal gesetzter Haken bliebe für immer.
it('nimmt einen Haken wieder weg', async () => {
  const ScryfallUebernehmen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({
    ScryfallVorschlag: () => Promise.resolve(einVorschlag()),
    ScryfallUebernehmen,
  })

  await holeVorschlag()

  const haken = await screen.findByLabelText(/rule text/)
  await userEvent.click(haken)
  await userEvent.click(screen.getByLabelText(/rarity/))
  await userEvent.click(haken)

  await userEvent.click(screen.getByRole('button', { name: texte.uebernehmen }))

  await waitFor(() => {
    // **Der Stand geht mit** (#1241) — er bindet den Vorschlag an die
    // Kartenwerte, gegen die verglichen wurde.
    expect(ScryfallUebernehmen).toHaveBeenCalledWith(
      '1',
      '5',
      'Lightning Bolt',
      ['rarity'],
      'stand-des-vorschlags',
    )
  })
})

// **Scheitert die Übernahme, bleibt der Vorschlag stehen** (§7, Regel 3).
//
// Wer die Meldung liest, soll sehen, worauf sie sich bezieht — und es noch
// einmal versuchen können, ohne neu zu holen.
it('zeigt einen Fehler der Übernahme im Kasten', async () => {
  zeige({
    ScryfallVorschlag: () => Promise.resolve(einVorschlag()),
    ScryfallUebernehmen: () =>
      Promise.resolve({
        fehler: { code: 'validierung', meldung: 'das Feld nimmt das nicht an' },
      }),
  })

  await holeVorschlag()

  await userEvent.click(await screen.findByLabelText(/rule text/))
  await userEvent.click(screen.getByRole('button', { name: texte.uebernehmen }))

  expect(
    await screen.findByText(/das Feld nimmt das nicht an/),
  ).toBeInTheDocument()

  // **Der Vorschlag steht noch da** — sonst müsste man neu holen.
  expect(screen.getByLabelText(/rule text/)).toBeInTheDocument()
})

// **Ein frisch geholter Stand sagt es, und ein leerer neuer Wert auch.**
//
// „Stand: 0 Stunden alt" wäre eine Zahl, die niemand braucht; und ein leeres
// Feld ohne Wort sähe aus, als fehlte die Zeile.
it('unterscheidet frisch von alt und leer von gefüllt', async () => {
  zeige({
    ScryfallVorschlag: () =>
      Promise.resolve(
        einVorschlag({
          vorschlag: {
            aenderungen: [{ feld: 'flavor text', alt: 'Etwas', neu: '' }],
            befunde: [],
            herkunft: 'https://scryfall.com/card/lea/161',
            alterStunden: 0,
            stand: '',
          },
        }),
      ),
  })

  await holeVorschlag()

  expect(await screen.findByText(texte.frisch)).toBeInTheDocument()
  expect(screen.getByText(`neu: ${texte.leer}`)).toBeInTheDocument()
})

// **Eine Antwort ohne Vorschlag und ohne Fehler zeigt nichts an.**
//
// Sie sollte nicht vorkommen — Go setzt bei einem Fehler `vorschlag: null`
// **und** einen Fehler. Der Zweig steht trotzdem da, weil die Bindung `null`
// erlaubt; ohne Lauf wäre er eine Behauptung darüber, was dann geschieht.
it('zeigt nichts bei einer Antwort ohne Vorschlag', async () => {
  zeige({
    ScryfallVorschlag: () => Promise.resolve({ vorschlag: null, fehler: null }),
  })

  await holeVorschlag()

  await waitFor(() => {
    expect(
      screen.queryByRole('button', { name: texte.uebernehmen }),
    ).not.toBeInTheDocument()
  })

  expect(screen.queryByText(texte.nichtsZuTun)).not.toBeInTheDocument()
})
