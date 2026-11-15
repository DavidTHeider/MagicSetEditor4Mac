import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { beschriftung } from '../navigation'
import { leisteneintrag } from '../test/leistenweg'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Palette } from './Palette'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'

function zeige(eigenschaften: Partial<Parameters<typeof Rahmen>[0]> = {}) {
  belegeApi({})
  const wechseln = vi.fn()
  const suchen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Rahmen
        aktuell="start"
        wechseln={wechseln}
        suchen={suchen}
        {...eigenschaften}
      >
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )

  return { wechseln, suchen }
}

afterEach(() => {
  raeumeApi()
})

describe('Rahmen', () => {
  // **Die Navigation kommt aus `navigation.ts`.** Eine Kopie daneben wäre grün
  // und zu kurz, und niemand sähe es.
  // **Die Kontextleiste** — Werkstück U4 von Phase 5.1 (#1415): der Rahmen
  // zeichnet sie selbst, unter dem Anwendungsnamen, denn sie gilt auf jeder
  // Seite.
  it('zeigt die Kontextleiste unter dem Anwendungsnamen', () => {
    zeige()

    const leiste = screen.getByRole('region', {
      name: kataloge.de.bausteine.kontext.bezeichnung,
    })
    const name = screen.getByText(kataloge.de.allgemein.anwendungsname)

    expect(
      name.compareDocumentPosition(leiste) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })

  // **Jeden Eintrag der Leiste, und das sind seit UX-02 Bereiche**
  // (#1674).
  // Hier stand `anwendungsseiten` — vierzehn Seiten, vierzehn Knöpfe. Die
  // Leiste führt jetzt sechs Bereiche und drei Dienste; wohin es **in** einem
  // Bereich weitergeht, sagt der Kopf der Seite.
  it('wechselt beim Klick auf einen Eintrag', async () => {
    const { wechseln } = zeige()

    await userEvent.click(await leisteneintrag('pakete'))

    expect(wechseln).toHaveBeenCalledWith('pakete')
  })

  it('öffnet die Palette mit ⌘K und schließt sie damit wieder', () => {
    zeige()

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })
    expect(screen.getByRole('dialog')).toBeInTheDocument()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  // **Wo man steht, steht am Inhalt** — N4 des Nachlaufs über 6.13
  // (#1906): die Hüllenprobe wartet nach einem Wechsel auf diese Marke statt
  // auf eine Frist, und die Zeile der Palette nennt ihr Ziel genauso.
  it('trägt die aktuelle Seite am Inhalt und das Ziel an der Palettenzeile', async () => {
    zeige({ aktuell: 'karten' })

    expect(screen.getByRole('main')).toHaveAttribute('data-seite', 'karten')

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.allgemein.palette.oeffnen,
      }),
    )

    const zeile = screen.getByRole('button', {
      name: beschriftung(kataloge.de, 'sets'),
    })
    expect(zeile).toHaveAttribute('data-seite', 'sets')
  })

  it('öffnet die Palette auch über die Schaltfläche', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.allgemein.palette.oeffnen,
      }),
    )

    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('ruft die Suche mit ⌘F', () => {
    const { suchen } = zeige()

    fireEvent.keyDown(window, { key: 'f', metaKey: true })

    expect(suchen).toHaveBeenCalledOnce()
  })

  // Ohne eigene Suchfunktion führt ⌘F zur Suchseite.
  //
  // **Eine Taste, die je nach Seite etwas anderes tut, ist keine Gewohnheit
  // mehr** — und „tut nichts" ist auch etwas anderes.
  //
  // **Ohne den Spreizoperator**, weil `exactOptionalPropertyTypes` ein
  // ausdrückliches `undefined` von einem fehlenden Feld unterscheidet — und
  // geprüft werden soll das fehlende.
  it('führt ohne eigene Suchfunktion zur Suchseite', () => {
    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={wechseln}>
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    fireEvent.keyDown(window, { key: 'f', metaKey: true })

    expect(wechseln).toHaveBeenCalledWith('suche')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('führt mit ⌘, zu den Einstellungen', () => {
    const { wechseln } = zeige()

    fireEvent.keyDown(window, { key: ',', metaKey: true })

    expect(wechseln).toHaveBeenCalledWith('einstellungen')
  })

  // **⌘Q und ⌘W bleiben dem Betriebssystem.** Wer jede ⌘-Kombination abfängt,
  // nimmt sie ihm weg — und §6 sagt sie ausdrücklich zu.
  it('lässt unbelegte Befehlstasten durch', () => {
    const { wechseln, suchen } = zeige()

    fireEvent.keyDown(window, { key: 'q', metaKey: true })
    fireEvent.keyDown(window, { key: 'w', metaKey: true })

    expect(wechseln).not.toHaveBeenCalled()
    expect(suchen).not.toHaveBeenCalled()
  })

  it('lässt einen Tastendruck ohne Befehlstaste durch', () => {
    const { wechseln } = zeige()

    fireEvent.keyDown(window, { key: ',' })

    expect(wechseln).not.toHaveBeenCalled()
  })

  it('bietet einen Sprung zum Inhalt', () => {
    zeige()

    const sprung = screen.getByRole('link', {
      name: kataloge.de.allgemein.navigation.sprungZumInhalt,
    })

    expect(sprung).toHaveAttribute('href', '#inhalt')
    expect(screen.getByRole('main')).toHaveAttribute('id', 'inhalt')
  })

  // **Der Inhalt nimmt den Fokus, ohne ein Halt zu sein** — UX-10.AK-02
  // (#1682): WebKit bewegt ihn beim Sprung zu einem Ziel, das keinen nehmen
  // kann, nicht, und nach einem Seitenwechsel landet er dort
  // (`useFokusNachSeitenwechsel`).
  it('lässt den Inhalt den Fokus nehmen, ohne ihn in die Tab-Folge zu stellen', () => {
    zeige()

    const inhalt = screen.getByRole('main')
    inhalt.focus()

    expect(inhalt).toHaveFocus()
    expect(inhalt).toHaveAttribute('tabindex', '-1')
  })

  it('nimmt den Tastaturhorcher beim Abbau wieder weg', () => {
    const abmelden = vi.spyOn(window, 'removeEventListener')

    const { unmount } = render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={vi.fn()}>
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    unmount()

    expect(abmelden).toHaveBeenCalledWith('keydown', expect.any(Function))
  })
})

