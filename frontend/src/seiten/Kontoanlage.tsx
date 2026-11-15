// **Ein eigenes Konto anlegen und sich damit anmelden** — Werkstück 6 von
// Phase 4.12
// (#1351).
//
// # Was diese Fläche sagt, bevor jemand tippt
//
// **Dass eine Adresse hier eine Kennung ist und kein nachgewiesener Kanal.**
// Das Programm verschickt nichts; es kann nicht wissen, ob es das Postfach
// gibt. Das zu verschweigen wäre eine Zusage, die niemand einlöst (§10) —
// deshalb steht es über dem Feld und nicht im Kleingedruckten.
//
// **Und dass das Kennwort nie zurückkommt.** Kein DTO trägt es, an keiner
// Fläche; wer es vergisst, setzt ein neues.
//
// # Ein Formular für zwei Wege
//
// **Weil beide dieselben zwei Felder brauchen** und niemand zwischen zwei
// Kästen suchen soll, welcher der richtige ist. Der Anzeigename steht nur
// beim Anlegen — beim Anmelden hätte er nichts zu bedeuten.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Bindungsfehler } from '../bindungsfehler'
import { Katalogfehler } from '../katalogfehler'
import { Kennwortfeld } from '../bausteine/Kennwortfeld'
import { Knopf } from '../bausteine/Knopf'
import { Nachweisfeld } from '../bausteine/Nachweisfeld'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'

import type { NutzerDTO } from '../api'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// Der Schlüssel des Vorgangs — **einer für beide Wege** (§7, Regel 4).
//
// **Weil sie einander ausschließen**: zwei Schlüssel ließen Anlegen und
// Anmelden mit derselben Kennung gleichzeitig laufen, und das zweite ginge
// gegen einen Stand, den es noch nicht gibt.
const KONTO = 'konto'

export function Kontoanlage({ geaendert }: { geaendert: () => void }) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer.konto

  const [kennung, setzeKennung] = useState('')
  const [kennwort, setzeKennwort] = useState('')
  const [name, setzeName] = useState('')

  // faktorGefragt sagt, ob die Anwendung einen zweiten Faktor verlangt hat.
  //
  // **Erst versuchen, dann fragen** (§6): wer keinen hat, sieht kein Feld.
  const [faktorGefragt, setzeFaktorGefragt] = useState(false)

  const [fehler, setzeFehler] = useState<string | undefined>(undefined)
  const [befund, setzeBefund] = useState<Record<string, string>>({})
  const [erfolg, setzeErfolg] = useState<string | undefined>(undefined)

  // **Ein Vorgang für beide Wege**: sie schließen einander aus, und getrennte
  // ließen Anlegen und Anmelden mit derselben Kennung gleichzeitig laufen.
  const vorgang = useVorgang()

  const fuehre = async (
    ruf: () => Promise<NutzerDTO | null>,
    satz: (nutzer: NutzerDTO) => string,
  ) => {
    setzeFehler(undefined)
    setzeBefund({})
    setzeErfolg(undefined)

    await vorgang.starte(KONTO, async () => {
      try {
        const nutzer = await ruf()

        // **Eine Antwort ohne Nutzer und ohne Fehler ist ein Fehler** (§8).
        // Sie kann nicht vorkommen — `nutzerNach` gibt entweder einen Nutzer
        // oder einen Fehler —, und **deshalb** steht sie hier: eine halbe
        // Antwort, die wie eine ganze aussieht, ließe die Seite „angelegt"
        // sagen, ohne dass etwas entstanden wäre.
        if (nutzer === null) {
          throw new Katalogfehler('kontoLeer')
        }

        setzeErfolg(satz(nutzer))
        setzeKennwort('')

        // **Geschrieben ist geschrieben** (§7, Regel 3): das Nachladen ist
        // Bestenfalls-Arbeit und macht das Konto nicht ungeschehen.
        geaendert()
      } catch (grund) {
        const meldung = meldungAus(grund, texte)

        // **Der Fehler steht am verursachenden Feld** (§6). Die Anwendung
        // nennt `kennung` oder `kennwort`; alles andere steht über dem
        // Formular, weil kein Feld gemeint ist.
        // **Ein fehlender zweiter Faktor ist keine Abweisung, sondern ein
        // fehlender Schritt.** Das Formular klappt sein Feld auf, statt
        // dieselbe Eingabe noch einmal zu verlangen.
        if (
          grund instanceof Bindungsfehler &&
          grund.schluessel === 'zweiterFaktor'
        ) {
          setzeFaktorGefragt(true)

          return
        }

        if (
          grund instanceof Bindungsfehler &&
          (grund.feld === 'kennung' || grund.feld === 'kennwort')
        ) {
          setzeBefund({ [grund.feld]: meldung })

          return
        }

        setzeFehler(meldung)
      }
    })
  }

  const eintragen = () =>
    fuehre(
      async () => {
        const antwort = await api().KontoEintragen(kennung, kennwort, name)

        return auspacken(antwort, () => antwort.nutzer)
      },
      (nutzer) => einsetzen(eigene.eingetragen, { name: nutzer.name }),
    )

  // **Der Nachweis reist als Argument und nicht über den Zustand.**
  //
  // **Weil `useState` erst beim nächsten Zeichnen gilt**: ein
  // `setzeNachweis(x); anmelden()` schickte den **vorigen** Wert — und der
  // erste Versuch hätte immer den leeren. Genau der Fehler, den §7 Regel 2
  // meint, nur eine Ebene tiefer.
  const anmelden = (nachweis = '') =>
    fuehre(
      async () => {
        const antwort = await api().KontoAnmelden(kennung, kennwort, nachweis)

        return auspacken(antwort, () => antwort.nutzer)
      },
      (nutzer) => einsetzen(eigene.angemeldet, { name: nutzer.name }),
    )

  const ruht = kennung.trim() === '' || kennwort === '' || vorgang.laeuft(KONTO)

  return (
    <section aria-label={eigene.eigenes} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß. */}
      <Erklaerungsabsatz
        text={eigene.eigenesErklaerung}
        klasse="max-w-prose text-leise"
      />
      <p className="max-w-prose text-leise">{eigene.kennwortNieZurueck}</p>

      <div className="flex max-w-xl flex-col gap-3">
        <Pflichtfeld
          beschriftung={eigene.kennung}
          wert={kennung}
          befund={befund.kennung}
          hinweis={eigene.kennungHinweis}
          kennung
          geaendert={setzeKennung}
        />

        <Kennwortfeld
          beschriftung={eigene.kennwort}
          wert={kennwort}
          befund={befund.kennwort}
          hinweis={eigene.kennwortHinweis}
          geaendert={setzeKennwort}
        />

        <Pflichtfeld
          beschriftung={eigene.anzeigename}
          wert={name}
          hinweis={eigene.anzeigenameHinweis}
          geaendert={setzeName}
        />

        {faktorGefragt && (
          <Nachweisfeld
            angehalten={vorgang.laeuft(KONTO)}
            abgeschickt={(gegeben) => {
              void anmelden(gegeben)
            }}
          />
        )}

        <div className="flex flex-wrap gap-3">
          <Knopf
            gewicht="haupt"
            angehalten={ruht}
            aufKlick={() => {
              void eintragen()
            }}
          >
            {eigene.eintragen}
          </Knopf>

          <Knopf
            gewicht="leise"
            angehalten={ruht}
            aufKlick={() => {
              void anmelden()
            }}
          >
            {eigene.anmelden}
          </Knopf>
        </div>
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
