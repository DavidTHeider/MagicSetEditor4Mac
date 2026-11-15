// **Die Lasche „Setangaben"** —
// #567.
//
// # Wofür
//
// Ein Set hat Angaben: Titel, Beschreibung, Künstler, Copyright, Setcode,
// Sprache, Symbol. Über `magic.mse-game` sind es **89 Felder**. Sie standen im
// Bestand und waren in der Oberfläche nirgends erreichbar.
//
// # Was MSE tut
//
//	void SetInfoEditor::onChangeSet() {
//	  setStyles(set->stylesheet, set->stylesheet->set_info_style);
//	  setData(set->data);
//	}
//
// `src/gui/control/native_look_editor.cpp`, Zeilen 187–190, Fassung f9d9356
// vom 2021-07-06 — eine eigene Lasche neben „Cards", mit denselben
// Steuerelementen wie der Karteneditor.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FeldbeschreibungDTO, SetangabenAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Setangaben } from './Setangaben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function form(
  name: string,
  teile: Partial<FeldbeschreibungDTO> = {},
): FeldbeschreibungDTO {
  return {
    name,
    aufschrift: name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    ...teile,
  }
}

function zeige(teile = {}) {
  const SetwertSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({
    Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
    Setangaben: () =>
      Promise.resolve({
        felder: [
          form('title', { aufschrift: 'Title' }),
          form('description', { mehrzeilig: true, aufschrift: 'Description' }),
          form('set code', { aufschrift: 'Set code' }),
        ],
        werte: [{ feld: 'title', wert: 'Aus der Saat' }],
        kopf: leererSetkopf,
        fehler: null,
      }),
    SetwertSetzen,
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Setangaben setID="1" verlaufsstand={0} />
    </Sprachraum>,
  )

  return SetwertSetzen
}

// leererSetkopf ist ein Setkopf ohne eine einzige Angabe (#1325).
//
// **Für die Läufe, die ihn nicht meinen**: die Formatangaben haben ihren
// eigenen Lauf in `setkopf.test.tsx`; hier stünde eine gefüllte Gestalt nur im
// Weg.
const leererSetkopf = {
  mseFassung: '',
  spielfassung: '',
  stilfassung: '',
  versionskontrolle: '',
  apprenticecode: '',
}

describe('Die Setangaben', () => {
  it('zeigt je Feld ein Eingabefeld mit seinem Wert', async () => {
    zeige()

    await waitFor(() => {
      expect(screen.getByLabelText('title')).toHaveValue('Aus der Saat')
    })

    // **Die Aufschrift steht daneben, der Name im `aria-label`** — dieselbe
    // Trennung wie im Karteneditor: der Name ist die Kennung.
    expect(screen.getByText('Title')).toBeInTheDocument()

    // **Und die Form folgt `type:`**: `description` ist mehrzeilig.
    expect(screen.getByLabelText('description').tagName).toBe('TEXTAREA')

    // **Ein Feld ohne Wert steht leer da**, nicht gar nicht: MSE zeigt jede
    // Angabe, die das Spiel erklärt.
    expect(screen.getByLabelText('set code')).toHaveValue('')
  })

  it('schreibt beim Verlassen des Feldes', async () => {
    const SetwertSetzen = zeige()

    const feld = await screen.findByLabelText('set code')
    await userEvent.type(feld, 'PRB')
    await userEvent.tab()

    await waitFor(() => {
      expect(SetwertSetzen).toHaveBeenCalledWith(
        '1',
        'set code',
        'PRB',
        expect.any(Array),
      )
    })
  })

  // **Ein unveränderter Wert wird nicht geschrieben.** Wer ein Feld anklickt
  // und wieder verlässt, legte sonst einen Schritt im Verlauf an, der nichts
  // ändert — und ⌘Z nähme ihn dann zurück, ohne dass etwas anders aussieht.
  it('schreibt nichts, wenn sich nichts geändert hat', async () => {
    const SetwertSetzen = zeige()

    await userEvent.click(await screen.findByLabelText('title'))
    await userEvent.tab()

    expect(SetwertSetzen).not.toHaveBeenCalled()
  })

  // **Ein Fehler erscheint am verursachenden Feld** (§6).
  it('meldet einen Fehler am Feld', async () => {
    zeige({
      SetwertSetzen: () => Promise.reject(new Error('Die Platte ist voll.')),
    })

    const feld = await screen.findByLabelText('set code')
    await userEvent.type(feld, 'PRB')
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Die Platte ist voll.',
      )
    })
  })

  // **Ein Ladefehler wird gezeigt und nicht verschluckt** (§8).
  it('meldet einen Ladefehler', async () => {
    belegeApi({
      Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
      Setangaben: () => Promise.reject(new Error('Kein Spielpaket.')),
    })

    render(
      <Sprachraum sprache="de">
        <Setangaben setID="1" verlaufsstand={0} />
      </Sprachraum>,
    )

    // **Der Satz sagt, was los ist; die Ursache steht dabei** ([Ladefeld]).
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Kein Spielpaket.')
    })
  })

  // **Ein Set ohne Angabenfelder erklärt den nächsten Schritt** (§6), statt
  // „Keine Daten" zu zeigen.
  it('erklärt den leeren Zustand', async () => {
    belegeApi({
      Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
      Setangaben: () =>
        Promise.resolve({
          felder: [],
          werte: [],
          kopf: leererSetkopf,
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Setangaben setID="1" verlaufsstand={0} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(
        screen.getByText(texte.seiten.setangaben.leerTitel),
      ).toBeInTheDocument()
    })
  })
})

