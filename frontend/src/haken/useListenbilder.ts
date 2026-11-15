// **Die kleinen Bilder der Kartenliste** — Werkstück 2 von Phase 3.5
// (#1166).
//
// # Nur für das Sichtfenster
//
// Ein Set kann 3.788 Karten haben. Alle Bilder zu holen wäre dasselbe wie alle
// Zeilen zu zeichnen — der Fehler, den `useSichtfenster` gerade behebt.
// Geholt wird deshalb, was zu sehen ist, und zwar in **einem** Ruf je Fenster.
//
// MSE2 macht es lazily über einen Hintergrundstrang, der je gezeichneter Zeile
// eine Anfrage stellt (`ImageCardList::OnGetItemImage` →
// `thumbnail_thread.request`, `src/gui/control/image_card_list.cpp`, Zeilen
// 92–104, Fassung `5deaeee` vom 2020-05-16). Dieselbe Regel, ein Ruf statt
// dreißig.
//
// # Und was einmal da ist, bleibt
//
// Wer auf und ab rollt, sieht dieselben Zeilen wieder. MSE2 hält sie in einer
// `map` je Dateinamen und zusätzlich auf der Platte
// (`thumbnail_thread.cpp`, `image_cache_dir`); hier genügt eine Tabelle im
// Speicher — sie stirbt mit dem Setwechsel, und genau das soll sie (§7,
// Regel 1).
import { useEffect, useRef, useState } from 'react'

import { api, auspacken } from '../api'

