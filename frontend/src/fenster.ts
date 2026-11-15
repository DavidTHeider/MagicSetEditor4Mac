// Die Ereignisse, die die native Menüleiste ins Fenster meldet (#162).
//
// # Warum die Namen hier noch einmal stehen
//
// Sie stehen einmal in `internal/app/fensternaht.go` und einmal hier. Eine
// gemeinsame Quelle gäbe es nur über die erzeugten Bindungen, und die tragen
// Methoden, keine Konstanten. Der Preis ist ein Wächter auf jeder Seite;
// bezahlt wird er dadurch, dass ein Tippfehler sonst eine Anwendung ergibt, in
// der ein Kürzel nichts tut — ohne Fehler, ohne Meldung, ohne roten Lauf.
import { useEffect } from 'react'

export const ereignisEinstellungen = 'menue:einstellungen'
export const ereignisSetSichern = 'menue:set-sichern'
export const ereignisSetSichernUnter = 'menue:set-sichern-unter'

// **Der Bestand, nicht das Set** — die ganze Datenbank
// (#799).
export const ereignisBestandSichern = 'menue:bestand-sichern'

// **Öffnen, Schließen, Drucken und die Hilfe** (#891).
export const ereignisSetNeu = 'menue:set-neu'
export const ereignisSetZuletzt = 'menue:set-zuletzt'
export const ereignisSetOeffnen = 'menue:set-oeffnen'
export const ereignisSetSchliessen = 'menue:set-schliessen'
export const ereignisDrucken = 'menue:drucken'
export const ereignisHilfe = 'menue:hilfe'

// **Das Setmenü und „Über"** — Werkstück 4 von Phase 3.7
// (#1168,
// Befund Z4-01).
//
// **`ereignisSeite` trägt einen Seitenschlüssel**; die übrigen tragen nichts.
export const ereignisRueckgaengig = 'menue:rueckgaengig'

// **Alle Karten anhaken** —
// #1242.
//
// MSEs `ItemList::doSelectAll` wählt **alle Karten**, nicht nur Text; das
// fachliche Gegenstück hier sind die Anhakhaken der Kartenliste.
export const ereignisAllesAuswaehlen = 'menue:alles-auswaehlen'
export const ereignisWiederherstellen = 'menue:wiederherstellen'
export const ereignisSeite = 'menue:seite'
export const ereignisUeber = 'menue:ueber'

// ereignisSeitenleiste klappt die Seitenleiste ein oder aus — Werkstück U14 von
// Phase 5.1 (#1490): der Menüeintrag unter *Darstellung* und ⌥⌘S.
export const ereignisSeitenleiste = 'menue:seitenleiste'

// **Zurück und Vorwärts** unter *Darstellung* — UX-03.AK-02 (#1675).
export const ereignisSeiteZurueck = 'menue:seite-zurueck'
export const ereignisSeiteVor = 'menue:seite-vor'

// **Navigator und Inspector** unter *Darstellung* — UXR-F14 (#2029): ⌥⌘[ und
// ⌥⌘], die Kürzel des Arbeitsplatzes. Gehört wird im `Arbeitsplatz` über
// `useMenueereignis`, nicht über `useFensterereignisse`: nur er weiß, ob die
// Seite eine Spalte hat und wie sie steht.
export const ereignisNavigator = 'menue:navigator'
export const ereignisInspector = 'menue:inspector'

// **Ein Klick im Menü des gewählten Objekts** — GESTE (#2358). Trägt
// `Objektausloesung`; gehört wird im `Objekthandlungsraum`, der weiß, ob die
// Epoche noch gilt.
export const ereignisObjekthandlung = 'menue:objekthandlung'

// **Darstellung zurücksetzen** unter *Darstellung* — UX-09.AK-03 (#1681).
// Gehört wird im `Gestaltungsraum`, nicht über `useFensterereignisse`: er
// hält, was zurückgesetzt wird.
export const ereignisDarstellungZuruecksetzen =
  'menue:darstellung-zuruecksetzen'

