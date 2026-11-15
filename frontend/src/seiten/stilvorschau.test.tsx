// **Die Stilseite zeigt eine Karte**
// (#643).
//
// §6: eine Änderung ist in der Vorschau zu sehen, ohne dass jemand etwas
// bestätigt. Für ein Kartenfeld hielt das seit je; für eine **Stiloption**
// nicht — sie liegt auf der Seite *Stil*, die Vorschau lag auf der Seite
// *Karten*.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stil } from './Stil'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

// bild ist ein winziges PNG — der Inhalt zählt nicht, nur dass eines da ist.
const bild = 'iVBORw0KGgo='

// vorschau gibt eine gerasterte Karte, wie die Fassade sie liefert.
function vorschau() {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild,
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

function zeige(teile: Partial<Api>) {
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () =>
      Promise.resolve({
        optionen: [
          {
            name: 'randfarbe',
            art: 'choice',
            beschreibung: '',
            wert: 'schwarz',
            moeglichkeiten: [
              { wert: 'schwarz', beschriftung: 'schwarz' },
              { wert: 'weiß', beschriftung: 'weiß' },
            ],
            eigeneFarbe: false,
            gesetzt: false,
            hatVorgabe: false,
            hatWahlbilder: false,
            vorgabe: '',
          },
          {
            name: 'randbreite',
            art: 'text',
            beschreibung: '',
            wert: '2',
            moeglichkeiten: [],
            eigeneFarbe: false,
            gesetzt: false,
            hatVorgabe: false,
            hatWahlbilder: false,
            vorgabe: '',
          },
        ],
      }),
    StiloptionSetzen: () => Promise.resolve({ fehler: null }),
    Karten: () =>
      Promise.resolve({
        karten: [{ id: '1', nummer: 0, werte: [] }],
      }),
    Vorschau: () => Promise.resolve(vorschau()),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Stil setID="7" geaendert={() => undefined} />
    </Sprachraum>,
  )
}

describe('Die Vorschau auf der Stilseite', () => {
  it('zeigt eine Karte des Sets', async () => {
    zeige({})

    expect(
      await screen.findByAltText(texte.seiten.stil.vorschauTitel),
    ).toHaveAttribute('src', expect.stringContaining(bild))
  })

  // **Ohne Bestätigung** (§6): das Umstellen einer Option rechnet die Vorschau
  // neu, und niemand drückt dafür etwas.
  it('rechnet nach einer Änderung neu', async () => {
    const Vorschau = vi.fn(() => Promise.resolve(vorschau()))

    zeige({ Vorschau })

    await screen.findByAltText(texte.seiten.stil.vorschauTitel)

    expect(Vorschau).toHaveBeenCalledTimes(1)

    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: /randfarbe/ }),
      'weiß',
    )

    await waitFor(() => {
      expect(Vorschau).toHaveBeenCalledTimes(2)
    })
  })

  // **Und beim Tippen ebenso** —
  // #590.
  //
  // Eine Auswahl schreibt beim Umstellen; ein **Textfeld** schrieb erst beim
  // Verlassen, und Verlassen ist eine Bestätigung. Getippt wird die Vorschau
  // deshalb über [App.Stilprobevorschau] gerechnet — der Bestand sieht davon
  // nichts, dieselbe Bauart wie im Karteneditor seit #288.
  it('zeigt eine getippte Stiloption, ohne sie zu schreiben', async () => {
    const Stilprobevorschau = vi.fn(() => Promise.resolve(vorschau()))
    const StiloptionSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ Stilprobevorschau, StiloptionSetzen })

    await screen.findByAltText(texte.seiten.stil.vorschauTitel)

    await userEvent.type(
      screen.getByRole('textbox', { name: /randbreite/ }),
      '5',
    )

    // **Erst das Abschlusssignal, dann die Verneinung** (§3): der Aufruf
    // belegt, dass der Horcher hängt — eine Verneinung davor träfe auch dann
    // zu, wenn niemand zuhört.
    await waitFor(() => {
      expect(Stilprobevorschau).toHaveBeenCalledWith('1', 'randbreite', '25', 1)
    })

    expect(StiloptionSetzen).not.toHaveBeenCalled()
  })

  // **Ein Set ohne Karten sagt es** (§6), statt leer zu bleiben.
  it('erklärt ein Set ohne Karten', async () => {
    zeige({ Karten: () => Promise.resolve({ karten: [] }) })

    expect(
      await screen.findByText(texte.seiten.stil.vorschauLeer),
    ).toBeInTheDocument()

    expect(
      screen.queryByAltText(texte.seiten.stil.vorschauTitel),
    ).not.toBeInTheDocument()
  })

  // **Ein Befund steht da, statt die Seite leer zu lassen** (§8).
  it('meldet einen Befund', async () => {
    zeige({
      Vorschau: () =>
        Promise.resolve({
          vorschau: null,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Rasterer schweigt.',
          },
        }),
    })

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Rasterer schweigt.',
    )
  })
})

// **Das alte Bild bleibt stehen, während das neue rechnet** — die Ausnahme von
// §7 Regel 1, und bis zum 2026-08-22 hielt sie kein Test.
//
// `useLaden` räumt seit dem Zweitblick über Phase L von sich aus, wenn sich das
// Bezugsobjekt ändert. Für die Vorschau wäre das falsch: ihre Abhängigkeiten
// ändern sich mit **jedem Tastendruck**, und ein Bild, das dabei verschwindet
// und wiederkommt, ist schlechter als ein einen Augenblick altes.
//
// **Ohne diesen Test wäre die Ausnahme beim ersten Aufräumen weggefallen** —
// sie steht als drittes Argument da und sieht wie ein Versehen aus.
it('lässt das alte Bild stehen, während das neue rechnet', async () => {
  let aufloesen: ((wert: ReturnType<typeof vorschau>) => void) | undefined

  const Stilprobevorschau = vi.fn(
    () =>
      new Promise<ReturnType<typeof vorschau>>((fertig) => {
        aufloesen = fertig
      }),
  )

  zeige({ Stilprobevorschau })

  const bild = await screen.findByAltText(texte.seiten.stil.vorschauTitel)
  const vorher = bild.getAttribute('src')

  await userEvent.type(screen.getByRole('textbox', { name: /randbreite/ }), '5')

  await waitFor(() => {
    expect(Stilprobevorschau).toHaveBeenCalled()
  })

  // **Der neue Lauf hängt noch** — und das Bild steht.
  expect(screen.getByAltText(texte.seiten.stil.vorschauTitel)).toHaveAttribute(
    'src',
    vorher,
  )

  aufloesen?.(vorschau())
})