export function useListenbilder(
  setID: string,
  sichtbare: readonly string[],

  // stand steigt bei jeder Änderung am Bestand.
  //
  // # Warum die Bilder ihn brauchen
  //
  // Wer ein Kartenbild **einsetzt oder ersetzt**, ändert die Datei, auf die
  // ein gemerktes Listenbild zeigt — der Kasten zeigte danach weiter das alte
  // oder blieb leer, bis jemand das Set wechselt. `geholt` merkt sich jede
  // Antwort und wurde nur beim Setwechsel geräumt. Angestrichen vom Nachlauf
  // über Phase 3.5 als Befund 5.
  //
  // **Ein Stand und keine Einzelmeldung**: welche Karte betroffen war, weiß
  // dieser Haken nicht, und ein Weg dorthin wäre eine Kante zwischen
  // Bildeditor und Kartenliste, die es nicht gibt.
  stand: number,
): Readonly<Record<string, string>> {
  // **Ein Zustand für beides — welches Set, und seine Bilder.**
  //
  // Ein zweiter Effekt, der beim Setwechsel `setzeBilder({})` ruft, wäre der
  // naheliegende Weg und ein falscher: `setState` im Effekt löst eine zweite
  // Zeichenrunde aus, in der die Liste kurz die Bilder des **vorigen** Sets
  // trägt. Steht das Set im Zustand, ist die Frage »gehören diese Bilder
  // hierher?« beim Lesen beantwortet und braucht keine Runde.
  const [gezeigt, setzeGezeigt] = useState<{
    fuer: string
    bilder: Record<string, string>
  }>({ fuer: setID, bilder: {} })

  // **Was schon **beantwortet** ist, wird nicht wieder gefragt** — auch nicht
  // das, was sich als bildlos herausgestellt hat. Ohne die zweite Hälfte
  // fragte jedes Rollen erneut nach jeder Karte ohne Illustration, und das
  // sind die meisten.
  //
  // **Beantwortet, nicht versucht.** Der erste Anlauf trug die Kennungen
  // **vor** dem Ruf ein und ließ sie auch dann stehen, wenn der Stapel
  // scheiterte: ein vorübergehender Fehler machte die Bilder bis zum
  // Setwechsel dauerhaft unsichtbar. Angestrichen vom Zweitblick über
  // Phase 3.5 als Befund 7.
  const geholt = useRef(new Set<string>())

  // laufend hält die Kennungen, nach denen gerade gefragt wird.
  //
  // **Damit dasselbe Fenster nicht zweimal fragt**, während die Antwort
  // unterwegs ist — und damit ein Fehlschlag sie wieder freigibt.
  const laufend = useRef(new Set<string>())

  // gesehenerStand ist das, woran das Gedächtnis hängt — Set **und** Stand.
  //
  // **Beides zusammen und nicht zwei Referenzen**: ein Wechsel ist ein
  // Wechsel, gleich welcher der beiden Werte sich ändert, und zwei getrennte
  // Prüfungen liefen beim ersten neuen Grund auseinander.
  const gesehenerStand = useRef(`${setID}/${String(stand)}`)

  // **Der Schlüssel ist der Inhalt, nicht die Kennung des Effekts.**
  //
  // `sichtbare` ist bei jedem Zeichnen ein neues Feld; als Abhängigkeit liefe
  // der Effekt endlos. Die verbundene Zeichenkette ändert sich nur, wenn sich
  // das Fenster wirklich verschoben hat.
  const fenster = sichtbare.join(',')

  // jetziger ist der Stand, für den gerade geholt wird.
  const jetziger = `${setID}/${String(stand)}`

  useEffect(() => {
    // **Bei jedem Wechsel beginnt das Gedächtnis von vorn.** Kennungen zählen
    // je Bestand und nicht je Set; ein gemerktes »hat kein Bild« träfe im
    // nächsten Set eine andere Karte. Und ein eingesetztes Bild macht ein
    // gemerktes ungültig, ohne dass die Kennung sich ändert.
    if (gesehenerStand.current !== jetziger) {
      gesehenerStand.current = jetziger
      geholt.current = new Set<string>()
      laufend.current = new Set<string>()
    }

    const offene = fenster
      .split(',')
      .filter(
        (id) =>
          id !== '' && !geholt.current.has(id) && !laufend.current.has(id),
      )

    if (offene.length === 0) {
      return
    }

    // **Erst, wenn das Rollen innehält** — E07 (#2279). Das erste Rollen
    // durch Black fragte für jedes Fenster, an dem die Liste vorbeizog, 87
    // Mal in vier Sekunden, und Go renderte die Bildchen gleichzeitig,
    // während WebKit die Bilder des Rollens zeichnen wollte. Ein neues
    // Fenster vor Ablauf der Ruhe hebt die Frage auf; als `laufend` gilt eine
    // Kennung erst, wenn wirklich gefragt wird.
    const zeitgeber = setTimeout(() => {
      fragen(offene)
    }, listenbilderruhe)

    return () => {
      clearTimeout(zeitgeber)
    }

    // fragen holt die Bildchen eines Fensters.
    function fragen(offene: string[]) {
      for (const id of offene) {
        laufend.current.add(id)
      }

      void (async () => {
        try {
          const antwort = await api().Kartenbilder(setID, offene)
          const neue = auspacken(antwort, () => antwort.bilder)

          // **Eine Antwort von vorhin fasst nichts mehr an** (§7, Regel 2).
          //
          // Der erste Anlauf prüfte das Set erst ganz am Ende, beim **Lesen**.
          // Das schützte weder die Buchführung noch den Zustandssetzer: eine
          // späte Antwort des alten Sets durfte Kennungen im neuen als geholt
          // markieren, Einträge eines neuen Laufs aus `laufend` löschen und den
          // ganzen Stand auf das alte Set zurückstellen. Angestrichen vom
          // Nachlauf über Phase 3.5 als Befund 2.
          //
          // **Der Vergleich steht vor jedem Schreiben**, nicht danach.
          if (gesehenerStand.current !== jetziger) {
            return
          }

          // **Beantwortet — jetzt wird gemerkt.**
          for (const id of offene) {
            geholt.current.add(id)
          }

          // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7,
          // Regel 2) — und der Stand ist hier das **Set**, nicht das Fenster.
          //
          // Der erste Anlauf zählte einen Lauf je Sichtfenster mit: wer weiter
          // rollte, während der vorige Stapel lief, verwarf dessen Antwort —
          // obwohl beide zum selben Set gehören. Beim Zurückrollen blieben die
          // Bilder leer, denn gemerkt waren die Kennungen trotzdem.
          // Angestrichen vom Zweitblick über Phase 3.5 als Befund 8.
          //
          // **Zwei Stapel desselben Sets ergänzen einander**, sie überholen sich
          // nicht: `setzeStand` mischt sie, und die Reihenfolge ihrer Antworten
          // ist gleichgültig.
          setzeGezeigt((bisher) =>
            bisher.fuer === setID
              ? { fuer: setID, bilder: { ...bisher.bilder, ...neue } }
              : { fuer: setID, bilder: neue },
          )
        } catch {
          // **Ein Bild, das nicht kommt, ist kein Fehler der Liste** (§8). Die
          // Zeile steht mit ihrem Namen da; ein roter Hinweis über einer
          // vollständigen Liste sagte etwas Falscheres als gar keiner.
          //
          // **Aber es bleibt fragbar.** Die Kennungen wandern nicht nach
          // `geholt`; wer noch einmal an dieser Stelle vorbeirollt, fragt
          // erneut (Befund 7).
        } finally {
          // **Auch hier nur für den eigenen Stand.** Ein alter Block, der nach
          // einem Wechsel aufräumte, löschte Einträge eines **neuen** Laufs —
          // und dessen Kennungen würden dann doppelt gefragt.
          if (gesehenerStand.current === jetziger) {
            for (const id of offene) {
              laufend.current.delete(id)
            }
          }
        }
      })()
    }
  }, [setID, fenster, jetziger])

  // **Bilder eines fremden Sets werden gar nicht erst gezeigt.** Zwischen dem
  // Wechsel und der ersten Antwort steht der alte Stand noch da; ihn hier zu
  // verwerfen ist eine Frage beim Lesen und kostet keine Zeichenrunde.
  return gezeigt.fuer === setID ? gezeigt.bilder : leer
}

// listenbilderruhe ist die Zeit in Millisekunden, die das Sichtfenster still
// stehen muss, bevor seine Bildchen geholt werden — kurz genug, dass sie beim
// Anhalten gleich da sind, lang genug, dass ein Rollen mit 60 Bildern je
// Sekunde nicht jedes Fenster fragt.
export const listenbilderruhe = 120

// leer ist die Tabelle, die es gibt, solange nichts geholt ist.
//
// **Eine gemeinsame und keine je Zeichnung**: ein frisches `{}` wäre bei jedem
// Lauf ein neuer Wert, und jeder Verbraucher mit `bilder` in seinen
// Abhängigkeiten zeichnete erneut.
const leer: Readonly<Record<string, string>> = {}