// **Ein Doppelklick im Finder** — die Kennung des geöffneten Sets
// (#891).
export const ereignisDateiGeoeffnet = 'datei:geoeffnet'

// **Und wenn sie sich nicht lesen ließ** — Pfad und Fehler.
//
// Der Satz entsteht hier und nicht in Go: Go kennt den Schlüssel, nicht die
// Sprache der Sitzung (§9).
export const ereignisDateiFehler = 'datei:fehler'

// **Der Fortschritt eines Ladevorgangs** (#1100, Stufe 2).
//
// Ein Ereignis und keine Rückgabe, weil ein Abbild von 188 MB länger dauert
// als ein Aufruf. Es trägt zwei Zahlen: gelesen und ganz — und `ganz` ist −1,
// wenn die Gegenstelle keine Länge nennt.
export const ereignisLadefortschritt = 'aktualisierung:fortschritt'

// **Der Stand eines Datei-Laufs** (FMT-03.AK-02, #1692) — ein Auftrag beim
// Beginn, bei jedem gezählten Stück und am Ende; dasselbe Wort wie
// `app.EreignisAuftrag`.
export const ereignisAuftrag = 'auftrag:stand'

// Fensterhandlungen sind die Antworten auf die Einträge der Menüleiste.
export interface Fensterhandlungen {
  einstellungen: () => void
  sichern: () => void
  sichernUnter: () => void
  bestandSichern: () => void
  setNeu: () => void

  // setZuletzt schlägt eines der zuletzt benutzten Sets auf (#891).
  setZuletzt: (setID: string) => void
  setOeffnen: () => void
  setSchliessen: () => void
  drucken: () => void
  hilfe: () => void

  // **Das Setmenü** (#1168). `rueckgaengig` und `wiederherstellen` rufen den
  // **fachlichen** Verlauf — denselben, den ⌘Z ruft.
  rueckgaengig: () => void
  wiederherstellen: () => void

  // seite schlägt eine Lasche auf — Suchen, Ersetzen, Ausgeben.
  seite: (schluessel: string) => void

  // ueber führt auf den Abschnitt „Über" der Hilfe.
  ueber: () => void
  seitenleiste: () => void
  seiteZurueck: () => void
  seiteVor: () => void

  // dateiGeoeffnet bekommt die Kennung des Sets, das der Finder aufschlug,
  // und die Epoche des Bestands, in dem es eingelesen wurde (J05, #2269).
  dateiGeoeffnet: (setID: string, epoche?: number) => void

  // dateiFehler bekommt den Pfad und den Fehler aus Go.
  dateiFehler: (pfad: string, fehler: Fehlerangabe) => void
}

// Fehlerangabe ist der Fehler, wie Go ihn über ein Ereignis meldet.
//
// **Nicht `Bindungsfehler`.** Der entsteht aus einer abgelehnten Bindung; hier
// kommt ein Verbund über die Ereignisnaht, und er trägt dieselben Angaben.
export interface Fehlerangabe {
  schluessel: string
  meldung: string
  werte: Record<string, string> | null
}

// Wailslaufzeit ist der Teil der Laufzeit, den wir benutzen.
//
// **Als eigene Beschreibung und nicht über einen Import der Laufzeit**: sie
// steht im Betrieb am `window` — seit Wails 3 legt `anbindung.ts` sie dorthin
// (ENG-02) —, und im Test steht sie nicht da. Ein fester Import machte aus
// jedem Seitentest einen, der eine Laufzeit braucht.
interface Wailslaufzeit {
  // TitelSetzen setzt den Titel des Fensters — AP-7 (#2364).
  TitelSetzen?: (titel: string) => void

  // **Mit Nutzlast.** Wails reicht die Werte durch, die Go mitgibt; die
  // Menüeinträge geben keine, das Öffnen einer Datei gibt zwei.
  EventsOn?: (
    name: string,
    rueckruf: (...daten: unknown[]) => void,
  ) => () => void

