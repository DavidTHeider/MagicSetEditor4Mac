// Der Paketbaum — MSE2s `PackageUpdateList::TreeItem`, portiert (#244).
//
// # Wofür
//
// Die Paketseite zeigte 329 Zeilen in einer Tabelle: ungefiltert, ungruppiert,
// ohne Suche. Wer ein bestimmtes Paket suchte, scrollte.
//
// **Die Gruppierung steht in den Daten und musste nicht erfunden werden.**
// Jedes Paket kann `installer group:` und `position hint:` nennen; über
// `full-magic-pack` tun es 416 von 526 beziehungsweise 495. MSE2 baut daraus
// einen Baum: der Pfad wird an `/` zerlegt, jede Stufe wird ein Knoten.
//
// `src/gui/package_update_list.cpp`, Zeilen 18–75 und 200–212, Fassung f9d9356
// vom 2021-07-06.
//
// # Warum hier und nicht in Go
//
// Weil gesucht wird. `App.Pakete` frischt den Katalog auf — es **liest jedes
// Paket auf der Platte**, bevor es antwortet; das bei jedem Tastendruck zu tun
// wäre eine Plattenrunde je Buchstabe. Go liefert Gruppe, Rang und Artrang, das
// Fenster baut daraus so oft einen Baum, wie es will.
//
// Die portierte **Reihenfolge der Arten** steht deswegen trotzdem in Go
// (`domain.Paket.Artrang`): sie ist eine Aussage über MSE-Pakete, kein
// Anzeigedetail.
import type { PaketDTO } from '../api'
import { vergleicher } from '../sprache/vergleich'

// Baumzeile ist eine Zeile der flachen Liste, die `toItems` erzeugt.
//
// **Flach und nicht verschachtelt** — das ist MSEs Form: `toItems` (Zeilen
// 69–75) sortiert die Kinder und schiebt sie samt Enkeln in einen Vektor. Eine
// Baumliste zeichnet Zeilen mit Einzug; sie braucht keine verschachtelte
// Struktur, und eine flache Liste lässt sich mit den Pfeiltasten durchlaufen,
// ohne dass jemand sie erst wieder plattmacht.
export interface Baumzeile {
  // schluessel ist über die ganze Liste eindeutig — der Gruppenpfad.
  schluessel: string

  // aufschrift ist, was in der Zeile steht: der Name der Stufe oder des Pakets.
  aufschrift: string

  // stufe ist die Einrücktiefe, von 0 an.
  stufe: number

  // paket steht nur an einer Paketzeile; ein Gruppenknoten hat keines.
  paket?: PaketDTO
}

// knoten ist ein Baumknoten während des Baus.
interface Knoten {
  aufschrift: string
  schluessel: string
  kinder: Knoten[]
  paket?: PaketDTO

  // artrang und rang sind die **kleinste** Einordnung unter diesem Knoten —
  // Zeilen 20–27, wo `add` die Werte des Pakets nach oben durchreicht, sofern
  // sie kleiner sind. Ohne diese Übernahme stünde jede Gruppe hinten, weil ein
  // frischer Knoten mit `TYPE_OTHER` und einer Million anfängt.
  artrang: number
  rang: number
}

// ArtrangUebrige ist `TYPE_OTHER` — der Anfangswert eines frischen Knotens.
const ArtrangUebrige = 100

// RangOhneHinweis ist `position_hint(1000000)` aus Zeile 52 derselben Fassung.
const RangOhneHinweis = 1000000

// paketbaum ordnet die Pakete zu einer flachen Liste mit Stufen.
//
// `ohneGruppe` ist die Aufschrift des Sammelknotens für Pakete, die keine
// `installer group:` nennen:
//
//	String group = ip->description->installer_group;
//	if (group.empty()) group = _("custom");
//
// Zeilen 204–205. **Sie kommt von außen**, weil sie ein Oberflächentext ist
// und in die Sprachkataloge gehört (§9) — `_(S)` ist bei MSE kein
// Übersetzungsaufruf, sondern ein Makro für Unicode-Literale
// (`src/util/string.hpp`, Zeile 58), und dort steht wörtlich „Custom".
//
// **Sie ist zugleich der Schlüssel des Knotens.** Ein Paket, dessen Gruppe
// genau so heißt, landete darin — MSE hat dieselbe Eigenschaft, und der Preis
// dafür ist kleiner als ein Sonderweg durch den ganzen Baum.
export function paketbaum(
  pakete: readonly PaketDTO[],
  ohneGruppe: string,
  sprache: string,
): Baumzeile[] {
  const wurzel = neuerKnoten('', '')

  for (const paket of pakete) {
    const weg = zerlegen(paket.gruppe)

    einfuegen(wurzel, paket, weg.length === 0 ? [ohneGruppe] : weg)
  }

  const zeilen: Baumzeile[] = []
  ausgeben(wurzel, zeilen, 0, vergleicheIn(sprache))

  return zeilen
}

