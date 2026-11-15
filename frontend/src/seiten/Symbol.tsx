// **Der Symboleditor** —
// #890,
// Werkstück 3 von Phase X.
//
// # Was MSE hier tut
//
// `src/gui/symbol/` (Fassung f9d9356 vom 2021-07-06): links die Teilliste,
// rechts die Zeichenfläche, oben eine Werkzeugleiste mit Grundformen,
// Punkten, Symmetrie und Auswahl. Gezeichnet wird mit der Maus.
//
// # Und was hier anders ist, mit Grund
//
// **Die Tastatur kommt zuerst.** §6 verlangt sie durchgängig, und das Issue
// nennt sie „die schwerste Zusage dieser Phase". Ein Zeichenwerkzeug, das man
// nachträglich tastaturfähig macht, wird es nie — deshalb ist **jeder Punkt
// ein eigenes fokussierbares Element** und keine Fläche, auf die man klickt.
//
// Das Ergebnis ist ein SVG, in dem Tabulator zwischen den Punkten wechselt und
// die Pfeiltasten sie verschieben. Wer eine Maus hat, benutzt dieselben
// Elemente.

import { Ladefehler } from '../bausteine/Ladehinweise'
import { useCallback, useEffect, useState } from 'react'

import type { SymbolteilDTO } from '../antworten'
import type { Handgriff } from '../haken/useZeichenverlauf'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Symbolflaeche } from './Symbolflaeche'
import { Symbolteilliste } from './Symbolteilliste'
import { Symbolwerkzeugleiste } from './Symbolwerkzeugleiste'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { Symmetriefeld } from './Symmetriefeld'
import { useSymmetrie } from '../haken/useSymmetrie'
import { useTeilliste } from '../haken/useTeilliste'
import { usePunktwege } from '../haken/usePunktwege'
import { useSymbolablage } from '../haken/useSymbolablage'
import { useSchubladenweite } from '../haken/useSchubladenweite'
import { Knopf } from '../bausteine/Knopf'
import { useZeichenverlauf } from '../haken/useZeichenverlauf'
import { useZeichenschutz } from '../haken/useZeichenschutz'
import { useSymbolzugaenge } from '../haken/useSymbolzugaenge'
import { useRuecknahmeweg } from '../ruecknahme'
import {
  type Bibliotheksbearbeitung,
  vorhandeneZeichnung,
} from './symbolablegen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { symbolworte } from './symbolworte'

interface SymbolEigenschaften {
  setID: string

  // ziel sagt, **welches** Symbol bearbeitet wird.
  //
  // **Ohne Ziel das Setsymbol** — so war es, seit es den Editor gibt. Mit Ziel
  // das Symbolfeld einer Karte
  // (#1339).
  //
  // **Der Editor kann das längst**: `Symbol(setID, name)` liest einen
  // **benannten** Symbolstand und `SymbolSetzen(setID, teile, unter)` schreibt
  // ihn unter einem Namen. Was fehlte, war der Weg von der Kartenzeile hierher.
  ziel?: { karteID: string; feld: string } | undefined

  // bibliothek sagt, dass ein Symbol der **Bibliothek** bearbeitet wird —
  // gesichert wird dann eine neue Fassung dort (DOM-02.AK-02).
  bibliothek?: Bibliotheksbearbeitung | undefined
}

// keineTeile ist die leere Zeichnung — **eine einzige, wiederverwendete**.
//
// **Ein `?? []` legte bei jedem Zeichnen ein neues Array an**, und jeder
// Effekt, der `gezeigte` in seinen Abhängigkeiten führt, liefe damit bei jedem
// Zeichnen erneut. Solange nichts geladen ist, wäre das ein Kreis, der sich
// selbst antreibt — gefunden am 2026-08-25 mit der Symmetrie
// (#1092).
const keineTeile: SymbolteilDTO[] = []

