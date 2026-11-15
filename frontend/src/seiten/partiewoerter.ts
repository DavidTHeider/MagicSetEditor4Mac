// **Die Wörter des Spielkerns im Fenster** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum das eine Datei ist und nicht zwei gleiche Funktionen
//
// **Weil es dieselbe Frage ist** (§2): `bibliothek` ist ein Wort des Kerns,
// und ein Fenster, das es unübersetzt zeigte, spräche in zwei Sprachen
// zugleich (§9). Der Tisch und die Handlungsleiste stellen die Frage beide —
// und die zweite Kopie hatte einen Zweig, den kein Lauf betreten konnte: sie
// wird nur mit den sieben bekannten Zonen gerufen, und ihr Rückfall lag tot
// daneben (§3).
//
// # Und warum es einen Rückfall überhaupt gibt
//
// **Weil fremde Daten der Normalfall sind** (§8). Ein Kern, der eine neue
// Zone einführte, ließe das Fenster sonst mit einer leeren Zeile dastehen —
// die Kennung ist schlechter als ein Wort und besser als nichts.

import { einsetzen } from '../sprache/einsetzen'
import { sitzwort } from './sitzwort'

import type { Partie, Sitz } from '../apipartie'
import type { Sprache } from '../sprache/kataloge'

// wortAus liest einen Schlüssel aus einem Katalog — oder gibt den Rückfall.
//
// # Warum `unknown` und nicht `Record<string, string>`
//
// **Weil ein Katalog auch Verbünde trägt.** Seit PLAY-03.AK-04
// (#1711)
// steht `tischArt` als Verbund im Spielkatalog, und der **ganze** Katalog
// wird hier hineingereicht. „Nur Zeichenketten" war schon vorher eine
// Behauptung über etwas, das wächst — jetzt ist sie falsch.
//
// **Die Prüfung ist keine Vorsicht, sondern die Antwort**: was kein Wort
// ist, ist keines, und dann gilt derselbe Rückfall wie für einen fehlenden
// Schlüssel (§8).
function wortAus(
  eigene: Record<string, unknown>,
  schluessel: string,
  rueckfall: string,
): string {
  const wort = eigene[schluessel]

  return typeof wort === 'string' ? wort : rueckfall
}

// zonenwort übersetzt eine Zonenkennung.
export function zonenwort(
  eigene: Record<string, unknown>,
  zone: string,
): string {
  return wortAus(eigene, `zone${gross(zone)}`, zone)
}

// kurzzonenwort gibt die Kurzform einer Zone, wo sie eine hat — sonst ihr
// Wort.
//
// **Für die Ablagen**, seit dem 2026-09-24: über einem Stapel von 63
// Einheiten Breite passt „Bibliothek" nicht, und der Name lief in den des
// Friedhofs daneben. Wo keine Kurzform im Katalog steht, bleibt es beim
// ganzen Wort — auch das ist derselbe Rückfall wie bei [zonenwort].
export function kurzzonenwort(
  eigene: Record<string, unknown>,
  zone: string,
): string {
  return wortAus(eigene, `zonekurz${gross(zone)}`, zonenwort(eigene, zone))
}

// standwort übersetzt den Stand einer Partie — derselbe Weg.
export function standwort(
  eigene: Record<string, unknown>,
  stand: string,
): string {
  return wortAus(eigene, `stand${gross(stand)}`, stand)
}

// zeigerwort übersetzt, was jemand am Tisch gerade tut — derselbe Weg.
//
// **Auch hier mit Rückfall** (§8): der Kern darf eine fünfte Präsenzart
// bekommen, und bis das Fenster sie kennt, ist die Kennung besser als eine
// leere Zeile.
export function zeigerwort(
  eigene: Record<string, unknown>,
  art: string,
): string {
  return wortAus(eigene, `zeiger${gross(art)}`, art)
}