describe('Befehlspalette', () => {
  it('wählt eine Seite aus und schließt sich', async () => {
    const { wechseln } = zeige()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })

    // **Im Dialog gesucht**, nicht auf der Seite: die Navigation trägt dieselben
    // Namen, und ein Test, der den ersten Treffer nimmt, klickt in die
    // Navigation und prüft die Palette nicht.
    await userEvent.click(
      within(screen.getByRole('dialog')).getByRole('button', {
        name: beschriftung(kataloge.de, 'sets'),
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('sets')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  // **Escape geht an ein Element und nicht an das Fenster** — UX-07
  // (#1679).
  //
  // Hier stand `fireEvent.keyDown(window, …)`, und das war eine Abkürzung:
  // ein echter Tastendruck hat immer ein **Element** als Ziel, nie das
  // Fenster. Solange der Dialog von Hand gebaut war, fiel es nicht auf —
  // sein Horcher las nur `key`. Mantines Horcher fragt das Ziel nach einem
  // Merkmal, und `window.getAttribute` gibt es nicht.
  //
  // **Der Lauf war damit großzügiger als der Betrieb** (§3), und die
  // Korrektur geht in die richtige Richtung: er stellt jetzt nach, was
  // wirklich geschieht. `⌘K` bleibt am Fenster, denn diesen Horcher hängt
  // `Rahmen` selbst dorthin.
  // **Einen Umlauf später** (G12, #1672): der Dialog wartet, ob jemand in
  // ihm die Taste für sich genommen hat — ein Hinweis, eine Rückfrage.
  it('schließt sich mit Escape', async () => {
    zeige()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })
    fireEvent.keyDown(document.body, { key: 'Escape' })

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  it('filtert nach der Eingabe', async () => {
    zeige()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })

    await userEvent.type(
      screen.getByRole('textbox', {
        name: kataloge.de.allgemein.palette.eingabe,
      }),
      'paket',
    )

    const dialog = within(screen.getByRole('dialog'))

    expect(
      dialog.getByRole('button', { name: beschriftung(kataloge.de, 'pakete') }),
    ).toBeInTheDocument()
    expect(
      dialog.queryByRole('button', {
        name: beschriftung(kataloge.de, 'karten'),
      }),
    ).not.toBeInTheDocument()
  })

  it('sagt es, wenn nichts passt', async () => {
    zeige()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })

    await userEvent.type(
      screen.getByRole('textbox', {
        name: kataloge.de.allgemein.palette.eingabe,
      }),
      'gibtsnicht',
    )

    expect(
      screen.getByText(kataloge.de.allgemein.palette.keineTreffer),
    ).toBeInTheDocument()
  })

  // **Die alte Eingabe verschwindet beim Öffnen.** Eine Palette, die sich die
  // letzte Suche merkt, zeigt beim nächsten Öffnen drei Einträge und wirkt
  // kaputt.
  it('beginnt bei jedem Öffnen leer', async () => {
    zeige()

    fireEvent.keyDown(window, { key: 'k', metaKey: true })
    await userEvent.type(
      screen.getByRole('textbox', {
        name: kataloge.de.allgemein.palette.eingabe,
      }),
      'paket',
    )

    fireEvent.keyDown(window, { key: 'k', metaKey: true })
    fireEvent.keyDown(window, { key: 'k', metaKey: true })

    expect(
      screen.getByRole('textbox', {
        name: kataloge.de.allgemein.palette.eingabe,
      }),
    ).toHaveValue('')
  })

  it('lässt eine andere Taste durch', () => {
    const schliessen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Palette
          schliessen={schliessen}
          waehlen={vi.fn()}
          waehleBefehl={vi.fn()}
        />
      </Sprachraum>,
    )

    fireEvent.keyDown(window, { key: 'a' })

    expect(schliessen).not.toHaveBeenCalled()
  })

  // Escape greift auch, wenn der Fokus die Palette verlassen hat.
  it('schließt sich mit Escape, gleich wo der Fokus steht', async () => {
    const schliessen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Palette
          schliessen={schliessen}
          waehlen={vi.fn()}
          waehleBefehl={vi.fn()}
        />
      </Sprachraum>,
    )

    fireEvent.keyDown(document.body, { key: 'Escape' })

    await waitFor(() => {
      expect(schliessen).toHaveBeenCalledOnce()
    })
  })
})

// **Die Regionen eines Arbeitsplatzes gehören dem Arbeitsplatz** — Werkstück
// U2 von Phase 5.1 (#1413): ⌥⌘[ und ⌥⌘] führen auf keine Seite. Der Rahmen
// fängt die Taste ab und tut sonst nichts; wer horcht, ist der Arbeitsplatz.
//
// **Der Beleg, dass jemand zuhört, steht davor** (§3): ⌘, führt auf die
// Einstellungen — erst dann sagt die Verneinung etwas.
it('führt bei ⌥⌘[ und ⌥⌘] auf keine Seite', () => {
  const { wechseln } = zeige()

  fireEvent.keyDown(window, { key: ',', metaKey: true })
  expect(wechseln).toHaveBeenCalledWith('einstellungen')

  fireEvent.keyDown(window, {
    key: '„',
    code: 'BracketLeft',
    metaKey: true,
    altKey: true,
  })
  fireEvent.keyDown(window, {
    key: '‘',
    code: 'BracketRight',
    metaKey: true,
    altKey: true,
  })

  expect(wechseln).toHaveBeenCalledTimes(1)
})
