// Die Paketseite: Herkunft, Zustimmung und die Absagen.
//
// **Zwei Dateien, weil eine über 400 Zeilen ginge** (§2). Hier stehen der Weg
// bis zur Zustimmung und die Meldungen; die Läufe, die sich überholen, stehen
// in `paketlauf.test.tsx`.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { belegeSeite, oeffneBezug, texte, zeige } from '../test/paketseite'
import { kurzform } from '../sprache/kurzform'

afterEach(() => {
  raeumeApi()
})

// fuelleAus tippt die vier Felder — **nacheinander**, wie jemand es täte.
//
// **Erst die Schublade** — seit UXR-17 (#1925) steht der Bezug hinter
// *Paket beziehen …* im Kopf der Seite.
async function fuelleAus() {
  await oeffneBezug()

  // **Erst leeren, dann tippen.** Seit #263 beginnt das Formular leer, und das
  // `clear` wäre entbehrlich — es bleibt trotzdem stehen: es macht den Test
  // unabhängig davon, ob eine Vorgabe da ist oder nicht, und genau diese
  // Abhängigkeit hat ihn beim Wegfall der Vorgabe rot gemacht.
  await userEvent.clear(screen.getByLabelText(texte.seiten.pakete.bezugAdresse))
  await userEvent.type(
    screen.getByLabelText(texte.seiten.pakete.bezugAdresse),
    'https://github.com/MagicSetEditorPacks/M15-Magic-Pack',
  )
  await userEvent.clear(screen.getByLabelText(texte.seiten.pakete.bezugStand))
  await userEvent.type(
    screen.getByLabelText(texte.seiten.pakete.bezugStand),
    '3.1.0',
  )
  await userEvent.clear(
    screen.getByLabelText(texte.seiten.pakete.bezugUnterverzeichnis),
  )
  await userEvent.type(
    screen.getByLabelText(texte.seiten.pakete.bezugUnterverzeichnis),
    'data',
  )
  await userEvent.type(
    screen.getByLabelText(texte.seiten.pakete.bezugName),
    'M15',
  )
}

describe('Vorlage beziehen', () => {
  // **Der Hinweis steht vor dem Bezug** (#139, §11): wer nicht weiß, woher eine
  // Vorlage kommt, kann ihre Rechtslage nicht einschätzen. Es wird nichts
  // geladen, bevor jemand zugestimmt hat.
  it('lädt erst, wenn jemand die Herkunft gesehen und zugestimmt hat', async () => {
    const PaketBeziehen = vi.fn(() =>
      Promise.resolve({
        ort: '/pakete/M15',
        stand: '3.1.0',
        pruefsumme: 'sha256:abc',
        uebergangen: [],
      }),
    )

    belegeSeite({
      PaketHerkunft: () =>
        Promise.resolve({
          quelle: 'MagicSetEditorPacks/M15-Magic-Pack',
          stand: '3.1.0',
        }),
      PaketBeziehen,
    })

    zeige()
    await fuelleAus()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )

    // Bis hierher ist nichts geladen worden.
    expect(PaketBeziehen).not.toHaveBeenCalled()
    expect(
      await screen.findByText(kurzform(texte.seiten.pakete.bezugLizenz)),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    )

    expect(PaketBeziehen).toHaveBeenCalledWith(
      'https://github.com/MagicSetEditorPacks/M15-Magic-Pack',
      '3.1.0',
      'data',
      'M15',
    )

    expect(await screen.findByText(/sha256:abc/)).toBeInTheDocument()
  })

  // Wer den Hinweis liest und ablehnt, lädt nichts.
  it('lädt nichts, wenn jemand ablehnt', async () => {
    const PaketBeziehen = vi.fn()

    belegeSeite({
      PaketHerkunft: () =>
        Promise.resolve({
          quelle: 'MagicSetEditorPacks/M15-Magic-Pack',
          stand: '3.1.0',
        }),
      PaketBeziehen,
    })

    zeige()
    await fuelleAus()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugAbbrechen }),
    )

    expect(PaketBeziehen).not.toHaveBeenCalled()
    expect(
      screen.queryByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    ).not.toBeInTheDocument()
  })

  // **Was nicht ausgepackt wurde, wird genannt** (#139).
  it('nennt die Dateien, die nicht mitgekommen sind', async () => {
    belegeSeite({
      PaketHerkunft: () =>
        Promise.resolve({
          quelle: 'MagicSetEditorPacks/M15-Magic-Pack',
          stand: '3.1.0',
        }),
      PaketBeziehen: () =>
        Promise.resolve({
          ort: '/pakete/M15',
          stand: '3.1.0',
          pruefsumme: 'sha256:abc',
          uebergangen: ['magicseteditor.exe', 'mse.exe'],
        }),
    })

    zeige()
    await fuelleAus()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    )

    expect(
      await screen.findByText(texte.seiten.pakete.bezugUebergangen),
    ).toBeInTheDocument()
    expect(screen.getByText('magicseteditor.exe')).toBeInTheDocument()
  })

  // Ein Zweig wird vom Backend abgewiesen; die Meldung erscheint am Feld.
  it('meldet, wenn der Stand keiner ist', async () => {
    belegeSeite({
      PaketHerkunft: () =>
        Promise.resolve({
          quelle: 'MagicSetEditorPacks/M15-Magic-Pack',
          stand: '3.1.0',
        }),
      PaketBeziehen: () =>
        Promise.resolve({
          ort: '',
          stand: '',
          pruefsumme: '',
          uebergangen: [],
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: '»main« ist keine Freigabe dieses Repositoriums.',
            feld: 'stand',
          },
        }),
    })

    zeige()
    await fuelleAus()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('keine Freigabe')
  })

  // Auch das Nachschlagen der Herkunft kann scheitern.
  it('meldet, wenn sich die Herkunft nicht holen lässt', async () => {
    belegeSeite({
      PaketHerkunft: () => Promise.reject(new Error('die Bindung fehlt')),
      PaketBeziehen: vi.fn(),
    })

    zeige()
    await fuelleAus()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Bindung fehlt',
    )
  })

  // Ein Bezug, der gar nicht antwortet, wird gemeldet.
  //
  // **Nicht dasselbe wie ein Fehler im Umschlag**: hier scheitert der Aufruf
  // selbst — die Bindung fehlt, oder das Fenster ist fort. Wer nur den einen
  // Weg prüft, hat den anderen nie ausgeführt.
  it('meldet einen Aufruf, der scheitert', async () => {
    belegeSeite({
      PaketHerkunft: () =>
        Promise.resolve({
          quelle: 'MagicSetEditorPacks/M15-Magic-Pack',
          stand: '3.1.0',
        }),
      PaketBeziehen: () => Promise.reject(new Error('die Bindung fehlt')),
    })

    zeige()
    await fuelleAus()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Bindung fehlt',
    )
  })

  // **Bezogen wird, was bestätigt wurde — nicht, was gerade dasteht.**
  //
  // Der Zweitblick über Phase 9 hat es gefunden: nach dem Hinweis blieben die
  // Felder änderbar, und »Beziehen« nahm die **aktuellen** Werte. Wer die
  // Herkunft eines Repositoriums liest und danach eine andere Adresse eintippt,
  // bezöge etwas, dem er nie zugestimmt hat (§7, Regeln 1 und 2).
})