  // OnFileDrop meldet die **absoluten** Pfade gezogener Dateien samt der
  // Stelle, an der losgelassen wurde
  // (#246).
  //
  // **Der zweite Parameter bleibt `false`.** Er ließe Wails eine CSS-Klasse an
  // Elementen mit einer eigenen Eigenschaft verwalten; welches Feld gemeint
  // ist, weiß die Oberfläche selbst — über die Stelle (`bildfeldAn`), seit
  // die Webansicht unter Wails 3 kein `dragover` mehr sieht (ENG-02).
  OnFileDrop?: (
    rueckruf: (x: number, y: number, pfade: string[]) => void,
    ueberZiel: boolean,
  ) => void

  OnFileDropOff?: () => void
}

// laufzeit gibt die Wails-Laufzeit, wenn eine da ist.
export function laufzeit(): Wailslaufzeit | undefined {
  return (globalThis as { runtime?: Wailslaufzeit }).runtime
}

// horche meldet einen Rückruf für ein Wails-Ereignis an — und wieder ab.
//
// **Ohne Laufzeit geschieht nichts, und das ist kein Fehler**: unter `vite
// preview` und in jedem Lauf gibt es keine, und ein Fenster ohne Meldeweg
// zeigt eben keinen fremden Tisch.
//
// **Hier und nicht in einem Haken**, seit zwei Haken dieselbe Naht brauchen:
// der Gast für Tischbild und Zug, das Gespräch für Chat und Präsenz. Zwei
// Kopien liefen beim ersten Unterschied auseinander.
export function horche(
  ereignis: string,
  nimm: (daten: unknown[]) => void,
): () => void {
  const abmelden = laufzeit()?.EventsOn?.(ereignis, (...daten) => {
    nimm(daten)
  })

  return () => {
    abmelden?.()
  }
}

