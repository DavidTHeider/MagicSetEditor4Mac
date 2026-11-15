// **Die Konsole sieht die Karte, die die Liste aufgeschlagen hat.**
//
// # Der Befund
//
// Wer in der Kartenliste eine Karte anklickt und danach die **Konsole**
// aufschlägt, bekam „Dafür muss eine Karte aufgeschlagen sein." — während
// dieselbe Karte daneben im Editor stand. Über die **Suche** geöffnet ging es;
// über die Liste nicht.
//
// Der Grund: `Karten.tsx` führte seine Auswahl **nur örtlich**. Die Konsole
// bekommt ihre Kartenkennung aus dem Zustand von `App`, und dorthin sagte
// niemand Bescheid. Ein Rückruf, der fehlt, sieht im Einzeltest der
// Kartenseite richtig aus: sie wählt ja. Erst hier zeigt sich, ob jemand
// zuhört.
//
// **Gefunden bei der Sichtprüfung im laufenden Fenster** beim Bauen der
// Bildschirmabzüge — nicht von einem Test
// (#892).
// Dieselbe Gestalt wie
// #1013:
// eine Fähigkeit, die auf dem naheliegenden Weg nicht zu erreichen ist.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { ausKarten } from './test/kartenproben'
import { leisteneintrag } from './test/leistenweg'

const texte = kataloge.de

beforeEach(() => {
  belegeApi({ ...belegeFenster('de') })
})

afterEach(() => {
  raeumeApi()
})

describe('Die Skriptkonsole', () => {
  it('sieht die Karte, die in der Liste gewählt wurde', async () => {
    const Karten = vi.fn(() =>
      Promise.resolve({
        // **Der Name steht in den Werten, nicht daneben.** Die Zeilenliste
        // liest ihn aus dem Feld `name`; ein Verbund mit `name:` und leeren
        // Werten gäbe „ohne Namen".
        karten: [
          { id: '11', nummer: 1, werte: [{ feld: 'name', wert: 'Erste' }] },
          { id: '22', nummer: 2, werte: [{ feld: 'name', wert: 'Zweite' }] },
        ],
      }),
    )

    const wege = ausKarten(Karten, '11')
    const Kartenzeilen = vi.fn(wege.Kartenzeilen)

    belegeApi({
      ...belegeFenster('de'),
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Probeset',
              spiel: 'magic',
              stil: 'magic-m15',
              karten: 2,
              herkunft: '/a.mse-set',
              herkunftFehlt: false,
            },
          ],
        }),
      ...wege,
      Kartenzeilen,
      Karten,
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    render(<App />)

    await userEvent.click(await leisteneintrag('sets'))
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Probeset — ${texte.seiten.sets.oeffnen}`,
      }),
    )

    await waitFor(() => {
      expect(Kartenzeilen).toHaveBeenCalled()
    })

    // **Die zweite Karte wählen** — und nicht die erste. Die erste ist beim
    // Öffnen ohnehin aufgeschlagen (#668); ein Test darauf sagte nur, dass
    // die Vorgabe ankommt, nicht dass ein Klick etwas bewirkt.
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Zweite — ${texte.seiten.karten.bearbeiten}`,
      }),
    )

    // **Die Konsole steht in der Schublade „Werkzeuge"** (UX-08, #1680).
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.allgemein.arbeitsplatz.schubladen.werkzeuge.knopf,
      }),
    )
    await userEvent.click(
      await screen.findByRole('tab', { name: texte.seiten.konsole.titel }),
    )

    // **Das Eingabefeld ist der Beleg.** Ohne aufgeschlagene Karte zeigt die
    // Konsole einen Satz und kein Feld.
    expect(
      await screen.findByLabelText(texte.seiten.karten.konsole.ausdruck),
    ).toBeInTheDocument()

    expect(
      screen.queryByText(texte.seiten.karten.konsole.keineKarte),
    ).not.toBeInTheDocument()
  })

  // **Und die Karte, die das Öffnen aufgeschlagen hat** — ohne einen Klick.
  //
  // Die Kartenseite schlägt die erste Karte beim Öffnen auf (#668), aber nur
  // **örtlich**: das Fenster erfuhr davon nichts, und die Konsole sagte
  // „Dafür muss eine Karte aufgeschlagen sein", während die Karte daneben im
  // Editor stand. Dieselbe Gestalt wie #892, nur ohne Klick.
  //
  // **Gefunden in der Sichtprüfung der Schubladen** (UX-08, #1680): seitdem
  // liegt die Konsole **über** dieser Karte, und der Widerspruch ist nicht
  // mehr zu übersehen.
  it('sieht die Karte, die beim Öffnen aufgeschlagen wurde', async () => {
    const wege = ausKarten(
      () =>
        Promise.resolve({
          karten: [
            { id: '11', nummer: 1, werte: [{ feld: 'name', wert: 'Erste' }] },
            { id: '22', nummer: 2, werte: [{ feld: 'name', wert: 'Zweite' }] },
          ],
        }),
      '11',
    )

    belegeApi({
      ...belegeFenster('de'),
      ...wege,
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Probeset',
              spiel: 'magic',
              stil: 'magic-m15',
              karten: 2,
              herkunft: '/a.mse-set',
              herkunftFehlt: false,
            },
          ],
        }),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    })

    render(<App />)

    await userEvent.click(await leisteneintrag('sets'))
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Probeset — ${texte.seiten.sets.oeffnen}`,
      }),
    )

    // **Die Liste steht** — und mit ihr die erste Karte im Editor.
    await screen.findByRole('button', {
      name: `Zweite — ${texte.seiten.karten.bearbeiten}`,
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.allgemein.arbeitsplatz.schubladen.werkzeuge.knopf,
      }),
    )
    await userEvent.click(
      await screen.findByRole('tab', { name: texte.seiten.konsole.titel }),
    )

    expect(
      await screen.findByLabelText(texte.seiten.karten.konsole.ausdruck),
    ).toBeInTheDocument()
  })
})