// zerlegen macht aus `magic/m15 style` die Stufen `['magic', 'm15 style']`.
//
// **Leere Stufen fallen weg.** Eine Gruppe `magic//stile` — geschrieben,
// gemeint war eine — legte sonst einen namenlosen Knoten an, den niemand
// anklicken kann und der in der Liste als leere Zeile steht.
function zerlegen(gruppe: string): string[] {
  return gruppe
    .split('/')
    .map((stufe) => stufe.trim())
    .filter((stufe) => stufe !== '')
}

function neuerKnoten(aufschrift: string, schluessel: string): Knoten {
  return {
    aufschrift,
    schluessel,
    kinder: [],
    artrang: ArtrangUebrige,
    rang: RangOhneHinweis,
  }
}

// einfuegen hängt ein Paket unter seinen Gruppenpfad — `TreeItem::add`,
// Zeilen 18–59.
function einfuegen(knoten: Knoten, paket: PaketDTO, weg: readonly string[]) {
  // Die Einordnung wandert nach oben, bevor es weitergeht (Zeilen 20–27).
  if (
    paket.artrang < knoten.artrang ||
    (paket.artrang === knoten.artrang && paket.rang < knoten.rang)
  ) {
    knoten.artrang = paket.artrang
    knoten.rang = paket.rang
  }

  // **Die Zerlegung entscheidet, nicht die Länge.** Der erste Anlauf fragte
  // `weg.length === 0` und holte die Stufe danach mit `weg[0] ?? ''` — unter
  // `noUncheckedIndexedAccess` ein zweiter Zweig für dieselbe Frage, den kein
  // Aufruf erreicht. Ein `undefined` aus der Zerlegung **ist** die Antwort.
  const [stufe, ...rest] = weg

  if (stufe === undefined) {
    // **Das Paket wird ein eigenes Kind, nicht dieser Knoten.**
    //
    // Hier weicht die Umsetzung von MSE ab, und der Grund ist der Fall
    // darunter: MSE legt das Paket **in** den Knoten seiner letzten Wegstufe —
    // deshalb braucht es die Sonderbehandlung „zwei Pakete mit demselben Pfad"
    // (Zeilen 43–49). Ein Paket immer als eigenes Kind zu führen, ergibt
    // dieselbe Anzeige und lässt den Sonderfall entfallen.
    //
    // **Sichtbar wird der Unterschied an der Aufschrift**: MSE zeigt für ein
    // Paket den letzten Wegabschnitt, hier steht sein Name. Bei `magic/mana`
    // mit zwei Symbolschriften darin zeigte MSE zweimal „mana", diese Liste
    // `magic-mana-large` und `magic-mana-small` — was die Frage beantwortet,
    // die jemand vor der Liste hat.
    const blatt = neuerKnoten(paket.name, knoten.schluessel + ' ' + paket.name)
    blatt.paket = paket
    blatt.artrang = paket.artrang
    blatt.rang = paket.rang
    knoten.kinder.push(blatt)

    return
  }

  const schluessel =
    knoten.schluessel === '' ? stufe : knoten.schluessel + '/' + stufe

  // **Nur Gruppenknoten werden wiederverwendet.** Ein Blatt trägt ein Paket;
  // es unter sich weitere Stufen hängen zu lassen, machte aus einer Zeile mit
  // Paket eine mit Paket **und** Kindern — MSE fängt denselben Fall in
  // Zeilen 43–49 ab.
  let kind = knoten.kinder.find(
    (vorhanden) =>
      vorhanden.aufschrift === stufe && vorhanden.paket === undefined,
  )

  if (kind === undefined) {
    kind = neuerKnoten(stufe, schluessel)
    knoten.kinder.push(kind)
  }

  einfuegen(kind, paket, rest)
}

// vergleiche ist `compare_pos_hint` (Zeilen 61–67): erst die Art, dann der
// Rang, zuletzt die Aufschrift.
function vergleicheIn(sprache: string) {
  const namen = vergleicher(sprache)

  return (a: Knoten, b: Knoten): number => vergleiche(a, b, namen)
}

function vergleiche(
  a: Knoten,
  b: Knoten,
  namen: (a: string, b: string) => number,
): number {
  if (a.artrang !== b.artrang) {
    return a.artrang - b.artrang
  }

  if (a.rang !== b.rang) {
    return a.rang - b.rang
  }

  // **Nach der Sprache des Fensters und kein `<`.** MSE vergleicht wxStrings;
  // hier stehen Paketnamen mit Umlauten und Bindestrichen nebeneinander, und
  // `'ä' < 'b'` ist in JavaScript falsch (F01: `sprache/vergleich.ts`).
  return namen(a.aufschrift, b.aufschrift)
}

