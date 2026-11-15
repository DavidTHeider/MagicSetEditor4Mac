// **Die Zwischenablage der Kartenlasche** — Werkstück 1 von Phase 3.7
// (#1168).
//
// **Gegen die ganze Seite und nicht gegen den Knopfkasten**: was hier zu
// prüfen ist, ist die Naht zwischen Anhakung, Griff und den zwei Bedienungen
// — und die verläuft durch `Karten.tsx`.
//
// **Seit UXR-06 (#1914) über das Menü „Diese Karte"**: der Kasten „Ablage
// und Bezug" mit seinen drei Knöpfen ist fort, die Einträge tragen dieselben
// Namen. Wo ein Lauf ihn nur öffnete, um auf die Seite zu warten, wartet er
// jetzt auf die Zeile der Liste.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { waehleKartenhandlung, waehleMehrere } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const eigene = texte.seiten.karten.ablage

function probeKarte(id: string, name: string): KarteDTO {
  return { id, nummer: 0, werte: [{ feld: 'name', wert: name }] }
}

const zwei = [probeKarte('1', 'Feuerball'), probeKarte('2', 'Mirri')]

function belegeSeite(teile: Partial<Api>) {
  return belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    ...ausKarten(() => Promise.resolve({ karten: zwei })),
    Karten: () => Promise.resolve({ karten: zwei }),
    ...teile,
  })
}

function zeige(bestandGeaendert = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="1"
        karteID={undefined}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={bestandGeaendert}
      />
    </Sprachraum>,
  )

  return bestandGeaendert
}

// ohneFehler ist die Antwort, die alle drei Wege im Erfolgsfall geben.
const ohneFehler = { fehler: null }

describe('Die Zwischenablage', () => {
  // **Ohne Haken die aufgeschlagene Karte** — wie MSE2, dessen Liste immer
  // eine gewählte Zeile hat.
  it('kopiert die aufgeschlagene Karte, wenn nichts angehakt ist', async () => {
    const kopieren = vi.fn(() => Promise.resolve(ohneFehler))

    belegeSeite({ AblageKopieren: kopieren })
    zeige()
    await waehleKartenhandlung(eigene.kopieren.replace('{zahl}', '1'))

    await waitFor(() => {
      expect(kopieren).toHaveBeenCalledWith('1', ['1'])
    })
  })

  // **Dieselbe Auswahl wie die Massenänderung.**
  it('kopiert, was angehakt ist', async () => {
    const kopieren = vi.fn(() => Promise.resolve(ohneFehler))

    belegeSeite({ AblageKopieren: kopieren })
    zeige()
    await waehleMehrere()

    for (const name of ['Feuerball', 'Mirri']) {
      await userEvent.click(
        await screen.findByRole('checkbox', {
          name: new RegExp(
            `${texte.seiten.karten.massen.auswaehlen}.*${name}|${name}.*${texte.seiten.karten.massen.auswaehlen}`,
          ),
        }),
      )
    }

    await waehleKartenhandlung(eigene.kopieren.replace('{zahl}', '2'))

    await waitFor(() => {
      expect(kopieren).toHaveBeenCalledWith('1', ['1', '2'])
    })
  })

  // **Regel 3**: geschrieben ist geschrieben — die Karten sind fort, also
  // wird nachgeladen.
  it('lädt nach dem Ausschneiden nach', async () => {
    const ausschneiden = vi.fn(() => Promise.resolve(ohneFehler))
    const geaendert = vi.fn()

    belegeSeite({ AblageAusschneiden: ausschneiden })
    zeige(geaendert)
    await waehleKartenhandlung(eigene.ausschneiden.replace('{zahl}', '1'))

    await waitFor(() => {
      expect(ausschneiden).toHaveBeenCalledWith('1', ['1'])
    })

    expect(geaendert).toHaveBeenCalled()
  })

  it('lädt nach dem Einfügen nach', async () => {
    const einfuegen = vi.fn(() =>
      Promise.resolve({ karten: ['9'], befunde: [], fehler: null }),
    )
    const geaendert = vi.fn()

    belegeSeite({ AblageEinfuegen: einfuegen })
    zeige(geaendert)
    await waehleKartenhandlung(eigene.einfuegen)

    await waitFor(() => {
      expect(einfuegen).toHaveBeenCalledWith('1')
    })

    expect(geaendert).toHaveBeenCalled()
  })

  // **Der Fehler steht an der Handlungszeile** (§6) und nicht als anonymer
  // Hinweis — und der Bestand wird **nicht** nachgeladen (§7, Regel 3).
  it('meldet einen Fehler und lädt nicht nach', async () => {
    const geaendert = vi.fn()

    belegeSeite({
      AblageEinfuegen: () =>
        Promise.resolve({
          karten: [],
          befunde: [],
          fehler: {
            code: 'validierung',
            meldung: 'Das gehört zu einem anderen Spiel.',
            schluessel: '',
            feld: '',
          },
        }),
    })
    zeige(geaendert)
    await waehleKartenhandlung(eigene.einfuegen)

    expect(
      await screen.findByText('Das gehört zu einem anderen Spiel.'),
    ).toBeInTheDocument()

    expect(geaendert).not.toHaveBeenCalled()
  })
})

