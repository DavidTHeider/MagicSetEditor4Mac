// **Eigenschaften und Werkzeuge fahren als Schubladen über die Kartenfläche**
// — UX-08.AK-01 und S06.NAT-02
// (#1680,
// #1686),
// nach den Artboards *Eigenschaften* und *Werkzeuge* des Entwurfs.
//
// # Warum durch die ganze Anwendung
//
// **Weil die Zusage eine Naht ist**: der Knopf im Kopf, der Seitenwechsel im
// Fenster, die Schublade über der Fläche und die Karten darunter. Jedes Stück
// für sich sähe richtig aus, auch wenn die Karten beim Öffnen verschwänden.
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from '../App'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { leisteneintrag } from '../test/leistenweg'

const texte = kataloge.de
const worte = texte.allgemein.arbeitsplatz.schubladen

beforeEach(() => {
  const wege = ausKarten(vi.fn(() => Promise.resolve({ karten: [] })))

  belegeApi({
    ...belegeFenster('de'),
    ...wege,
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '7',
            name: 'Probeset',
            spiel: 'magic',
            stil: 'magic-m15',
            karten: 0,
            herkunft: '/a.mse-set',
            herkunftFehlt: false,
          },
        ],
      }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Setangaben: () =>
      Promise.resolve({
        felder: [],
        werte: [],
        kopf: {
          mseFassung: '',
          spielfassung: '',
          stilfassung: '',
          versionskontrolle: '',
          apprenticecode: '',
        },
      }),
    Statistik: () =>
      Promise.resolve({
        statistik: { karten: 0, felder: 0, leere: 0, belegung: [] },
      }),
    Verteilung: () => Promise.resolve({ verteilung: [] }),
  })
})

afterEach(() => {
  raeumeApi()
})

// **Nach der Schublade wird mit ihrem Namen gefragt, nie nach „einem Dialog".**
// In diesem Aufbau steht die Einführung vom ersten Start offen, ebenfalls
// ein Dialog; „kein Dialog mehr" wäre nie wahr — oder nur, weil Mantine sie
// beim selben Escape mitschließt.
async function setOeffnen() {
  render(<App />)

  await userEvent.click(await leisteneintrag('sets'))
  await userEvent.click(
    await screen.findByRole('button', {
      name: `Probeset — ${texte.seiten.sets.oeffnen}`,
    }),
  )

  // **Die Kartenfläche steht**, bevor irgendeine Schublade aufgeht — auf sie
  // kommt es an: sie muss es danach noch.
  return screen.findByText(texte.seiten.karten.leerTitel)
}

