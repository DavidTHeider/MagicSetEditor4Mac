// **Die Ablage einer gezeichneten Zeichnung** —
// #890.
//
// **Eine eigene Datei**, weil `Symbol.tsx` über die Zeilengrenze wuchs (§2).
// Der Schnitt folgt der Frage: dort, **wie die Seite aussieht**, hier, **was
// beim Ablegen geschieht**.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import type { SymbolteilDTO } from '../antworten'

// abgelegteDatei schreibt die Zeichnung und gibt ihren Dateinamen.
//
// **`unter` nennt die Datei, die überschrieben werden soll.** Leer heißt: eine
// neue anlegen.
export async function abgelegteDatei(
  setID: string,
  teile: SymbolteilDTO[],
  unter: string,
): Promise<string> {
  const antwort = await api().SymbolSetzen(setID, teile, unter)

  return auspacken(antwort, () => antwort.name)
}

// Symbolziel sagt, **welches** Feld die Zeichnung trägt.
//
// **Ohne Ziel das Setsymbol** — so war es, seit es den Editor gibt. Mit Ziel
// das Symbolfeld einer Karte
// (#1339).
export interface Symbolziel {
  karteID: string
  feld: string
}

// Symbolzielstand ist ein Ziel **samt dem Set, in dem es gilt**.
//
// **Das Set gehört dazu**, seit Befund 3 des Zweitblicks über Phase 4.11:
// geräumt wurde das Ziel nur beim Aufschlagen einer anderen Karte, und die
// vier Setwahlwege taten es nicht. Danach zeigte es auf eine Karte des
// **vorigen** Sets — beim Speichern lag die Symboldatei im neuen Set, und das
// Feldschreiben wurde zu Recht abgewiesen. Ein halber Stand (§7, Regel 1).
export interface Symbolzielstand extends Symbolziel {
  setID: string
}

// geraeumtesZiel gibt das Ziel, das **behalten** werden darf.
//
// **Verbergen ist nicht Räumen** (§7, Regel 1). [zielImSet] blendet ein
// fremdes Ziel aus; der Zustand blieb aber stehen, und wer Set A verließ und
// später wiederkam, fand es vor. Nach einem Zurückspielen des Bestands kann
// dieselbe Kennung inzwischen zu einer **anderen** Karte gehören — angestrichen
// vom Nachlauf über Phase 4.11.
//
// **Warum beide nebeneinander stehen**: das eine beantwortet „was zeige ich
// jetzt?", das andere „was behalte ich?". Sie sehen gleich aus und werden an
// verschiedenen Stellen gebraucht — beim Zeichnen und beim Wechsel.
export function geraeumtesZiel(
  stand: Symbolzielstand | undefined,
  setID: string | undefined,
): Symbolzielstand | undefined {
  if (stand === undefined || setID === undefined || stand.setID !== setID) {
    return undefined
  }

  return stand
}

// zielImSet gibt das Ziel, **wenn es zu diesem Set gehört**.
//
// **Eine Ableitung und kein Effekt.** Ein Effekt liefe nach dem Zeichnen: für
// einen Anstrich stünde das alte Ziel neben der neuen Set-ID, und ein
// Speichern in dieser Zeit träfe die falsche Karte. Wer den Stand ableitet,
// hat den Fehler nicht.
export function zielImSet(
  stand: Symbolzielstand | undefined,
  setID: string | undefined,
): Symbolziel | undefined {
  if (stand === undefined || setID === undefined || stand.setID !== setID) {
    return undefined
  }

  return { karteID: stand.karteID, feld: stand.feld }
}

// abgelegtesSymbol schreibt die Zeichnung und trägt sie in ihr Feld ein.
//
// **Nacheinander und nicht nebeneinander.** Der Feldwert zeigt auf die Datei;
// wer ihn setzt, bevor sie liegt, hat einen Verweis ins Leere. Und scheitert
// das Ablegen, wird gar nicht erst gesetzt (§7, Regel 3).
//
// **Der Name wird gemeldet, sobald er dasteht.** Scheitert das Feldschreiben,
// ist die Datei schon geschrieben; ohne den gemerkten Namen legte der nächste
// Versuch eine zweite an, und die erste bliebe als verwaistes Beiwerk liegen
// (#890, Befund 1 des Nachlaufs).
//
// **Ein leeres Blatt räumt den Feldwert.** Wer alle Teile löscht, will kein
// Symbol mehr; eine leere Datei zu schreiben hieße, dieselbe Verwaisung mit
// anderem Anlass anzulegen (Befund 3).
export async function abgelegtesSymbol(
  setID: string,
  teile: SymbolteilDTO[],
  unter: string,
  melde: (name: string) => void,
  ziel?: Symbolziel,
): Promise<string> {
  const dateiname =
    teile.length === 0 ? '' : await abgelegteDatei(setID, teile, unter)

  melde(dateiname)

  // **Dasselbe Nacheinander für beide Ziele.** Die Datei liegt, dann zeigt
  // ein Feldwert auf sie — welches Feld es ist, ändert daran nichts.
  const gesetzt =
    ziel === undefined
      ? await api().SetwertSetzen(setID, 'symbol', dateiname, null)
      : await api().FeldSetzen(setID, ziel.karteID, ziel.feld, dateiname, null)

  auspacken(gesetzt, () => undefined)

  return dateiname
}

