// Die Vorrichtung, die jede Seite im Rahmen zeigt.
//
// **Eine eigene Datei**, weil zwei Prüfdateien sie brauchen und weil
// `barrierefreiheit.test.tsx` sonst über die 400 Zeilen aus §2 liefe. Der
// Schnitt folgt einer Frage: hier steht, **wie** eine Seite entsteht — dort
// und in `seitensprache.test.tsx`, **was** von ihr verlangt wird.
//
// Getrennt am 2026-08-22.
import {
  stillePflege,
  stilleDatenpflege,
  stillerRuecksprung,
} from './bestandspflegeprobe'
import { render } from '@testing-library/react'
import { vi } from 'vitest'

import type { Seitenschluessel } from '../navigation'
import { Rahmen } from '../rahmen/Rahmen'
import { Seiteninhalt } from '../seiten/Seiteninhalt'
import { Einstellungsraum } from '../haken/einstellungsraum'
import type { Sprache } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import type { EinstellungDTO } from '../api'
import type { Einstellungsstand } from '../haken/useEinstellungen'
import { stillerSpeicherlauf } from './speicherlauf'
import { belegeKartenwege, probeZeile } from './kartenproben'
import { belegeApi } from './api'
import { belegeAnmeldung } from './anmeldebelegung'
import { MitOffenemSet } from './offenesset'
import { seitenteile } from './seitenteile'

// einstellungszeilen sind der Stand, den die Einstellungsseite im Betrieb
// bekommt — dieselben Schlüssel und Werte wie `internal/domain/einstellung.go`.
//
// # Warum sie nicht leer sind
//
// **Weil eine leere Vorrichtung eine leere Seite prüft.** Mit
// `eintraege: []` zeichnete die Einstellungsseite in
// `barrierefreiheit.test.tsx` und `auszeichnung.test.tsx` ihren **leeren
// Zustand**, und beide Prüfungen wären grün über eine Seite, die es so nicht
// gibt — die Dienstschalter, die Wahl je Frage und die Zugangsdaten sähe axe
// **nie**.
//
// Das ist genau die Falle, die der Absatz über `belegeProbeset` unten für
// #288
// beschreibt: eine grüne Prüfung über eine Seite, die nicht gerendert wurde.
const einstellungszeilen: EinstellungDTO[] = [
  {
    schluessel: 'sprache',
    wert: 'de',
    wirksam: 'de',
    werte: ['de', 'en'],
    bereich: 'anwendung',
  },
  {
    schluessel: 'erscheinungsbild',
    wert: 'system',
    wirksam: 'system',
    werte: ['system', 'hell', 'dunkel'],
    bereich: 'anwendung',
  },
  {
    schluessel: 'schriftgroesse',
    wert: 'normal',
    wirksam: 'normal',
    werte: ['klein', 'normal', 'gross', 'sehrgross'],
    bereich: 'anwendung',
  },
  {
    schluessel: 'aktualisierung',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'melden', 'holen'],
    bereich: 'anwendung',
  },
]

const dienstzeilen: EinstellungDTO[] = (
  [
    'scryfall',
    'mtgapi',
    'mtgjson',
    'scrydex',
    'cardtrader',
    'anthropic',
    'openai',
    'xai',
  ] as const
).map((kennung) => ({
  schluessel: `dienst.${kennung}`,
  wert: kennung === 'scryfall' ? 'an' : 'aus',
  wirksam: kennung === 'scryfall' ? 'an' : 'aus',
  werte: ['aus', 'an'],
  bereich: 'dienste',
}))

const fragezeilen: EinstellungDTO[] = [
  {
    schluessel: 'frage.suche',
    wert: 'scryfall',
    wirksam: 'scryfall',
    werte: ['scryfall', 'mtgapi', 'mtgjson', 'scrydex', 'cardtrader'],
    bereich: 'dienste',
  },
]