// handlungswort übersetzt eine Befehlsart oder einen Nutzlastschlüssel —
// #1623.
//
// **Ohne `gross`**, und das ist der Unterschied zu den dreien darüber: der
// Schlüssel ist die Art selbst (`objekt.tappen`), weil ein Punkt in einem
// Bezeichner nichts zu suchen hat und `objektTappen` daneben eine zweite
// Schreibweise wäre, die mit dem Kern auseinanderläuft.
//
// **Mit Rückfall** (§8): ein Kern, der einen Befehl dazubekommt, lässt das
// Fenster sonst mit einer leeren Beschriftung dastehen. Dass keiner fehlt,
// hält `internal/werkzeuge/befehlsformen_test.go` — die Kennung ist der
// Rückfall, falls er einmal nicht greift.
export function handlungswort(
  eigene: Record<string, unknown>,
  art: string,
): string {
  return wortAus(eigene, art, art)
}

// sitzname gibt den Namen zu einer Sitzkennung — die Kennung, wo keiner steht.
//
// **Einmal und nicht viermal.** Er stand in `Partietisch`, in
// `Legendenwahlen`, in `Mischpflichten` und in `Ausloeserreihung`, und alle
// vier beantworten dieselbe Frage (§2). Zusammengezogen beim Bau von `M9`
// (#1611),
// als die vierte dazukam.
//
// **Mit Rückfall auf die Kennung** (§8): ein von Hand geänderter Bestand
// kann einen Sitz ausgetragen haben, und eine leere Stelle im Satz wäre
// schlechter als ein Bezeichner.
export function sitzname(partie: Partie, kennung: string): string {
  return partie.sitze.find((sitz) => sitz.kennung === kennung)?.name ?? kennung
}

// gross macht aus `bibliothek` ein `Bibliothek` — der Katalogschlüssel.
function gross(wort: string): string {
  return `${wort.charAt(0).toUpperCase()}${wort.slice(1)}`
}

// Sitzworte sind die zwei Katalogeinträge, die ein Sitzkopf braucht.
//
// **Genau zwei und kein `Record<string, string>`.** Die Schlüssel stehen
// fest; über eine offene Abbildung bräuchte jeder Zugriff einen Rückfall,
// den kein Aufruf betritt.
export interface Sitzworte {
  standLeben: string
  standGift: string
  zugAktiv: string

  // **Die zwei aus PLAY-03**, die auf die Leinwand gehören — CR 728.1 und
  // CR 903.8. Der Commanderschaden gehört nicht dazu; der Grund steht bei
  // [sitzkopf].
  standStrahlung: string
  standSteuer: string
}

// sitzkopf baut den Satz, der am Sitz auf der Leinwand steht — PLAY-03,
// gezeichnet als `tsitz-kopf` im Artboard `Tisch`.
//
// **Nicht `sitzwort`**: so heißt seit G8 der **Name** eines Sitzes samt
// seinem Profil (`seiten/sitzwort.ts`). Zwei Bedeutungen für ein Wort wären
// zwei Auskünfte über dieselbe Sache.
//
// **Hier und nicht im Zeichner**: die Kataloge liegen in `sprache/`, und ein
// Renderer, der sie läse, hätte die Oberflächensprache. `sitzschilder` sagt,
// **wo** der Satz hinkommt; dies sagt, **was** er sagt.
//
// **Gift steht nur, wo es welches gibt.** Eine Null an jedem Sitz wäre eine
// Auskunft über nichts, und sie stünde auf jedem Tisch — auch auf denen, wo
// niemand je vergiftet wird.
//
// **Ein unbekannter Sitz bekommt seine Kennung** (§8). Besser eine Kennung
// als ein leeres Schild, das aussieht wie ein Fehler im Zeichner.
export function sitzkopf(
  partie: Partie,
  eigene: Sitzworte,
  sitz: string,
  name: (sitz: Sitz) => string,
): string {
  const gefunden = partie.sitze.find((einer) => einer.kennung === sitz)

  if (gefunden === undefined) {
    return sitz
  }

  const teile = [
    name(gefunden),
    `${String(gefunden.leben)} ${eigene.standLeben}`,
  ]

  if (gefunden.vergiftung > 0) {
    teile.push(`${String(gefunden.vergiftung)} ${eigene.standGift}`)
  }

  // **Strahlung und Steuer nach derselben Regel** — CR 728.1 und CR 903.8
  // (PLAY-03, #1711).
  //
  // **Die Zahl zuerst, wie beim Gift daneben** — und nicht „Steuer 2" wie
  // im Artboard. Auf einer Zeile, die an jedem Sitz steht, ist eine
  // Schreibweise besser als zwei; der Entwurf zeichnet den Wert dort als
  // Chip in einer anderen Anordnung.
  if (gefunden.strahlung > 0) {
    teile.push(`${String(gefunden.strahlung)} ${eigene.standStrahlung}`)
  }

  if (gefunden.commandersteuer > 0) {
    teile.push(`${String(gefunden.commandersteuer)} ${eigene.standSteuer}`)
  }

  // **Der Commanderschaden steht hier nicht.** Eine Zeile auf der Leinwand
  // trägt keine Liste — sie steht an jedem Sitz und wird bei vier Sitzen
  // viermal gelesen. Er steht in der Spielerleiste.

  // **„Am Zug" steht nur an einem Sitz** — der Entwurf setzt „wartet" an die
  // übrigen, und das wäre an drei von vier Sitzen dreimal dieselbe Auskunft.
  //
  // **Und gar nicht, wo es keinen Zug gibt**: ein Sandkasten aus G1 bis G8
  // geht nicht zugweise zu, und `aktiv` ist dort leer.
  if (partie.zug.aktiv !== '' && partie.zug.aktiv === sitz) {
    teile.push(eigene.zugAktiv)
  }

  return teile.join(' · ')
}

