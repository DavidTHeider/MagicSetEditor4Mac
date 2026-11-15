// **Was die Menüleiste im Fenster auslöst** — Werkstück 4 von Phase 3.7
// (#1168,
// Befund Z4-01).
//
// # Warum gegen die ganze Anwendung
//
// **Weil die Zusage eine Naht ist**: „Undo/Redo rufen außerhalb von
// Texteingaben denselben Verlauf wie ⌘Z". Der Verlauf hängt am Fenster, die
// Ereignisse kommen aus Go, und dazwischen liegt `Anwendung.tsx`. Ein Test
// gegen den Haken allein prüfte die eine Hälfte.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import type { BewegungAntwort } from './formen'
import {
  ereignisAllesAuswaehlen,
  ereignisRueckgaengig,
  ereignisSeite,
  ereignisUeber,
  ereignisWiederherstellen,
} from './fenster'
import { kataloge } from './sprache/kataloge'
import { effekteNachholen } from './test/effekte'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { ausKarten } from './test/kartenproben'
import { leisteneintrag } from './test/leistenweg'

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

// horcher sammelt die Rückrufe, die das Fenster anmeldet.
//
// **Eine Laufzeit-Attrappe und kein echtes Wails**: `EventsOn` ist die ganze
// Naht, die es hier braucht — was Wails sonst noch kann, ist für diese Frage
// ohne Belang.
function laufzeitAttrappe(): Map<string, (...daten: unknown[]) => void> {
  const angemeldet = new Map<string, (...daten: unknown[]) => void>()

  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
      angemeldet.set(name, rueckruf)

      return () => angemeldet.delete(name)
    },
  }

  return angemeldet
}

const texte = kataloge.de

// zeige baut die Anwendung und gibt die angemeldeten Rückrufe.
function zeige() {
  const horcher = laufzeitAttrappe()

  render(<App />)

  return horcher
}

// beimSet schlägt das Set auf, das `belegeSet` anbietet.
//
// **Für den Verlauf ist das kein Beiwerk**: `useVerlauf` hält still, solange
// keine `setID` feststeht (§7, Regel 4 braucht einen Schlüssel). Wer die
// Menüeinträge ohne offenes Set prüfte, sähe zu, wie nichts geschieht, und
// hielte es für die Zusage.
async function beimSet() {
  await userEvent.click(await leisteneintrag('sets'))

  await userEvent.click(
    await screen.findByRole('button', { name: /^Gewähltes Set —/ }),
  )

  // **Gesucht im Inhalt**: seit UX-03.AK-02 steht die aufgeschlagene Karte
  // auch in der Brotkrume darüber.
  await within(screen.getByRole('main')).findByText('Karte aus 42')
}

// belegeSet ist die Fassade mit einem Set und einer Karte darin.
function belegeSet(weiteres: Parameters<typeof belegeApi>[0] = {}) {
  const karten = (setID: string) =>
    Promise.resolve({
      karten: [
        {
          id: '1',
          nummer: 0,
          werte: [{ feld: 'name', wert: `Karte aus ${setID}` }],
        },
      ],
    })

  belegeApi({
    ...belegeFenster('de'),
    Startbefund: () =>
      Promise.resolve({
        offen: true,
        grund: '',
        schluessel: '',
        fassung: '9.9.9',
        rueckspielrest: '',
        rueckweg: null,
      }),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '42',
            name: 'Gewähltes Set',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            karten: 1,
            herkunftFehlt: false,
          },
        ],
      }),
    ...ausKarten(karten),
    Karten: karten,
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    ...weiteres,
  })
}

