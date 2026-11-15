// **Der Katalog der Einstellungen, wie ihn `docs/EINSTELLUNGEN.md` führt** —
// UXR-F2 (#2008).
//
// # Warum er aus dem Dokument gelesen wird
//
// **Weil die Werte einer Einstellung in Go stehen** (`internal/domain`), und
// ein Lauf über „jede Einstellung" sie nicht aus einer Liste hier abschreiben
// soll — die wäre die ungeprüfte. Das Dokument ist die **erzeugte** Tafel:
// `internal/werkzeuge/einstellungsschema_test.go` vergleicht sie Zeile für
// Zeile mit `internal/domain`, und eine neue Einstellung ohne Zeile macht ihn
// rot. Wer hier liest, liest damit den Katalog.
//
// # Was er nicht ist
//
// **Kein Ersatz für die Antwort des Bestands**: `wirksam`, Herkunft und Ebene
// kommen aus der Auflösung zur Laufzeit. Die Vorrichtung füllt sie mit dem
// Wert, damit ein Eintrag vollständig ist; was sie bedeuten, prüfen die Läufe
// zur Herkunft.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import type { EinstellungDTO } from '../api'

// Katalogzeile ist eine Zeile der Tafel.
export interface Katalogzeile {
  schluessel: string

  // tafel sagt, in welcher der beiden Tafeln sie steht: der des Profils und
  // des Geräts oder der des Menschen.
  tafel: 'profil' | 'mensch'

  // art ist `liste` bei aufgezählten Werten, sonst `geprueft` oder `frei` —
  // dann gibt es keine Werteliste.
  art: 'liste' | 'geprueft' | 'frei'
  werte: string[]

  // vorgabe ist leer, wo das Dokument einen Strich schreibt.
  vorgabe: string
  bereich: string
  ebene: string
  wirkung: string
}

const datei = join(import.meta.dirname, '../../../docs/EINSTELLUNGEN.md')

// zellen zerlegt eine Tabellenzeile in ihre Zellen.
function zellen(zeile: string): string[] {
  return zeile
    .split('|')
    .slice(1, -1)
    .map((zelle) => zelle.trim())
}

// ohneCode nimmt die Rückstriche weg: `de` wird zu de.
function ohneCode(zelle: string): string {
  return zelle.replace(/^`|`$/g, '')
}

function werteVon(zelle: string): Pick<Katalogzeile, 'art' | 'werte'> {
  if (zelle === '*geprüft*') {
    return { art: 'geprueft', werte: [] }
  }

  if (zelle === '*frei*') {
    return { art: 'frei', werte: [] }
  }

  return {
    art: 'liste',
    werte: [...zelle.matchAll(/`([^`]+)`/g)].map((treffer) => treffer[1] ?? ''),
  }
}

// tafelNach gibt die Zeilen unter einer Überschrift, ohne den Kopf.
function tafelNach(dokument: string, ueberschrift: string): string[] {
  const rest = dokument.split(`\n${ueberschrift}\n`)[1] ?? ''
  const zeilen: string[] = []

  for (const zeile of rest.split('\n')) {
    if (zeile.startsWith('## ')) {
      break
    }

    if (zeile.startsWith('| `')) {
      zeilen.push(zeile)
    }
  }

  return zeilen
}

