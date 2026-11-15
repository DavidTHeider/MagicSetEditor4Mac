// Die Einstellungen wirken auf die ganze Anwendung —
// #245.
//
// Eine eigene Datei neben `App.test.tsx`, weil beide sonst über 400 Zeilen
// gingen (§2). Der Schnitt folgt der Frage: dort steht, dass die Anwendung
// zeichnet; hier, dass sie es **so** tut, wie es im Bestand steht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { anzeigeschriften, groessen } from './erscheinungsbild'
import { kataloge } from './sprache/kataloge'
import { belegeApi, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
  document.documentElement.removeAttribute('data-erscheinungsbild')
  document.documentElement.style.removeProperty('--schriftgroesse')
  document.documentElement.style.removeProperty('--font-anzeige')
})

const texte = kataloge.de

// antwort baut eine Einstellungsantwort mit den gegebenen Werten.
interface Wahl {
  sprache?: string
  erscheinungsbild?: string
  schriftgroesse?: string
  schriftart?: string
}

function antwort(wahl: Wahl) {
  return () =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'sprache',
          wert: wahl.sprache ?? '',
          wirksam: wahl.sprache ?? '',
          // **Die Liste, wie Go sie schickt** (`docs/EINSTELLUNGEN.md`): mit
          // neun Werten ist die Sprache eine Auswahlliste. Zwei Werte wären
          // seit UXR-F2 (#2008) zwei Segmente — eine Vorrichtung, die
          // großzügiger ist als der Betrieb.
          werte: ['erben', 'de', 'en', 'es', 'fr', 'it', 'ja', 'ru', 'zh'],
          bereich: 'anwendung',
        },
        {
          schluessel: 'erscheinungsbild',
          wert: wahl.erscheinungsbild ?? 'system',
          wirksam: wahl.erscheinungsbild ?? 'system',
          werte: ['system', 'hell', 'dunkel'],
          bereich: 'anwendung',
        },
        {
          schluessel: 'schriftgroesse',
          wert: wahl.schriftgroesse ?? 'normal',
          wirksam: wahl.schriftgroesse ?? 'normal',
          werte: ['klein', 'normal', 'gross', 'sehrgross'],
          bereich: 'anwendung',
        },
        {
          schluessel: 'schriftart',
          wert: wahl.schriftart ?? 'pergament',
          wirksam: wahl.schriftart ?? 'pergament',
          werte: ['pergament', 'fantasy', 'antike', 'kalligrafie'],
          bereich: 'anwendung',
        },
      ],
    })
}

// zurEinstellungsseite geht über die Navigationsleiste dorthin.
async function zurEinstellungsseite() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: texte.seiten.einstellungen.titel,
    }),
  )
}

// zumThema geht in die Gruppe, in der das Erscheinungsbild steht.
//
// **Seit UX-04 liegt es nicht mehr neben der Sprache**
// (#1676):
// der Entwurf trennt „Sprache, Start, Kürzel" von „Thema, Schrift, Fenster".
async function zumThema() {
  await userEvent.click(
    await screen.findByRole('tab', {
      name: texte.seiten.einstellungen.gruppen.erscheinungsbild,
    }),
  )
}

