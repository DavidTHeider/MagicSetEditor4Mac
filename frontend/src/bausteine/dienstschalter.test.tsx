// **Die Zeilen eines Dienstes tragen ihren Namen** —
// #1540.
//
// # Der Befund
//
// **Im Fenster stand `dienst.openai.adresse`.** `Dienstschalter` reicht seit
// Phase 4.8 eine eigene Beschriftung durch — die **Eigennamen** der Dienste,
// die nicht übersetzt werden —, und für alles, was dort nicht steht, fiel es
// auf den **rohen Schlüssel** zurück. Werkstück 3 von A18 (#1533) hat die
// Adressen danebengelegt; sie sind keine Eigennamen und stehen im Katalog.
//
// **Gefunden hat es die Sichtprüfung am laufenden Fenster** (2026-09-13) —
// kein Lauf, denn jeder sucht sein Feld über die Beschriftung, die er selbst
// erwartet.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Dienstschalter } from './Dienstschalter'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { kurzform } from '../sprache/kurzform'

import type { EinstellungDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// eineZeile baut den Eintrag, den die Seite reicht.
function eineZeile(schluessel: string, wert: string): EinstellungDTO {
  return {
    schluessel,
    wert,
    wirksam: wert,
    bereich: 'dienste',
    herkunft: 'profil',
    ebene: 'profil',

    // **Eine Adresse hat keine Werteliste** — jede ist eine andere, und
    // `Wahl` zeichnet dafür ein Textfeld statt einer Auswahl.
    werte: schluessel.endsWith('.adresse') ? [] : ['aus', 'an'],
  }
}

// zeige stellt die Schalter hin.
function zeige(...zeilen: EinstellungDTO[]) {
  zeigeMit({}, ...zeilen)
}

// zeigeMit stellt sie unter einer eigenen Aufschrift hin.
function zeigeMit(
  eigenes: { titel?: string; hinweis?: string },
  ...zeilen: EinstellungDTO[]
) {
  render(
    <Sprachraum sprache="de">
      <Dienstschalter
        dienste={zeilen}
        angehalten={() => false}
        setze={() => undefined}
        {...eigenes}
      />
    </Sprachraum>,
  )
}

// **Der Schalter trägt den Eigennamen des Dienstes.**
//
// **Der Steuerlauf** (§10): ohne ihn sagte die Zeile darunter nicht, ob die
// Beschriftung überhaupt noch ankommt.
it('beschriftet den Schalter mit dem Eigennamen', () => {
  zeige(eineZeile('dienst.openai', 'an'))

  expect(screen.getByLabelText('OpenAI')).toBeInTheDocument()
})

// **Und die Adresse ihren Namen aus dem Katalog.**
it('beschriftet die Adresse aus dem Katalog', () => {
  zeige(eineZeile('dienst.openai.adresse', 'http://127.0.0.1:8099'))

  expect(
    screen.getByLabelText(eigene.namen['dienst.openai.adresse']),
  ).toBeInTheDocument()

  expect(screen.queryByText('dienst.openai.adresse')).not.toBeInTheDocument()
})

// **Die Aufschrift darüber sagt, worüber diese Schalter urteilen** —
// UX-04.AK-04
// (#1676).
//
// **Im Fenster stand „Fremde Kartendienste" über den Sprachmodellen** —
// gefunden in der Sichtprüfung am 2026-09-21, als dieselben Schalter in
// zwei Gruppen standen. Eine Aufschrift, die etwas anderes nennt als das,
// was darunter steht, ist schlechter als keine.
it('nimmt eine eigene Aufschrift an', () => {
  zeigeMit(
    { titel: eigene.kiAnbieter, hinweis: eigene.kiAnbieterHinweis },
    eineZeile('dienst.openai', 'an'),
  )

  expect(
    screen.getByRole('region', { name: eigene.kiAnbieter }),
  ).toBeInTheDocument()

  expect(
    screen.getByText(kurzform(eigene.kiAnbieterHinweis)),
  ).toBeInTheDocument()

  // **Und die fremde steht dann nicht mehr da.** Ohne diese Zeile wäre die
  // Zusicherung darüber auch dann grün, wenn beide Aufschriften
  // nebeneinander stünden (§10).
  expect(screen.queryByText(eigene.dienste)).not.toBeInTheDocument()
})

// **Ohne eigene Aufschrift bleibt es bei den Kartendiensten** — der
// Steuerlauf dazu: sonst sagte der Lauf darüber nicht, ob der Rückfall noch
// greift.
it('fällt ohne eigene Aufschrift auf die Kartendienste zurück', () => {
  zeige(eineZeile('dienst.openai', 'an'))

  expect(
    screen.getByRole('region', { name: eigene.dienste }),
  ).toBeInTheDocument()
})