// **Die Eingabetaste lädt die Seite nicht neu.**
//
// Ein `form` ohne `onSubmit` schickt bei Enter ab — im Browser der Vorschau
// heißt das: die Seite lädt neu, und alles Ungespeicherte ist weg. Das Feld
// schreibt beim Verlassen; Enter soll nichts tun.
describe('Die Eingabetaste', () => {
  it('schickt das Formular nicht ab', async () => {
    zeige()

    const feld = await screen.findByLabelText('set code')

    // **Ausgelöst und nicht getippt.** Ein Formular schickt bei Enter nur
    // dann von selbst ab, wenn es genau **eine** Eingabe hat oder einen
    // Absendeknopf — dieses hat drei und keinen. Der Weg dahin ist trotzdem
    // erreichbar: ein Knopf ohne `type` ist einer, und der nächste, der hier
    // einen einsetzt, hat den Neuladefehler ohne diesen Wächter.
    // **`fireEvent` gibt zurück, ob die Vorgabe blieb** — `false` heißt:
    // jemand hat `preventDefault` gerufen. Ein eigener Zuhörer daneben liefe
    // **vor** React, dessen Behandlung an der Wurzel hängt, und sähe die
    // Vorgabe noch unangetastet.
    const formular = feld.closest('form')
    if (formular === null) {
      throw new Error('das Feld steht in keinem Formular')
    }

    const geblieben = fireEvent.submit(formular)

    expect(geblieben).toBe(false)
  })
})

// **Nach ⌘Z steht wieder der alte Wert im Feld.**
//
// # Der Fall
//
// Die Eingaben stehen auf `defaultValue` — sie halten ihren Entwurf selbst.
// React baut sie nur neu auf, wenn ihr Schlüssel sich ändert; sonst bleibt im
// Feld der zurückgenommene Text stehen, während im Bestand der alte steht.
//
// **Eine Anzeige, die etwas behauptet, was nicht gilt** — und der nächste
// Klick auf ein anderes Feld schriebe sie zurück.
//
// In der laufenden Anwendung gesehen: ⌘Z nahm den Wert aus dem Bestand, und
// das Feld zeigte ihn weiter
// (#567).
describe('Nach einer Rücknahme', () => {
  it('zeigt im Feld, was im Bestand steht — ohne die Lasche neu zu bauen', async () => {
    const erste: SetangabenAntwort = {
      felder: [form('set code', { aufschrift: 'Set code' })],
      werte: [{ feld: 'set code', wert: 'PRB' }],
      kopf: leererSetkopf,
      fehler: null,
    }

    const zweite: SetangabenAntwort = {
      felder: [form('set code', { aufschrift: 'Set code' })],
      werte: [],
      kopf: leererSetkopf,
      fehler: null,
    }

    let runde = 0

    belegeApi({
      Vorbehalte: () => Promise.resolve({ vorbehalte: [], fehler: null }),
      Setangaben: () => Promise.resolve(runde === 0 ? erste : zweite),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Setangaben setID="1" verlaufsstand={0} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByLabelText('set code')).toHaveValue('PRB')
    })

    // **Die Rücknahme**: der Bestand gibt jetzt nichts mehr, und der
    // Verlaufsstand zählt weiter.
    runde = 1
    rerender(
      <Sprachraum sprache="de">
        <Setangaben setID="1" verlaufsstand={1} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByLabelText('set code')).toHaveValue('')
    })
  })
})

// **Kein Kasten im Kasten** — AP-12
// (#2369),
// Befund B-28: die Setangaben stehen in der Schublade *Eigenschaften* und
// trugen darin noch einmal einen eigenen Rahmen; Stil, Symbol und
// Schlüsselwörter daneben tun es nicht.
describe('die Setangaben in der Schublade', () => {
  it('stehen ohne eigenen Rahmen', async () => {
    zeige()

    const titel = await screen.findByDisplayValue('Aus der Saat')
    let kasten: HTMLElement | null = titel
    while (kasten !== null) {
      expect(kasten).not.toHaveClass('rounded-lg', 'border')
      kasten = kasten.parentElement
    }
  })
})
