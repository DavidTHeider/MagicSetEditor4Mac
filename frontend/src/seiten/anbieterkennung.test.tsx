// **Die Kennung, unter der sich diese Anwendung bei einem Anbieter ausweist**
// — Nacharbeit zu Werkstück 4 von Phase 4.12
// (#1351).
//
// # Der Befund
//
// **Die Anmeldung war unerreichbar.** Die Nutzerseite verwies auf „die
// Einstellungen"; dort gab es kein Feld, und `domain.PruefeEinstellung` wies
// den Schlüssel `anbieter.<name>.kennung` sogar ab. Vier Anbieter, ein
// fertiger OAuth-Lauf — und kein Weg hinein.
//
// **Gefunden vom Auftraggeber am 2026-09-06** an der laufenden Anwendung: er
// suchte die Stelle, auf die der Satz verweist, und fand nichts.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Einstellungen } from './Einstellungen'
import { Sprachraum } from '../sprache/kontext'
import { alsSatz } from '../test/auszeichnung'
import { belegeApi, raeumeApi } from '../test/api'
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from '../test/bestandspflegeprobe'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'
import { kurzform, langform, ohneAuszeichnung } from '../sprache/kurzform'

const texte = kataloge.de.seiten.einstellungen

afterEach(() => {
  raeumeApi()
})

// eineKennung ist die Einstellung eines Anbieters — **ohne Werteliste**, weil
// jeder Anbieter eine andere Gestalt vergibt.
function eineKennung(anbieter: string, wert = ''): EinstellungDTO {
  return {
    schluessel: `anbieter.${anbieter}.kennung`,
    bereich: 'anmeldung',
    wert,
    wirksam: wert,
    werte: [],
  }
}

function zeige(anmeldung: EinstellungDTO[], setze = vi.fn()) {
  belegeApi({})

  render(
    <Sprachraum sprache="de">
      <Einstellungen
        eintraege={[]}
        dienste={[]}
        fragen={[]}
        ki={[]}
        vorgaben={[]}
        anmeldung={anmeldung}
        ursache={undefined}
        meldefehler={undefined}
        setze={setze}
        laeuft={() => false}
        bestandspflege={stillePflege()}
        ruecksprung={stillerRuecksprung()}
        pflege={stilleDatenpflege()}
        setID={undefined}
      />
    </Sprachraum>,
  )

  return setze
}

describe('Die Anbieterkennung in den Einstellungen', () => {
  it('steht in einem eigenen Gebiet', async () => {
    zeige([eineKennung('github')])

    await userEvent.click(
      screen.getByRole('tab', { name: texte.gruppen.nutzer }),
    )

    // **Und der Hinweis sagt, was BEIM ANBIETER zu tun ist** — ohne den
    // ersten Schritt gibt es nichts einzutragen.
    //
    // **Sichtbar der erste Satz, der Rest hinter dem Fragezeichen** — seit
    // UX-15 (#1771).
    expect(
      await screen.findByText(alsSatz(kurzform(texte.anmeldungHinweis))),
    ).toBeInTheDocument()

    // **Und nichts geht verloren.**
    expect(
      screen
        .getAllByRole('tooltip', { hidden: true })
        .map((eines) => eines.textContent),
    ).toContain(langform(ohneAuszeichnung(texte.anmeldungHinweis)))

    // **Und wohin das Geheimnis gehört** — Befund 3 des Zweitblicks über
    // Phase 4.12: GitHub verlangt eines, und wo es hinsoll, stand nirgends.
    expect(
      await screen.findByText(alsSatz(kurzform(texte.anmeldungGeheimnis))),
    ).toBeInTheDocument()
  })

  it('lässt sich eintragen', async () => {
    const setze = zeige([eineKennung('github')])

    await userEvent.click(
      screen.getByRole('tab', { name: texte.gruppen.nutzer }),
    )

    // **Über den Katalognamen und nicht über den Schlüssel**: das Feld heißt
    // „Kennung bei GitHub", und genau das liest jemand vor dem Bildschirm.
    const feld = screen.getByRole('textbox', {
      name: texte.namen['anbieter.github.kennung'],
    })

    // **Geschrieben wird mit ⏎ oder beim Verlassen** — seit OPS-02 (#1719)
    // hält ein Freitextfeld einen Entwurf, statt jeden Buchstaben zu
    // schreiben; hier stand „ein Zeichen genügt".
    await userEvent.type(feld, 'Ov23{Enter}')

    expect(setze).toHaveBeenCalledWith('anbieter.github.kennung', 'Ov23')
  })

  // **Ohne Anbieter bleibt die Gruppe, und die Felder gehen.**
  //
  // # Hier stand „Ohne Anbieter kein Gebiet"
  //
  // **Seit UX-04 ist eine Gruppe ein Ort und kein Inhalt**
  // (#1676):
  // die neun stehen fest, weil jemand sie **sucht**. Was sie leer zurückließe,
  // wäre eine Lasche ohne Inhalt — und genau deshalb steht in *Nutzer und
  // Profil* auch ohne Anbieter der Satz, der erklärt, wo eine Anmeldung
  // herkommt.
  it('lässt die Gruppe stehen und die Felder fort, wenn es keine Anbieter gibt', async () => {
    zeige([])

    await userEvent.click(
      screen.getByRole('tab', { name: texte.gruppen.nutzer }),
    )

    expect(
      screen.queryByRole('textbox', { name: /github/i }),
    ).not.toBeInTheDocument()
  })
})
