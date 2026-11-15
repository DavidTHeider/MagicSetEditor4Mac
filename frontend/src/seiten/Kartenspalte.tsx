// **Die Listenspalte der Kartenlasche** — herausgelöst, weil `Karten.tsx` mit
// ihr über 400 Zeilen ginge (§2).
//
// Der Schnitt folgt der Frage: dort steht, **was** die Lasche tut, hier, wie
// die linke Spalte aussieht — Liste und Notizen, so wie MSE2 sie stapelt
// (`splitter->SplitHorizontally(card_list, nodes_panel, -40)`,
// `src/gui/set/cards_panel.cpp`, Zeile 55, Fassung `a7336d6` vom 2020-06-14).
import { useState } from 'react'

import type { KarteDTO, KartenzeileDTO } from '../api'
import { inEingabe } from '../eingabefeld'
import type { Kartenablagegriff } from '../haken/useKartenablage'
import { useNurLesend } from '../haken/schreibrecht'
import { ablagetaste } from './ablagetasten'
import { useSprache, useTexte } from '../sprache/kontext'
import { useSchnellfilter } from '../haken/useSchnellfilter'
import { Schnellfilter } from './Schnellfilter'
import { useKartenvorauswahl } from '../haken/useKartenvorauswahl'
import { Statistikfilterhinweis } from './Statistikfilterhinweis'
import { Kartenliste } from './Kartenliste'
import { Listenspalten } from './Listenspalten'
import type { Listengestalt } from './listengestalt'
import { gefiltert, imSet, wirksam } from './markenauswahl'
import { Kartenmarken } from './Kartenmarken'
import { Markenfilter } from './Markenfilter'
import { Kartennotizen } from './Kartennotizen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Knopf } from '../bausteine/Knopf'
import { Auswahlfilter } from '../bausteine/Auswahlfilter'

interface KartenspalteEigenschaften {
  // spalten sind die zusätzlichen Spalten der Liste und der Wahlkasten
  // darunter — Werkstück 3 von Phase 3.7
  // (#1168).
  listengestalt: Listengestalt

  // ablage ist der Griff auf die Zwischenablage
  // (#1168).
  //
  // **Die Kürzel hängen an dieser Spalte**, nicht an der Liste darin — sie
  // ist, was MSE2 `CardsPanel` nennt, und dort steht die Fallunterscheidung:
  // Liste, Editor oder Notizen, je nachdem, wo der Fokus steht.
  ablage: Kartenablagegriff

  kennung: string
  setID: string
  karten: readonly KartenzeileDTO[]
  gewaehlt: string | undefined
  waehlen: (karteID: string) => void
  angehakt: readonly string[]
  anhaken: (karteID: string, an: boolean) => void

  // geoeffnet ist die aufgeschlagene Karte **mit ihren Werten** —
  // `undefined`, solange keine offen oder noch keine geladen ist.
  //
  // **Und nicht zusätzlich die Zeile aus der Liste.** Sie steht früher zur
  // Verfügung, und genau das war der Fehler: der Notizkasten baute sich mit
  // ihr auf, bevor die Werte da waren.
  geoeffnet: KarteDTO | undefined

  // stand steigt bei jeder Änderung — die Liste liest mit ihm nach. **Im
  // Schlüssel von Notiz- und Markenkasten steht er nicht mehr** (U04, U05,
  // #2121): er nahm dem, der tippte, den Entwurf.
  stand: number

  notizGeaendert: () => void
}

