// **Die Felder des Stils** — die Zusatzfelder einer Karte im Formular
// (#2298).
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO } from '../api'
import type { ZusatzfeldDTO } from '../stilwahl'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stilfelder } from './Stilfelder'
import type { Feldfehler } from './feldkennung'

const texte = kataloge.de.seiten.karten

function form(teile: Partial<FeldbeschreibungDTO> = {}): FeldbeschreibungDTO {
  return {
    name: 'vermerk',
    aufschrift: 'Vermerk',
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    ...teile,
  }
}

function zusatzfeld(teile: Partial<ZusatzfeldDTO> = {}): ZusatzfeldDTO {
  return { form: form(), wert: '', gespeichert: false, ...teile }
}

function zeige({
  fehler,
  melden = vi.fn(),
  geglueckt = vi.fn(),
}: {
  fehler?: Feldfehler
  melden?: (befund: Feldfehler | undefined) => void
  geglueckt?: () => void
} = {}) {
  render(
    <Sprachraum sprache="de">
      <Stilfelder
        wege={{
          setID: '7',
          karteID: '3',
          vorgang: { starte: (_, was) => was(), laeuft: () => false },
          texte: kataloge.de,
          melden,
          geglueckt,
        }}
        stand={0}
        fehler={fehler}
      />
    </Sprachraum>,
  )

  return { melden, geglueckt }
}

afterEach(() => {
  raeumeApi()
})

const lesefehler = {
  code: 'unbekannt' as const,
  schluessel: 'unbekannt',
  meldung: 'Der Stil ist nicht zu lesen.',
  feld: '',
  werte: {},
}

