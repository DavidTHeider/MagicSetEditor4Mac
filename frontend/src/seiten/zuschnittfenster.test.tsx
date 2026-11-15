// **Der Ausschnitt wird gewählt** —
// #576.
//
// Bis hierher schnitt das Einsetzen mit MSEs Vorgabewahl zu und war fertig.
// Wer ein Bild einsetzte, dessen Motiv nicht oben links sitzt, bekam den
// falschen Teil und hatte keinen Weg, das zu ändern.
//
// **Was hier geprüft wird, ist die Naht** — nicht die Regel: wohin ein Zug
// führt, entscheidet Go (`ImageSlice::constrain`,
// `ImageSliceSelector::onMotion`), und dort steht es geprüft. Hier zählt, dass
// die Wahl ankommt, dass eine überholte Antwort nicht einträgt und dass
// „nichts ändern" genau die Vorgabewahl einsetzt.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AuswahlDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { zuschnittantwort } from '../test/zuschnitt'
import { Zuschnittfenster } from './Zuschnittfenster'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function zeige(belegung: Record<string, unknown> = {}) {
  const uebernehmen = vi.fn()
  const abbrechen = vi.fn()

  belegeApi({
    Zuschnittgrenze: (
      _qb: number,
      _qh: number,
      _zb: number,
      _zh: number,
      auswahl: AuswahlDTO,
    ) => Promise.resolve(auswahl),
    ...belegung,
  })

  render(
    <Sprachraum sprache="de">
      <Zuschnittfenster
        zuschnitt={zuschnittantwort()}
        laeuft={false}
        uebernehmen={uebernehmen}
        abbrechen={abbrechen}
      />
    </Sprachraum>,
  )

  return { uebernehmen, abbrechen }
}

