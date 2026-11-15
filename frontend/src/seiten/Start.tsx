// Die Startseite: was die Anwendung ist und was als Nächstes zu tun ist.
//
// # Warum sie mehr ist als eine Überschrift
//
// Wer die Anwendung öffnet, will wissen, was er tun kann, nicht was wir
// gerade bauen — ein Satz über den **Entwicklungsstand** wäre eine Notiz an
// uns.
//
// # Ein Schritt, der nirgendwohin führt, steht nicht da
//
// „Karten bearbeiten" erscheint erst, wenn ein Set offen ist. Ein Knopf, der
// nur eine Meldung erzeugt, ist eine Sackgasse mit Beschriftung — und §6
// verlangt vom leeren Zustand, dass er den nächsten Schritt **erklärt**, nicht
// dass er alle Schritte zeigt.
import { useCallback } from 'react'

import { api, type Portfolioantwort } from '../api'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { useLaden } from '../haken/useLaden'
import { Ueberblick, type Ueberblickszahlen } from './Ueberblick'
import { problemlage, type Problemlage } from './problemlage'
import { auspacken } from '../umschlag'
import { Aufmerksamkeit, useAufmerksamkeit } from './Aufmerksamkeit'
import { Hintergrundaufgaben } from './Hintergrundaufgaben'
import { Weiterarbeiten } from './Weiterarbeiten'
import type { Seitenschluessel } from '../navigation'
import { useTexte } from '../sprache/kontext'

interface StartEigenschaften {
  wechseln: (seite: Seitenschluessel) => void

  // waehleSet schlägt ein zuletzt benutztes Set auf (#891).
  waehleSet: (setID: string, karteID?: string) => void

  // bestandsstand lässt die Liste neu laden, wenn sich etwas geändert hat.
  bestandsstand: number
}

