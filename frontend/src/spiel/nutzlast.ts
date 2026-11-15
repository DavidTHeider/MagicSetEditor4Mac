// **Aus eingesammelten Werten wird eine Nutzlast** —
// #1623.
//
// # Warum das eine eigene Datei ist
//
// **Weil es die Stelle ist, an der die Gestalt zählt und nicht das
// Aussehen.** Ein Eingabefeld gibt eine Zeichenkette; der Kern liest `menge`
// als Zahl und `ziele` als Liste. Wer beides im Bauteil mischte, hätte die
// Umrechnung im Baum — und eine `"1"` statt einer `1` fällt erst an der
// Naht auf, wo die Meldung von etwas anderem spricht (§3: die rohe Form
// gehört in den Lauf).

import { alsText } from './feldwerte'

import type { Befehlsform } from './befehlsformen'
import type { Partie, Spielobjekt } from '../apipartie'

// Werte sind, was die Form eingesammelt hat — je Feldname eine Zeichenkette.
//
// **Eine Liste für die, die mehrere tragen** (Ziele, Angriffe, Blöcke).
export type Werte = Record<string, string | string[]>

// zahlfelder sind die Schlüssel, die der Kern als Zahl liest.
//
// **Aus der Form und nicht aus einer zweiten Liste**: `art: 'zahl'` steht
// schon da, und zwei Listen liefen auseinander.
function istZahl(form: Befehlsform, name: string): boolean {
  return form.felder.some((feld) => feld.name === name && feld.art === 'zahl')
}

// sichtbareFelder gibt die Felder, die eine Form gerade wirklich fragt.
//
// **Eine Stelle gibt es nur in einer geordneten Zone** (Befund F04,
// #1748).
// Welche Felder das sind, hängt an den **wirksamen** Werten und nicht an den
// eingegebenen: wer nichts angefasst hat, hat trotzdem eine Zone — die
// Vorgabe. Ein `werte.zone ?? 'hand'` schriebe die Vorgabe ein zweites Mal
// hin; läuft sie auseinander, fragt das Formular etwas anderes, als es
// schickt.
//
// **Zwei Verbraucher, eine Antwort**: das Bauteil zeichnet danach, und
// `nutzlastAus` schickt danach. Zwei Entscheidungen nebeneinander wären zwei
// Wahrheiten über denselben Befehl.
export function sichtbareFelder(
  form: Befehlsform,
  werte: Werte,
): Befehlsform['felder'] {
  const wirksam = new Map<string, string>()

  for (const feld of form.felder) {
    wirksam.set(feld.name, alsText(werte[feld.name] ?? feld.vorgabe ?? ''))
  }

  // **Ein unbekannter Name gibt die leere Zeichenkette**, und das ist kein
  // toter Zweig: ein Prädikat darf nach einem Feld fragen, das diese Form
  // nicht hat — die Antwort ist dann „nichts gewählt".
  const nachschlag = (name: string) => wirksam.get(name) ?? ''

  return form.felder.filter(
    (feld) => feld.nurWenn === undefined || feld.nurWenn(nachschlag),
  )
}

// startwert ist, was ein Feld zeigt und schickt, solange niemand gewählt hat.
//
// **Ein Sitz beginnt beim eigenen** — P3 (#2371): ein Auswahlfeld ohne
// leeren Eintrag zeigt den ersten Sitz, und ohne diese Vorgabe schickte die
// Form eine leere Angabe, während der erste Sitz dastand.
export function startwert(
  feld: Befehlsform['felder'][number],
  partie: Partie,
): string {
  return feld.vorgabe ?? (feld.art === 'sitz' ? partie.fuer : '')
}

// nutzlastAus baut die Nutzlast eines Befehls.
//
// **Ein leeres Feld fällt heraus** und wird nicht als `""` geschickt: der
// Kern liest „nicht gesagt" und „leer gesagt" verschieden — `ruf` ohne Ruf
// ist ein gültiger Münzwurf, `ruf: ""` wäre einer mit einem Ruf, den
// niemand kennt.
export function nutzlastAus(
  form: Befehlsform,
  werte: Werte,
  objekt: Spielobjekt | undefined,
  partie: Partie,
): Record<string, unknown> {
  const nutzlast: Record<string, unknown> = {}

  if (form.gruppe === 'karte' && objekt !== undefined) {
    nutzlast.objekt = objekt.kennung
  }

  // **Ein Feld, das die Form nicht fragt, geht nicht hinaus** (Befund F04).
  // Sonst schickte die Vorgabe weiter eine Stelle, die niemand sieht — und
  // das ist genau der Fall aus S05.NAT-02: wer von der Bibliothek auf die
  // Hand umstellt, nimmt die vorher gewählte Stelle damit aus dem Auftrag.
  for (const feld of sichtbareFelder(form, werte)) {
    const wert = werte[feld.name] ?? startwert(feld, partie)

    // **Ohne Kennung** (#2480): sie vergibt der Kern. Das Fenster sieht nur
    // Sichtkennungen und kann nicht wissen, welche Kennung frei ist.
    if (feld.art === 'spielstein') {
      nutzlast.felder = spielsteinfelder(
        Array.isArray(wert) ? wert.join('|') : wert,
      )

      continue
    }

    if (Array.isArray(wert)) {
      nutzlast[feld.name] = paareOder(feld.art, wert)

      continue
    }

    if (wert === '') {
      continue
    }

    nutzlast[feld.name] = istZahl(form, feld.name) ? Number(wert) : wert
  }

  return nutzlast
}

// paareOder macht aus einer Liste von Kennungen das, was der Kern erwartet.
//
// **Angriffe und Blöcke sind Paare** (CR 508.1a, CR 509.1a), und die zweite
// Hälfte steht in der Kennung: `kreatur|ziel`. Die Form setzt sie zusammen,
// weil ein `<select>` je Zeile genau einen Wert gibt.
function paareOder(art: string, wert: string[]): unknown {
  if (art !== 'angriffe' && art !== 'bloecke') {
    return wert
  }

  const zweites = art === 'angriffe' ? 'ziel' : 'angreifer'

  return wert.map((zeile) => {
    // **An der ersten Stelle geteilt und nicht überall**: eine
    // Objektkennung sieht aus wie `sitz1/k1/0`, und niemand verbietet ihr
    // ein Trennzeichen. Ein `split('|')` warf die zweite Hälfte weg.
    const trenner = zeile.indexOf('|')

    if (trenner < 0) {
      return { kreatur: zeile, [zweites]: '' }
    }

    return {
      kreatur: zeile.slice(0, trenner),
      [zweites]: zeile.slice(trenner + 1),
    }
  })
}

// spielsteinfelder macht aus vier Angaben die Felder eines MSE-Pakets.
//
// **Dieselben Feldnamen wie in einer Vorlage** — der Grund steht bei
// `domain.Spielobjekt.Felder`: eine zweite Deutung derselben Felder liefe
// beim ersten neuen auseinander.
//
// **Leere fallen heraus**, damit ein Spielstein ohne Werte keine leere
// Stärke bekommt (CR 208.3: wo keine steht, steht keine).
function spielsteinfelder(wert: string): Record<string, string> {
  const [name, typ, staerke, widerstand] = wert.split('|')

  const felder: Record<string, string> = {}

  for (const [schluessel, eines] of [
    ['name', name],
    ['super type', typ],
    ['power', staerke],
    ['toughness', widerstand],
  ] as const) {
    if (eines !== undefined && eines !== '') {
      felder[schluessel] = eines
    }
  }

  return felder
}