// einstellungskatalog liest beide Tafeln.
export function einstellungskatalog(): Katalogzeile[] {
  const dokument = readFileSync(datei, 'utf8')

  const profil = tafelNach(
    dokument,
    '## Die Tafel des Profils und des Geräts',
  ).map((zeile): Katalogzeile => {
    const [
      schluessel = '',
      bereich = '',
      ebene = '',
      werte = '',
      vorgabe = '',
      wirkung = '',
    ] = zellen(zeile)

    return {
      schluessel: ohneCode(schluessel),
      tafel: 'profil',
      ...werteVon(werte),
      vorgabe: vorgabe === '—' ? '' : ohneCode(vorgabe),
      bereich: ohneCode(bereich),
      ebene: ohneCode(ebene),
      wirkung,
    }
  })

  const mensch = tafelNach(dokument, '## Die Tafel des Menschen').map(
    (zeile): Katalogzeile => {
      const [schluessel = '', werte = '', vorgabe = ''] = zellen(zeile)

      return {
        schluessel: ohneCode(schluessel),
        tafel: 'mensch',
        ...werteVon(werte),
        vorgabe: vorgabe === '—' ? '' : ohneCode(vorgabe),
        bereich: 'anwendung',
        ebene: 'nutzer',
        wirkung: 'sofort',
      }
    },
  )

  return [...profil, ...mensch]
}

// alsEintrag macht aus einer Zeile, was die Seite von Go bekommt.
//
// **Der Wert ist die Vorgabe, sonst der erste der Liste**: ein frischer
// Bestand. Wer einen anderen braucht, überschreibt ihn.
export function alsEintrag(zeile: Katalogzeile, wert?: string): EinstellungDTO {
  const gelten =
    wert ?? (zeile.vorgabe !== '' ? zeile.vorgabe : (zeile.werte[0] ?? ''))

  return {
    schluessel: zeile.schluessel,
    wert: gelten,
    wirksam: gelten,
    werte: zeile.werte,
    bereich: zeile.bereich,
    ebene: zeile.ebene,
    herkunft: 'vorgabe',
    vorgabe: zeile.vorgabe,
    wirkung: zeile.wirkung,
  }
}

// profilEintraege sind die Einstellungen, die die Einstellungsseite zeigt —
// die Tafel des Profils und des Geräts.
export function profilEintraege(): EinstellungDTO[] {
  return einstellungskatalog()
    .filter((zeile) => zeile.tafel === 'profil')
    .map((zeile) => alsEintrag(zeile))
}

// Rolle ist, was eine Einstellung auf der Seite sein soll — so, wie ein
// Vorleseprogramm es nennt.
export type Rolle = 'textbox' | 'switch' | 'radiogroup' | 'combobox'

// anPaare sind die Wörter, mit denen eine Einstellung an und aus heißt.
const anPaare = [
  ['an', 'aus'],
  ['sichtbar', 'verborgen'],
]

// erwartetesElement ist die Regel R7 aus `docs/UIREVIEW.md`, in den Worten
// des Issues (#2008): *Schalter für zwei Werte, die an und aus sind, Segmente
// bis vier, Auswahl darüber.*
//
// **Sie steht hier ein zweites Mal und nicht als Aufruf von `wahlart.ts`**:
// ein Lauf, der die Regel aus dem Code holte, prüfte den Code gegen sich
// selbst. Wer sie ändert, ändert beide Stellen mit Absicht.
export function erwartetesElement(
  werte: readonly string[],
  schluessel?: string,
): Rolle {
  // **Die Zeitzone hat keine Werteliste und trotzdem eine Auswahl** — UXR-F1
  // (#2007): Go prüft sie, nennt sie aber nicht, und das Fenster führt die
  // Liste selbst (`Zeitzonenwahl`).
  if (werte.length === 0) {
    return schluessel === 'zeitzone' ? 'combobox' : 'textbox'
  }

  if (
    werte.length === 2 &&
    anPaare.some((paar) => paar.every((wort) => werte.includes(wort)))
  ) {
    return 'switch'
  }

  return werte.length >= 2 && werte.length <= 4 ? 'radiogroup' : 'combobox'
}

// elementVon liest, was ein Knoten mit `data-einstellung` ist.
export function elementVon(knoten: Element): Rolle {
  if (knoten.getAttribute('role') === 'switch') {
    return 'switch'
  }

  if (knoten.getAttribute('role') === 'radiogroup') {
    return 'radiogroup'
  }

  return knoten.tagName === 'SELECT' ||
    knoten.getAttribute('role') === 'combobox'
    ? 'combobox'
    : 'textbox'
}
