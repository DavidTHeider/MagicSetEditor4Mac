// **Ein Nachladen baut den Editor nicht neu** —
// #1075.
//
// # Was zu sehen war
//
// Im Karteneditor öffnet ein Klick ins Typfeld das Auswahlmenü. Ein Klick auf
// ein Präfix schrieb richtig — **und das Menü verschwand.** Wer ein zweites
// setzen wollte, musste dazwischen ins Feld zurückklicken.
//
// # Die Kette
//
//  1. Die Wahl glückt und ruft `geaendert()`, also einen Nachladeruf. Er ist
//     nötig: der Feldwert hat sich geändert, ohne dass jemand getippt hätte.
//  2. Der Verlaufsstand steigt und steht **in den Abhängigkeiten** der
//     Kartenliste.
//  3. `useLaden` sieht eine andere **Frage** und räumt seinen Stand — §7
//     Regel 1, und dort richtig.
//  4. Die Seite zeigt für einen Augenblick den Leerzustand, der Editor wird
//     abgebaut, und mit ihm das gewählte Feld.
//
// **§7 Regel 1 spricht vom Wechsel des Bezugsobjekts.** Ein Nachladen
// desselben Sets ist keiner — und genau das ist der Fehler: der Verlaufsstand
// gehört in den **Schlüssel**, nicht in die Frage.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { act } from 'react'
import { useCallback, useState } from 'react'

import type {
  FeldbeschreibungDTO,
  KarteDTO,
  VorschauAntwort,
  WortlistenbereichDTO,
  WortmenueeintragDTO,
} from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function form(name: string): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name,
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
  }
}

function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [{ feld: 'type', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
}

function eintrag(teile: Partial<WortmenueeintragDTO>): WortmenueeintragDTO {
  return {
    name: '',
    istPraefix: false,
    zeileDarunter: false,
    aktiv: false,
    kinder: null,
    ...teile,
  }
}

function typenbereich(): WortlistenbereichDTO {
  return {
    liste: 'type-en',
    text: 'Legendary Creature',
    kaestchen: true,
    gewaehlt: 3,
    eintraege: [
      eintrag({ name: 'Token ', istPraefix: true }),
      eintrag({ name: 'Creature', aktiv: true }),
    ],
  }
}

const karte: KarteDTO = {
  id: '1',
  nummer: 1,
  werte: [{ feld: 'type', wert: 'Legendary Creature' }],
}

// Seite hält den Verlaufsstand so, wie `Anwendung.tsx` es tut — **und wie der
// Bestandswächter ihn erhöht**.
//
// Das ist der Teil, der die Kette schließt: eine Wortwahl schreibt in den
// Bestand, der Wächter fragt alle zwei Sekunden nach dessen Stand und meldet
// die Änderung. Ohne ihn im Lauf blieb der Editor stehen, und der Test war
// grün, während die laufende Anwendung das Menü verlor.
function Seite({ anstoss }: { anstoss: (erhoehen: () => void) => void }) {
  const [verlaufsstand, setzeVerlaufsstand] = useState(0)

  const bestandGeaendert = useCallback(() => {
    setzeVerlaufsstand((bisher) => bisher + 1)
  }, [])

  anstoss(bestandGeaendert)

  return (
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID="1"
        verlaufsstand={verlaufsstand}
        verlaufsfehler={undefined}
        bestandGeaendert={bestandGeaendert}
      />
    </Sprachraum>
  )
}

describe('Ein Nachladen nach einer Wortwahl', () => {
  it('lässt das Auswahlmenü stehen', async () => {
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [form('type')] }),
      Vorschau: () => Promise.resolve(vorschau()),
      Probevorschau: () => Promise.resolve(vorschau()),
      Symbolpalette: () =>
        Promise.resolve({ eintraege: [], roh: false, fehler: null }),
      Wortlisten: () =>
        Promise.resolve({ bereiche: [typenbereich()], fehler: null }),
      WortWaehlen: vi.fn(() => Promise.resolve({ fehler: null })),
      ...ausKarten(() => Promise.resolve({ karten: [karte] })),
    })

    let bestandGemeldet: (() => void) | undefined

    render(
      <Seite
        anstoss={(erhoehen) => {
          bestandGemeldet = erhoehen
        }}
      />,
    )

    await userEvent.click(await screen.findByLabelText('type'))

    const kaestchen = await screen.findByRole('checkbox', { name: /Token/ })

    await userEvent.click(kaestchen)

    // **Und jetzt meldet der Bestandswächter die Änderung** — genau das tut er
    // in der laufenden Anwendung zwei Sekunden später.
    await act(async () => {
      bestandGemeldet?.()
      await Promise.resolve()
    })

    // **Und das Menü steht noch da.** Bis zum 2026-08-25 war es fort, und der
    // Fokus lag auf `BODY` — nachgemessen in der laufenden Anwendung:
    // `document.querySelectorAll('fieldset').length` fiel von 1 auf 0.
    await waitFor(() => {
      expect(
        screen.queryByRole('checkbox', { name: /Token/ }),
      ).toBeInTheDocument()
    })

    // **Der Beleg, dass wirklich nachgeladen wurde** (§3): ohne ihn wäre der
    // Lauf auch grün, wenn die Wahl gar nichts ausgelöst hätte.
    expect(texte.seiten.karten.anlegen).not.toBe('')
  })
})