export function Symbol({ setID, ziel, bibliothek }: SymbolEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol
  const worte = symbolworte(texte.seiten.karten, bibliothek)
  const weite = useSchubladenweite()

  const [name, setzeName] = useState('')
  const [fehler, setzeFehler] = useState<string>()

  const vorgang = useVorgang()

  // vorhandenes holt die Zeichnung, die schon im Set steht.
  //
  // **Erst der Feldwert, dann die Zeichnung.** `set.symbol` nennt den
  // Dateinamen; ohne ihn gibt es nichts zu laden, und ein Aufruf mit leerem
  // Namen wäre eine Frage nach einer Datei, die niemand gemeint hat.
  //
  // **Ohne diesen Weg wäre der Editor jedes Mal leer** — wer ein Symbol
  // gezeichnet und das Set geschlossen hat, fände beim Wiederkommen nichts und
  // finge von vorn an.
  const bibliotheksID = bibliothek?.id
  const holen = useCallback(
    () => vorhandeneZeichnung(setID, ziel, bibliotheksID),
    [setID, ziel, bibliotheksID],
  )

  const vorhandenes = useLaden(holen, [setID, ziel, bibliotheksID])

  // **Eine Quelle der Wahrheit, und jede Änderung geht über sie.**
  //
  // Der erste Anlauf hielt die geladene Zeichnung in `vorhandenes.daten` und
  // änderte `teile` — das blieb leer, und wer ein vorhandenes Symbol öffnete
  // und einen Punkt bewegte, sah nichts geschehen. Angestrichen vom Zweitblick
  // über Phase X.
  //
  // **Seit dem 2026-08-31 hält der Verlauf sie** — Werkstück F von Phase 4.11
  // (#1340).
  // Der Zustand ist derselbe geblieben; was dazukommt, ist der Stapel
  // darunter. Jede Änderung geht deshalb über `eintragen` und nennt dabei
  // ihren **Handgriff**: ohne Namen wüsste eine Rücknahme nicht, was sie
  // trifft.
  // **Höchstens eine Umformung zugleich, und keine überholte Antwort**
  // (§7, Regeln 2 und 4). Jede Handlung des Editors schickt den **ganzen**
  // Zeichenstand hinaus und setzt die Antwort ganz ein; zwei davon parallel
  // verlieren einander. Angestrichen als Befund 2 des Zweitblicks über Phase
  // 4.11 — die Begründung steht am Haken.
  const schutz = useZeichenschutz()

  const verlauf = useZeichenverlauf(vorhandenes.daten?.teile ?? keineTeile)
  // **Ein Wechsel des Bezugsobjekts räumt, was unterwegs ist** (§7, Regel 1).
  //
  // Dieselben Abhängigkeiten wie beim Laden: wer ein anderes Symbol
  // aufschlägt, will nicht, dass die Grundform von vorhin noch ankommt und in
  // die neue Zeichnung fällt.
  //
  // **Im Effekt und nicht beim Zeichnen.** Das Abräumen ist eine Wirkung, kein
  // Ergebnis; beim Zeichnen ausgeführt liefe es auch dann, wenn React die
  // Komponente nur noch einmal anstreicht.
  useEffect(() => {
    schutz.raeume()
  }, [schutz, setID, ziel])
  const gezeigte = verlauf.teile

  // **⌘Z zielt auf die offene Zeichnung, solange es eine gibt.**
  //
  // MSE2 hat dafür ein eigenes Fenster mit eigenem `ActionStack`; hier ist der
  // Editor eine Seite, und beide Verläufe hängen am selben Menüeintrag. Der
  // Grund und die Abweichung stehen in `ruecknahme.tsx`.
  useRuecknahmeweg({
    zuruecknehmen: verlauf.zuruecknehmen,
    wiederherstellen: verlauf.wiederherstellen,
  })

  // eintragen ist der eine Weg in den Verlauf — und räumt die Meldung mit.
  const eintragen = useCallback(
    (neue: SymbolteilDTO[], handgriff: Handgriff) => {
      setzeFehler(undefined)
      verlauf.eintragen(neue, handgriff)
    },
    [verlauf],
  )

  // **Ein Ladefehler steht dort, wo auch die anderen stehen.**
  //
  // `useLaden` führt seinen eigenen; die Seite zeigte nur ihren. Eine
  // beschädigte `.mse-symbol` oder ein Fehler beim Lesen der Setangaben sah
  // damit aus wie „noch nichts gezeichnet" — und fremde Daten sind der
  // Normalfall (§8). Angestrichen vom Zweitblick über Phase X.
  //
  // **Der eigene zuerst**: er ist die Antwort auf die letzte Handlung, und die
  // Ladefrage ist längst beantwortet, wenn jemand einen Knopf drückt.
  const meldung = fehler ?? vorhandenes.fehler

  // **Die Symmetrie** —
  // #1092.
  //
  // `service.SymmetrieAbbilden` war portiert, geprüft und hatte **keinen
  // produktiven Aufrufer**; der Editor bot weder Art noch Kopienzahl an und
  // zeichnete keine Kopien.
  const klagen = useCallback(
    (ursache: unknown) => {
      setzeFehler(meldungAus(ursache, texte))
    },
    [texte],
  )

  // **Gruppen, Reihenfolge und die Mehrfachwahl** —
  // #1168.
  //
  // **Die Wahl steht hier und nicht in der Liste**: die Zeichenfläche zeigt
  // dieselbe Wahl, und zwei Zustände für eine Frage liefen auseinander.
  const liste = useTeilliste({ teile: gezeigte, eintragen, klagen, schutz })

  const { gewaehlt } = liste

  // waehlen und nurEines sind die beiden Wege in dieselbe Wahl.
  //
  // **`nurEines` für die Zeichenfläche**: dort trifft ein Klick eine Form, und
  // eine Mehrfachwahl über die Fläche wäre MSEs `SELECT_IF_OUTSIDE` — das
  // gehört zum Ziehen und ist hier nicht gebaut.
  const { waehlen } = liste

  const nurEines = useCallback(
    (stelle: number) => {
      waehlen(stelle, false)
    },
    [waehlen],
  )

  // **Die beiden Wege, die Teile hinzufügen** — Grundform und Nachzeichnen.
  const zugaenge = useSymbolzugaenge({
    gezeigte,
    eintragen,
    schutz,
    waehlen,
    klagen,
  })

  const { hinzufuegen, nachzeichnen } = zugaenge

  // **Die Punktwege** — verschieben, setzen, einfügen, löschen, Griffe ziehen.
  const punkte = usePunktwege({ teile: gezeigte, eintragen, klagen, schutz })

  const symmetrie = useSymmetrie({
    teile: gezeigte,
    gewaehlt,
    eintragen,
    klagen,
    schutz,
  })

  const loeschen = useCallback(
    (stelle: number) => {
      eintragen(
        gezeigte.filter((_, wo) => wo !== stelle),
        { marke: `loeschen:${String(stelle)}` },
      )
      liste.nachWegfall(stelle)
    },
    [eintragen, gezeigte, liste],
  )

  const ablage = useSymbolablage({
    setID,
    ziel,
    bibliothek,
    teile: gezeigte,
    geladenerName: vorhandenes.daten?.name ?? '',
    name,
    setzeName,
    setzeFehler,
    vorgang,
  })

  return (
    <section className="flex flex-col gap-4">
      <header className="flex flex-col gap-1">
        {/* **„Größer“ in der Schublade** — AP-12 (#2369), B-26: MSE2 hat
            ein eigenes Fenster, das mit der Fläche wächst (`window.cpp` 68).
            Ein Umschalter: der Name bleibt, `aria-pressed` sagt den Stand. */}
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-xl font-semibold">{worte.ueberschrift}</h2>
          {weite !== undefined && (
            <Knopf
              gewicht="leise"
              gedrueckt={weite.weit}
              aufKlick={weite.umschalten}
            >
              {texte.seiten.karten.symbol.groesser}
            </Knopf>
          )}
        </div>
        <Erklaerungsabsatz
          text={worte.erklaerung}
          klasse="text-sm text-leise"
        />
      </header>

      <Symbolwerkzeugleiste
        hinzufuegen={(art, ecken, eng) => void hinzufuegen(art, ecken, eng)}
        nachzeichnen={() => void nachzeichnen()}
        zuruecknehmen={verlauf.zuruecknehmen}
        wiederherstellen={verlauf.wiederherstellen}
        kannZuruecknehmen={verlauf.kannZuruecknehmen}
        kannWiederherstellen={verlauf.kannWiederherstellen}
      />

      {/* **Nebeneinander erst, wenn der Behälter Platz hat** — nicht das
          Fenster: die Seite steht in der Schublade *Eigenschaften* (UX-08,
          #1680), und dort stünde die Fläche sonst neben der Teilliste in
          einem Drittel ihrer Breite. */}
      <div className="flex flex-col gap-4 @3xl:flex-row">
        <Symbolflaeche
          gross={weite?.weit === true}
          teile={gezeigte}
          gewaehlt={gewaehlt}
          mitgewaehlt={liste.wahl}
          verschieben={punkte.verschieben}
          setzenAuf={punkte.setzenAuf}
          zugBeginnt={punkte.zugBeginnt}
          waehlen={nurEines}
          einfuegen={(teil, punkt) => {
            void punkte.punktEinfuegen(teil, punkt)
          }}
          loeschen={(teil, punkt) => {
            void punkte.punktLoeschen(teil, punkt)
          }}
          griffAuf={punkte.griffAuf}
          kopien={symmetrie.kopien}
          mitteAuf={symmetrie.mitteAuf}
          symmetriegriffAuf={symmetrie.griffAuf}
        />

        <Symmetriefeld
          teil={gezeigte[gewaehlt]}
          anlegen={symmetrie.anlegen}
          aufloesen={symmetrie.aufloesen}
          setzen={symmetrie.setzen}
        />

        <Symbolteilliste
          teile={gezeigte}
          geladen={vorhandenes.daten !== undefined}
          fehlschlag={vorhandenes.fehler !== undefined}
          wahl={liste.wahl}
          waehlen={waehlen}
          loeschen={loeschen}
          gruppieren={liste.gruppieren}
          aufloesen={liste.aufloesen}
          verschieben={liste.verschieben}
          kannGruppieren={liste.kannGruppieren}
          kannAufloesen={liste.kannAufloesen}
        />
      </div>

      <details className="text-sm text-leise">
        <summary className="cursor-pointer font-medium">
          {eigene.tastatur}
        </summary>
        <p className="mt-1">{eigene.tastaturhilfe}</p>
      </details>

      <div className="flex items-center gap-3">
        <Aktionsknopf
          aktion="sichern"
          aufKlick={() => void ablage.ablegen()}
          angehalten={
            gezeigte.length === 0 &&
            (bibliothek !== undefined || ablage.abgelegterName === '')
          }
          gewicht="haupt"
        >
          {worte.sichern}
        </Aktionsknopf>

        {name !== '' && (
          <p className="text-sm text-leise">
            {einsetzen(worte.gesichert, { name })}
          </p>
        )}
      </div>

      {fehler === undefined && vorhandenes.fehler !== undefined ? (
        <Ladefehler
          fehler={vorhandenes.fehler}
          wiederholen={vorhandenes.nachladen}
        />
      ) : (
        meldung !== undefined && (
          <p role="alert" className="text-sm text-fehler">
            {meldung}
          </p>
        )
      )}
    </section>
  )
}
