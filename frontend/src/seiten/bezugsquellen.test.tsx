// Die Bezugsquellen —
// #263.
//
// Der Abnahmekasten lautet: „die Paketseite bietet sie an — **ein Klick statt
// einer Adresse**". Was er nicht sagt und was trotzdem gilt: der Klick
// überspringt die **Zustimmung** nicht (#139, #56).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import {
  belegeSeite,
  oeffneBezug,
  zeige as zeigeSeite,
} from '../test/paketseite'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

const quellen = [
  {
    name: 'Basic-M15-Magic-Pack',
    adresse: 'https://github.com/MagicSetEditorPacks/Basic-M15-Magic-Pack',
    stand: 'df1bf728440c3b3cfb0ec38a3d5b484d39c35cad',
    unterverzeichnis: '',
    schluessel: 'basicM15',
    megabyte: 316,
  },
  {
    name: 'M15-Magic-Pack',
    adresse: 'https://github.com/MagicSetEditorPacks/M15-Magic-Pack',
    stand: '3f7f8ceed0c1fcc0d193d752c646aaaea7bece85',
    unterverzeichnis: '',
    schluessel: 'm15',
    megabyte: 855,
  },
]

// zeige schlägt die Seite auf und öffnet die Schublade — seit UXR-17
// (#1925) stehen die Quellen hinter *Paket beziehen …* im Kopf.
async function zeige(belegung: Partial<Api> = {}) {
  belegeSeite({
    Bezugsquellen: vi.fn(() => Promise.resolve({ quellen })),
    ...belegung,
  })
  zeigeSeite()
  await oeffneBezug()
}

describe('Die Bezugsquellen', () => {
  it('zeigt jede Quelle mit Name, Größe und Satz', async () => {
    await zeige()

    // **`^M15` und nicht `M15`**: `Basic-M15-Magic-Pack` enthält denselben
    // Text, und beide sind Knöpfe.
    const kachel = await screen.findByRole('button', {
      name: /^M15-Magic-Pack/,
    })

    expect(kachel).toHaveTextContent('855')
    expect(kachel).toHaveTextContent(texte.seiten.pakete.quellen.m15)
  })

  // **Das Formular beginnt leer.** Bis #263 stand eine Vorgabe darin, die sich
  // nicht erklärte — und deren `unterverzeichnis: 'data'` die Schriften
  // zurückließ.
  it('beginnt mit einem leeren Formular', async () => {
    await zeige()

    await screen.findByRole('button', { name: /^M15-Magic-Pack/ })

    expect(screen.getByLabelText(texte.seiten.pakete.bezugAdresse)).toHaveValue(
      '',
    )
    expect(screen.getByLabelText(texte.seiten.pakete.bezugStand)).toHaveValue(
      '',
    )

    // **Adresse, Stand, Verzeichnis und Name sind Kennungen** — die
    // Rechtschreibkorrektur von macOS machte aus einem Stand ein Wort.
    for (const feld of [
      texte.seiten.pakete.bezugAdresse,
      texte.seiten.pakete.bezugStand,
      texte.seiten.pakete.bezugUnterverzeichnis,
      texte.seiten.pakete.bezugName,
    ]) {
      expect(screen.getByLabelText(feld)).toHaveAttribute('spellcheck', 'false')
    }
  })

  it('trägt eine Quelle mit einem Klick ein', async () => {
    await zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: /^M15-Magic-Pack/ }),
    )

    expect(screen.getByLabelText(texte.seiten.pakete.bezugAdresse)).toHaveValue(
      'https://github.com/MagicSetEditorPacks/M15-Magic-Pack',
    )
    expect(screen.getByLabelText(texte.seiten.pakete.bezugStand)).toHaveValue(
      '3f7f8ceed0c1fcc0d193d752c646aaaea7bece85',
    )
    expect(screen.getByLabelText(texte.seiten.pakete.bezugName)).toHaveValue(
      'M15-Magic-Pack',
    )
  })

  // **Ohne Unterverzeichnis — und das ist der Kern von #263.**
  //
  // Die Packs tragen ihre Vorlagen unter `data` und ihre Schriften daneben
  // (`Magic - Fonts`). Wer nur `data` holt, bekommt Vorlagen ohne Schriften,
  // und jede Karte zeigt „Schrift fehlt", ohne dass jemand sähe, warum.
  it('lässt das Unterverzeichnis leer, damit die Schriften mitkommen', async () => {
    await zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: /^M15-Magic-Pack/ }),
    )

    expect(
      screen.getByLabelText(texte.seiten.pakete.bezugUnterverzeichnis),
    ).toHaveValue('')
  })

  // **Ein Klick bezieht nicht.** Er trägt ein; danach steht dieselbe Frage wie
  // vorher — „Herkunft anzeigen", dann „Beziehen". Das ist die Zusage aus #139
  // und #56, und ohne diesen Test wäre sie eine Absichtserklärung.
  it('bezieht nichts, sondern trägt nur ein', async () => {
    const PaketBeziehen = vi.fn()
    const PaketHerkunft = vi.fn()

    await zeige({ PaketBeziehen, PaketHerkunft })

    await userEvent.click(
      await screen.findByRole('button', { name: /^M15-Magic-Pack/ }),
    )

    expect(PaketHerkunft).not.toHaveBeenCalled()
    expect(PaketBeziehen).not.toHaveBeenCalled()
  })

  // **Eine gegebene Zustimmung wird von einem Quellenwechsel hinfällig**
  // (§7, Regel 1). Ohne das stünde der Hinweis zu Quelle A über einem
  // Formular mit Quelle B — und „Beziehen" holte B.
  it('räumt eine gegebene Zustimmung weg', async () => {
    await zeige({
      PaketHerkunft: vi.fn(() =>
        Promise.resolve({ quelle: 'irgendwo', stand: 'abc' }),
      ),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: /^Basic-M15-Magic-Pack/ }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )

    expect(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.bezugStarten,
      }),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: /^M15-Magic-Pack/ }),
    )

    expect(
      screen.queryByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    ).not.toBeInTheDocument()
  })

  it('meldet einen Fehler beim Laden der Quellen am Ort', async () => {
    await zeige({
      Bezugsquellen: vi.fn(() =>
        Promise.resolve({
          quellen: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Die Liste schweigt.',
          },
        }),
      ),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Die Liste schweigt.',
    )
  })
})
