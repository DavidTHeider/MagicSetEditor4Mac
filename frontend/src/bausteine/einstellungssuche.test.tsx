// **Eine Zeile der Einstellungsseite: Herkunft, Wirksames, Zurücksetzen** —
// Werkstück U11 von Phase 5.1
// (#1422).
//
// # Hier stand einmal auch die Suche
//
// **Sie ist am 2026-09-21 umgebaut worden** (UX-04.AK-03,
// #1676):
// der Trefferkasten zeichnete zu jedem Treffer eine **zweite Kopie**
// desselben Bedienelements, und das Artboard markiert den Kasten
// stattdessen an seinem Platz. Was die drei Läufe darüber geprüft haben,
// steht seitdem woanders:
//
//   - **wonach verglichen wird** in
//     [`einstellungstreffer.test.ts`](./einstellungstreffer.test.ts),
//   - **was die Seite daraus macht** in
//     [`seiten/trefferanzeige.test.tsx`](../seiten/trefferanzeige.test.tsx).
//
// Was hier bleibt, ist die **Zeile selbst** — und die gab es schon vor der
// Suche.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Wahl } from '../seiten/Wahl'
import { Sprachraum } from '../sprache/kontext'
import { dienstnamen } from './dienstnamen'
import type { EinstellungDTO } from '../api'

const schrift: EinstellungDTO = {
  schluessel: 'schriftgroesse',
  wert: 'gross',
  wirksam: 'gross',
  werte: ['normal', 'gross'],
  bereich: 'anwendung',
  ebene: 'profil',
  herkunft: 'profil',
  vorgabe: 'normal',
}

const scryfall: EinstellungDTO = {
  schluessel: 'dienst.scryfall',
  wert: 'an',
  wirksam: 'an',
  werte: ['aus', 'an'],
  bereich: 'dienste',
  ebene: 'profil',
  herkunft: 'vorgabe',
  vorgabe: 'an',
}