// ausgeben macht aus dem Baum die flache Liste — `toItems`, Zeilen 69–75.
//
// **Ein Gruppenknoten mit genau einem Paket und sonst nichts wird durch dieses
// Paket ersetzt.** MSE legt das Paket in diesem Fall **auf** den Knoten
// (Zeilen 39–42); der Baum ist damit eine Stufe flacher als eine Umsetzung,
// die jedem Paket ein eigenes Kind gibt.
//
// **Das ist kein Rand, sondern der Normalfall**: über `full-magic-pack`
// tragen **298 der 329 Gruppen genau ein Paket** — 90 %, gezählt am
// 2026-08-19. Der erste Anlauf gab jedem ein eigenes Kind und nannte das
// „dieselbe Anzeige"; der Zweitblick über Phase G hat beides angestrichen, die
// Behauptung und die Zusatzebene.
//
// **Die Aufschrift bleibt der Paketname**, und das ist die Abweichung, die
// bleibt: MSE zeichnet den letzten Wegabschnitt, und bei mehreren Paketen
// einer Gruppe stehen dann mehrere Zeilen mit demselben Wort. Der Paketname
// beantwortet die Frage, die jemand vor der Liste hat.
//
// **Ohne Kaskade.** Ein Knoten, dessen einziges Kind wieder eine Gruppe ist,
// bleibt stehen — sonst fiele `a/b/c` auf zwei Stufen zusammen, und MSE gibt
// dort drei.
function ausgeben(
  knoten: Knoten,
  zeilen: Baumzeile[],
  stufe: number,
  ordnung: (a: Knoten, b: Knoten) => number,
) {
  const geordnet = [...knoten.kinder].sort(ordnung)

  for (const kind of geordnet) {
    const alleine = einzelnesPaket(kind)

    if (alleine !== undefined) {
      zeilen.push(zeile(alleine, stufe))

      continue
    }

    zeilen.push(zeile(kind, stufe))

    ausgeben(kind, zeilen, stufe + 1, ordnung)
  }
}

// sichtbareZeilen lässt weg, was unter einer eingeklappten Gruppe liegt.
//
// **Der Grund ist gemessen**: über `full-magic-pack` stehen 526 Pakete in 332
// Gruppen, davon 364 unter dem einen Knoten `magic`. Ein Baum, der alles
// zeigt, ist rund 800 Zeilen lang — dieselbe Liste wie vorher, nur mit Einzug.
//
// **Ein Prädikat und keine Menge.** Der Aufrufer entscheidet, was „offen"
// heißt — beim Suchen ist es *alles*, sonst die Menge der aufgeklappten
// Schlüssel. Eine Menge hier zwänge ihn, für den Suchfall eine mit allen
// Schlüsseln zu bauen, die er dafür erst aus dem Ergebnis sammeln müsste.
//
// **Über die Stufe und nicht über den Schlüssel.** Die Liste ist flach und in
// Tiefensuche geordnet; wer eine Gruppe zuklappt, überspringt jede folgende
// Zeile, die tiefer steht als sie. Damit braucht der Vergleich keinen Pfad,
// und ein Enkel bleibt automatisch verborgen, wenn sein Elternteil es ist.
export function sichtbareZeilen(
  zeilen: readonly Baumzeile[],
  offen: (schluessel: string) => boolean,
): Baumzeile[] {
  const sichtbar: Baumzeile[] = []

  // verborgenAb ist die Stufe, ab der gerade übersprungen wird — `undefined`,
  // solange nichts zugeklappt ist.
  let verborgenAb: number | undefined

  for (const zeile of zeilen) {
    if (verborgenAb !== undefined && zeile.stufe > verborgenAb) {
      continue
    }

    verborgenAb = undefined
    sichtbar.push(zeile)

    if (zeile.paket === undefined && !offen(zeile.schluessel)) {
      verborgenAb = zeile.stufe
    }
  }

  return sichtbar
}

// einzelnesPaket gibt das Kind, wenn dieser Knoten **nur** dieses eine Paket
// trägt — sonst `undefined`.
function einzelnesPaket(knoten: Knoten): Knoten | undefined {
  if (knoten.paket !== undefined || knoten.kinder.length !== 1) {
    return undefined
  }

  const einziges = knoten.kinder[0]

  return einziges?.paket === undefined ? undefined : einziges
}

// zeile macht aus einem Knoten seine Zeile.
//
// **Ohne `paket`, wenn keines da ist** — nicht mit `undefined`.
// `exactOptionalPropertyTypes` unterscheidet beides, und eine Zeile mit
// `paket: undefined` ist keine Gruppenzeile, sondern eine Paketzeile ohne
// Paket.
function zeile(knoten: Knoten, stufe: number): Baumzeile {
  const gebaut: Baumzeile = {
    schluessel: knoten.schluessel,
    aufschrift: knoten.aufschrift,
    stufe,
  }

  if (knoten.paket !== undefined) {
    gebaut.paket = knoten.paket
  }

  return gebaut
}
