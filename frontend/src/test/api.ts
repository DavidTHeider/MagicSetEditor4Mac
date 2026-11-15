// Die Fassade für Tests (AGENTS.md §3).
//
// **Teilweise belegt, und was fehlt, meldet sich.** Ein Test belegt genau die
// Methoden, die sein Fall braucht; jede andere scheitert mit ihrem Namen statt
// `undefined` zu liefern.
//
// Der Grund steht in §3 unter „eine Vorrichtung, die großzügiger ist als der
// Betrieb": eine Fassade, die stillschweigend nichts zurückgibt, macht aus einem
// vergessenen Beleg eine leere Seite — und dann sucht jemand im Zustand statt in
// der Vorrichtung.
import { act } from '@testing-library/react'
import { epocheVergessen } from '../bestandsepoche'

import { api, fassade, type Api } from '../api'
import {
  ohneAnbietergeheimnisse,
  ohneZugaenge,
  vorgabegrenzen,
} from './apigrenzen'
import { ohneBefehle, stillerAbbruch, stillesErgebnis } from './apibefehle'
import { ohneBogenprobelauf, ohneVorlagenprobelauf } from './apiprobelauf'
import { angemeldeteSitzung } from './apisitzung'
import {
  ohneDeckbezuege,
  ohneDienstkatalog,
  stillerVermerk,
  uebersichtsvorgaben,
} from './apiuebersicht'

interface MitGo {
  go?: { app?: { Bindungen?: Partial<Api> } }
}

// antwortAbwarten wartet, bis eine Antwort der Fassade angekommen **ist**.
//
// **Das Gegenstück zur Verzögerung in [belegeApi].** Wer dort eine
// Makroaufgabe später antwortet, kommt mit `await Promise.resolve()` nicht
// weiter: das leert die Mikroaufgaben, und die Antwort steht in der Schlange
// dahinter.
//
// # Wofür man es braucht und wofür nicht
//
// **Für eine Verneinung.** „Es wurde nichts eingetragen" ist erst dann eine
// Aussage, wenn die Antwort da war — vorher trifft sie ohnehin zu (§3). Für
// eine **Bejahung** ist `waitFor` das bessere Werkzeug: es kehrt zurück,
// sobald der Fall eintritt, statt eine feste Frist zu warten.
export async function antwortAbwarten(): Promise<void> {
  await act(async () => {
    await new Promise((fertig) => {
      setTimeout(fertig, 0)
    })
  })
}

// Ende ist ein festgehaltener Ausgang — geglückt oder gescheitert.
type Ende = { gut: true; wert: unknown } | { gut: false; grund: unknown }

// fangen ruft eine Methode und hält ihren Ausgang fest, ohne ihn zu werfen.
//
// **Sofort und nicht später** — siehe [belegeApi]. Ein synchroner Wurf zählt
// dabei wie eine Ablehnung: die Fassade sagt `Promise<T>` zu, und ein
// `.catch()` am Aufrufer fängt einen synchronen Wurf nicht.
function fangen(
  methode: (...werte: unknown[]) => unknown,
  werte: unknown[],
): { ende: Promise<Ende>; roh: unknown } {
  try {
    const roh = methode(...werte)

    return {
      roh,
      ende: Promise.resolve(roh).then(
        (wert): Ende => ({ gut: true, wert }),
        (grund: unknown): Ende => ({ gut: false, grund }),
      ),
    }
  } catch (grund) {
    return { roh: undefined, ende: Promise.resolve({ gut: false, grund }) }
  }
}

// abbrechbarWie gibt der verzögerten Antwort das `cancel` der rohen — **wie
// die Wails-Laufzeit**, deren Rufe abbrechbar sind (BK-07d, #2118). Ohne das
// verlöre die Vorrichtung den Abbruch, den der Betrieb weiterreicht.
function abbrechbarWie<T>(antwort: Promise<T>, roh: unknown): Promise<T> {
  const abbruch = (roh as { cancel?: unknown } | undefined)?.cancel

  if (typeof abbruch === 'function') {
    ;(antwort as Promise<T> & { cancel?: () => unknown }).cancel = () =>
      (abbruch as () => unknown).call(roh)
  }

  return antwort
}

