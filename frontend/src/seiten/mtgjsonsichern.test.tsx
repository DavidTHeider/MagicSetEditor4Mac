// Die Ausfuhr eines Sets als MTG JSON.
//
// **Über die Lasche `Ausgeben` und nicht über die Setliste**
// (#677): der
// Knopf stand an jeder Setzeile, zwischen drei anderen. Jetzt steht er dort,
// wo alle Ausgänge stehen — mit einem Satz daneben, wofür er ist.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgeben } from './Ausgeben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probeSet(teile: Partial<SetDTO> = {}): SetDTO {
  return {
    id: '4',
    name: 'Mein Probeset',
    spiel: 'magic',
    stil: 'm15',
    herkunft: '/pfad/probeset.mse-set',
    karten: 2,
    herkunftFehlt: false,
    ...teile,
  }
}

function ausfuhr(teile: Record<string, unknown> = {}) {
  return Promise.resolve({
    pfad: '/tmp/PRB.json',
    karten: 2,
    herkunftFehlt: false,
    befunde: [],
    abgebrochen: false,
    ...teile,
  })
}

// zeige schlägt den Ausgabeassistenten auf und wählt MTG JSON.
//
// **Erst gewählt, dann gesichert** — seit UXR-09
// (#1917)
// steht der Knopf nicht mehr neben vier anderen Ausgängen, sondern in der
// Fläche des Ziels, das unter „Wohin?" gewählt ist.
async function zeige() {
  render(
    <Sprachraum sprache="de">
      <Ausgeben setID="4" seite="ausgeben" wechseln={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('radio', {
      name: texte.seiten.sets.mtgjsonUeberschrift,
    }),
  )
}

describe('Als MTG JSON sichern', () => {
  // **Das Spiel kommt aus dem Set**, nicht aus einer Auswahl daneben: welche
  // Zuordnung gilt, entscheidet das Set, aus dem geschrieben wird.
  it('schreibt das gewählte Set mit seinem Spiel', async () => {
    const MTGJSONSchreiben = vi.fn(() => ausfuhr())

    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      MTGJSONSchreiben,
    })

    await zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.seiten.sets.mtgjsonSichern} — Mein Probeset`,
      }),
    )

    expect(MTGJSONSchreiben).toHaveBeenCalledWith('4', 'magic')
    expect(
      await screen.findByText(/\/tmp\/PRB\.json/, { exact: false }),
    ).toBeInTheDocument()
  })

  // **Was nicht abbildbar war, steht beim Erfolg** — wie beim Öffnen (§8).
  it('nennt, was beim Schreiben nicht abbildbar war', async () => {
    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      MTGJSONSchreiben: () =>
        ausfuhr({
          befunde: ['PRB → super type: MSE führt beides in einem Feld'],
        }),
    })

    await zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.seiten.sets.mtgjsonSichern} — Mein Probeset`,
      }),
    )

    expect(
      await screen.findByText(texte.seiten.sets.mtgjsonVerluste),
    ).toBeInTheDocument()
    expect(
      screen.getByText('PRB → super type: MSE führt beides in einem Feld'),
    ).toBeInTheDocument()
  })

  // Ein abgebrochener Dialog ist kein Fehler und keine Meldung.
  it('sagt nichts, wenn niemand einen Ort gewählt hat', async () => {
    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      MTGJSONSchreiben: () => ausfuhr({ abgebrochen: true }),
    })

    await zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.seiten.sets.mtgjsonSichern} — Mein Probeset`,
      }),
    )

    expect(
      screen.queryByText(texte.seiten.sets.mtgjsonGesichert, { exact: false }),
    ).not.toBeInTheDocument()
  })

  // Ein Fehler erscheint als Meldung, nicht als leerer Bildschirm.
  it('meldet einen Fehler beim Schreiben', async () => {
    belegeApi({
      // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
      EinstellungSetzen: () => Promise.resolve({}),
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      MTGJSONSchreiben: () =>
        Promise.resolve({
          pfad: '',
          karten: 0,
          befunde: [],
          abgebrochen: false,
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'kein Schreibrecht',
            feld: '',
          },
        }),
    })

    await zeige()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.seiten.sets.mtgjsonSichern} — Mein Probeset`,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'kein Schreibrecht',
    )
  })
})
