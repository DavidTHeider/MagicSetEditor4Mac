// Was die Paketliste zeigt — Suche, Artfilter und Ordnung (#244).
//
// # Warum eine eigene Datei
//
// Weil hier die Entscheidungen stehen und nicht die Klassennamen: *welche
// Zeilen jetzt dastehen* ist die Frage, an der man sich vertut. Der Baum
// daneben ([paketbaum]) beantwortet eine andere — *wie sie zusammengehören* —
// und ist eine Portierung; diese hier ist es nicht und sagt das.
import type { PaketDTO, Paketherkunftsart } from '../api'
import { paketbaum, type Baumzeile } from './paketbaum'
import { vergleicher } from '../sprache/vergleich'

// Ordnung ist die gewählte Reihenfolge.
//
// **`gruppen` ist der Normalfall und MSEs Antwort**: der Baum aus
// `installer group:`. Die drei flachen Ordnungen daneben beantworten Fragen,
// auf die ein Baum keine Antwort gibt — „was ist alt", „was kommt von dort".
export type Ordnung = 'gruppen' | 'name' | 'version' | 'herkunft'

// Sichtwahl ist, was die Bedienenden eingestellt haben.
export interface Sichtwahl {
  suche: string

  // art ist leer für „alle Arten".
  art: string

  // herkunft ist leer für „alle Herkünfte" — DOM-01.AK-01 (#1696).
  herkunft: Paketherkunftsart | ''

  ordnung: Ordnung

  // auswahl zeigt genau diese Pakete — die Sammelzeile der Aufmerksamkeit
  // (AP-8, #2365). **Über den Namen**, wie das Arbeitsobjekt eines Pakets
  // im ganzen Programm heißt; `name` ist der Satz, der sie nennt.
  auswahl?: { name: string; namen: readonly string[] } | undefined
}

// leereWahl ist die Sicht ohne Filter.
export const leereWahl: Sichtwahl = {
  suche: '',
  art: '',
  herkunft: '',
  ordnung: 'gruppen',
}

// sicht gibt die Zeilen, die die Liste zeigt.
//
// `ohneGruppe` ist die Aufschrift des Sammelknotens — sie kommt aus den
// Sprachkatalogen und reist durch bis in [paketbaum].
//
// **Erst filtern, dann ordnen** — nicht umgekehrt. Wer den Baum baut und
// danach Zeilen wegwirft, lässt Gruppen ohne Inhalt stehen; wer die Pakete
// filtert und den Baum danach baut, bekommt den Weg zu jedem Treffer
// geschenkt.
export function sicht(
  pakete: readonly PaketDTO[],
  wahl: Sichtwahl,
  ohneGruppe: string,
  sprache: string,
): Baumzeile[] {
  const gefiltert = pakete.filter((paket) => passt(paket, wahl))

  if (wahl.ordnung === 'gruppen') {
    return paketbaum(gefiltert, ohneGruppe, sprache)
  }

  return flach(gefiltert, wahl.ordnung, sprache)
}

// passt sagt, ob ein Paket durch Suche und Artfilter kommt.
function passt(paket: PaketDTO, wahl: Sichtwahl): boolean {
  if (wahl.auswahl !== undefined && !wahl.auswahl.namen.includes(paket.name)) {
    return false
  }

  if (wahl.art !== '' && paket.art !== wahl.art) {
    return false
  }

  if (wahl.herkunft !== '' && paket.herkunftsart !== wahl.herkunft) {
    return false
  }

  const frage = wahl.suche.trim().toLowerCase()
  if (frage === '') {
    return true
  }

  // **Vier Angaben und nicht nur der Name.** Wer „m15" tippt, meint den Stil;
  // wer „example.invalid" tippt, meint alles von dort; wer eine Art tippt,
  // meint eine Art. Eine Suche, die nur den Namen liest, zwingt zum Raten,
  // wonach sie sucht.
  return [paket.name, paket.art, paket.herkunft, paket.gruppe].some((angabe) =>
    angabe.toLowerCase().includes(frage),
  )
}

// flach ordnet ohne Gruppen — eine Stufe, jede Zeile ein Paket.
function flach(
  pakete: readonly PaketDTO[],
  ordnung: Ordnung,
  sprache: string,
): Baumzeile[] {
  return [...pakete].sort(ordnungsvergleich(ordnung, sprache)).map((paket) => ({
    schluessel: paket.art + '/' + paket.name,
    aufschrift: paket.name,
    stufe: 0,
    paket,
  }))
}

// ordnungsvergleich gibt die Ordnung für die flache Liste — in der Sprache
// des Fensters (F01, `sprache/vergleich.ts`).
function ordnungsvergleich(ordnung: Ordnung, sprache: string) {
  const vergleiche = vergleicher(sprache)

  return (a: PaketDTO, b: PaketDTO): number => {
    if (ordnung === 'version') {
      return absteigendMitLeerHinten(
        a.version,
        b.version,
        a.name,
        b.name,
        vergleiche,
      )
    }

    if (ordnung === 'herkunft') {
      return aufsteigendMitLeerHinten(
        a.herkunft,
        b.herkunft,
        a.name,
        b.name,
        vergleiche,
      )
    }

    return vergleiche(a.name, b.name)
  }
}

type Vergleich = (a: string, b: string) => number

// absteigendMitLeerHinten ordnet neu vor alt — und „keine Angabe" ganz hinten.
//
// **Ein Paket ohne `version:` ist nicht das älteste**, man weiß es nicht. Wer
// die leere Zeichenkette einfach mitsortierte, machte aus einer fehlenden
// Angabe eine Aussage.
function absteigendMitLeerHinten(
  a: string,
  b: string,
  namea: string,
  nameb: string,
  vergleiche: Vergleich,
): number {
  if (a === '' || b === '') {
    return leerHinten(a, b, namea, nameb, vergleiche)
  }

  return vergleiche(b, a) || vergleiche(namea, nameb)
}

// aufsteigendMitLeerHinten ordnet a–z — und „keine Angabe" ganz hinten.
function aufsteigendMitLeerHinten(
  a: string,
  b: string,
  namea: string,
  nameb: string,
  vergleiche: Vergleich,
): number {
  if (a === '' || b === '') {
    return leerHinten(a, b, namea, nameb, vergleiche)
  }

  return vergleiche(a, b) || vergleiche(namea, nameb)
}

// leerHinten entscheidet, wenn mindestens eine Angabe fehlt.
function leerHinten(
  a: string,
  b: string,
  namea: string,
  nameb: string,
  vergleiche: Vergleich,
): number {
  if (a === '' && b === '') {
    return vergleiche(namea, nameb)
  }

  return a === '' ? 1 : -1
}

// arten nennt die Arten, die im Bestand vorkommen — geordnet.
//
// **Aus dem Bestand und nicht aus einer Aufzählung.** Sieben Arten sind
// möglich, installiert sind meist drei; eine feste Liste böte vier Filter an,
// die nichts finden. Und ein Paket einer achten Art, das jemand ins
// Verzeichnis legt, wäre über einen festen Filter nicht erreichbar.
export function arten(pakete: readonly PaketDTO[], sprache: string): string[] {
  return [...new Set(pakete.map((paket) => paket.art))].sort(
    vergleicher(sprache),
  )
}