export function Start({
  wechseln,
  waehleSet,
  bestandsstand,
}: StartEigenschaften) {
  const texte = useTexte()
  const { start } = texte.seiten

  // **Die Zahlen des Überblicks** (#1166, Werkstück 3).
  //
  // **Ein Ladelauf für alle vier.** Drei nebeneinander gäben drei Zustände,
  // die einzeln eintreffen — und die Startseite spränge dreimal.
  //
  // **Ein Lesefehler hält nichts auf** (§8): was nicht zu holen war, zählt
  // null, und die übrigen Zahlen stehen da.
  //
  // # `allSettled` und nicht `all`
  //
  // Der erste Anlauf nahm `Promise.all`. Lehnt **eine** Bindung ihren Promise
  // ab — ein Abbruch über die Wails-Naht tut das —, verwirft `useLaden` das
  // **ganze** Ergebnis: keine Teilzahlen, kein Überblick. Der Kommentar sagte
  // das Gegenteil, und der Wächter daneben prüfte einen **fachlichen**
  // Fehlerwert in einem **erfüllten** Umschlag, den dieser Weg gar nicht
  // auspackt — der behauptete Transportweg blieb unbetreten.
  //
  // Angestrichen vom Zweitblick über Phase 3.5 als Befund 6.
  const zahlen = useLaden<Ueberblickszahlen>(
    useCallback(async () => {
      // **Die Exemplare zählt das Fenster** (U15, #2160): ein Ausschnitt
      // von einer Zeile samt Stückzahl, statt der ganzen Sammlung, nur um
      // ihre Mengen zu summieren.
      const [sets, sammlung, decks] = await Promise.allSettled([
        api().Sets(),
        api().Sammlungsfenster('', '', '', 0, 1),
        api().Decks(''),
      ])

      const gelesen = sets.status === 'fulfilled' ? sets.value.sets : []

      return {
        sets: gelesen.length,
        karten: gelesen.reduce((summe, eines) => summe + eines.karten, 0),
        exemplare: sammlung.status === 'fulfilled' ? sammlung.value.stueck : 0,
        decks: decks.status === 'fulfilled' ? decks.value.decks.length : 0,
      }
    }, []),
    // **Der Stand liest nach und fragt nicht neu** (U10, #2125): in den
    // Abhängigkeiten räumte er, und der Überblick verschwand bei jedem
    // Schreiben, bis die Antwort kam.
    [],
    false,
    bestandsstand,
  )

  // **Die Prüfstände der Sets kommen für sich** (AP-8, #2365) — gelesen,
  // nicht geprüft. **Die Zahlen warten nicht auf sie**: im selben Ladelauf
  // stand die Kennzahl bei 50 Sets kalt erst nach 526 ms (P95, `kennzahlen09`,
  // Grenze 500 ms aus PERF02); getrennt füllen sich die Unterzeilen danach in
  // ihrer schon stehenden Zeile.
  const probleme = useLaden<Problemlage>(
    useCallback(async () => {
      const antwort = await api().Setprobleme()

      return problemlage(auspacken(antwort, () => antwort.staende))
    }, []),
    [],
    false,
    bestandsstand,
  )

  // **Das Portfolio ebenso** (AP-8, Teil 2b): es rechnet über die ganze
  // Sammlung, und die Kachel „Sammlung“ trägt seine Summe als Unterzeile.
  // Gelesen, was im Bestand liegt — kein Preis wird dafür geholt.
  const portfolio = useLaden<Portfolioantwort>(
    useCallback(async () => {
      const antwort = await api().Portfolio()

      return auspacken(antwort, () => antwort)
    }, []),
    [],
    false,
    bestandsstand,
  )

  // **Die Aufmerksamkeit fragt hier und nicht in ihrer Liste** (AP-8, Teil
  // 2b): die Kachel „Decks“ zählt die Decks mit fehlenden Karten aus
  // derselben Antwort. Zweimal gefragt hieße zweimal jedes Deck mit der
  // Sammlung abgleichen und jedes Paketskript prüfen.
  const aufmerksamkeit = useAufmerksamkeit(bestandsstand)

  return (
    <>
      <Seitenkopf titel={start.ueberschrift} erklaerung={start.was} />

      {/* **Der Überblick steht immer** (#1325). Eine Bedingung wie
          `hatArbeit &&` ließe die Seite beim Eintreffen der Antwort
          **umspringen**: „noch nicht geladen" sähe aus wie „kein Set".

          **Und das gilt auch, solange er lädt** (AP-4, #2361, B-12): hier
          stand trotzdem eine solche Bedingung, `zahlen.daten !== undefined
          &&`, und alles darunter rückte beim Eintreffen um eine Kachelzeile.
          Jetzt stehen Platzhalter in seiner Endhöhe da.

          Die Einführung ist ein Fenster beim Start
          (`bausteine/Einfuehrung.tsx`), und ob es wiederkehrt, ist eine
          Einstellung. Was der Überblick bei leerem Bestand zeigt, ist seine
          eigene Sache — vier Zahlen, von denen drei null sind, sagen
          zutreffend, dass hier noch nichts liegt. */}
      <Ueberblick
        zahlen={zahlen.daten}
        probleme={probleme.daten}
        portfolio={portfolio.daten}
        aufmerksamkeit={aufmerksamkeit.daten}
        wechseln={wechseln}
        waehleSet={waehleSet}
        nachladen={probleme.nachladen}
      />

      {/* **Zwei Fragen nebeneinander**, wie im Artboard *Main* (UXR-03,
          #1911): *woran arbeite ich weiter?* und *was wartet?*. Untereinander
          standen sie mit dem Überblick und den Hintergrundaufgaben als vier
          gleichrangige Abschnitte, und die Seite rollte auf das Doppelte der
          Fensterhöhe. **Ab `lg`**, denn darunter gäbe jede Spalte einem
          Namen und seinem Knopf keine Zeile mehr — dann stehen sie
          untereinander, Weiterarbeiten zuerst.

          **Womit habe ich zuletzt gearbeitet?** ist die erste Frage — und die
          Ablage führt dieselbe Liste, aber ein Menü sieht nur, wer es
          aufklappt (#891). **Was Aufmerksamkeit braucht** kommt aus Befunden,
          nicht aus einer Vermutung. */}
      <div className="grid items-start gap-6 lg:grid-cols-2">
        <Weiterarbeiten
          wechseln={wechseln}
          waehleSet={waehleSet}
          stand={bestandsstand}
        />
        <Aufmerksamkeit
          waehleSet={waehleSet}
          uebersicht={aufmerksamkeit}
          setbefunde={probleme.daten?.mitProblemen}
        />
      </div>

      {/* **Was im Hintergrund läuft**, fragt sich nur, solange etwas läuft
          (#1422) — und es steht nur da, solange das Buch etwas führt. */}
      <Hintergrundaufgaben stand={bestandsstand} />
    </>
  )
}
