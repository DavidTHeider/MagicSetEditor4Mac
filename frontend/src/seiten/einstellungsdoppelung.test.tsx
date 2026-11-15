// **Keine Einstellung steht zweimal auf der Seite** — Befund der
// Sichtprüfung vom 2026-09-21, UX-04
// (#1676).
//
// # Der Anlass
//
// Nach dem Umbau auf neun Gruppen standen die fünf Kartendienste **zweimal**
// in der Gruppe *Dienste*: einmal als gewöhnliche Wahl mit dem **rohen
// Schlüssel** als Aufschrift (`dienst.scryfall`), einmal richtig benannt
// über den `Dienstschalter` (*Scryfall*). In *KI* standen so die drei
// Sprachmodelle und **vier Adressen** doppelt.
//
// # Warum kein Lauf es gesehen hat
//
// **Weil jeder von ihnen ein Element über seinen Namen sucht und es
// findet.** Dass daneben ein zweites steht, kam in keiner Zusicherung vor —
// und „zeichnet keine Auswahlmöglichkeit ohne Text" ging durch, weil die
// Optionen `aus` und `an` heißen; sie waren nur unübersetzt.
//
// # Gezählt wird der Schlüssel und nicht die Aufschrift
//
// **Die Gegenprobe hat den ersten Anlauf widerlegt.** Er zählte die
// Aufschriften, und gegen den alten Stand wurde er **nicht** rot: die zwei
// Felder trugen zwei **verschiedene** Namen — `dienst.scryfall` und
// *Scryfall*. Doppelt war die Einstellung, nicht ihr Name.
//
// Deshalb trägt jedes Bedienelement in [Wahl] seinen Schlüssel als
// `data-einstellung`, und dieser Lauf zählt ihn. Er prüft damit die Klasse
// und nicht die fünf Fälle: jede Einstellung, die die Seite bekommt, steht
// **genau einmal** da.
//
// # Und keine Aufschrift zweimal
//
// **Zwei verschiedene Felder mit derselben Aufschrift** — in *Verbindungen*
// standen zwei Paare „Adresse"/„Port", zwei Felder je Paar und nicht
// auseinanderzuhalten, wer Feld für Feld tabbte
// (#1767).
// Der letzte Lauf unten zählt dafür **jedes** Eingabefeld der Gruppe, nicht
// nur die mit Schlüssel.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import {
  alleDienste,
  eintraege,
  zeigeEinstellungen as zeige,
} from '../test/einstellungsseite'
import { einstellungsgruppen } from './einstellungsgruppen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

