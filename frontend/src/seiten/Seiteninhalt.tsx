// Der Inhalt je Seite.
//
// **Eine Stelle, die jede Seite kennt** — und der Schalter darüber ist
// vollständig, weil `Seitenschluessel` ein Aufzählungstyp ist: eine neue Seite
// in `navigation.ts` macht diese Datei zum Typfehler, nicht zu einer leeren
// Fläche im Betrieb.
import { Activity } from 'react'

import { istSetseite, type Seitenschluessel } from '../navigation'
import type { Bestandspflege } from '../haken/useBestandspflege'
import type { Pflegelauf } from '../haken/useDatenpflege'
import type { Ruecksprungstand } from '../haken/useRuecksprung'
import type { Einstellungsstand } from '../haken/useEinstellungen'
import { Abgeschaltet } from './Abgeschaltet'
import { useVerborgeneSeiten } from '../abschaltung'
import { Anwendungsinhalt } from './Anwendungsinhalt'
import { Setbereich } from './Setbereich'
import { useAufgehobeneSeiten } from '../haken/useAufgehobeneSeiten'
import { useHerkunft } from '../haken/useHerkunft'
import { type Speicherlauf } from '../haken/useSetSpeichern'

// **Alle drei Angaben sind Pflicht.**
//
// Der erste Anlauf machte `waehleSet` freiwillig und setzte einen Rückfall, der
// nichts tut — ein Zweig, den nur ein Test betritt, und eine Setliste, deren
// Klicks im Betrieb ins Leere gingen, falls jemand die Angabe vergisst. Der
// Typ sagt es jetzt, statt es aufzufangen.
interface SeiteninhaltEigenschaften {
  seite: Seitenschluessel

  // **Für den Abschnitt „Über" der Hilfe** (#1168, Z4-01).
  //
  // Sie stehen hier und nicht im Zustand der Hilfeseite, weil der
  // **Menüeintrag** sie auslöst — und der spricht mit dem Fenster, nicht mit
  // einer Seite, die vielleicht gar nicht gezeichnet ist.
  fassung: string
  zuUeber: boolean
  ueberAngekommen: () => void
  einfuehrungZeigen: () => void
  setID: string | undefined

  // speichern ist **einer** für das ganze Fenster (#162, Zweitblick über
  // Phase 6): ⌘S und die Knöpfe der Setliste teilen sich seine Vorgangsliste,
  // sonst gälte §7 Regel 4 je Aufrufer statt je Set.
  speichern: Speicherlauf

  // karteID ist die **Vorgabe** für die Kartenseite — die Karte, die ein
  // Treffer aufgeschlagen hat. Nicht die gerade gewählte: welche das ist, führt
  // die Kartenseite selbst.
  karteID: string | undefined

  // bestandGeaendert meldet, dass dieses Fenster den Bestand geändert hat —
  // ein Stilwechsel etwa. Die Laschen legen den Stand in ihre
  // Ladeabhängigkeiten und laden beim Zurückwechseln neu.
  bestandGeaendert: () => void

  // wechseln führt auf eine andere Seite — die Startseite bietet die ersten
  // Schritte an, und ein Schritt ohne Ziel wäre eine Beschreibung.
  wechseln: (seite: Seitenschluessel) => void

  waehleSet: (setID: string, karteID?: string) => void
  oeffneKarte: (setID: string, karteID: string) => void

  // symbolziel sagt, welches Symbol der Editor bearbeitet — ohne Ziel das
  // Setsymbol (#1339).
  symbolziel?: { karteID: string; feld: string } | undefined

  // oeffneSymbol führt zum Symboleditor für ein Kartenfeld.
  oeffneSymbol: (karteID: string, feld: string) => void

  // einstellungen sind die Anwendungseinstellungen (#245) — sie werden auf der
  // Einstellungsseite gezeigt und dort geändert.
  //
  // **Die Sprache steht nicht mehr daneben.** Bis #245 reichte das Fenster sie
  // bis hierher durch, damit die Einstellungsseite sie in ihrer Auswahlliste
  // zeigen konnte; jetzt steht sie mit den anderen Einstellungen darin.
  einstellungen: Einstellungsstand

  // bestandspflege ist das Sichern und Zurückspielen des Bestands (#940).
  //
  // **Von oben und nicht hier gebaut**: das Sichern hängt auch an der
  // Menüleiste, und beide Wege müssen sich denselben Vorgang teilen (§7,
  // Regel 4).
  bestandspflege: Bestandspflege

  // ruecksprung sind die vier Rückkehrarten und die verbuchten Rückkehren
  // (O7, #1431).
  ruecksprung: Ruecksprungstand

  // pflege ist das Aufräumen verwaister Zeilen (O8, #1432).
  pflege: Pflegelauf

  meldefehler: string | undefined

  // verlaufsstand steigt nach jeder geglückten Rücknahme (#52). Die Kartenseite
  // legt ihn in ihre Ladeabhängigkeiten: **was sich geändert hat, steht im
  // Bestand**, und gelesen wird es von dort.
  verlaufsstand: number

