// **Der Griff auf die Zwischenablage** — Werkstück 1 von Phase 3.7
// (#1168).
//
// # Ein Haken, weil es zwei Bedienungen sind
//
// Kopieren, Ausschneiden und Einfügen sind über die Knöpfe der Handlungszeile
// **und** über ⌘C/⌘X/⌘V in der Kartenliste erreichbar. Zwei Umsetzungen
// nebeneinander wären zwei Antworten auf dieselbe Frage — eine Änderung an der
// einen ließe die andere stehen, und niemand sähe, welche gilt.
//
// # ported: die drei hängen in MSE2 am **fokussierten** Bedienelement
//
//	#define CUT_COPY_PASTE(op,return) \
//	  int id = focused_control(this); \
//	  if      (id == ID_EDITOR)    { return editor->op();    } \
//	  else if (id == ID_CARD_LIST) { return card_list->op(); } \
//	  else if (id == ID_NOTES)     { return notes->op();     } \
//	  else                         { return false;           }
//
// `src/gui/set/cards_panel.cpp`, Zeilen 371–382, Fassung f9d9356 vom
// 2021-07-06. Deshalb hängen die Kürzel hier an der **Liste** und nicht am
// Fenster: wer im Kartentext steht und ⌘C drückt, meint den Text.
import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { EinfuegeAntwort } from '../apiablage'
import { auspacken, type Umschlag } from '../umschlag'
import type { Vorgang } from './useVorgang'

// Kartenablagegriff ist, was eine Bedienung braucht — und sonst nichts.
export interface Kartenablagegriff {
  kopieren: () => void
  ausschneiden: () => void
  einfuegen: () => void

  // laeuft sperrt die zweite Bedienung, solange die erste unterwegs ist
  // (§7, Regel 4).
  laeuft: boolean

  // anzahl ist, wie viele Karten Kopieren und Ausschneiden meinen — die
  // Aufschrift nennt sie, damit niemand raten muss.
  anzahl: number
}

interface Ablagelage {
  setID: string

  // angehakt sind die Karten der Massenänderung — **dieselbe Auswahl** trägt
  // die Zwischenablage. Eine zweite daneben wäre eine zweite Antwort auf
  // „welche Karten meine ich", und niemand könnte sagen, welche gilt.
  angehakt: readonly string[]

  // aufgeschlagen ist die offene Karte — sie gilt, wenn kein Haken steht.
  //
  // **Wie MSE2**: `doCopy` fragt `getSelection`, und die Liste hat immer eine
  // gewählte Zeile (`src/gui/control/card_list.cpp`, Zeilen 146–157, Fassung
  // f9d9356 vom 2021-07-06).
  aufgeschlagen: string | undefined

  // vorgang ist der **des Aufrufers**, nicht ein eigener. Derselbe Grund wie
  // beim Kartenverweis: ein Schloss in einem Bauteil, das der Bezugswechsel
  // abbaut, lebt nur so lange wie die Instanz.
  vorgang: Vorgang

  // aufFehler meldet nach oben — an die Handlungszeile, die den Wechsel
  // überlebt.
  aufFehler: (meldung: string | undefined) => void

  // geaendert meldet, dass der Bestand sich geändert hat.
  geaendert: () => void
}

export function useKartenablage({
  setID,
  angehakt,
  aufgeschlagen,
  vorgang,
  aufFehler,
  geaendert,
}: Ablagelage): Kartenablagegriff {
  const texte = useTexte()

  const karteIDs =
    angehakt.length > 0
      ? angehakt
      : aufgeschlagen === undefined
        ? []
        : [aufgeschlagen]

  // **Ein Vorgang je Set** (§7, Regel 4). Nicht je Tätigkeit: Kopieren und
  // Einfügen sind zugleich bedienbar, und zwei Schlüssel sperrten einander
  // dann nicht.
  const schluessel = `ablage-${setID}`

  // inDieAblage ruft eine der drei Bindungen und meldet, was zurückkommt.
  //
  // **`danach` läuft auch nach einem Fehler, wenn schon geschrieben wurde**
  // (§7, Regel 3) — Befund 4 des Zweitblicks über Phase 3.7.
  //
  // # Warum das nötig ist
  //
  // Einfügen legt Karte für Karte an, Ausschneiden löscht Karte für Karte.
  // Scheitert ein späterer Datensatz, **bleiben die früheren geschrieben**;
  // die Go-Seite gibt sie deshalb mit zurück. Der erste Anlauf ließ
  // `auspacken` bei **jedem** Fehler abbrechen und lud nicht nach: die Liste
  // zeigte dann gelöschte Karten weiter oder unterschlug angelegte, bis
  // jemand die Lasche wechselte.
  //
  // **Die Meldung bleibt trotzdem stehen.** Ein Teilerfolg ist kein Erfolg;
  // wer nicht liest, was schiefging, wiederholt es.
  const inDieAblage = (
    ruf: () => Promise<Umschlag>,
    danach: () => void,
    schonGeschrieben: (antwort: Umschlag) => boolean = () => false,
  ) => {
    void vorgang.starte(schluessel, async () => {
      let antwort: Umschlag | undefined

      try {
        antwort = await ruf()

        auspacken(antwort, () => antwort)
        aufFehler(undefined)
        danach()
      } catch (ursache: unknown) {
        aufFehler(meldungAus(ursache, texte))

        if (antwort !== undefined && schonGeschrieben(antwort)) {
          danach()
        }
      }
    })
  }

  return {
    kopieren: () => {
      inDieAblage(
        () => api().AblageKopieren(setID, [...karteIDs]),
        // **Kopieren ändert den Bestand nicht.** Ein Nachladen hier wäre ein
        // Ruf ohne Anlass — und einer, der scheitern kann.
        () => undefined,
      )
    },

    ausschneiden: () => {
      inDieAblage(
        () => api().AblageAusschneiden(setID, [...karteIDs]),
        // **Regel 3**: geschrieben ist geschrieben. Die Karten sind fort;
        // scheitert das Nachladen, bleibt es dabei.
        geaendert,
        // **Kein Teilschnitt möglich** — seit dem 2026-08-27 löscht die
        // Go-Seite alle Karten in **einer** Transaktion (Befund 4 des
        // Zweitblicks über Phase 3.7). Scheitert sie, ist keine fort.
      )
    },

    einfuegen: () => {
      inDieAblage(
        () => api().AblageEinfuegen(setID),
        geaendert,
        // **Eine Karte, die entstanden ist, verschwindet nicht** — sie steht
        // in der Antwort, auch wenn das Füllen danach scheiterte.
        // **Einfügen bleibt schrittweise**, und das ist kein Versäumnis:
        // eine Karte, die entstanden ist, **existiert** — sie zu verstecken
        // wäre schlimmer als sie zu zeigen. Die Go-Seite nennt sie deshalb
        // auch im Fehlerfall, und hier wird danach nachgeladen.
        (antwort) => (antwort as EinfuegeAntwort).karten.length > 0,
      )
    },

    laeuft: vorgang.laeuft(schluessel),
    anzahl: karteIDs.length,
  }
}
