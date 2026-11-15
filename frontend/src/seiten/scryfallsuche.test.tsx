// **Die Kartensuche bei Scryfall** — Werkstück 1 von Phase 3.6
// (#1167).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SuchtrefferDTO } from '../apiscryfall'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { oeffneAnlegen } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { Scryfallsuche } from './Scryfallsuche'

const texte = kataloge.de.seiten.karten.scryfallsuche

function treffer(teile: Partial<SuchtrefferDTO> = {}): SuchtrefferDTO {
  return {
    kennung: 'abc',
    name: 'Lightning Bolt',
    setcode: 'lea',
    setname: 'Limited Edition Alpha',
    nummer: '161',
    seltenheit: 'common',
    manakosten: '{R}',
    typzeile: 'Instant',
    bild: 'https://bilder/klein.jpg',
    ...teile,
  }
}

function zeige(waehlen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Scryfallsuche waehlen={waehlen} />
    </Sprachraum>,
  )

  return waehlen
}

afterEach(() => {
  raeumeApi()
})

describe('die Kartensuche', () => {
  // **Gesucht wird auf Anweisung und nicht beim Tippen.**
  //
  // Scryfall nennt eine Grenze, und §16 verlangt, sie **unter** der genannten
  // Zahl einzuhalten; eine Suche je Tastendruck ginge dagegen.
  it('sucht erst auf Knopfdruck', async () => {
    const gesucht = vi.fn().mockResolvedValue({
      treffer: [treffer()],
      weitere: false,
      alterStunden: 0,
      befunde: [],
      fehler: null,
    })

    belegeApi({ ScryfallSuche: gesucht })

    zeige()

    await userEvent.type(screen.getByLabelText(texte.frage), 'bolt')

    expect(gesucht).not.toHaveBeenCalled()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    await waitFor(() => {
      expect(gesucht).toHaveBeenCalledWith('bolt')
    })
  })

  // **Und die Eingabetaste sucht auch** (§6): wer tippt, will nicht zur Maus.
  it('sucht bei der Eingabetaste', async () => {
    const gesucht = vi.fn().mockResolvedValue({
      treffer: [],
      weitere: false,
      alterStunden: 0,
      befunde: [],
      fehler: null,
    })

    belegeApi({ ScryfallSuche: gesucht })

    zeige()

    await userEvent.type(screen.getByLabelText(texte.frage), 'bolt{Enter}')

    await waitFor(() => {
      expect(gesucht).toHaveBeenCalledWith('bolt')
    })
  })

  // **Ein Treffer zeigt, woran man ihn erkennt.**
  //
  // Name, Set und Nummer — und das Bild, denn eine Karte erkennt man am Bild
  // und nicht am Setcode.
  it('zeigt Name, Set und Bild', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    const zeile = await screen.findByRole('button', { name: /Lightning Bolt/ })

    expect(zeile).toHaveTextContent('Limited Edition Alpha')
    expect(zeile).toHaveTextContent('LEA 161')

    expect(screen.getByRole('presentation', { hidden: true })).toHaveAttribute(
      'src',
      'https://bilder/klein.jpg',
    )
  })

  // **Ein Klick gibt den Treffer nach oben.**
  it('reicht den gewählten Treffer weiter', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    const gewaehlt = zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))
    await userEvent.click(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    )

    expect(gewaehlt).toHaveBeenCalledWith(treffer())
  })

  // **Nichts gefunden ist eine Antwort und kein leerer Bereich** (§6).
  it('sagt, wenn nichts gefunden wurde', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByText(texte.nichtsGefunden)).toBeInTheDocument()
  })

  // **»Es gibt mehr« steht dabei.**
  //
  // Eine Liste, die schweigend endet, sieht aus wie eine vollständige — und
  // nachgeladen wird nicht.
  it('sagt, wenn Scryfall mehr hat', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: true,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByText(texte.mehr)).toBeInTheDocument()
  })

  // **Ein Fehler erscheint und verdeckt nichts.**
  it('meldet einen Fehler', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: { code: 'unbekannt', schluessel: '', meldung: 'kein Netz' },
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByRole('alert')).toHaveTextContent('kein Netz')
  })

  // **Und eine Karte ohne Bild bekommt keines.**
  it('kommt ohne Bild aus', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer({ bild: '' })],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    await screen.findByRole('button', { name: /Lightning Bolt/ })

    expect(
      screen.queryByRole('presentation', { hidden: true }),
    ).not.toBeInTheDocument()
  })
})

