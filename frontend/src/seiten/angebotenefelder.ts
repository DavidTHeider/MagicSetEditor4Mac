// **Welche Felder der Karteneditor anbietet** —
// #562.
//
// Eigene Datei, weil `Karteneditor.tsx` mit ihr über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, was das Fenster **tut**,
// hier, **welche** Felder es zeigt.
import type { FeldbeschreibungDTO, WertDTO } from '../api'
import type { FeldflaecheDTO } from '../formen'

// notizenfeld ist der Name, unter dem die Notizen einer Karte stehen.
//
// **Hier und in `Kartennotizen.tsx` derselbe Name** — er steht so in jeder
// `.mse-set` und wird nicht übersetzt (§9).
export const notizenfeld = 'notes'

export interface Angebotenesfeld {
  feld: string
  wert: string
}

// **Angeboten wird, was auf der Karte liegt** — nicht, was die Karte führt.
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–157, Fassung f9d9356 vom
// 2021-07-06. Eine frisch angelegte Karte trägt **keinen** Feldwert — die
// Vorgaben stehen im Spielpaket, nicht im Bestand. Der Editor listete
// `karte.werte` und zeigte deshalb ein leeres Formular: eine neue Karte ließ
// sich nicht füllen.
//
// **In der Reihenfolge des Spielpakets.** MSE baut die Werteliste einer Karte
// aus `game.card_fields` (`src/data/card.cpp`, Zeile 37, Fassung f9d9356 vom
// 2021-07-06) und zeigt sie in dieser Ordnung; die Ebenen des Plans stehen
// dagegen in Zeichenreihenfolge, und die setzt den Hintergrund vor den Namen.
//
// **Und was die Karte darüber hinaus führt, bleibt stehen.** Ein Set aus MTG
// JSON trägt Felder, für die die gewählte Vorlage keine Ebene hat; sie zu
// verstecken hieße, sie stillschweigend zu verlieren.
export function angeboteneFelder(
  werte: readonly WertDTO[],
  formen: readonly FeldbeschreibungDTO[],
  flaechen: readonly FeldflaecheDTO[],
  // reihenfolge ist eine feste Ordnung der gezeichneten Felder — ohne sie
  // gilt die Leseordnung der Flächen (`formularruhe.ts`, Befund 9).
  reihenfolge: readonly string[] = leseordnung(flaechen),
): Angebotenesfeld[] {
  const gefuehrt = new Map(werte.map((wert) => [wert.feld, wert.wert]))

  const nachForm = new Map(formen.map((form) => [form.name, form]))

  const felder: Angebotenesfeld[] = []
  const einzeln = (feld: string, vorgabe: string) => {
    if (!felder.some((eintrag) => eintrag.feld === feld)) {
      // **Der Wert der Karte schlägt den Anfangswert.** Der Anfangswert ist,
      // was **ohne** Wert gilt — nicht, was statt seiner gilt.
      felder.push({ feld, wert: gefuehrt.get(feld) ?? vorgabe })
    }
  }

  // **Ein Feld, das aus anderen entsteht, wird durch sie ersetzt**
  // (#696).
  //
  // `text` trägt bei `magic.mse-game` `save value: false` und setzt sich über
  // `combined_editor` aus Regeltext und Flavortext zusammen. Der Editor bot
  // einen Kasten „Text" an — wer hineintippte, sah die Karte unverändert, und
  // die Eingabe war beim nächsten Öffnen weg.
  //
  // **An seiner Stelle**, nicht am Ende: es steht dort, wo der Text auf der
  // Karte liegt, und seine Teile stehen untereinander in derselben
  // Reihenfolge wie dort.
  //
  // **Ein Teil ohne Erklärung bleibt weg.** Er hätte kein Steuerelement,
  // keine Aufschrift und keinen Anfangswert — angeboten wäre er ein leeres
  // Kästchen ohne Namen.
  const dazu = (feld: string, vorgabe: string) => {
    const teile = nachForm.get(feld)?.teilfelder ?? []
    if (teile.length === 0) {
      // **Was niemand bearbeiten darf, steht nicht als Eingabefeld da**
      // (#708).
      //
      //	bool is_enabled(ValueViewerP const& v) {
      //	  return v->getField()->editable && v->isVisible();
      //	}
      //
      // `src/gui/control/card_editor.cpp`, Zeilen 155–156, Fassung f9d9356
      // vom 2021-07-06. Der Kommentar oben führte diese Regel schon an —
      // umgesetzt war davon nur die **Sichtbarkeit**. 24 der 249 Kartenfelder
      // von `magic.mse-game` tragen `editable: false`, und jedes stand als
      // Eingabefeld da, das der Schreibweg danach abwies.
      //
      // **Ein Feld ohne Erklärung bleibt stehen**: es ist eines, das das Spiel
      // nicht kennt, und die Karte führt es ohnehin.
      if (nachForm.get(feld)?.editierbar === false) {
        return
      }

      einzeln(feld, vorgabe)

      return
    }

    for (const teil of teile) {
      const form = nachForm.get(teil)
      if (form !== undefined) {
        einzeln(teil, form.vorgabe)
      }
    }
  }

  // **Zuerst, was der Stil zeichnet — in der Ordnung, in der es auf der Karte
  // liegt** (#674).
  //
  // # adapted, und der Grund gehört dazu
  //
  // Die Reihenfolge des Spielpakets ist die von MSE, und sie stimmt: `Card`
  // baut seine Werteliste aus `game.card_fields` (`src/data/card.cpp`,
  // Zeile 37, Fassung f9d9356 vom 2021-07-06). **Nur zeigt MSE gar kein
  // Formular** — dort liegen die Felder **auf** der Karte (`ValueViewer`,
  // `src/gui/control/card_editor.cpp`), und niemand liest sie von oben nach
  // unten.
  //
  // `magic.mse-game` erklärt `border color`, `corners` und `card color` vor
  // `name`. In einer senkrechten Liste heißt das: die ersten drei Felder, die
  // jemand sieht, sind Rahmenfarbe, Eckenform und Kartenfarbe — und der
  // **Name** steht an vierter Stelle. Die Liste ist eine Anpassung dieses
  // Projekts; dann gehört auch ihre Ordnung hierher.
  //
  // **Von oben nach unten, bei gleicher Höhe von links nach rechts** — wie
  // jemand die Karte ansieht, und wie er die Vorschau daneben liest.
  // **Über die Flächen und nicht über die Formen**: jede Fläche hat ihren
  // Platz in der Leseordnung, eine Form hätte ihn nur, wenn sie gezeichnet
  // wird. Ein Rückgriff für den anderen Fall wäre ein Zweig, den kein Aufruf
  // betreten kann (§3).
  for (const name of reihenfolge) {
    const form = nachForm.get(name)
    if (form !== undefined) {
      dazu(name, form.vorgabe)
    }
  }

  // **Und danach, was die Karte führt, ohne dass der Stil es zeichnet** — in
  // der Reihenfolge des Spiels. Ein Set aus MTG JSON trägt solche Felder; sie
  // zu verstecken hieße, sie stillschweigend zu verlieren.
  for (const form of formen) {
    if (gefuehrt.has(form.name)) {
      dazu(form.name, form.vorgabe)
    }
  }

  // **Ein Feld ohne Form hat keinen Anfangswert** — es ist eines, das das
  // Spiel nicht erklärt, und die Karte führt es ohnehin.
  for (const wert of werte) {
    // **`notes` ist keines dieser Felder.**
    //
    // Es ist eine Eigenschaft der Karte und kein Feldwert des Spiels:
    // `String notes;` steht in `Card` **neben** `data` (`src/data/card.hpp`,
    // Zeile 39, Fassung `641acff` vom 2020-05-19), und MSE2 gibt ihm einen
    // eigenen Kasten unter der Kartenliste (`cards_panel.cpp`, Zeilen 39–50).
    //
    // Bis zum 2026-08-27 fiel es in diese Schleife: bei einer Karte, die eines
    // trägt, stand am Ende der Feldliste ein einzeiliges Kästchen mit der
    // Aufschrift »notes« — an einer Karte ohne stand gar keines, und niemand
    // konnte eine Notiz **anlegen**. Der Kasten dafür ist `Kartennotizen.tsx`.
    if (wert.feld === notizenfeld) {
      continue
    }

    dazu(wert.feld, '')
  }

  return felder
}