describe('Die Kürzel der Zwischenablage', () => {
  it.each([
    ['c', 'AblageKopieren'],
    ['x', 'AblageAusschneiden'],
    ['v', 'AblageEinfuegen'],
  ] as const)('⌘%s ruft %s', async (taste, methode) => {
    const gerufen = vi.fn(() =>
      Promise.resolve({ karten: [], befunde: [], fehler: null }),
    )

    belegeSeite({ [methode]: gerufen })
    zeige()

    const zeile = await screen.findByRole('button', { name: /Feuerball/ })

    fireEvent.keyDown(zeile, { key: taste, metaKey: true })

    await waitFor(() => {
      expect(gerufen).toHaveBeenCalled()
    })
  })

  // **Auf dem Ankreuzfeld gehört ⌘C den Karten.**
  //
  // Es steht in jeder Zeile, und wer eine Karte anhakt, hat den Fokus darauf
  // — der Tastendruck danach ist der wahrscheinlichste überhaupt.
  it('kopiert, wenn der Fokus auf einem Ankreuzfeld steht', async () => {
    const kopieren = vi.fn(() => Promise.resolve(ohneFehler))

    belegeSeite({ AblageKopieren: kopieren })
    zeige()
    await waehleMehrere()

    const [haken] = await screen.findAllByRole('checkbox', {
      name: new RegExp(texte.seiten.karten.massen.auswaehlen),
    })

    if (haken === undefined) {
      throw new Error('kein Ankreuzfeld in der Liste')
    }

    await userEvent.click(haken)
    fireEvent.keyDown(haken, { key: 'c', metaKey: true })

    await waitFor(() => {
      expect(kopieren).toHaveBeenCalledWith('1', ['1'])
    })
  })

  // **Im Textfeld gehört ⌘C dem Textfeld** — die Notizen stehen unter der
  // Liste und liegen im selben Ereignisweg.
  it('lässt ⌘C im Notizfeld in Ruhe', async () => {
    const kopieren = vi.fn(() => Promise.resolve(ohneFehler))

    belegeSeite({ AblageKopieren: kopieren })
    zeige()

    const notizen = await screen.findByLabelText(
      texte.seiten.karten.notizen.ueberschrift,
    )

    fireEvent.keyDown(notizen, { key: 'c', metaKey: true })

    // **Und ein Beleg, dass jemand zuhört** (§3): derselbe Weg mit dem
    // Listenknopf ruft — ohne ihn hieße „nicht gerufen" nur, dass der Horcher
    // nicht hing.
    fireEvent.keyDown(
      await screen.findByRole('button', { name: /Feuerball/ }),
      { key: 'c', metaKey: true },
    )

    await waitFor(() => {
      expect(kopieren).toHaveBeenCalledTimes(1)
    })
  })
})

