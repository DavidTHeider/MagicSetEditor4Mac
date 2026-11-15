// Die Anwendungseinstellungen —
// #245.
//
// # Warum aus dem Bestand und nicht aus dem Browserspeicher
//
// Bis #245 lag die Sprache im `localStorage`. Der Grund dort war gut — „ein
// Bestand, der ohne Datenbank nicht anzeigbar ist, wäre eine Anwendung, die
// bei einem Migrationsfehler auch noch stumm ist" — und wird **anders**
// eingelöst: lässt sich der Bestand nicht lesen, gelten die Vorgabewerte und
// die Systemsprache, also genau das, was ein frischer Rechner zeigt.
//
// Dafür gewinnt der Bestand drei Dinge: **einen Ort** (der MCP-Server hat
// keinen `localStorage`), **einen Sicherungsbegriff** (wer seine Datei kopiert,
// nimmt seine Einstellungen mit) und **eine Prüfung**
// (`domain.PruefeEinstellung`).
//
// # Warum nicht [useLaden]
//
// Weil dieser Haken **über** dem `Sprachraum` läuft: `App` braucht die Sprache,
// um ihn herzustellen. `useLaden` ruft `useTexte()` und bekäme dort den
// Vorgabewert — Deutsch —, und ein Ladefehler stünde in der englischen
// Oberfläche auf Deutsch. Genau diese Falle steht in `App.tsx` beschrieben und
// war dort mit dem Satz geschlossen: „über dieser Zeile gibt es keinen Haken
// mehr, der einen Katalog braucht."
//
// Deshalb reicht dieser Haken die **Ursache** durch, nicht den Satz. Formuliert
// wird sie auf der Einstellungsseite — innerhalb des Sprachraums.
import { useCallback, useEffect, useRef, useState } from 'react'

import { api, auspacken, type EinstellungDTO } from '../api'
import { useEinstellungsraum } from './einstellungsraum'
import { useLauf } from './useLauf'
import { useVorgang } from './useVorgang'
import { vergissVorrangreihe } from './useVorrangreihe'

// Einstellungsstand ist, was die Oberfläche über die Einstellungen weiß.
export interface Einstellungsstand {
  // wert gibt, was **gilt** — leer, solange nichts geladen ist.
  //
  // **Und nicht, was gewählt ist**, seit Werkstück 9 von Phase 4.12
  // (#1351,
  // Punkt 9): steht bei einem Profil „wie der Nutzer", gilt der Wert des
  // Menschen. Jeder Aufrufer hier — Sprache, Erscheinungsbild, Schriftgröße,
  // Anzeigeschrift, Einführung — will wissen, was gilt.
  //
  // **Die Wahl selbst steht in `eintraege`**, und dort holt sie die
  // Einstellungsseite: die Auswahlliste zeigt „wie der Nutzer", solange
  // niemand etwas anderes gewählt hat.
  wert: (schluessel: string) => string

  // eintraege sind die Einstellungen in Anzeigereihenfolge, samt ihrer Werte.
  eintraege: EinstellungDTO[]

  // imBereich gibt die Einstellungen eines Bereichs — `anwendung` für die
  // Einstellungsseite, `bildausfuhr` für die Bildseite (#491).
  imBereich: (bereich: string) => EinstellungDTO[]

  // vorgaben sind die, mit denen ein neuer Datensatz anfängt — Werkstück 9
  // von Phase 4.12
  // (#1351).
  vorgaben: () => EinstellungDTO[]

  // dienstschalter sind die vier „darf gefragt werden" — Phase 4.8.
  dienstschalter: () => EinstellungDTO[]

  // fragewahlen sind die „wer beantwortet was".
  //
  // **Zwei Wege statt eines Bereichs**: beide stehen unter `dienste`, und sie
  // sind zweierlei ([0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md),
  // Entscheidung 1).
  fragewahlen: () => EinstellungDTO[]

  // kizeilen sind die zwei der lokalen KI-Laufzeit — A7, #1446.
  //
  // **Ein dritter Weg im selben Bereich**, aus demselben Grund wie die zwei
  // darüber: „darf gefragt werden" und „wo lauscht sie" sind nicht „wer
  // beantwortet was", und eine gemeinsame Liste zeigte drei Dinge als eines.
  kizeilen: () => EinstellungDTO[]

  // geladen sagt, ob die Antwort **schon einmal** da war.
  //
  // **Nicht „lädt gerade nicht"**, und der Unterschied ist teuer bezahlt: `App`
  // zeichnet nichts, solange die Sprache nicht feststeht — und jede Änderung
  // lädt erneut. Beim ersten Anlauf verschwand damit bei **jedem** Umschalten
  // das ganze Fenster für einen Wimpernschlag und baute sich neu auf. Gefunden
  // hat es der Test „schreibt eine Änderung und liest neu": nach dem
  // Umschalten war die Auswahlliste nicht mehr da.
  //
  // Ein **Fehler** zählt als geladen: dann gelten die Vorgabewerte, und warten
  // hilft nicht mehr.
  geladen: boolean

