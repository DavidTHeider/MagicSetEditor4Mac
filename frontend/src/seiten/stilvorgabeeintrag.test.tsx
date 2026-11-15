// Der Vorgabeeintrag einer Farboption
// (#764).
//
// # Was er ist
//
// **Der Eintrag, der sagt „nimm, was die Vorlage rechnet".** MSE stellt ihn
// einer Farboption mit `default:` voran und lässt ihn gewählt, solange niemand
// etwas gesetzt hat:
//
//	void DropDownColorList::select(size_t item) {
//	  if (isDefault(item)) {
//	    cve.change( Defaultable<Color>());
//
// `src/gui/value/color.cpp`, Zeilen 118–120, Fassung f9d9356 vom 2021-07-06 —
// die Wahl **entfernt** den gespeicherten Wert.
//
// **Ohne `default:` gibt es ihn nicht** (`color.cpp`, Zeile 38, dieselbe
// Fassung), und das ist der Fall, den der zweite Test hält: `wert` trägt auch
// den Rückfall und wäre fast nie leer — hinge der Eintrag daran, stünde er bei
// fast jeder Option.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { type Api } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stiloptionen } from './Stiloptionen'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

// farboption gibt eine Farbwahl mit zwei Farben.
function farboption(teile: {
  vorgabe: string
  gesetzt: boolean
  wert: string

  // **Vorhanden heißt nicht ausgerechnet.** Ohne eigene Angabe folgt
  // `hatVorgabe` dem Wert; der eine Fall, in dem beides auseinanderfällt, sagt
  // es ausdrücklich.
  hatVorgabe?: boolean
}) {
  return {
    optionen: [
      {
        name: 'randfarbe',
        art: 'color',
        beschreibung: '',
        moeglichkeiten: [
          { wert: 'rgb(0,0,0)', beschriftung: 'schwarz' },
          { wert: 'rgb(255,255,255)', beschriftung: 'weiß' },
        ],
        eigeneFarbe: false,

        // **Vorhanden heißt nicht ausgerechnet.** Ohne eigene Angabe folgt
        // `hatVorgabe` dem Wert — die Fälle, in denen es auseinanderfällt,
        // sagen es ausdrücklich.
        hatVorgabe: teile.vorgabe !== '',
        hatWahlbilder: false,
        ...teile,
      },
    ],
  }
}

function zeige(
  antwort: ReturnType<typeof farboption>,
  teile: Partial<Api> = {},
) {
  belegeApi({ Stiloptionen: () => Promise.resolve(antwort), ...teile })

  render(
    <Sprachraum sprache="de">
      <Stiloptionen setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )
}

it('stellt den Vorgabeeintrag voran und wählt ihn, solange nichts gesetzt ist', async () => {
  zeige(
    farboption({ vorgabe: 'rgb(0,0,0)', gesetzt: false, wert: 'rgb(0,0,0)' }),
  )

  const liste = await screen.findByRole('combobox', { name: /randfarbe/ })

  // **Zuoberst**, wie bei MSE.
  expect(liste).toHaveTextContent(
    new RegExp(`^${texte.seiten.stil.vorgabe}schwarz`),
  )

  // **Und gewählt** — nicht die Farbe, die zufällig dieselbe ist.
  expect(
    screen.getByRole('option', { name: texte.seiten.stil.vorgabe }),
  ).toBeInTheDocument()
  expect((liste as HTMLSelectElement).selectedIndex).toBe(0)
})