export function Kartenspalte({
  listengestalt,
  ablage,
  kennung,
  setID,
  karten,
  gewaehlt,
  waehlen,
  angehakt,
  anhaken,
  geoeffnet,
  stand,
  notizGeaendert,
}: KartenspalteEigenschaften) {
  // **Die gewählte Marke steht hier** (#1208) — bei der Liste, die sie
  // filtert, und nicht eine Ebene höher: sie ändert nichts am Bestand und
  // gehört zu dieser Spalte wie die Rollstelle.
  //
  // # Und ein Setwechsel räumt sie — durch das Abbauen
  //
  // §7 Regel 1 verlangt, dass ein Wechsel des Bezugsobjekts den offenen
  // Kontext mitschließt. Hier steht **kein** Schlüssel und kein Rückruf, der
  // das täte; es geschieht, weil `Karten` diese Spalte gar nicht erst
  // zeichnet, solange die neue Liste unterwegs ist
  // (`{karten.daten && karten.daten.length > 0 && …}`). Der Zustand geht mit
  // dem Bauteil.
  //
  // **Dasselbe gilt für den Ausklapper „Arbeitsstand"** weiter unten: er ist
  // ein unkontrolliertes `<details>` und behielte sein `open`, wenn er
  // stehen bliebe.
  //
  // **Das ist tragend und unsichtbar, und deshalb steht es hier.** Wer die
  // Liste beim Laden stehen lassen will — gegen das Flackern —, nimmt beide
  // Zusagen mit weg, ohne eine Zeile anzufassen, die davon spricht.
  //
  // Gehalten wird es von `kontextraeumen.test.tsx`. Der Zweitblick über die
  // Nacharbeit hat beides als Befund erhoben
  // (#1278,
  // Befunde 4 und 6); nachgestellt im laufenden Fenster war der Filter nach
  // dem Setwechsel leer (17 Zeilen statt 1) und der Ausklapper zu — und in
  // einem anderen DOM-Knoten. Der Befund traf also nicht zu, die **Lücke im
  // Nachweis** aber schon.
  const sprache = useSprache()
  const [markenfilter, setzeMarkenfilter] = useState('')

  // **Die wirksame Marke ist abgeleitet** (#1208): eine, die es nicht mehr
  // gibt, filtert nicht — sonst stünde die Liste leer da, und der Knopf für
  // »Alle« wäre mit ihr verschwunden.
  const angeboteneMarken = imSet(listengestalt.marken, sprache)
  const wirksameMarke = wirksam(angeboteneMarken, markenfilter)

  // **Der Notizfehler steht hier und nicht im Kasten** (#1215, Befund 3).
  //
  // Der Kasten wird beim Kartenwechsel abgebaut; diese Spalte nicht. Ein
  // Schreibfehler, der nach dem Wechsel eintrifft, hätte dort niemanden mehr,
  // dem er etwas sagen könnte.
  // **Die Liste kommt geordnet an**: `Karten` ordnet **einmal** für alle.
  // Stellte erst dieses letzte Kind um, arbeiteten die aufgeschlagene Karte,
  // die Nachfolgerin nach dem Löschen und die Zwischenablage weiter mit der
  // Reihenfolge aus der Datei.
  const [notizfehler, setzeNotizfehler] = useState<string>()
  const texte = useTexte()
  const nurLesend = useNurLesend()

  // **Die Kästchen stehen erst im Auswahlmodus** — UXR-05
  // (#1913).
  //
  // Bis dahin trug jede gezeichnete Zeile eines. Das Sichtfenster zeichnet
  // die Zeilen im Bild und je acht davor und dahinter (`useSichtfenster`) —
  // bei einem großen Set rund zwei Dutzend, und jede war ein Feld mehr auf
  // der Fläche; R3 aus `docs/UIREVIEW.md` lässt zwölf zu. Gebraucht werden
  // sie für einen seltenen Weg, mehrere Karten zugleich zu ändern oder zu
  // kopieren; das Artboard *Karten* zeigt in der Liste keines.
  //
  // **Eine Anhakung von außen schaltet den Modus ein** — „Alle Karten
  // anhaken" aus der Menüleiste. Beim Zeichnen und nicht in einem Effekt,
  // aus demselben Grund wie die Wahl in `Karten.tsx`: einen Herzschlag lang
  // stünde sonst eine angehakte Liste ohne Kästchen da.
  //
  // **Und er bleibt an, bis jemand ihn abschaltet** — sonst verschwände das
  // letzte Kästchen unter dem Zeiger, sobald es abgehakt ist. Ein Setwechsel
  // räumt ihn mit, wie den Markenfilter: diese Spalte wird dabei abgebaut.
  const [auswahl, setzeAuswahl] = useState(false)

  if (angehakt.length > 0 && !auswahl) {
    setzeAuswahl(true)
  }

  // auswahlUmschalten verlässt den Modus **samt Anhakung**. Versteckt stehen
  // bliebe sie wirksam — ⌘C kopierte Karten, deren Haken niemand mehr sieht.
  const auswahlUmschalten = () => {
    if (auswahl) {
      for (const karteID of angehakt) {
        anhaken(karteID, false)
      }
    }

    setzeAuswahl(!auswahl)
  }

  // **Der Filter der Statistik und die Karten mit Befund liegen unter allen
  // anderen** — UXR-F6 (#2013) und AP-8 (#2365). Die Statistik wählt die
  // Karten eines Balkens (`StatsFilter`, `src/gui/set/stats_panel.cpp`
  // Zeilen 536–545, Fassung f9d9356 vom 2021-07-06); Marke und Schnellfilter
  // grenzen davon weiter ein. **Die Zeilen darüber** sagen, dass sie da sind.
  const vorauswahl = useKartenvorauswahl(setID, karten, gewaehlt, waehlen)
  const { statistik, befund } = vorauswahl
  const kartenDerGruppe = vorauswahl.karten

  // **Der Schnellfilter liegt über dem Markenfilter**: erst entscheidet Go,
  // welche Karten zur Anfrage passen, dann nimmt die Marke weiter weg (U5).
  //
  // **Und er fragt neu, wenn sich ändert, was er sieht** — die Werte (`stand`)
  // und die sichtbaren Spalten (Befund 1 des Zweitblicks über Phase G).
  const filter = useSchnellfilter(setID, stand, listengestalt.spalten)
  const nachMarke = gefiltert(
    kartenDerGruppe,
    listengestalt.marken,
    wirksameMarke,
  )
  const sichtbar =
    filter.treffer === undefined
      ? nachMarke
      : nachMarke.filter((karte) => filter.treffer?.has(karte.id))

  return (
    <div
      id={kennung}
      // **Die Liste haftet, solange die Spalten nebeneinander stehen**
      // (#667):
      // das Formular ist bei `magic-m15` fast 5.000 Punkte hoch, und wer weit
      // unten ein Feld bearbeitet, hatte die Liste längst aus dem Bild
      // geschoben.
      //
      // `self-start`, weil ein gedehntes Rasterfeld nicht haften kann.
      // **Die Liste bekommt die Höhe, die übrig ist** (#1271).
      //
      // Gemessen am 2026-08-28 bei 1280 × 800: die Spalte war auf `70vh`
      // gedeckelt — 560 px —, und darunter standen Notizkasten (150), Marken
      // (150) und Spaltenwahl (42) **offen**. Für die Liste blieben rund 180
      // px, also **vier Zeilen** von 103 Karten. Die Liste ist auf dieser
      // Lasche das, womit jemand anfängt; vier Zeilen sind kein Anfang.
      //
      // Jetzt: `calc(100vh-4rem)` statt `70vh`, und Notiz und Marken stehen
      // eingeklappt beieinander. Nachgemessen bleiben der Liste rund 620 px.
      //
      // **Seit UX-15 (#1771) haftet sie nicht mehr, sie steht**: der
      // Arbeitsplatz gibt jeder Spalte die Höhe bis an den Fensterrand, und
      // die Seite rollt nicht mehr unter ihr weg. `min-h-0` lässt die Spalte
      // schrumpfen, bis sie in ihre Region passt — dann rollt die Liste und
      // nicht die Region.
      className="flex min-h-0 flex-col rounded-lg border border-kante p-2"
    >
      {/* **Die Sammelstelle für ⌘C, ⌘X und ⌘V.**

          `display: contents` — sie hat keinen eigenen Kasten und lässt das
          Flexlayout darüber unberührt; sie ist nur da, um die Ereignisse ihrer
          Kinder entgegenzunehmen. **Nicht am `<section>` selbst**, denn
          `role="presentation"` nähme dem Bereich seine Landmarke, und der
          Listenschalter zeigt mit `aria-controls` darauf. */}
      <div
        role="presentation"
        className="contents"
        // **⌘C, ⌘X und ⌘V über Karten** (#1168).
        //
        // **Hier und nicht am Fenster**: MSE2 führt die drei am fokussierten
        // Bedienelement (`CardsPanel::doCopy` über `CUT_COPY_PASTE`,
        // `src/gui/set/cards_panel.cpp`, Zeilen 371–382, Fassung f9d9356 vom
        // 2021-07-06). Diese Spalte trägt Liste **und** Notizen — also genau
        // die Wahl, die das Makro trifft.
        //
        // **Die Näherung steht dazu**: MSE2 fragt drei Kennungen ab, wir fragen
        // „steht der Fokus in einem Feld, das tippt?". Wer im Notizkasten ⌘C
        // drückt, bekommt seinen Text; wer daneben steht, seine Karten. Was
        // MSE2 zusätzlich kann und wir nicht — ein ⌘C, das *nichts* tut, weil
        // der Fokus auf keinem der drei liegt —, wäre ein Unterschied ohne
        // Wirkung: hier gibt es nichts Viertes in dieser Spalte.
        //
        // **Kein Fensterhorcher.** Ein `window.addEventListener` hinge erst
        // nach dem Zeichnen — ein Tastendruck davor ginge ungesehen verloren.
        onKeyDown={(ereignis) => {
          const tat = ablagetaste({
            taste: ereignis.key,
            meta: ereignis.metaKey,
            imEingabefeld: inEingabe(ereignis.target),
            nurLesend,
          })

          if (tat === null) {
            return
          }

          // **Erst hier**, nicht davor: ⌘C ohne Treffer gehört dem Browser.
          ereignis.preventDefault()
          ablage[tat]()
        }}
      >
        {/* **Die Statistik zuerst** (UXR-F6): ihr Filter liegt unter allen
            anderen, und seine Zeile steht über ihnen. Ohne Wahl zeichnet sie
            nichts. */}
        {befund !== undefined && (
          <div className="mb-2">
            <Auswahlfilter
              name={befund.name}
              loesen={texte.seiten.karten.auswahlLoesen}
              aufLoesen={vorauswahl.loeseBefund}
            />
          </div>
        )}
        <Statistikfilterhinweis
          filter={statistik}
          gesamt={karten.length}
          gezeigt={sichtbar.length}
        />
        {/* **Der Filter steht über der Liste** (#1208) — dort, wo man sieht,
            was er wegnimmt. Ohne Marken im Set zeichnet er nichts. */}
        <Schnellfilter
          filter={filter}
          gesamt={karten.length}
          gezeigt={sichtbar.length}
        />
        <div className="mb-2">
          <Knopf
            gewicht="leise"
            klein
            gedrueckt={auswahl}
            aufKlick={auswahlUmschalten}
          >
            {texte.seiten.karten.massen.mehrereWaehlen}
          </Knopf>
        </div>
        <Markenfilter
          marken={angeboteneMarken}
          gewaehlt={wirksameMarke}
          waehlen={setzeMarkenfilter}
        />

        <Kartenliste
          spalten={listengestalt.zusaetze}
          marken={listengestalt.marken}
          setID={setID}
          stand={stand}
          karten={sichtbar}
          gewaehlt={gewaehlt}
          waehlen={waehlen}
          angehakt={angehakt}
          anhaken={anhaken}
          mitKaestchen={auswahl}
        />

        {/* **Die Notizen stehen unter der Liste** — genau dort, wo MSE2 sie hat.
          Sie gehören zur **Karte** und nicht zu ihren Feldwerten (#1166).

          # Der Schlüssel hängt an der **geladenen** Karte

          Der Kasten hält seinen Text im Zustand und nimmt den Anfangstext beim
          Aufbau aus den Werten. Ein Klick in die Liste ändert die **gewählte**
          Kennung sofort — sie kommt aus den Zeilen, die schon dastehen —, die
          **Werte** kommen danach. Stand `aufgeschlagen.id` im Schlüssel, baute
          sich der Kasten mit den Werten der **vorigen** Karte auf und blieb
          dabei: nach jedem Wechsel stand ein leerer Kasten da, obwohl jede
          Karte des kanonischen `Vanguard` ein `notes` trägt.

          **Deshalb `geoeffnet.id`**, und deshalb steht der Kasten erst da,
          wenn die Werte da sind. Das ist zugleich §7 Regel 1 und Regel 2 in
          einem: ein Wechsel räumt ihn, und was er zeigt, gehört zu dem Stand,
          aus dem es stammt. */}
        {geoeffnet !== undefined && (
          <details className="mt-2 shrink-0 border-t border-kante pt-2">
            <summary className="cursor-pointer text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
              {texte.seiten.karten.arbeitsstand}
            </summary>

            <Erklaerungsabsatz
              text={texte.seiten.karten.arbeitsstandHinweis}
              klasse="mt-1 text-sm text-leise"
            />

            <Kartennotizen
              key={geoeffnet.id}
              setID={setID}
              karteID={geoeffnet.id}
              werte={geoeffnet.werte}
              geaendert={notizGeaendert}
              aufFehler={setzeNotizfehler}
            />

            <Kartenmarken
              key={`marken/${geoeffnet.id}`}
              setID={setID}
              karteID={geoeffnet.id}
              marken={listengestalt.marken[geoeffnet.id] ?? keineMarken}
              // **Die Gestalt lädt nach und nicht die Karte** (#1208): die
              // Marken reisen mit den **Zeilen** der Liste, und
              // `notizGeaendert` holt nur die aufgeschlagene Karte. Mit ihm
              // blieb die eben gesetzte Marke unsichtbar, bis jemand das Set
              // wechselte — im laufenden Fenster gesehen, nicht überlegt (§3).
              geaendert={listengestalt.geaendert}
            />
          </details>
        )}

        {/* **Die Spaltenwahl steht unter der Liste** (#1168) — dort, wo man
            sie sieht, während man die Liste ansieht. */}
        <Listenspalten
          setID={setID}
          spalten={listengestalt.spalten}
          sortierfeld={listengestalt.sortierfeld}
          aufsteigend={listengestalt.aufsteigend}
          fehler={listengestalt.fehler}
          ohneVorlage={listengestalt.ohneVorlage}
          vorgang={listengestalt.vorgang}
          geaendert={listengestalt.geaendert}
        />

        {notizfehler !== undefined && (
          <p role="alert" className="mt-1 text-sm text-fehler">
            {notizfehler}
          </p>
        )}
      </div>
    </div>
  )
}

// keineMarken ist die geteilte leere Liste — **eine** und nicht je Zeichnung
// eine neue, damit ein Vergleich nach Gleichheit nicht bei jedem Durchgang
// scheitert.
const keineMarken: readonly string[] = []
