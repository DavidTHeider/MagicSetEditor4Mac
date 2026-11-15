// Das **Quelltextfeld** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Der Griff
//
// Ein `<textarea>` kann keine Farben. Also liegen zwei Dinge übereinander:
// unten ein `<pre>`, das denselben Text in Abschnitten zeigt, oben ein
// **durchsichtiges** Feld, das die Eingaben nimmt und den Schreibzeiger führt.
//
// **Beide tragen dieselbe Schrift, dieselbe Zeilenhöhe und dieselben
// Abstände.** Läuft das auseinander, steht die Farbe neben dem Zeichen —
// deshalb kommen die Maße aus **einer** Konstante und nicht aus zwei
// Klassenlisten.
//
// # Woher die Farben kommen
//
// Aus dem Lexer in Go, über die Bindung `Skriptfarben`. Ein zweiter Lexer in
// TypeScript liefe an dem Tag auseinander, an dem einer von beiden eine Regel
// lernt — und die Einfärbung zeigte eine Sprache, die es nicht gibt (§10).
//
// # Zwei Sprachen, und jede hat ihre eigene Bindung
//
// **Eine Paketdatei ist kein MSE-Script.** `game: magic` ist ein Schlüssel mit
// einem Wert, kein Ausdruck; der Skript-Lexer liest daraus Bezeichner und
// Doppelpunkte und färbt nichts Sinnvolles. Deshalb nimmt dieses Feld eine
// **Sprache**.
//
// Bis zum 2026-08-28 sagte der Vorlagenbereich `keine` — Zeilennummern ja,
// Farbe nein —, und das war richtig herum geirrt: keine Farbe ist besser als
// eine falsche. Seit
// #1181 gibt
// es die richtige, und ein `script:`-Wert darin trägt die Arten der Sprache.
//
// **Zwei Bindungen und kein Schalter darin**: welche gefragt wird, entscheidet
// der Aufrufer. Ein Parameter „welche Sprache" wäre eine fachliche
// Verzweigung in `internal/app`, und dort gehört keine hin (§2).
//
// # Was er nicht kann
//
// **Zeilenumbruch.** Das Feld läuft waagerecht über, statt umzubrechen: eine
// umgebrochene Zeile hätte zwei Zeilenhöhen und eine Nummer, und die
// Nummernspalte daneben zählte falsch.
//
// **Das ist eine eigene Entscheidung und keine übernommene.** Ein
// `wxTE_DONTWRAP` gibt es in MSE2 nicht; es baut seine Eingabe als
// **einzeiliges** `wxTextCtrl`, allein mit `wxTE_PROCESS_ENTER`
// (`console_panel.cpp`, Zeile 399, Fassung f9d9356 vom 2021-07-06). Dort
// bricht nichts um, weil dort nur eine Zeile steht — das sagt über ein
// mehrzeiliges Feld nichts.
import { holeFarben } from './quelltextfarben'
import { useLaden } from '../haken/useLaden'
import { Textsuche } from './Textsuche'
import { Vorschlagsliste } from './Vorschlagsliste'
import { useSchreibzeiger } from './useSchreibzeiger'
import { useSprung } from './quelltextsprung'
import { abschnitte, type Bereich } from './quelltextanzeige'
import {
  anzeigeklassen,
  eingabeklassen,
  nummernklassen,
  zeilenhoehe,
} from './quelltextklassen'
import { useCallback, useEffect, useRef, useState } from 'react'

import type { Skriptstelle } from '../api'
import type { Sprungziel } from './quelltextsprung'

// Sprache sagt, wonach eingefärbt wird.
//
// **`keine` ist kein Mangel, sondern eine Aussage**: was hier steht, ist keine
// Sprache, die dieses Programm kennt — eine Konsole etwa nimmt beliebigen Text.
export type Sprache = 'script' | 'paket' | 'keine'

interface QuelltextfeldEigenschaften {
  // bezeichnung ist die Aufschrift des Felds.
  bezeichnung: string

  // sprache sagt, wonach eingefärbt wird — Vorgabe ist MSE-Script.
  sprache?: Sprache

  // namen sind die Namen der Umgebung — für die Vervollständigung.
  //
  // **Leer heißt: es gibt keine**, und dann schlägt das Feld nichts vor. Es
  // holt sie nicht selbst: welche gelten, hängt an der Karte, und das weiß nur
  // die Seite (#1164).
  namen?: readonly string[] | undefined

  wert: string
  setzeWert: (wert: string) => void

  // stelle ist ein gemeldeter Fehler — er wird angestrichen.
  stelle?: Skriptstelle | null