it('entfernt den gespeicherten Wert, wenn jemand die Vorgabe wählt', async () => {
  const StiloptionZuruecksetzen = vi.fn(() => Promise.resolve({ fehler: null }))
  const StiloptionSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige(
    farboption({
      vorgabe: 'rgb(0,0,0)',
      gesetzt: true,
      wert: 'rgb(255,255,255)',
    }),
    { StiloptionZuruecksetzen, StiloptionSetzen },
  )

  const liste = await screen.findByRole('combobox', { name: /randfarbe/ })

  // **Erst der Beleg, dass die Liste schreibt** (§3): eine gewöhnliche Wahl
  // geht durch denselben Horcher. Eine Verneinung davor träfe auch dann zu,
  // wenn niemand zuhört.
  await userEvent.selectOptions(liste, 'rgb(0,0,0)')
  expect(StiloptionSetzen).toHaveBeenCalledWith('1', 'randfarbe', 'rgb(0,0,0)')

  await userEvent.selectOptions(liste, texte.seiten.stil.vorgabe)

  expect(StiloptionZuruecksetzen).toHaveBeenCalledWith('1', 'randfarbe')
})

it('zeigt keinen Vorgabeeintrag, wenn die Vorlage keine Vorgabe nennt', async () => {
  zeige(farboption({ vorgabe: '', gesetzt: false, wert: 'rgb(0,0,0)' }))

  const liste = await screen.findByRole('combobox', { name: /randfarbe/ })

  expect(liste).not.toHaveTextContent(texte.seiten.stil.vorgabe)
  expect((liste as HTMLSelectElement).options).toHaveLength(2)
})

// **Und ein gescheitertes Zurücksetzen meldet sich** — §7 Regel 3: nachgeladen
// wird dann gar nicht, sonst überschriebe ein Lesefehler die Meldung, die
// sagt, was zu tun ist.
it('meldet einen Fehler beim Zurücksetzen', async () => {
  const StiloptionZuruecksetzen = vi.fn(() =>
    Promise.resolve({
      fehler: {
        code: 'validierung' as const,
        schluessel: 'validierung',
        meldung: 'Die Option gibt es nicht.',
      },
    }),
  )

  zeige(
    farboption({
      vorgabe: 'rgb(0,0,0)',
      gesetzt: true,
      wert: 'rgb(255,255,255)',
    }),
    { StiloptionZuruecksetzen },
  )

  const liste = await screen.findByRole('combobox', { name: /randfarbe/ })

  await userEvent.selectOptions(liste, texte.seiten.stil.vorgabe)

  // **Gegen den Katalog und nicht gegen den rohen Satz aus Go** (§9): die
  // Oberfläche zeigt die übersetzte Meldung zum Schlüssel, nicht den Text der
  // Bindung.
  expect(await screen.findByRole('alert')).toHaveTextContent(
    texte.fehler.validierung,
  )
})

// **Vorhanden heißt nicht ausgerechnet** — aus dem Zweitblick über Phase L.
//
// `magic-m15-bigtext.mse-style` schreibt `default: black` **ohne**
// Anführungszeichen. Das ist ein Skriptfehler („der Name »black« ist hier nicht
// bekannt"), der Wert fällt auf `initial: black` zurück, und `vorgabe` bleibt
// leer.
//
// **MSE zeigt den Eintrag dort trotzdem**: `hasDefault()` fragt nach dem
// Skript, nicht nach seinem Ergebnis (`src/gui/value/color.cpp`, Zeile 38,
// Fassung f9d9356 vom 2021-07-06).
//
// Bis zum 2026-08-22 hing der Eintrag an `vorgabe !== ''` und verschwand
// deshalb genau bei der einen Option des Packs, die ihn gebraucht hätte.
it('zeigt den Eintrag auch, wenn die Vorgabe sich nicht ausrechnet', async () => {
  zeige(
    farboption({
      vorgabe: '',
      hatVorgabe: true,
      gesetzt: false,
      wert: 'rgb(0,0,0)',
    }),
  )

  const liste = await screen.findByRole('combobox', { name: /randfarbe/ })

  expect(
    screen.getByRole('option', { name: texte.seiten.stil.vorgabe }),
  ).toBeInTheDocument()
  expect((liste as HTMLSelectElement).selectedIndex).toBe(0)
})
