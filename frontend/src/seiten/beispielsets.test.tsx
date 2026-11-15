// **Ein Beispielset ist einen Klick entfernt** —
// #702.
//
// Der erste Weg einer frisch geladenen Anwendung führte durch einen
// Dateidialog — auf einen Rechner, auf dem noch kein `.mse-set` liegt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { belegeApi, raeumeApi } from '../test/api'
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

function zeige(
  sets: { name: string; pfad: string }[],
  oeffnen = vi.fn(() =>
    Promise.resolve({
      angelegt: ['9'],
      unvollstaendig: false,
      setID: '9',
      karten: 107,
      befunde: [],
      schonOffen: false,
      abgebrochen: false,
      fehler: null,
    }),
  ),
) {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [] }),
    Beispielsets: () => Promise.resolve({ sets }),
    BeispielsetOeffnen: oeffnen,
  })

  const gewaehlt = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={gewaehlt}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return { oeffnen, gewaehlt }
}

describe('Die mitgelieferten Beispielsets', () => {
  it('stehen mit einem Knopf je Set da', async () => {
    zeige([
      { name: 'Vanguard', pfad: '/x/Vanguard.mse-set' },
      { name: 'Probe', pfad: '/x/Probe.mse-set' },
    ])
    await einlesenAufklappen()

    expect(
      await screen.findByRole('button', { name: 'Vanguard' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Probe' })).toBeInTheDocument()
  })

  // **Derselbe Ausgang wie beim Dateidialog**: das Set liegt danach im
  // Bestand und meldet sich mit seiner Kartenzahl. Ein Sprung mitten hinein
  // wäre ein anderes Verhalten als bei `Set öffnen …` — zwei Wege, die
  // dasselbe tun und sich verschieden anfühlen, sind einer zu viel.
  it('schlagen das Set auf, ohne nach einer Datei zu fragen', async () => {
    const nutzer = userEvent.setup()
    const { oeffnen } = zeige([
      { name: 'Vanguard', pfad: '/x/Vanguard.mse-set' },
    ])
    await einlesenAufklappen()

    await nutzer.click(await screen.findByRole('button', { name: 'Vanguard' }))

    // **Gewartet wird auf das Abschlusssignal** (§3): der Spion läuft erst
    // nach der Antwort, und die Zusicherung steht danach.
    await waitFor(() => {
      expect(oeffnen).toHaveBeenCalledWith('/x/Vanguard.mse-set')
    })

    expect(await screen.findByRole('status')).toHaveTextContent('107')
  })

  // **Ohne Bündel steht der Abschnitt nicht da.** Ein leerer Kasten
  // „Beispielsets" wäre ein Versprechen ohne Inhalt — und beim Entwickeln der
  // Normalfall.
  it('bleiben weg, wenn keines mitkommt', async () => {
    zeige([])
    await einlesenAufklappen()

    expect(
      await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
    ).toBeInTheDocument()

    expect(
      screen.queryByText(texte.seiten.sets.beispielsets),
    ).not.toBeInTheDocument()
  })
})
