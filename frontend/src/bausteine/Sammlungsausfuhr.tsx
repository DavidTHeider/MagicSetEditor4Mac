// **Die Sammlung ausgeben und zurücklesen** — Werkstück 4 von Phase 4.5
// (#1292).
//
// # Warum ein eigener Kasten neben der Einfuhr
//
// Weil es eine andere Frage ist: dort *was habe ich noch nicht*, hier *was
// habe ich, und wie bekomme ich es woanders hin*. Der Schnitt folgt der Frage
// (§2).
//
// # Warum die Ausgabe in einem Feld steht und keine Datei wird
//
// **Weil ein Dateidialog eine Fähigkeit ist, die drei der vier Flächen nicht
// haben** — und weil ein Textfeld etwas kann, das eine Datei nicht kann: man
// sieht hinein. Wer eine Datei will, kopiert und legt sie ab; wer nur
// nachsehen will, was drinsteht, muss nichts speichern.
import { useId, useRef, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Einfuhrvorschau, braucheVorschau } from './Einfuhrvorschau'
import { Teilen } from './Teilen'

import type {
  SammelvorschauDTO,
  Zusammenfuehrung,
} from '../antwortenlisteneinfuhr'

export interface SammlungsausfuhrEigenschaften {
  // aufEingelesen läuft, nachdem ein CSV zurückgelesen wurde.
  aufEingelesen: () => void
}

