// Die Paketseite: was geschieht, wenn zwei Läufe einander überholen (§7).
//
// **Getrennt von `pakete.test.tsx`**, weil eine Datei über 400 Zeilen ginge
// (§2) — und weil es eine andere Frage ist: dort die Zustimmung, hier die
// Reihenfolge.
import { act, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { belegeSeite, oeffneBezug, texte, zeige } from '../test/paketseite'

afterEach(() => {
  raeumeApi()
})

// bezugsergebnis gibt einen geglückten Bezug.
function bezugsergebnis() {
  return Promise.resolve({
    ort: '/pakete/M15',
    stand: '3.1.0',
    pruefsumme: 'sha256:abc',
    uebergangen: [],
  })
}

// fuelleAus tippt die vier Felder — **nacheinander**, wie jemand es täte.
//
// **Erst die Schublade** — seit UXR-17 (#1925) steht der Bezug hinter
// *Paket beziehen …* im Kopf der Seite.
async function fuelleAus() {
  await oeffneBezug()

  // **Erst leeren, dann tippen.** Seit dem Vorlagenpack als Vorgabe stehen die
  // drei Felder schon gefüllt da; `type` hängte sonst hinten an.
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

describe('Vorlage beziehen — überholte Läufe', () => {
  it('bezieht, was bestätigt wurde — nicht, was danach eingetippt wird', async () => {
    const PaketBeziehen = vi.fn(() => bezugsergebnis())

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

    // Der Hinweis steht. Jetzt ändert jemand die Adresse.
    const adresse = screen.getByLabelText(texte.seiten.pakete.bezugAdresse)
    await userEvent.clear(adresse)
    await userEvent.type(adresse, 'https://beispiel.test/etwas-anderes')

    // **Die Zustimmung ist damit hinfällig** — es gibt nichts mehr zu beziehen.
    expect(
      screen.queryByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    ).not.toBeInTheDocument()

    expect(PaketBeziehen).not.toHaveBeenCalled()
  })

  // **Eine verspätete Antwort trägt sich nicht in einen geräumten Stand ein.**
  //
  // §7 Regel 2. Der Nachlauf zum Zweitblick über Phase 9 hat es gefunden: der
  // erste Test änderte die Eingabe erst **nach** der Antwort. Kommt sie später,
  // räumt die Eingabeänderung erst den Kontext — und die alte Antwort trägt ihn
  // wieder ein, mit einer Herkunft, die zu den neuen Feldern nicht passt.
  it('trägt eine verspätete Herkunft nicht mehr ein', async () => {
    let antworte: (wert: { quelle: string; stand: string }) => void = () =>
      undefined

    belegeSeite({
      PaketHerkunft: () =>
        new Promise<{ quelle: string; stand: string }>((aufloesen) => {
          antworte = aufloesen
        }),
      PaketBeziehen: vi.fn(),
    })

    zeige()
    await fuelleAus()

    // **Fragen, Feld ändern und antworten in einem Zug**, damit React alles
    // zusammen verarbeitet: eine Auflösung außerhalb von `act` erzeugt eine
    // Warnung und einen Zustand, den niemand gezeichnet hat.
    await act(async () => {
      await userEvent.click(
        screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
      )

      const adresse = screen.getByLabelText(texte.seiten.pakete.bezugAdresse)
      await userEvent.clear(adresse)

      antworte({ quelle: 'MagicSetEditorPacks/M15-Magic-Pack', stand: '3.1.0' })
    })

    expect(
      screen.queryByRole('button', { name: texte.seiten.pakete.bezugStarten }),
    ).not.toBeInTheDocument()
  })

  // **Auch ein verspäteter Fehler gehört zu seinem Stand** (§7, Regel 2).
  //
  // Die unauffälligere Hälfte: eine Meldung zu einer Adresse, die niemand mehr
  // im Feld stehen hat, sieht aus wie ein Fehler der neuen.
  it('zeigt einen verspäteten Fehler nicht mehr an', async () => {
    let scheitere: (grund: Error) => void = () => undefined

    belegeSeite({
      PaketHerkunft: () =>
        new Promise<{ quelle: string; stand: string }>((_, ablehnen) => {
          scheitere = ablehnen
        }),
      PaketBeziehen: vi.fn(),
    })

    zeige()
    await fuelleAus()

    await act(async () => {
      await userEvent.click(
        screen.getByRole('button', { name: texte.seiten.pakete.bezugPruefen }),
      )

      await userEvent.clear(
        screen.getByLabelText(texte.seiten.pakete.bezugAdresse),
      )

      scheitere(new Error('die Bindung fehlt'))
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  // Ohne Adresse und Stand gibt es nichts zu prüfen.
  //
  // **Die Schaltfläche bleibt trotzdem bedienbar.** Eine ausgegraute ist per
  // Tastatur nicht erreichbar, und wer sie ansteuert, erfährt nicht, warum
  // nichts geschieht — sie sagt es beim Klick.
  it('sagt, was fehlt, statt sich abzuschalten', async () => {
    const PaketHerkunft = vi.fn(() =>
      Promise.resolve({ quelle: 'egal', stand: 'egal' }),
    )

    belegeSeite({ PaketHerkunft, PaketBeziehen: vi.fn() })

    zeige()
    await oeffneBezug()

    // **Der Stand wird geleert.** Seit dem Vorlagenpack als Vorgabe stehen
    // Adresse, Stand und Unterverzeichnis schon da; unvollständig wird das
    // Formular erst, wenn jemand ein Feld leert.
    await userEvent.clear(screen.getByLabelText(texte.seiten.pakete.bezugStand))

    const knopf = screen.getByRole('button', {
      name: texte.seiten.pakete.bezugPruefen,
    })
    expect(knopf).toBeEnabled()

    await userEvent.click(knopf)

    expect(PaketHerkunft).not.toHaveBeenCalled()
    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.seiten.pakete.bezugUnvollstaendig,
    )
  })
})
