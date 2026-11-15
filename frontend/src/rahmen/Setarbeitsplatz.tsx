// Der Arbeitsplatz eines Sets: sein Name, seine Laschen, sein Inhalt.
//
// # Warum es ihn gibt
//
// Fünf der neun Seiten sprechen über das **offene Set** — Karten, Suche,
// Statistik, Bilder, Druckbogen —, vier über die **Anwendung**. Bis
// #288 standen
// alle neun in derselben Spalte, und nirgends stand, welches Set gerade gilt.
// Wer zwei Sets offen hatte, sah den Unterschied erst an den Karten.
//
// # Eine Laschenleiste und keine Knopfreihe
//
// `role="tablist"` sagt einem Vorleseprogramm „Lasche 2 von 5" und lässt die
// Pfeiltasten wandern — die gewohnte Bedienung, und §6 verlangt sie
// durchgängig. Eine Reihe Knöpfe sähe gleich aus und wäre es nicht.
//
// # Der Name ist eine Zugabe
//
// Scheitert das Laden, bleiben die Laschen stehen. Die Karten sind der Grund,
// warum jemand hier ist; der Name macht es bequemer (§8).
import { useId, useRef, type ReactNode, type RefObject } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { lasche } from '../bausteine/laschenstil'
import { Leerzustand } from '../bausteine/Leerzustand'
import { Setleiste } from './Setleiste'
import { useGelesenesSet } from '../haken/offenesset'
import {
  beschriftung,
  nachbarlasche,
  type Seitenschluessel,
  type Setschluessel,
} from '../navigation'
import { laschengruppen, setlaschen } from '../laschengruppen'
import {
  schubladen,
  schubladeVon,
  type Schubladenschluessel,
} from '../schubladen'
import { useTexte } from '../sprache/kontext'

interface SetarbeitsplatzEigenschaften {
  setID: string | undefined
  aktuell: Seitenschluessel
  wechseln: (seite: Seitenschluessel) => void

  // verlaufsstand und oeffneKarte braucht die Leiste am Kopf (AP-9).
  verlaufsstand: number
  oeffneKarte: (setID: string, karteID: string) => void

  // children bekommt das **gesetzte** Set.
  //
  // **Als Funktion und nicht als Knoten**: ohne Set zeigt der Arbeitsplatz
  // einen leeren Zustand und ruft sie gar nicht. Damit ist `setID` in jeder
  // Lasche eine Zeichenkette, und die fünf Wächter „kein Set gewählt", die
  // dort standen, sind weg — es waren Zweige, die kein Aufruf betreten konnte
  // (§3).
  children: (setID: string) => ReactNode
}

// laschenreihe ist die Ordnung der Laschen als Schlüssel — die Form, in der
// [nachbarlasche] rechnet.
const laschenreihe = setlaschen.map((lasche) => lasche.schluessel)