// setsymbolname gibt den Dateinamen aus `set.symbol` — leer, wenn keiner steht.
export async function setsymbolname(setID: string): Promise<string> {
  const angaben = await api().Setangaben(setID)
  const werte = auspacken(angaben, () => angaben.werte)

  return werte.find((wert) => wert.feld === 'symbol')?.wert ?? ''
}

// kartensymbolname gibt den Dateinamen aus einem Symbolfeld einer Karte.
//
// **Über die Karte und nicht über die Setangaben** — sonst stünde hier der
// Wert des Sets unter dem Namen des Kartenfeldes
// (#1339).
export async function kartensymbolname(
  karteID: string,
  feld: string,
): Promise<string> {
  const antwort = await api().Karte(karteID)
  const karte = auspacken(antwort, () => antwort.karte)

  return karte.werte.find((wert) => wert.feld === feld)?.wert ?? ''
}

// Bibliotheksbearbeitung sagt, **welches Symbol der Bibliothek** der Editor
// bearbeitet (DOM-02.AK-02,
// #1697).
export interface Bibliotheksbearbeitung {
  id: string
  name: string

  // gesichert meldet die neue Fassung samt den Sets, die die vorige tragen.
  gesichert: (veraltet: string[]) => void
}

// vorhandeneZeichnung holt, was der Editor aufschlägt — das Symbol der
// Bibliothek, das Setsymbol oder das Symbolfeld einer Karte.
//
// **Aus `Symbol.tsx` herausgelöst**, als die Bibliothek als drittes Ziel dazu
// kam und die Datei über die Zeilengrenze aus §2 gegangen wäre.
export async function vorhandeneZeichnung(
  setID: string,
  ziel: Symbolziel | undefined,
  bibliothek: string | undefined,
): Promise<{ teile: SymbolteilDTO[]; name: string }> {
  if (bibliothek !== undefined) {
    const antwort = await api().BibliothekssymbolTeile(bibliothek)

    return auspacken(antwort, () => ({
      teile: antwort.teile,
      name: antwort.name,
    }))
  }

  // **Woher der Dateiname kommt, hängt am Ziel** (#1339): ohne Ziel aus den
  // Setangaben, mit Ziel aus dem Feldwert der Karte. Beides ist derselbe
  // Griff — ein Feldwert, der auf eine Datei zeigt.
  //
  // **Zwei Anweisungen statt eines Ternärs mit `await`.** V8 rechnet den
  // stillen `else`-Zweig eines `if` aus der Differenz zweier Zähler; steht
  // davor ein `await` in einem Ternär, kommt dabei eine **negative** Zahl
  // heraus (gemessen am 2026-08-31: `[38, -34]`), und der Bericht meldet
  // einen unbetretenen Zweig, den jeder Lauf betritt. Die Form hier ist
  // dieselbe Logik ohne diesen Fallstrick.
  let dateiname = ''

  if (ziel === undefined) {
    dateiname = await setsymbolname(setID)
  } else {
    dateiname = await kartensymbolname(ziel.karteID, ziel.feld)
  }

  if (dateiname === '') {
    return { teile: [], name: '' }
  }

  const antwort = await api().Symbol(setID, dateiname)

  return auspacken(antwort, () => ({
    teile: antwort.teile,
    name: antwort.name,
  }))
}

// inBibliothekGesichert legt eine neue Fassung ab und meldet die Sets, die
// die vorige tragen.
export async function inBibliothekGesichert(
  bibliothek: Bibliotheksbearbeitung,
  teile: SymbolteilDTO[],
): Promise<string> {
  const antwort = await api().BibliothekssymbolSichern(bibliothek.id, teile)

  bibliothek.gesichert(auspacken(antwort, () => antwort.veraltet))

  return bibliothek.name
}