// sitzkopfFuer bindet, was sich am Tisch nicht ändert — Partie, Katalog und
// die Namen der Profile — und gibt die Funktion, die der Zeichner braucht.
//
// **Hier und nicht als Pfeil im JSX.** Ein Pfeil dort ist eine Funktion, die
// kein Lauf betritt: die Fläche reicht sie an den Zeichner weiter, und der
// zeichnet unter jsdom nicht. Sie wäre eine unbetretene Zeile, die niemand
// betreten *kann* — und damit entweder ein fehlender Test oder toter Code
// (§3). So ist es eine Funktion mit einem eigenen Lauf.
//
// **Und sie nimmt die Profilnamen statt einer Namensfunktion**, aus demselben
// Grund: eine Funktion, die der Aufrufer im JSX baut, ist wieder ein Pfeil,
// den niemand betritt. Den Namen bildet `sitzwort` (G8), damit auf der
// Leinwand derselbe steht wie in der Sichtwahl daneben.
export function sitzkopfFuer(
  partie: Partie,
  eigene: Sitzworte,
  namen: Map<string, string>,
): (sitz: string) => string {
  return (sitz) =>
    sitzkopf(partie, eigene, sitz, (einer) => sitzwort(einer, namen))
}

// Partienamen sind die drei Wendungen, mit denen eine Partie heißt.
interface Partienamen {
  partieGegen: string
  partieZwischen: string
  partieVon: string
}

// partiename sagt, wie eine Partie über dem Tisch heißt — die Brotkrume,
// UX-03.AK-02 (#1675).
//
// **Aus der Sicht dessen, der hinsieht**: wer an einem Sitz spielt, spielt
// **gegen** die anderen, wie im Artboard *Tisch* („Partie gegen Bo"); wer
// zusieht, sieht eine Partie **zwischen** allen; und wer allein spielt, hat
// kein Gegenüber.
//
// **Aufgezählt in der Sprache des Fensters** (`Intl.ListFormat`) — „und"
// hier, „, and" dort.
export function partiename(
  eigene: Partienamen,
  partie: Partie,
  sprache: Sprache,
): string {
  const liste = new Intl.ListFormat(sprache, { type: 'conjunction' })
  const namen = (sitze: Sitz[]) => liste.format(sitze.map((sitz) => sitz.name))

  const spielt = partie.sitze.some((sitz) => sitz.kennung === partie.fuer)
  const gegner = partie.sitze.filter((sitz) => sitz.kennung !== partie.fuer)

  if (!spielt) {
    return einsetzen(eigene.partieZwischen, { namen: namen(partie.sitze) })
  }

  if (gegner.length === 0) {
    return einsetzen(eigene.partieVon, { namen: namen(partie.sitze) })
  }

  return einsetzen(eigene.partieGegen, { namen: namen(gegner) })
}
