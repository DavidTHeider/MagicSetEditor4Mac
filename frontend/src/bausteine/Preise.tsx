// **Die Preise eines Drucks** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Geholt wird auf Anweisung
//
// Beim Aufschlagen wird **gelesen**, was dasteht — nicht gefragt. Ein Programm,
// das für dreihundert Zeilen dreihundert Abrufe auslöst, überrascht seine
// Nutzenden (§16, und dasselbe Argument wie beim Paketbezug).
//
// # Und jede Zahl trägt ihr Alter
//
// Die Quelle sagt über ihre Preise ausdrücklich, dass sie für keinen einsteht.
// Ein Preis ohne Alter und Quelle wäre hier eine Behauptung.
import { useState } from 'react'

import type { Preisreihe, PreisstandDTO } from '../antworten'

import { Aufklappkasten } from './Aufklappkasten'
import { Knopf } from './Knopf'
import { Fremdverlauf } from './Fremdverlauf'
import { Marktangebote } from './Marktangebote'
import { Marktlinke } from './Marktlinke'
import { Preisverlauf } from './Preisverlauf'
import { api } from '../fassade'
import { Bindungsfehler } from '../bindungsfehler'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { geldAusText } from '../geld'
import { alsKalendertag } from '../zeitpunkt'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'

interface PreiseEigenschaften {
  printing: string

  // aufGeholt sagt dem Aufrufer, dass neue Preise dastehen.
  //
  // # Warum das nach oben gemeldet wird
  //
  // Die **Übersicht** rechnet aus denselben Preisen. Ohne diese Meldung stünde
  // sie nach einem Abruf weiter auf „noch keine Preise geholt" — gesehen im
  // laufenden Fenster, und kein Lauf hatte es gezeigt: beide Bauteile für sich
  // waren richtig, die Naht zwischen ihnen fehlte.
  aufGeholt: () => void
}

// reiheVon nimmt die Reihe aus dem Stand, den jemand angeklickt hat.
//
// **Alle sieben Teile und keiner erraten** (A3, #1442): mit einer zweiten
// Quelle wäre ein Ergänzen hinter der Naht die Antwort auf eine Frage, die
// niemand gestellt hat.
function reiheVon(preis: PreisstandDTO): Preisreihe {
  return {
    printing: preis.printing,
    quelle: preis.quelle,
    markt: preis.markt,
    art: preis.art,
    waehrung: preis.waehrung,
    finish: preis.finish,
    zustand: preis.zustand,
    sprache: preis.sprache,
  }
}

// reihenschluessel macht aus einer Reihe eine Zeichenkette zum Vergleichen.
function reihenschluessel(reihe: Preisreihe): string {
  return JSON.stringify(reihe)
}