// **Jedes Bauteil, das sich selbst zeichnet, kommt hier vor** — sonst
// prüfte dieser Lauf die vier Kartendienste und nicht die Klasse. Vier
// Wege gibt es: der `Dienstschalter` (zweimal, in *Dienste* und in *KI*),
// die `Dienstfragen` und die `KILaufzeit`.
const sprachmodell: EinstellungDTO[] = [
  {
    schluessel: 'dienst.anthropic',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
  {
    schluessel: 'dienst.anthropic.adresse',
    wert: '',
    wirksam: '',
    werte: [],
    bereich: 'dienste',
  },
]

const kizeilen: EinstellungDTO[] = [
  {
    schluessel: 'ki.laufzeit',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
  {
    schluessel: 'ki.adresse',
    wert: '',
    wirksam: '',
    werte: [],
    bereich: 'dienste',
  },
]

const fragen: EinstellungDTO[] = [
  {
    schluessel: 'frage.suche',
    wert: 'scryfall',
    wirksam: 'scryfall',
    werte: ['scryfall', 'mtgapi'],
    bereich: 'dienste',
  },
]

// **Und einer, den dieses Fenster nicht kennt.** Er kann entstehen, wenn Go
// einen Dienst mehr führt als der Katalog hier — `gruppeFuer` gibt ihm dann
// die erste Gruppe. **Die Gegenprobe hat gezeigt, dass er hierher gehört**:
// ein Schnitt `=== 'dienste'` statt `!== 'ki'` ließ ihn aus beiden Listen
// fallen, und ohne diese Zeile blieb der Lauf dabei grün.
const unbekannt: EinstellungDTO[] = [
  {
    schluessel: 'dienst.neuer',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
]

const dienste = [...alleDienste, ...sprachmodell, ...unbekannt]

// schluesselDerGruppe gibt jeden Schlüssel, den die offene Gruppe zeigt.
function schluesselDerGruppe(): string[] {
  const feld = screen.getByRole('tabpanel')

  return [...feld.querySelectorAll('[data-einstellung]')].map(
    (eines) => eines.getAttribute('data-einstellung') ?? '',
  )
}

// aufschriften gibt zu jedem dieser Felder seinen zugänglichen Namen.
//
// **Den zugänglichen Namen und nicht das Attribut**: ein Feld trägt ihn über
// `aria-label` **oder** über ein `<label for>`, und beide Wege kommen auf
// dieser Seite vor. Wer nur das Attribut läse, hielte jedes Feld ohne eines
// für unbeschriftet.
function aufschriften(): string[] {
  const feld = screen.getByRole('tabpanel')

  return [...feld.querySelectorAll('[data-einstellung]')].map((eines) => {
    const beschriftung = (eines as HTMLInputElement).labels?.item(0)

    return (
      eines.getAttribute('aria-label') ??
      beschriftung?.textContent.trim() ??
      eines.id
    )
  })
}

// feldnamen gibt den zugänglichen Namen **jedes** Eingabefelds der offenen
// Gruppe — über `aria-labelledby`, `aria-label` oder ein `<label>`.
function feldnamen(): string[] {
  const feld = screen.getByRole('tabpanel')
  const rollen = ['textbox', 'spinbutton', 'combobox', 'checkbox', 'switch']

  return rollen.flatMap((rolle) =>
    within(feld).queryAllByRole(rolle).map(feldname),
  )
}

function feldname(eines: HTMLElement): string {
  const bezug = eines.getAttribute('aria-labelledby')
  if (bezug !== null) {
    return bezug
      .split(' ')
      .map((kennung) => document.getElementById(kennung)?.textContent ?? '')
      .join(' ')
  }

  const beschriftung = (eines as HTMLInputElement).labels?.[0]

  return (
    eines.getAttribute('aria-label') ?? beschriftung?.textContent.trim() ?? ''
  )
}

// jedeGruppe schlägt der Reihe nach jede Gruppe auf.
async function jedeGruppe(
  pruefe: (gruppe: string) => void,
  womit: EinstellungDTO[] = dienste,
) {
  zeige({ eintraege, dienste: womit, fragen, ki: kizeilen })

  for (const gruppe of einstellungsgruppen) {
    await userEvent.click(
      screen.getByRole('tab', { name: eigene.gruppen[gruppe] }),
    )

    // **Eingeklappt heißt erreichbar, nicht fort** — die Abschnitte gehen
    // auf, bevor gezählt wird (UXR-19, UXR-21: die Adressen der KI).
    for (const name of Object.values(eigene.abschnitte)) {
      const kasten = within(screen.getByRole('tabpanel')).queryByRole(
        'button',
        { name, expanded: false },
      )
      if (kasten !== null) {
        await userEvent.click(kasten)
      }
    }

    pruefe(eigene.gruppen[gruppe])
  }
}

describe('die Einstellungsseite zeigt jede Einstellung einmal', () => {
  afterEach(() => {
    raeumeApi()
  })

  // **Je Gruppe und nicht über die ganze Seite.** Eine Gruppe ist zugleich
  // ein Blatt: was in einer anderen steht, sieht niemand daneben.
  it('zeichnet keine Einstellung doppelt', async () => {
    await jedeGruppe((gruppe) => {
      const gesehen = schluesselDerGruppe()

      const doppelte = gesehen.filter(
        (eines, stelle) => gesehen.indexOf(eines) !== stelle,
      )

      expect(doppelte, `in der Gruppe „${gruppe}"`).toEqual([])
    })
  })

  // **Und keine geht verloren.** Eine Gruppe, deren Feld leer bleibt, wäre
  // still grün — dann prüfte der Lauf darüber nichts und sagte es nicht
  // (§10). Gezählt wird deshalb über alle neun: **so viele Bedienelemente,
  // wie die Seite Einstellungen bekommen hat.**
  it('zeigt jede Einstellung, die sie bekommt', async () => {
    let gesehen = 0

    await jedeGruppe(() => {
      gesehen += schluesselDerGruppe().length
    })

    expect(gesehen).toBe(
      eintraege.length + dienste.length + fragen.length + kizeilen.length,
    )
  })

  // **Und keine Aufschrift ist ein roher Schlüssel.** Ein Feld, das
  // `dienst.scryfall` heißt, hat seinen Namen nicht gefunden — das war die
  // sichtbare Hälfte desselben Befunds.
  //
  // **Ohne den unbekannten Dienst**, und das ist keine Ausnahme, sondern
  // die Frage: er trägt seinen Schlüssel **mit Absicht**, weil das Fenster
  // keinen Namen für ihn hat
  // (#1540).
  // Der Befund war ein Dienst, den das Fenster kennt und trotzdem nicht
  // benennt.
  it('zeigt keine Aufschrift, die ein Schlüssel ist', async () => {
    await jedeGruppe(
      (gruppe) => {
        for (const aufschrift of aufschriften()) {
          expect(aufschrift, `in der Gruppe „${gruppe}"`).not.toMatch(
            /^(dienst|ki|netz|vorgabe|richtlinie|anbieter|frage)\./,
          )
        }
      },
      [...alleDienste, ...sprachmodell],
    )
  })

  // **Keine Aufschrift zweimal in einer Gruppe** (#1767) — über jedes
  // Eingabefeld, auch die ohne Einstellungsschlüssel: die Adresse und der Port
  // eines Dienstes sind keine Einstellung und standen doch doppelt da.
  it('trägt keine Aufschrift zweimal', async () => {
    // **Die zwei Dienste stehen still** — dann zeigen sie Adresse und Port.
    const steht = (art: 'rest' | 'mcp') =>
      Promise.resolve({
        stand: {
          art,
          laeuft: false,
          adresse: '',
          tls: false,
          marke: '',
          namen: [],
          adressen: [],
          fingerabdruck: '',
          zeugnisAussteller: '',
          zeugnisBis: '',
          lauscher: [],
          fehlende: [],
        },
        fehler: null,
      })
    belegeApi({ RESTStand: () => steht('rest'), MCPStand: () => steht('mcp') })

    const ueberall: string[] = []
    zeige({ eintraege, dienste, fragen, ki: kizeilen })

    for (const gruppe of einstellungsgruppen) {
      await userEvent.click(
        screen.getByRole('tab', { name: eigene.gruppen[gruppe] }),
      )

      // **Auch was Fachleute einstellen** — dort stehen die zwei Dienste.
      const fachleute = within(screen.getByRole('tabpanel')).queryByRole(
        'button',
        { name: eigene.fachleute },
      )
      if (fachleute !== null) {
        await userEvent.click(fachleute)
      }
      await screen.findAllByRole('tab')

      const namen = feldnamen()
      ueberall.push(...namen)
      const doppelte = namen.filter(
        (eines, stelle) => namen.indexOf(eines) !== stelle,
      )

      expect(doppelte, `in der Gruppe „${eigene.gruppen[gruppe]}"`).toEqual([])
    }

    // **Und jemand hört zu** (§3): die vier Felder der zwei Dienste standen
    // wirklich da — sonst wäre das Nein oben keines.
    expect(ueberall.filter((name) => name.includes('Port'))).toHaveLength(2)
  })
})
