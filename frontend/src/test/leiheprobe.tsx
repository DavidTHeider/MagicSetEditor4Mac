// **Die Vorrichtung der Leihläufe** — Werkstück G6 von Phase 5.5
// (#1463).
//
// **Eine eigene Datei**, weil zwei Läufe sie brauchen und `deckleihe.test.tsx`
// mit dem zweiten über die Zeilengrenze aus §2 ginge (#1960).
import { render, screen, within } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'

import { Deckleihe } from '../seiten/Deckleihe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'
import { einDeck } from '../seiten/partieprobe'
import { einKontakt, eineListe } from '../seiten/peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import type { Leihe } from '../apileihe'

export const eigene = kataloge.de.seiten.spielen

// eineLeihe gibt eine Leihe mit den genannten Abweichungen.
export function eineLeihe(abweichend: Partial<Leihe> = {}): Leihe {
  return {
    kennung: 'l1',
    verleiher: '12D3KooWAnna',
    entleiher: '12D3KooWBert',
    abdruck: 'abdruckdesabzugs',
    bindungsart: 'partie',
    woran: 'p1',
    rechte: ['spielen'],
    ausgestellt: '2026-09-17T12:00:00Z',
    bis: '2026-09-18T12:00:00Z',
    modus: 'zuende',
    verliehen: true,
    verbraucht: '',
    widerrufen: false,
    karten: 40,
    ...abweichend,
  }
}

// leer ist, was die Naht ohne Leihe schickt.
export const leer = { leihen: [], pfad: '', fehler: null }

// gewaehlt ist die Antwort eines Dateidialogs, in dem jemand einen Ort
// bestätigt hat (UXR-F15, #2030).
export function gewaehlt(pfad: string) {
  return { pfad, abgebrochen: false, fehler: null }
}

// weggeklickt ist die Antwort eines Dateidialogs, der ohne Wahl zuging — ein
// Abbruch und kein Fehler.
export const weggeklickt = { pfad: '', abgebrochen: true, fehler: null }

// umschlagGewaehlt ist dieselbe Antwort von `UmschlagWaehlen`: mit der Gestalt
// eines angesehenen Umschlags, die das Annehmen nicht braucht.
export function umschlagGewaehlt(pfad: string) {
  return { ...gewaehlt(pfad), objekte: [], urheber: '', fingerabdruck: '' }
}

// zielgruppe und umschlaggruppe sind die beiden Wahlen — die Gruppe trägt
// den Namen, der früher an einem Textfeld hing.
export const zielgruppe = () =>
  screen.getByRole('group', { name: eigene.leiheZiel })

export const umschlaggruppe = () =>
  screen.getByRole('group', { name: eigene.leiheUmschlag })

// waehleZiel öffnet den Dialog fürs Ziel und wartet, bis der Pfad
// **dasteht** — auf die Wirkung und nicht auf den Ruf (§3).
export async function waehleZiel(
  hand: ReturnType<typeof nutzerereignis.setup>,
  pfad: string,
): Promise<void> {
  // **Der einzige Knopf der Gruppe**: seine Aufschrift wechselt mit der Wahl.
  await hand.click(within(zielgruppe()).getByRole('button'))
  await within(zielgruppe()).findByText(pfad)
}

// waehleUmschlag dasselbe für den Umschlag, der angenommen wird. **Die Gruppe
// steht erst, wenn die Annahme aufgeklappt ist** (`oeffneAnnahme`).
export async function waehleUmschlag(
  hand: ReturnType<typeof nutzerereignis.setup>,
  pfad: string,
): Promise<void> {
  await hand.click(within(umschlaggruppe()).getByRole('button'))
  await within(umschlaggruppe()).findByText(pfad)
}

// leiheApi belegt die Fassade samt Kontakten und Decks.
//
// **Die Seite lädt beide bei jedem Aufstellen** (UX-08.AK-04, #1680): ohne
// Beleg stünde neben dem, worum es in einem Lauf geht, ein Ladefehler.
export function leiheApi(teile: Partial<Api>): Api {
  return belegeApi({
    Peerkontakte: () =>
      Promise.resolve(
        eineListe([
          einKontakt({
            kennung: '12D3KooWBert',
            eigenerName: 'Bert',
            stand: 'bestaetigt',
          }),
        ]),
      ),
    Decks: () =>
      Promise.resolve({ decks: [einDeck('7', 'Mono Rot')], fehler: null }),
    ...teile,
  })
}

// waehleBertUndDeck wählt, statt zu tippen, was früher getippt wurde.
export async function waehleBertUndDeck(
  hand: ReturnType<typeof nutzerereignis.setup>,
): Promise<void> {
  await hand.click(await screen.findByRole('radio', { name: /^Bert · / }))
  await hand.selectOptions(
    screen.getByLabelText(eigene.leiheDeck),
    await screen.findByRole('option', { name: 'Mono Rot' }),
  )
}

// oeffneAnnahme klappt „Eine Leihe annehmen" auf — seit UXR-13 (#1921)
// steht das Feld für den Umschlag eingeklappt unter dem Verleihen: offen ist
// eine Gruppe, nicht drei (R3).
export async function oeffneAnnahme(
  hand: ReturnType<typeof nutzerereignis.setup>,
): Promise<void> {
  await hand.click(
    screen.getByRole('button', { name: eigene.leiheAnnehmenTitel }),
  )
}

// zeige zeichnet die Seite und gibt die Hand.
export function zeige(geaendert?: () => void) {
  render(
    <Sprachraum sprache="de">
      <Deckleihe geaendert={geaendert} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}
