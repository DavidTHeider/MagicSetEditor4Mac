// **Ein Paket aus einem Gerüst anlegen** —
// #1350.
//
// # Warum es hier steht und nicht in `Paketliste`
//
// **Weil sie zwei Fragen beantworten** (§2): die Liste zeigt, was da ist,
// dieses Formular legt etwas an. Und weil die Liste mit dem Formular über die
// Zeilengrenze ginge.
//
// # Was das Fenster nicht selbst entscheidet
//
// **Welche Arten anlegbar sind, sagt die Anwendung** (`PaketArten`). Zwei von
// sieben Paketendungen lassen sich leer anlegen; eine Aufzählung hier wäre
// eine zweite Wahrheit, und sie stünde still falsch da, sobald eine dritte
// dazukommt.
import { useEffect, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Feld } from '../bausteine/Feld'
import { Knopf } from '../bausteine/Knopf'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { Suchauswahl } from '../bausteine/Suchauswahl'
import { Bindungsfehler } from '../bindungsfehler'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Anlegbarkeit, PaketDTO } from '../api'

const feldklassen =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Eigenschaften {
  // spiele sind die installierten Spielpakete — ein Stil gehört zu einem.
  spiele: readonly PaketDTO[]

  // angelegt meldet nach oben, dass die Liste neu zu lesen ist.
  angelegt: () => void
}

