// **Aufnehmen und Einlesen in Schubladen** — UXR-10
// (#1918),
// nach dem Artboard `Sammlung`: im Kopf führt *+ Aufnehmen*, daneben steht
// das Einlesen, und der Navigator trägt die Sicht.
//
// # Der Anlass
//
// **Am nativen Fenster gemessen, 2026-10-02** (`docs/UIREVIEW.md`): bei
// leerer Sammlung standen im Navigator — 240 Punkte breit — das Formular
// mit vierzehn Feldern, die Listeneinfuhr, die Ordnung und die Ausgabe
// offen untereinander: **18 Felder** (R3 lässt zwölf zu). Mit dem ersten
// Exemplar klappte alles in einen Kasten, und der Bericht der Einfuhr, die
// das ausgelöst hatte, ging mit (vermerkt in `huellenprobe/fix-f03.mjs`).
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { verstoesse } from '../test/axe'
import { ohneDublette } from '../test/einfuhr'
import {
  bestand,
  belegeSammlung,
  felder,
  hauptknoepfeJeRegion,
  oeffneAufnehmen,
  oeffneEinfuhr,
  texte,
  zeigeSammlung,
} from '../test/sammlungsseite'

// **jsdom rechnet kein Layout**: die Fensterliste liest ihre Höhe aus
// `clientHeight`, und die stellen wir.
beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => 600,
  })
})

afterEach(() => {
  raeumeApi()
})

// grenze ist R3 aus `docs/UIREVIEW.md`.
const grenze = 12

describe('Die Sammlung bei leerem Bestand', () => {
  it('zeichnet höchstens zwölf Felder', async () => {
    belegeSammlung([])
    zeigeSammlung()

    await screen.findByText(texte.leerTitel)

    expect(felder()).toBeLessThanOrEqual(grenze)
  })

  // **Der leere Zustand nennt den nächsten Schritt als Knopf**
  // (`docs/UX-ARCHITEKTUR.md`, Zustände) — bis hierher war er das offene
  // Formular im Navigator.
  it('bietet beide Wege als Knopf an', async () => {
    belegeSammlung([])
    zeigeSammlung()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.neuesTitel }),
    )
    const schublade = await screen.findByRole('dialog', {
      name: texte.neuesTitel,
    })
    expect(
      within(schublade).getByLabelText(texte.drucksuche.karte),
    ).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.listeTitel }),
    )
    expect(
      within(
        await screen.findByRole('dialog', { name: texte.einfuhrTitel }),
      ).getByLabelText(texte.listeFeld),
    ).toBeInTheDocument()
  })

  // **Der Bericht bleibt stehen** — vorher klappte der Kasten mit dem ersten
  // Exemplar zu, und das Einfuhrergebnis verschwand mit ihm.
  it('zeigt den Bericht einer Einfuhr auch, wenn danach etwas dasteht', async () => {
    const gelesen = vi
      .fn()
      .mockResolvedValueOnce({ exemplare: [], fehler: null })
      .mockResolvedValue({ exemplare: bestand(2), fehler: null })

    belegeSammlung([], {
      Sammlung: gelesen,
      ListeVorschauen: ohneDublette,
      ListeDeuten: () =>
        Promise.resolve({
          deutung: {
            zeilen: [
              {
                roh: '4 Lightning Bolt (M11) 149',
                zeilennummer: 1,
                menge: 4,
                name: 'Lightning Bolt',
                setcode: 'm11',
                nummer: '149',
                finish: 'nonfoil',
                treffer: [
                  {
                    printing: 'druck-1',
                    name: 'Lightning Bolt',
                    setcode: 'm11',
                    setname: 'Magic 2011',
                    nummer: '149',
                    seltenheit: 'common',
                    bild: '',
                  },
                ],
                drucke: 1,
                eindeutig: true,
                befund: '',
                weitere: false,
                fehler: null,
              },
            ],
            befunde: [],
            alterSekunden: 0,
          },
          fehler: null,
        }),
      ListeEinlesen: () =>
        Promise.resolve({
          bericht: {
            angelegt: [1],
            erhoeht: 0,
            ausgelassen: 0,
            stueck: 4,
            befunde: [],
          },
          fehler: null,
        }),
    })
    const geschrieben = vi.fn()
    zeigeSammlung(geschrieben)

    await screen.findByText(texte.leerTitel)
    const schublade = await oeffneEinfuhr()

    await userEvent.type(
      within(schublade).getByLabelText(texte.listeFeld),
      '4 Lightning Bolt (M11) 149',
    )
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.listeDeuten }),
    )
    await userEvent.click(
      await within(schublade).findByRole('button', {
        name: texte.listeEinlesen,
      }),
    )

    await screen.findByText('Karte 0001')
    expect(
      within(
        screen.getByRole('dialog', { name: texte.einfuhrTitel }),
      ).getByText('1 neu, 0 erhöht, 4 Stück aufgenommen.'),
    ).toBeInTheDocument()

    // **Und die Einfuhr meldet sich dem Fenster als eigenes Schreiben** —
    // Nachlauf zum Zweitblick über Phase 6, zweiter Teil, R3: sonst las
    // der Abgleich des offenen Decks nicht neu.
    expect(geschrieben).toHaveBeenCalled()
  })
})

