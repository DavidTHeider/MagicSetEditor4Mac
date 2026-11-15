// **Der Inhalt einer Setlasche** — der Schalter über die Laschen eines
// Sets.
//
// **Eine eigene Datei**, seit `Seiteninhalt.tsx` mit den Schubladen über die
// Zeilengrenze aus §2 wuchs (UX-08, #1680). Der Schnitt folgt der Frage: dort
// steht, **wo** eine Seite erscheint — Anwendung, Arbeitsplatz, Schublade —,
// hier, **welche** Seite zu welchem Schlüssel gehört.
import type { ReactNode } from 'react'

import { OhneKarten } from '../bausteine/OhneKarten'
import { useGelesenesSet } from '../haken/offenesset'
import type { Seitenschluessel, Setschluessel } from '../navigation'
import { Ausgeben } from './Ausgeben'
import { Ersetzen } from './Ersetzen'
import { Karten } from './Karten'
import { Schluesselwoerter } from './Schluesselwoerter'
import { Setangaben } from './Setangaben'
import { Skriptkonsole } from './Skriptkonsole'
import { Statistik } from './Statistik'
import { Stil } from './Stil'
import { Suche } from './Suche'
import { Symbolfach } from './Symbolfach'
import { Zufallspack } from './Zufallspack'