// **Die der KI** — seit A7, #1446. Sie stehen im Bereich
// `dienste` und sind weder Schalter eines Kartendienstes noch eine Wahl.
const kizeilen: EinstellungDTO[] = [
  {
    schluessel: 'ki.laufzeit',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
  // **Apples Gerätemodell** — AI-05, #2042. Es fängt **automatisch** an:
  // wer es auf dem Mac aktiviert hat, bekommt es ohne weiteren Schritt.
  {
    schluessel: 'ki.geraetemodell',
    wert: 'automatisch',
    wirksam: 'automatisch',
    werte: ['automatisch', 'aus'],
    bereich: 'dienste',
  },
  // **Das Modell des Vorrangs bei Ollama** — #2054. Leer heißt: das erste,
  // das in fester Form antworten kann.
  {
    schluessel: 'ki.vorrangmodell',
    wert: '',
    wirksam: '',
    werte: [],
    bereich: 'dienste',
  },
  {
    schluessel: 'ki.adresse',
    wert: '',
    wirksam: '',
    werte: [],
    bereich: 'dienste',
  },

  // **Und die Kontoschau** — T11, #1476. Sie steht im selben Bereich und
  // fängt ebenso aus an.
  {
    schluessel: 'konto.schau',
    wert: 'aus',
    wirksam: 'aus',
    werte: ['aus', 'an'],
    bereich: 'dienste',
  },
]

export const einstellungsprobe: Einstellungsstand = {
  wert: (schluessel) =>
    [...einstellungszeilen, ...dienstzeilen, ...fragezeilen, ...kizeilen].find(
      (zeile) => zeile.schluessel === schluessel,
    )?.wirksam ?? '',
  eintraege: [
    ...einstellungszeilen,
    ...dienstzeilen,
    ...fragezeilen,
    ...kizeilen,
  ],
  imBereich: (bereich) =>
    einstellungszeilen.filter((zeile) => zeile.bereich === bereich),
  // **Leer, und das ist die Vorgabe dieser Vorrichtung** — Werkstück 9 von
  // Phase 4.12 (#1351): wer die Vorgaben meint, belegt sie selbst.
  vorgaben: () => [],
  dienstschalter: () => dienstzeilen,
  kizeilen: () => kizeilen,
  fragewahlen: () => fragezeilen,
  geladen: true,
  ursache: undefined,
  fehlschlag: undefined,
  setze: () => undefined,
  laeuft: () => false,
}

// probeset belegt die Fassade so weit, dass jede Lasche etwas zu zeigen hat.
//
// **Mit offenem Set und nicht ohne.** Bis
// #288 lief
// die Prüfung mit `setID={undefined}`; seither zeigt der Arbeitsplatz dann
// seinen leeren Zustand, und **fünf der zehn Seiten** wären gar nicht geprüft
// worden. Eine grüne Prüfung über eine Seite, die nicht gerendert wurde, ist
// genau die zu kurze Aufzählung, gegen die diese Datei geschrieben ist.
// **`leer` gibt dasselbe Set ohne eine Karte** — G02 (#2249): ein leeres
// Set ist ein eigener Leerzustand jeder Setseite, und erst er zeigt, ob sie
// erklärt, was fehlt.
export function belegeProbeset({ leer = false }: { leer?: boolean } = {}) {
  const karten = leer ? 0 : 1

  belegeApi({
    // **Und was die übrigen Seiten lesen** (G02) — `seitenteile.ts`.
    ...seitenteile(),

    // **Und was die Nutzerseite braucht** — Werkstück 8 bis 15 von Phase 4.12
    // (#1351).
    //
    // **Ohne sie zeigte sie „Die Methode Nutzer ist nicht belegt." und hatte
    // keine einzige Lasche.** Jeder Wächter, der über `seiten` läuft — die
    // Auszeichnungen und die Barrierefreiheit —, prüfte damit seit ihrer
    // Entstehung eine Fehlermeldung statt einer Seite. Genau die Falle, die
    // zwanzig Zeilen weiter unten für das offene Set beschrieben ist, und sie
    // ist wieder zugeschnappt.
    ...belegeAnmeldung(),

    // **Die Zugangsstände, die es im Betrieb gibt** — nicht die leere Liste
    // aus `ohneZugaenge`.
    //
    // **Sonst zeichnet `Zugaenge` gar nichts**, und damit prüfen axe und
    // `auszeichnung.test.tsx` einen Bereich, den es auf dem Bildschirm gibt
    // und in der Vorrichtung nicht — dieselbe Falle, die zwanzig Zeilen
    // weiter oben für `einstellungsprobe` beschrieben ist.
    //
    // **Zwei Dienste brauchen Zugangsdaten** (`internal/domain/dienste.go`):
    // MTGJSON eine Marke, Scrydex Schlüssel und Team-Kennung.
    Zugaenge: () =>
      Promise.resolve({
        zugaenge: [
          {
            dienst: 'mtgjson',
            angaben: ['marke'],
            hinterlegt: [],
            vollstaendig: false,
          },
          {
            dienst: 'scrydex',
            angaben: ['schluessel', 'team'],
            hinterlegt: ['schluessel'],
            vollstaendig: false,
          },
        ],
        fehler: null,
      }),

    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '7',
            name: 'Probeset',
            spiel: 'magic',
            stil: 'magic-m15',
            karten,
            herkunft: '/a.mse-set',
            herkunftFehlt: false,
          },
        ],
      }),
    Karten: () =>
      Promise.resolve({
        karten: leer
          ? []
          : [
              {
                id: '1',
                nummer: 0,
                werte: [{ feld: 'name', wert: 'Feuerball' }],
              },
            ],
      }),
    ...belegeKartenwege(leer ? [] : [probeZeile({ id: '1' })]),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Vorschau: () => Promise.resolve({ vorschau: null }),
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Statistik: () =>
      Promise.resolve({
        statistik: { karten, felder: karten, leere: 0, belegung: [] },
      }),
    Verteilung: () => Promise.resolve({ verteilung: [] }),
    Suchfelder: () => Promise.resolve({ felder: [] }),
    Bildnamen: () => Promise.resolve({ namen: [] }),

    // **Die Deckseite** (#1293). Sie holt beim Erscheinen drei Listen — ohne
    // Beleg scheitert die Fassade mit dem Namen der Methode, und das ist die
    // Zusage von `test/api.ts`: kein `undefined`, sondern ein benannter
    // Fehlschlag.
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
  })
}

