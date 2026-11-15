// Die Kartenliste eines Sets, samt den Vorbehalten aus dem Import.
//
// **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7, Regel 1):
// Karten und Vorbehalte hängen beide an `setID`, und `useLaden` lädt beide
// neu, sobald sie sich ändert. Ein Rest des vorigen Sets bleibt nicht stehen.
import { useState } from 'react'

import { useKartendaten } from '../haken/useKartendaten'
import { useAktuellesSet } from '../haken/useAktuellesSet'
import { useKartenwerkzeug } from '../haken/useKartenwerkzeug'
import { Arbeitsplatz } from '../bausteine/Arbeitsplatz'
import { useArbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'
import { useVorgang } from '../haken/useVorgang'
import { useKartenanlage } from '../haken/useScryfallkarte'
import { useTexte } from '../sprache/kontext'
import { gewaehlteKarte, leereKarte, nachrueckendeKarte } from './kartenauswahl'
import { useSeitenweg } from '../seitenweg'
import type { KartenEigenschaften } from './kartenpropse'
import { Kartenspalte } from './Kartenspalte'
import { Kartenarbeitsflaeche } from './Kartenarbeitsflaeche'
import { geordnet } from './listenordnung'
import { Kartenwerkzeuge } from './Kartenwerkzeuge'
import { useScryfallvorschlag } from '../haken/useScryfallvorschlag'
import { Ladefeld } from './Ladefeld'
import { Kartenleer } from './Kartenleer'
import { Kartenanhang } from './Kartenanhang'
import { grundansicht } from './vorschauansicht'
import { useKartenskripte } from '../haken/useKartenskripte'
import { useAllesAnhaken } from '../haken/useAllesAnhaken'
import { useKartenblaettern } from '../haken/useKartenblaettern'
import { useKartenmeldung } from '../haken/useKartenmeldung'

// listenkennung ist die Kennung der Listenspalte — `aria-controls`.
const listenkennung = 'kartenliste'

export function Karten({
  setID,
  karteID,
  karteGewaehlt,
  symbolZeichnen,
  bestandGeaendert,
  verlaufsstand,
  verlaufsfehler,
  zuruecknehmen,
}: KartenEigenschaften) {
  const texte = useTexte()

  // **Der Wechsel des Sets räumt die gewählte Karte mit** (§7, Regel 1). Ohne
  // das stünde der Editor einer Karte offen, die zum neuen Set nicht gehört —
  // und sein erstes Schreiben ginge an ein fremdes Set.
  //
  // Die Vorgabe von außen wird im selben Zug übernommen: sie kommt zusammen mit
  // dem Set, und zwei getrennte Prüfungen räumten sie in derselben Runde wieder
  // weg, in der sie gesetzt wurde.
  const [gewaehlt, setzeGewaehlt] = useState(karteID)

  // **Die Anhakung** — geräumt beim Setwechsel, gesetzt vom Menüeintrag. Was
  // sie leistet und warum, steht bei [useAnhakung].
  const [angehakt, setzeAngehakt] = useState<readonly string[]>([])

  // anlegefehler ist die Meldung der **Handlungszeile** — sie steht an der
  // Handlung, die sie verursacht hat (§6), und nicht als anonymer Hinweis über
  // dem Fenster.
  //
  // **Sie gehört nicht mehr allein dem Anlegen** (#1167): der Verweiskasten
  // meldet hierher, weil er den Kartenwechsel nicht überlebt und seine Meldung
  // sonst verlöre. Der Name bleibt, damit der Diff lesbar ist; was er trägt,
  // sagt dieser Satz.
  const [anlegefehler, setzeAnlegefehler] = useState<string>()

  // **Die Liste lässt sich wegklappen**
  // (#697).
  //
  // Gemessen an der laufenden Anwendung bei 1.440 Punkten: die Liste nimmt
  // 224, und was bleibt, teilen sich Formular und Vorschau zu je 407 — für
  // zwei Formularspalten reicht das nicht. Wer sie wegklappt, bekommt 679 und
  // damit zwei Spalten à 324.
  //
  // **Nicht gemerkt, und das ist eine Entscheidung.** Der Ort dafür wären die
  // Einstellungen des Bestands (#245),
  // nicht ein `localStorage`; ein Schalter für eine Spalte ist eine
  // Schemaänderung nicht wert. Er steht offen da, sooft jemand die Lasche
  // öffnet — und das ist der Zustand, in dem der Weg zur nächsten Karte einen
  // Klick kostet.
  // **Der Arbeitsplatz** (U5, #1416): Navigator, Zentrum, Inspector — mit dem
  // Gedächtnis je Profil aus U4. Das Formular des Editors geht per Portal in
  // den Inspector; `inspectorZiel` ist sein Knoten.
  const gedaechtnis = useArbeitsplatzgedaechtnis()
  const [inspectorZiel, setzeInspectorZiel] = useState<HTMLElement | null>(null)
  // ansichtsort ist der Platz der Schalter in der Werkzeugzeile (P03).
  const [ansichtsort, setzeAnsichtsort] = useState<HTMLDivElement | null>(null)

  const vorgang = useVorgang()

  // **Regel 2 aus §7: eine Antwort gehört zu dem Stand, aus dem sie stammt**
  // (#716).
  //
  // `Karten` bleibt beim Setwechsel **dieselbe** Komponente: der Block darunter
  // räumt Auswahl und Anhakung, aber eine laufende Handlung weiß davon nichts.
  // Wer eine Karte anlegt und währenddessen das Set wechselt, bekam die Karte
  // des **alten** Sets aufgeschlagen, die Liste nachgeladen und den Bestand als
  // geändert gemeldet — angestrichen vom Zweitblick durch Codex am 2026-08-21.
  //
  // **Für Erfolg und Fehler gleichermaßen.** Die Meldung ist die unauffälligere
  // Hälfte: „ließ sich nicht anlegen" zu einem Set, das niemand mehr offen hat,
  // sieht aus wie ein Fehler des aktuellen.
  const { nochAktuell, nurAktuell } = useAktuellesSet(setID)

  const [vorlage, setzeVorlage] = useState({ setID, karteID })

  if (vorlage.setID !== setID || vorlage.karteID !== karteID) {
    setzeVorlage({ setID, karteID })
    setzeGewaehlt(karteID)

    if (vorlage.setID !== setID) {
      setzeAngehakt([])
      setzeAnlegefehler(undefined)
    }
  }

  // **Die Wahl gilt auch im Fenster** (#892) — **jede**, nicht nur die per
  // Klick: die Skriptkonsole wertet in ihrem Kontext aus, und sie meint auch
  // die erste Karte, die das Öffnen aufschlägt. Warum als Effekt, steht am
  // Haken.
  useKartenmeldung(gewaehlt, karteID, karteGewaehlt)

  const {
    karten,
    liste: rohliste,
    formen,
    geoeffnete,
  } = useKartendaten({
    setID,
    verlaufsstand,
    gewaehlt,
  })

  // **Ein geöffnetes Set schlägt seine erste Karte auf**
  // (#668).
  //
  //	void SetWindow::onChangeSet() {
  //	  …
  //	  // only after setSet select a card
  //	  FOR_EACH(p, panels) {
  //	    p->selectFirstCard();
  //	  }
  //
  // `src/gui/set/window.cpp`, Zeilen 305–320, weiter über
  // `CardsPanel::selectFirstCard` und `ItemList::selectFirst`, Fassung
  // f9d9356 vom 2021-07-06.
  //
  // **Beim Zeichnen und nicht in einem Effekt** — dieselbe Begründung wie beim
  // Wechselwächter darüber: ein Effekt liefe nach dem Zeichnen, und einen
  // Herzschlag lang stünde „keine Karte gewählt" da.
  //
  // **An „noch nichts gewählt" und nicht an „nicht in der Liste".**
  //
  // Der erste Anlauf fragte, ob die *aufgeschlagene* Karte in der Liste steht
  // — und griff damit in jeden Zwischenstand, in dem sie es noch nicht tut.
  // Wer eine Karte anlegt, bekommt sie zugewiesen, **bevor** die Liste
  // nachgeladen ist (§7, Regel 3); der Zweig sprang zurück auf die erste, und
  // die neue Karte war für einen Herzschlag nicht die offene.
  //
  // **Wer eine Karte mitgibt, behält sie** — `gewaehlt` ist dann gesetzt.
  //
  // **Und eine Karte, die verschwindet, holt keine neue herbei**: nach einem
  // ⌘Z auf das Anlegen bleibt `gewaehlt` stehen und zeigt ins Leere, und die
  // Seite sagt „keine Karte gewählt", statt eine fremde aufzuschlagen.
  // **Und erst, wenn die Liste zu diesem Set gehört.** Ohne `laedt` schlüge ein
  // Setwechsel die erste Karte des **alten** Sets auf.
  //
  // **`useLaden` lässt die vorige Liste nur beim Nachladen stehen**: beim
  // Wechsel des Bezugsobjekts räumt er (§7, Regel 1), und deshalb
  // verschwindet die ganze Spalte unten für einen Herzschlag.
  // **Das gewählte Feld gehört über den Editor** (#1075): der wurde bei
  // jedem Verlaufsstand neu gebaut, und mit dem Neuaufbau ging das
  // Auswahlmenü — gemessen: `fieldset` fiel von 1 auf 0. Seit U01 (#2119)
  // baut ihn nur ein Kartenwechsel neu; die Wahl bleibt trotzdem hier, denn
  // auch dann soll sie nicht von vorn anfangen.
  const feldwahl = useState<string>()

  // **Maßstab und Überlagerungen der Vorschau ebenso** (UXR-06, #1914): wer
  // auf 200 % prüfte, stand nach dem nächsten Feld wieder bei „Anpassen".
  const ansicht = useState(grundansicht)

  // **Die aufgeschlagene Karte hängt nicht an der Reihenfolge** — sie wird
  // über ihre Kennung gesucht. Deshalb steht sie hier und nicht unten: die
  // Handlungen der Lasche brauchen sie, und die Ordnung kommt erst mit ihnen.
  const inBearbeitung = gewaehlteKarte(rohliste, gewaehlt)

  // **Über den Karten steht die aufgeschlagene** — UX-03.AK-02, wie im
  // Artboard *Karten*; eine ohne Namen lässt „Karten" stehen.
  useSeitenweg(
    inBearbeitung?.name ? [{ beschriftung: inBearbeitung.name }] : undefined,
  )

  // **Die Handlungen der Lasche** (#1168) — Zwischenablage und Listengestalt.
  // Sie stehen in `kartenwerkzeug.ts`, weil diese Datei sonst über 400 Zeilen
  // ginge (§2); der Schnitt folgt einer Frage: hier, **wie** die Lasche
  // angeordnet ist, dort, **womit** sie arbeitet.
  // **Der Scryfall-Vorschlag gehört hierher**, seit die Handlungen getrennt
  // sind (#1240):
  // der Anlegekasten oben liest seinen Namen, der Verweiskasten an der Karte
  // seine Kennung. Der Grund steht am Haken.
  const vorschlag = useScryfallvorschlag(gewaehlt)

  const { ablage, listengestalt } = useKartenwerkzeug({
    setID,
    angehakt,
    aufgeschlagen: inBearbeitung?.id,
    verlaufsstand,
    vorgang,
    nurAktuell,
    setzeAnlegefehler,
    nachladen: karten.nachladen,
    bestandGeaendert,
  })

  // **Die berechnete Ordnung gilt für die ganze Lasche.** In `Kartenspalte`,
  // dem letzten Kind, wäre sie nur eine **Darstellung**: die erste
  // aufgeschlagene Karte, die Nachfolgerin nach dem Löschen und die
  // Reihenfolge der Zwischenablage kämen weiter aus `cards.reihenfolge`.
  // MSE2 führt dafür **eine** `sorted_list`, aus der Auswahl, Vorher/Nachher,
  // Löschnachfolge und Kopierreihenfolge lesen
  // (`src/gui/control/item_list.cpp`, Zeilen 37–64 und 87–100, Fassung
  // `f9d9356` vom 2021-07-06).
  const liste = geordnet(rohliste, listengestalt.ordnung)

  const erste = karten.laedt ? undefined : liste[0]

  if (gewaehlt === undefined && erste !== undefined) {
    setzeGewaehlt(erste.id)
  }

  const gewaehlteKarten = liste.filter((karte) => angehakt.includes(karte.id))

  // **Alle Karten anhaken** — der Menüeintrag aus dem Zweig *Set* (#1242).
  useAllesAnhaken(liste, setzeAngehakt)

  // **⇞ und ⇟ blättern** (#1290) — auch während des Tippens; warum, steht beim
  // Haken.
  useKartenblaettern(liste, gewaehlt, setzeGewaehlt)

  // **Wie eine Karte entsteht, steht woanders** (§2). Beide Wege — leer
  // anlegen und aus Scryfall holen — enden gleich: die neue Karte wird
  // aufgeschlagen, die Liste lädt nach, der Bestand meldet sich. Zweimal
  // derselbe Verbund nebeneinander war die Zeile, die diese Datei über die
  // Grenze trug.
  const kartenskripte = useKartenskripte(setID, () => {
    karten.nachladen()
    bestandGeaendert()
  })

  const { anlegen, ausScryfall } = useKartenanlage({
    setID,
    vorgang,
    texte,
    nochAktuell,
    aufgeschlagen: (neue) => {
      setzeGewaehlt(neue)
      karten.nachladen()
      bestandGeaendert()
    },
    gemeldet: setzeAnlegefehler,
  })

  return (
    <div className="flex h-full flex-col gap-4">
      {/* **Die Meldung steht dort, wo ⌘Z wirkt** (§6) — nicht als anonymer
          Hinweis über dem Fenster. */}
      {verlaufsfehler !== undefined && (
        <p role="alert" className="text-fehler">
          {verlaufsfehler}
        </p>
      )}

      {/* **Die Handlungen der Lasche stehen oben** — in eigener Datei (§2):
          dort **was**, hier **wie** sie angeordnet sind. */}
      <Kartenwerkzeuge
        setID={setID}
        kartenskripte={kartenskripte}
        vorgang={vorgang}
        ausScryfall={(name, art) => void ausScryfall(name, art)}
        anlegen={() => void anlegen()}
        bestandGeaendert={bestandGeaendert}
        anlegefehler={anlegefehler}
        gewaehlt={vorschlag.treffer?.name ?? ''}
        waehlenTreffer={vorschlag.waehlen}
        ansichtsort={karten.daten?.length ? setzeAnsichtsort : undefined}
      />

      <Ladefeld
        laedt={karten.laedt && karten.daten === undefined}
        fehler={karten.fehler}
        leer={karten.daten?.length === 0}
      >
        <Kartenleer ablage={ablage} />
      </Ladefeld>

      {/* **Drei Regionen: Navigator, Zentrum, Inspector** — der Arbeitsplatz
          aus U2 (0076), mit der Kartenliste links, der Vorschau in der Mitte
          und den Feldern rechts: **die Karte ist das Hauptobjekt** (§6) und
          steht in der Mitte. Unter 1.200 Punkten klappt der Navigator von
          selbst ein; was jemand klappt, merkt sich das Profil (U4). */}
      {karten.daten && karten.daten.length > 0 && (
        <Arbeitsplatz
          bezeichnung={texte.seiten.karten.arbeitsplatz}
          gedaechtnis={gedaechtnis}
          schalterort={ansichtsort}
          navigator={{
            bezeichnung: texte.seiten.karten.listeBezeichnung,
            inhalt: (
              <Kartenspalte
                listengestalt={listengestalt}
                ablage={ablage}
                kennung={listenkennung}
                setID={setID}
                // **Die geordnete Liste** und nicht `karten.daten`, der Stand
                // aus der Datei — sonst erreichte die Ordnung die Spalte gar
                // nicht.
                karten={liste}
                gewaehlt={gewaehlt}
                waehlen={setzeGewaehlt}
                angehakt={angehakt}
                anhaken={(kennung, an) => {
                  setzeAngehakt((bisher) =>
                    an
                      ? [...bisher, kennung]
                      : bisher.filter((eine) => eine !== kennung),
                  )
                }}
                geoeffnet={geoeffnete.daten}
                stand={verlaufsstand}
                notizGeaendert={() => {
                  geoeffnete.nachladen()
                }}
              />
            ),
          }}
          zentrum={
            <Kartenarbeitsflaeche
              symbolZeichnen={symbolZeichnen}
              setID={setID}
              inBearbeitung={inBearbeitung}
              karten={liste}
              geoeffnete={geoeffnete.daten}
              formen={formen.daten ?? []}
              verlaufsstand={verlaufsstand}
              vorgang={vorgang}
              ablage={ablage}
              vorschlag={vorschlag.treffer}
              feldwahl={feldwahl}
              ansicht={ansicht}
              aufFehler={nurAktuell(setID, (meldung: string | undefined) => {
                setzeAnlegefehler(meldung)
              })}
              waehlen={nurAktuell(setID, setzeGewaehlt)}
              anlegen={() => void anlegen()}
              nachruecken={nurAktuell(setID, (geloescht: string) => {
                setzeGewaehlt((jetzt) =>
                  jetzt === geloescht
                    ? nachrueckendeKarte(liste, geloescht)
                    : jetzt,
                )
              })}
              nachladen={nurAktuell(setID, () => {
                karten.nachladen()
              })}
              fehlerRaeumen={nurAktuell(setID, () => {
                setzeAnlegefehler(undefined)
              })}
              bestandGeaendert={nurAktuell(setID, () => {
                bestandGeaendert()
              })}
              geaendert={() => {
                karten.nachladen()
                geoeffnete.nachladen()
              }}
              leereKarte={leereKarte}
              formularZiel={inspectorZiel}
            />
          }
          inspector={{
            bezeichnung: texte.seiten.karten.inspectorBezeichnung,
            inhalt: <div ref={setzeInspectorZiel} className="contents" />,
          }}
        />
      )}

      <Kartenanhang
        setID={setID}
        gewaehlteKarten={gewaehlteKarten}
        nachladen={karten.nachladen}
        karten={liste}
        verlaufsstand={verlaufsstand}
        zuruecknehmen={zuruecknehmen}
      />
    </div>
  )
}