// belegeApi legt die genannten Methoden ins Fenster und gibt sie zurück.
//
// Der Rückgabewert ist die **vollständige** Fassade: die belegten Methoden und
// für alle übrigen eine, die sich meldet.
//
// # Jede Antwort kommt eine Makroaufgabe später
//
// **Weil sie das im Betrieb auch tut.** Ein Aufruf geht über die Wails-IPC;
// zwischen dem Ruf und der Antwort liegt immer mindestens eine Aufgabe. Eine
// Vorrichtung, die synchron antwortet, ist großzügiger als der Betrieb (§3) —
// und dann ist jeder Griff, der zu früh kommt, trotzdem grün.
//
// **Gemessen am 2026-09-02** (#1321):
// mit der Verzögerung fielen **20 von 2.186** Tests in fünf Dateien um. Jeder
// von ihnen griff in einen Baum, den es noch nicht gab — eine Auswahlliste vor
// ihren Möglichkeiten, ein Kästchen vor seinen Zeilen, ein Feld, das eine
// spätere Antwort neu gezeichnet hatte.
//
// **Der Anlass war ein Flackerer**, kein Verdacht: `formatzeile.test.tsx` war
// einer von acht vollen Läufen rot, mit „Unable to find a label with the text
// of: Sideboard ohne Grenze". Der Test wartete auf einen Bereich, den es auch
// im **leeren** Zustand gibt.
//
// **Gerufen wird sofort, geantwortet später.** Wer den Aufruf mitverzögerte,
// misst etwas anderes: ein Spion auf einer Bindung wäre dann noch nicht
// gerufen, und acht weitere Tests fielen um, ohne dass etwas daran falsch
// wäre. Der erste Anlauf hat genau das getan.
//
// **Das Gegenstück zum Warten heißt [antwortAbwarten].**
export function belegeApi(teile: Partial<Api>): Api {
  const gebaut = fassade(
    Object.fromEntries(
      Object.entries({
        AutoErsetzen: ohneErsetzung,
        // **Die Setliste lädt die Pakete für ein neues Set** — Beiwerk für
        // jeden Lauf über die Liste. Ohne Beleg stünde seit dem Assistenten
        // (UX-05, #1677) ein zweiter Ladefehler neben dem, um den es geht.
        Pakete: ohnePakete,
        // **Und die Setangaben fragen, was dem Set an Paketen fehlt**
        // (DOM-01.AK-04, #1696) — die häufigste Antwort: nichts.
        Paketbedarf: ohnePaketbedarf,
        // **Der Handelsraum fragt nach den Handeln mit dem Gegenüber** —
        // Beiwerk für jeden Lauf, der eines aufschlägt (UX-08.AK-04, #1680).
        HandelVorgaenge: ohneHandel,
        // **Und der Bereich Handeln nach allen** (TRADE-01, #1715).
        HandelUebersicht: ohneHandelsliste,
        Zugaenge: ohneZugaenge,
        ...uebersichtsvorgaben,
        DeckBenutzt: stillerVermerk,
        Dienstkatalog: ohneDienstkatalog,
        ExemplarDecks: ohneDeckbezuege,
        Grenzen: vorgabegrenzen,
        Befehlskatalog: ohneBefehle,
        BefehlAusfuehren: stillesErgebnis,
        BefehlAbbrechen: stillerAbbruch,
        Anbietergeheimnisse: ohneAnbietergeheimnisse,
        Platz: ohnePlatz,
        // **Schalterreife und Wiederanlauf** (OPS-10 #1908, F04 #2343) fragt
        // der Rahmen bei jedem Zeichnen; meist ohne Marke und ohne Abbruch.
        Schalter: () => ({ schalter: [] }),
        Wiederanlauf: () => ({ abbrueche: [], fehler: null }),
        Sitzungsstand: angemeldeteSitzung,
        Ruecksetzungsarten: ohneRuecksetzungsarten,
        Kostenstand: ohneKostenstand,

        // **Der Ausgabeassistent fragt „Wie es wird"** (UXR-F7, #2015) —
        // beim Aufschlagen jedes Ziels und bei jeder Wahl.
        BogenProbelauf: ohneBogenprobelauf,
        VorlageProbelauf: ohneVorlagenprobelauf,

        // **Das Fenster der Sammlung reicht durch, was `Sammlung` belegt** —
        // Werkstück #1500.
        //
        // **Ohne die Sicht nachzubauen.** Filtern und Ordnen geschehen in Go;
        // sie hier ein zweites Mal zu schreiben wäre eine zweite Wahrheit, und
        // ein Test liefe dann gegen sie statt gegen die echte. Wer prüft, dass
        // eine Sicht **wirkt**, belegt `Sammlungsfenster` selbst und sieht
        // nach, was hereinkam.
        Sammlungsfenster: durchgereichteSammlung,
        Kostenposten: ohneKostenposten,
        Tokenablaufwahlen: ohneAblaufwahlen,
        Arbeitskontext: ohneArbeitskontext,
        Nutzer: ohneNutzer,
        Nutzerbilder: ohneNutzerbilder,

        // **Der Gestaltungsraum fragt nach dem Fensterbild** (UX-09, #1681),
        // bei jedem Zeichnen der Anwendung. Die häufigste Antwort: keines.
        Fensterbild: ohneFensterbild,

        // **Der Turnierassistent fragt nach den Formaten** (DOM-04, #1699),
        // sobald die Turnierseite steht. Die häufigste Antwort: keine.
        Formate: ohneFormate,

        // **Der Karteneditor fragt nach den Feldern des Stils** (#2298) — bei
        // jeder Karte. Die häufigste Antwort: keine.
        Kartenzusatzfelder: ohneZusatzfelder,
        ...teile,
      }).map(([name, methode]) => [
        name,
        (...werte: unknown[]) => {
          // **Gerufen wird sofort, geantwortet wird später** — so wie über
          // die IPC. Wer den Aufruf mitverzögert, misst etwas anderes.
          //
          // **Und der Fänger hängt sofort dran**, nicht erst im `setTimeout`:
          // eine Zusage, die abgelehnt wird, während niemand zuhört, meldet
          // Node als unbehandelte Ablehnung. Der erste Anlauf tat das, und
          // der Lauf war rot bei 2.190 grünen Tests und 25 Fehlern daneben.
          const gefangen = fangen(
            methode as (...w: unknown[]) => unknown,
            werte,
          )

          return abbrechbarWie(
            new Promise((fertig, gescheitert) => {
              setTimeout(() => {
                void gefangen.ende.then((ende) => {
                  if (ende.gut) {
                    fertig(ende.wert)
                  } else {
                    // **Weitergereicht wie erhalten.** Womit ein Beleg ablehnt,
                    // entscheidet der Test — die Vorrichtung darf daraus keinen
                    // `Error` machen, sonst prüfte niemand mehr, was die Seite
                    // mit einer Ablehnung ohne `Error` tut.
                    // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
                    gescheitert(ende.grund)
                  }
                })
              }, 0)
            }),
            gefangen.roh,
          )
        },
      ]),
    ),
  )

  ;(globalThis as unknown as MitGo).go = { app: { Bindungen: gebaut } }

  return gebaut
}