  // ursache ist der rohe Fehler beim **Laden** — der Satz dazu entsteht im
  // Sprachraum.
  ursache: unknown

  // fehlschlag ist ein gescheitertes **Schreiben**, mit der Einstellung, an
  // der es geschah — damit die Meldung dort steht und nicht als Ladefehler
  // über der Seite (§6; gefunden am nativen Fenster für OPS-02, #1719). Ein
  // geglücktes Schreiben räumt ihn.
  fehlschlag: { schluessel: string; ursache: unknown } | undefined

  setze: (schluessel: string, wert: string) => void
  laeuft: (schluessel: string) => boolean
}

// Ergebnis ist, was ein Lauf mitbringt — Einträge oder eine Ursache.
interface Ergebnis {
  eintraege?: EinstellungDTO[]
  ursache?: unknown
}

// **`erben` ist eine Wahl und kein Wert, der gilt** — `domain.Erben`: es
// gilt dann, was beim Menschen gewählt ist, und das weiß nur der Bestand.
const erben = 'erben'

// nachgetragen setzt einen angenommenen Wert in seinen Eintrag — **soweit
// das Fenster ihn kennt** (Nachlauf über Phase 6.12, N6).
//
// `wirksam` ist eine Faltung in `internal/domain`, und das Fenster rechnet
// sie nicht nach. Nachgetragen wird deshalb nur, was in der Auswahl steht:
// eine freie Eingabe zeigt der Bestand anders, als sie geschrieben wurde —
// ein Proxy ohne Kennwort. Und bei `erben` ändert sich die Wahl, nicht das
// Geltende.
//
// **Eine Näherung, und sie steht nur im Fehlerfall da**: eine Richtlinie,
// bei der die strengere Wahl des Menschen gilt, oder „nur lokal" über den
// Diensten sieht sie nicht. Das nächste geglückte Lesen ersetzt sie, und
// der Fehler steht so lange daneben.
function nachgetragen(
  eintrag: EinstellungDTO,
  wert: string | undefined,
): EinstellungDTO {
  if (wert === undefined || !eintrag.werte.includes(wert)) {
    return eintrag
  }

  return wert === erben
    ? { ...eintrag, wert }
    : { ...eintrag, wert, wirksam: wert }
}

// **`stand` sagt, wann neu zu lesen ist**
// (Zweitblick über Phase S). Die Einstellungen wurden **einmal** geladen; nach
// einem Zurückspielen stand der Wert des ersetzten Bestands da. Der Zähler
// kommt von oben, weil nur die Anwendung weiß, wann sich alles geändert hat.
// Fragehaltung sagt, was der Haken mit dem Bestand vorhat.
export type Fragehaltung = 'fragen' | 'ohne'

// **Die Haltung sagt, ob überhaupt gefragt werden darf.**
//
// # Warum diese Wahl sein muss
//
// Ging der Bestand nicht auf, sind die Nähte der Anwendung **nicht gesetzt**.
// Ein Aufruf darüber läuft in einen Nullzeiger; Wails fängt ihn ab und ruft
// den Rückweg **nie** — das Fenster wartet für immer, und die
// Wiederherstellungsseite käme nie zum Vorschein.
//
// Gemessen am 2026-08-25 an einer laufenden Anwendung mit einer unbrauchbaren
// Bestandsdatei: `process message error: app.Bindungen.Einstellungen ->
// runtime error: invalid memory address or nil pointer dereference`, und das
// Fenster blieb weiß.
//
// **Ohne Frage gelten die Vorgabewerte** — dieselben, die ein Fehler beim
// Laden ergäbe (#962).
//
// **Der Bezug sagt, wessen Einstellungen es sind** — Bestand und Profil (U08,
// #2124). Scheitert ein Nachlesen bei **demselben** Bezug, bleibt der
// bestätigte Stand stehen; bei einem anderen wird geräumt.
export function useEinstellungen(
  stand = 0,
  haltung: Fragehaltung = 'fragen',
  bezug = '',
): Einstellungsstand {
  // **Unter einem Einstellungsraum gilt der geteilte Stand** — Befund G14 des
  // Zweitblicks über Phase 6.3 (`einstellungsraum.tsx`). Der eigene wird
  // trotzdem gebaut, weil ein Haken nicht bedingt gerufen werden darf; er
  // fragt dann nicht.
  const geteilt = useEinstellungsraum()
  const eigener = useEigenerStand(
    stand,
    geteilt === undefined ? haltung : 'ohne',
    bezug,
  )

  return geteilt ?? eigener
}