describe('Das Zuschnittfenster', () => {
  // **Die vier Zahlenfelder sind die Tastaturbedienung** (§6) — und sie sind
  // MSEs: `left`, `top`, `width`, `height` (`ImageSliceWindow`,
  // `src/gui/image_slice_window.cpp`, Zeilen 96–104, Fassung f9d9356 vom
  // 2021-07-06).
  it('zeigt die Vorgabewahl in den vier Zahlenfeldern', () => {
    zeige()

    expect(
      screen.getByLabelText(texte.seiten.karten.zuschnitt.links),
    ).toHaveValue(0)
    expect(
      screen.getByLabelText(texte.seiten.karten.zuschnitt.oben),
    ).toHaveValue(0)
    expect(
      screen.getByLabelText(texte.seiten.karten.zuschnitt.breite),
    ).toHaveValue(125)
    expect(
      screen.getByLabelText(texte.seiten.karten.zuschnitt.hoehe),
    ).toHaveValue(100)
  })

  // **Wer nichts ändert, setzt die Vorgabewahl ein** — der Abnahmekasten aus
  // #576, der am leichtesten stillschweigend bricht.
  it('reicht ohne Änderung die Vorgabewahl durch', async () => {
    const { uebernehmen } = zeige()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    )

    expect(uebernehmen).toHaveBeenCalledWith({
      x: 0,
      y: 0,
      breite: 125,
      hoehe: 100,
    })
  })

  // **Eine geänderte Zahl geht durch Gos Regel**, und was zurückkommt, steht
  // im Feld — nicht das Getippte.
  it('lässt Go über eine geänderte Zahl entscheiden', async () => {
    const Zuschnittgrenze = vi.fn(() =>
      Promise.resolve({ x: 40, y: 0, breite: 125, hoehe: 100 }),
    )

    const { uebernehmen } = zeige({ Zuschnittgrenze })

    const links = screen.getByLabelText(texte.seiten.karten.zuschnitt.links)
    await userEvent.clear(links)
    await userEvent.type(links, '999')

    await waitFor(() => {
      expect(links).toHaveValue(40)
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.karten.zuschnitt.uebernehmen,
      }),
    )

    expect(uebernehmen).toHaveBeenCalledWith({
      x: 40,
      y: 0,
      breite: 125,
      hoehe: 100,
    })
  })

  // **Die gezogene Kante wird genannt** — MSEs `PREFER_WIDTH` und
  // `PREFER_HEIGHT`: wer die Breite setzt, behält die Breite.
  it('nennt die gezogene Kante', async () => {
    const Zuschnittgrenze = vi.fn(() =>
      Promise.resolve({ x: 0, y: 0, breite: 60, hoehe: 48 }),
    )

    zeige({ Zuschnittgrenze })

    const breite = screen.getByLabelText(texte.seiten.karten.zuschnitt.breite)
    await userEvent.clear(breite)
    await userEvent.type(breite, '6')

    await waitFor(() => {
      expect(Zuschnittgrenze).toHaveBeenCalledWith(
        400,
        100,
        100,
        80,
        expect.anything(),
        false,
        'breite',
      )
    })
  })

  // **Das Seitenverhältnis lässt sich lösen** — MSEs „force to fit"
  // (`aspect_fixed`, Zeile 105). Angekreuzt heißt **gebunden**; die Ankreuzung
  // ist die Vorgabe, weil MSE `aspect_fixed = true` beginnt.
  it('reicht das gelöste Seitenverhältnis durch', async () => {
    const Zuschnittgrenze = vi.fn(() =>
      Promise.resolve({ x: 0, y: 0, breite: 125, hoehe: 100 }),
    )

    zeige({ Zuschnittgrenze })

    const wahl = screen.getByLabelText(
      texte.seiten.karten.zuschnitt.seitenverhaeltnis,
    )
    expect(wahl).toBeChecked()

    await userEvent.click(wahl)

    await waitFor(() => {
      expect(Zuschnittgrenze).toHaveBeenCalledWith(
        400,
        100,
        100,
        80,
        expect.anything(),
        true,
        '',
      )
    })
  })

  // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
  //
  // Jede Bewegung fragt Go. Käme die Antwort von vorhin nach der von jetzt an,
  // spränge der Rahmen zurück — ohne dass etwas fehlgeschlagen wäre.
  it('verwirft eine überholte Antwort', async () => {
    const antworten: ((wert: AuswahlDTO) => void)[] = []

    zeige({
      Zuschnittgrenze: () =>
        new Promise<AuswahlDTO>((aufloesen) => {
          antworten.push(aufloesen)
        }),
    })

    const links = screen.getByLabelText(texte.seiten.karten.zuschnitt.links)

    // **Zwei Änderungen und kein `clear`.** Bis Werkstück U3 von Phase 5.1
    // (#1414) erzeugte das Leeren selbst einen Ruf — `Number('')` ist `0` —,
    // und dieser Test lebte davon. Ein Test, dessen Mechanik an einem Fehler
    // hängt, wird grün, wenn der Fehler behoben wird (§10).
    fireEvent.change(links, { target: { value: '7' } })

    await waitFor(() => {
      expect(antworten.length).toBe(1)
    })

    fireEvent.change(links, { target: { value: '8' } })

    await waitFor(() => {
      expect(antworten.length).toBe(2)
    })

    // **Der jüngste zuerst, der älteste danach** — genau die Reihenfolge, die
    // ohne Wächter den alten Stand einträgt.
    antworten[1]?.({ x: 8, y: 0, breite: 125, hoehe: 100 })

    await waitFor(() => {
      expect(links).toHaveValue(8)
    })

    antworten[0]?.({ x: 999, y: 0, breite: 125, hoehe: 100 })

    // **Die Verneinung braucht die abgewartete Antwort** (§3): vorher träfe
    // sie ohnehin zu.
    await antwortAbwarten()

    expect(links).toHaveValue(8)
  })

  it('lässt sich abbrechen', async () => {
    const { abbrechen } = zeige()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.karten.zuschnitt.abbrechen,
      }),
    )

    expect(abbrechen).toHaveBeenCalled()
  })
})

// **Und die übrigen beiden Zahlenfelder** — jedes für sich, weil jedes eine
// andere Kante nennt.
describe('Die Zahlenfelder für Oben und Höhe', () => {
  it('reichen ihre Änderung mit der richtigen Kante durch', async () => {
    const Zuschnittgrenze = vi.fn(() =>
      Promise.resolve({ x: 0, y: 5, breite: 125, hoehe: 100 }),
    )

    zeige({ Zuschnittgrenze })

    const oben = screen.getByLabelText(texte.seiten.karten.zuschnitt.oben)
    await userEvent.clear(oben)
    await userEvent.type(oben, '5')

    await waitFor(() => {
      expect(Zuschnittgrenze).toHaveBeenCalledWith(
        400,
        100,
        100,
        80,
        // **Der Wert selbst ist hier nicht die Frage** — `clear` schickt
        // erst eine leere Eingabe, und was daraus wird, entscheidet Go. Was
        // dieses Feld beisteuert, ist die **Kante**.
        expect.anything(),
        false,
        '',
      )
    })

    // **Die Antwort abgewartet, nicht nur den Ruf** (#1809): sie setzt die
    // vier Felder neu, und ein Tippen davor liefe gegen den alten Stand.
    await antwortAbwarten()

    const hoehe = screen.getByLabelText(texte.seiten.karten.zuschnitt.hoehe)
    await userEvent.clear(hoehe)
    await userEvent.type(hoehe, '9')

    await waitFor(() => {
      expect(Zuschnittgrenze).toHaveBeenCalledWith(
        400,
        100,
        100,
        80,
        expect.anything(),
        false,
        'hoehe',
      )
    })
  })
})