// zeigeSeite rendert eine Seite im Rahmen, wie die Anwendung es tut.
//
// **Als Helfer und nicht dreimal abgeschrieben**: die drei Prüfungen hier
// sollen denselben Baum sehen. Wo der Aufbau dreimal dasteht, sieht eine davon
// irgendwann einen anderen — und dann prüft sie etwas, das es nicht gibt.
export function zeigeSeite(
  schluessel: Seitenschluessel,
  sprache: Sprache,

  // wechseln ist der Seitenwechsel, den die Seite bekommt.
  //
  // **Als Parameter, damit ein Lauf ihn ansehen kann**: seit UX-02 führt der
  // Kopf einer Seite in die übrigen Seiten ihres Bereichs, und ohne diesen
  // Spion ließe sich nicht prüfen, wohin.
  wechseln: (seite: Seitenschluessel) => void = vi.fn(),

  // raum ist ein geteilter Einstellungsstand — für die Läufe, die sehen
  // wollen, was ein Schalter verbirgt (OPS-10, #1908).
  raum?: Einstellungsstand,
) {
  return render(
    <Sprachraum sprache={sprache}>
      <Einstellungsraum stand={raum}>
        {/* **Das offene Set liest das Fenster** (`Anwendung`) und reicht es
            hinunter — ohne diesen Raum stand der Arbeitsplatz hier „ohne
            Namen", und keine Lasche wusste, wie viele Karten es hat (G02). */}
        <MitOffenemSet setID="7">
          <Rahmen aktuell={schluessel} wechseln={wechseln}>
            <Seiteninhalt
              speichern={stillerSpeicherlauf()}
              wechseln={wechseln}
              fassung="0.0.0-test"
              zuUeber={false}
              ueberAngekommen={vi.fn()}
              einfuehrungZeigen={vi.fn()}
              bestandGeaendert={vi.fn()}
              seite={schluessel}
              setID="7"
              karteID={undefined}
              einstellungen={einstellungsprobe}
              bestandspflege={stillePflege()}
              ruecksprung={stillerRuecksprung()}
              pflege={stilleDatenpflege()}
              meldefehler={undefined}
              oeffnungsanstoss={false}
              oeffnungAngestossen={vi.fn()}
              neuanstoss={false}
              neuAngestossen={vi.fn()}
              bestandsstand={0}
              verlaufsstand={0}
              verlaufsfehler={undefined}
              waehleSet={vi.fn()}
              setEntfernt={vi.fn()}
              oeffneKarte={vi.fn()}
              oeffneSymbol={vi.fn()}
            />
          </Rahmen>
        </MitOffenemSet>
      </Einstellungsraum>
    </Sprachraum>,
  )
}

// dokumentDer zeichnet eine Seite wie [zeigeSeite] und gibt **das ganze
// Dokument** zurück, nicht den Ausschnitt: seit UX-08 (#1680) stehen sieben
// Setseiten in einer Schublade, und die zeichnet Mantine in ein Portal
// daneben. Ein Wächter über `container` sähe von *Stil* nur die Karten
// darunter und hieße grün.
export function dokumentDer(
  seite: Seitenschluessel,
  sprache: Sprache,
): HTMLElement {
  return zeigeSeite(seite, sprache).baseElement
}
