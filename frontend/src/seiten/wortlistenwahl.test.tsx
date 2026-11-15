// **Die Auswahllisten eines Feldes im Karteneditor**
// (#1074).
//
// # Wofür
//
// `magic.mse-game` führt 51 Wortlisten mit 432 Wörtern. Wer den Typ einer
// Karte setzen will, tippte ihn bisher ab und hoffte auf die Schreibung; MSE
// bietet seit jeher ein Menü an (`DropDownWordList`, `src/gui/value/text.cpp`,
// Zeilen 109–300, Fassung f9d9356 vom 2021-07-06).
//
// # Was hier geprüft wird
//
// Der Weg **durch** die Naht: was die Bindung schickt, wird zu Kästchen und
// einer Liste — und ein Klick geht mit **Bereich und Weg** zurück, nicht mit
// einem fertigen Wert. Wer den Wert im Fenster zusammensetzte, baute MSEs
// Auswahllogik ein zweites Mal nach.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  FehlerDTO,
  FeldbeschreibungDTO,
  KarteDTO,
  VorschauAntwort,
  WortlistenbereichDTO,
  WortmenueeintragDTO,
} from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function vorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [{ feld: 'type', links: 0, oben: 0, breite: 10, hoehe: 5 }],
      nennungen: [],
      vorgaben: [],
    },
  }
}

function form(name: string): FeldbeschreibungDTO {
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
  }
}

function eintrag(teile: Partial<WortmenueeintragDTO>): WortmenueeintragDTO {
  return {
    name: '',
    istPraefix: false,
    zeileDarunter: false,
    aktiv: false,
    kinder: null,
    ...teile,
  }
}

// typenbereich bildet `type-en` nach: zwei Präfixe, drei Punkte, eine Gruppe.
function typenbereich(): WortlistenbereichDTO {
  return {
    liste: 'type-en',
    text: 'Legendary Creature',
    kaestchen: true,
    gewaehlt: 3,
    eintraege: [
      eintrag({ name: 'Token ', istPraefix: true }),
      eintrag({ name: 'Legendary ', istPraefix: true, aktiv: true }),
      eintrag({ name: 'Creature', aktiv: true }),
      eintrag({ name: 'Instant' }),
      eintrag({
        name: 'Nichtstandard',
        kinder: [eintrag({ name: 'Plane' }), eintrag({ name: 'Scheme' })],
      }),
    ],
  }
}

const beschriftung = 'Auswahl aus type-en'

function zeige(
  bereiche: WortlistenbereichDTO[],
  antwort: { fehler: FehlerDTO | null } = { fehler: null },
) {
  const WortWaehlen = vi.fn(() => Promise.resolve(antwort))

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    Symbolpalette: () =>
      Promise.resolve({ eintraege: [], roh: false, fehler: null }),
    Wortlisten: () => Promise.resolve({ bereiche, fehler: null }),
    WortWaehlen,
  })

  const karte: KarteDTO = {
    id: '5',
    nummer: 0,
    werte: [{ feld: 'type', wert: 'Legendary Creature' }],
  }

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte}
        formen={[form('type')]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )

  return WortWaehlen
}

// **Die Auswahl gehört zum Feld, in dem jemand steht** — wie die
// Symbolpalette: 24 Felder zugleich zu fragen wären 24 Läufe durch die
// Skriptumgebung, von denen 23 leer zurückkämen.
async function stehImTypfeld() {
  await userEvent.click(await screen.findByLabelText('type'))
}