describe('Die Einstellungen der Anwendung', () => {
  // **Das Erscheinungsbild steht am Wurzelelement**, nicht in React: es gilt
  // für alles, auch für den Rollbalken.
  it('schreibt das gewählte Erscheinungsbild ans Wurzelelement', async () => {
    belegeApi({
      Einstellungen: antwort({ sprache: 'de', erscheinungsbild: 'dunkel' }),
    })

    render(<App />)

    // **Gewartet wird auf die Wirkung, nicht auf die Überschrift.**
    //
    // Sie steht da, sobald die Anwendung zeichnet, und hat mit dem Ladelauf
    // nichts zu tun. Unter Last erschien sie, bevor die Antwort eingetragen
    // war — der Lauf war rot, etwa einmal in fünf vollen Pipelines, und nie
    // allein. Ein Test, der eine asynchrone Wirkung an einer fremden Marke
    // festmacht, prüft die Reihenfolge des Rechners.
    await waitFor(() => {
      expect(
        document.documentElement.getAttribute('data-erscheinungsbild'),
      ).toBe('dunkel')
    })
  })

  it('nimmt das Attribut für die Systemeinstellung weg', async () => {
    document.documentElement.setAttribute('data-erscheinungsbild', 'hell')

    belegeApi({ Einstellungen: antwort({ sprache: 'de' }) })

    render(<App />)

    // **Und hier auf das Verschwinden** — dieselbe Regel, andersherum: das
    // Attribut steht anfangs noch da.
    await waitFor(() => {
      expect(
        document.documentElement.hasAttribute('data-erscheinungsbild'),
      ).toBe(false)
    })
  })

  it('setzt die gewählte Schriftgröße', async () => {
    belegeApi({
      Einstellungen: antwort({ sprache: 'de', schriftgroesse: 'sehrgross' }),
    })

    render(<App />)

    await waitFor(() => {
      expect(
        document.documentElement.style.getPropertyValue('--schriftgroesse'),
      ).toBe(groessen.sehrgross)
    })
  })

  // **Die Anzeigeschrift geht denselben Weg wie die Größe** — Werkstück J zu
  // #1325.
  // Geprüft wird am Wurzelelement, weil sie dort steht: eine Zusicherung über
  // eine Komponente sagte nichts über die Überschrift, die daneben liegt.
  it('setzt die gewählte Anzeigeschrift', async () => {
    belegeApi({
      Einstellungen: antwort({ sprache: 'de', schriftart: 'fantasy' }),
    })

    render(<App />)

    await waitFor(() => {
      expect(
        document.documentElement.style.getPropertyValue('--font-anzeige'),
      ).toBe(anzeigeschriften.fantasy)
    })
  })

  // **Geschrieben wird über die Naht, und danach wird neu gelesen** — was
  // gilt, sagt der Bestand. Wer den Wert örtlich einträgt, zeigt eine Wahl,
  // die vielleicht abgewiesen wurde.
  it('schreibt eine Änderung und liest neu', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({}))
    const Einstellungen = vi
      .fn()
      .mockImplementationOnce(antwort({ sprache: 'de' }))
      .mockImplementation(
        antwort({ sprache: 'de', erscheinungsbild: 'dunkel' }),
      )

    belegeApi({ Einstellungen, EinstellungSetzen })

    render(<App />)
    await zurEinstellungsseite()
    await zumThema()

    // **Drei Werte sind seit UXR-F2 (#2008) drei Segmente.**
    await userEvent.click(
      await screen.findByRole('radio', {
        name: texte.seiten.einstellungen.werte['erscheinungsbild.dunkel'],
      }),
    )

    expect(EinstellungSetzen).toHaveBeenCalledWith('erscheinungsbild', 'dunkel')

    // **Gewartet, nicht sofort geprüft.** Der zweite Lauf geht über die
    // Fassade; das Segment steht erst danach auf dem neuen Wert.
    await waitFor(() => {
      expect(
        screen.getByRole('radio', {
          name: texte.seiten.einstellungen.werte['erscheinungsbild.dunkel'],
        }),
      ).toBeChecked()
    })
  })

  // **Ein gescheitertes Schreiben lässt den alten Wert stehen** — und das ist
  // richtig, denn er steht ja noch. Gemeldet wird es an der Seite.
  it('meldet ein gescheitertes Schreiben und behält den Wert', async () => {
    belegeApi({
      Einstellungen: antwort({ sprache: 'de' }),
      EinstellungSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            meldung: 'Der Bestand nimmt es nicht.',
          },
        }),
    })

    render(<App />)
    await zurEinstellungsseite()
    await zumThema()

    await userEvent.click(
      await screen.findByRole('radio', {
        name: texte.seiten.einstellungen.werte['erscheinungsbild.dunkel'],
      }),
    )

    expect(
      screen.getByRole('radio', {
        name: texte.seiten.einstellungen.werte['erscheinungsbild.system'],
      }),
    ).toBeChecked()
  })

  // **Ohne Wahl und ohne bekannte Systemsprache gilt Deutsch** (§9,
  // Ausgangssprache). Unter jsdom ist `navigator.language` `en-US`; der Fall
  // „das System spricht etwas, das wir nicht führen" braucht deshalb einen
  // eigenen Aufbau.
  it('fällt auf Deutsch zurück, wenn das System eine fremde Sprache spricht', async () => {
    const urspruenglich = Object.getOwnPropertyDescriptor(
      window.navigator,
      'language',
    )

    Object.defineProperty(window.navigator, 'language', {
      // **Nicht mehr `fr-FR`**: Französisch führt das Fenster seit LOC-01.
      // Portugiesisch hat ein Sprachpaket im Pack, aber keinen Katalog.
      value: 'pt-BR',
      configurable: true,
    })

    belegeApi({ Einstellungen: antwort({}) })

    render(<App />)

    expect(
      await screen.findByRole('heading', {
        name: texte.seiten.start.ueberschrift,
      }),
    ).toBeInTheDocument()

    if (urspruenglich) {
      Object.defineProperty(window.navigator, 'language', urspruenglich)
    }
  })

  // **Eine Sprachwahl wechselt die Oberfläche und meldet die Leiste nach.**
  //
  // Der Weg geht durch den Bestand: schreiben, neu lesen, neue Sprache. Ohne
  // diesen Test stünde nur fest, dass die Sprache beim **Start** ankommt.
  it('wechselt die Sprache über die Einstellungsseite', async () => {
    const SpracheSetzen = vi.fn(() => Promise.resolve({}))
    const Einstellungen = vi
      .fn()
      .mockImplementationOnce(antwort({ sprache: 'de' }))
      .mockImplementation(antwort({ sprache: 'en' }))

    belegeApi({
      Einstellungen,
      SpracheSetzen,
      EinstellungSetzen: () => Promise.resolve({}),
    })

    render(<App />)
    await zurEinstellungsseite()
    await userEvent.selectOptions(
      await screen.findByRole('combobox', {
        name: texte.seiten.einstellungen.namen.sprache,
      }),
      'en',
    )

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3, #1321).
    // `toHaveBeenCalled` trifft zu, sobald gerufen wurde — die Antwort ist
    // dann noch unterwegs, und was sie bewirkt, steht noch nicht da.
    expect(
      await screen.findByRole('heading', {
        name: kataloge.en.seiten.einstellungen.titel,
      }),
    ).toBeInTheDocument()

    expect(SpracheSetzen).toHaveBeenCalledWith('en')
  })
})