export function Setarbeitsplatz({
  setID,
  aktuell,
  wechseln,
  verlaufsstand,
  oeffneKarte,
  children,
}: SetarbeitsplatzEigenschaften) {
  const texte = useTexte()

  // **Die Leiste hält ihre Knöpfe**, damit die Pfeiltasten den Fokus setzen
  // können. Über `document.getElementById` ginge es auch — und träfe die
  // Lasche eines anderen Fensters, sobald es zwei gibt.
  const leiste = useRef<HTMLDivElement>(null)

  // **Die Kennungen erzeugt React** (#1553) — aus demselben Grund wie die
  // Leiste darüber: ein zweiter Arbeitsplatz gäbe sonst zweimal dieselbe.
  const laschen = useId()
  const feldkennung = useId()

  // **Gelesen hat das Fenster** (U13, #2158) — einmal für Brotkrumen und
  // Arbeitsplatz, statt je einmal.
  const satz = useGelesenesSet(setID)

  // **Die Fläche unter einer Schublade ist die Kartenfläche** (UX-08,
  // #1680). Steht eine Seite in einer Schublade, markiert die Leiste die
  // Karten — sie sind es, die darunter stehen, und auf sie kehrt Escape
  // zurück.
  const flaeche: Seitenschluessel =
    schubladeVon(aktuell) === undefined ? aktuell : 'karten'

  // **Der Ausgabeassistent ist keine Lasche** (UXR-09, #1917): er steht im
  // Arbeitsplatz, aber die Leiste wählt nichts, und das Feld darunter nennt
  // sich nicht nach einer Lasche.
  const istLasche = laschenreihe.includes(flaeche)

  // **Der Knopf nimmt den Fokus, bevor seine Schublade aufgeht** (S06.NAT-02).
  // WebKit gibt einem Knopf beim Klick keinen — die Gewohnheit von macOS —,
  // und die Schublade merkte sich sonst `body` als den Ort, an den sie ihn
  // beim Schließen zurückgibt. Gefunden in der Sichtprüfung im Browser.
  const griffe: Record<
    Schubladenschluessel,
    RefObject<HTMLButtonElement | null>
  > = {
    eigenschaften: useRef<HTMLButtonElement>(null),
    werkzeuge: useRef<HTMLButtonElement>(null),
  }

  // **Der Schritt ist ein Knopf** (UX-15, #1771) und kein Satz, der auf eine
  // Datei verweist: wer schon Sets hat, soll nicht lesen, er müsse erst eine
  // Datei holen.
  if (setID === undefined) {
    return (
      <Leerzustand
        titel={texte.allgemein.arbeitsplatz.keinSet}
        erklaerung={texte.allgemein.arbeitsplatz.keinSetErklaerung}
      >
        <Knopf
          gewicht="haupt"
          aufKlick={() => {
            wechseln('sets')
          }}
        >
          {texte.allgemein.arbeitsplatz.zuDenSets}
        </Knopf>
      </Leerzustand>
    )
  }

  // aufTaste bewegt den Fokus zwischen den Laschen — das Muster, das ein
  // Vorleseprogramm bei `role="tablist"` erwartet.
  //
  // **Gezählt wird von der Lasche, auf der man steht** — nicht von der
  // gewählten. Solange die gewählte die einzige im Tabweg war, war das
  // dasselbe; auf dem Ausgabeassistenten ist keine gewählt (UXR-09, #1917),
  // und die Pfeile führten sonst nirgendwohin.
  const aufTaste = (von: Setschluessel, ereignis: React.KeyboardEvent) => {
    const richtung =
      ereignis.key === 'ArrowRight' ? 1 : ereignis.key === 'ArrowLeft' ? -1 : 0

    if (richtung === 0) {
      return
    }

    ereignis.preventDefault()

    const naechste = nachbarlasche(laschenreihe, von, richtung)

    // **Eine Lasche hat immer einen Nachbarn** — `von` steht in der Reihe,
    // und `nachbarlasche` gibt `undefined` nur für eine, die fehlt. Der Typ
    // kennt die Zusage nicht.
    /* v8 ignore next 3 -- unerreichbar: `von` ist eine Lasche der Reihe */
    if (naechste === undefined) {
      return
    }

    wechseln(naechste)

    // **Der Fokus wandert mit.** Ohne ihn bliebe er auf der alten Lasche, und
    // die nächste Pfeiltaste zählte von dort — man käme nie weiter als eins.
    leiste.current
      ?.querySelector<HTMLElement>(`[data-lasche="${naechste}"]`)
      ?.focus()
  }

  return (
    <div className="flex min-h-full flex-col">
      <header className="mb-5 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            {satz?.name ?? texte.allgemein.arbeitsplatz.ohneNamen}
          </h1>

          {satz !== undefined && (
            <p className="mt-1 text-leise">
              {/* **Die Zahl der Karten steht in der Leiste darunter** (AP-9)
                  — hier stand sie doppelt, und bei einer Karte als „1
                  Karten“. */}
              {`${satz.spiel} · ${satz.stil}`}
            </p>
          )}
        </div>

        {/* **Die Schubladen öffnen hier** — wie im Entwurf rechts im Kopf,
            neben dem Namen, über den sie sprechen (UX-08, #1680). Jede öffnet
            bei ihrer ersten Seite; wer eine andere will, wählt sie in der
            Schublade oder über ihr Kürzel. */}
        <div className="flex gap-2">
          {schubladen.map((schublade) => (
            <Knopf
              key={schublade.schluessel}
              gewicht="neben"
              griff={griffe[schublade.schluessel]}
              aufKlick={() => {
                griffe[schublade.schluessel].current?.focus()
                wechseln(schublade.laschen[0])
              }}
            >
              {
                texte.allgemein.arbeitsplatz.schubladen[schublade.schluessel]
                  .knopf
              }
            </Knopf>
          ))}

          {/* **Ausgeben steht neben den Schubladen** — wie im Artboard
              *Karten* (UXR-09, #1917). Er schlägt den Ausgabeassistenten
              auf, wie er zuletzt dastand; *Bilder* und *Druckbogen* sind
              seine Ziele und keine Laschen mehr.

              **Im Nebengewicht** seit AP-3 (#2360): der Kopf steht über
              jeder Seite des Sets, und jede hat ihre eigene führende
              Handlung — „+ Karte anlegen“, „Nachschlagen“. Je Blick führt
              einer (`Zustaende09`); das Ausgeben selbst führt auf seiner
              Seite (`Wege09`). */}
          <Knopf
            aufKlick={() => {
              wechseln('ausgeben')
            }}
          >
            {beschriftung(texte, 'ausgeben')}
          </Knopf>
        </div>
      </header>

      {/* **Die Leiste am Setkopf** — AP-9 (#2366), `Bereiche09`: über den
          Laschen, für jede Seite des Sets dieselbe. */}
      {satz !== undefined && (
        <div className="mb-4">
          <Setleiste
            setID={setID}
            karten={satz.karten}
            verlaufsstand={verlaufsstand}
            wechseln={wechseln}
            oeffneKarte={oeffneKarte}
          />
        </div>
      )}

      {/* **Eine Leiste, zwei Gruppen** — fünf waren es seit #1166; drei
          davon sind seit UX-08 die beiden Schubladen (#1680).
          
          # Warum die Gruppen `role="presentation"` tragen
          
          ARIA lässt in einer `tablist` nur `tab` als Kind zu; ein `group`
          dazwischen machte die Leiste ungültig, und ein Vorleseprogramm
          verlöre die Zählung „Lasche 4 von 12". Die Umschläge sind deshalb
          ausdrücklich durchsichtig.
          
          # Die Gruppe hört man an ihrer ersten Lasche
          
          **Die Aufschrift selbst ist für die Vorlesehilfe verborgen**, und
          die erste Lasche jeder Gruppe nennt sie als Beschreibung — #1991.
          Bis zum 2026-10-01 stand sie als lesbarer Text da, und VoiceOver
          las in der Abnahme-VM **jede** Lasche der ersten Gruppe mit der
          Aufschrift der zweiten: „Cards OUTPUT AND PRINT". Gemessen mit
          einer WKWebView-Probe: nur mit dem gebauten CSS, ohne `text-xs`
          nicht — VoiceOver hängt nahen Text nach der Lage an. Mit dieser
          Bauform liest es „Cards BUILD CARDS, selected, tab, 1 of 6" und
          „Export OUTPUT AND PRINT, tab, 5 of 6", die übrigen ohne Zusatz.
          
          **Und die Alternative ist geprüft und verworfen**: ein `aria-label`
          je Lasche — »Konsole, Technik« — trüge die Gruppe in jede Ansage.
          Das verlängert **jede** der zwölf um die Angabe, die in den meisten
          Fällen niemand braucht, und es überschreibt den sichtbaren Namen,
          nach dem jede Bedienhilfe sucht.
          
          **Und die Leiste bricht um.** `flex-wrap` statt einer Zeile, die bei
          1.440 Punkten die letzte Lasche abschnitt — ein Knopf, den niemand
          sieht, ist keiner.

          # Die Aufschrift steht neben den Laschen

          **Nicht darüber** — UXR-02 (#1910): über ihnen nahm sie auf jeder
          Seite des Sets eine eigene Zeile, und das Artboard *Karten* zeichnet
          über dem Arbeitsplatz gar keine Laschenzeile. Verborgen und an der
          ersten Lasche beschrieben bleibt sie, wie oben steht. */}
      <div
        ref={leiste}
        role="tablist"
        aria-label={texte.allgemein.arbeitsplatz.laschen}
        className="flex flex-wrap items-center gap-x-5 gap-y-2"
      >
        {laschengruppen.map((gruppe) => (
          <div
            key={gruppe.schluessel}
            role="presentation"
            className="flex items-center gap-2"
          >
            <span
              id={`${laschen}-gruppe-${gruppe.schluessel}`}
              aria-hidden="true"
              className="px-1 text-xs font-medium uppercase tracking-wide text-leise"
            >
              {texte.allgemein.arbeitsplatz.gruppen[gruppe.schluessel]}
            </span>

            <div role="presentation" className="flex gap-1">
              {gruppe.laschen.map((schluessel: Setschluessel, stelle) => {
                const gewaehlt = schluessel === flaeche

                return (
                  <button
                    key={schluessel}
                    type="button"
                    role="tab"
                    id={`${laschen}-${schluessel}`}
                    data-lasche={schluessel}
                    aria-selected={gewaehlt}
                    aria-describedby={
                      stelle === 0
                        ? `${laschen}-gruppe-${gruppe.schluessel}`
                        : undefined
                    }
                    aria-controls={feldkennung}
                    // **Ohne gewählte Lasche steht die erste im Tabweg** —
                    // sonst wäre die Leiste auf dem Ausgabeassistenten mit der
                    // Tastatur nicht zu erreichen.
                    tabIndex={
                      gewaehlt || (!istLasche && schluessel === laschenreihe[0])
                        ? 0
                        : -1
                    }
                    onClick={() => {
                      wechseln(schluessel)
                    }}
                    // **Die Taste hängt an der Lasche, nicht an der Leiste.**
                    // Eine Leiste mit Tastenbehandlung müsste selbst den Fokus
                    // nehmen können (`jsx-a11y/interactive-supports-focus`) —
                    // und dann stünde eine Station mehr im Tabweg, die nichts
                    // tut. Das ARIA-Muster für Laschen sieht die Pfeile ohnehin
                    // an der gewählten Lasche vor.
                    onKeyDown={(ereignis) => {
                      aufTaste(schluessel, ereignis)
                    }}
                    // **Eine Füllung und kein Reiterblatt.**
                    //
                    // Die Lasche sah aus wie ein Registerblatt, das oben am
                    // Feld darunter klebt — und das trägt nur, solange alle
                    // Laschen in **einer** Zeile stehen. Sobald die Leiste
                    // umbricht, klebt die gewählte Lasche irgendwo in der
                    // ersten Zeile und das Feld beginnt unter der zweiten;
                    // die Verbindung, die das Blatt behauptet, gibt es nicht
                    // mehr. Im Fenster nachgesehen am 2026-08-27 (#1166).
                    className={lasche}
                  >
                    {beschriftung(texte, schluessel)}
                  </button>
                )
              })}
            </div>
          </div>
        ))}
      </div>

      {/* **Kein Kasten um das Feld** — UX-15 (#1771). Gemessen am
          2026-09-24 in WebKit bei 1.440 × 900: Rand und Innenabstand des
          Kastens nahmen dem Kartenarbeitsplatz 50 Punkte, er maß 1.045 statt
          1.095, und der Navigator klappte unter `engeGrenze` (1.072) von
          selbst ein — die Kartenliste fehlte auf dem Fenster, für das das
          Artboard *Karten* gezeichnet ist. Dort steht der Arbeitsplatz ohne
          Rahmen unter dem Kopf; die Leiste trennt eine Linie. */}
      {/* **Ein Laschenfeld nur unter einer gewählten Lasche** — der
          Ausgabeassistent trägt seine eigene Überschrift (UXR-09, #1917).
          Es bleibt **dasselbe** Element: ein anderes baute den Inhalt
          darunter neu, und mit ihm jede aufgehobene Seite. */}
      <div
        id={feldkennung}
        role={istLasche ? 'tabpanel' : undefined}
        aria-labelledby={istLasche ? `${laschen}-${flaeche}` : undefined}
        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- ein Laschenfeld nimmt den Fokus; ohne gewählte Lasche ist es keines und nimmt keinen
        tabIndex={istLasche ? 0 : undefined}
        className="mt-3 flex-1 border-t border-kante pt-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      >
        {children(setID)}
      </div>
    </div>
  )
}