describe('Die Schubladen des Sets', () => {
  it('öffnet die Eigenschaften über der Kartenfläche, bei Stil', async () => {
    const flaeche = await setOeffnen()

    await userEvent.click(
      screen.getByRole('button', { name: worte.eigenschaften.knopf }),
    )

    const schublade = await screen.findByRole('dialog', {
      name: worte.eigenschaften.titel,
    })

    const laschen = within(schublade).getAllByRole('tab')
    expect(laschen.map((lasche) => lasche.textContent)).toEqual([
      texte.seiten.stil.titel,
      texte.seiten.setangaben.titel,
      texte.seiten.symbol.titel,
      texte.seiten.schluesselwoerter.titel,
    ])

    const stil = within(schublade).getByRole('tab', {
      name: texte.seiten.stil.titel,
    })
    expect(stil).toHaveAttribute('aria-selected', 'true')

    // **Der Fokus steht auf der gewählten Lasche** — und nicht auf dem ×,
    // das Mantine sonst als erstes Bedienelement nähme. Von dort wandern die
    // Pfeile, und Tab führt in die Seite.
    await waitFor(() => {
      expect(stil).toHaveFocus()
    })

    // **Die Karten bleiben darunter** — dasselbe Element, nicht ein neues.
    expect(flaeche).toBeInTheDocument()

    // **Und sie sind gesperrt, solange die Schublade steht**: kein Fokus,
    // kein Klick, kein Vorleseprogramm darunter. Mantine setzt kein
    // `aria-hidden` auf den Rest (`Schublade.tsx`) — `inert` ist die Sperre.
    expect(flaeche.closest('[inert]')).not.toBeNull()
  })

  it('gibt die Karten frei, wenn die Schublade zugeht', async () => {
    const flaeche = await setOeffnen()

    await userEvent.click(
      screen.getByRole('button', { name: worte.eigenschaften.knopf }),
    )
    await screen.findByRole('dialog', { name: worte.eigenschaften.titel })

    await userEvent.click(
      screen.getByRole('button', { name: worte.schliessen }),
    )

    await waitFor(() => {
      expect(
        screen.queryByRole('dialog', { name: worte.eigenschaften.titel }),
      ).toBeNull()
    })
    expect(flaeche.closest('[inert]')).toBeNull()
  })

  it('wechselt die Lasche in der Schublade und bleibt dabei offen', async () => {
    await setOeffnen()

    await userEvent.click(
      screen.getByRole('button', { name: worte.eigenschaften.knopf }),
    )

    const schublade = await screen.findByRole('dialog', {
      name: worte.eigenschaften.titel,
    })

    await userEvent.click(
      within(schublade).getByRole('tab', {
        name: texte.seiten.setangaben.titel,
      }),
    )

    await waitFor(() => {
      expect(
        within(
          screen.getByRole('dialog', { name: worte.eigenschaften.titel }),
        ).getByRole('tab', { name: texte.seiten.setangaben.titel }),
      ).toHaveAttribute('aria-selected', 'true')
    })
  })

  it('schließt mit Escape, und der Fokus steht wieder auf dem Knopf', async () => {
    const flaeche = await setOeffnen()

    const knopf = screen.getByRole('button', {
      name: worte.werkzeuge.knopf,
    })
    await userEvent.click(knopf)

    const schublade = await screen.findByRole('dialog', {
      name: worte.werkzeuge.titel,
    })

    // **Escape gilt, wenn der Fokus darin steht** (`Schublade.tsx`) — also
    // erst, wenn Mantine ihn hineingesetzt hat.
    await waitFor(() => {
      expect(schublade).toContainElement(document.activeElement as HTMLElement)
    })

    expect(
      within(schublade)
        .getAllByRole('tab')
        .map((lasche) => lasche.textContent),
    ).toEqual([
      texte.seiten.statistik.titel,
      texte.seiten.pack.titel,
      texte.seiten.konsole.titel,
    ])

    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(
        screen.queryByRole('dialog', { name: worte.werkzeuge.titel }),
      ).toBeNull()
    })

    await waitFor(() => {
      expect(knopf).toHaveFocus()
    })

    expect(flaeche).toBeInTheDocument()
  })

  // **Der Fokus kehrt zum Knopf zurück, auch wenn der Klick ihn nicht
  // fokussiert hat.** WebKit gibt einem Knopf beim Klick keinen Fokus — die
  // Gewohnheit von macOS —, und dann merkte sich die Schublade `body` als
  // Auslöser. Gefunden in der Sichtprüfung im Browser: nach Escape stand der
  // Fokus nirgends. `fireEvent.click` klickt wie WebKit, ohne zu fokussieren;
  // `userEvent.click` fokussiert wie Chrome und sähe den Fehler nicht.
  it('gibt den Fokus an den Knopf zurück, auch ohne Fokus beim Klick', async () => {
    await setOeffnen()

    const knopf = screen.getByRole('button', {
      name: worte.eigenschaften.knopf,
    })
    // **Vorher steht der Fokus woanders**, und `fireEvent.click` bringt ihn
    // nicht dorthin — anders als `userEvent.click`.
    expect(knopf).not.toHaveFocus()
    fireEvent.click(knopf)

    const schublade = await screen.findByRole('dialog', {
      name: worte.eigenschaften.titel,
    })
    await waitFor(() => {
      expect(schublade).toContainElement(document.activeElement as HTMLElement)
    })

    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(
        screen.queryByRole('dialog', { name: worte.eigenschaften.titel }),
      ).toBeNull()
    })
    await waitFor(() => {
      expect(knopf).toHaveFocus()
    })
  })

  // **Ein Kürzel wechselt die Schublade, und der Fokus geht mit.** Ohne das
  // stünde er auf einer Lasche, die es nicht mehr gibt — also auf `body` —,
  // und weder Escape noch die Pfeile kämen an.
  it('wechselt per Kürzel in die andere Schublade und nimmt den Fokus mit', async () => {
    await setOeffnen()

    const knopf = screen.getByRole('button', {
      name: worte.eigenschaften.knopf,
    })
    await userEvent.click(knopf)

    const stil = within(
      await screen.findByRole('dialog', { name: worte.eigenschaften.titel }),
    ).getByRole('tab', { name: texte.seiten.stil.titel })
    await waitFor(() => {
      expect(stil).toHaveFocus()
    })

    // **⌥⌘8 ist die Statistik** — in der Schublade *Werkzeuge*.
    fireEvent.keyDown(window, {
      key: '8',
      code: 'Digit8',
      metaKey: true,
      altKey: true,
    })

    const statistik = within(
      await screen.findByRole('dialog', { name: worte.werkzeuge.titel }),
    ).getByRole('tab', { name: texte.seiten.statistik.titel })
    await waitFor(() => {
      expect(statistik).toHaveFocus()
    })

    // **Und Escape führt zum Knopf, mit dem alles anfing.**
    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(
        screen.queryByRole('dialog', { name: worte.werkzeuge.titel }),
      ).toBeNull()
    })
    await waitFor(() => {
      expect(knopf).toHaveFocus()
    })
  })

  it('führt die Seiten der Schubladen nicht mehr in der Leiste', async () => {
    await setOeffnen()

    const leiste = screen.getByRole('tablist', {
      name: texte.allgemein.arbeitsplatz.laschen,
    })

    for (const verstaut of [
      texte.seiten.stil.titel,
      texte.seiten.konsole.titel,
      texte.seiten.statistik.titel,
    ]) {
      expect(within(leiste).queryByRole('tab', { name: verstaut })).toBeNull()
    }

    expect(
      within(leiste).getByRole('tab', { name: texte.seiten.karten.titel }),
    ).toHaveAttribute('aria-selected', 'true')
  })
})