export function Sammlungsausfuhr({
  aufEingelesen,
}: SammlungsausfuhrEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const feldID = useId()

  const [text, setzeText] = useState('')

  // form ist die Gestalt, in der der Text zuletzt entstanden ist.
  //
  // **Sie entscheidet, ob zurückgelesen werden darf** — Befund 9 des
  // Zweitblicks über Phase 4.5: ein Listentext ist **verlustbehaftet** und
  // gehört nicht in den CSV-Leser.
  //
  // **`undefined` heißt: von Hand getippt.** Wer selbst ein CSV einfügt, soll
  // es zurücklesen können — die Form kennt dann nur der Text selbst, und der
  // CSV-Leser sagt es, wenn die Kopfzeile fehlt.
  const [form, setzeForm] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()
  const [bericht, setzeBericht] = useState<string>()

  // befunde sind die Zeilen, die nicht durchkamen (Befund 4).
  const [befunde, setzeBefunde] = useState<string[]>([])

  // ausgelassen zählt, was beim Auslassen stehen blieb (DOM-03.AK-01).
  const [ausgelassen, setzeAusgelassen] = useState(0)

  // vorgemerkt ist die Vorschau, die auf die Wahl wartet (DOM-03.AK-01,
  // #1698) — sie gilt dem Text im Feld, und eine Änderung dort räumt sie.
  //
  // **Samt dem Text, den sie beschreibt** — Befund 3 des Zweitblicks über
  // Phase 6.6: *Einlesen* liest genau dieses CSV und nicht, was inzwischen
  // im Feld steht.
  const [vorgemerkt, setzeVorgemerkt] = useState<{
    vorschau: SammelvorschauDTO
    csv: string
  } | null>(null)

  // textstand zählt die Wechsel des Textes mit (§7, Regel 2): eine Antwort
  // auf einen Text, der inzwischen ein anderer ist, wird verworfen — Erfolg
  // wie Fehler.
  const textstand = useRef(0)

  const vorgang = useVorgang()

  // ausgeben holt die Sammlung in der gewünschten Form.
  //
  // **Ohne Laufzähler, anders als die Listeneinfuhr** (§7, Regel 2). Der
  // Nachlauf über Phase 4.5 hat die Gestalt dort angestrichen, und nachgesehen
  // wurde auch hier: `vorgang.starte('ausgeben', …)` weist einen zweiten Ruf
  // ab, solange der erste läuft (Regel 4) — zwei Ausgaben können sich also
  // nicht überholen. Ein Zähler daneben hätte einen Zweig, den kein Aufruf
  // betritt.
  const ausgeben = (form: string) => {
    void vorgang.starte('ausgeben', async () => {
      try {
        const antwort = await api().SammlungAusgeben(form)

        setzeFehler(undefined)
        setzeBericht(undefined)
        setzeBefunde([])
        setzeVorgemerkt(null)
        setzeForm(form)
        setzeText(auspacken(antwort, () => antwort.text))
        textstand.current += 1
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  // zurueck fragt erst, was das CSV zusammenführte — DOM-03.AK-01 (#1698).
  // **Trifft keine Zeile eine vorhandene, geht es ohne Zwischenschritt
  // weiter**; sonst steht die Vorschau da, und gelesen wird erst mit der
  // Wahl. Die eigene Ausfuhr ein zweites Mal träfe jede Zeile.
  const zurueck = () => {
    const csv = text
    const stand = textstand.current

    void vorgang.starte('zurueck', async () => {
      try {
        const antwort = await api().SammlungCSVVorschauen(csv)
        const vorschau = auspacken(antwort, () => antwort.vorschau)

        if (textstand.current !== stand) {
          return
        }

        if (braucheVorschau(vorschau)) {
          setzeFehler(undefined)
          setzeVorgemerkt({ vorschau, csv })

          return
        }
      } catch (ursache: unknown) {
        if (textstand.current === stand) {
          setzeFehler(meldungAus(ursache, texte))
        }

        return
      }

      await zuruecklesen(csv, 'addieren')
    })
  }

  // mitWahl liest ein, was die Vorschau zeigte — **das CSV, das sie
  // beschreibt**, so zusammengeführt, wie gewählt.
  const mitWahl = (csv: string, modus: Zusammenfuehrung) => {
    setzeVorgemerkt(null)

    void vorgang.starte('zurueck', () => zuruecklesen(csv, modus))
  }

  // zuruecklesen liest ein CSV zurück.
  const zuruecklesen = async (csv: string, modus: Zusammenfuehrung) => {
    try {
      // **Ohne Herkunft, und das ist eine Aussage** (Befund F03,
      // [#1747]). Die Bindung nimmt sie als zweiten Parameter; hier steht
      // kein Dateiname, weil hier keine Datei war — der Text kommt aus
      // diesem Feld, eingefügt oder aus der eigenen Ausfuhr. Leer heißt
      // „niemand hat es gesagt", und dann wird an der Zeile nichts
      // behauptet.
      //
      // **Der Dateiweg bekommt seine Herkunft, wenn es ihn gibt**: das
      // Artboard `Einlesen` des Entwurfs zeigt den Namen über der Vorschau,
      // und dieser Schritt entsteht mit der Vorschau, nicht vorher.
      const antwort = await api().SammlungCSVEinlesen(csv, '', modus)

      // **Geschrieben ist geschrieben** (§7, Regel 3). Dieselbe Naht wie
      // bei der Listeneinfuhr schickt Bericht und Fehler zugleich, und
      // dieselbe Hülle warf den Bericht weg: `auspacken` wirft, bevor
      // jemand ihn liest (Nachlauf über Phase 4.5, Befund 3).
      const gelesen = antwort.bericht

      setzeBefunde(gelesen?.befunde ?? [])
      setzeAusgelassen(gelesen?.ausgelassen ?? 0)
      setzeBericht(
        gelesen === null
          ? undefined
          : einsetzen(eigene.listeBericht, {
              angelegt: String(gelesen.angelegt.length),
              erhoeht: String(gelesen.erhoeht),
              stueck: String(gelesen.stueck),
            }),
      )

      // **Nachgeladen wird nur, wenn etwas geschrieben wurde.** Ohne
      // Bericht ist gar nichts passiert, und ein Lesefehler überschriebe
      // die Meldung, die sagt, was zu tun ist.
      if (gelesen !== null) {
        aufEingelesen()
      }

      auspacken(antwort, () => gelesen)

      setzeFehler(undefined)
    } catch (ursache: unknown) {
      setzeFehler(meldungAus(ursache, texte))
    }
  }

  return (
    // **Überschrift und Rahmen trägt der Kasten darum** (UXR-10): die Ausgabe
    // steht eingeklappt in der Schublade *Einlesen und ausgeben*.
    <section aria-label={eigene.ausfuhrTitel} className="flex flex-col gap-3">
      <Erklaerungsabsatz
        text={eigene.ausfuhrErklaerung}
        klasse="max-w-3xl text-xs text-leise"
      />

      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          onClick={() => {
            ausgeben('csv')
          }}
          disabled={vorgang.laeuft('ausgeben')}
          className={ausfuhrknopf}
        >
          {eigene.ausfuhrAlsCSV}
        </button>

        <button
          type="button"
          onClick={() => {
            ausgeben('liste')
          }}
          disabled={vorgang.laeuft('ausgeben')}
          className={ausfuhrknopf}
        >
          {eigene.ausfuhrAlsListe}
        </button>

        {/* **Zurücklesen nur, wenn etwas dasteht.** Ein Knopf, der eine leere
            Eingabe schickte, bekäme einen Fehler, den niemand verursacht hat. */}
        {text !== '' && form !== 'liste' && (
          <button
            type="button"
            onClick={zurueck}
            disabled={vorgang.laeuft('zurueck')}
            className={ausfuhrknopf}
          >
            {eigene.ausfuhrZurueck}
          </button>
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {/* **Die Vorschau steht vor dem Lesen** — DOM-03.AK-01 (#1698), und
          nur, wenn eine Zeile eine vorhandene träfe. */}
      {vorgemerkt !== null && (
        <Einfuhrvorschau
          vorschau={vorgemerkt.vorschau}
          laeuft={vorgang.laeuft('zurueck')}
          aufEinlesen={(modus) => {
            mitWahl(vorgemerkt.csv, modus)
          }}
          aufAbbrechen={() => {
            setzeVorgemerkt(null)
          }}
        />
      )}

      {bericht !== undefined && <p className="text-sm text-text">{bericht}</p>}

      {ausgelassen > 0 && (
        <p className="text-sm text-text">
          {einsetzen(eigene.listeAusgelassen, {
            ausgelassen: String(ausgelassen),
          })}
        </p>
      )}

      {/* **Was nicht durchkam, steht da** (Befund 4). */}
      {befunde.map((befund) => (
        <p key={befund} className="text-sm text-hinweis">
          {befund}
        </p>
      ))}

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium" id={feldID}>
          {eigene.ausfuhrErgebnis}
        </span>
        {/* **Beschreibbar und nicht nur lesbar**: hier kommt auch ein CSV
            herein, das jemand anderswo abgelegt hat. */}
        <textarea
          aria-labelledby={feldID}
          value={text}
          rows={4}
          placeholder={eigene.ausfuhrLeer}
          onChange={(ereignis) => {
            setzeText(ereignis.target.value)
            textstand.current += 1

            // **Getippter Text hat keine Form** (Befund 9): der Rückleseknopf
            // kommt wieder, und der CSV-Leser sagt, wenn die Kopfzeile fehlt.
            //
            // # Warum das so bleibt, obwohl der Nachlauf es anstrich
            //
            // Der Nachlauf über Phase 4.5 sah die Klasse offen: schon ein
            // angehängtes Zeichen bietet einen **verlustbehafteten**
            // Listentext wieder als CSV an. Nachgemessen wurde, was dann
            // passiert — `sammelliste.kopfLesen` weist ihn mit „die Spalte
            // \"printing\" fehlt — ohne sie ist eine Zeile kein Exemplar" ab.
            // **Es geht nichts verloren**; es kostet einen Klick und gibt
            // einen Satz, der sagt, was fehlt.
            //
            // Die Alternative — die Herkunft über Bearbeitungen hinweg raten —
            // wäre in beide Richtungen falsch: ein zu einem gültigen CSV
            // umgeschriebener Text bekäme den Knopf nie zurück.
            //
            // **Was gefehlt hat, war der Beleg**, nicht die Entscheidung: der
            // Satz oben stand ungeprüft da. Jetzt hält ihn ein Lauf.
            setzeForm(undefined)

            // **Und die Vorschau gilt dem Text von eben** (§7, Regel 1).
            setzeVorgemerkt(null)
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>

      {/* **Die Sammlung als CSV teilen** — NET-04 (#1708). */}
      <Teilen art="sammlung" bezug="" />
    </section>
  )
}

// ausfuhrknopf ist die Gestalt der drei Knöpfe.
const ausfuhrknopf =
  'rounded-sm border border-kante px-3 py-1 text-sm focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-50 '