// **Und der gefundene Name landet im Anlegekasten** — Werkstück 1 von
// Phase 3.6.
//
// # Was hier steht und nicht oben
//
// Oben steht, was die **Suche** tut; hier, dass sie an der Kartenlasche hängt
// und dass ihr Ergebnis irgendwohin führt. Eine Suche, deren Treffer man
// anklicken kann und bei der dann nichts geschieht, wäre in beiden Prüfungen
// grün.
//
// **Was sie noch nicht tut**, steht am Code: den genauen **Druck** wählen.
// `ScryfallKarteAnlegen` nimmt einen Namen entgegen, und daraus einen
// Druckbezug zu machen ändert den Vertrag an allen vier Flächen.
describe('die Suche an der Kartenlasche', () => {
  it('belegt den Anlegekasten mit dem gefundenen Namen', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [] }),
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    render(
      <Sprachraum sprache="de">
        {/* **Über die ganze Lasche und nicht über die Kopfzeile allein**
            (#1240): der gewählte Treffer wohnt seit der Umgruppierung in
            `useScryfallvorschlag`, und die Zusage dieses Laufs ist genau die
            Naht zwischen Suche und Anlegekasten. Eine Vorrichtung, die den
            Zustand selbst hielte, wäre großzügiger als der Betrieb (§3). */}
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    // **Der Scryfall-Weg steht hinter seinem Aufklapp** (#1240).
    await oeffneAnlegen()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))
    await userEvent.click(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    )

    await waitFor(() => {
      expect(
        screen.getByLabelText(kataloge.de.seiten.karten.scryfallkarte.name),
      ).toHaveValue('Lightning Bolt')
    })
  })

  // **Und eine gescheiterte Suche räumt ihn wieder** — Befund 1 des Nachlaufs
  // über Phase 3.6.
  //
  // Die Wahl stammt aus einer Trefferliste, die es nicht mehr gibt. Sie stehen
  // zu lassen hieße, mit dem Ergebnis der vorigen Suche weiterzuarbeiten.
  it('räumt den Anlegekasten, wenn die nächste Suche scheitert', async () => {
    const gesucht = vi
      .fn()
      .mockResolvedValueOnce({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      })
      .mockResolvedValue({
        treffer: [],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'die Gegenstelle antwortet nicht',
        },
      })

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [] }),
      ScryfallSuche: gesucht,
    })

    render(
      <Sprachraum sprache="de">
        {/* **Über die ganze Lasche und nicht über die Kopfzeile allein**
            (#1240): der gewählte Treffer wohnt seit der Umgruppierung in
            `useScryfallvorschlag`, und die Zusage dieses Laufs ist genau die
            Naht zwischen Suche und Anlegekasten. Eine Vorrichtung, die den
            Zustand selbst hielte, wäre großzügiger als der Betrieb (§3). */}
        <Karten
          karteGewaehlt={vi.fn()}
          setID="1"
          karteID={undefined}
          verlaufsstand={0}
          verlaufsfehler={undefined}
          bestandGeaendert={vi.fn()}
        />
      </Sprachraum>,
    )

    // **Der Scryfall-Weg steht hinter seinem Aufklapp** (#1240).
    await oeffneAnlegen()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))
    await userEvent.click(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    )

    await waitFor(() => {
      expect(
        screen.getByLabelText(kataloge.de.seiten.karten.scryfallkarte.name),
      ).toHaveValue('Lightning Bolt')
    })

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    await waitFor(() => {
      expect(
        screen.getByLabelText(kataloge.de.seiten.karten.scryfallkarte.name),
      ).toHaveValue('')
    })
  })
})
