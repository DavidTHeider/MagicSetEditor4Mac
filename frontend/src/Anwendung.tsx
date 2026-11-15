// **Die Anwendung mit offenem Bestand** — Rahmen, Seiten, Meldungsband.
//
// **Eine eigene Datei**, weil `App.tsx` über die Zeilengrenze wuchs, als der
// Weg vom Finder dazukam (§2). Der Schnitt folgt der Frage: dort steht, **ob**
// überhaupt gezeichnet wird, hier, **was** dann dasteht. Am 2026-09-03 sind
// `useFensterhandlungen` und `Starteinfuehrung` herausgezogen.
import { useCallback, useState } from 'react'

import { Starteinfuehrung } from './bausteine/Starteinfuehrung'
import { useAktuellesSet } from './haken/useAktuellesSet'
import { useEinfuehrung } from './haken/useEinfuehrung'
import { useFensterhandlungen } from './haken/useFensterhandlungen'
import { useMenueziele } from './haken/useMenueziele'
import { useSeitenleiste } from './haken/useSeitenleiste'
import { Rahmen } from './rahmen/Rahmen'
import { Seiteninhalt } from './seiten/Seiteninhalt'
import { useFensterereignisse } from './fenster'
import { useAbgelegteDateien } from './haken/useAbgelegteDateien'
import { useDateioeffnung } from './haken/useDateioeffnung'
import { OffenesSet, useFensterset } from './haken/offenesset'
import { useSetwahl } from './haken/useSetwahl'
import { useBestandspflege } from './haken/useBestandspflege'
import { useDatenpflege } from './haken/useDatenpflege'
import { useRuecksprung } from './haken/useRuecksprung'
import { useVorgang } from './haken/useVorgang'
import { useSetsicherung } from './haken/useSetsicherung'
import { useVerlauf } from './haken/useVerlauf'
import { Anschlussraum } from './anschlussraum'
import { useSeitengang } from './haken/useSeitengang'
import { useSymbolziel } from './haken/useSymbolziel'
import { useKartenoeffnung } from './haken/useKartenoeffnung'
import { useRuecknahmeziel } from './ruecknahme'
import { ruecknahmeweg } from './ruecknahmeweg'
import { type Einstellungsstand } from './haken/useEinstellungen'
import { type Sprache } from './sprache/kataloge'
import { useMeldungsband } from './haken/useMeldungsband'
import { belegungAus } from './rahmen/ziffernbelegung'