// **Ein halb geglücktes Einfügen** — Befund 4 des Zweitblicks über Phase 3.7.
describe('Einfügen, das auf halbem Weg scheitert', () => {
  // # Warum das kein Randfall ist
  //
  // Eingefügt wird Karte für Karte: erst entsteht sie, dann kommen ihre
  // Werte. Scheitert die dritte, sind die ersten beiden **da** — und die
  // Go-Seite nennt sie in derselben Antwort, in der der Fehler steht.
  //
  // Der erste Anlauf ließ `auspacken` bei jedem Fehler abbrechen und lud
  // nicht nach: die Liste zeigte die neuen Karten erst, wenn jemand die
  // Lasche wechselte. Das ist §7, Regel 3 — geschrieben ist geschrieben.
  //
  // **Die Meldung bleibt trotzdem stehen.** Ein Teilerfolg ist kein Erfolg.
  it('lädt nach und meldet trotzdem', async () => {
    const geaendert = vi.fn()

    belegeSeite({
      AblageEinfuegen: () =>
        Promise.resolve({
          karten: ['9', '10'],
          befunde: [],
          fehler: {
            code: 'konflikt',
            schluessel: 'fehler.konflikt',
            meldung: 'die dritte Karte ging nicht',
            feld: '',
          },
        }),
    })

    zeige(geaendert)

    await waehleKartenhandlung(eigene.einfuegen)

    // **Über den Text und nicht über die Rolle** — die Seite trägt mehrere
    // Meldungsflächen, und die Zusage ist, dass **diese** Meldung dasteht.
    expect(
      await screen.findByText(/die dritte Karte ging nicht/),
    ).toBeInTheDocument()

    await waitFor(() => {
      expect(geaendert).toHaveBeenCalled()
    })
  })
})

// **Ein gescheiterter Schnitt lädt nicht nach** — die Kehrseite von Befund 4
// des Zweitblicks über Phase 3.7.
describe('Ausschneiden, das scheitert', () => {
  // # Warum das die Zusage ist und nicht ihr Gegenteil
  //
  // Seit dem 2026-08-27 löscht die Go-Seite alle Karten in **einer**
  // Transaktion: scheitert sie, ist **keine** fort. Nachzuladen gäbe es dann
  // nichts — und ein Ladefehler überschriebe die Meldung, die sagt, was zu
  // tun ist (§7, Regel 3).
  it('meldet und lässt den Bestand in Ruhe', async () => {
    const geaendert = vi.fn()

    belegeSeite({
      AblageAusschneiden: () =>
        Promise.resolve({
          fehler: {
            code: 'konflikt',
            schluessel: 'fehler.konflikt',
            meldung: 'der Schnitt ging nicht',
            feld: '',
          },
        }),
    })

    zeige(geaendert)

    await waehleKartenhandlung(eigene.ausschneiden.replace('{zahl}', '1'))

    expect(
      await screen.findByText(/der Schnitt ging nicht/),
    ).toBeInTheDocument()

    expect(geaendert).not.toHaveBeenCalled()
  })
})

// **Ein leeres Set hat keine Kartenspalte** — und damit weder die
// Handlungszeile noch ⌘V. Gefunden hat es die Hüllenprobe (`kopieren`): in
// ein neues Set ließ sich nichts einfügen. Der Leerzustand bietet es an.
describe('Ein leeres Set', () => {
  it('bietet Einfügen an und lädt danach nach', async () => {
    const einfuegen = vi.fn(() =>
      Promise.resolve({ karten: ['9'], befunde: [], fehler: null }),
    )
    const geaendert = vi.fn()

    belegeSeite({
      ...ausKarten(() => Promise.resolve({ karten: [] })),
      Karten: () => Promise.resolve({ karten: [] }),
      AblageEinfuegen: einfuegen,
    })
    zeige(geaendert)

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.einfuegen }),
    )

    await waitFor(() => {
      expect(einfuegen).toHaveBeenCalledWith('1')
    })

    expect(geaendert).toHaveBeenCalled()
  })
})
