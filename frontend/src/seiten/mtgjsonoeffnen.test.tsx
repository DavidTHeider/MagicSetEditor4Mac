// Der Weg für MTG JSON in der Setliste.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { belegeApi, raeumeApi } from '../test/api'
import { abgebrochen } from '../test/laufauftrag'
import { Sets } from './Sets'

// einlesenAufklappen öffnet den Kasten mit den Einfuhrwegen.
//
// **Seit dem 2026-09-03 sind sie zugeklappt** — Werkstück C von Phase 4.10
// (#1325):
// vier Wege standen über der Setliste, und das erste Set begann bei 915
// Bildpunkten, also unter der Falz. Was diese Tests prüfen, steht darin und
// ändert sich nicht; erreicht wird es über einen Klick mehr.
async function einlesenAufklappen() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.sets.einlesenTitel,
    }),
  )
}

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function leereListe() {
  return Promise.resolve({ sets: [] })
}

function ausgang(teile: Record<string, unknown> = {}) {
  return Promise.resolve({
    setID: '1',
    karten: 2,
    befunde: [],
    schonOffen: false,
    angelegt: [],
    unvollstaendig: false,
    abgebrochen: false,
    ...teile,
  })
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )
}

describe('MTG JSON öffnen', () => {
  // **Das Spiel wird gewählt, nicht geraten.** MTG JSON beschreibt
  // Magic-Karten, nicht MSE-Kartenfelder; dazwischen liegt eine Zuordnung, und
  // die gehört zum Spiel.
  it('öffnet eine Datei mit dem gewählten Spiel', async () => {
    const MTGJSONOeffnen = vi.fn(() => ausgang())

    belegeApi({
      Sets: leereListe,
      MTGJSONSpiele: () => Promise.resolve(['magic', 'probespiel']),
      MTGJSONOeffnen,
    })

    zeige()
    await einlesenAufklappen()

    const wahl = await screen.findByLabelText(texte.seiten.sets.mtgjsonSpiel)
    await userEvent.selectOptions(wahl, 'probespiel')

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.sets.mtgjsonDatei }),
    )

    expect(MTGJSONOeffnen).toHaveBeenCalledWith('probespiel')
  })

  // **Ein Abbruch ist keine Meldung wert** — und schon gar keine
  // Fehlermeldung. Seit *Set öffnen …* über die Vorschau geht (FMT-03), ist
  // der Dialog von MTG JSON der Weg, der in der Setliste abbrechen kann.
  it('sagt nichts, wenn niemand eine Datei wählt', async () => {
    const MTGJSONOeffnen = vi.fn(() =>
      ausgang({ setID: '', abgebrochen: true }),
    )

    belegeApi({
      Sets: leereListe,
      MTGJSONSpiele: () => Promise.resolve(['magic']),
      MTGJSONOeffnen,
    })

    zeige()
    await einlesenAufklappen()
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.mtgjsonDatei,
      }),
    )

    // **Erst der Beleg, dass der Weg gegangen ist** (§3): gerufen, und der
    // Knopf ist wieder frei — dann erst die Verneinung.
    await waitFor(() => {
      expect(MTGJSONOeffnen).toHaveBeenCalledOnce()
      expect(
        screen.getByRole('button', { name: texte.seiten.sets.mtgjsonDatei }),
      ).toBeEnabled()
    })
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.queryByText(/Karten? geöffnet$/)).not.toBeInTheDocument()
  })

  // Ein Verzeichnis ist die andere Gestalt von MTG JSON — und unter macOS ein
  // anderer Dialog.
  it('öffnet ein Verzeichnis über einen eigenen Weg', async () => {
    const MTGJSONVerzeichnisOeffnen = vi.fn(() => ausgang({ karten: 5 }))

    belegeApi({
      Sets: leereListe,
      MTGJSONSpiele: () => Promise.resolve(['magic']),
      MTGJSONVerzeichnisOeffnen,
    })

    zeige()
    await einlesenAufklappen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.mtgjsonVerzeichnis,
      }),
    )

    expect(MTGJSONVerzeichnisOeffnen).toHaveBeenCalledWith('magic')
  })

  // **Ohne Zuordnung wird nichts angeboten, das nicht ginge.** Eine
  // Schaltfläche, die zuverlässig eine Fehlermeldung liefert, ist keine.
  it('erklärt, wenn es für kein Spiel eine Zuordnung gibt', async () => {
    belegeApi({ Sets: leereListe, MTGJSONSpiele: () => Promise.resolve([]) })

    zeige()
    await einlesenAufklappen()

    expect(
      await screen.findByText(texte.seiten.sets.mtgjsonKeinSpiel),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: texte.seiten.sets.mtgjsonDatei }),
    ).not.toBeInTheDocument()
  })

  // **Befunde und Erfolg schließen einander nicht aus** (§8) — auch hier.
  it('nennt die Befunde eines Imports', async () => {
    belegeApi({
      Sets: leereListe,
      MTGJSONSpiele: () => Promise.resolve(['magic']),
      MTGJSONOeffnen: () =>
        ausgang({ befunde: ['PRB.json:PRB → Karte 1 → uuid: unbekannt'] }),
    })

    zeige()
    await einlesenAufklappen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.mtgjsonDatei,
      }),
    )

    expect(
      await screen.findByText('PRB.json:PRB → Karte 1 → uuid: unbekannt'),
    ).toBeInTheDocument()
  })

  // Ein Fehler beim Nachsehen der Spiele nimmt nur den JSON-Weg, nicht die
  // Seite: eine `.mse-set` lässt sich weiterhin öffnen.
  it('hält die Seite, wenn die Spiele nicht zu holen sind', async () => {
    belegeApi({
      Sets: leereListe,
      MTGJSONSpiele: () => Promise.reject(new Error('nicht erreichbar')),
    })

    zeige()
    await einlesenAufklappen()

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.seiten.sets.setOeffnen }),
      ).toBeInTheDocument()
    })

    expect(
      screen.queryByRole('button', { name: texte.seiten.sets.mtgjsonDatei }),
    ).not.toBeInTheDocument()
  })
})