export function Anwendung({
  sprache,
  fassung,
  rueckspielrest,
  einstellungen,
  meldefehler,
  bestandErsetzt,
  fremderstand,
}: {
  sprache: Sprache

  // fassung ist die laufende Fassung — sie steht im Abschnitt „Über" der
  // Hilfe (#1168).
  fassung: string

  // rueckspielrest sagt, dass ein abgebrochenes Zurückspielen beim Start nicht
  // abgeschlossen werden konnte — sonst leer (#1170).
  rueckspielrest: string

  einstellungen: Einstellungsstand

  // meldefehler ist, was beim Melden der Sprache an Go schieflief. **Gemeldet
  // wird in `App`** und nicht hier — sonst meldete das Wiederherstellungsfenster
  // nichts, und seine Menüleiste bliebe deutsch (Zweitblick über Phase 6.2).
  // Gezeigt gehört der Befund dort, wo jemand die Sprache wählt.
  meldefehler: string | undefined

  // bestandErsetzt meldet nach oben, dass **alles** ein anderes ist.
  //
  // Er sitzt über dem `Sprachraum`, weil die Einstellungen — und damit die
  // Sprache — dort geladen werden. Ein Zurückspielen kann sie ändern.
  bestandErsetzt: () => void

  // fremderstand zählt, was ein anderer Prozess oder die Leitung geschrieben
  // hat (#156). **Gezählt wird in `App`** und nicht hier: dort hängen auch die
  // Einstellungen daran, und ein Schalter, den die Kommandozeile umlegt, soll
  // im offenen Fenster wirken (Befund 2 des Gesamtreviews über 0.6.0).
  fremderstand: number
}) {
  // **Ob die Hilfe zum Abschnitt „Über" rollen soll** (#1168).
  //
  // **Ein Zustand und kein Anker in der Adresse.** Diese Anwendung hat keine
  // Adressleiste; ein `#ueber` stünde nirgends, und beim zweiten Ruf desselben
  // Eintrags täte der Browser nichts, weil sich die Adresse nicht ändert.
  const [zuUeber, setzeZuUeber] = useState(false)

  // **⌘O und „Öffnen …" stoßen an** (#891). Die Setliste hört darauf und
  // öffnet denselben Dateidialog wie ihr Knopf; eine zweite Fassung derselben
  // Frage liefe mit der ersten auseinander.
  //
  // **Und sie verbraucht ihn.** Die Seite entsteht neu, wenn jemand auf sie
  // wechselt — der Anstoß kommt an, bevor es die Seite gibt.
  const [oeffnungsanstoss, setzeOeffnungsanstoss] = useState(false)

  // **Und ⌘N** (#891) — dasselbe Muster: die Setliste setzt den Fokus ins
  // Namensfeld und verbraucht den Anstoß.
  const [neuanstoss, setzeNeuanstoss] = useState(false)
  const [setID, setzeSetID] = useState<string>()

  // **Die Seite samt ihrem Gang** — ⌘[ und ⌘] (UX-03.AK-02, #1675).
  const gang = useSeitengang(setID)
  const { aktuell, wechseln: setzeAktuell } = gang
  const [karteID, setzeKarteID] = useState<string>()

  // **Welches Symbol der Editor bearbeitet** (#1339) — und dass es sein Set
  // nicht überlebt (§7, Regel 1). Beides steht am Haken, weil beides
  // zusammengehört; die Begründung steht dort.
  const symbolziel = useSymbolziel(setID)

  // **Wohin eine aufgeschlagene Karte führt** — seit den Schubladen hängt es
  // davon ab, wo man steht (UX-08, #1680); die Regel steht am Haken.
  const { oeffneKarte, oeffneSymbol } = useKartenoeffnung({
    setzeSetID,
    setzeKarteID,
    symbolziel,
    setzeAktuell,
  })

  // **⌘Z und ⇧⌘Z stehen hier**, weil das Set hier steht: der Rahmen kennt keine
  // Bezugsobjekte, und die Kartenseite kennt die Tastenkürzel des Fensters
  // nicht (#52).
  const verlauf = useVerlauf(setID)

  // **Wohin ⌘Z zielt** (#1340). Der Symboleditor bringt einen eigenen Verlauf
  // mit; solange er dasteht, meint der Menübefehl **ihn**. Die Regel und ihr
  // Grund stehen in `ruecknahmeweg.ts`.
  const menueverlauf = ruecknahmeweg(useRuecknahmeziel(), {
    zuruecknehmen: verlauf.zuruecknehmen,
    wiederherstellen: verlauf.wiederherstellen,
  })

  // **Was dieses Fenster selbst ändert, zählt mit.**
  //
  // `useBestandsstand` fragt in Abständen nach; ein Stilwechsel auf der Lasche
  // „Stil" wäre damit erst beim nächsten Abruf sichtbar. Wer ihn auslöst, weiß
  // es sofort — und die Kartenlasche lädt beim Zurückwechseln neu (#288).
  const [eigenerstand, setzeEigenerstand] = useState(0)

  // **Die Einführung beim Start** (#1325) — wann sie steht, entscheidet
  // [Starteinfuehrung]; hier steht nur, ob sie schon geschlossen wurde.
  const einfuehrung = useEinfuehrung()

  // bestandGeaendert zählt den eigenen Stand hoch — dieselbe Zahl, die der
  // MCP-Server bewegt (#156).
  const bestandGeaendert = useCallback(() => {
    setzeEigenerstand((bisher) => bisher + 1)
  }, [])

  // **⌘S steht aus demselben Grund hier wie ⌘Z**: es gilt fürs ganze Fenster,
  // und welches Set gemeint ist, weiß nur diese Stelle (#162).
  const sicherung = useSetsicherung(setID)
  const { sichernJetzt, sichernUnter } = sicherung
  const speichern = sicherung.speichern

  const einstellungenOeffnen = useCallback(() => {
    setzeAktuell('einstellungen')
  }, [setzeAktuell])

  // **Vier Wege zu einem offenen Set, ein Laufstand** (#1111): ein neues Set
  // räumt die Karte des vorigen (§7, Regel 1); der Grund steht am Haken.
  const setwahl = useSetwahl(
    setzeSetID,
    setzeKarteID,
    useCallback(() => {
      setzeAktuell('karten')
    }, [setzeAktuell]),
    useCallback(() => {
      setzeEigenerstand((bisher) => bisher + 1)
    }, []),
  )

  // **Der Bestand ist kein Set** (#799): eigener Vorgang, eigene Meldung.
  //
  // **Und ein Vorgang für beide Wege darauf** (#940): sichern und
  // zurückspielen greifen auf dieselbe Datei, und Regel 4 aus §7 gilt je
  // Datensatz — nicht je Schaltfläche.
  const pflege = useBestandspflege(() => {
    // **Ein Zurückspielen wechselt das Bezugsobjekt vollständig** (§7,
    // Regel 1): das offene Set und die aufgeschlagene Karte gibt es im neuen
    // Bestand vielleicht gar nicht, und stehen bleiben dürfen sie nicht.
    setzeSetID(undefined)
    setzeKarteID(undefined)
    // **Was gerade aufgeht, gehört zum alten Bestand** (§7, Regel 2; #2292).
    setwahl.lauf.naechster()
    setzeEigenerstand((bisher) => bisher + 1)
    bestandErsetzt()
  }, bestandErsetzt)
  const bestand = pflege.sicherung

  // **Und was zurückgehen kann** (O7, #1431) — eigener Vorgang je Epoche, weil
  // eine Quittung je Rückkehr gilt und nicht für den ganzen Bestand.
  const ruecksprung = useRuecksprung(useVorgang())

  // **Und das Aufräumen** (O8, #1432) — eigener Vorgang, denn Nachsehen und
  // Aufräumen teilen sich denselben Datensatz: den Bestand.
  const datenpflege = useDatenpflege(useVorgang())

  // **Ein Klick auf das Kreuz räumt alles** (#799) — warum, steht in
  // `useMeldungsband.ts`, das die Darstellung dazunimmt.
  const meldungRaeumen = useCallback(() => {
    bestand.raeumen()
    speichern.raeumen()
  }, [bestand, speichern])

  // **Die native Menüleiste ist der zweite Weg zu denselben Handlungen**, nicht
  // der einzige: im Fenster gehen die Kürzel über `tastatur.ts`, in der
  // Browservorschau gibt es gar keine Leiste (#162).
  // **Ein Doppelklick im Finder landet hier** (#891). Das Set ist zu diesem
  // Zeitpunkt schon im Bestand — Go hat es eingelesen und meldet die Kennung.
  const datei = useDateioeffnung(sprache, setwahl.aufschlagen, setwahl.lauf)

  // **Die vier Handlungen der Ablage und der Hilfe** (#891).
  //
  // Sie stehen hier und nicht im Rahmen: welches Set offen ist, weiß diese
  // Ebene, und „schließen" heißt genau das — das **Set**, nicht das Fenster.
  const setOeffnen = useCallback(() => {
    setzeAktuell('sets')
    setzeOeffnungsanstoss(true)
  }, [setzeAktuell])

  const oeffnungAngestossen = useCallback(() => {
    setzeOeffnungsanstoss(false)
  }, [])

  const waehleSet = setwahl.waehleSet

  const setNeu = useCallback(() => {
    setzeAktuell('sets')
    setzeNeuanstoss(true)
  }, [setzeAktuell])

  const neuAngestossen = useCallback(() => {
    setzeNeuanstoss(false)
  }, [])

  // **⌘W schließt das Set, nicht das Fenster.**
  //
  // Eine Abweichung von MSE2, und sie ist begründet: dort gibt es gar kein
  // „Schließen" (`src/gui/set/window.cpp`, Fassung f9d9356 vom 2021-07-06).
  // Unter macOS ist ⌘W trotzdem belegt erwartet, und diese Anwendung hat ein
  // Fenster — es zu schließen hieße, sie zu beenden. Dafür gibt es ⌘Q.
  const setSchliessen = useCallback(() => {
    setzeSetID(undefined)
    setzeKarteID(undefined)
    setzeAktuell('sets')
  }, [setzeAktuell])

  // **Ein gelöschtes Set kann nicht das offene bleiben** (§7, Regel 1)
  // (#1218),
  // **und nur, wenn es das offene war — das von jetzt**, nicht das vom Beginn
  // des Löschens (Nachlauf über die Folgen vor 0.8.0, Befund 1). Wer ein
  // anderes Set löscht oder inzwischen aufschlug, behält seine Arbeit.
  const offenesJetzt = useAktuellesSet(setID ?? '')
  const setEntfernt = useCallback(
    (entfernt: string) => {
      if (offenesJetzt.nochAktuell(entfernt)) {
        setSchliessen()
      }
    },
    [offenesJetzt, setSchliessen],
  )

  const { drucken, hilfe, seite, ueber } = useMenueziele(
    setzeAktuell,
    setzeZuUeber,
  )

  // **Die Seitenleiste** (#1490): Knopf, ⌥⌘S und Menüeintrag schalten
  // dieselbe Einstellung.
  const seitenleiste = useSeitenleiste(einstellungen)

  // **Was das Fenster an Befehlen anmeldet** — in eigener Datei (§2), seit
  // die Einführung dazukam. Der Schnitt folgt der Frage: hier steht, **was**
  // die Anwendung zeichnet, dort, **welche Befehle** sie entgegennimmt — und
  // dass jeder von ihnen die Einführung schließt (#1325).
  const fensterhandlungen = useFensterhandlungen(
    {
      einstellungenOeffnen,
      sichernJetzt,
      sichernUnter,
      bestandSichern: bestand.ausMenue,
      setOeffnen,
      setNeu,
      waehleSet,
      setSchliessen,
      drucken,
      hilfe,
      rueckgaengig: menueverlauf.zuruecknehmen,
      wiederherstellen: menueverlauf.wiederherstellen,
      seite,
      ueber,
      seitenleiste: seitenleiste.umschalten,
      seiteZurueck: gang.zurueck,
      seiteVor: gang.vor,
      dateiGeoeffnet: datei.dateiGeoeffnet,
      dateiFehler: datei.dateiFehler,
    },
    einfuehrung.schliessen,
  )

  useFensterereignisse(fensterhandlungen)

  // **Eine gezogene Datei** (#891). Der Horcher steht hier und nicht auf der
  // Kartenseite: Wails meldet ein Ablegen **einmal fürs ganze Fenster**, und
  // ein Horcher eine Ebene tiefer bekam nichts mit, wenn keine Karte offen
  // war — und machte aus einer gezogenen `.mse-set` ein Bild.
  useAbgelegteDateien(datei.dateiGezogen)

  // **Das offene Set liest das Fenster einmal** (U13, #2158), mit dem Stand
  // der Seiten. **Die Meldung des Fensters** (#180) steht im Band.
  const stand = verlauf.stand + fremderstand + eigenerstand
  const { offenesSet, objekt } = useFensterset(setID, stand)
  const band = useMeldungsband(
    {
      sprache,
      seite: aktuell,
      setname: objekt?.name,
      rueckspielrest,
      dateimeldung: datei.meldung,
      verlaufsfehler: verlauf.fehler,
      bestandsfehler: bestand.fehler,
      bestandspfad: bestand.pfad,
      setID,
      speicherungFuer: speichern.fuer,
      speicherfehler: speichern.fehler,
      speicherpfad: speichern.speicherung?.pfad,
    },
    meldungRaeumen,
  )

  return (
    // **An der Stelle des Fragments** — warum hier, steht in `anschlussraum.tsx`.
    <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
      <Starteinfuehrung
        einstellungen={einstellungen}
        zugeklappt={einfuehrung.zugeklappt}
        erneut={einfuehrung.erneut}
        setID={setID}
        wechseln={setzeAktuell}
        schliessen={einfuehrung.schliessen}
        gewaehlt={bestandGeaendert}
      />

      <Rahmen
        aktuell={aktuell}
        wechseln={setzeAktuell}
        zuruecknehmen={verlauf.zuruecknehmen}
        wiederherstellen={verlauf.wiederherstellen}
        sichern={sichernJetzt}
        oeffnen={setOeffnen}
        neu={setNeu}
        schliessen={setSchliessen}
        drucken={drucken}
        sichernUnter={sichernUnter}
        seiteZurueck={gang.zurueck}
        seiteVor={gang.vor}
        meldung={band.meldung}
        meldungSchliessen={band.raeumen}
        seitenleiste={seitenleiste}
        objekt={objekt}
        palettenwege={{ setID, waehleSet, oeffneKarte }}
        ziffern={belegungAus(einstellungen.wert('kuerzel.ziffern'))}
      >
        <OffenesSet wert={offenesSet}>
          <Seiteninhalt
            seite={aktuell}
            wechseln={setzeAktuell}
            fassung={fassung}
            zuUeber={zuUeber}
            ueberAngekommen={() => {
              setzeZuUeber(false)
            }}
            einfuehrungZeigen={einfuehrung.zeigen}
            bestandGeaendert={bestandGeaendert}
            setID={setID}
            speichern={speichern}
            karteID={karteID}
            meldefehler={meldefehler}
            // **Beide Stände zusammen** — ⌘Z und MCP sind dieselbe Frage (#156).
            verlaufsstand={stand}
            verlaufsfehler={verlauf.fehler}
            zuruecknehmen={verlauf.zuruecknehmen}
            oeffnungsanstoss={oeffnungsanstoss}
            oeffnungAngestossen={oeffnungAngestossen}
            neuanstoss={neuanstoss}
            neuAngestossen={neuAngestossen}
            bestandsstand={fremderstand + eigenerstand}
            einstellungen={einstellungen}
            bestandspflege={pflege}
            ruecksprung={ruecksprung}
            pflege={datenpflege}
            waehleSet={waehleSet}
            setEntfernt={setEntfernt}
            oeffneKarte={oeffneKarte}
            symbolziel={symbolziel.ziel}
            oeffneSymbol={oeffneSymbol}
          />
        </OffenesSet>
      </Rahmen>
    </Anschlussraum>
  )
}
