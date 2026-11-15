// **Die Symbolpalette im Karteneditor**
// (#564).
//
// # Wofür
//
// Im Feld für die Manakosten steht `2RR`. Auf der Karte werden daraus Symbole;
// im Feld bleiben es Buchstaben, und wer nicht weiß, dass `T` das Tap-Symbol
// ist, findet es nicht. MSE bietet ein Menü an, und es steht **im Paket**
// (`insert symbol menu:`, `magic-mana-large.mse-symbol-font`, ab Zeile 841).
//
// # Was eingefügt wird
//
//	String code = style().symbol_font.font->insertSymbolCode(id);
//	if (!style().always_symbol) {
//	  code = _("<sym>") + code + _("</sym>");
//	}
//
// `TextValueEditor::onCommand`, `src/gui/value/text.cpp`, Zeilen 589–593,
// Fassung f9d9356 vom 2021-07-06.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type {
  FeldbeschreibungDTO,
  KarteDTO,
  PalettenEintragDTO,
  VorschauAntwort,
} from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function vorschau(felder: readonly string[]): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: felder.map((feld) => ({
        feld,
        links: 0,
        oben: 0,
        breite: 10,
        hoehe: 5,
      })),
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

function eintrag(teile: Partial<PalettenEintragDTO>): PalettenEintragDTO {
  return {
    art: 'code',
    code: '',
    beschriftung: '',
    rueckfrage: '',
    bild: 'data:image/png;base64,BBBB',
    kinder: [],
    ...teile,
  }
}

function karte(werte: KarteDTO['werte'] = []): KarteDTO {
  return { id: '5', nummer: 0, werte }
}

function zeige(
  eintraege: PalettenEintragDTO[],
  roh: boolean,
  werte: KarteDTO['werte'] = [],
) {
  const FeldSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

  belegeApi({
    Vorschau: () => Promise.resolve(vorschau(['kosten'])),
    Probevorschau: () => Promise.resolve(vorschau(['kosten'])),
    Symbolpalette: () => Promise.resolve({ eintraege, roh, fehler: null }),
    FeldSetzen,
  })

  render(
    <Sprachraum sprache="de">
      <Karteneditor
        setID="1"
        karte={karte(werte)}
        formen={[form('kosten')]}
        geaendert={() => undefined}
      />
    </Sprachraum>,
  )

  return FeldSetzen
}

// **Die Palette gehört zum Feld, in dem jemand steht.** Welche Symbolschrift
// gilt, steht am Feldstil; sie für alle 24 Felder zugleich zu holen wären 24
// Läufe durch die Skriptumgebung, von denen 22 leer zurückkämen. MSE bindet das
// Menü genauso an den gerade offenen Editor (`TextValueEditor::getMenu`,
// `src/gui/value/text.cpp`, Zeilen 616–623, Fassung f9d9356 vom 2021-07-06).
//
// # Ein Klick ist noch keine Palette
//
// **Wer danach synchron zugreift, prüft einen Baum, den die Antwort erst
// aufbauen muss.** Am 2026-09-16 fiel einer von 3.728 Läufen genau daran um,
// während zwei Vorprüfungen zugleich liefen; allein war er dreimal grün.
//
// **Ein Warten hier hilft nicht**, und das ist der zweite Teil der Lehre: der
// erste Anlauf legte einen Spion auf die Naht und wartete auf ihn — der läuft
// **vor** der Antwort, und die drei Tests fielen weiter um. §3 nennt genau
// diese Gestalt: „ein Warten auf den Ruf statt auf die Antwort". Gewartet
// wird auf die **Wirkung**, und die steht an der Aufrufstelle, weil nur sie
// weiß, welche es ist.
async function stehIm(feld: string) {
  await userEvent.click(await screen.findByLabelText(feld))
}

