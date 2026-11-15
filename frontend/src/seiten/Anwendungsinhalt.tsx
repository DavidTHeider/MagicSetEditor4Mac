// **Der Schalter über die Seiten, die das Programm meinen** — Start, Setliste,
// Pakete, Einstellungen, Hilfe.
//
// **Eine eigene Datei**, weil `Seiteninhalt.tsx` mit beiden Schaltern über die
// Zeilengrenze wuchs (§2). Der Schnitt folgt der Frage: dort steht, **welcher
// von beiden** gemeint ist, hier, **welche Seite der Anwendung**.
import type { Bestandspflege } from '../haken/useBestandspflege'
import type { Pflegelauf } from '../haken/useDatenpflege'
import type { Ruecksprungstand } from '../haken/useRuecksprung'
import type { Einstellungsstand } from '../haken/useEinstellungen'
import { type Anwendungsschluessel, type Seitenschluessel } from '../navigation'
import { Einstellungen } from './Einstellungen'
import { Hilfe } from './Hilfe'
import { Nutzerseite } from './Nutzerseite'
import { Pakete } from './Pakete'
import { Decks } from './Decks'
import { Handeln } from './Handeln'
import { Spielen } from './Spielen'
import { Turniere } from './Turniere'
import { Sammlung } from './Sammlung'
import { type Speicherlauf } from '../haken/useSetSpeichern'
import { Sets } from './Sets'
import { Start } from './Start'
import { Vorlagen } from './Vorlagen'

// Anwendungsinhalt ist der Schalter über die Seiten, die das Programm meinen.
//
// **Beide Schalter sind vollständig**, weil beide Schlüssellisten
// Aufzählungstypen sind: eine neue Seite in `navigation.ts` macht diese Datei
// zum Typfehler, nicht zu einer leeren Fläche im Betrieb.
export function Anwendungsinhalt({
  seite,
  herkunft,
  fassung,
  zuUeber,
  ueberAngekommen,
  einfuehrungZeigen,
  wechseln,
  setID,
  speichern,
  waehleSet,
  einstellungen,
  bestandspflege,
  ruecksprung,
  pflege,
  meldefehler,
  oeffnungsanstoss,
  oeffnungAngestossen,
  neuanstoss,
  neuAngestossen,
  bestandsstand,
  bestandGeaendert,
  setEntfernt,
}: {
  seite: Anwendungsschluessel

  // herkunft ist die Seite vor der Hilfe (UXR-22) — die Hilfe zeigt sie
  // obenan.
  herkunft?: Seitenschluessel | undefined

  // **Für den Abschnitt „Über" der Hilfe** (#1168).
  fassung: string
  zuUeber: boolean
  ueberAngekommen: () => void
  einfuehrungZeigen: () => void

  wechseln: (seite: Seitenschluessel) => void
  setID: string | undefined
  speichern: Speicherlauf
  waehleSet: (setID: string, karteID?: string) => void

  // einstellungen sind die Anwendungseinstellungen (#245) — sie werden auf der
  // Einstellungsseite gezeigt und dort geändert.
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

  // oeffnungsanstoss stößt den Dateidialog der Setliste an (#891).
  oeffnungsanstoss: boolean
  oeffnungAngestossen: () => void

  // neuanstoss setzt den Fokus ins Namensfeld des neuen Sets — ⌘N (#891).
  neuanstoss: boolean
  neuAngestossen: () => void

  // bestandsstand lässt „Zuletzt benutzt" neu laden (#891).
  bestandsstand: number

  // bestandGeaendert meldet, dass dieses Fenster den Bestand geändert hat —
  // die Sammlung ruft es nach jedem eigenen Schreiben (R3, #2366).
  bestandGeaendert: () => void

  // setEntfernt sagt, dass ein Set gelöscht wurde (#1218).
  setEntfernt: (setID: string) => void
}) {
  switch (seite) {
    case 'start':
      return (
        <Start
          wechseln={wechseln}
          waehleSet={waehleSet}
          bestandsstand={bestandsstand}
        />
      )

    case 'sets':
      return (
        <Sets
          waehlen={waehleSet}
          speichern={speichern}
          anstoss={oeffnungsanstoss}
          anstossVerbraucht={oeffnungAngestossen}
          neuanstoss={neuanstoss}
          neuAngestossen={neuAngestossen}
          geloescht={setEntfernt}
          wechseln={wechseln}
          bestandsstand={bestandsstand}
        />
      )

    case 'pakete':
      return <Pakete wechseln={wechseln} bestandsstand={bestandsstand} />

    case 'vorlagen':
      return <Vorlagen wechseln={wechseln} bestandsstand={bestandsstand} />

    case 'sammlung':
      return (
        <Sammlung
          bestandsstand={bestandsstand}
          bestandGeaendert={bestandGeaendert}
        />
      )

    case 'decks':
      return <Decks wechseln={wechseln} bestandsstand={bestandsstand} />

    case 'turniere':
      return <Turniere bestandsstand={bestandsstand} />

    case 'spielen':
      return <Spielen wechseln={wechseln} bestandsstand={bestandsstand} />

    case 'handeln':
      return <Handeln bestandsstand={bestandsstand} />

    case 'hilfe':
      return (
        <Hilfe
          fassung={fassung}
          herkunft={herkunft}
          zuUeber={zuUeber}
          angekommen={ueberAngekommen}
          einfuehrungZeigen={einfuehrungZeigen}
        />
      )

    case 'nutzer':
      return <Nutzerseite bestandsstand={bestandsstand} />

    case 'einstellungen':
      return (
        <Einstellungen
          eintraege={einstellungen.imBereich('anwendung')}
          /* **Die Suche durchsucht alles, was dieses Fenster kennt** —
             UX-04.AK-03 und NET-02.NAT-03. Gezeigt wird weiterhin nur, was
             die sechs Listen darunter führen; `netz.proxy.http` zeichnet
             `Netzweg` selbst und stand in keiner von ihnen. */
          alleEinstellungen={einstellungen.eintraege}
          dienste={einstellungen.dienstschalter()}
          fragen={einstellungen.fragewahlen()}
          ki={einstellungen.kizeilen()}
          vorgaben={einstellungen.vorgaben()}
          anmeldung={einstellungen.imBereich('anmeldung')}
          ursache={einstellungen.ursache}
          meldefehler={meldefehler}
          setze={einstellungen.setze}
          laeuft={einstellungen.laeuft}
          bestandspflege={bestandspflege}
          ruecksprung={ruecksprung}
          pflege={pflege}
          setID={setID}
        />
      )
  }
}
