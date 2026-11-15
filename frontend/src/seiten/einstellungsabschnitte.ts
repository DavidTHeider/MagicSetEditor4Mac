// **Abschnitte in einer Gruppe der Einstellungen** — UXR-19
// (#1927).
//
// # Der Anlass
//
// **Am nativen Fenster gemessen, 2026-10-02** (`docs/UIREVIEW.md`, Commit
// ddadb180): die Einstellungen zeigten in *Allgemein* **21 Felder** und
// rollten **3,9 Bildschirme**. Das Prüfmaß erlaubt zwölf Felder, sonst
// „gruppiert und eingeklappt bis auf eine Gruppe" (R3), und höchstens 2,5
// Bildschirme (R4). Zehn der 21 waren die Belegung von ⌘0 bis ⌘9 — einmal
// eingestellt und immer gezeichnet —, und die Kürzeltafel darüber allein ist
// höher als das Fenster.
//
// # Was hier steht
//
// **Welche Einstellungen einer Gruppe hinter einem Aufklapper stehen**, je
// Gruppe in der Reihenfolge des Zeichnens. Was in keinem Abschnitt steht,
// bleibt offen — das ist die eine Gruppe, die R3 offen lässt.
//
// **Schlüssel und nicht Präfixe**, aus demselben Grund wie in
// `einstellungsgruppen.ts`: eine Regel über Präfixe griffe beim nächsten
// Schlüssel still daneben.
//
// **Auch Schlüssel, die ein Beiwerk zeichnet.** `kuerzel.ziffern` steht in
// keiner Wahl, sondern in der Kürzeltafel. Die Suche muss trotzdem wissen,
// welchen Kasten sie aufhält — sonst führte „Belegung" zu einem Feld, das
// gar nicht gezeichnet ist.
//
// # Was er nicht ist
//
// **Kein zweiter Kasten für Fachleute.** Der steht in `fachschluessel`
// (UX-04.AK-04) und sagt „das brauchst du nicht"; ein Abschnitt sagt „das
// gehört zusammen und wird selten gebraucht". Beides darf in einer Gruppe
// stehen, aber kein Schlüssel in beiden — das hält der Lauf daneben.
import type { Einstellungsgruppe as Gruppe } from './einstellungsgruppen'
import type { Katalog } from '../sprache/kataloge'

// abschnittsnamen sind alle Abschnitte, über alle Gruppen.
export const abschnittsnamen = [
  'eigeneUebersetzungen',
  'kuerzel',
  'fenster',
  'spalten',
  'handel',
  'ordner',
  'protokoll',
  'messung',
  'speicher',
  'werkzeuge',
  'adressen',
  'dienstkatalog',
  'zugaenge',
  'eingriffe',
  'kosten',
  'mitnehmen',
  'kimodelle',
  'eigeneFarben',
  'wirkung',
] as const

export type Abschnittsname = (typeof abschnittsnamen)[number]

// abschnittstitel ist der Name am Knopf eines Abschnitts.
//
// **Wo das Bauteil schon einen Namen trägt, ist es dieser** — die eigenen
// Übersetzungen und die Tastenkürzel heißen in der Hilfe und in ihrem
// Bauteil so, die Werkzeuge ohne Fenster trugen ihn als Überschrift. Ein zweiter Katalogeintrag mit demselben Satz liefe beim
// nächsten Übersetzen auseinander.
export function abschnittstitel(texte: Katalog, name: Abschnittsname): string {
  const eigene = texte.seiten.einstellungen
  const titel: Record<Abschnittsname, string> = {
    eigeneUebersetzungen: eigene.eigenuebersetzung.titel,
    kuerzel: eigene.kuerzelUeberschrift,
    fenster: eigene.abschnitte.fenster,
    spalten: eigene.abschnitte.spalten,
    handel: eigene.abschnitte.handel,
    ordner: eigene.abschnitte.ordner,
    protokoll: eigene.abschnitte.protokoll,
    messung: eigene.abschnitte.messung,
    speicher: eigene.abschnitte.speicher,
    werkzeuge: texte.seiten.technik.ueberschrift,
    adressen: eigene.abschnitte.adressen,
    dienstkatalog: eigene.katalog,
    zugaenge: eigene.zugaenge,
    eingriffe: eigene.abschnitte.eingriffe,
    kosten: texte.seiten.kosten.titel,
    mitnehmen: eigene.abschnitte.mitnehmen,
    kimodelle: eigene.kimodelleTitel,
    eigeneFarben: eigene.gestaltung.merkmale,
    wirkung: texte.seiten.technik.metriken,
  }

  return titel[name]
}

// Abschnittsplan sagt, welche Schlüssel ein Abschnitt trägt.
export interface Abschnittsplan {
  name: Abschnittsname
  schluessel: readonly string[]

