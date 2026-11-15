// Die Prüfungen des Karteneditors — **je ein Test für jede der vier Regeln aus
// AGENTS.md §7**, wie #40 es verlangt.
//
// Jeder Fall ist der, der ohne die Regel schiefgeht. Ein Test, der nur den
// geraden Weg geht, hält die Regel für überflüssig.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO, SchreibAntwort, VorschauAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probekarte(teile: Partial<KarteDTO> = {}): KarteDTO {
  return {
    id: '5',
    nummer: 0,
    werte: [
      { feld: 'name', wert: 'Feuerball' },
      { feld: 'typ', wert: 'Zauber' },
    ],
    ...teile,
  }
}

function vorschau(
  bild = '/vorschau/AAAA.png',
  befunde: string[] = [],
): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild,
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde,
      flaechen: [],
      nennungen: [],
      vorgaben: [],
    },
  }
}

function zeige(karte = probekarte()) {
  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte}
        formen={[]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Karteneditor', () => {
  it('zeigt je Feldwert ein Eingabefeld und die Vorschau', async () => {
    belegeApi({
      // **Die Feldformen** — seit #560 lädt der Editor sie (leer genügt hier).
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorschau: () => Promise.resolve(vorschau()),
    })
    zeige()

    expect(screen.getByRole('textbox', { name: 'name' })).toHaveValue(
      'Feuerball',
    )
    expect(screen.getByRole('textbox', { name: 'typ' })).toHaveValue('Zauber')

    await waitFor(() => {
      expect(
        screen.getByRole('img', { name: texte.seiten.karten.vorschau }),
      ).toBeInTheDocument()
    })
  })

  // **Regel 3, erste Hälfte: nach einem geglückten Schreiben wird nachgeladen.**
  //
  // Ohne das zeigte die Vorschau den Stand von vorher — und §6 sagt, eine
  // Änderung an einem Feld sei in der Vorschau zu sehen, ohne dass jemand etwas
  // bestätigt.
  it('lädt die Vorschau nach einem geglückten Schreiben nach', async () => {
    const bilder = vi
      .fn<() => Promise<VorschauAntwort>>()
      .mockResolvedValueOnce(vorschau('/vorschau/ERST.png'))
      .mockResolvedValue(vorschau('/vorschau/DANN.png'))

    belegeApi({
      Vorschau: bilder,
      FeldSetzen: () => Promise.resolve({} satisfies SchreibAntwort),
    })

    zeige()

    await waitFor(() => {
      expect(screen.getByRole('img')).toHaveAttribute(
        'src',
        '/vorschau/ERST.png',
      )
    })

    await userEvent.type(screen.getByRole('textbox', { name: 'name' }), 'X')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('img')).toHaveAttribute(
        'src',
        '/vorschau/DANN.png',
      )
    })
  })

  // **Regel 3, zweite Hälfte: nach einem gescheiterten Schreiben wird nicht
  // nachgeladen.**
  //
  // Sonst überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
  it('lädt nach einem gescheiterten Schreiben nicht nach', async () => {
    const bilder = vi
      .fn<() => Promise<VorschauAntwort>>()
      .mockResolvedValue(vorschau())

    belegeApi({
      Vorschau: bilder,
      // **Beim Tippen fragt die Vorschau ein Probebild an** — seit E01 (#2238)
      // steht ihr Scheitern als Meldung an der Vorschau, und dann gäbe es zwei.
      Probevorschau: () => Promise.resolve(vorschau()),
      FeldSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'Dieses Feld hat sich inzwischen geändert.',
          },
        }),
    })

    zeige()

    await waitFor(() => {
      expect(bilder).toHaveBeenCalledTimes(1)
    })

    // **Auf das Feld gewartet, nicht auf den Ruf** (#1809): es steht erst,
    // wenn die Felder der Karte gezeichnet sind.
    await userEvent.type(
      await screen.findByRole('textbox', { name: 'name' }),
      'X',
    )
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        // **Der Satz aus dem Katalog, nicht der aus Go** (#133). Übersetzt
        // wird über den Schlüssel; die genauere Formulierung aus Go bleibt der
        // Rückfall für einen Schlüssel, den der Katalog nicht kennt.
        kataloge.de.fehler.konflikt,
      )
    })

    // **Und die Vorschau wurde nicht noch einmal geholt.**
    expect(bilder).toHaveBeenCalledTimes(1)
  })

  // **Der Fehler erscheint am verursachenden Feld** (§6, zweite Zusage aus #40).
  it('setzt die Meldung an das Feld, das sie verursacht hat', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      // **Beim Tippen fragt die Vorschau ein Probebild an** — seit E01 (#2238)
      // steht ihr Scheitern als Meldung an der Vorschau, und dann gäbe es zwei.
      Probevorschau: () => Promise.resolve(vorschau()),
      FeldSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung',
            schluessel: 'pflichtfeld',
            meldung: 'Ein Feld braucht einen Namen.',
            feld: 'typ',
          },
        }),
    })

    zeige()

    await userEvent.type(screen.getByRole('textbox', { name: 'name' }), 'X')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })

    // Die Bindung nennt `typ` — also steht die Meldung dort, nicht an `name`.
    const typfeld = screen.getByRole('textbox', { name: 'typ' })
    expect(typfeld).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('textbox', { name: 'name' })).toHaveAttribute(
      'aria-invalid',
      'false',
    )
  })

  // **Regel 4: ein Vorgang je Datensatz.**
  //
  // Ein zweiter Aufruf, während der erste läuft, ginge gegen einen Stand, den
  // das Backend noch nicht kennt. Das Feld ist deshalb gesperrt, solange
  // geschrieben wird.
  it('sperrt das Feld, solange geschrieben wird', async () => {
    let freigeben: (() => void) | undefined

    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      // **Beim Tippen fragt die Vorschau ein Probebild an** — seit E01 (#2238)
      // steht ihr Scheitern als Meldung an der Vorschau, und dann gäbe es zwei.
      Probevorschau: () => Promise.resolve(vorschau()),
      FeldSetzen: () =>
        new Promise<SchreibAntwort>((aufloesen) => {
          freigeben = () => {
            aufloesen({})
          }
        }),
    })

    zeige()

    await userEvent.type(screen.getByRole('textbox', { name: 'name' }), 'X')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).toBeDisabled()
    })

    // Das andere Feld bleibt bedienbar — zwei Felder derselben Karte dürfen
    // gleichzeitig gehen.
    expect(screen.getByRole('textbox', { name: 'typ' })).not.toBeDisabled()

    freigeben?.()

    await waitFor(() => {
      expect(screen.getByRole('textbox', { name: 'name' })).not.toBeDisabled()
    })
  })

  it('meldet einen Fehler der Vorschau, statt leer zu bleiben', async () => {
    belegeApi({
      Vorschau: () =>
        Promise.resolve({
          vorschau: null,
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            // Der Satz aus Go reist mit, sichtbar wird aber der Katalogsatz
            // zum Schlüssel `unbekannt` (#133).
            meldung: 'Die Karte ließ sich nicht zeichnen.',
          },
        }),
    })

    zeige()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        kataloge.de.fehler.unbekannt,
      )
    })
  })

  // Eine Antwort ohne Bild **und** ohne Fehler ist ein Widerspruch — er wird
  // gemeldet, statt eine leere Fläche zu zeigen.
  it('meldet eine Antwort ohne Bild und ohne Fehler', async () => {
    belegeApi({
      // **Die Feldformen** — seit #560 lädt der Editor sie (leer genügt hier).
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      Vorschau: () => Promise.resolve({ vorschau: null }),
    })

    zeige()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'ohne Bild und ohne Fehler',
      )
    })
  })

  it('zeigt die Befunde der Vorschau', async () => {
    belegeApi({
      Vorschau: () =>
        Promise.resolve(
          vorschau('/vorschau/AAAA.png', [
            'Ebene 3: die Schrift MPlantin ist nicht zu finden',
          ]),
        ),
    })

    zeige()

    await waitFor(() => {
      expect(
        screen.getByText('Ebene 3: die Schrift MPlantin ist nicht zu finden'),
      ).toBeInTheDocument()
    })
  })
  // **Die Eingabetaste lädt die Seite nicht neu.**
  //
  // Ein Formular wird von der Eingabetaste abgeschickt — und in einer
  // Wails-Anwendung heißt das: das Fenster lädt neu, und der offene Zustand ist
  // weg. Der Abfangkanal dafür ist eine Zeile und ohne diesen Test ungeprüft.
  //
  // Ausgelöst wird das Ereignis unmittelbar: jsdom schickt ein Formular ohne
  // Absendeknopf nicht von sich aus ab, und ein Test, der auf diese Eigenheit
  // baut, prüfte jsdom statt den Abfangkanal.
  it('fängt das Absenden des Formulars ab', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      FeldSetzen: () => Promise.resolve({} satisfies SchreibAntwort),
    })

    zeige()

    await waitFor(() => {
      expect(screen.getByRole('img')).toBeInTheDocument()
    })

    const formular = screen
      .getByRole('textbox', { name: 'name' })
      .closest('form')
    if (!formular) {
      throw new Error('kein Formular gefunden')
    }

    const ereignis = new Event('submit', { bubbles: true, cancelable: true })
    fireEvent(formular, ereignis)

    expect(ereignis.defaultPrevented).toBe(true)
  })

  // Ein Fehler, der **kein** Bindungsfehler ist, landet am bearbeiteten Feld.
  //
  // Er kommt vor: die Fassade lehnt mit einer gewöhnlichen Meldung ab, wenn eine
  // Methode nicht belegt ist — und im Betrieb bricht auch mal die Bindung selbst
  // weg. Ohne Feldbezug muss die Meldung trotzdem irgendwo stehen, und das eben
  // bearbeitete Feld ist die einzige Stelle, die etwas damit zu tun hat.
  it('setzt einen Fehler ohne Feldbezug an das bearbeitete Feld', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      // **Beim Tippen fragt die Vorschau ein Probebild an** — seit E01 (#2238)
      // steht ihr Scheitern als Meldung an der Vorschau, und dann gäbe es zwei.
      Probevorschau: () => Promise.resolve(vorschau()),
      FeldSetzen: () => Promise.reject(new Error('die Bindung ist weg')),
    })

    zeige()

    await userEvent.type(screen.getByRole('textbox', { name: 'name' }), 'X')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('die Bindung ist weg')
    })

    expect(screen.getByRole('textbox', { name: 'name' })).toHaveAttribute(
      'aria-invalid',
      'true',
    )
  })

  // Und eine Ablehnung, die gar kein Fehlerobjekt ist, wird lesbar gemacht.
  it('macht auch aus einer Nicht-Fehler-Ablehnung eine Meldung', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(vorschau()),
      // **Beim Tippen fragt die Vorschau ein Probebild an** — seit E01 (#2238)
      // steht ihr Scheitern als Meldung an der Vorschau, und dann gäbe es zwei.
      Probevorschau: () => Promise.resolve(vorschau()),
      // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
      FeldSetzen: () => Promise.reject('etwas ganz anderes'),
    })

    zeige()

    await userEvent.type(screen.getByRole('textbox', { name: 'typ' }), 'X')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('etwas ganz anderes')
    })
  })
})