  // sprungziel setzt den Zeiger an eine Stelle — siehe `quelltextsprung.ts`.
  sprungziel?: Sprungziel | null

  // aufZeiger meldet, wo der Schreibzeiger steht — der Inspector der
  // Werkstatt sagt dazu, woher der Name kommt (U7, #1418).
  aufZeiger?: ((stelle: number) => void) | undefined

  // zeilen ist die Höhe des Felds in Zeilen.
  zeilen?: number

  platzhalter?: string

  // aufAusfuehren läuft bei ⌘↩ oder Strg+↩ im Feld — Werkstück U13 von
  // Phase 5.1 (#1424). Wer nichts übergibt, hat ein Feld ohne Lauf.
  aufAusfuehren?: (() => void) | undefined
}

export function Quelltextfeld({
  bezeichnung,
  wert,
  setzeWert,
  stelle = null,
  zeilen = 8,
  platzhalter,
  sprache = 'script',
  namen = [],
  sprungziel = null,
  aufZeiger,
  aufAusfuehren,
}: QuelltextfeldEigenschaften) {
  // **Der Rückruf steht in einer Referenz**, weil der Tastenhorcher unten in
  // einem Ref-Rückruf ohne Abhängigkeiten hängt — er sieht sonst den Rückruf
  // vom ersten Zeichnen.
  const ausfuehren = useRef(aufAusfuehren)

  useEffect(() => {
    ausfuehren.current = aufAusfuehren
  }, [aufAusfuehren])
  // **Die Farben kommen nach, der Text steht sofort.** Zwischen Anschlag und
  // Antwort zeigt die Anzeige den neuen Text mit den **alten** Abschnitten —
  // einen Augenblick lang falsch gefärbt und nie falsch geschrieben. Genau
  // dafür ist `behalten` da; die Kartenvorschau macht es aus demselben Grund
  // so.
  //
  // **Ein Fehler lässt den Text stehen.** Die Farbe ist eine Zutat, kein
  // Inhalt; wer ihn verlöre, weil eine Bindung nicht antwortet, hätte einen
  // Editor, der beim ersten Aussetzer leer ist.
  const farbstand = useLaden(
    async () => holeFarben(sprache, wert),
    [wert, sprache],
    true,
  )

  const merkmale = farbstand.daten ?? []

  // **Die Suche gehört ins Feld** (#1164). ⌘F führt in dieser Anwendung sonst
  // zur Volltextsuche über die Karten (§6); in einem Textfeld heißt dieselbe
  // Taste seit je „suche **hier**". Das ist keine gebrochene Gewohnheit,
  // sondern die stärkere — wer in einer 400-Zeilen-Vorlage steht, meint nicht
  // die Karten.
  const [suchtGerade, setzeSuchtGerade] = useState(false)
  const eingabe = useRef<HTMLTextAreaElement>(null)

  // gezeichnetesFeld ist dasselbe Feld als Zustand — damit die
  // Vorschlagsliste es **ohne Prüfung** bekommt.
  const [gezeichnetesFeld, setzeGezeichnetesFeld] =
    useState<HTMLTextAreaElement | null>(null)

  // **⌘F wird am Feld selbst abgefangen — nativ, nicht über React.**
  //
  // # Warum nativ
  //
  // Der Rahmen horcht auf ⌘F am **Fenster** und schickt zur Volltextsuche über
  // die Karten (§6). Aufhalten lässt er sich nur, wenn das Ereignis ihn nicht
  // erreicht — `preventDefault` genügt dafür nicht, es hält den **Browser**
  // auf und nicht den zweiten Horcher.
  //
  // Und `stopPropagation` an einem React-Ereignis genügt auch nicht: React
  // hängt seine Horcher an die Wurzel, und dort ist das Ereignis schon
  // unterwegs. Gemessen am 2026-08-26 — der Horcher am Fenster lief trotzdem.
  //
  // # Warum ein Ref-Rückruf und kein Effekt
  //
  // Weil der Effekt ein `feld === null` bräuchte, das kein Lauf betritt: nach
  // dem Zeichnen steht die Referenz. Ein Ref-Rückruf, der eine Aufräumfunktion
  // **zurückgibt**, wird seit React 19 nie mit `null` gerufen — der Zweig
  // entfällt, statt in `.covignore` zu landen (§3).
  const anTasten = useCallback((feld: HTMLTextAreaElement) => {
    eingabe.current = feld
    setzeGezeichnetesFeld(feld)

    const aufTaste = (ereignis: KeyboardEvent) => {
      // **⌘ oder Strg, aber nicht beide**: ⌃⌘F ist das Vollbild von macOS
      // (#1994), und die Suche fing es ab.
      if (ereignis.key === 'f' && ereignis.metaKey !== ereignis.ctrlKey) {
        ereignis.preventDefault()
        ereignis.stopPropagation()
        setzeSuchtGerade(true)
      }

      // **⌘↩ wertet aus** (#1424) — ein Enter ohne Taste dazu bleibt der
      // Zeilenumbruch.
      if (ereignis.key === 'Enter' && (ereignis.metaKey || ereignis.ctrlKey)) {
        ereignis.preventDefault()
        ausfuehren.current?.()
      }
    }

    feld.addEventListener('keydown', aufTaste)

    return () => {
      feld.removeEventListener('keydown', aufTaste)
    }
  }, [])

  const [treffer, setzeTreffer] = useState<Bereich | null>(null)

  // **Wo der Zeiger steht, und was daran hängt** — Wort, Vorschläge und das
  // Klammerpaar. Alle drei beantworten dieselbe Frage; die Rechnung dazu
  // steht in `useSchreibzeiger.ts`.
  const zeiger = useSchreibzeiger(wert, merkmale, namen)

  useSprung(() => eingabe.current, sprungziel, zeiger.setzeStelle, aufZeiger)
  const { wort, vorschlaege, paar } = zeiger

  // uebernehmen setzt einen Vorschlag an die Stelle des angefangenen Wortes.
  //
  // **Das Feld kommt als Argument und wird nicht aus der Referenz geholt.**
  // Dort wäre es `HTMLTextAreaElement | null`, und der `null`-Zweig ist
  // keiner: die Vorschlagsliste steht nur da, wenn das Feld gezeichnet ist.
  // Ein Zweig ohne Fall gehört weg (§3).
  const uebernehmen = (feld: HTMLTextAreaElement, name: string) => {
    const neu = wert.slice(0, wort.von) + name + wert.slice(wort.bis)

    setzeWert(neu)

    // **Der Zeiger steht hinter dem eingesetzten Namen.** Wer sonst
    // weitertippt, schriebe mitten hinein.
    const dahinter = wort.von + name.length

    zeiger.setzeStelle(dahinter)

    feld.focus()
    feld.setSelectionRange(dahinter, dahinter)
  }

  // **Der Treffer wird angezeigt, nicht ausgewählt.**
  //
  // Der erste Anlauf setzte die Auswahl in der Eingabe und holte dafür den
  // Fokus zurück — und damit ging **jeder zweite Tastendruck ins Textfeld
  // statt ins Suchfeld**. Gefunden von der Zusicherung über die Auswahl: sie
  // fand 14 statt 5, weil nach dem ersten Buchstaben niemand mehr im Suchfeld
  // stand.
  //
  // Die Anzeige darunter kann es ohnehin besser: sie hebt den Fund hervor,
  // während der Fokus bleibt, wo jemand tippt.
  //
  // **Ein fester Rückruf.** Ohne ihn entstünde bei jedem Zeichnen ein neuer,
  // und der Effekt in der Suche liefe erneut.
  const zeigeTreffer = useCallback((von: number, bis: number) => {
    setzeTreffer({ von, bis })
  }, [])

  // **Beim Schließen geht die Auswahl in die Eingabe** — dann ist der Fokus
  // dort richtig aufgehoben, und wer weitertippt, steht am Fund.
  const sucheSchliessen = useCallback(() => {
    setzeSuchtGerade(false)

    const feld = eingabe.current

    if (feld !== null && treffer !== null) {
      feld.focus()
      feld.setSelectionRange(treffer.von, treffer.bis)
    }

    setzeTreffer(null)
  }, [treffer])

  const zeilenText = wert.split('\n')

  return (
    // **Ein `div` und kein `label`.** Die Beschriftung hängt am `aria-label`
    // des Feldes; das Element hier ordnet nur an. Als `label` trüge es Knöpfe
    // in sich, und ein `label` darf keine tragen — ein Klick darauf gälte
    // zusätzlich seinem Feld.
    <div
      className="flex flex-col gap-1"
      // **Geräumt wird, wenn der Fokus die ganze Gruppe verlässt** — nicht
      // schon, wenn er das Textfeld verlässt.
      //
      // # Warum das der Unterschied ist
      //
      // Ein Vorschlag ist ein Knopf. Wer ihn mit der Tastatur ansteuert, nimmt
      // dem Feld den Fokus — und ein `onBlur` am Feld räumte die Liste weg,
      // **bevor** der Knopf gedrückt ist. Mit der Maus fiel das nicht auf, weil
      // der Knopf dort schon beim Drücken greift; mit der Tastatur war die
      // Vervollständigung gar nicht zu bedienen (§6).
      //
      // Gefunden von genau dem Lauf, der sie über die Tastatur annimmt.
      onBlur={(ereignis) => {
        // **Die Vorschläge gehen mit dem Fokus, der gemeldete Zeiger nicht.**
        // Wer im Inspector „Zur Definition" klickt, nimmt dem Feld zuerst
        // den Fokus — meldete das Feld hier −1, verschwände der Knopf unter
        // dem Mauszeiger, bevor der Klick ankommt. Gefunden im Seitenlauf
        // der Werkstatt (U7): `oeffnen` wurde nie gerufen.
        if (!ereignis.currentTarget.contains(ereignis.relatedTarget)) {
          zeiger.setzeStelle(-1)
        }
      }}
    >
      <span className="order-1 text-sm font-medium">{bezeichnung}</span>

      {suchtGerade && (
        <Textsuche
          wert={wert}
          setzeWert={setzeWert}
          zeigeTreffer={zeigeTreffer}
          schliessen={sucheSchliessen}
        />
      )}

      <div className="order-4 flex overflow-hidden rounded-sm border border-kante focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-ring">
        {/* **Die Nummern stehen außerhalb des Felds.** Innen wären sie Text,
            den man auswählen und überschreiben kann. */}
        <ol aria-hidden="true" className={nummernklassen}>
          {zeilenText.map((_, nummer) => (
            // Der Index ist hier der Schlüssel, und das ist richtig: eine
            // Zeilennummer **ist** ihre Stelle in der Liste.
            <li key={nummer}>{nummer + 1}</li>
          ))}
        </ol>

        <div className="relative flex-1 overflow-auto">
          {/* **Die Höhe steht am `pre` und nicht am Feld.** Die Eingabe liegt
              absolut darüber; ihr `rows` sagt nichts über die Größe des
              Kastens. Ohne diese Zeile war die Konsole mit `zeilen={3}` genau
              eine Zeile hoch — gefunden bei der Sichtprüfung im laufenden
              Fenster. */}
          <pre
            data-testid="quelltextanzeige"
            aria-hidden="true"
            className={anzeigeklassen}
            style={{ minHeight: `${String(zeilen * zeilenhoehe)}rem` }}
          >
            {abschnitte(wert, merkmale, stelle, treffer, paar)}
          </pre>

          {/* **Durchsichtig, aber nicht unsichtbar**: der Schreibzeiger
              braucht eine Farbe, sonst sieht niemand, wo er steht. */}
          <textarea
            ref={anTasten}
            aria-label={bezeichnung}
            value={wert}
            rows={zeilen}
            placeholder={platzhalter}
            spellCheck={false}
            wrap="off"
            onChange={(ereignis) => {
              setzeWert(ereignis.target.value)
              zeiger.setzeStelle(ereignis.target.selectionStart)
              aufZeiger?.(ereignis.target.selectionStart)
            }}
            onSelect={(ereignis) => {
              zeiger.setzeStelle(ereignis.currentTarget.selectionStart)
              aufZeiger?.(ereignis.currentTarget.selectionStart)
            }}

            className={eingabeklassen}
          />
        </div>
      </div>

      {/* **Die Liste steht im Baum hinter dem Feld und im Bild darüber.**
          Beides hat einen Grund, und sie widersprechen sich nicht:

          - **Im Bild oben**, weil unter dem Feld die Knöpfe liegen; eine
            Leiste, die sie beim Tippen verschöbe, ließe jemanden danebengreifen.
          - **Im Baum dahinter**, weil die Tabulatortaste dem Baum folgt und
            nicht dem Bild. Davor gelegen, war die Leiste mit `Tab` gar nicht
            zu erreichen — der Weg führte am Feld vorbei zum nächsten Knopf.
            Gemessen im laufenden Fenster: `Tab` landete auf »Auswerten«.

          Dieselbe Reihenfolge liest auch eine Sprachausgabe vor: erst das
          Feld, dann seine Vorschläge.

          **Erst wenn das Feld steht.** Vor dem ersten Zeichnen gibt es keines,
          in das ein Vorschlag ginge — und damit auch keinen `null`-Zweig
          weiter oben. */}
      {gezeichnetesFeld !== null && vorschlaege.length > 0 && (
        <Vorschlagsliste
          vorschlaege={vorschlaege}
          uebernehmen={(name) => {
            uebernehmen(gezeichnetesFeld, name)
          }}
        />
      )}
    </div>
  )
}