// leseordnung gibt je Feld seinen Platz in der Leseordnung der Karte.
//
// # Zeilen, nicht Bildpunkte
//
// Ein Vergleich nach `oben` allein steht auf dem falschen Fuß: `card symbol`
// beginnt bei 29, `casting cost` bei 29 und `name` bei **30** — ein
// Bildpunkt, und der Name stünde hinter den Kosten. Auf der Karte stehen alle
// drei nebeneinander in der Titelzeile.
//
// Deshalb werden zuerst **Zeilen** gebildet: Felder, deren senkrechte Spanne
// sich überlappt, gehören zusammen. Innerhalb einer Zeile gilt links vor
// rechts.
//
// # Der Grund kommt zuletzt
//
// `card color`, `border color`, `foil layer` und `corners` decken bei
// `magic-m15` **die ganze Karte** — 375 × 523, während das größte gesetzte
// Feld auf 37 % davon kommt. Sie liegen nicht an einer Stelle der Karte, sie
// **sind** die Karte. In einer Leseordnung stehen sie am Ende.
export function leseordnung(flaechen: readonly FeldflaecheDTO[]): string[] {
  // **Die Karte ist die Hülle aller Flächen**, nicht die größte einzelne.
  //
  // Der erste Anlauf maß gegen die größte Fläche — und bei einem Stil, dessen
  // Felder alle gleich groß sind, war damit **jedes** ein Grund. Die Hülle
  // gibt es immer, und sie ist das, was jemand ansieht.
  const kartenflaeche =
    Math.max(0, ...flaechen.map((flaeche) => flaeche.links + flaeche.breite)) *
    Math.max(0, ...flaechen.map((flaeche) => flaeche.oben + flaeche.hoehe))

  // **Die erste Fläche eines Feldes entscheidet.** Ein Feld kann mehrere
  // haben — ein zweiseitiger Rahmen zeichnet denselben Namen zweimal; dann
  // gilt die obere.
  const erste = new Map<string, FeldflaecheDTO>()
  for (const flaeche of flaechen) {
    if (!erste.has(flaeche.feld)) {
      erste.set(flaeche.feld, flaeche)
    }
  }

  const gesetzt: FeldflaecheDTO[] = []
  const grund: FeldflaecheDTO[] = []

  for (const flaeche of erste.values()) {
    if (flaeche.breite * flaeche.hoehe >= AnteilAlsGrund * kartenflaeche) {
      grund.push(flaeche)

      continue
    }

    gesetzt.push(flaeche)
  }

  gesetzt.sort((links, rechts) => links.oben - rechts.oben)

  const ordnung: { feld: string; platz: number }[] = []
  let zeile = 0

  // **Die Zeile misst sich am ersten Feld, nicht am bisher tiefsten.**
  //
  // Der erste Anlauf zog die Untergrenze mit jedem Feld nach — und bei
  // `magic-m15` wuchs sie über das halbe Kartenbild: `watermark` reicht von
  // 321 bis 487, und alles darunter bis zur Fußzeile fiel in **eine** Zeile.
  // Der Regeltext stand dann hinter dem Sammlerzeichen.
  //
  // Ein Feld gehört zur Zeile, wenn es **vor deren Mitte** beginnt. Das ist
  // die Regel, nach der auch ein Setzer eine Grundlinie bestimmt.
  let mitte = -Infinity

  for (const flaeche of gesetzt) {
    if (flaeche.oben >= mitte) {
      zeile += 1
      mitte = flaeche.oben + flaeche.hoehe / 2
    }

    // Die Zeile wiegt schwerer als die Spalte; `1e6` ist weiter, als eine
    // Karte je breit wird.
    ordnung.push({ feld: flaeche.feld, platz: zeile * 1e6 + flaeche.links })
  }

  ordnung.sort((links, rechts) => links.platz - rechts.platz)

  return [
    ...ordnung.map((eintrag) => eintrag.feld),
    ...grund.map((flaeche) => flaeche.feld),
  ]
}