  // zuruecknehmen nimmt **einen genannten** Schritt zurück — aus der
  // Verlaufsliste der Kartenseite (U6, #1417). Wahlfrei: wer keinen Verlauf
  // zeigt, reicht nichts durch.
  zuruecknehmen?: ((vorgang: string) => void) | undefined

  // verlaufsfehler erscheint auf der Kartenseite — dort, wo ⌘Z wirkt, und nicht
  // als anonymer Hinweis über dem Fenster (§6).
  verlaufsfehler: string | undefined

  // oeffnungsanstoss stößt den Dateidialog der Setliste an (#891) — ⌘O und der
  // Ablageeintrag „Öffnen …" setzen ihn, die Setliste verbraucht ihn.
  oeffnungsanstoss: boolean
  oeffnungAngestossen: () => void

  // neuanstoss setzt den Fokus ins Namensfeld des neuen Sets — ⌘N (#891).
  neuanstoss: boolean
  neuAngestossen: () => void

  // bestandsstand lässt „Zuletzt benutzt" neu laden (#891).
  bestandsstand: number

  // setEntfernt sagt, dass ein Set gelöscht wurde (#1218). War es das offene,
  // gibt es kein offenes mehr (§7, Regel 1).
  setEntfernt: (setID: string) => void
}

export function Seiteninhalt({
  seite,
  fassung,
  zuUeber,
  ueberAngekommen,
  einfuehrungZeigen,
  wechseln,
  bestandGeaendert,
  setID,
  speichern,
  karteID,
  waehleSet,
  oeffneKarte,
  symbolziel,
  oeffneSymbol,
  einstellungen,
  bestandspflege,
  ruecksprung,
  pflege,
  meldefehler,
  verlaufsstand,
  verlaufsfehler,
  zuruecknehmen,
  oeffnungsanstoss,
  oeffnungAngestossen,
  neuanstoss,
  neuAngestossen,
  bestandsstand,
  setEntfernt,
}: SeiteninhaltEigenschaften) {
  // **Welche Seiten stehen bleiben** (UX-03.AK-02, #1675) — jede besuchte,
  // verborgen, bis man zurückkommt. Warum aufheben und nicht zwischenspeichern,
  // steht am Haken. **Ein Zurückspielen räumt sie** (#1666, M9 a).
  const aufgehoben = useAufgehobeneSeiten(seite, setID, bestandspflege.ersetzt)

  // **Von wo man in die Hilfe kam** (UXR-22, #1930) — hier, weil nur diese
  // Stelle die Folge der Seiten sieht; die Hilfe bleibt aufgehoben stehen.
  const herkunft = useHerkunft(seite)

  // **Eine abgeschaltete Seite steht nicht da** (OPS-10, #1908) — an ihrer
  // Stelle der Satz, warum, und der Weg zurück.
  const abgeschaltet = useVerborgeneSeiten().has(seite)

  return (
    <>
      {/* **Eine Setseite steht im Arbeitsplatz, eine Anwendungsseite nicht.**
          Der Arbeitsplatz trägt den Namen des Sets und die Laschen; ohne ihn
          stünde „Statistik" da, ohne zu sagen, worüber sie spricht (#288).
          **Sein Schlüssel ist das Set**: ein anderes baut ihn neu (§7,
          Regel 1). */}
      {aufgehoben.setseite !== undefined && (
        <Activity mode={istSetseite(seite) ? 'visible' : 'hidden'}>
          <Setbereich
            key={setID ?? ''}
            seite={aufgehoben.setseite}
            aufgehoben={aufgehoben}
            setID={setID}
            wechseln={wechseln}
            karteID={karteID}
            oeffneKarte={oeffneKarte}
            symbolziel={symbolziel}
            oeffneSymbol={oeffneSymbol}
            bestandGeaendert={bestandGeaendert}
            verlaufsstand={verlaufsstand}
            verlaufsfehler={verlaufsfehler}
            zuruecknehmen={zuruecknehmen}
          />
        </Activity>
      )}

      {abgeschaltet && <Abgeschaltet seite={seite} />}

      {aufgehoben.anwendung.map((welche) => (
        <Activity
          key={welche}
          mode={welche === seite && !abgeschaltet ? 'visible' : 'hidden'}
        >
          <Anwendungsinhalt
            seite={welche}
            herkunft={herkunft}
            fassung={fassung}
            zuUeber={zuUeber}
            ueberAngekommen={ueberAngekommen}
            einfuehrungZeigen={einfuehrungZeigen}
            wechseln={wechseln}
            setID={setID}
            speichern={speichern}
            waehleSet={waehleSet}
            einstellungen={einstellungen}
            bestandspflege={bestandspflege}
            ruecksprung={ruecksprung}
            pflege={pflege}
            meldefehler={meldefehler}
            oeffnungsanstoss={oeffnungsanstoss}
            oeffnungAngestossen={oeffnungAngestossen}
            neuanstoss={neuanstoss}
            neuAngestossen={neuAngestossen}
            bestandsstand={bestandsstand}
            bestandGeaendert={bestandGeaendert}
            setEntfernt={setEntfernt}
          />
        </Activity>
      ))}
    </>
  )
}