describe('Die Schubladen der Sammlung', () => {
  it('führt im Kopf mit Aufnehmen, und das Formular steht in der Schublade', async () => {
    belegeSammlung(bestand(400))
    zeigeSammlung()

    await screen.findByText('Karte 0001')

    // **Nicht auf der Seite**, sondern hinter dem führenden Knopf.
    expect(screen.queryByLabelText(texte.drucksuche.karte)).toBeNull()
    expect(hauptknoepfeJeRegion()).toBeLessThanOrEqual(1)

    const knopf = screen.getByRole('button', { name: texte.aufnehmenOeffnen })
    expect(knopf).toHaveClass('bg-marke')

    const schublade = await oeffneAufnehmen()

    expect(
      within(schublade).getByLabelText(texte.drucksuche.karte),
    ).toBeInTheDocument()
    expect(knopf.closest('[inert]')).not.toBeNull()
    expect(hauptknoepfeJeRegion(schublade)).toBe(1)

    // **Auch die Schublade ist keine Formularwand** (R3): sieben Felder
    // offen, das Übrige unter „Weitere Angaben".
    expect(felder(schublade)).toBeLessThanOrEqual(grenze)
  })

  it('liest über die zweite Schublade ein, die Ausgabe eingeklappt darunter', async () => {
    belegeSammlung(bestand(400))
    zeigeSammlung()

    const schublade = await oeffneEinfuhr()

    expect(
      within(schublade).getByLabelText(texte.listeFeld),
    ).toBeInTheDocument()
    expect(
      within(schublade).getByRole('button', { name: texte.ausfuhrTitel }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(
      within(schublade).queryByRole('button', { name: texte.ausfuhrAlsCSV }),
    ).toBeNull()
  })

  // **Der Fokus steht im Feld „Karte“** — AP-10 (#2367), Kernweg 6a mit
  // dem Budget 4: ⌘3, „Aufnehmen …“, Name, ↩, „Aufnehmen“. Stünde er auf
  // dem Schließknopf, käme ein Griff dazu.
  it('gibt beim Öffnen dem Feld „Karte“ den Fokus', async () => {
    belegeSammlung(bestand(3))
    zeigeSammlung()

    const schublade = await oeffneAufnehmen()

    await waitFor(() => {
      expect(
        within(schublade).getByLabelText(texte.drucksuche.karte),
      ).toHaveFocus()
    })
  })

  // **Schließen verwirft nichts** — dieselbe Zusage wie beim Aufklappen
  // (UX-08 S11.NAT-03).
  it('behält eine halb getippte Eingabe über das Schließen hinweg', async () => {
    belegeSammlung(bestand(3))
    zeigeSammlung()

    const schublade = await oeffneAufnehmen()
    await userEvent.type(await druckfeld(schublade), 'abc')

    await userEvent.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })
    expect(
      screen.getByRole('button', { name: texte.aufnehmenOeffnen }),
    ).toHaveFocus()

    const wieder = await oeffneAufnehmen()
    expect(await druckfeld(wieder)).toHaveValue('abc')
  })

  // **Ein Fehler steht, wo geklickt wurde** (§6) — unter der Schublade sähe
  // ihn niemand.
  it('meldet einen Fehler beim Aufnehmen in der Schublade', async () => {
    belegeSammlung(bestand(3), {
      ExemplarAnlegen: () =>
        Promise.resolve({
          id: 0,
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'so nicht',
          },
        }),
    })
    zeigeSammlung()

    const schublade = await oeffneAufnehmen()
    await userEvent.type(await druckfeld(schublade), 'x')
    await userEvent.click(
      within(schublade).getByRole('button', { name: texte.aufnehmen }),
    )

    expect(await within(schublade).findByRole('alert')).toBeInTheDocument()
  })

  it('klappt die Ordnung im Navigator ein', async () => {
    belegeSammlung(bestand(400))
    zeigeSammlung()

    await screen.findByText('Karte 0001')

    expect(
      screen.getByRole('button', { name: texte.ordnungTitel }),
    ).toHaveAttribute('aria-expanded', 'false')
    expect(
      screen.queryByRole('button', { name: texte.ordnungDubletten }),
    ).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: texte.ordnungTitel }),
    )
    expect(
      screen.getByRole('button', { name: texte.ordnungDubletten }),
    ).toBeInTheDocument()
  })

  it('hält WCAG 2.2 AA in beiden Schubladen ein', async () => {
    belegeSammlung(bestand(3))
    zeigeSammlung()

    expect(await verstoesse(await oeffneAufnehmen())).toEqual([])

    await userEvent.keyboard('{Escape}')
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    expect(await verstoesse(await oeffneEinfuhr())).toEqual([])
  })
})