// ohneErsetzung ist die Vorbelegung für `AutoErsetzen`.
//
// # Warum diese eine Methode vorbelegt ist und die anderen nicht
//
// Sie wird bei **jedem Tastendruck** gerufen
// (#897) —
// nicht, weil ein Test sie meint, sondern weil das Feld sie fragt. Ohne
// Vorbelegung scheiterte jeder Test, der irgendwo tippt, an einer Methode, die
// mit seinem Fall nichts zu tun hat.
//
// **Sie gibt zurück, was hereinkam** — der ehrliche Fall „das Spiel hat keine
// Regeln", und derselbe, den die Go-Seite ohne Spielpaket gibt. Ein Test, der
// die Ersetzung **meint**, belegt sie selbst und überschreibt diese hier.
//
// **Das ist die eine Ausnahme von „was fehlt, meldet sich"**, und sie steht
// hier mit ihrem Grund, statt die Regel stillschweigend aufzuweichen.
function ohneErsetzung(_setID: string, text: string, stelle: number) {
  return Promise.resolve({ text, stelle, fehler: null })
}

// durchgereichteSammlung gibt, was `Sammlung` gibt — samt beider Zahlen.
const durchgereichteSammlung = async () => {
  const antwort = await api().Sammlung()

  return {
    exemplare: antwort.exemplare,
    gesamt: antwort.exemplare.length,
    insgesamt: antwort.exemplare.length,
    fehler: antwort.fehler,
  }
}

// ohneKostenstand ist die Vorbelegung für `Kostenstand` — Werkstück O12 von
// Phase 5.2
// (#1439).
//
// **Aus demselben Grund wie [ohnePlatz]**: die Bestandslasche fragt sie beim
// Aufschlagen. **Alles null und ohne Währung** — ein Konto ohne Freigabe hat
// keine, und eine erfundene wäre schlimmer als keine.
const leererBetrag = { wert: 0, waehrung: '' }

const ohneKostenstand = (_konto: string, zeitraum: string) =>
  Promise.resolve({
    konto: 'anbieter',
    zeitraum: zeitraum === '' ? 'tag' : zeitraum,
    grenze: leererBetrag,
    reserviert: leererBetrag,
    verbraucht: leererBetrag,
    unklar: leererBetrag,
    verfuegbar: leererBetrag,
    fehler: null,
  })

const ohneKostenposten = () => Promise.resolve({ posten: [], fehler: null })

// ohneRuecksetzungsarten ist die Vorbelegung für `Ruecksetzungsarten` —
// Werkstück O4 von Phase 5.2
// (#1428).
//
// **Aus demselben Grund wie [ohnePlatz]**: die Bestandslasche fragt sie beim
// Aufschlagen, und ohne Beleg stünde in jedem Lauf ein zweiter Fehlerkasten.
// **Die sechs echten und keine leere Liste**: eine leere Auswahl ist ein
// Zustand, den es im Betrieb nicht gibt.
const ohneRuecksetzungsarten = () =>
  Promise.resolve([
    'layout',
    'profileinstellungen',
    'nutzereinstellungen',
    'profil',
    'nutzer',
    'anwendung',
  ])