function useEigenerStand(
  stand: number,
  haltung: Fragehaltung,
  bezug: string,
): Einstellungsstand {
  const vorgang = useVorgang()
  const lauf = useLauf()
  const [eintraege, setzeEintraege] = useState<EinstellungDTO[]>([])
  const [geladen, setzeGeladen] = useState(false)
  const [ursache, setzeUrsache] = useState<unknown>()
  const [fehlschlag, setzeFehlschlag] =
    useState<Einstellungsstand['fehlschlag']>()

  // **Was seit dem letzten geglückten Lesen angenommen wurde** (Nachlauf
  // über Phase 6.12, N6). Zwei Schreibvorgänge auf verschiedene Schlüssel
  // teilen sich einen Ladezähler: das ältere Nachladen wird verworfen, und
  // trug nur das jüngere seinen Wert nach, ging der ältere verloren.
  const angenommen = useRef(new Map<string, string>())

  // bestaetigt ist der Bezug, zu dem die Einträge zuletzt **gelesen** wurden
  // — `undefined`, solange nie (U08, #2124).
  const bestaetigt = useRef<string | undefined>(undefined)

  // bezugJetzt ist der Bezug, der **jetzt** gilt.
  const bezugJetzt = useRef(bezug)
  useEffect(() => {
    bezugJetzt.current = bezug
  }, [bezug])

  // Ergebnis ist, was ein Lauf mitbringt: Einträge **oder** eine Ursache.
  //
  // **Getrennt vom Eintragen**, damit der Effekt darüber entscheiden kann, ob
  // er es noch will (§7, Regel 2).
  const holen = useCallback(async (): Promise<Ergebnis> => {
    try {
      const antwort = await api().Einstellungen()

      return { eintraege: auspacken(antwort, () => antwort.einstellungen) }
    } catch (fehler: unknown) {
      return { ursache: fehler }
    }
  }, [])

  // eintragen übernimmt ein Ergebnis — **wenn es noch das aktuelle ist**.
  //
  // **§7, Regel 2: eine Antwort gehört zu dem Stand, aus dem sie stammt.** Der
  // Fall ist erreichbar und nicht theoretisch: wer eine Einstellung umschaltet,
  // während der erste Lauf noch unterwegs ist, bekäme sonst den **alten** Stand
  // eingetragen, nachdem der neue schon dastand — und die Auswahl spränge
  // zurück, ohne dass etwas fehlgeschlagen wäre.
  //
  // **Ein Lesefehler nach einem geglückten Schreiben nimmt keinen
  // bestätigten Stand zurück** (Zweitblick über Phase 6.12, Z9). Hier stand
  // nur `?? []`, und ein Nachladen, das danach scheiterte, ließ gar keine
  // Einstellungen übrig — ein eben abgeschalteter Bereich erschien wieder,
  // obwohl er im Bestand aus war. Jetzt bleibt der letzte Stand, und
  // **alles, was seit dem letzten Lesen angenommen wurde**, gilt darin —
  // nicht nur das jüngste (N6). Der Fehler steht trotzdem da.
  const eintragen = useCallback(
    (
      meiner: number,
      ergebnis: Ergebnis,
      nachSchreiben = false,
      lesebezug = bezug,
    ) => {
      if (!lauf.istAktuell(meiner)) {
        return
      }

      if (ergebnis.eintraege) {
        angenommen.current.clear()
        bestaetigt.current = lesebezug
        setzeEintraege(ergebnis.eintraege)
      } else if (
        nachSchreiben ||
        (ergebnis.ursache !== undefined && bestaetigt.current === lesebezug)
      ) {
        // **Derselbe Bezug: der bestätigte Stand bleibt** (U08, #2124) —
        // ein Lesefehler setzte sonst Sprache, Erscheinungsbild und
        // Bereiche still auf die Vorgaben zurück. Der Fehler steht trotzdem
        // da. Was seit dem Lesen geschrieben wurde, gilt darin (N6). **Nur
        // bei einem Fehler**: ein Bestand, der gar nicht aufging (`ohne`),
        // bringt keine Ursache, und dann gilt das System.
        const geschrieben = new Map(angenommen.current)

        setzeEintraege((bisher) =>
          bisher.map((eintrag) =>
            nachgetragen(eintrag, geschrieben.get(eintrag.schluessel)),
          ),
        )
      } else {
        // **Ein anderer Lesefehler räumt weiter ab** — wer den Bestand nicht
        // lesen kann, soll nicht mit einem Stand von vorhin weiterarbeiten;
        // die Sprache fällt dann auf die des Systems zurück
        // (`wiederherstellung.test.tsx`).
        setzeEintraege([])
      }

      setzeUrsache(ergebnis.ursache)

      // **Auch bei einem Fehler, und das ist der Kern**: er zählt als geladen.
      // Eine Anwendung, die auf eine Antwort wartet, die nicht kommt, ist
      // stumm — und stumm war genau das, was der Browserspeicher verhindern
      // sollte.
      setzeGeladen(true)
    },
    [lauf, bezug],
  )

  // **Der Lauf steht in einem Effekt, sein Ergebnis in `then`.**
  //
  // Ein `await` unmittelbar im Effektkörper riefe `setState` **synchron** im
  // ersten Lauf, sobald die Fassade ein schon erfülltes Versprechen gibt — und
  // genau das tut jede Vorrichtung. React nennt das kaskadierende Zeichnungen,
  // und der Linter weist es ab.
  useEffect(() => {
    const meiner = lauf.naechster()

    // **Ein Weg und nicht zwei.** Ein `eintragen` unmittelbar im Effektkörper
    // wäre ein Zustandswechsel im Effekt selbst — React nennt das
    // kaskadierende Zeichnungen, und der Linter weist es ab. Ein erfülltes
    // Versprechen trägt dasselbe Ergebnis eine Mikroaufgabe später ein.
    const ergebnis: Promise<Ergebnis> =
      haltung === 'ohne' ? Promise.resolve({}) : holen()

    void ergebnis.then((geholt) => {
      eintragen(meiner, geholt, false, bezug)
    })
  }, [haltung, holen, eintragen, lauf, stand, bezug])

  const setze = (schluessel: string, wert: string) => {
    // **Ein Abschluss nach einem Profilwechsel gehört nicht mehr hierher**
    // (§7, Regel 1 und 2; Zweitblick über C und D, Befund 4).
    const meinBezug = bezug

    // **Ein Vorgang je Einstellung** (§7, Regel 4). Zwei Klicks auf dieselbe
    // Auswahl gingen sonst gegen einen Stand, den das Backend noch nicht kennt.
    void vorgang.starte(schluessel, async () => {
      try {
        auspacken(
          await api().EinstellungSetzen(schluessel, wert),
          () => undefined,
        )
      } catch (fehler: unknown) {
        if (bezugJetzt.current !== meinBezug) {
          return
        }

        // **Gefangen und gezeigt, nicht durchgereicht.** Der erste Anlauf ließ
        // ihn laufen — vitest meldete eine unbehandelte Ablehnung, und die
        // Seite sagte gar nichts: die Auswahl sprang zurück, und niemand
        // erfuhr, warum. §8 verlangt das Gegenteil.
        setzeFehlschlag({ schluessel, ursache: fehler })

        // **Nach einem gescheiterten Schreiben wird nicht nachgeladen**
        // (§7, Regel 3): sonst überschriebe ein Lesefehler die Meldung, die
        // sagt, was zu tun ist.
        return
      }

      if (bezugJetzt.current !== meinBezug) {
        return
      }

      setzeFehlschlag(undefined)

      // **Eine KI-Einstellung ändert, wen der Vorrang fragt** (#2054) — die
      // gemerkte Reihe am Formular gilt dann nicht mehr.
      if (schluessel.startsWith('ki.')) {
        vergissVorrangreihe()
      }

      // **Neu geladen und nicht örtlich gesetzt.** Was gilt, rechnet der
      // Bestand (`nachgetragen`); gemerkt wird der Wert für den Fall, dass
      // das Lesen scheitert.
      angenommen.current.set(schluessel, wert)

      const meiner = lauf.naechster()

      eintragen(meiner, await holen(), true)
    })
  }

  return {
    wert: (schluessel) =>
      eintraege.find((eintrag) => eintrag.schluessel === schluessel)?.wirksam ??
      '',
    eintraege,
    imBereich: (bereich) =>
      eintraege.filter((eintrag) => eintrag.bereich === bereich),

    // **Die Vorgaben am Bereich und nicht am Präfix** — der Bereich kommt aus
    // `internal/domain`, und wer eine Vorgabe hinzufügt, trägt ihn dort ein.
    vorgaben: () =>
      eintraege.filter((eintrag) => eintrag.bereich === 'vorgaben'),

    // **Schalter und Wahlen stehen im selben Bereich und sind zweierlei**
    // (Phase 4.8, [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md))
    // — getrennt am Präfix, denn die Schlüssel kommen aus `internal/domain`.
    dienstschalter: () =>
      eintraege.filter((eintrag) => eintrag.schluessel.startsWith('dienst.')),
    fragewahlen: () =>
      eintraege.filter((eintrag) => eintrag.schluessel.startsWith('frage.')),
    kizeilen: () =>
      eintraege.filter((eintrag) => eintrag.schluessel.startsWith('ki.')),
    geladen,
    ursache,
    fehlschlag,
    setze,
    laeuft: vorgang.laeuft,
  }
}
