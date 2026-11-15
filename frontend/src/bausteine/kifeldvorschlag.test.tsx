// **Ein Vorschlag am Feld** — AI-05.AK-06
// (#2042).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIFeldvorschlag } from './KIFeldvorschlag'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { KIVorschlagAntwort } from '../antworten'
import type { KIVorschlagsbezug } from '../antwortenmarkt'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

function antwort(teil: Partial<KIVorschlagAntwort> = {}): KIVorschlagAntwort {
  return {
    kennung: 'v1',
    anbieter: 'apple',
    modell: 'system',
    felder: [{ feld: 'name', werte: ['Nebelhirsch'], bisher: 'Alt' }],
    befunde: [],
    ablauf: '2026-10-04T01:00:00Z',
    uebernommen: 0,
    belege: [],
    korpusstaende: [],
    einschraenkungen: [],
    weg: 'vorrang',
    ausgang: '',
    rueckfaelle: null,
    fehler: null,
    ...teil,
  }
}

function zeige(geaendert = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <KIFeldvorschlag
        karteID="4"
        feld="name"
        aufschrift="Name"
        werte={[{ feld: 'rule text', wert: 'Fliegend' }]}
        formen={new Map([['rule text', { art: 'text' }]])}
        geaendert={geaendert}
      />
    </Sprachraum>,
  )

  return geaendert
}

const knopf = () =>
  screen.getByRole('button', {
    name: einsetzen(eigene.kiFeldVorschlagenFuer, { feld: 'Name' }),
  })

// **Gefragt wird mit Vorrang und nur nach diesem Feld** — samt der Karte im
// Auftrag; übernommen wird nur dieses Feld, und erst auf Druck.
it('fragt mit Vorrang und übernimmt auf Druck', async () => {
  const gefragt = vi.fn<
    (karte: string, bezug: KIVorschlagsbezug) => Promise<KIVorschlagAntwort>
  >(() => Promise.resolve(antwort()))
  const uebernommen = vi.fn(() => Promise.resolve({ felder: 1, fehler: null }))
  belegeApi({ KIVorschlagen: gefragt, KIVorschlagUebernehmen: uebernommen })

  const geaendert = zeige()
  await userEvent.click(knopf())

  expect(await screen.findByText('Nebelhirsch')).toBeInTheDocument()
  expect(
    screen.getByText(
      einsetzen(eigene.kiFeldVon, {
        anbieter: 'apple/system',
        weg: eigene.kiFeldWegVorrang,
      }),
    ),
  ).toBeInTheDocument()

  const bezug = gefragt.mock.calls[0]?.[1]
  expect(bezug?.anbieter).toBe('vorrang')
  expect(bezug?.nur).toEqual(['name'])
  expect(bezug?.auftrag).toContain('rule text: Fliegend')
  expect(uebernommen).not.toHaveBeenCalled()

  await userEvent.click(
    screen.getByRole('button', { name: eigene.kiFeldUebernehmen }),
  )

  await vi.waitFor(() => {
    expect(geaendert).toHaveBeenCalledOnce()
  })
  expect(uebernommen).toHaveBeenCalledWith('v1', ['name'])
})

// **Der Flavortext bekommt seinen eigenen Satz.**
it('fragt den Flavortext mit seinem Satz', async () => {
  const gefragt = vi.fn<
    (karte: string, bezug: KIVorschlagsbezug) => Promise<KIVorschlagAntwort>
  >(() =>
    Promise.resolve(
      antwort({
        felder: [{ feld: 'flavor text', werte: ['Still.'], bisher: '' }],
      }),
    ),
  )
  belegeApi({ KIVorschlagen: gefragt })

  render(
    <Sprachraum sprache="de">
      <KIFeldvorschlag
        karteID="4"
        feld="flavor text"
        aufschrift="Flavortext"
        werte={[]}
        formen={new Map()}
        geaendert={vi.fn()}
      />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.kiFeldVorschlagenFuer, { feld: 'Flavortext' }),
    }),
  )

  expect(await screen.findByText('Still.')).toBeInTheDocument()
  expect(gefragt.mock.calls[0]?.[1].auftrag).toBe(eigene.kiFeldAuftragFlavor)
})

