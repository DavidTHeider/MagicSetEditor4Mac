// **Eine Einstellung mit dem Bedienelement ihrer Werte** — herausgelöst aus
// `Einstellungen.tsx` für Phase 4.8
// (#1321).
//
// **Warum eine eigene Datei:** die Dienstschalter brauchen dasselbe
// Bedienelement mit einer **anderen Beschriftung** — ihr Name ist ein
// Eigenname und steht in keinem Sprachkatalog (0063, Entscheidung 5). Eine
// zweite Auswahlliste daneben wäre eine zweite Antwort auf dieselbe Frage.
//
// # Die Zahl der Werte wählt das Bedienelement — UXR-F2
//
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`. Welches, steht in `wahlart.ts`: ein
// Schalter für zwei Werte, die an und aus sind; Segmente für zwei bis vier;
// eine Auswahlliste darüber; ohne Werteliste ein Textfeld.
//
// **Der Erbwert ist ein Wert unter anderen.** `erben` — „wie beim Nutzer" —
// steht in den Segmenten wie vorher in der Auswahlliste: als eigener Eintrag,
// erreichbar von jedem anderen. Ein Schalter hätte dafür keinen Platz; eine
// Einstellung, die erben kann, hat darum mindestens drei Werte und wird nie
// einer (`wahlkatalog.test.tsx` zählt es über den Katalog nach).
//
// **Ein Schalter braucht keinen Knopf „Zurücksetzen"**: umlegen ist dasselbe,
// wie bei den Bereichsschaltern (UXR-19, #1927). Bei Segmenten bleibt er — er
// nennt, was die Vorgabe ist, und das steht an keinem Segment.
import { useId } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { Einstellungsherkunft } from '../bausteine/Einstellungsherkunft'
import { Feldhilfe } from '../bausteine/Feldhilfe'
import { Schalter } from '../bausteine/Schalter'
import { Segmente } from '../bausteine/Segmente'
import { kurzform, langform, ohneAuszeichnung } from '../sprache/kurzform'
import { Satz } from '../bausteine/Satz'
import { Zeitzonenwahl } from '../bausteine/Zeitzonenwahl'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { bedienung as bedienungVon } from './wahlart'
import { Sprachstaende } from './Sprachstaende'
import { eingabeklassen, erbenWert, Freitext } from './Wahlfreitext'

// Wahlbarer ist, was [Wahl] braucht — gleich, ob daraus ein Schalter, Segmente,
// eine Auswahlliste oder ein Textfeld wird.
//
// **Eine Gestalt und kein bestimmter Verbund** — Werkstück 8 von Phase 4.12
// (#1351):
// `EinstellungDTO` verlangt ein `bereich`, und die Einstellungen, die dem
// **Menschen** gehören, tragen keines. Ein erfundenes daran wäre eine Angabe,
// die nichts bedeutet; verlangt werden drei Felder, der Rest ist wahlweise.
export interface Wahlbarer {
  schluessel: string
  wert: string
  werte: readonly string[]

  // **Wirksames, Ebene, Herkunft und Vorgabe** — Werkstück U11 von Phase 5.1
  // (#1422). Wahlweise: die Stiloptionen und Exportoptionen, die dieselbe
  // Auswahlliste benutzen, haben keine Herkunft.
  wirksam?: string
  ebene?: string
  herkunft?: string
  wirkung?: string
  vorgabe?: string
}

// zeitzoneSchluessel ist die Einstellung, die eine Auswahl mit Suche bekommt —
// `domain.EinstellungZeitzone`.
const zeitzoneSchluessel = 'zeitzone'

interface Wahleigenschaften {
  eintrag: Wahlbarer
  // gesucht markiert das Feld als Treffer der Suche — UX-04.AK-03
  // (#1676).
  //
  // **Eine Marke und kein zweites Feld**: der Entwurf zeigt den Kasten an
  // seinem Platz mit dem Wort *gesucht* daneben. Eine Kopie in einer
  // Trefferliste wäre zweimal dasselbe Bedienelement — ein Befund (#1768).
  gesucht?: boolean | undefined

  angehalten: boolean
  aendern: (wert: string) => void

  // beschriftung überschreibt den Namen aus dem Katalog.
  //
  // **Für Namen, die keine Übersetzung haben** — die Dienste heißen
  // „Scryfall" und „Scrydex", und das sind Eigennamen (§9).
  //
  // **`undefined` ausdrücklich erlaubt** —
  // #1540:
  // ein Aufrufer, der für **manche** Zeilen einen Eigennamen hat, soll die
  // übrigen durchreichen können, statt selbst auf den rohen Schlüssel
  // zurückzufallen. Unter `exactOptionalPropertyTypes` sind „weggelassen" und
  // „undefined übergeben" zwei Dinge, und dieser Aufruf ist der zweite.
  beschriftung?: string | undefined

  // werteNamen überschreibt die Namen der Werte.
  //
  // **Ohne sie bräuchte jeder Dienst zwei Katalogeinträge**, die alle „Aus"
  // und „An" sagen — acht Zeilen für zwei Wörter.
  werteNamen?: Record<string, string>
}

// Wahl ist eine Einstellung mit dem Bedienelement ihrer Werte.
//
// **Die Werte kommen aus der Antwort, nicht aus einer Liste hier.** Sonst
// stünden die erlaubten Werte an zwei Stellen, und die hier wäre die
// ungeprüfte: ein Wert, den `domain.PruefeEinstellung` abweist, käme in die
// Liste und beim Klick nicht an. Sichtbar wäre nur, dass nichts geschieht.
export function Wahl({
  eintrag,
  gesucht,
  angehalten,
  aendern,
  beschriftung: eigene,
  werteNamen,
}: Wahleigenschaften) {
  const texte = useTexte()
  const wahlID = useId()
  const nameID = useId()
  const bedienung = bedienungVon(eintrag.werte)

  // **Der Typ sagt die Wahrheit, auch hier.** Ein Zugriff über `as keyof`
  // behauptet, jeder Schlüssel habe einen Eintrag, und macht aus einem
  // fehlenden Text ein stilles `undefined` — React zeichnet dafür eine Option
  // **ohne Text**, ohne zu klagen.
  //
  // Genau so standen die vier Dienstschalter im Fenster: leere Auswahllisten,
  // richtiger Wert, kein Hinweis. `dienst.scryfall.aus` gibt es in keinem
  // Katalog, und `Dienstschalter` reichte `werteNamen` nicht weiter. Gefunden
  // hat es die Sichtprüfung am laufenden Fenster in Phase 4.8
  // (#1321),
  // kein Test — die Tests suchen das Feld über seine Beschriftung und prüfen
  // den **Wert**.
  const werte: Record<string, string | undefined> =
    texte.seiten.einstellungen.werte

  const namen: Record<string, string | undefined> =
    texte.seiten.einstellungen.namen

  // **Der rohe Wert ist die letzte Reihe, nicht der Normalfall.** Dass jeder
  // Wert einen Namen hat, hält `internal/werkzeuge/einstellungstexte_test.go`
  // gegen die Go-Seite. Was hier steht, ist die Zusage für den Fall, dass er
  // doch einmal fehlt: lieber `aus` lesen als einen leeren Kasten sehen.
  const beschriftung = (wert: string) =>
    werteNamen?.[wert] ?? werte[`${eintrag.schluessel}.${wert}`] ?? wert

  const name = eigene ?? namen[eintrag.schluessel] ?? eintrag.schluessel

  // Dieselbe Wahrheit für den Hinweis — die Dienstschalter haben keinen. Der
  // Linter hat es angestrichen: „the types have no overlap".
  const hinweise: Record<string, string | undefined> =
    texte.seiten.einstellungen.hinweise

  const hinweis = hinweise[eintrag.schluessel]

  // **Der lange Teil steht hinter dem Fragezeichen** — UX-15.AK-02
  // (#1771).
  const langer =
    hinweis === undefined ? undefined : langform(ohneAuszeichnung(hinweis))

  // **Ohne Werteliste ein Textfeld** — Werkstück 9 von Phase 4.12
  // (#1351),
  // Punkt 8.
  //
  // **Ein Verzeichnis lässt sich nicht aufzählen**, und eine leere
  // Auswahlliste wäre ein Bedienelement, das etwas verspricht, was nicht
  // kommt. Die Entscheidung, dass es keine Werte gibt, steht in
  // `internal/domain`; hier steht nur, was daraus folgt.
  //
  // **Das Textfeld und die Auswahlliste stehen im selben Beschriftungsfeld**;
  // der Schalter und die Segmente zeichnen ihre Beschriftung selbst, weiter
  // unten.
  const bedienelement =
    bedienung.art === 'text' ? (
      eintrag.schluessel === zeitzoneSchluessel ? (
        // **Eine Einstellung ohne Werteliste, deren Werte sich trotzdem
        // aufzählen lassen** — UXR-F1
        // (#2007):
        // Go prüft die Zeitzone gegen die Zonendatenbank, nennt sie aber
        // nicht. Die Liste steht im Fenster, und `zeitzonenliste_test.go`
        // hält, dass Go jede Zeile annimmt.
        <Zeitzonenwahl
          id={wahlID}
          bezeichnung={name}
          wert={eintrag.wert}
          angehalten={angehalten}
          aendern={aendern}
          einstellung={eintrag.schluessel}
        />
      ) : (
        // **Neu aufgebaut, wenn ein anderer Wert gilt** — so zeigt der
        // Entwurf nach einer geglückten Änderung, was im Bestand steht.
        <Freitext
          key={eintrag.wert}
          id={wahlID}
          schluessel={eintrag.schluessel}
          wert={eintrag.wert}
          // **Ein freies Feld, das erbt, zeigt den Erbwert übersetzt** —
          // #1993:
          // als Platzhalter, mit dem Wert, der gerade gilt. Vorher stand
          // `erben` roh im Feld, und VoiceOver las es vor.
          erbe={
            eintrag.vorgabe !== erbenWert
              ? undefined
              : eintrag.wirksam === undefined
                ? beschriftung(erbenWert)
                : einsetzen(texte.seiten.einstellungen.erbeMitWert, {
                    erbe: beschriftung(erbenWert),
                    wert: eintrag.wirksam,
                  })
          }
          angehalten={angehalten}
          aendern={aendern}
        />
      )
    ) : (
      <select
        id={wahlID}
        data-einstellung={eintrag.schluessel}
        value={eintrag.wert}
        disabled={angehalten}
        onChange={(ereignis) => {
          aendern(ereignis.target.value)
        }}
        className={eingabeklassen}
      >
        {eintrag.werte.map((wert) => (
          <option key={wert} value={wert}>
            {beschriftung(wert)}
          </option>
        ))}
      </select>
    )

  // **Einmal gelesen, damit die Verengung in den Rückruf reicht**: eine
  // Konstante bleibt `string`, wo ein Feld im Rückruf wieder `undefined` sein
  // könnte — und ein `?? ''` dort wäre ein Zweig, den kein Lauf erreicht.
  const vorgabe = eintrag.vorgabe

  // **Die Marke steht am Namen, wo immer der Name steht** — ein Treffer der
  // Suche (UX-04.AK-03).
  const suchmarke = gesucht === true && (
    <span className="rounded-full border border-marke px-2 py-0.5 text-xs font-normal text-marke">
      {texte.seiten.einstellungen.gesuchtMarke}
    </span>
  )

  // **Der Name und das Element, so wie die Art sie braucht.**
  //
  // - **Ein Schalter trägt den Namen als Beschriftung**: das Kästchen steht
  //   davor, wie bei den Bereichsschaltern.
  // - **Segmente sind eine Gruppe** und kein Beschriftungsfeld: ein `<label>`
  //   um mehrere Optionsfelder gehörte dem ersten, und ein Klick auf den
  //   Namen wählte es. Die Gruppe nennt den Namen über `aria-labelledby`.
  // - **Textfeld und Auswahlliste** sind ein Beschriftungsfeld wie immer.
  const bedienteil =
    bedienung.art === 'schalter' ? (
      <label className="flex items-center gap-2 text-sm font-medium">
        <Schalter
          schluessel={eintrag.schluessel}
          an={eintrag.wert === bedienung.paar.an}
          angehalten={angehalten}
          aendern={(an) => {
            aendern(an ? bedienung.paar.an : bedienung.paar.aus)
          }}
        />
        {name}
        {suchmarke}
      </label>
    ) : bedienung.art === 'segmente' ? (
      <div className="flex flex-col gap-1">
        <span
          id={nameID}
          className="flex items-center gap-2 text-sm font-medium"
        >
          {name}
          {suchmarke}
        </span>

        <Segmente
          beschriftetVon={nameID}
          schluessel={eintrag.schluessel}
          optionen={eintrag.werte.map((wert) => ({
            wert,
            name: beschriftung(wert),
          }))}
          wert={eintrag.wert}
          angehalten={angehalten}
          aendern={aendern}
        />
      </div>
    ) : (
      <label htmlFor={wahlID} className="flex max-w-xs flex-col gap-1">
        <span className="flex items-center gap-2 text-sm font-medium">
          {name}
          {suchmarke}
        </span>

        {bedienelement}
      </label>
    )

  // **Unter dem Schalter eingerückt** — Hinweis und Herkunft beginnen unter
  // dem Namen und nicht unter dem Kästchen.
  const einzug = bedienung.art === 'schalter' ? 'ml-6' : undefined

  return (
    // **Der Ring am ganzen Kasten und nicht nur am Feld** — der Entwurf
    // zeichnet `fokus-zeigen` um die Gruppe aus Beschriftung, Feld und
    // Hinweis; ein Rahmen allein um das Eingabefeld sähe aus wie Fokus.
    <div
      data-einstellungszeile={eintrag.schluessel}
      className={
        gesucht === true
          ? '-mx-3 rounded-lg border border-marke px-3 py-2'
          : undefined
      }
    >
      {bedienteil}

      {/* **Über `Satz` und nicht roh** (#1100): die Kataloge benutzen `**…**`
          und `` `…` ``, und roh gezeigt läse man wörtlich `**aus**`. */}
      {/* **Ein Hinweis, den es nicht gibt, ist kein leerer Absatz.** Die
          Dienste tragen ihren Hinweis am Bereich, nicht an jeder Zeile. */}
      {/* **Ein Satz sichtbar, der Rest hinter dem Fragezeichen** —
          UX-15.AK-02 und AK-03
          (#1771).
          Gemessen am 2026-09-22: die Hinweise sind im Mittel 171 Zeichen
          lang, ihre **ersten Sätze** 25. Der erste Satz hilft beim
          Entscheiden, der Rest begründet — und den liest man einmal. */}
      <div className={einzug}>
        {hinweis !== undefined && (
          <p className="mt-1 flex max-w-prose items-start gap-2 text-sm text-leise">
            <span>
              <Satz text={kurzform(hinweis)} />
            </span>

            {/* **Einmal gelesen und nicht zweimal gerufen**: ein zweiter Ruf
              mit `?? ''` daneben wäre ein Zweig, den kein Lauf erreicht
              (§3). */}
            {langer !== undefined && <Feldhilfe zu={name} text={langer} />}
          </p>
        )}
        <Einstellungsherkunft eintrag={eintrag} beschriftung={beschriftung} />
        {/* **Je Sprache, wie weit sie ist** (H04, #2258). */}
        {eintrag.schluessel === 'sprache' && <Sprachstaende />}
      </div>
      {vorgabe !== undefined &&
        eintrag.wert !== vorgabe &&
        bedienung.art !== 'schalter' && (
          <div className="mt-2">
            <Knopf
              gewicht="neben"
              angehalten={angehalten}
              bezeichnung={einsetzen(
                texte.seiten.einstellungen.zuruecksetzenVon,
                {
                  name,
                },
              )}
              aufKlick={() => {
                aendern(vorgabe)
              }}
            >
              {einsetzen(texte.seiten.einstellungen.zuruecksetzen, {
                vorgabe: beschriftung(vorgabe),
              })}
            </Knopf>
          </div>
        )}
    </div>
  )
}