  // oben stellt ihn vor das Beiwerk der Gruppe statt darunter (AP-10).
  oben?: boolean
}

// abschnitteJeGruppe — eine Gruppe ohne Eintrag hat keinen Abschnitt.
export const abschnitteJeGruppe: Readonly<
  Partial<Record<Gruppe, readonly Abschnittsplan[]>>
> = {
  // **Allgemein: was jeder braucht, steht offen** — Sprache, Darstellung,
  // Zeitzone, Aktualisierung, Erste Schritte. Eine eigene Übersetzung legt
  // an, wer eine Sprache vermisst (LOC-02, #1688); die Tafel der Kürzel
  // steht ein zweites Mal in der Hilfe, und ihre Belegung ändert man einmal.
  allgemein: [
    { name: 'eigeneUebersetzungen', schluessel: ['sprache.eigene'] },
    { name: 'kuerzel', schluessel: ['kuerzel.ziffern'] },
  ],

  // **Erscheinungsbild: Thema und Schrift stehen offen** — sie stellt man
  // um, um zu sehen, wie es aussieht. Wie das Fenster aufgeht und wo die
  // Leiste steht, stellt man einmal ein; wie breit die Spalten eines
  // Arbeitsplatzes sind, stellt man meist mit der Maus am Arbeitsplatz
  // selbst ein, und hier steht es für den, der einen Wert eintippen will.
  //
  // **Die eigenen Farben stehen seit #2465 eingeklappt** — ein Feld für
  // Merkmale wie `--mse-marke`, das einstellt, wer einmal etwas Eigenes
  // will; das Fensterbild darüber bleibt offen, weil man es ansieht.
  erscheinungsbild: [
    { name: 'eigeneFarben', schluessel: ['gestaltung'] },
    {
      name: 'fenster',
      schluessel: [
        'seitenleiste',
        'leistenlage',
        'fensterstart',
        'fensterbreite',
        'fensterhoehe',
      ],
    },
    {
      name: 'spalten',
      schluessel: [
        'arbeitsplatz.navigator',
        'arbeitsplatz.inspektor',
        'arbeitsplatz.navigatorbreite',
        'arbeitsplatz.inspektorbreite',
        'arbeitsplatz.ansicht.sets',
        'arbeitsplatz.ansicht.decks',
        'arbeitsplatz.ansicht.sammlung',
      ],
    },
  ],

  // **KI: ob gefragt wird, steht offen** — UXR-21 (#1929). Die Schalter der
  // Anbieter, die Laufzeit, das Gerätemodell, Kontingent und Modelle. Wohin
  // gefragt wird, stellt man einmal ein oder nie: am Fenster gemessen
  // rollte die Gruppe mit den vier Adressen 3,2 Bildschirme, eng 4,2.
  //
  // **Seit #2465 auch die Liste der Modelle** (1.024 × 700: 2,9): sie sagt,
  // was feststeht, und wird nachgeschlagen; ob gefragt wird, entscheiden die
  // Schalter, die Laufzeit und das Kontingent darüber.
  ki: [
    { name: 'kimodelle', schluessel: [] },
    {
      name: 'adressen',
      schluessel: [
        'dienst.anthropic.adresse',
        'dienst.openai.adresse',
        'dienst.xai.adresse',
        'ki.adresse',
      ],
    },
  ],

  // **Verbindungen: die Schnittstellen zuerst** — AP-10 (#2367), Befund
  // B-19. Die Ein-Klick-Einrichtung von `mse`, `mse-mcp` und `mse-rest`
  // stand unter *Diagnose* hinter Protokollen und Messung; hier steht sie bei
  // dem Netz, über das ein Dienst lauscht, und dort, wo die Hilfe den Dienst
  // starten lässt.
  verbindungen: [{ name: 'werkzeuge', schluessel: [], oben: true }],

  // **Dienste: wer gefragt wird, steht offen** — die Kartendienste, die
  // Frage, das Marktkonto und „nur lokal". Die vier Schranken des Handels
  // und sein Schalter urteilen über einen Weg, den man einmal festlegt.
  //
  // **Seit #2465 auch der Katalog und die Zugangsdaten**: am Fenster trugen
  // sie 3.200 von 4.200 Punkten (5,2 Bildschirme). Der Katalog wird
  // nachgeschlagen, ein Zugang einmal hinterlegt.
  dienste: [
    { name: 'dienstkatalog', schluessel: [] },
    { name: 'zugaenge', schluessel: [] },
    {
      name: 'handel',
      schluessel: [
        'richtlinie.handel',
        'richtlinie.handel.markt',
        'richtlinie.handel.menge',
        'richtlinie.handel.wert',
        'richtlinie.handel.frist',
      ],
    },
  ],

  // **Daten und Sicherung: der Bestand steht offen** — was Platz belegt,
  // die Sicherung, das Zurückgehen. Wohin hochgeladen, ausgeführt,
  // gesichert, diagnostiziert und Bilder geschrieben wird, und was bei einem
  // gleichnamigen Bild geschieht, legt man einmal fest.
  //
  // **Seit #2465 offen nur noch Platz und Sicherung** (3,7 und eng 5,0):
  // Zurückgehen, Aufräumen und Zurücksetzen sind dieselbe Sorte Eingriff
  // und stehen in einem Abschnitt; was ein Auftrag kosten darf und das
  // Mitnehmen der Einstellungen legt man einmal fest.
  daten: [
    { name: 'eingriffe', schluessel: [] },
    { name: 'kosten', schluessel: [] },
    { name: 'mitnehmen', schluessel: [] },
    {
      name: 'ordner',
      schluessel: [
        'vorgabe.hochladepfad',
        'vorgabe.ausfuhrpfad',
        'vorgabe.sicherungspfad',
        'vorgabe.diagnosepfad',
        'bildausfuhr.ordner',
        'bildkonflikt',
      ],
    },
  ],

  // **Diagnose: der Zustand und der Fehlerbericht stehen offen** — wer
  // hierherkommt, fragt, warum gerade etwas nicht geht. Wie ausführlich das
  // Protokoll schreibt, wie gemessen wird und was zwischengespeichert und
  // indiziert ist, fragt man selten — und jedes für sich. Die Werkzeuge ohne
  // Fenster stehen seit AP-10 unter *Verbindungen*.
  //
  // **Seit #2465 auch, was seit dem Start gewirkt hat** (eng 2,7): eine
  // Statistik, nach der man fragt, wenn der Zustand darüber nicht reicht.
  diagnose: [
    { name: 'wirkung', schluessel: [] },
    {
      name: 'protokoll',
      schluessel: [
        'protokoll.stufe',
        'protokoll.datei',
        'protokoll.system',
        'protokoll.ausgabe',
        'protokoll.format',
        'protokoll.grenze',
        'protokoll.aufbewahrung',
        'protokoll.komponenten',
        'protokoll.ort',
      ],
    },
    {
      name: 'messung',
      schluessel: [
        'messung.spur',
        'messung.abtastung',
        'messung.spurgrenze',
        'messung.aufbewahrung',
        'messung.profile',
      ],
    },
    { name: 'speicher', schluessel: ['zwischenspeicher.grenze'] },
  ],
}