// Lascheninhalt ist der Schalter über die Laschen eines Sets.
//
// **`setID` ist hier eine Zeichenkette** — der Arbeitsplatz ruft ihn nur mit
// offenem Set. Damit ist der Wächter „kein Set gewählt", der in fünf Laschen
// stand, keine Frage mehr, sondern eine Zusage des Typs (§3).
export function Lascheninhalt({
  seite,
  wechseln,
  setID,
  karteID,
  oeffneKarte,
  symbolziel,
  oeffneSymbol,
  bestandGeaendert,
  verlaufsstand,
  verlaufsfehler,
  zuruecknehmen,
}: {
  seite: Setschluessel
  wechseln: (seite: Seitenschluessel) => void
  setID: string
  karteID: string | undefined
  oeffneKarte: (setID: string, karteID: string) => void

  // symbolziel sagt, welches Symbol der Editor bearbeitet — ohne Ziel das
  // Setsymbol (#1339).
  symbolziel?: { karteID: string; feld: string } | undefined

  // oeffneSymbol führt zum Symboleditor für ein Kartenfeld.
  oeffneSymbol: (karteID: string, feld: string) => void
  bestandGeaendert: () => void
  verlaufsstand: number
  verlaufsfehler: string | undefined
  zuruecknehmen?: ((vorgang: string) => void) | undefined
}) {
  // **Was mit den Karten arbeitet, sagt über sich, dass keine da sind**
  // (G02, #2249) — `OhneKarten`. Die Zahl kommt aus dem Raum des offenen
  // Sets, den das Fenster ohnehin liest; ein eigener Ruf je Lasche wäre ein
  // zweiter Stand derselben Frage (U13).
  const karten = useGelesenesSet(setID)?.karten
  const mitHinweis = (inhalt: ReactNode) => (
    <>
      <OhneKarten karten={karten} wechseln={wechseln} />
      {inhalt}
    </>
  )

  switch (seite) {
    case 'karten':
      return (
        <Karten
          setID={setID}
          karteID={karteID}
          karteGewaehlt={(gewaehlte) => {
            oeffneKarte(setID, gewaehlte)
          }}
          symbolZeichnen={oeffneSymbol}
          verlaufsstand={verlaufsstand}
          verlaufsfehler={verlaufsfehler}
          bestandGeaendert={bestandGeaendert}
          zuruecknehmen={zuruecknehmen}
        />
      )

    case 'setangaben':
      // **Der Verlaufsstand steht nicht im Schlüssel** (U03, #2120): er baute
      // die Felder neu auf, und wer gerade tippte, verlor Entwurf und Fokus.
      // Eine Rücknahme erreicht das Feld über `Setangabezeile`.
      //
      // **Die Kennung des Sets steht darin** —
      // #749:
      // §7 Regel 1 verlangt, dass ein Wechsel des Bezugsobjekts den offenen
      // Kontext räumt. Ohne sie blieb eine Fehlermeldung des einen Sets unter
      // dem nächsten stehen — und die Meldungen dieser Seiten tragen keinen
      // Bezug, an dem man das sähe.
      return (
        <Setangaben key={setID} setID={setID} verlaufsstand={verlaufsstand} />
      )

    case 'schluesselwoerter':
      // **Der Schlüssel ist das Set** — ein Wechsel räumt Wahl, Editor und
      // Meldung mit ab (§7, Regel 1), und damit fällt auch Regel 2 mit: die
      // Antwort eines überholten Schreibens findet niemanden mehr.
      //
      // **Der Verlaufsstand steht nicht darin** (UXR-F3, #2009): er lädt die
      // Liste neu und baut ihren Editor neu auf, aber das gewählte Wort bleibt
      // gewählt — wer eine Rücknahme ausführt, will weitermachen, wo er war.
      // Die Seite bis dahin, deren Felder auf `defaultValue` stehen, räumte
      // mit dem Verlaufsstand im Schlüssel auch die Wahl ab.
      return (
        <Schluesselwoerter
          key={setID}
          setID={setID}
          verlaufsstand={verlaufsstand}
        />
      )

    case 'konsole':
      // **Die aufgeschlagene Karte ist der Kontext.** Ohne eine gibt es kein
      // `card`, und die Seite sagt es, statt einen leeren Ausdruck anzubieten.
      //
      // **Und sie steht deshalb im Schlüssel**, nicht nur das Set: ein
      // Kartenwechsel ließ sonst Ausdruck, Ergebnis und Fehler stehen (§7,
      // Regel 1), und eine verspätete Antwort für die vorige Karte wurde
      // unter der neuen eingetragen (Regel 2). Ein `key` löst beide zugleich
      // — `docs/SCHREIBWEGE.md`. Angestrichen vom Zweitblick durch Codex über
      // Phase W.
      return (
        <Skriptkonsole
          key={`${setID}/${karteID ?? ''}`}
          karteID={karteID ?? ''}
          zuDenKarten={wechseln}
        />
      )

    case 'pack':
      return mitHinweis(<Zufallspack key={setID} setID={setID} />)

    case 'symbol':
      // **Der Schlüssel trägt das Ziel.** Die Zeichenfläche steht auf
      // geladenem Zustand; ohne Neuaufbau zeigte sie beim Wechsel von einem
      // Feld zum nächsten weiter die alte Zeichnung (§7, Regel 1).
      return (
        <Symbolfach
          key={`${setID}/${symbolziel?.karteID ?? ''}/${symbolziel?.feld ?? ''}`}
          setID={setID}
          ziel={symbolziel}
        />
      )

    case 'stil':
      return mitHinweis(
        <Stil
          key={setID}
          setID={setID}
          geaendert={bestandGeaendert}
          wechseln={wechseln}
        />,
      )

    case 'suche':
      // **Im offenen Set beginnt die Suche dort** (AP-11, #2368) — und ein
      // anderes Set beginnt neu (§7, Regel 1).
      return mitHinweis(
        <Suche key={setID} oeffnen={oeffneKarte} offenesSet={setID} />,
      )

    case 'ersetzen':
      return mitHinweis(
        <Ersetzen key={setID} setID={setID} geaendert={bestandGeaendert} />,
      )

    case 'statistik':
      return mitHinweis(<Statistik key={setID} setID={setID} />)

    // **Drei Seiten, ein Assistent** (UXR-09, #1917): `bilder` und
    // `druckbogen` schlagen ihn mit ihrem Ziel auf (`ausgabe.ts`).
    case 'bilder':
    case 'druckbogen':
    case 'ausgeben':
      return <Ausgeben setID={setID} seite={seite} wechseln={wechseln} />
  }
}