// **Eine Ablehnung ist ein eigener Ausgang** — kein leerer Vorschlag, kein
// Fehler; sie sagt, was das Modell sagte.
it('zeigt eine Ablehnung als eigenen Ausgang', async () => {
  belegeApi({
    KIVorschlagen: () =>
      Promise.resolve(
        antwort({
          kennung: '',
          felder: [],
          ausgang: 'abgelehnt',
          befunde: ['Nein.'],
        }),
      ),
  })

  zeige()
  await userEvent.click(knopf())

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFeldAbgelehnt, { grund: 'Nein.' }),
    ),
  ).toBeInTheDocument()
  expect(screen.queryByRole('alert')).toBeNull()
})

// **Kein Wert für dieses Feld ist eine Auskunft**, kein leerer Kasten.
it('sagt, wenn für das Feld nichts kam', async () => {
  belegeApi({
    KIVorschlagen: () =>
      Promise.resolve(antwort({ felder: [], weg: 'rueckfall' })),
  })

  zeige()
  await userEvent.click(knopf())

  expect(await screen.findByText(eigene.kiFeldNichts)).toBeInTheDocument()
})

// **Ein Fehler steht am Feld** (§6) — beim Fragen wie beim Übernehmen.
it('zeigt Fehler am Feld', async () => {
  const fehler = {
    code: 'validierung' as const,
    schluessel: 'x',
    meldung: 'Kaputt.',
  }
  belegeApi({
    KIVorschlagen: () => Promise.resolve(antwort()),
    KIVorschlagUebernehmen: () => Promise.resolve({ felder: 0, fehler }),
  })

  const geaendert = zeige()
  await userEvent.click(knopf())
  await userEvent.click(
    await screen.findByRole('button', { name: eigene.kiFeldUebernehmen }),
  )

  expect(await screen.findByRole('alert')).toBeInTheDocument()
  expect(geaendert).not.toHaveBeenCalled()
})

it('zeigt einen Fehler beim Fragen', async () => {
  belegeApi({
    KIVorschlagen: () =>
      Promise.resolve(
        antwort({
          fehler: {
            code: 'validierung',
            schluessel: 'x',
            meldung: 'Kein Anbieter.',
          },
        }),
      ),
  })

  zeige()
  await userEvent.click(knopf())

  expect(await screen.findByRole('alert')).toBeInTheDocument()
  // **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742).
  expect(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.kieinrichtung.oeffnen,
    }),
  ).toBeInTheDocument()
})

// **Ein unbekannter Weg steht als Wort da** — und fällt nicht weg.
it('nennt einen unbekannten Weg wörtlich', async () => {
  belegeApi({
    KIVorschlagen: () => Promise.resolve(antwort({ weg: 'gewaehlt' })),
  })

  zeige()
  await userEvent.click(knopf())

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFeldVon, {
        anbieter: 'apple/system',
        weg: eigene.kiFeldWegGewaehlt,
      }),
    ),
  ).toBeInTheDocument()
})

it('nennt einen Weg, den es nicht kennt, wörtlich', async () => {
  belegeApi({
    KIVorschlagen: () => Promise.resolve(antwort({ weg: 'irgendwie' })),
  })

  zeige()
  await userEvent.click(knopf())

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFeldVon, {
        anbieter: 'apple/system',
        weg: 'irgendwie',
      }),
    ),
  ).toBeInTheDocument()
})

// **Ein leerer Wert ist kein Vorschlag** — gemessen am echten Gerätemodell
// (Vermerk 0166): einmal von vier Läufen kam ein Name ohne Text.
it('nimmt einen leeren Wert nicht als Vorschlag', async () => {
  belegeApi({
    KIVorschlagen: () =>
      Promise.resolve(
        antwort({ felder: [{ feld: 'name', werte: ['  '], bisher: '' }] }),
      ),
  })

  zeige()
  await userEvent.click(knopf())

  expect(await screen.findByText(eigene.kiFeldNichts)).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: eigene.kiFeldUebernehmen }),
  ).toBeNull()
})

// **Warum nichts kam, steht da** — die Befunde der Schemaprüfung, sonst sähe
// jemand nur „kein Vorschlag" und müsste raten.
it('nennt die Befunde, wenn für das Feld nichts kam', async () => {
  belegeApi({
    KIVorschlagen: () =>
      Promise.resolve(antwort({ felder: [], befunde: ['name: zu lang'] })),
  })

  zeige()
  await userEvent.click(knopf())

  expect(await screen.findByText(/name: zu lang/)).toBeInTheDocument()
})