export function Preise({ printing, aufGeholt }: PreiseEigenschaften) {
  const texte = useTexte()
  const darstellung = useDarstellung()
  const eigene = texte.seiten.sammlung

  const [runde, setzeRunde] = useState(0)
  const [holfehler, setzeHolfehler] = useState<string>()

  // holbefunde sind die Begleitbefunde des letzten Holens.
  //
  // **Getrennt vom Fehler, und das ist der Punkt** (#1202):
  // Go reicht unlesbare Preisfelder ausdrücklich bis zur Antwort, und sie
  // kommen **mit** einem Erfolg wie **neben** einem Fehler. Wer nur den
  // Fehler zeigte, ließe jemanden einen geglückten Abruf sehen, in dem eine
  // Reihe fehlt.
  const [holbefunde, setzeHolbefunde] = useState<readonly string[]>([])

  // reihe ist die Reihe, deren Verlauf offen steht — `undefined` heißt: keine.
  //
  // **Eine und nicht mehrere.** Drei offene Verläufe nebeneinander wären drei
  // Listen ohne Beschriftung; wer vergleichen will, klappt um.
  //
  // **Als Paar und nicht als `waehrung/finish`.** Der erste Anlauf legte beides
  // in eine Zeichenkette und trennte sie beim Zeichnen wieder — mit zwei
  // `?? ''` für einen Fall, den es nicht gibt (§3).
  const [reihe, setzeReihe] = useState<Preisreihe>()

  const vorgang = useVorgang()

  const stand = useLaden(async () => {
    const antwort = await api().Preise(printing)

    return auspacken(antwort, () => antwort)
  }, [printing, runde])

  const staende = stand.daten?.staende ?? []

  // **Die Befunde beider Wege zusammen.** Das Lesen bringt welche mit, das
  // Holen auch; sie kommen aus derselben Quelle und gehören in dieselbe Liste.
  const befunde = [...(stand.daten?.befunde ?? []), ...holbefunde]

  // **Eine Meldung und nicht zwei Bedingungen.** Der erste Anlauf fragte
  // erst, ob eine der beiden dasteht, und wählte danach noch einmal — zwei
  // Verzweigungen über dieselbe Frage, von denen eine nie beide Seiten sah.
  const meldung = holfehler ?? stand.fehler

  const holen = () => {
    void vorgang.starte(printing, async () => {
      try {
        const antwort = await api().PreiseHolen(printing)

        auspacken(antwort, () => antwort.staende)
        setzeHolfehler(undefined)
        setzeHolbefunde(antwort.befunde)
        setzeRunde((zahl) => zahl + 1)
        aufGeholt()
      } catch (ursache: unknown) {
        setzeHolfehler(meldungAus(ursache, texte))

        // **Auch ein Fehler bringt Befunde mit.** Ein Abruf, der an einer
        // Reihe scheitert, hat die anderen gelesen — und wer nur den Fehler
        // zeigte, verschwiege sie.
        setzeHolbefunde(
          ursache instanceof Bindungsfehler ? ursache.befunde : [],
        )
      }
    })
  }

  return (
    <section className="flex flex-wrap items-center gap-3">
      <span className="text-sm font-medium">{eigene.preiseTitel}</span>

      {staende.length === 0 && (
        <span className="text-sm text-leise">{eigene.keinePreise}</span>
      )}

      {staende.map((preis) => (
        <Knopf
          key={`${preis.waehrung}/${preis.finish}`}
          bezeichnung={`${eigene.verlauf} ${preis.waehrung} ${preis.finish}`}
          aufKlick={() => {
            setzeReihe((vorher: Preisreihe | undefined) =>
              vorher !== undefined &&
              reihenschluessel(vorher) === reihenschluessel(reiheVon(preis))
                ? undefined
                : reiheVon(preis),
            )
          }}
        >
          <span className="text-sm">
            <span className="font-medium">
              {geldAusText(preis.wert, preis.waehrung, darstellung)}
            </span>{' '}
            <span className="text-leise">
              ({preis.finish} ·{' '}
              {alterstext(eigene, preis.alterStunden, preis.quelle)}
              {/* **Abrufzeit ist nicht Inhaltsstand** (F04, #2246): wo die
                  Quelle den Tag nennt, steht er daneben. */}
              {preis.beobachtet !== '' &&
                ` · ${einsetzen(eigene.preisQuellstand, {
                  tag: alsKalendertag(
                    preis.beobachtet.slice(0, 10),
                    darstellung,
                  ),
                })}`}
              )
            </span>
          </span>
        </Knopf>
      ))}

      <Knopf
        bezeichnung={`${eigene.preiseHolen} ${printing}`}
        aufKlick={holen}
        angehalten={vorgang.laeuft(printing)}
      >
        {vorgang.laeuft(printing) ? eigene.preiseLaufen : eigene.preiseHolen}
      </Knopf>

      {befunde.length > 0 && (
        <div className="basis-full text-sm text-leise">
          <p>{eigene.preisbefunde}</p>
          <ul className="mt-1 space-y-1 text-xs">
            {befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </div>
      )}

      {meldung !== undefined && (
        <span role="alert" className="text-sm text-fehler">
          {meldung}
        </span>
      )}

      {reihe !== undefined && <Preisverlauf reihe={reihe} />}

      {/* **Angebote, fremder Verlauf und die Wege hinaus eingeklappt** —
          UXR-10 (#1918): zwei Abrufe ins Netz samt Erklärungen, und das
          Artboard *Sammlung* zeigt im Inspector keinen davon. Gezeichnet
          und gefragt wird erst offen. */}
      <div className="basis-full">
        <Aufklappkasten titel={eigene.marktUndVerlauf}>
          <div className="flex flex-col gap-3">
            {/* **Der `key` räumt beim Wechsel ab** (§7, Regel 1 und 2
                zugleich, docs/SCHREIBWEGE.md): ein anderer Druck baut das
                Bauteil neu, und damit verschwinden geholte Angebote samt
                einer laufenden Antwort. */}
            <Marktangebote key={printing} printing={printing} />

            {/* **Und derselbe `key` aus demselben Grund** (§7, Regel 1 und
                2). */}
            <Fremdverlauf
              key={`verlauf-${printing}`}
              printing={printing}
              aufGeholt={aufGeholt}
            />

            {/* **Wohin diese Karte außerhalb führt** (T8, #1473) — zuletzt,
                weil die Kästen darüber sagen, was sie **hier** wert ist. */}
            <Marktlinke key={`links-${printing}`} printing={printing} name="" />
          </div>
        </Aufklappkasten>
      </div>
    </section>
  )
}

// alterstext sagt, wie alt ein Preis ist.
//
// **Unter einer Stunde heißt „gerade eben"** und nicht „vor 0 h" — eine Null
// in einer Altersangabe liest sich wie ein Fehler.
// veraltetAb ist die Schwelle in Stunden: eine Woche. **Veraltet ist nicht
// null** (U8, #1419) — ein alter Preis steht als alter Preis da, und „kein
// Preis" sagt etwas anderes.
const veraltetAb = 7 * 24

function alterstext(
  eigene: { preisAlter: string; preisFrisch: string; preisVeraltet: string },
  stunden: number,
  quelle: string,
): string {
  if (stunden === 0) {
    return einsetzen(eigene.preisFrisch, { quelle })
  }

  if (stunden > veraltetAb) {
    return einsetzen(eigene.preisVeraltet, {
      tage: String(Math.floor(stunden / 24)),
      quelle,
    })
  }

  return einsetzen(eigene.preisAlter, { stunden: String(stunden), quelle })
}
