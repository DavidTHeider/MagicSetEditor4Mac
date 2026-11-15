// **Die Setliste nach dem Artboard *Sets*** — UXR-04
// (#1912).
//
// # Was gemessen war
//
// Am nativen Fenster: drei Hauptknöpfe auf der Seite, *Set öffnen …* führte
// im Kopf, und Anlage und Einfuhr standen **unter** der Liste — mit einem Set
// im Probebestand noch im Bild, mit dreißig unter der Falz, wo nur ⌘N sie
// fand. Das Artboard stellt die Liste links und das Neue rechts daneben, im
// Kopf führt *+ Neues Set*, *Datei öffnen …* steht daneben.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

function satz(nummer: number): SetDTO {
  return {
    id: String(nummer),
    name: `Set ${String(nummer)}`,
    spiel: 'probe',
    stil: 'probe-schlicht',
    herkunft: `/pfad/set${String(nummer)}.mse-set`,
    karten: nummer,
    herkunftFehlt: false,
  }
}

function paket(art: string, name: string) {
  return {
    art,
    name,
    anzeigename: name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 0,
    artrang: 0,
    herkunftsart: 'mitgeliefert' as const,
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
    hauptdatei: '',
  }
}

function zeige() {
  belegeApi({
    Sets: () => Promise.resolve({ sets: [satz(1)] }),
    Pakete: () =>
      Promise.resolve({
        pakete: [
          paket('mse-game', 'probe'),
          paket('mse-style', 'probe-schlicht'),
        ],
        befunde: [],
        fehler: null,
      }),
  })

  return render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={vi.fn()}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )
}

// kopf sind die Handlungen des Seitenkopfs — rechts oben, dort, wo jede
// Seite sie trägt. Links daneben stehen Titel und Erklärung samt ihrem „?".
function kopf(behaelter: HTMLElement) {
  const gefunden = behaelter.querySelector('header')?.lastElementChild
  if (!(gefunden instanceof HTMLElement)) {
    throw new Error('kein Seitenkopf')
  }

  return gefunden
}

describe('Der Kopf der Setliste', () => {
  // **Kein gefüllter Knopf im Kopf** — AP-3 (#2360): »Neues Set« springt
  // in das Anlegen, das daneben offen steht, und dessen »Weiter« ist der
  // Schritt, der führt (`Wege09`). Zwei gefüllte für denselben Weg sagten
  // nicht, welcher dran ist; je Blick führt einer (`Zustaende09`).
  it('trägt »Neues Set« und »Set öffnen …« im Kopf im Nebengewicht', async () => {
    const { container } = zeige()
    await screen.findByRole('region', { name: worte.titel })

    const knoepfe = within(kopf(container)).getAllByRole('button')
    expect(knoepfe.map((knopf) => knopf.textContent)).toEqual([
      worte.setOeffnen,
      `+${worte.neuKopf}`,
    ])

    const haupt = knoepfe.filter((knopf) =>
      /(^|\s)bg-marke(\s|$)/.test(knopf.className),
    )
    expect(haupt).toEqual([])

    // **Er kündigt ⌘N an** — es tut dasselbe: zum Namen des neuen Sets.
    expect(
      within(kopf(container)).getByRole('button', { name: worte.neuKopf }),
    ).toHaveAttribute('aria-keyshortcuts', 'Meta+N')
  })

  it('springt mit »Neues Set« zum Namen — auch ein zweites Mal', async () => {
    const { container } = zeige()
    const neu = within(kopf(container)).getByRole('button', {
      name: worte.neuKopf,
    })

    await userEvent.click(neu)

    const name = await screen.findByLabelText(worte.neuName)
    await waitFor(() => {
      expect(name).toHaveFocus()
    })

    // **Ein zweiter Druck findet das Feld wieder**, nachdem jemand woanders
    // war — sonst täte der Knopf beim zweiten Mal nichts Sichtbares.
    await userEvent.click(
      screen.getByLabelText(kataloge.de.bausteine.listenansicht.suche),
    )
    expect(name).not.toHaveFocus()

    await userEvent.click(neu)
    await waitFor(() => {
      expect(screen.getByLabelText(worte.neuName)).toHaveFocus()
    })
  })
})

describe('Liste und Anlage', () => {
  // **Nebeneinander, sobald Platz ist** — die Liste links, das Neue rechts,
  // wie im Artboard. Geprüft wird die Zusage im Baum: beide stehen in einem
  // Raster, das ab `lg` zwei Spalten hat, und die Liste vorn — im engen
  // Fenster steht sie damit zuerst. Wie breit es wird, zeigt nur das Fenster.
  it('stehen in einem Raster, die Liste vorn', async () => {
    zeige()

    const liste = await screen.findByRole('region', { name: worte.titel })
    const einfuhr = screen.getByRole('button', { name: worte.einlesenTitel })
    const raster = liste.closest('div.grid')

    expect(raster?.contains(einfuhr)).toBe(true)
    expect(raster?.className).toMatch(
      /(^|\s)lg:grid-cols-\[minmax\(0,1fr\)_21rem\](\s|$)/,
    )
    expect(
      liste.compareDocumentPosition(einfuhr) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()

    // **Die Einfuhr steht eingeklappt in der Spalte daneben** — offen ist
    // eine Gruppe, nicht zwei (R3). **Das Anlegen steht dort nicht mehr**:
    // ab dem ersten Set liegt es in einer Schublade (AP-10,
    // `setanlageschublade.test.tsx`).
    expect(einfuhr).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('region', { name: worte.neuTitel })).toBeNull()
  })
})