// useFensterereignisse hört auf die Einträge der nativen Menüleiste.
//
// **Ohne Laufzeit tut es nichts, und das ist kein Fehler**: im Browser der
// Vorschau (`npm run vorschau`) gibt es keine Menüleiste, und die Kürzel im
// Fenster gehen ohnehin über `tastatur.ts`. Die Naht ist der **zweite** Weg zu
// derselben Handlung, nicht der einzige.
export function useFensterereignisse(handlungen: Fensterhandlungen): void {
  const {
    einstellungen,
    sichern,
    sichernUnter,
    bestandSichern,
    setNeu,
    setZuletzt,
    setOeffnen,
    setSchliessen,
    drucken,
    hilfe,
    rueckgaengig,
    wiederherstellen,
    seite,
    ueber,
    seitenleiste,
    seiteZurueck,
    seiteVor,
    dateiGeoeffnet,
    dateiFehler,
  } = handlungen

  useEffect(() => {
    const laufende = laufzeit()
    if (!laufende?.EventsOn) {
      return
    }

    const abmelden = [
      laufende.EventsOn(ereignisEinstellungen, einstellungen),
      laufende.EventsOn(ereignisSetSichern, sichern),
      laufende.EventsOn(ereignisSetSichernUnter, sichernUnter),
      laufende.EventsOn(ereignisBestandSichern, bestandSichern),
      laufende.EventsOn(ereignisSetNeu, setNeu),

      // **Mit Nutzlast** — welches Set gemeint ist, weiß der Menüeintrag.
      laufende.EventsOn(ereignisSetZuletzt, (...daten) => {
        const [kennung] = daten
        if (typeof kennung === 'string' && kennung !== '') {
          setZuletzt(kennung)
        }
      }),
      laufende.EventsOn(ereignisSetOeffnen, setOeffnen),
      laufende.EventsOn(ereignisSetSchliessen, setSchliessen),
      laufende.EventsOn(ereignisDrucken, drucken),
      laufende.EventsOn(ereignisHilfe, hilfe),

      // **Die Rücknahme über die Leiste** (#1168). Sie ruft **denselben**
      // Verlauf wie ⌘Z — das ist die Zusage des Werkstücks, und sie ist hier
      // eingelöst: das Fenster hält ihn schon, und der Menüeintrag greift
      // danach statt neben ihn.
      laufende.EventsOn(ereignisRueckgaengig, rueckgaengig),
      laufende.EventsOn(ereignisWiederherstellen, wiederherstellen),
      laufende.EventsOn(ereignisUeber, ueber),
      laufende.EventsOn(ereignisSeitenleiste, seitenleiste),
      laufende.EventsOn(ereignisSeiteZurueck, seiteZurueck),
      laufende.EventsOn(ereignisSeiteVor, seiteVor),

      // **Die Nutzlast wird geprüft, nicht geglaubt** — dieselbe Regel wie
      // beim zuletzt benutzten Set darüber.
      //
      // **Hier nur die Art und nicht der Wert**, anders als dort: ob die
      // Zeichenkette eine Seite *benennt*, weiß der Aufrufer, und er prüft es
      // (`Anwendung.tsx`). Ein `!== ''` daneben wäre dieselbe Zusage an zwei
      // Stellen (§10) — und ein Zweig, den kein Lauf von seinem Nachbarn
      // unterscheiden kann.
      laufende.EventsOn(ereignisSeite, (...daten) => {
        const [schluessel] = daten
        if (typeof schluessel === 'string') {
          seite(schluessel)
        }
      }),

      // **Die Nutzlast wird geprüft, nicht geglaubt.** Sie kommt über die
      // Ereignisnaht als `unknown`; ein blindes `as` machte aus einer
      // geänderten Go-Seite eine Seite, die stumm nichts tut.
      laufende.EventsOn(ereignisDateiGeoeffnet, (...daten) => {
        const [kennung, epoche] = daten
        if (typeof kennung === 'string' && kennung !== '') {
          dateiGeoeffnet(
            kennung,
            typeof epoche === 'number' ? epoche : undefined,
          )
        }
      }),

      laufende.EventsOn(ereignisDateiFehler, (...daten) => {
        const [pfad, fehler] = daten
        if (typeof pfad === 'string' && istFehlerangabe(fehler)) {
          dateiFehler(pfad, fehler)
        }
      }),
    ]

    return () => {
      for (const eines of abmelden) {
        eines()
      }
    }
  }, [
    einstellungen,
    sichern,
    sichernUnter,
    bestandSichern,
    setNeu,
    setZuletzt,
    setOeffnen,
    setSchliessen,
    drucken,
    hilfe,
    rueckgaengig,
    wiederherstellen,
    seite,
    ueber,
    seitenleiste,
    seiteZurueck,
    seiteVor,
    dateiGeoeffnet,
    dateiFehler,
  ])
}

// istFehlerangabe sagt, ob eine Nutzlast die Gestalt eines Go-Fehlers hat.
function istFehlerangabe(wert: unknown): wert is Fehlerangabe {
  return (
    typeof wert === 'object' &&
    wert !== null &&
    typeof (wert as { schluessel?: unknown }).schluessel === 'string'
  )
}

// useMenueereignis meldet **einen** Horcher an der nativen Menüleiste an.
//
// # Warum es das neben [useFensterereignisse] gibt
//
// **Weil der Zustand nicht immer der Anwendung gehört.** Die Anhakhaken der
// Kartenliste liegen in `Karten.tsx`; ein Rückruf dafür in
// [Fensterhandlungen] müsste durch zwei Ebenen gereicht werden, und die
// Anwendung trüge einen Zustand, den sie nicht benutzt
// (#1242).
//
// **Der Horcher hängt so lange, wie die Seite steht** — wer die Karten
// verlässt, meldet ab. Das ist richtig: das Ereignis meint die Liste, die
// gerade da ist.
//
// # Ohne Laufzeit tut es nichts
//
// Wie [useFensterereignisse] und aus demselben Grund: im Browser der Vorschau
// gibt es keine Menüleiste.
export function useMenueereignis(name: string, handlung: () => void): void {
  useEffect(() => {
    const laufende = laufzeit()
    if (!laufende?.EventsOn) {
      return
    }

    return laufende.EventsOn(name, handlung)
  }, [name, handlung])
}
