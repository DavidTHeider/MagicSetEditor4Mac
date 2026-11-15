// **Überholte Läufe der Einstellungen** (§7, Regel 2) —
// #245.
//
// Eine eigene Datei neben `anwendungseinstellungen.test.tsx`, weil beide
// zusammen über 400 Zeilen gingen (§2). Der Schnitt folgt der Frage: dort
// steht, dass die Einstellungen wirken; hier, dass eine Antwort zu dem Stand
// gehört, aus dem sie stammt.
import { StrictMode } from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { groessen } from './erscheinungsbild'
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
          // **Die Liste, wie Go sie schickt**: mit neun Werten ist die
          // Sprache eine Auswahlliste (UXR-F2, #2008).
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

describe('Überholte Läufe der Einstellungen', () => {
  //
  // Der Fall ist erreichbar und nicht theoretisch: wer eine Einstellung
  // umschaltet, während der erste Lauf noch unterwegs ist, bekäme sonst den
  // **alten** Stand eingetragen, nachdem der neue schon dastand — die Auswahl
  // spränge zurück, ohne dass etwas fehlgeschlagen wäre.
  // **Unter `StrictMode`, und das ist der Betrieb**: `main.tsx` hängt die
  // Anwendung genau so ein. React ruft Effekte dort beim Aufbau **zweimal** —
  // zwei Ladeläufe, deren Reihenfolge beim Zurückkommen niemand zusagt.
  //
  // Hier kommt der erste **nach** dem zweiten zurück. Trüge er sein Ergebnis
  // noch ein, stünde die Schriftgröße auf `normal`, obwohl im Bestand
  // `sehrgross` steht — und niemand sähe einen Fehler.
  it('trägt einen überholten Ladelauf nicht mehr ein', async () => {
    let erstenLaufBeenden: () => void = () => undefined

    const Einstellungen = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((erfuellen) => {
            erstenLaufBeenden = () => {
              erfuellen(antwort({ sprache: 'de' })())
            }
          }),
      )
      .mockImplementation(
        antwort({ sprache: 'de', schriftgroesse: 'sehrgross' }),
      )

    belegeApi({ Einstellungen })

    render(
      <StrictMode>
        <App />
      </StrictMode>,
    )

    // **Erst der zweite Lauf** — er kommt sofort zurück.
    //
    // **Und darauf wird gewartet, nicht auf die Überschrift.** Sie steht da,
    // sobald die Anwendung zeichnet, und hat mit dem Ladelauf nichts zu tun:
    // unter Last erschien sie, bevor die Antwort eingetragen war, und der Test
    // war rot — einmal in etwa fünfzehn Gesamtläufen, nie allein. Ein Test, der
    // eine asynchrone Wirkung an einer fremden Marke festmacht, prüft die
    // Reihenfolge des Rechners.
    await waitFor(() => {
      expect(
        document.documentElement.style.getPropertyValue('--schriftgroesse'),
      ).toBe(groessen.sehrgross)
    })

    // **Jetzt erst der erste** — und er findet einen Lauf vor, der nicht mehr
    // seiner ist.
    erstenLaufBeenden()

    await waitFor(() => {
      expect(Einstellungen).toHaveBeenCalledTimes(2)
    })

    // **Auch hier auf die Wirkung gewartet**, nicht auf die Zahl der Aufrufe:
    // dass der überholte Lauf **zurück** ist, heißt nicht, dass React schon
    // gezeichnet hat. Dieselbe Klasse wie oben, und sie ist beim dritten
    // roten Gesamtlauf in Folge behoben worden — an **allen** Stellen, nicht
    // an der genannten (§10).
    await waitFor(() => {
      expect(
        document.documentElement.style.getPropertyValue('--schriftgroesse'),
      ).toBe(groessen.sehrgross)
    })
  })

  // **Die Sprachwahl wirkt auch auf die native Menüleiste.** Sie liegt in Go
  // und weiß von React nichts; ohne den Aufruf stünde die Oberfläche in der
  // neuen Sprache und die Leiste in der alten.
  it('meldet die Sprache an die Menüleiste', async () => {
    const SpracheSetzen = vi.fn(() => Promise.resolve({}))

    belegeApi({ Einstellungen: antwort({ sprache: 'en' }), SpracheSetzen })

    render(<App />)

    // **Auf den Spion warten, nicht auf die Überschrift** (§3).
    //
    // Die englische Überschrift kommt aus dem Sprachkontext; die Meldung an
    // die Menüleiste ist ein eigener Effekt. Die eine ist kein Beleg für die
    // andere — und unter Last war sie es nicht: dieser Fall war im vollen
    // Lauf rot und einzeln grün.
    await waitFor(() => {
      expect(SpracheSetzen).toHaveBeenCalledWith('en')
    })

    // **Und die Seite steht danach auf Englisch** — die zweite Zusage, jetzt
    // an einer Stelle, an der die erste schon eingelöst ist.
    expect(
      await screen.findByRole('heading', {
        name: kataloge.en.seiten.start.ueberschrift,
      }),
    ).toBeInTheDocument()
  })

  // **Eine überholte Meldung wird nicht mehr eingetragen** (§7, Regel 2).
  //
  // Zwei Sprachwechsel kurz hintereinander: der erste Aufruf an die Menüleiste
  // kommt zurück, **nachdem** der zweite Wechsel schon feststeht. Trüge er sein
  // Ergebnis noch ein, stünde an der Einstellungsseite ein Fehler zu einer
  // Sprache, die niemand mehr eingestellt hat — oder eine Entwarnung, die für
  // den vorigen Stand galt.
  //
  // **Seit dem Nachlauf über Phase 6.2 geht der zweite Aufruf erst hinaus,
  // wenn der erste zurück ist** (`sprache/menue.ts`): Go setzt die Leiste bei
  // jedem Ruf, und zwei zugleich konnten dort umgekehrt ankommen. Hier wartete
  // der Lauf bis dahin auf zwei gleichzeitige Aufrufe — genau das, was jetzt
  // nicht mehr geschieht.
  it.each([
    ['ein überholter Fehlschlag', true],
    ['ein überholter Erfolg', false],
  ])('trägt %s nicht mehr ein', async (_name, scheitert) => {
    let erstenAufrufBeenden: () => void = () => undefined

    const SpracheSetzen = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((erfuellen, ablehnen) => {
            erstenAufrufBeenden = () => {
              if (scheitert) {
                ablehnen(new Error('zu spät'))

                return
              }

              erfuellen({})
            }
          }),
      )
      .mockImplementation(() => Promise.resolve({}))

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

    // **Die Oberfläche steht schon auf Englisch** — der Wechsel ist
    // eingetragen, während der erste Aufruf noch in Go hängt.
    await waitFor(() => {
      expect(
        screen.getByRole('heading', {
          name: kataloge.en.seiten.einstellungen.titel,
        }),
      ).toBeInTheDocument()
    })

    // **Der zweite wartet**, bis der erste zurück ist.
    expect(SpracheSetzen).toHaveBeenCalledTimes(1)

    // **Jetzt erst kommt der erste zurück** — und findet einen Lauf vor, der
    // nicht mehr seiner ist. Danach geht der zweite hinaus.
    erstenAufrufBeenden()

    await waitFor(() => {
      expect(SpracheSetzen).toHaveBeenLastCalledWith('en')
    })

    expect(SpracheSetzen).toHaveBeenCalledTimes(2)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