describe('Die Symbolpalette', () => {
  it('zeigt je Symbol eine Schaltfläche mit seinem Bild', async () => {
    zeige(
      [
        eintrag({ code: 'W', beschriftung: 'Weißes Mana' }),
        eintrag({ art: 'line', bild: '' }),
        eintrag({ code: 'U', beschriftung: 'Blaues Mana' }),
      ],
      true,
    )

    await stehIm('kosten')

    const knopf = await screen.findByRole('button', { name: 'Weißes Mana' })
    expect(knopf.querySelector('img')).toHaveAttribute(
      'src',
      'data:image/png;base64,BBBB',
    )

    expect(
      screen.getByRole('button', { name: 'Blaues Mana' }),
    ).toBeInTheDocument()
  })

  // **`always symbol: true` fügt den rohen Code ein.**
  it('fügt den rohen Code ein, wo das Feld nur Symbole trägt', async () => {
    const FeldSetzen = zeige(
      [eintrag({ code: 'W', beschriftung: 'Weiß' })],
      true,
    )

    await stehIm('kosten')
    await userEvent.click(await screen.findByRole('button', { name: 'Weiß' }))

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith('1', '5', 'kosten', 'W', [
        '',
        'W',
      ])
    })
  })

  // **Und sonst wird ausgezeichnet.**
  it('fasst den Code in `<sym>`, wo das Feld Fließtext trägt', async () => {
    const FeldSetzen = zeige(
      [eintrag({ code: 'T', beschriftung: 'Tap' })],
      false,
    )

    await stehIm('kosten')
    await userEvent.click(await screen.findByRole('button', { name: 'Tap' }))

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith(
        '1',
        '5',
        'kosten',
        '<sym>T</sym>',
        ['', '<sym>T</sym>'],
      )
    })
  })

  // **Angehängt und nicht ersetzt.** Wer `2R` getippt hat und `R` klickt, will
  // `2RR` — nicht `R`.
  it('hängt an den vorhandenen Wert an', async () => {
    const FeldSetzen = zeige(
      [eintrag({ code: 'R', beschriftung: 'Rot' })],
      true,
      [{ feld: 'kosten', wert: '2R' }],
    )

    await stehIm('kosten')
    await userEvent.click(await screen.findByRole('button', { name: 'Rot' }))

    await waitFor(() => {
      expect(FeldSetzen).toHaveBeenCalledWith('1', '5', 'kosten', '2RR', [
        '2R',
        '2RR',
      ])
    })
  })

  // **Ein Untermenü wird eine Gruppe.**
  //
  // `magic-mana-large` ordnet so die halben und die hybriden Symbole. Über
  // `full-magic-pack` sind es 433 Untermenüs unter 5.219 Einträgen.
  it('macht aus einem Untermenü eine Gruppe', async () => {
    zeige(
      [
        eintrag({
          art: 'submenu',
          beschriftung: 'Halb',
          kinder: [eintrag({ code: '1/2', beschriftung: 'Ein halbes' })],
        }),
      ],
      true,
    )

    await stehIm('kosten')

    const gruppe = await screen.findByRole('group', { name: 'Halb' })
    expect(
      within(gruppe).getByRole('button', { name: 'Ein halbes' }),
    ).toBeInTheDocument()
  })

  // **Ein `custom` steht noch nicht.**
  //
  // MSE öffnet dafür ein Eingabefeld (`wxGetTextFromUser`,
  // `InsertSymbolMenu::getCode`, `src/data/symbol_font.cpp`, Zeile 462,
  // Fassung f9d9356 vom 2021-07-06). Über das ganze Pack sind es 62 von 5.219
  // Einträgen; sie wegzulassen ist ehrlicher als eine Schaltfläche, die etwas
  // anderes einfügt als das, was dransteht.
  it('lässt einen Eintrag mit Rückfrage aus', async () => {
    zeige(
      [
        eintrag({
          art: 'custom',
          beschriftung: 'Beliebig',
          rueckfrage: 'Wie viel?',
        }),
        eintrag({ code: 'W', beschriftung: 'Weiß' }),
      ],
      true,
    )

    await stehIm('kosten')

    expect(
      await screen.findByRole('button', { name: 'Weiß' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: 'Beliebig' }),
    ).not.toBeInTheDocument()
  })

  // **Ein Code ohne Bild zeigt seinen Code.**
  //
  // Ein Menü nennt Codes, die eine Bedingung gerade weggenommen hat; MSE zeigt
  // dann einen Eintrag ohne Kachel (`makeMenuItem`, Zeilen 508–518). Eine leere
  // Schaltfläche wäre ein Loch in der Reihe.
  it('zeigt den Code, wo kein Bild da ist', async () => {
    zeige([eintrag({ code: 'T', beschriftung: 'Tap', bild: '' })], true)

    await stehIm('kosten')

    const knopf = await screen.findByRole('button', { name: 'Tap' })
    expect(knopf).toHaveTextContent('T')
    expect(knopf.querySelector('img')).toBeNull()
  })

  // **Eine Auswahl bekommt nichts untergeschoben.**
  //
  // MSE hängt das Einfügemenü an den `TextValueEditor` und an keinen anderen
  // (`TextValueEditor::getMenu`, `src/gui/value/text.cpp`, Zeilen 616–623,
  // Fassung f9d9356 vom 2021-07-06). Über die 190 Stile von `M15-Magic-Pack`
  // trägt **kein** `choice`-Feld eine Symbolschrift — alle 79, die eine haben,
  // sind Textfelder. Der Wächter ist für das Paket, das es anders macht: eine
  // Auswahl nähme sonst einen Wert an, den ihre eigene Liste nicht kennt.
  it('schreibt nichts in eine Auswahl', async () => {
    const FeldSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Vorschau: () => Promise.resolve(vorschau(['kosten'])),
      Probevorschau: () => Promise.resolve(vorschau(['kosten'])),
      Symbolpalette: () =>
        Promise.resolve({
          eintraege: [eintrag({ code: 'W', beschriftung: 'Weiß' })],
          roh: true,
          fehler: null,
        }),
      FeldSetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte()}
          formen={[
            { ...form('kosten'), art: 'choice', wahlen: ['eins', 'zwei'] },
          ]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    await stehIm('kosten')
    await userEvent.click(await screen.findByRole('button', { name: 'Weiß' }))

    expect(FeldSetzen).not.toHaveBeenCalled()
  })

  // **Ein Feld ohne Symbolschrift bekommt keine Palette** — kein leerer Kasten,
  // keine Überschrift ohne Inhalt.
  it('zeigt nichts, wo es nichts anzubieten gibt', async () => {
    zeige([], false)

    await stehIm('kosten')

    // **Eine Verneinung braucht einen Beleg, dass die Antwort da war** (§3).
    // Die leere Palette baut nichts auf — also wird das Warten an etwas
    // festgemacht, das in **beiden** Fällen kommt: das Bild der Vorschau.
    // Hier stand die Werkzeugleiste des Maßstabs, die mit dem Bild kam; seit
    // UXR-06 (#1914) ist der Maßstab eine Auswahl und steht schon, während
    // das Bild lädt — das Bild selbst ist der Beleg, auf den sie stand.
    await screen.findByRole('img', { name: texte.seiten.karten.vorschau })

    expect(
      screen.queryByRole('group', { name: texte.seiten.karten.symbole }),
    ).not.toBeInTheDocument()
  })
})
