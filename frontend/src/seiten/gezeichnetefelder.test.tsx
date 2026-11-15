// **Bedienbar ist, was auf der Karte liegt.**
//
// # Wofür
//
// Eine frisch angelegte Karte trägt keinen einzigen Feldwert — der Bestand
// schreibt eine leere Zeile, die Vorgaben stehen im Spielpaket. Der Editor
// listete `karte.werte` und zeigte deshalb **nichts**: eine neue Karte ließ
// sich nicht füllen
// (#562).
//
// # Warum die Flächen und nicht die Feldliste des Spiels
//
// MSE fragt vor jedem Feld:
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–157, Fassung f9d9356 vom
// 2021-07-06. Bedienbar ist, was **dieser Stil** zeichnet — nicht, was das
// Spiel kennt: `magic.mse-game` erklärt Felder, die keine Vorlage setzt.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'

import type { FeldbeschreibungDTO, KarteDTO, VorschauAntwort } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

function flaeche(feld: string) {
  return { feld, links: 0, oben: 0, breite: 10, hoehe: 5 }
}

function mitFlaechen(felder: readonly string[]): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: felder.map(flaeche),
      nennungen: [],
      vorgaben: [],
    },
  }
}

function form(
  name: string,
  teile: Partial<FeldbeschreibungDTO> = {},
): FeldbeschreibungDTO {
  return {
    name,
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
    aufschrift: name,
    ...teile,
  }
}

function zeige(karte: KarteDTO, formen: readonly FeldbeschreibungDTO[]) {
  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte}
        formen={formen}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )
}