// ohnePlatz ist die Vorbelegung für `Platz` — Werkstück 4 von Phase 4.13
// (#1352).
//
// **Aus demselben Grund wie [ohneZugaenge]**: die Bestandslasche fragt sie
// beim Aufschlagen.
// ohneAblaufwahlen ist die Antwort einer Uhr ohne Vorschläge — die Zugriffstoken
// fragen sie beim Aufbau (Werkstück O2 von Phase 5.2, #1426), und ohne Beleg
// meldete jede Seite mit einem technischen Profil eine unbelegte Methode.
function ohneAblaufwahlen() {
  return Promise.resolve({ wahlen: [], fehler: null })
}

// **Die Kontextleiste hängt an jeder Anwendung** — Werkstück U4 von Phase 5.1
// (#1415): wer `App` zeichnet, ohne sie zu belegen, bekäme sonst in jedem Lauf
// einen Fehlerkasten „Arbeitskontext ist nicht belegt". Die drei Vorgaben sind
// ein Mensch allein in seinem Profil, ohne Bild.
const ohneArbeitskontext = () =>
  Promise.resolve({
    nutzerID: '1',
    profilID: '1',
    eignerID: '1',
    rolle: 'verwalter',
    nurLesend: false,
    revision: 0,
    fehler: null,
  })

const ohneNutzer = () => Promise.resolve({ nutzer: [], fehler: null })
const ohnePaketbedarf = () => Promise.resolve({ befunde: [] })
const ohneZusatzfelder = () => Promise.resolve({ felder: [], fehler: null })

const ohnePakete = () =>
  Promise.resolve({ pakete: [], befunde: [], fehler: null })
const ohneHandel = () => Promise.resolve({ vorgaenge: [], fehler: null })
const ohneHandelsliste = () => Promise.resolve({ zeilen: [], fehler: null })

const ohneNutzerbilder = () =>
  Promise.resolve({ bilder: {}, badges: {}, fehler: null })
const ohneFensterbild = () => Promise.resolve({ adresse: '', fehler: null })
const ohneFormate = () => Promise.resolve({ formate: [], fehler: null })

function ohnePlatz() {
  return Promise.resolve({ posten: [], fehler: null })
}

// raeumeApi nimmt die Belegung wieder weg.
//
// **In `afterEach`**, sonst sieht der nächste Test die Belegung des vorigen —
// und ein Test, der grün ist, weil ein anderer vor ihm lief, ist keiner.
export function raeumeApi(): void {
  delete (globalThis as unknown as MitGo).go
  // **Ein frisches Fenster kennt keine Epoche** (J05, #2269).
  epocheVergessen()
}

// belegeFenster legt, was **jedes** Zeichnen von `App` braucht.
//
// # Warum es das gibt
//
// `App` fragt beim Zeichnen zwei Dinge, ohne dass ein Test sie erwähnt: die
// Einstellungen (seit #245 zeichnet sie erst, wenn die Sprache feststeht) und
// `SpracheSetzen` — die Naht zur nativen Menüleiste, gerufen aus einem Effekt.
//
// **Die zweite fehlte in sechs von acht App-Tests**, und das ist eine
// Fragilität für sich: die Fassade lässt eine unbelegte Methode **scheitern**
// (`fassade.ts`, Zeile 109), und die Tests hingen daran, dass diese Ablehnung
// unsichtbar bleibt. Wer morgen den Meldeweg ändert, macht sechs Tests rot,
// ohne etwas kaputtgemacht zu haben.
//
// **Was sie nicht ist: die Behebung eines Flackers.** Gefunden wurde sie beim
// Nachgehen eines roten Laufs in dreiundzwanzig
// (`appsichern.test.tsx > tut nichts, solange kein Set offen ist`). Die
// Vermutung, die unbelegte Methode sei die Ursache, hat sich **nicht**
// bestätigt: ohne den Beleg laufen die sieben Fälle der Datei sieben von
// sieben Mal durch. Der Flacker steht als eigenes Werkstück, und dieser Helfer
// behauptet nicht, ihn zu beheben (§10).
//
// **Sie steht hier und nicht in jeder Datei**, damit der nächste App-Test sie
// nicht vergessen kann — hineingestreut, damit die eigenen Belegungen sie
// weiterhin überschreiben können.

// **`belegeFenster` steht in `fensterbelegung.ts`** und wird hier
// weitergereicht (O8, #1432): die Datei ging mit ihr über die Zeilengrenze aus
// §2, und jeder Test führt sie von hier ein — den Einstiegspunkt zu ändern
// hieße, in vierhundert Dateien einen Pfad zu tauschen, ohne dass sich etwas
// ändert.
export { belegeFenster } from './fensterbelegung'
