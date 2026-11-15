// Ein gerechneter Vorgabewert ist im Editor zu sehen —
// #563.
//
// 142 der 249 Felder von `magic.mse-game` tragen ein `default:`. Die
// **gezeichnete** Karte trug sie; das **Feld daneben** war leer. Wer die Karte
// ansah, sah eine Farbe; wer das Feld ansah, sah nichts.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Feldzeile } from './Feldzeile'
import { Karteneditor } from './Karteneditor'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

function zeige(teile: Partial<Parameters<typeof Feldzeile>[0]> = {}) {
  return render(
    <Sprachraum sprache="de">
      <Feldzeile
        feld="card color"
        wert=""
        form={{
          name: 'card color',
          aufschrift: 'Card color',
          art: 'text',
          beschreibung: '',
          wahlen: [],
          mehrzeilig: false,
          vorgabe: '',
          teilfelder: [],
          reihe: '',
          stelle: 0,
          editierbar: true,
          gespeichert: true,
        }}
        fehler={undefined}
        laeuft={false}
        palette={undefined}
        wortlisten={[]}
        wortWaehlen={vi.fn()}
        hergeleitet={undefined}
        gewaehlt={vi.fn()}
        getippt={vi.fn()}
        schreiben={vi.fn()}
        einfuegen={vi.fn()}
        bildWaehlen={vi.fn()}
        {...teile}
      />
    </Sprachraum>,
  )
}

describe('Ein hergeleiteter Feldwert', () => {
  // **Der erste Abnahmekasten**: die Karte zeigt eine Farbe, das Feld zeigt
  // sie auch — und **als Herleitung gekennzeichnet**, nicht als getippter
  // Wert.
  it('steht am Feld und ist als Herleitung benannt', () => {
    zeige({ hergeleitet: 'blau' })

    expect(screen.getByText(/blau/)).toBeInTheDocument()
    expect(
      screen.getByText(new RegExp(texte.seiten.karten.hergeleitet)),
    ).toBeInTheDocument()
  })

  // **Und im Feld selbst, als Platzhalter.** Ein Hinweis unter dem Feld sagt,
  // *dass* etwas hergeleitet ist; der Platzhalter zeigt es **dort**, wo der
  // Wert stünde — grau, also unterscheidbar von getipptem Text.
  it('steht als Platzhalter im Feld', () => {
    zeige({ hergeleitet: 'blau' })

    expect(screen.getByRole('textbox', { name: 'card color' })).toHaveAttribute(
      'placeholder',
      'blau',
    )
  })

  // **Ohne Herleitung steht nichts da.** Ein Hinweis „hier ist nichts
  // hergeleitet" wäre Lärm an 107 der 249 Felder.
  it('bleibt ohne Herleitung stumm', () => {
    zeige()

    expect(
      screen.queryByText(new RegExp(texte.seiten.karten.hergeleitet)),
    ).not.toBeInTheDocument()
  })

  // **Ein eigener Wert verdrängt sie.** Kommt eine Vorgabe für ein Feld, das
  // die Karte führt, ist das ein Fehler weiter unten — die Zeile zeigt dann
  // trotzdem den eigenen Wert und keinen Hinweis, statt beides zu behaupten.
  it('zeigt bei eigenem Wert keinen Hinweis', () => {
    zeige({ wert: 'rot', hergeleitet: 'blau' })

    expect(screen.getByRole('textbox', { name: 'card color' })).toHaveValue(
      'rot',
    )
    expect(
      screen.queryByText(new RegExp(texte.seiten.karten.hergeleitet)),
    ).not.toBeInTheDocument()
  })
})

// **Und der Weg durch den Editor** — die Naht, nicht nur die Zeile.
//
// Die Vorgaben kommen mit der **Vorschau**, nicht mit `Kartenfelder`: ein
// `default:` ist ein Skript über `card`, und dieselbe Vorlage gibt für zwei
// Karten zwei Werte. Ohne diesen Test stünde nur fest, dass die Zeile eine
// Herleitung zeigen **kann**.
describe('Der Editor reicht die Herleitungen durch', () => {
  it('zeigt die Herleitung aus der Vorschau am Feld', async () => {
    belegeApi({
      Vorschau: () =>
        Promise.resolve({
          vorschau: {
            stil: '',
            gezeichnet: '',
            bild: '/vorschau/AAAA.png',
            breite: 10,
            hoehe: 14,
            dpi: 150,
            befunde: [],
            flaechen: [],
            nennungen: [],
            vorgaben: [{ feld: 'card color', wert: 'blau' }],
          },
        }),
      Symbolpalette: () => Promise.resolve({ eintraege: [], roh: false }),
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={{
            id: '1',
            nummer: 0,
            werte: [{ feld: 'card color', wert: '' }],
          }}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(new RegExp(texte.seiten.karten.hergeleitet)),
    ).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'card color' })).toHaveAttribute(
      'placeholder',
      'blau',
    )
  })
})