describe('eine Zeile mit Herkunft', () => {
  it('nennt die Herkunft als Marke — und das Zurücksetzen nur abseits der Vorgabe', () => {
    const aendern = vi.fn()
    const { rerender } = render(
      <Sprachraum sprache="de">
        <Wahl eintrag={schrift} angehalten={false} aendern={aendern} />
      </Sprachraum>,
    )

    // **Der Wert steht im Feld, die Herkunft als Marke** (AP-11, #2368);
    // der ganze Satz für das Vorleseprogramm.
    expect(screen.getByText('von diesem Profil')).toBeInTheDocument()
    expect(
      screen.getByText('Gesetzt von diesem Profil, Ebene Profil.'),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: /zurücksetzen/ }),
    ).toHaveTextContent('Zurücksetzen auf „Normal“')

    rerender(
      <Sprachraum sprache="de">
        <Wahl eintrag={scryfall} angehalten={false} aendern={aendern} />
      </Sprachraum>,
    )

    expect(
      screen.queryByRole('button', { name: /zurücksetzen/ }),
    ).not.toBeInTheDocument()
    // **Die Vorgabe trägt keine Marke.**
    expect(document.querySelector('[data-herkunftsmarke]')).toBeNull()

    // **Ohne Herkunft keine Zeile** — die Vorrichtung darf sie weglassen.
    const ohneHerkunft: EinstellungDTO = {
      schluessel: schrift.schluessel,
      wert: schrift.wert,
      wirksam: schrift.wirksam,
      werte: schrift.werte,
      bereich: schrift.bereich,
    }
    rerender(
      <Sprachraum sprache="de">
        <Wahl eintrag={ohneHerkunft} angehalten={false} aendern={aendern} />
      </Sprachraum>,
    )
    expect(document.querySelector('[data-herkunftsmarke]')).toBeNull()
  })

  // **Fragen tragen Dienstnamen, und ein Schlüssel ohne Namen steht roh da** —
  // die beiden Ränder der Beschriftung.
  // **Die Werte einer Frage tragen Dienstnamen, Unbekanntes steht roh.**
  //
  // **Ohne die Suche gerendert**, seit der Trefferkasten weg ist: es waren
  // immer Aussagen über die **Zeile**, und die Suche war nur der Weg
  // dorthin.
  it('deutet Fragen mit Dienstnamen und nennt Unbekanntes roh', () => {
    render(
      <Sprachraum sprache="de">
        <Wahl
          eintrag={{
            schluessel: 'frage.preise',
            wert: 'mtgjson',
            wirksam: 'mtgjson',
            werte: ['scryfall', 'mtgjson'],
            ebene: 'mars',
            herkunft: 'wolke',
            vorgabe: 'scryfall',
          }}
          werteNamen={dienstnamen}
          angehalten={false}
          aendern={vi.fn()}
        />
      </Sprachraum>,
    )

    // **Zwei Dienste sind zwei Segmente** (UXR-F2, #2008).
    expect(
      screen.getByRole('radio', { name: dienstnamen.scryfall ?? '' }),
    ).toBeInTheDocument()

    // Unbekannte Herkunft und Ebene stehen roh.
    expect(screen.getByText('von wolke')).toBeInTheDocument()
    expect(
      screen.getByText('Gesetzt von wolke, Ebene mars.'),
    ).toBeInTheDocument()
  })

  // **Und ein Schlüssel ohne Namen trägt sich selbst.** Er kann nur aus
  // einer neueren Go-Seite kommen; eine Zeile ohne Beschriftung wäre
  // schlimmer als ein roher Schlüssel.
  it('beschriftet einen unbekannten Schlüssel mit sich selbst', () => {
    render(
      <Sprachraum sprache="de">
        <Wahl
          eintrag={{
            schluessel: 'geheim.knopf',
            wert: 'a',
            werte: ['a', 'b'],
          }}
          angehalten={false}
          aendern={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByLabelText('geheim.knopf')).toBeInTheDocument()
  })

  // **Ohne Wirksames zählt der Wert** — eine `Wahlbare` ohne `wirksam`, wie
  // Stil- und Exportoptionen sie liefern, aber mit Herkunft. Zu sehen ist es,
  // wo die Wahl erbt (AP-11, #2368).
  it('nimmt den Wert, wo kein Wirksames steht', () => {
    render(
      <Sprachraum sprache="de">
        <Wahl
          eintrag={{
            schluessel: 'x',
            wert: 'erben',
            werte: ['erben'],
            herkunft: 'wolke',
            ebene: 'mars',
          }}
          angehalten={false}
          aendern={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(screen.getByText('Es gilt „erben“.')).toBeInTheDocument()
  })

  // **Ein leerer wirksamer Wert ist keiner** — „Sprache" endet bei „erben",
  // und der Satz sagt, dass der Rechner entscheidet, statt „“ zu zeigen.
  it('sagt bei leerem Wirksamen, dass der Rechner vorgibt', () => {
    render(
      <Sprachraum sprache="de">
        <Wahl
          eintrag={{
            schluessel: 'sprache',
            wert: 'erben',
            wirksam: '',
            werte: ['erben', 'de', 'en'],
            herkunft: 'vorgabe',
            ebene: 'nutzer',
          }}
          angehalten={false}
          aendern={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByText('Es gilt, was der Rechner vorgibt.'),
    ).toBeInTheDocument()
  })

  // **Und das Zurücksetzen setzt zurück.**
  //
  // **Der Knopf wurde bis zum 2026-09-21 über die Trefferliste gedrückt**;
  // mit ihr fiel der Lauf weg, und die Abdeckung hat es angestrichen — die
  // Zeile `aendern(vorgabe)` stand danach unbetreten da. Sie gehört zur
  // Zeile und nicht zur Suche.
  it('setzt auf die Vorgabe zurück', async () => {
    const aendern = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Wahl
          eintrag={{
            schluessel: 'schriftgroesse',
            wert: 'gross',
            wirksam: 'gross',
            werte: ['normal', 'gross'],
            ebene: 'profil',
            herkunft: 'profil',
            vorgabe: 'normal',
          }}
          angehalten={false}
          aendern={aendern}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', {
        name: 'Schriftgröße auf die Vorgabe zurücksetzen',
      }),
    )

    expect(aendern).toHaveBeenCalledWith('normal')
  })
})