describe('Die Felder, die der Editor anbietet', () => {
  it('bietet ein leeres Feld an, das die Karte nicht führt', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(mitFlaechen(['name', 'rule text'])),
      Probevorschau: () => Promise.resolve(mitFlaechen(['name', 'rule text'])),
    })

    zeige({ id: '5', nummer: 0, werte: [] }, [form('name'), form('rule text')])

    await waitFor(() => {
      expect(screen.getByLabelText('name')).toHaveValue('')
    })

    expect(screen.getByLabelText('rule text')).toHaveValue('')
  })

  // **Und was die Karte führt, bleibt stehen** — auch wenn dieser Stil es nicht
  // zeichnet. Ein Set aus MTG JSON trägt Felder, für die die gewählte Vorlage
  // keine Ebene hat; sie zu verstecken hieße, sie stillschweigend zu verlieren.
  //
  // **Das Beispiel war bis zum 2026-08-27 `notes`**, und das taugt nicht mehr:
  // es ist kein unbekanntes Feld, sondern eine **Eigenschaft jeder Karte**
  // (`String notes;`, `src/data/card.hpp`, Zeile 39) und hat seit #1166 seinen
  // eigenen Kasten unter der Kartenliste. Ein Name, der inzwischen etwas
  // anderes bedeutet, prüft nicht mehr das, wofür der Lauf da ist.
  // **Seit UX-05 steht es unter „Experte"** — ein Feld, das das Spielpaket
  // nicht erklärt, gehört nicht zwischen Kartenname und Regeltext
  // (#1677).
  // **Behalten heißt weiterhin behalten**: der Weg dorthin ist ein Klick,
  // und genau den geht dieser Lauf.
  it('behält ein Feld, das die Karte führt, aber niemand zeichnet', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(mitFlaechen(['name'])),
      Probevorschau: () => Promise.resolve(mitFlaechen(['name'])),
    })

    zeige(
      {
        id: '5',
        nummer: 0,
        werte: [
          { feld: 'name', wert: 'Feuerball' },
          { feld: 'mtgjson id', wert: 'aus MTG JSON' },
        ],
      },
      [form('name')],
    )

    await waitFor(() => {
      expect(screen.getByLabelText('name')).toHaveValue('Feuerball')
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.seiten.karten.tiefeExperte,
      }),
    )

    expect(screen.getByLabelText('mtgjson id')).toHaveValue('aus MTG JSON')
  })

  // **Ein Feld ohne Wert zeigt seinen Anfangswert.**
  //
  //	value( !field->initial.empty() ? field->initial
  //	     : initial_first_choice    ? field->choices->choiceName(0) … )
  //
  // `src/data/field/choice.cpp`, Zeilen 279–283, Fassung f9d9356 vom
  // 2021-07-06. Ein `select` ohne Wert zeigt sonst seine **erste** Option:
  // `rarity` beginnt mit `basic land` und hat `initial: common`, und der
  // Editor behauptete damit eine Seltenheit, die die Karte nicht trägt und
  // die auch nicht auf ihr steht.
  it('zeigt den Anfangswert eines Auswahlfeldes', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(mitFlaechen(['rarity'])),
      Probevorschau: () => Promise.resolve(mitFlaechen(['rarity'])),
    })

    zeige({ id: '5', nummer: 0, werte: [] }, [
      form('rarity', {
        art: 'choice',
        wahlen: ['basic land', 'common', 'rare'],
        vorgabe: 'common',
        teilfelder: [],
        reihe: '',
        stelle: 0,
        editierbar: true,
        gespeichert: true,
      }),
    ])

    await waitFor(() => {
      expect(screen.getByLabelText('rarity')).toHaveValue('common')
    })
  })

  // **Und der Wert der Karte schlägt ihn.** Sonst zeigte eine Karte, deren
  // Seltenheit ausdrücklich `basic land` ist, `common` — der Anfangswert ist,
  // was ohne Wert gilt, nicht was statt seiner gilt.
  it('lässt den Wert der Karte den Anfangswert schlagen', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(mitFlaechen(['rarity'])),
      Probevorschau: () => Promise.resolve(mitFlaechen(['rarity'])),
    })

    zeige(
      {
        id: '5',
        nummer: 0,
        werte: [{ feld: 'rarity', wert: 'basic land' }],
      },
      [
        form('rarity', {
          art: 'choice',
          wahlen: ['basic land', 'common', 'rare'],
          vorgabe: 'common',
          teilfelder: [],
          reihe: '',
          stelle: 0,
          editierbar: true,
          gespeichert: true,
        }),
      ],
    )

    await waitFor(() => {
      expect(screen.getByLabelText('rarity')).toHaveValue('basic land')
    })
  })

  // **Beschriftet wird mit der Aufschrift, nicht mit dem Namen.**
  //
  //	if (caption.default_.empty()) caption.default_ = name_to_caption(name);
  //
  // `src/data/field.cpp`, Zeile 69, Fassung f9d9356 vom 2021-07-06. Der Name
  // ist, was Skripte ansprechen; die Aufschrift ist, was dransteht — und in
  // einem Paket mit `localized caption:` steht sie in der Sprache der
  // Oberfläche.
  it('beschriftet mit der Aufschrift und nicht mit dem Namen', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(mitFlaechen(['casting cost'])),
      Probevorschau: () => Promise.resolve(mitFlaechen(['casting cost'])),
    })

    zeige({ id: '5', nummer: 0, werte: [] }, [
      form('casting cost', { aufschrift: 'Casting cost' }),
    ])

    await waitFor(() => {
      expect(screen.getByText('Casting cost')).toBeInTheDocument()
    })

    // **Und der Name bleibt die Kennung.** `aria-label` trägt ihn, denn daran
    // hängt, welches Feld geschrieben wird — und welches ein Klick in die
    // Vorschau aufschlägt.
    expect(screen.getByLabelText('casting cost')).toBeInTheDocument()
  })

  // **Jedes Feld einmal.** Ein Feldstil kann mehrere Ebenen für dasselbe Feld
  // legen — Text und Schatten etwa; zwei Eingabefelder mit demselben
  // `aria-label` wären zwei Wahrheiten über einen Wert.
  it('zeigt ein Feld mit mehreren Ebenen nur einmal', async () => {
    belegeApi({
      Vorschau: () => Promise.resolve(mitFlaechen(['name', 'name'])),
      Probevorschau: () => Promise.resolve(mitFlaechen(['name', 'name'])),
    })

    zeige({ id: '5', nummer: 0, werte: [] }, [form('name')])

    await waitFor(() => {
      expect(screen.getAllByLabelText('name')).toHaveLength(1)
    })
  })
})
