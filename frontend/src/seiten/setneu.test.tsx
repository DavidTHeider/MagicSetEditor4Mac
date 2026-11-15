// **Ein neues, leeres Set** —
// #891.
//
// # Warum über `Sets` und nicht über `SetNeu` allein
//
// Die Form steht auf der Setliste, und das Anlegen führt dorthin zurück: das
// neue Set wird **aufgeschlagen**. Ein Lauf gegen die Form allein sagte über
// diese Naht nichts.
import { render, screen, waitFor } from '@testing-library/react'
import { Activity } from 'react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sets } from './Sets'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { waehleZeile } from '../test/suchwahl'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.sets

// stilwahl ist die Auswahl des Stils — seit UXR-04 (#1912) kein Optionsfeld je
// Stil mehr, mit dem Full-Magic-Pack waren es 355.
const stilwahl = () => screen.findByRole('combobox', { name: worte.neuStil })

// pakete ist ein Katalog mit einem Spiel und zwei Stilen — einer davon gehört
// einem anderen Spiel.
function pakete() {
  return {
    pakete: [
      {
        art: 'mse-game',
        name: 'probe',
        anzeigename: 'Probe',
        version: '',
        pfad: '',
        herkunft: '',
        pruefsumme: '',
        braucht: [],
        gruppe: '',
      },
      {
        art: 'mse-style',
        name: 'probe-schlicht',
        anzeigename: 'Probe-schlicht',
        version: '',
        pfad: '',
        herkunft: '',
        pruefsumme: '',
        braucht: [],
        gruppe: '',
      },
      {
        art: 'mse-style',
        name: 'magic-m15',
        anzeigename: 'Magic-m15',
        version: '',
        pfad: '',
        herkunft: '',
        pruefsumme: '',
        braucht: [],
        gruppe: '',
      },
    ],
    fehler: null,
  }
}

function zeige(waehlen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={waehlen}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return waehlen
}

