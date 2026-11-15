// **Die verfügbaren Namen** — Werkstück 1 und 4 von Phase 3.3
// (#1164).
//
// > Für jemanden, der MSE-Script **nicht** kennt, ist ein leeres Feld keine
// > Einladung.
//
// **Aus der wirklichen Umgebung**, nicht aus einer gepflegten Liste: was die
// Vorlage dazulegt, steht mit darin. Eine Liste daneben liefe an dem Tag
// auseinander, an dem die Bibliothek eine Funktion dazubekommt — und böte
// etwas an, das nicht läuft.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Skriptkonsole } from './Skriptkonsole'

const texte = kataloge.de.seiten.karten.konsole

describe('Die verfügbaren Namen', () => {
  afterEach(raeumeApi)

  // **Auf Wunsch und nicht beim Aufschlagen**: es sind neunzig und mehr, und
  // wer nur `1 + 1` rechnen will, hat nicht danach gefragt.
  //
  // **Der Wunsch ist seit UXR-08 das Aufklappen** des Kastens „Verfügbare
  // Namen" (#1916) — bis dahin ein eigener Knopf „Namen anzeigen", der sich
  // nicht wieder schließen ließ.
  it('werden erst auf Klick geholt', async () => {
    const gefragt = vi.fn(() =>
      Promise.resolve({ namen: ['card', 'length', 'set'], fehler: null }),
    )

    belegeApi({ Skriptnamen: gefragt })

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    expect(gefragt).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: texte.namen }))

    // **Die Zahl steht neben dem Feld, das sie ändert** — sie kommt aus
    // `Namensliste`, seit die Liste eingrenzt und deckelt.
    expect(
      await screen.findByText('Es stehen 3 von 3 da — tippe, um einzugrenzen.'),
    ).toBeInTheDocument()
    expect(gefragt).toHaveBeenCalledWith('7')
  })

  // **Ein Klick setzt den Namen ein.** Abschreiben ist der Weg, auf dem sich
  // ein Tippfehler einschleicht — und ein Tippfehler in einem Namen sieht aus
  // wie ein Fehler der Sprache.
  it('setzen sich per Klick in den Ausdruck', async () => {
    belegeApi({
      Skriptnamen: () =>
        Promise.resolve({ namen: ['card', 'length'], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: texte.namen }))
    await userEvent.click(await screen.findByRole('button', { name: 'length' }))

    expect(screen.getByLabelText(texte.ausdruck)).toHaveValue('length')
  })

  // **Ein Fehler beim Holen macht die Konsole nicht kaputt.**
  //
  // Sie ist zum Rechnen da; eine fehlende Liste ist ein Mangel und kein
  // Grund, das Feld zu sperren (§8).
  it('melden einen Fehler und lassen die Konsole stehen', async () => {
    belegeApi({
      Skriptnamen: () =>
        Promise.resolve({
          namen: [],
          fehler: {
            code: 'nichtGefunden' as const,
            schluessel: 'nichtGefunden',
            meldung: 'die Karte gibt es nicht.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Skriptkonsole karteID="7" zuDenKarten={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: texte.namen }))

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.getByLabelText(texte.ausdruck)).toBeEnabled()
  })
})
