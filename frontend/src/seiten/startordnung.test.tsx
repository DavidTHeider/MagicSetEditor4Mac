// **Was auf der Startseite führt und was dahinter gehört** — UXR-03
// (#1911).
//
// # Der Befund
//
// Gemessen am 2026-10-02 am nativen Fenster (`docs/UIREVIEW.md`, Commit
// `ddadb180`): **13 Überschriften** auf einer Seite, Rolltiefe 1,9 und eng
// 2,3 — viele gleichrangige Abschnitte untereinander: der Überblick mit einer
// Überschrift je Zahl, vier Listen unter *Weiterarbeiten*, die
// Aufmerksamkeit und ein leeres Buch der Hintergrundaufgaben.
//
// **Gemessen mit einem einzigen Probeset.** Ein Bestand, mit dem jemand
// arbeitet, hat sechs Schritte je Vorgang (`service.Hoechstens`), etwas
// Angeheftetes und mehr als drei Hinweise — und erst dort zeigt sich, dass
// jeder angeheftete Eintrag einen Hauptknopf trug (R1).
//
// # Was hier gilt
//
// Drei Überschriften — die Seite, *Weiterarbeiten*, *Was Aufmerksamkeit
// braucht* —, und **eine führende Handlung je Region**. Das Artboard *Main*
// zeigt beide Bereiche **nebeneinander**; ob sie es tun, rechnet jsdom nicht
// (§3) — das sieht die Hüllenprobe `uirevue`.
import { render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type { AufmerksamkeitAntwort } from '../apiaufmerksamkeit'
import type {
  ArbeitsschrittDTO,
  WeiterarbeitenAntwort,
} from '../apiweiterarbeiten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

function schritt(
  vorgang: ArbeitsschrittDTO['vorgang'],
  stelle: number,
): ArbeitsschrittDTO {
  return {
    objekt: {
      art: 'set',
      kennung: `${vorgang}-${String(stelle)}`,
      name: `Set ${vorgang} ${String(stelle)}`,
      angeheftet: false,
    },
    vorgang,
    zeitpunkt: `2026-09-09T${String(20 - stelle).padStart(2, '0')}:00:00Z`,
  }
}

const sechs = (vorgang: ArbeitsschrittDTO['vorgang']) =>
  Array.from({ length: 6 }, (_, stelle) => schritt(vorgang, stelle))

const weiterarbeiten: WeiterarbeitenAntwort = {
  zuletzt: schritt('geoeffnet', 0),
  geoeffnet: sechs('geoeffnet'),
  eingefuehrt: sechs('eingefuehrt'),
  ausgegeben: sechs('ausgegeben'),
  angeheftet: [
    { art: 'set', kennung: 'a', name: 'Kernset', angeheftet: true },
    { art: 'deck', kennung: 'b', name: 'Mono Rot', angeheftet: true },
  ],
  fehler: null,
}

// **Sechs Decks und drei Pakete** — seit AP-8 (#2365) teilen Pakete eine
// Zeile: neun Objekte, sieben Zeilen.
const aufmerksamkeit: AufmerksamkeitAntwort = {
  hinweise: Array.from({ length: 9 }, (_, stelle) => {
    const paket = stelle >= 6

    return {
      art: paket ? ('paketBefund' as const) : ('deckFehlend' as const),
      objekt: {
        art: paket ? ('paket' as const) : ('deck' as const),
        kennung: `objekt-${String(stelle)}`,
        name: `Objekt ${String(stelle)}`,
        angeheftet: false,
      },
      anzahl: 1,
      grund: '',
      ort: '',
      schluessel: '',
      werte: [],
    }
  }),
  durchgesehen: { sets: 18, decks: 1, pakete: 540 },
  befunde: [],
  fehler: null,
}

async function hinstellen() {
  belegeApi({
    Weiterarbeiten: () => Promise.resolve(weiterarbeiten),
    Aufmerksamkeit: () => Promise.resolve(aufmerksamkeit),
  })

  render(
    <Sprachraum sprache="de">
      <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={0} />
    </Sprachraum>,
  )

  await screen.findByRole('button', { name: /^Weiter an/ })
  await screen.findByRole('button', { name: 'Alle 7 zeigen' })
  await screen.findByRole('list', { name: texte.ueberblick.titel })
}

describe('die Ordnung der Startseite', () => {
  it('steht unter drei Überschriften, auch mit vollem Bestand', async () => {
    await hinstellen()

    expect(
      screen.getAllByRole('heading').map((kopf) => kopf.textContent),
    ).toEqual([
      texte.ueberschrift,
      texte.weiterarbeiten.titel,
      `${texte.aufmerksamkeit.titel} 9`,
    ])
  })

  // **R1 aus `docs/UIREVIEW.md`**, so gezählt wie die Messung: Knöpfe im
  // Hauptgewicht (`bg-marke`) je Region.
  it('führt in jeder Region mit höchstens einem Hauptknopf', async () => {
    await hinstellen()

    const regionen = screen.getAllByRole('region')
    expect(regionen.length).toBeGreaterThan(1)

    for (const region of regionen) {
      const haupt = within(region)
        .queryAllByRole('button')
        .filter((knopf) => knopf.classList.contains('bg-marke'))

      expect(
        haupt.length,
        region.getAttribute('aria-label') ?? '',
      ).toBeLessThanOrEqual(1)
    }
  })

  // **R2 aus `docs/UIREVIEW.md`**: höchstens zwanzig Bedienelemente, ohne
  // dass jemand etwas aufklappt. jsdom kennt kein „im Bild"; gezählt wird,
  // was gezeichnet ist — die strengere Zahl. **Vorher 51**: vier Kacheln,
  // „Weiter an …", zwanzig Zeilen mit je zwei Knöpfen und die
  // Aufmerksamkeit.
  it('zeichnet mit vollem Bestand höchstens zwanzig Bedienelemente', async () => {
    await hinstellen()

    expect(screen.getAllByRole('button').length).toBeLessThanOrEqual(20)
  })

  // **Weiterarbeiten und Aufmerksamkeit stehen nebeneinander**, wie im
  // Artboard — ab einem Fenster, das breit genug ist. Geprüft wird die
  // Zusage im Baum: beide stehen in **einem** Raster, das ab `lg` zwei
  // Spalten hat. Wie breit es tatsächlich wird, zeigt nur das Fenster.
  it('stellt Weiterarbeiten und Aufmerksamkeit in ein Raster', async () => {
    await hinstellen()

    const weiter = screen.getByRole('region', {
      name: texte.weiterarbeiten.titel,
    })
    const achtung = screen.getByRole('region', {
      name: texte.aufmerksamkeit.titel,
    })
    const raster = weiter.parentElement

    expect(raster).not.toBeNull()
    expect(achtung.parentElement).toBe(raster)
    expect(raster?.className).toMatch(/(^|\s)lg:grid-cols-2(\s|$)/)
  })
})