// AnteilAlsGrund sagt, ab welchem Anteil an der Kartenfläche ein Feld als
// **Grund** gilt und nicht als Stelle auf ihr.
//
// **0,9 und nicht 1**: bei `magic-m15` decken `card color`, `border color`,
// `foil layer` und `corners` je 375 × 523 von einer Hülle von 375 × 536 —
// 97,6 %. Das nächstkleinere Feld ist die Illustration mit **36 %**;
// dazwischen ist Platz für jeden Stil, der seinen Rahmen um ein paar Punkte
// kleiner zeichnet.
const AnteilAlsGrund = 0.9

// bildfelderAus gibt die Bildfelder einer Vorlage — Werkstück A12 von
// Phase 5.3 (#1451).
//
// **Aus den Formen und nicht aus den Werten der Karte.** Ein Bildfeld, das
// noch leer ist, ist trotzdem eines — und genau in ein leeres will man ein
// erzeugtes Bild setzen.
//
// **Über die Feldbeschreibung und nicht über den Namen**: ein Feld heißt in
// einer Vorlage `mainframe image` und in einer anderen `art`; was es ist,
// steht im Spielpaket.
export function bildfelderAus(
  formen: readonly FeldbeschreibungDTO[],
): string[] {
  return formen.filter((form) => form.art === 'image').map((form) => form.name)
}

// kartennameAus gibt den Namen einer Karte — leer, wenn sie keinen trägt.
//
// **Er belegt die Scryfall-Suche vor** (#1005).
// Er ist ein Vorschlag, keine Festlegung: eine eigene Karte heißt oft anders
// als die, deren Bild man will.
export function kartennameAus(werte: readonly WertDTO[]): string {
  return werte.find((wert) => wert.feld === 'name')?.wert ?? ''
}