// imAbschnitt sind alle Schlüssel, die in einem Abschnitt stehen.
const imAbschnitt: ReadonlySet<string> = new Set(
  Object.values(abschnitteJeGruppe).flatMap((plaene) =>
    plaene.flatMap((plan) => plan.schluessel),
  ),
)

// ausserAbschnitten lässt aus, was ein Abschnitt zeichnet — ein Beiwerk,
// das es sonst mitzeichnete, stünde doppelt (UXR-21, #1929).
export function ausserAbschnitten<T extends { schluessel: string }>(
  eintraege: readonly T[],
): T[] {
  return eintraege.filter((eintrag) => !imAbschnitt.has(eintrag.schluessel))
}

// Aufschlag ist der Wunsch einer Übergabe, einen Abschnitt aufzuschlagen —
// AP-10 (#2367): die Gruppe, in der er steht, und ein Zeichen je Wunsch,
// damit derselbe Abschnitt ein zweites Mal aufgehen kann.
export interface Aufschlag {
  gruppe: string
  zeichen: string
}

// aufschlagAus macht aus einer Übergabe den Wunsch — keinen ohne Abschnitt.
export function aufschlagAus(
  gewuenscht: { gruppe: string; abschnitt?: string | undefined },
  nummer: number,
): Aufschlag | undefined {
  return gewuenscht.abschnitt === undefined
    ? undefined
    : {
        gruppe: gewuenscht.gruppe,
        zeichen: `${gewuenscht.abschnitt}#${String(nummer)}`,
      }
}

// mitAufschlag gibt dem gewünschten Abschnitt einer Gruppe sein Zeichen.
export function mitAufschlag<T extends { name: string }>(
  plaene: readonly T[],
  gruppe: string,
  aufschlag: Aufschlag | undefined,
): (T & { aufschlagen?: string })[] {
  return plaene.map((plan) =>
    aufschlag?.gruppe === gruppe &&
    aufschlag.zeichen.startsWith(`${plan.name}#`)
      ? { ...plan, aufschlagen: aufschlag.zeichen }
      : plan,
  )
}
