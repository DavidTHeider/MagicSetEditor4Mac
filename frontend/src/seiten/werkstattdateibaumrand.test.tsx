// **Die Ränder des Dateibaums** — Abbrechen, Filter und späte Antworten
// (FMT-02.AK-01, #1691).
// Die Hauptwege stehen in `werkstattdateibaum.test.tsx`.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, raeumeApi } from '../test/api'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

import type { Api } from '../api'

const texte = kataloge.de.seiten.vorlagen
const pfad = '/pakete/probe.mse-style'

function zeige(eigene: Partial<Api> = {}) {
  const lesen = vi.fn((_paket: string, datei: string) =>
    Promise.resolve({ inhalt: `# ${datei}\n`, stand: 'alt', fehler: null }),
  )

  belegeWerkstatt({
    Paketdateien: () =>
      Promise.resolve({
        dateien: ['style', 'card_fields'],
        beiwerk: ['rahmen.png'],
        fehler: null,
      }),
    PaketdateiLesen: lesen,
    ...eigene,
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )

  return { lesen }
}

async function oeffne(datei: string) {
  await userEvent.click(await screen.findByRole('button', { name: datei }))
  await screen.findByLabelText(texte.inhalt)
}

function aktionen() {
  return within(screen.getByRole('group', { name: texte.dateiaktionen }))
}

// handgriff wählt einen Handgriff aus dem Menü »Datei«. **Seit UXR-18
// (#1926) stehen die vier hinter einem Knopf** und nicht mehr als Reihe über
// dem Editor — sie sind selten, und der Editor ist es nicht.
async function handgriff(name: string) {
  await userEvent.click(aktionen().getByRole('button', { name: texte.datei }))
  await userEvent.click(aktionen().getByRole('menuitem', { name }))
}

describe('Die Ränder des Dateibaums', () => {
  afterEach(raeumeApi)

  it('bricht jede Eingabe ab — mit Esc, mit dem Knopf und an der Rückfrage', async () => {
    zeige()
    await oeffne('card_fields')

    await handgriff(texte.dateiNeu)
    await userEvent.type(screen.getByLabelText(texte.dateiName), 'x{Escape}')
    expect(screen.queryByLabelText(texte.dateiName)).not.toBeInTheDocument()

    await handgriff(texte.dateiUmbenennen)
    await userEvent.click(
      aktionen().getByRole('button', { name: texte.dateiAbbrechen }),
    )
    expect(
      screen.queryByLabelText(texte.dateiNeuerName),
    ).not.toBeInTheDocument()

    await handgriff(texte.dateiLoeschen)
    await userEvent.click(
      aktionen().getByRole('button', { name: texte.dateiAbbrechen }),
    )
    expect(screen.queryByText(texte.dateiLoeschenFrage)).not.toBeInTheDocument()
    expect(
      aktionen().getByRole('button', { name: texte.datei }),
    ).toBeInTheDocument()
  })

  it('legt auch per Mausklick an — und sperrt den Knopf ohne Namen', async () => {
    const anlegen = vi.fn(() => Promise.resolve({ stand: 'neu', fehler: null }))
    zeige({ PaketdateiAnlegen: anlegen })
    await oeffne('card_fields')

    await handgriff(texte.dateiNeu)
    const knopf = aktionen().getByRole('button', { name: texte.dateiAnlegen })
    expect(knopf).toBeDisabled()

    await userEvent.type(screen.getByLabelText(texte.dateiName), ' neu ')
    await userEvent.click(knopf)

    expect(anlegen).toHaveBeenCalledWith(pfad, 'neu', '')
  })

  it('findet Beiwerk über den Filter des Baums', async () => {
    zeige()
    await screen.findByRole('button', { name: 'card_fields' })

    await userEvent.type(screen.getByLabelText(texte.baumSuche), 'rahmen')

    expect(
      screen.getByRole('button', { name: `rahmen.png — ${texte.beiwerk}` }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'card_fields' }),
    ).not.toBeInTheDocument()
  })

  // **Eine späte Antwort wählt nichts um** (§7, Regel 2) — wer inzwischen
  // eine andere Datei aufgeschlagen hat, bleibt dort, für Erfolg und Fehler.
  it.each([
    ['eine geglückte', true],
    ['eine gescheiterte', false],
  ])('lässt %s späte Antwort die Auswahl stehen', async (_art, glueckt) => {
    let fertig: () => void = () => undefined
    const { lesen } = zeige({
      PaketdateiUmbenennen: () =>
        new Promise((erfuellen) => {
          fertig = () => {
            erfuellen(
              glueckt
                ? { fehler: null }
                : {
                    fehler: {
                      code: 'konflikt' as const,
                      schluessel: '',
                      meldung: 'klemmt',
                      werte: {},
                    },
                  },
            )
          }
        }),
    })
    await oeffne('card_fields')

    await handgriff(texte.dateiUmbenennen)
    await userEvent.type(
      screen.getByLabelText(texte.dateiNeuerName),
      'x{Enter}',
    )
    await userEvent.click(screen.getByRole('button', { name: 'style' }))
    await waitFor(() => {
      expect(lesen).toHaveBeenLastCalledWith(pfad, 'style')
    })

    // **Und was an der neuen Datei offen ist, bleibt offen** — der Zustand
    // hat einen Platz, und eine späte Antwort darf ihn nicht überschreiben.
    await handgriff(texte.dateiNeu)

    fertig()
    await antwortAbwarten()

    expect(lesen).toHaveBeenLastCalledWith(pfad, 'style')
    expect(aktionen().queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByLabelText(texte.dateiName)).toBeInTheDocument()
  })
})