// **Auch eine Auswahl und eine Ankreuzfläche zeigen die Herleitung** —
// der Zweitblick über Phase G hat es gefunden.
//
// Der erste Anlauf reichte sie nur als **Platzhalter** durch, und einen
// Platzhalter haben nur Textfelder. Bei `choice` und `boolean` stand der
// gerechnete Wert im Hinweis darunter, während das Steuerelement seinen
// eigenen ersten Eintrag zeigte und die Karte mit dem gerechneten gezeichnet
// wurde: **drei Antworten auf dieselbe Frage.**
//
// Der Fall aus dem Bericht: eine M15-Karte mit `super type: Creature` und ohne
// eigenes `type symbol` bekommt als Vorgabe `creature`; die Auswahl zeigte
// `front triangle`.
describe('Eine Herleitung an Auswahl und Ankreuzfläche', () => {
  it('steht in der Auswahl', () => {
    zeige({
      feld: 'type symbol',
      wert: '',
      hergeleitet: 'creature',
      form: {
        name: 'type symbol',
        aufschrift: 'Type symbol',
        art: 'choice',
        beschreibung: '',
        wahlen: ['front triangle', 'creature'],
        mehrzeilig: false,
        vorgabe: '',
        teilfelder: [],
        reihe: '',
        stelle: 0,
        editierbar: true,
        gespeichert: true,
      },
    })

    expect(screen.getByRole('combobox', { name: 'type symbol' })).toHaveValue(
      'creature',
    )
  })

  it('steht an der Ankreuzfläche', () => {
    zeige({
      feld: 'shape',
      wert: '',
      hergeleitet: 'yes',
      form: {
        name: 'shape',
        aufschrift: 'Shape',
        art: 'boolean',
        beschreibung: '',
        wahlen: [],
        mehrzeilig: false,
        vorgabe: '',
        teilfelder: [],
        reihe: '',
        stelle: 0,
        editierbar: true,
        gespeichert: true,
      },
    })

    expect(screen.getByRole('checkbox', { name: 'shape' })).toBeChecked()
  })

  // **Und ein eigener Wert schlägt sie.** Wer `no` schreibt, meint `no`, auch
  // wenn die Herleitung `yes` sagt.
  it('lässt den eigenen Wert die Herleitung schlagen', () => {
    zeige({
      feld: 'shape',
      wert: 'no',
      hergeleitet: 'yes',
      form: {
        name: 'shape',
        aufschrift: 'Shape',
        art: 'boolean',
        beschreibung: '',
        wahlen: [],
        mehrzeilig: false,
        vorgabe: '',
        teilfelder: [],
        reihe: '',
        stelle: 0,
        editierbar: true,
        gespeichert: true,
      },
    })

    expect(screen.getByRole('checkbox', { name: 'shape' })).not.toBeChecked()
  })
})

// **Eine Herleitung ändert sich, und das Steuerelement geht mit.**
//
// Der Nachlauf über Phase G hat gefunden, dass der Befund von oben nur beim
// **ersten Einhängen** behoben war: `defaultValue` und `defaultChecked` sind
// unkontrolliert, und ein späterer Wechsel erreicht sie nicht mehr.
//
// **Das ist keine Randlage, sondern der Normalfall.** Die Probevorschau
// rechnet die Herleitungen bei jedem Tastendruck neu; wer die Manakosten
// ändert, ändert damit die hergeleitete Farbe. Danach zeigte die Karte die
// neue, der Hinweis unter dem Feld die neue — und die Auswahlliste die alte.
describe('Eine Herleitung, die sich ändert', () => {
  const auswahlform = {
    name: 'card color',
    aufschrift: 'Card color',
    art: 'choice',
    beschreibung: '',
    wahlen: ['white', 'red'],
    mehrzeilig: false,
    vorgabe: '',
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
  }

  it('bewegt die Auswahl mit', () => {
    const { rerender } = zeige({
      feld: 'card color',
      wert: '',
      hergeleitet: 'white',
      form: auswahlform,
    })

    expect(screen.getByRole('combobox', { name: 'card color' })).toHaveValue(
      'white',
    )

    rerender(
      <Sprachraum sprache="de">
        <Feldzeile
          feld="card color"
          wert=""
          form={auswahlform}
          fehler={undefined}
          laeuft={false}
          palette={undefined}
          wortlisten={[]}
          wortWaehlen={vi.fn()}
          hergeleitet="red"
          gewaehlt={vi.fn()}
          getippt={vi.fn()}
          schreiben={vi.fn()}
          einfuegen={vi.fn()}
          bildWaehlen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('combobox', { name: 'card color' })).toHaveValue(
      'red',
    )
  })

  it('bewegt die Ankreuzfläche mit', () => {
    const form = {
      name: 'shape',
      aufschrift: 'Shape',
      art: 'boolean',
      beschreibung: '',
      wahlen: [],
      mehrzeilig: false,
      vorgabe: '',
      teilfelder: [],
      reihe: '',
      stelle: 0,
      editierbar: true,
      gespeichert: true,
    }

    const { rerender } = zeige({
      feld: 'shape',
      wert: '',
      hergeleitet: 'no',
      form,
    })

    expect(screen.getByRole('checkbox', { name: 'shape' })).not.toBeChecked()

    rerender(
      <Sprachraum sprache="de">
        <Feldzeile
          feld="shape"
          wert=""
          form={form}
          fehler={undefined}
          laeuft={false}
          palette={undefined}
          wortlisten={[]}
          wortWaehlen={vi.fn()}
          hergeleitet="yes"
          gewaehlt={vi.fn()}
          getippt={vi.fn()}
          schreiben={vi.fn()}
          einfuegen={vi.fn()}
          bildWaehlen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByRole('checkbox', { name: 'shape' })).toBeChecked()
  })
})