describe('Die Wortlistenwahl', () => {
  it('zeigt Präfixe als Kästchen und den Rest als Liste', async () => {
    zeige([typenbereich()])
    await stehImTypfeld()

    // **Der Name kommt ohne sein Leerzeichen an.** Im Paket heißt das Präfix
    // `Token `, damit es vor dem nächsten Wort steht; eine Beschriftung
    // normalisiert Leerraum, und sichtbar ist ohnehin keiner.
    const token = await screen.findByRole('checkbox', { name: 'Token' })
    expect(token).not.toBeChecked()

    // **Das aktive Präfix ist angekreuzt** — es steht im Wert.
    expect(screen.getByRole('checkbox', { name: 'Legendary' })).toBeChecked()

    // **Und die Punkte stehen in einer Liste**, deren Wert der Weg ist.
    expect(screen.getByRole('combobox', { name: beschriftung })).toHaveValue(
      '2',
    )
  })

  it('schickt Bereich und Weg, nicht den fertigen Wert', async () => {
    const WortWaehlen = zeige([typenbereich()])
    await stehImTypfeld()

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Token' }),
    )

    await waitFor(() => {
      expect(WortWaehlen).toHaveBeenCalledWith(
        '1',
        '5',
        'type',
        0,
        [0],
        ['Legendary Creature'],
      )
    })
  })

  it('schickt beim Wechsel in der Liste den Weg des Eintrags', async () => {
    const WortWaehlen = zeige([typenbereich()])
    await stehImTypfeld()

    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: beschriftung }),
      '3',
    )

    await waitFor(() => {
      expect(WortWaehlen).toHaveBeenCalledWith(
        '1',
        '5',
        'type',
        0,
        [3],
        ['Legendary Creature'],
      )
    })
  })

  it('reicht ein Untermenü als eingerückten Eintrag durch', async () => {
    const WortWaehlen = zeige([typenbereich()])
    await stehImTypfeld()

    const liste = await screen.findByRole('combobox', { name: beschriftung })

    // **Der Gruppenkopf ist abgeschaltet** — bei MSE öffnet er ein
    // Untermenü und wird selbst nie gewählt.
    expect(screen.getByRole('option', { name: 'Nichtstandard' })).toBeDisabled()

    await userEvent.selectOptions(liste, '4.1')

    await waitFor(() => {
      expect(WortWaehlen).toHaveBeenCalledWith(
        '1',
        '5',
        'type',
        0,
        [4, 1],
        ['Legendary Creature'],
      )
    })
  })

  it('zeigt »nichts davon«, wenn kein Eintrag passt', async () => {
    zeige([
      {
        liste: 'type-en',
        text: 'Artifact',
        kaestchen: false,
        gewaehlt: -1,
        eintraege: [eintrag({ name: 'Creature' })],
      },
    ])
    await stehImTypfeld()

    expect(
      await screen.findByRole('combobox', { name: beschriftung }),
    ).toHaveValue('')

    expect(
      screen.getByRole('option', {
        name: texte.seiten.karten.wortlisten.keine,
      }),
    ).toBeInTheDocument()
  })

  // **Ein gescheitertes Schreiben erscheint am Feld** (§6, §7 Regel 3).
  //
  // Nicht als anonymer Hinweis irgendwo — und die Karte wird **nicht**
  // nachgeladen: ein Lesefehler überschriebe sonst die Meldung, die sagt, was
  // zu tun ist.
  it('meldet einen Fehler am verursachenden Feld', async () => {
    zeige([typenbereich()], {
      fehler: {
        code: 'konflikt',
        schluessel: 'probe',
        meldung: 'Diese Karte hat sich inzwischen geändert.',
        feld: 'type',
      },
    })

    await stehImTypfeld()

    await userEvent.click(
      await screen.findByRole('checkbox', { name: 'Token' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Diese Karte hat sich inzwischen geändert.',
    )
  })

  it('zeigt nichts, wo ein Feld keine Auswahl hat', async () => {
    zeige([])
    await stehImTypfeld()

    await screen.findByLabelText('type')

    // **Nach ihrem Namen gefragt** (UXR-06, #1914): seit der Maßstab der
    // Vorschau eine Auswahl ist, steht neben dem Formular eine zweite.
    expect(
      screen.queryByRole('combobox', { name: beschriftung }),
    ).not.toBeInTheDocument()
  })

  it('zeigt eine Liste ohne Punkte nur als Kästchen', async () => {
    zeige([
      {
        liste: 'type-en',
        text: 'Token ',
        kaestchen: true,
        gewaehlt: 0,
        eintraege: [eintrag({ name: 'Token ', istPraefix: true, aktiv: true })],
      },
    ])
    await stehImTypfeld()

    expect(await screen.findByRole('checkbox', { name: 'Token' })).toBeChecked()

    expect(
      screen.queryByRole('combobox', { name: beschriftung }),
    ).not.toBeInTheDocument()
  })
})
