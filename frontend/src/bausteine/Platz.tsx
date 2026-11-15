// **Was Platz belegt, und wie er frei wird** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Der Anlass, gemessen
//
// **Am 2026-09-07 auf dem Rechner des Auftraggebers gezählt**: neben einem
// Bestand von 127 MB lagen **dreißig** Sicherungen vor Migrationen — zusammen
// über zwei Gigabyte, und nichts sagte es. Sie entstehen bei jedem Start, der
// migriert, und niemand räumte sie weg.
//
// # Warum nur ein Posten einen Knopf hat
//
// **Weil nur einer weg darf.** Der Bestand ist die Anwendung, die Bilder
// gehören zu Karten, und ein Paket entfernt man auf der Seite *Pakete* — dort
// steht daneben, welche Sets es benutzen. Ein Knopf, der hier alles
// wegräumte, nähme jemandem seine Arbeit.
import { useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import type { PlatzpostenDTO } from '../apiplatz'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

// platzvorgang ist die eine Kennung für jedes Wegräumen auf dieser Seite.
//
// **Eine feste und keine je Posten**: die Antwort ist die ganze Zählung, also
// ist der Datensatz die Anzeige (§7, Regel 4).
const platzvorgang = 'platz'

export function Platz() {
  const texte = useTexte()
  const eigene = texte.seiten.bestandspflege

  const stand = useLaden(holen, [])

  const [geraeumt, setzeGeraeumt] = useState<PlatzpostenDTO[] | undefined>(
    undefined,
  )
  const [fehler, setzeFehler] = useState<string>()

  // **Ein Vorgang je Anzeige, nicht je Posten** (§7, Regel 4) — Nachlauf zum
  // Zweitblick über Phase 4.13 (2026-09-07).
  //
  // Hier stand der Posten als Kennung, und das war der falsche Datensatz:
  // **jede Antwort trägt die ganze Zählung** und ersetzt den gesamten Stand.
  // Zwei Posten durften deshalb parallel laufen, und dann entschied die
  // Reihenfolge der Antworten über die angezeigte Zahl.
  //
  // **`useLauf` stand hier daneben und reichte nicht**, und der Nachlauf hat
  // gesagt, warum: es bevorzugt die später **gestartete** Antwort. Der neuere
  // **Stand** kann aber von der früher gestarteten kommen — wer zuerst klickt
  // und zuletzt löscht, hat die jüngere Zählung, und sie wurde verworfen.
  //
  // Zwei Wege waren zu haben: die Antworten vergleichen (woran? sie tragen
  // keinen Stand) oder **gar nicht erst parallel laufen lassen**. Das zweite
  // ist das ehrlichere: das Wegräumen dauert Millisekunden, und ein zweiter
  // Knopf, der solange angehalten ist, überrascht niemanden.
  //
  // **Und `useLauf` ist damit fort.** Es kann nicht mehr überholt werden —
  // ein zweiter Lauf beginnt erst, wenn der erste zu Ende ist. Ein Wächter,
  // dessen Zweig kein Aufruf betritt, ist eine Behauptung darüber, was im
  // Ernstfall geschieht (§3); die Abdeckung hat ihn angestrichen.
  const vorgang = useVorgang()

  const posten = geraeumt ?? stand.daten ?? []

  // fragt ist der Posten, dessen Wegräumen gerade bestätigt werden will —
  // OPS-06.AK-04 (#1723): **gelöscht wird erst nach einem Ja**, denn was
  // hier fortgeht, kommt nicht zurück.
  const [fragt, setzeFragt] = useState<string>()

  const freigeben = (welcher: string) =>
    vorgang.starte(platzvorgang, async () => {
      try {
        const antwort = await api().PlatzFreigeben(welcher)

        // **Die Zählung zuerst, auch wenn gleich ein Fehler kommt** (§7,
        // Regel 3): was weggeräumt wurde, **ist** weg. Eine Antwort, die nur
        // den Fehler zeigt und die alte Zahl stehen lässt, behauptet einen
        // Stand, den es nicht gibt.
        if (antwort.posten.length > 0) {
          setzeGeraeumt(antwort.posten)
        }

        // **Und dann der Fehler** — `auspacken` wirft ihn als
        // `Bindungsfehler`, damit `meldungAus` ihn im Fang unten über seinen
        // Schlüssel übersetzt. Ihn hier von Hand zu lesen gäbe den Satz aus
        // Go statt des Katalogtextes.
        auspacken(antwort, () => undefined)
        setzeFehler(undefined)

        // **Weggeräumt ist weggeräumt** (§7, Regel 3; #1970): kam die
        // Zählung danach nicht, gilt die alte nicht mehr. Gezählt wird über
        // den gewöhnlichen Weg, und was dort scheitert, steht an der Zählung.
        if (antwort.nachladefehler) {
          setzeGeraeumt(undefined)
          stand.nachladen()
        }
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  if (stand.fehler !== undefined) {
    return (
      <p role="alert" className="text-fehler">
        {texte.bausteine.zustaende.fehler} {stand.fehler}
      </p>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {/* **Über `Satz` und nicht als Text**: der Katalogtext trägt
          Hervorhebungen, und roh gezeigt stünden die Sternchen da (§9). */}
      <Erklaerungsabsatz
        text={eigene.platzErklaerung}
        klasse="max-w-prose text-leise"
      />

      <ul className="flex max-w-xl flex-col gap-2">
        {posten.map((einer) => (
          <li
            key={einer.posten}
            className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-kante px-3 py-2"
          >
            {/* **Eine Zeile, ein Textknoten.** Der erste Anlauf setzte die
                Größe in ein eigenes `<span>`, und damit war die Zeile für
                jeden Leser — Mensch wie Prüfung — in Stücke zerlegt. */}
            <span>{zeile(eigene, einer)}</span>

            {einer.raeumbar && einer.dateien > 1 && (
              <Knopf
                gewicht="leise"
                bezeichnung={einsetzen(eigene.platzFreigebenFuer, {
                  posten: namevon(eigene, einer.posten),
                })}
                angehalten={vorgang.laeuft(platzvorgang)}
                aufKlick={() => {
                  setzeFragt(einer.posten)
                }}
              >
                {eigene.platzFreigeben}
              </Knopf>
            )}

            {fragt === einer.posten && (
              <div className="w-full">
                <Entscheidung
                  frage={einsetzen(eigene.platzFreigebenFrage, {
                    posten: namevon(eigene, einer.posten),
                  })}
                  erklaerung={eigene.platzFreigebenErklaerung}
                  bestaetigen={eigene.platzLoeschenJa}
                  aktion="loeschen"
                  abbrechen={texte.bausteine.aktionen.abbrechen}
                  laeuft={vorgang.laeuft(platzvorgang)}
                  aufBestaetigen={() => {
                    setzeFragt(undefined)
                    void freigeben(einer.posten)
                  }}
                  aufAbbrechen={() => {
                    setzeFragt(undefined)
                  }}
                />
              </div>
            )}
          </li>
        ))}
      </ul>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}

// holen liest die Zählung.
async function holen(): Promise<PlatzpostenDTO[]> {
  const antwort = await api().Platz()

  return auspacken(antwort, () => antwort.posten)
}

// zeile schreibt einen Posten als einen Satz.
function zeile(
  eigene: {
    platzNamen: Record<string, string>
    platzDateien: string
  },
  posten: PlatzpostenDTO,
): string {
  const dateien =
    posten.dateien > 0
      ? ' · ' +
        einsetzen(eigene.platzDateien, { anzahl: String(posten.dateien) })
      : ''

  return `${namevon(eigene, posten.posten)}: ${inBytes(posten.bytes)}${dateien}`
}

// namevon gibt den Namen eines Postens aus dem Katalog.
//
// **Ein Posten, den der Katalog nicht kennt, steht mit seinem Schlüssel da.**
// Der Fall ist echt: die Liste kommt aus Go, und ein fünfter Posten dort ist
// einen Katalogeintrag früher da als hier. Ein leerer Name wäre eine Zeile,
// die niemand zuordnen kann (§6).
function namevon(
  eigene: { platzNamen: Record<string, string> },
  posten: string,
): string {
  const namen: Record<string, string | undefined> = eigene.platzNamen

  return namen[posten] ?? posten
}

// inBytes schreibt eine Zahl so, wie ein Mensch sie liest.
//
// **Tausender und keine Zweierpotenzen.** Der Finder zeigt „127 MB" für
// dieselbe Datei, und eine Zahl, die von seiner abweicht, sät Zweifel an
// beiden — dieselbe Rechnung wie in `cmd/mse/platz.go`.
export function inBytes(zahl: number): string {
  const tausend = 1000

  if (zahl < tausend) {
    return `${String(zahl)} B`
  }

  let wert = zahl

  for (const einheit of ['kB', 'MB', 'GB', 'TB']) {
    wert /= tausend

    if (wert < tausend) {
      return `${wert.toFixed(1)} ${einheit}`
    }
  }

  return `${(wert / tausend).toFixed(1)} PB`
}