describe('die Felder des Stils', () => {
  // **Ohne Felder steht nichts da** — die meisten Stile bringen keines mit,
  // das sich setzen ließe, und ein leerer Kasten wäre ein Rätsel.
  it('zeigt ohne Zusatzfelder nichts', async () => {
    const gefragt = vi.fn(() => Promise.resolve({ felder: [], fehler: null }))

    belegeApi({ Kartenzusatzfelder: gefragt })
    zeige()

    await waitFor(() => {
      expect(gefragt).toHaveBeenCalledWith('7', '3')
    })
    expect(
      screen.queryByRole('region', { name: texte.stilfelder }),
    ).not.toBeInTheDocument()
  })

  // **Bild- und Symbolfelder fehlen** — sie hängen an Wegen eines
  // Kartenfelds; das Textfeld daneben steht mit seinem Wert da.
  it('zeigt die Felder mit ihrem Wert, ohne Bild- und Symbolfelder', async () => {
    belegeApi({
      Kartenzusatzfelder: () =>
        Promise.resolve({
          felder: [
            zusatzfeld({ wert: 'Vorabdruck', gespeichert: true }),
            zusatzfeld({
              form: form({ name: 'bild', aufschrift: 'Bild', art: 'image' }),
            }),
            zusatzfeld({
              form: form({
                name: 'zeichen',
                aufschrift: 'Zeichen',
                art: 'symbol',
              }),
            }),
          ],
          fehler: null,
        }),
    })
    zeige()

    expect(
      await screen.findByRole('region', { name: texte.stilfelder }),
    ).toBeInTheDocument()
    expect(screen.getByText(texte.stilfelderHilfe)).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'vermerk' })).toHaveValue(
      'Vorabdruck',
    )
    expect(screen.queryByText('Bild')).not.toBeInTheDocument()
    expect(screen.queryByText('Zeichen')).not.toBeInTheDocument()
  })

  // **Geschrieben wird beim Verlassen**, wie bei jedem Textfeld — und danach
  // wird die Liste nachgeladen und die Vorschau benachrichtigt.
  it('schreibt einen Wert beim Verlassen und lädt nach', async () => {
    const gefragt = vi.fn(() =>
      Promise.resolve({ felder: [zusatzfeld()], fehler: null }),
    )
    const gesetzt = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({ Kartenzusatzfelder: gefragt, ZusatzwertSetzen: gesetzt })
    const { geglueckt } = zeige()

    await userEvent.type(
      await screen.findByRole('textbox', { name: 'vermerk' }),
      'Neu',
    )
    await userEvent.tab()

    await waitFor(() => {
      expect(geglueckt).toHaveBeenCalled()
    })
    expect(gesetzt).toHaveBeenCalledWith('7', '3', 'vermerk', 'Neu')
    expect(gefragt).toHaveBeenCalledTimes(2)
  })

  // **Ein Fehler kommt am Feld an** — über `melden`, wie beim Kartenfeld.
  it('meldet einen gescheiterten Wert am Feld', async () => {
    belegeApi({
      Kartenzusatzfelder: () =>
        Promise.resolve({ felder: [zusatzfeld()], fehler: null }),
      ZusatzwertSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'kein Zusatzfeld',
            feld: 'feld',
            werte: {},
          },
        }),
    })
    const { melden, geglueckt } = zeige()

    await userEvent.type(
      await screen.findByRole('textbox', { name: 'vermerk' }),
      'x',
    )
    await userEvent.tab()

    await waitFor(() => {
      expect(melden).toHaveBeenCalled()
    })
    expect(geglueckt).not.toHaveBeenCalled()
  })

  // **Und ein Fehler, den das Formular hält, steht an seinem Feld.**
  it('zeigt den Fehler des Formulars an seinem Feld', async () => {
    belegeApi({
      Kartenzusatzfelder: () =>
        Promise.resolve({ felder: [zusatzfeld()], fehler: null }),
    })
    zeige({ fehler: { feld: 'vermerk', meldung: 'Das geht so nicht.' } })

    expect(await screen.findByText('Das geht so nicht.')).toBeInTheDocument()
  })

  // **„Keine Zusatzfelder" und „nicht lesbar" sind zwei Antworten** —
  // scheitert das Lesen, steht die Meldung da, statt dass der Kasten wie bei
  // einem Stil ohne Felder verschwindet (Befund 7 des Nachlaufs über E/F).
  it('zeigt, dass die Felder nicht zu lesen waren', async () => {
    belegeApi({
      Kartenzusatzfelder: () =>
        Promise.resolve({ felder: [], fehler: lesefehler }),
    })
    zeige()

    const meldung = await screen.findByRole('alert')
    expect(meldung).toHaveTextContent(kataloge.de.bausteine.zustaende.fehler)
    expect(
      screen.getByRole('region', { name: texte.stilfelder }),
    ).toBeInTheDocument()
  })

  // **Geschrieben ist geschrieben** (§7, Regel 3): scheitert das Nachladen,
  // wird es gemeldet, und der Wert bleibt stehen.
  it('meldet ein gescheitertes Nachladen und behält den Wert', async () => {
    let runde = 0
    belegeApi({
      Kartenzusatzfelder: () => {
        runde += 1
        return Promise.resolve(
          runde === 1
            ? { felder: [zusatzfeld()], fehler: null }
            : { felder: [], fehler: lesefehler },
        )
      },
      ZusatzwertSetzen: () => Promise.resolve({ fehler: null }),
    })
    const { geglueckt, melden } = zeige()

    const feld = await screen.findByRole('textbox', { name: 'vermerk' })
    await userEvent.type(feld, 'Neu')
    await userEvent.tab()

    const meldung = await screen.findByRole('alert')
    expect(meldung).toHaveTextContent(kataloge.de.bausteine.zustaende.fehler)
    expect(geglueckt).toHaveBeenCalled()
    expect(melden).not.toHaveBeenCalled()
    expect(screen.getByRole('textbox', { name: 'vermerk' })).toHaveValue('Neu')
  })

  // **Ein Enter schickt nichts ab** — ein Formular mit einem einzigen
  // Textfeld sendet sich in WebKit selbst, und die Seite lüde neu.
  it('hält das Absenden des Kastens an', async () => {
    belegeApi({
      Kartenzusatzfelder: () =>
        Promise.resolve({ felder: [zusatzfeld()], fehler: null }),
    })
    zeige()

    const feld = await screen.findByRole('textbox', { name: 'vermerk' })
    const formular = feld.closest('form')
    if (formular === null) {
      throw new Error('das Feld steht in keinem Formular')
    }

    // `false` heißt: jemand hat `preventDefault` gerufen.
    expect(fireEvent.submit(formular)).toBe(false)
  })
})