// **Ein halb geglückter Import nennt, was er zurückgelassen hat** (#155).
//
// # Warum das zählt
//
// Jedes Set einer AllSets-Datei liegt in einer eigenen Transaktion. Was vor
// dem Fehler durchkam, steht im Bestand — und wer das nicht erfährt, findet
// die Sets später in der Liste und hält sie für vollständig.
describe('ein abgebrochener Import', () => {
  it('nennt die Sets, die trotzdem entstanden sind, und lädt die Liste nach', async () => {
    const Sets = vi
      .fn()
      .mockResolvedValueOnce({ sets: [] })
      .mockResolvedValue({
        sets: [
          {
            id: '1',
            name: 'Halb da',
            spiel: 'magic',
            stil: '',
            herkunft: '/tmp/alle.json',
            karten: 2,
            herkunftFehlt: false,
          },
        ],
      })

    belegeApi({
      Sets,
      MTGJSONSpiele: () => Promise.resolve(['magic']),
      MTGJSONOeffnen: () =>
        Promise.resolve({
          setID: '',
          karten: 0,
          befunde: [],
          schonOffen: false,
          angelegt: ['1'],
          unvollstaendig: true,
          abgebrochen: false,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            meldung: 'Karte 1: Ein Feld braucht einen Namen.',
          },
        }),
    })

    zeige()
    await einlesenAufklappen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.sets.mtgjsonDatei,
      }),
    )

    // **Der Fehler bleibt ein Fehler** — und der Teilerfolg steht daneben.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Ein Feld braucht einen Namen.',
    )
    // **Über alle Standbereiche und nicht über „den" einen** (#1321). Die
    // Seite hat mehr als einen: neben dem Ergebnis steht der Ladehinweis von
    // Scryfall, und der ist noch da, solange seine Antwort unterwegs ist.
    // `getByRole('status')` war damit immer mehrdeutig — sichtbar wurde es
    // erst, als `test/api.ts` aufhörte, synchron zu antworten (§3).
    expect(
      screen
        .getAllByRole('status')
        .map((einer) => einer.textContent)
        .join(' | '),
    ).toContain('1 Sets liegen trotzdem im Bestand')

    // **Und die Liste zeigt sie.** Ohne das Nachladen im `finally` stünde dort
    // der Stand von vorhin.
    expect(await screen.findByText('Halb da')).toBeInTheDocument()
  })
})

// **Ein angehaltenes MTG JSON ist keine Fehlermeldung** (FMT-03, #1692) —
// die Sets davor sind ganz, und der Teilerfolg sagt, wie viele.
it('nennt nach einem Anhalten die Sets davor und keinen Fehler', async () => {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [] }),
    MTGJSONSpiele: () => Promise.resolve(['magic']),
    MTGJSONOeffnen: () =>
      Promise.resolve({
        setID: '',
        karten: 0,
        befunde: [],
        schonOffen: false,
        abgebrochen: false,
        angelegt: ['1', '2'],
        unvollstaendig: true,
        fehler: abgebrochen,
      }),
  })

  zeige()
  await einlesenAufklappen()
  await userEvent.click(
    await screen.findByRole('button', { name: texte.seiten.sets.mtgjsonDatei }),
  )

  expect(
    await screen.findByText(/2 Sets liegen trotzdem im Bestand/),
  ).toBeInTheDocument()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})