describe('Die Menüleiste im Fenster', () => {
  // **Rückgängig ruft den fachlichen Verlauf** — dieselbe Bindung wie ⌘Z.
  //
  // Das ist die Zusage des Werkstücks: die Leiste greift **nach** dem
  // Verlauf, nicht daneben.
  it('ruft mit Rückgängig und Wiederherstellen den Verlauf', async () => {
    // **Eine Bewegung, die nichts bewegt** — mehr braucht dieser Fall nicht:
    // geprüft ist, dass der Menüeintrag den Verlauf **ruft**. Was eine
    // Bewegung auslöst, steht in `verlauf.test.ts`.
    const stand: BewegungAntwort = {
      bewegt: false,
      karteID: '',
      feld: '',
      umfang: 0,
      urheber: '',
      fehler: null,
    }

    const zurueck = vi.fn(() => Promise.resolve(stand))
    const vor = vi.fn(() => Promise.resolve(stand))

    belegeSet({ Zuruecknehmen: zurueck, Wiederherstellen: vor })

    const horcher = zeige()

    await beimSet()

    horcher.get(ereignisRueckgaengig)?.()

    await waitFor(() => {
      expect(zurueck).toHaveBeenCalled()
    })

    horcher.get(ereignisWiederherstellen)?.()

    await waitFor(() => {
      expect(vor).toHaveBeenCalled()
    })
  })

  // **Ein Seitenschlüssel führt auf die Seite.**
  it('schlägt die genannte Seite auf', async () => {
    belegeSet()

    const horcher = zeige()

    await waitFor(() => {
      expect(horcher.has(ereignisSeite)).toBe(true)
    })

    horcher.get(ereignisSeite)?.('pakete')

    expect(
      await screen.findByRole('heading', {
        name: texte.seiten.pakete.titel,
        level: 1,
      }),
    ).toBeInTheDocument()
  })

  // **Ein Schlüssel, den es nicht gibt, führt nirgendwohin.**
  //
  // §8 und dieselbe Regel wie bei jeder Nutzlast über die Ereignisnaht: sie
  // wird geprüft, nicht geglaubt. Ein blindes `as` machte aus einem
  // Tippfehler in Go ein leeres Fenster — `Seiteninhalt` hat für einen
  // Schlüssel, den es nicht gibt, keinen Zweig und gibt nichts zurück.
  //
  // **Erst ein guter Schlüssel, dann die schlechten**, und dazwischen ein
  // Nachholen (§3): eine Zusicherung unmittelbar nach dem Rückruf trifft auf
  // einen Baum, den React noch nicht neu gezeichnet hat, und ist grün, ohne
  // etwas geprüft zu haben. Die Gegenprobe hat es gezeigt — mit entschärfter
  // Prüfung blieb sie zunächst grün.
  it('bleibt stehen, wenn der Schlüssel keine Seite benennt', async () => {
    belegeSet()

    const horcher = zeige()

    await waitFor(() => {
      expect(horcher.has(ereignisSeite)).toBe(true)
    })

    // **Der Beleg, dass jemand zuhört** (§3): dieselbe Naht, ein gültiger
    // Schlüssel, eine sichtbare Wirkung.
    horcher.get(ereignisSeite)?.('pakete')

    await screen.findByRole('heading', {
      name: texte.seiten.pakete.titel,
      level: 1,
    })

    horcher.get(ereignisSeite)?.('gibtsnicht')
    horcher.get(ereignisSeite)?.('')
    horcher.get(ereignisSeite)?.(42)

    await effekteNachholen()

    expect(screen.getByRole('heading', { level: 1 })).toHaveAccessibleName(
      texte.seiten.pakete.titel,
    )
  })

  // **„Über" führt auf die Hilfe und rollt zum Abschnitt.**
  //
  // AGENTS.md §11 sagt seit jeher, die Rechtelage stehe „in der Anwendung
  // selbst unter »Über«" — sie stand dort bis zum 2026-08-27 nicht.
  //
  // **Gemessen wird am Rollen und nicht an der Seite.** Der Abschnitt steht
  // auf der Hilfeseite, gleich wie man dorthin kam; wer nur ihn zusicherte,
  // bliebe grün, wenn der Menüeintrag bloß die Seite aufschlüge und die
  // ganze zweite Hälfte des Weges fehlte. Die Gegenprobe hat genau das
  // gezeigt. **Dass wirklich gerollt wird, sieht kein Test** — jsdom rechnet
  // kein Layout (`test/einrichtung.ts`); geprüft ist, dass der Anker gefunden
  // wird und **an welchem Abschnitt** er hängt.
  it('führt mit Über auf den Abschnitt und nennt die Fassung', async () => {
    const gerollt = vi.spyOn(Element.prototype, 'scrollIntoView')

    belegeSet()

    const horcher = zeige()

    await waitFor(() => {
      expect(horcher.has(ereignisUeber)).toBe(true)
    })

    horcher.get(ereignisUeber)?.()

    const ueberschrift = await screen.findByRole('heading', {
      name: texte.hilfe.ueber.ueberschrift,
    })

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3, #1321).
    expect(await screen.findByText('9.9.9')).toBeInTheDocument()

    // **Der Anker hängt am Abschnitt**, nicht irgendwo auf der Seite.
    expect(gerollt).toHaveBeenCalled()
    expect(gerollt.mock.instances[0]).toContainElement(ueberschrift)

    // **Und die Rechtelage steht dabei** — der Satz, den §11 verlangt.
    expect(screen.getByText(/Wizards of the Coast/)).toBeInTheDocument()

    gerollt.mockRestore()
  })

  // **Alle Karten anhaken** —
  // #1242.
  //
  //	void ItemList::doSelectAll() { … }   // wählt dort ALLE Karten
  //
  // `src/gui/control/item_list.cpp`, Zeilen 43–64, Fassung `f9d9356` vom
  // 2021-07-06. Hier stand als Begründung, *Alles auswählen* sei nur Sache
  // der Textmarke — das stimmte nicht, und der Nachlauf über Phase 3.7 hat es
  // angestrichen.
  //
  // **Die sichtbare Wirkung ist die Massenänderung**: sie erscheint, wenn
  // etwas angehakt ist. An den Anhakhaken hängt sie, und die Zwischenablage
  // dazu.
  it('hakt mit »Alle Karten anhaken« jede Karte an', async () => {
    belegeSet()

    const horcher = zeige()

    await beimSet()
    await effekteNachholen()

    horcher.get(ereignisAllesAuswaehlen)?.()

    expect(
      await screen.findByText(texte.seiten.karten.massen.ueberschrift),
    ).toBeInTheDocument()
  })
})
