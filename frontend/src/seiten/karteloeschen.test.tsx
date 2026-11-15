// **Eine Karte lässt sich löschen — nach einer Rückfrage**
// (#621).
//
// #562 hat das
// Anlegen gebaut und das Löschen vertagt: *„eine gelöschte Karte ohne ⌘Z ist
// Datenverlust mit einem Klick."* Seit der Verlauf eine eigene Art für ganze
// Karten trägt, ist der Weg zurück da — und die Rückfrage sagt es dazu.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { oeffneKartenhandlungen } from '../test/kartenlasche'
import { belegeApi, raeumeApi } from '../test/api'
import { Karten } from './Karten'
import { ausKarten } from '../test/kartenproben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function karte(id: string, name: string): KarteDTO {
  return { id, nummer: Number(id) - 1, werte: [{ feld: 'name', wert: name }] }
}

function zeige(
  teile: Partial<Api>,
  bestandGeaendert = () => undefined,
  bestand: () => KarteDTO[] = () => [karte('1', 'Erste'), karte('2', 'Zweite')],
) {
  belegeApi({
    ...ausKarten(() => Promise.resolve({ karten: bestand() })),
    Karten: () => Promise.resolve({ karten: bestand() }),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Probevorschau: () => Promise.resolve({ vorschau: null }),
    // **Ein eigener `Karten`-Beleg wird mitprojiziert** (#891). Die Seite holt
    // `Kartenzeilen` und `Karte`; ohne diese Zeile bliebe die Vorgabe darüber
    // stehen, und der Test prüfte Karten, die er nicht beschrieben hat.
    ...(teile.Karten ? ausKarten(teile.Karten, '7') : {}),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={vi.fn()}
        setID="7"
        karteID="1"
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={bestandGeaendert}
      />
    </Sprachraum>,
  )
}

describe('Karte löschen', () => {
  it('fragt nach, löscht und räumt die Wahl', async () => {
    const KarteLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
    const bestandGeaendert = vi.fn()
    let bestand = [karte('1', 'Erste'), karte('2', 'Zweite')]

    zeige({ KarteLoeschen }, bestandGeaendert, () => bestand)

    await oeffneKartenhandlungen()

    const loeschen = await screen.findByRole('menuitem', {
      name: texte.seiten.karten.loeschen,
    })

    // **Der erste Klick löscht nicht** — er fragt.
    await userEvent.click(loeschen)

    expect(KarteLoeschen).not.toHaveBeenCalled()
    expect(
      screen.getByText(texte.seiten.karten.loeschenFrage),
    ).toBeInTheDocument()

    bestand = [karte('2', 'Zweite')]

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    expect(KarteLoeschen).toHaveBeenCalledWith('7', '1', false)

    // **Der offene Kontext ist geräumt** (§7, Regel 1): kein Editor mehr auf
    // eine Karte, die es nicht mehr gibt, und keine Frage, die zu ihr gehörte
    // — und der Bestand daneben weiß davon.
    await waitFor(() => {
      expect(bestandGeaendert).toHaveBeenCalled()
    })

    expect(screen.queryByDisplayValue('Erste')).not.toBeInTheDocument()
    expect(
      screen.queryByText(texte.seiten.karten.loeschenFrage),
    ).not.toBeInTheDocument()
  })

  // **Nach dem Löschen ist eine Karte aufgeschlagen, nicht keine** — so hält es
  // MSE2: es merkt sich die *Stelle* der gelöschten Karte und wählt danach, was
  // dort nun steht (`src/gui/control/card_list.cpp:78–89`,
  // `src/gui/control/item_list.cpp:87–100`, Fassung f9d9356 vom 2021-07-06).
  it('schlägt die nachrückende Karte auf', async () => {
    const KarteLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
    let bestand = [karte('1', 'Erste'), karte('2', 'Zweite')]

    zeige({ KarteLoeschen }, undefined, () => bestand)

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )

    bestand = [karte('2', 'Zweite')]

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    // Die zweite Karte rückt auf die Stelle der ersten — und wird aufgeschlagen.
    expect(await screen.findByDisplayValue('Zweite')).toBeInTheDocument()
  })

  // **Und war es die letzte, gilt die letzte, die bleibt.** `selectItemPos`
  // greift bei `pos >= size` auf `sorted_list.back()` zurück.
  it('schlägt die vorige Karte auf, wenn die letzte gelöscht wird', async () => {
    const KarteLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
    let bestand = [karte('1', 'Erste'), karte('2', 'Zweite')]

    zeige({ KarteLoeschen }, undefined, () => bestand)

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: `Zweite — ${texte.seiten.karten.bearbeiten}`,
      }),
    )

    // **Das Menü ist mit dem Klick auf die Zeile zugegangen** — ein Klick
    // daneben schließt es. Für die zweite Karte geht es neu auf.
    await oeffneKartenhandlungen()
    await userEvent.click(
      screen.getByRole('menuitem', { name: texte.seiten.karten.loeschen }),
    )

    bestand = [karte('1', 'Erste')]

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    expect(await screen.findByDisplayValue('Erste')).toBeInTheDocument()
  })

  // **Und war es die einzige, ist keine aufgeschlagen** — die Liste ist leer,
  // und der leere Zustand des Sets sagt, was als Nächstes zu tun ist (§6).
  it('schlägt nichts auf, wenn die letzte Karte gelöscht wird', async () => {
    const KarteLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))
    let bestand = [karte('1', 'Erste')]

    zeige({ KarteLoeschen }, undefined, () => bestand)

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )

    bestand = []

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    expect(
      await screen.findByText(texte.seiten.karten.leerTitel),
    ).toBeInTheDocument()
  })

  // **Regel 2 aus §7: eine Antwort gehört zu dem Stand, aus dem sie stammt.**
  //
  // Wer während des Löschens eine andere Karte aufschlägt, hat einen neuen
  // offenen Kontext. Die Antwort auf das Löschen darf ihn nicht anfassen —
  // sonst springt die Wahl auf die Nachfolgerin einer Karte, die gar nicht die
  // aufgeschlagene war.
  it('lässt eine inzwischen aufgeschlagene Karte in Ruhe', async () => {
    let loesen = () => {
      // bis der Aufruf sie ersetzt: die Zusage steht noch nicht
    }

    const KarteLoeschen = vi.fn(
      () =>
        new Promise<{ fehler: null }>((erfuellen) => {
          loesen = () => {
            erfuellen({ fehler: null })
          }
        }),
    )

    // **Drei Karten, und die aufgeschlagene ist nicht die Nachfolgerin.** Mit
    // zweien fiele beides zusammen, und der Test bliebe grün, ohne den Fall zu
    // enthalten (§10).
    const bestand = [
      karte('1', 'Erste'),
      karte('2', 'Zweite'),
      karte('3', 'Dritte'),
    ]
    const bestandGeaendert = vi.fn()

    zeige({ KarteLoeschen }, bestandGeaendert, () => bestand)

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    // Mitten im Lauf die dritte aufschlagen — nicht die nachrückende zweite.
    await userEvent.click(
      screen.getByRole('button', {
        name: `Dritte — ${texte.seiten.karten.bearbeiten}`,
      }),
    )

    expect(await screen.findByDisplayValue('Dritte')).toBeInTheDocument()

    loesen()

    // **Auf den Abschluss warten, nicht auf den Zustand**: eine Zusicherung,
    // die schon vor der Antwort zutrifft, prüft nichts.
    await waitFor(() => {
      expect(bestandGeaendert).toHaveBeenCalled()
    })

    // **Und dann noch das Nachladen** (§3, #1321): der Ruf ist geschehen, die
    // neue Liste ist deswegen noch nicht da.
    expect(await screen.findByDisplayValue('Dritte')).toBeInTheDocument()
  })

  it('bricht ab, ohne zu löschen', async () => {
    const KarteLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ KarteLoeschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenNein }),
    )

    expect(KarteLoeschen).not.toHaveBeenCalled()

    // **Und die Frage ist weg, der Fokus wieder auf dem Menü** — der
    // Auslöser, aus dem sie kam.
    expect(
      screen.queryByRole('group', {
        name: texte.seiten.karten.loeschenFrage,
      }),
    ).toBeNull()
    expect(
      screen.getByRole('button', { name: texte.seiten.karten.dieseKarte }),
    ).toHaveFocus()
  })

  it('meldet einen Befund an der Stelle, an der geklickt wurde', async () => {
    const KarteLoeschen = vi.fn(() =>
      Promise.resolve({
        fehler: {
          code: 'nichtGefunden' as const,
          schluessel: 'nichtGefunden',
          meldung: 'Die Karte gibt es nicht.',
        },
      }),
    )

    zeige({ KarteLoeschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    )

    // **Die Meldung steht am verursachenden Ort** (§6), und die Frage bleibt
    // stehen — wer den Befund liest, will nicht neu anfangen müssen.
    await waitFor(() => {
      expect(screen.getByText(texte.fehler.nichtGefunden)).toBeInTheDocument()
    })

    expect(
      screen.getByRole('button', { name: texte.seiten.karten.loeschenJa }),
    ).toBeInTheDocument()
  })

  it('räumt die Frage, wenn die Karte wechselt', async () => {
    const KarteLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({ KarteLoeschen })

    await oeffneKartenhandlungen()

    await userEvent.click(
      await screen.findByRole('menuitem', {
        name: texte.seiten.karten.loeschen,
      }),
    )

    expect(
      screen.getByText(texte.seiten.karten.loeschenFrage),
    ).toBeInTheDocument()

    // **Eine andere Karte aufschlagen** — die Frage gehörte der ersten
    // (§7, Regel 1). Bliebe sie stehen, löschte der nächste Klick die falsche.
    await userEvent.click(
      screen.getByRole('button', {
        name: `Zweite — ${texte.seiten.karten.bearbeiten}`,
      }),
    )

    await waitFor(() => {
      expect(
        screen.queryByText(texte.seiten.karten.loeschenFrage),
      ).not.toBeInTheDocument()
    })
  })
})