// **Ein Zug am Rahmen geht durch Gos Ziehregel.**
//
// Was das Fenster beisteuert, ist der Maßstab — geprüft in
// `zuschnittrahmen.test.tsx`. Hier zählt, dass der Zug an der Bindung ankommt
// und dass die Antwort den Rahmen setzt.
describe('Ein Zug am Rahmen', () => {
  it('geht durch Zuschnittziehen', async () => {
    Element.prototype.setPointerCapture = vi.fn()

    const Zuschnittziehen = vi.fn(() =>
      Promise.resolve({ x: 30, y: 0, breite: 125, hoehe: 100 }),
    )

    zeige({ Zuschnittziehen })

    // **Über die Marke**: der Rahmen ist eine Zeigerhilfe und vor
    // Hilfsmitteln verborgen (G17).
    const rahmen = document.querySelector<HTMLElement>('[data-griff="rahmen"]')
    if (rahmen === null) {
      throw new Error('kein Rahmen')
    }

    fireEvent.pointerDown(rahmen, { clientX: 0, clientY: 0, pointerId: 1 })
    fireEvent.pointerMove(rahmen, { clientX: 30, clientY: 0, pointerId: 1 })
    fireEvent.pointerUp(rahmen, { pointerId: 1 })

    await waitFor(() => {
      expect(Zuschnittziehen).toHaveBeenCalled()
    })

    await waitFor(() => {
      expect(
        screen.getByLabelText(texte.seiten.karten.zuschnitt.links),
      ).toHaveValue(30)
    })
  })
  // **Ein leeres Feld ist kein Ausschnitt bei null** — Werkstück U3 von
  // Phase 5.1 (#1414).
  //
  // # Warum die vier Felder **ungeprüft** bleiben
  //
  // Weil die Regel in Go steht und **berichtigt statt abweist**:
  // `render.Zuschnittwahl.Begrenze` (portiert aus `ImageSlice::constrain`)
  // schiebt jede Fläche in die Quelle zurück, und der Rahmen zeigt danach,
  // was gilt. Eine zweite Fassung im Fenster wäre eine zweite Antwort auf
  // dieselbe Frage (§15).
  //
  // # Was hier trotzdem falsch war
  //
  // `Number('')` ist `0`. Wer das Feld leerte, um eine neue Zahl zu tippen,
  // schickte damit einen Zuschnitt bei null los — und der Rahmen sprang.
  it('schickt kein leeres Feld an die Grenzrechnung', () => {
    const Zuschnittgrenze = vi.fn((...args: unknown[]) =>
      Promise.resolve(args[4] as AuswahlDTO),
    )

    zeige({ Zuschnittgrenze })

    fireEvent.change(
      screen.getByLabelText(texte.seiten.karten.zuschnitt.links),
      { target: { value: '' } },
    )

    expect(Zuschnittgrenze).not.toHaveBeenCalled()
  })

  // **Und eine Null, die jemand tippt, geht durch** — sie ist ein gültiger
  // Ursprung, und der Unterschied zum leeren Feld ist der ganze Punkt.
  it('lässt eine getippte Null durch', async () => {
    const Zuschnittgrenze = vi.fn((...args: unknown[]) =>
      Promise.resolve(args[4] as AuswahlDTO),
    )

    zeige({ Zuschnittgrenze })

    const oben = screen.getByLabelText(texte.seiten.karten.zuschnitt.oben)

    // **Erst weg von der Null und dann zurück.** React meldet keine Änderung,
    // wenn der Wert derselbe bleibt — ein `0` auf ein `0` wäre kein Ereignis
    // und damit kein Beleg.
    fireEvent.change(oben, { target: { value: '5' } })

    // **Gewartet wird auf die Wirkung und nicht auf den Ruf** (§3): erst
    // wenn die Antwort im Feld steht, ist ein `0` darauf eine Änderung.
    await waitFor(() => {
      expect(oben).toHaveValue(5)
    })

    fireEvent.change(oben, { target: { value: '0' } })

    await waitFor(() => {
      expect(Zuschnittgrenze).toHaveBeenCalledTimes(2)
    })

    expect(Zuschnittgrenze.mock.calls[1]?.[4]).toMatchObject({ y: 0 })
  })
})
