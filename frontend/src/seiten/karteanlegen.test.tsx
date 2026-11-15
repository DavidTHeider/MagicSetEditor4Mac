// **Eine Karte lässt sich anlegen**
// (#562).
//
// Bis hierher konnte die Anwendung Sets **öffnen** und Felder ändern; eine neue
// Karte gab es nirgends. Wer ein Set anlegen wollte, musste es woanders
// schreiben.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function karte(id: string, name: string): KarteDTO {
  return { id, nummer: Number(id) - 1, werte: [{ feld: 'name', wert: name }] }
}

function zeige(teile: Partial<Api>, bestandGeaendert = () => undefined) {
  belegeApi({
    // **Der Verlauf antwortet leer** — der Anhang zeigt ihn seit U6 immer,
    // und unbelegt stünde hier eine zweite Meldung neben der geprüften.
    Verlauf: () => Promise.resolve({ schritte: [], fehler: null }),
    ...ausKarten(() => Promise.resolve({ karten: [karte('1', 'Erste')] })),
    Karten: () => Promise.resolve({ karten: [karte('1', 'Erste')] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    // **Eine Vorschau mit Bild und nicht `null`.** Seit
    // #668
    // steht eine Karte offen, sobald die Liste da ist — und eine Antwort
    // ohne Bild **und** ohne Fehler ist ein Befund, den die Seite meldet.
    // Eine Vorrichtung, die ärmer ist als der Betrieb, prüft den falschen
    // Weg (§3).
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    // **Ein eigener `Karten`-Beleg wird mitprojiziert** (#891). Die Seite holt
    // `Kartenzeilen` und `Karte`; ohne diese Zeile bliebe die Vorgabe darüber
    // stehen, und der Test prüfte Karten, die er nicht beschrieben hat.
    ...(teile.Karten ? ausKarten(teile.Karten, '7') : {}),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="7"
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={bestandGeaendert}
      />
    </Sprachraum>,
  )
}

// vorschau gibt eine Antwort, wie sie der Rasterer gibt.
function vorschau() {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [],
      nennungen: [],
      vorgaben: [],
    },
  }
}

describe('Karte anlegen', () => {
  it('legt an und lädt die Liste nach', async () => {
    const KarteAnlegen = vi.fn(() => Promise.resolve({ karteID: '2' }))
    const Karten = vi
      .fn()
      .mockResolvedValueOnce({ karten: [karte('1', 'Erste')] })
      .mockResolvedValue({
        karten: [karte('1', 'Erste'), karte('2', '')],
      })

    zeige({ KarteAnlegen, Karten })

    await screen.findByRole('button', {
      name: `Erste — ${texte.seiten.karten.bearbeiten}`,
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    expect(KarteAnlegen).toHaveBeenCalledWith('7')

    // **Die neue Karte steht in der Liste** — nachgeladen, nicht dazugedacht.
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: `${texte.seiten.karten.ohneName} — ${texte.seiten.karten.bearbeiten}`,
        }),
      ).toBeInTheDocument()
    })
  })

  // **Und sie wird aufgeschlagen.** Wer eine Karte anlegt, will sie füllen —
  // nicht erst in einer Liste von dreitausend suchen.
  it('schlägt die neue Karte auf', async () => {
    const Karten = vi
      .fn()
      .mockResolvedValueOnce({ karten: [karte('1', 'Erste')] })
      .mockResolvedValue({ karten: [karte('1', 'Erste'), karte('2', '')] })

    zeige({ KarteAnlegen: () => Promise.resolve({ karteID: '2' }), Karten })

    await screen.findByRole('button', {
      name: `Erste — ${texte.seiten.karten.bearbeiten}`,
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: `${texte.seiten.karten.ohneName} — ${texte.seiten.karten.bearbeiten}`,
        }),
      ).toHaveAttribute('aria-current', 'true')
    })
  })

  // **Der Kopf des Arbeitsplatzes erfährt davon.**
  //
  // Er nennt die Kartenzahl, und die kam aus dem Setbestand des Fensters. In
  // der gebauten Anwendung stand nach dem Anlegen „1 Karten" über einer Liste
  // mit zweien — eine Anzeige, die etwas behauptet, was daneben widerlegt wird.
  it('meldet die neue Karte dem Fenster', async () => {
    const gemeldet = vi.fn()
    const Karten = vi
      .fn()
      .mockResolvedValueOnce({ karten: [karte('1', 'Erste')] })
      .mockResolvedValue({ karten: [karte('1', 'Erste'), karte('2', '')] })

    zeige(
      { KarteAnlegen: () => Promise.resolve({ karteID: '2' }), Karten },
      gemeldet,
    )

    await screen.findByRole('button', {
      name: `Erste — ${texte.seiten.karten.bearbeiten}`,
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    await waitFor(() => {
      expect(gemeldet).toHaveBeenCalled()
    })
  })

  // **Und nur dann.** Ein gescheitertes Anlegen hat den Bestand nicht
  // geändert; wer trotzdem meldete, ließe das Fenster ohne Anlass nachladen.
  it('meldet nichts, wenn das Anlegen scheitert', async () => {
    const gemeldet = vi.fn()

    zeige(
      { KarteAnlegen: () => Promise.reject(new Error('Die Platte ist voll.')) },
      gemeldet,
    )

    await screen.findByRole('button', {
      name: `Erste — ${texte.seiten.karten.bearbeiten}`,
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    expect(gemeldet).not.toHaveBeenCalled()
  })

  // **Ein gescheitertes Anlegen wird gemeldet**, nicht verschluckt (§8).
  it('meldet einen Fehlschlag', async () => {
    zeige({
      KarteAnlegen: () => Promise.reject(new Error('Die Platte ist voll.')),
    })

    await screen.findByRole('button', {
      name: `Erste — ${texte.seiten.karten.bearbeiten}`,
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.anlegen }),
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Die Platte ist voll.',
      )
    })
  })
})