describe('Ein neues Set', () => {
  it('legt es an und schlägt es auf', async () => {
    const SetAnlegen = vi.fn().mockResolvedValue({ id: '9', fehler: null })

    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue(pakete()),
      SetAnlegen,
    })

    const waehlen = zeige()

    // **Der Stil wird gewählt** — die Liste zeigt nur die des Spiels, sobald
    // das Feld offen ist (UXR-F1, #2007). **Ein Spiel ist kein Schritt**
    // (UX-05.AK-04): es gibt nur eines mit Stil, und es beginnt beim Aussehen.
    await userEvent.click(await stilwahl())
    expect(screen.queryByRole('option', { name: /Magic-m15/ })).toBeNull()
    await userEvent.click(screen.getByRole('option', { name: /Probe-schl/ }))
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))

    await userEvent.type(
      await screen.findByLabelText(worte.neuName),
      'Mein Set',
    )
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )

    await waitFor(() => {
      expect(SetAnlegen).toHaveBeenCalledWith(
        'Mein Set',
        'probe',
        'probe-schlicht',
      )
    })

    // **Und es wird aufgeschlagen** — wer ein Set anlegt, will hinein.
    expect(waehlen).toHaveBeenCalledWith('9')
  })

  it('meldet einen Fehler und behält, was dasteht', async () => {
    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue(pakete()),
      SetAnlegen: vi.fn().mockRejectedValue(new Error('Das Spiel fehlt.')),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: worte.neuUeberspringen }),
    )
    const feld = await screen.findByLabelText(worte.neuName)

    await userEvent.type(feld, 'Mein Set')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das Spiel fehlt.',
    )

    // **Geschrieben ist geschrieben** (§7, Regel 3): der Name bleibt stehen,
    // damit niemand ihn neu tippt.
    expect(feld).toHaveValue('Mein Set')
  })

  // **Ein Wechsel des Spiels räumt den Stil** (§7, Regel 1).
  //
  // Ein Stil gehört zu einem Spiel; wer ihn behielte, schickte einen, den es
  // dort nicht gibt.
  // **Der Stil folgt dem Spiel** — und wird dabei **vorgewählt**, nicht
  // geleert.
  //
  //	stylesheet_list->showData<StyleSheet>(game->name() + _("-*"));
  //	stylesheet_list->select(settings.gameSettingsFor(*game).default_stylesheet);
  //
  // `src/gui/new_window.cpp`, Zeilen 65–66, Fassung f9d9356 vom 2021-07-06.
  //
  // **Bis zum 2026-08-25 stand hier `''`**, und das Set entstand ohne Stil —
  // MSEs `done()` legt in dem Fall gar keines an (Zeile 91). Angestrichen vom
  // Zweitblick über Phase Y als Befund 5.
  //
  // Was gleich bleibt: der Stil des **vorigen** Spiels darf nicht stehen
  // bleiben. `probe-schlicht` gehört zu `probe`, nicht zu `magic`.
  it('wählt beim Spielwechsel den Stil des neuen Spiels', async () => {
    const SetAnlegen = vi.fn().mockResolvedValue({ id: '9', fehler: null })

    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue({
        pakete: [
          ...pakete().pakete,
          {
            art: 'mse-game',
            name: 'magic',
            anzeigename: 'Magic',
            version: '',
            pfad: '',
            herkunft: '',
            pruefsumme: '',
            braucht: [],
            gruppe: '',
          },
        ],
        fehler: null,
      }),
      SetAnlegen,
    })

    zeige()

    // **Erst das eine Spiel samt Stil, dann zurück und das andere.**
    await userEvent.click(await screen.findByRole('radio', { name: /Probe/ }))
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))
    await waehleZeile(await stilwahl(), 'Probe-schlicht probe-schlicht')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuSchrittZurueck }),
    )

    await userEvent.click(await screen.findByRole('radio', { name: /Magic/ }))
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))

    // **Der Stil des neuen Spiels steht vorgewählt**, nicht der alte.
    expect(await stilwahl()).toHaveValue('Magic-m15')
    await userEvent.click(screen.getByRole('button', { name: worte.neuWeiter }))

    await userEvent.type(await screen.findByLabelText(worte.neuName), 'Zwei')
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    )

    await waitFor(() => {
      expect(SetAnlegen).toHaveBeenCalledWith('Zwei', 'magic', 'magic-m15')
    })
  })

  // **Ohne Stilpaket ebenso** — angestrichen vom Nachlauf des Zweitblicks
  // über Phase Y.
  //
  // Das Fenster verlangte den Stil nur, **solange es einen gab**: hat ein
  // installiertes Spiel keinen passenden Stil, war `gewaehlterStil` leer und
  // der Knopf trotzdem freigegeben. `Set::validate` weist ein Set ohne Stil ab
  // (`src/data/set.cpp:143`), und der Dienst tut es jetzt auch — die Seite
  // soll es vorher sagen, statt eine Meldung aus dem Fachkern zu zeigen.
  it('sagt es, wenn das Spiel kein Stilpaket hat', async () => {
    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue({
        pakete: [
          {
            art: 'mse-game',
            name: 'einsam',
            anzeigename: 'Einsam',
            version: '',
            pfad: '',
            herkunft: '',
            pruefsumme: '',
            braucht: [],
            gruppe: '',
          },
        ],
        fehler: null,
      }),
    })

    zeige()

    expect(
      await screen.findByText(/Zu einsam ist kein Stilpaket installiert/),
    ).toBeInTheDocument()

    // **Und der Knopf ist angehalten** — auch mit Namen.
    await userEvent.click(
      screen.getByRole('button', { name: worte.neuUeberspringen }),
    )
    await userEvent.type(
      await screen.findByLabelText(worte.neuName),
      'Geht nicht',
    )

    expect(
      screen.getByRole('button', { name: worte.neuAnlegen }),
    ).toBeDisabled()
  })

  // **Ohne Spielpaket steht kein Satz mit einem Loch darin** — Werkstück C
  // von Phase 4.10 (#1325).
  //
  // Der Abzug der Setseite vom 2026-09-03 zeigte „**Zu ist kein Stilpaket
  // installiert.**" — der Platzhalter `{spiel}` war leer, weil noch kein
  // Spiel dastand. Ein Satz über die Stile *eines* Spiels ist ohne dieses
  // Spiel keine Auskunft; was dann gilt, sagt `neuOhneSpiele`.
  it('nennt ohne Spielpaket das fehlende Spiel und nicht die fehlenden Stile', async () => {
    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue({ pakete: [], fehler: null }),
    })

    zeige()

    expect(await screen.findByText(worte.neuOhneSpiele)).toBeInTheDocument()

    // **Und kein zweiter Satz daneben**, der von einem Spiel spricht, das es
    // nicht gibt.
    expect(screen.queryByText(/ist kein Stilpaket installiert/)).toBeNull()
  })

  // **Und während die Pakete noch kommen, steht er auch nicht da.**
  //
  // Das ist der Zustand, den der Abzug erwischt hat: 331 Pakete brauchen
  // einen Augenblick, und solange `pakete.daten` `undefined` ist, zeigt die
  // Seite ihr Formular mit leerer Spielliste. `gewaehltesSpiel` ist dann `''`
  // — und der Platzhalter im Satz bleibt ein Loch.
  //
  // **Ein Ladezustand ist kein Befund über die Pakete**, und ein Satz, der so
  // tut, ist schlimmer als keiner.
  it('zeigt den Stilsatz nicht, solange die Pakete noch kommen', async () => {
    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockReturnValue(new Promise(() => undefined)),
    })

    zeige()

    // **Gewartet wird auf die Überschrift** — sie steht, sobald die Seite
    // gezeichnet ist, und danach ist der Satz entweder da oder nicht.
    expect(await screen.findByText(worte.neuTitel)).toBeInTheDocument()

    expect(screen.queryByText(/ist kein Stilpaket installiert/)).toBeNull()
  })

  // **Ohne Spielpaket gibt es nichts anzulegen, und die Seite sagt es.**
  it('sagt es, wenn kein Spiel installiert ist', async () => {
    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue({ pakete: [], fehler: null }),
    })

    zeige()

    expect(await screen.findByText(worte.neuOhneSpiele)).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: worte.neuAnlegen }),
    ).not.toBeInTheDocument()
  })

  // **⌘N setzt den Fokus ins Namensfeld.**
  it('nimmt den Anstoß entgegen', async () => {
    const verbraucht = vi.fn()

    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue(pakete()),
    })

    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
          neuanstoss
          neuAngestossen={verbraucht}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByLabelText(worte.neuName)).toHaveFocus()
    })

    expect(verbraucht).toHaveBeenCalled()
  })

  // **Beim Wiedereinblenden greift es nicht nach dem Fokus** — angestrichen
  // vom Zweitblick über Phase 6.3 (G11): `<Activity>` lässt den Effekt beim
  // Einblenden neu laufen, und das Namensfeld zog den Fokus aus der Leiste.
  it('greift beim Wiedereinblenden nicht nach dem Fokus', async () => {
    belegeApi({
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      Pakete: vi.fn().mockResolvedValue(pakete()),
    })
    const seite = (sichtbar: boolean) => (
      <Sprachraum sprache="de">
        <button type="button">Leiste</button>
        <Activity mode={sichtbar ? 'visible' : 'hidden'}>
          <Sets
            wechseln={vi.fn()}
            waehlen={vi.fn()}
            speichern={stillerSpeicherlauf()}
            neuanstoss
            neuAngestossen={vi.fn()}
          />
        </Activity>
      </Sprachraum>
    )

    const { rerender } = render(seite(true))
    await waitFor(() => {
      expect(screen.getByLabelText(worte.neuName)).toHaveFocus()
    })

    rerender(seite(false))
    const leiste = screen.getByRole('button', { name: 'Leiste' })
    leiste.focus()
    rerender(seite(true))

    expect(leiste).toHaveFocus()
  })
})