export function Paketanlage({ spiele, angelegt }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.pakete

  const [arten, setzeArten] = useState<readonly Anlegbarkeit[]>([])
  const [art, setzeArt] = useState('mse-game')
  const [name, setzeName] = useState('')
  const [vollname, setzeVollname] = useState('')
  const [spiel, setzeSpiel] = useState('')
  const [laeuft, setzeLaeuft] = useState(false)
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  // **feldfehler steht neben `fehler` und nicht darin** — Befund 1 des
  // Nachlaufs über Phase 4.10. Die Anwendung meldet ihren Fehler **mit** dem
  // Feld, das ihn verursacht hat; das Fenster warf den Bezug weg und zeigte
  // einen anonymen Hinweis über dem Formular. §6 sagt das Gegenteil.
  const [feldfehler, setzeFeldfehler] = useState<
    Record<string, string> | undefined
  >(undefined)
  const [erfolg, setzeErfolg] = useState<string | undefined>(undefined)

  // **Ein Lauf, der seinen Stand mitzählt** (§7, Regel 2): wer das Formular
  // schnell zweimal öffnet, bekäme sonst die Antwort des ersten Laufs in den
  // zweiten Stand.
  useEffect(() => {
    let gilt = true

    // **Ein Ausgang und eine Wache** — nicht zwei.
    //
    // Der erste Anlauf prüfte `gilt` im Erfolgs- **und** im Fehlerfall; das
    // sind vier Zweige für eine Frage, und einer davon ließ sich nur mit
    // einem Lauf herstellen, der abgeräumt wird und **dann** scheitert.
    // Erst rechnen, dann einmal fragen, ob das Ergebnis noch jemanden
    // angeht (§7, Regel 2).
    const holen = async () => {
      let ausgang: { arten?: readonly Anlegbarkeit[]; fehler?: string }

      try {
        ausgang = { arten: await api().PaketArten() }
      } catch (grund) {
        // **Ein Fehler steht am Formular und nicht in der Konsole** (§8).
        // Ohne dieses `catch` ging die Ablehnung ins Leere: der Lauf meldete
        // sie als unbehandelt, und die Seite zeigte ein Formular ohne Arten,
        // dessen Knopf nichts tut.
        ausgang = { fehler: meldungAus(grund, texte) }
      }

      if (!gilt) {
        return
      }

      setzeArten(ausgang.arten ?? [])
      setzeFehler(ausgang.fehler)
    }

    void holen()

    return () => {
      gilt = false
    }

    // **`texte` steht hier, weil der Fehlerweg es liest.** Die Folge ist, dass
    // ein Sprachwechsel die Arten neu holt — das ist ein Aufruf, der eine feste
    // Liste zurückgibt, und dafür steht die Meldung danach in der richtigen
    // Sprache. Sie aus dem Effekt herauszuhalten hieße, den rohen Grund im
    // Zustand zu halten und erst beim Zeichnen zu übersetzen: mehr Zustand für
    // weniger.
  }, [texte])

  // **Ob ein Spiel gebraucht wird, sagt die Art** — und die Anwendung sagt,
  // welche Art es braucht. Ein `art === 'mse-style'` hier wäre dieselbe
  // zweite Wahrheit eine Zeile tiefer.
  const brauchtSpiel =
    arten.find((eine) => eine.art === art)?.brauchtSpiel ?? false

  const anlegen = async () => {
    setzeLaeuft(true)
    setzeFehler(undefined)
    setzeFeldfehler(undefined)
    setzeErfolg(undefined)

    try {
      const antwort = await api().PaketAnlegen(art, name, vollname, spiel)

      auspacken(antwort, () => antwort.paket)

      setzeErfolg(
        einsetzen(eigene.anlegenGelungen, {
          name: antwort.paket?.name ?? name,
        }),
      )
      setzeName('')
      setzeVollname('')

      // **Geschrieben ist geschrieben** (§7, Regel 3): das Nachladen der
      // Liste ist Bestenfalls-Arbeit und macht das Paket nicht ungeschehen.
      angelegt()
    } catch (grund) {
      // **Über den Katalog und nicht über `err.message`** — sonst stünde die
      // Meldung aus Go auf einer deutschen Seite englisch da, und ein
      // `Katalogfehler` gäbe seinen Schlüssel statt seines Satzes (§9).
      const satz = meldungAus(grund, texte)

      // **Nennt der Fehler ein Feld, steht er dort** und nicht über dem
      // Formular. Der Schlüssel entscheidet, welcher Satz es wird — seit
      // #1350 hat die Namensregel einen eigenen, damit sie nicht als „Die
      // Eingabe ist nicht gültig" endet.
      if (grund instanceof Bindungsfehler && grund.feld !== undefined) {
        setzeFeldfehler({ [grund.feld]: satz })

        return
      }

      setzeFehler(satz)
    } finally {
      setzeLaeuft(false)
    }
  }

  // **Jede Art bei ihrem Namen** (FMT-02, #1691) — hier stand „kein Spiel,
  // also ein Stil“, und Include wie Exportvorlage hießen „Stil“. Eine Art,
  // die die Anwendung neu anbietet und die hier noch keinen Namen hat, steht
  // mit ihrer Endung da und nicht unter einem falschen.
  const artnamen: Readonly<Record<string, string>> = {
    'mse-game': eigene.anlegenArtSpiel,
    'mse-style': eigene.anlegenArtStil,
    'mse-include': eigene.anlegenArtInclude,
    'mse-export-template': eigene.anlegenArtExport,
  }

  return (
    // **Ohne eigene Überschrift** — seit UXR-17 (#1925) steht die Anlage
    // eingeklappt unter dem Arbeitsplatz, und der Schalter des Kastens trägt
    // den Namen. Der Bereich behält ihn als `aria-label`.
    <section aria-label={eigene.anlegen} className="flex flex-col gap-3">
      <div className="flex flex-wrap items-start gap-3">
        <Feld beschriftung={eigene.anlegenArt}>
          {(verbindung) => (
            <select
              {...verbindung}
              value={art}
              onChange={(ereignis) => {
                setzeArt(ereignis.target.value)
              }}
              className={feldklassen}
            >
              {arten.map((eine) => (
                <option key={eine.art} value={eine.art}>
                  {artnamen[eine.art] ?? eine.art}
                </option>
              ))}
            </select>
          )}
        </Feld>

        <Pflichtfeld
          beschriftung={eigene.anlegenName}
          wert={name}
          befund={feldfehler?.name}
          hinweis={eigene.anlegenNameHinweis}
          kennung
          geaendert={setzeName}
        />

        {/* **Der Hinweis steht unter seinem Feld** — über `Feld`, wie beim
            Kurznamen, und hängt über `aria-describedby` daran, nicht im
            `label`: sonst würde er Teil des Feldnamens. Bis UXR-17 (#1925)
            stand er als eigenes Glied der umbrechenden Zeile und landete bei
            Stil und Exportvorlage links unter dem Formular, getrennt von
            seinem Feld (Hüllenprobe `fmt02d`). */}
        <Feld
          beschriftung={eigene.anlegenVollname}
          hinweis={eigene.anlegenVollnameHinweis}
        >
          {(verbindung) => (
            <input
              {...verbindung}
              value={vollname}
              onChange={(ereignis) => {
                setzeVollname(ereignis.target.value)
              }}
              className={feldklassen}
            />
          )}
        </Feld>

        {brauchtSpiel && (
          <Feld
            beschriftung={eigene.anlegenSpiel}
            hinweis={eigene.anlegenSpielHinweis}
          >
            {(verbindung) => (
              // **Eine Auswahl mit Suche** — UXR-F1
              // (#2007),
              // R7: die Spiele wachsen mit dem, was jemand bezieht. **Die
              // erste Zeile ist „kein Spiel"** — leer, wie die Anwendung es
              // erwartet.
              <Suchauswahl
                id={verbindung.id}
                bezeichnung={eigene.anlegenSpiel}
                eintraege={[
                  { wert: '', name: eigene.anlegenSpielLeer },
                  ...spiele.map((eines) => ({
                    wert: eines.name,
                    name: eines.name,
                  })),
                ]}
                wert={spiel}
                aendern={setzeSpiel}
                beschriebenVon={verbindung['aria-describedby']}
                ungueltig={verbindung['aria-invalid']}
              />
            )}
          </Feld>
        )}
      </div>

      {/* **Der Knopf steht unter den Feldern**, nicht als letztes Glied der
          umbrechenden Zeile: die Felder sind verschieden hoch, seit ihre
          Hinweise darunter stehen. */}
      <div>
        <Knopf
          gewicht="haupt"
          angehalten={laeuft || name.trim() === ''}
          aufKlick={() => {
            void anlegen()
          }}
        >
          {eigene.anlegenKnopf}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {erfolg !== undefined && (
        <p role="status" className="text-leise">
          {erfolg}
        </p>
      )}
    </section>
  )
}
