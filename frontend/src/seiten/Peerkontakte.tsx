// **Wen dieses Profil kennt** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum die Liste erst auf Nachfrage kommt
//
// **Weil sie eine Frage an den Bestand ist, die niemand gestellt hat.**
// Dieselbe Gestalt wie bei der Identität nebenan, und aus demselben Grund:
// wer die Technikseite aufschlägt, will meist etwas anderes.
//
// # Und warum nur eine Zeile offen steht
//
// **UXR-16** (#1924):
// jede Zeile war ein offenes Formular, und jedes wartende Gegenüber trug sein
// „Stimmt überein" — mit acht Gegenübern vier Hauptknöpfe in einer Region
// (R1 in `docs/UIREVIEW.md`). **Zu Anfang ist die erste offen, die auf einen
// Vergleich wartet**, wie im Artboard *Gegenüber*, das neben der Liste genau
// einen bestätigen lässt; gibt es keine und nur ein Gegenüber, dieses. **Wer
// eine Zeile öffnet oder an ihr etwas tut, hält sie offen** — sonst spränge
// nach dem Bestätigen die nächste Wartende an ihre Stelle, und die Zwecke
// des eben Bestätigten wären fort.
//
// # Und warum nach jeder Änderung die ganze Liste kommt
//
// **Weil eine Änderung mehr als eine Zeile trifft.** Eine Sperre räumt die
// Rechte mit; eine Rotation nebenan setzt **jeden** Kontakt zurück. Wer nur
// die geänderte Zeile einträgt, ließe die übrigen mit einem Stand stehen, den
// es nicht mehr gibt.
import { useEffect, useId, useRef, useState } from 'react'

import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Einladungstausch } from '../bausteine/Einladungstausch'
import { Faltliste } from '../bausteine/Faltliste'
import { Knopf } from '../bausteine/Knopf'
import { Kontaktaufnahme } from '../bausteine/Kontaktaufnahme'
import { Nachbarn } from '../bausteine/Nachbarn'
import { Peerkontaktzeile } from '../bausteine/Peerkontaktzeile'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { Peerkontakt, Peerkontakteantwort } from '../apipeerkontakt'

interface Peerkontakteeigenschaften {
  // eingebettet heißt: in einem Kasten, den jemand eben aufgeklappt hat —
  // dann **ist** er die Nachfrage, und die Liste kommt sofort (UX-03.AK-03,
  // #1675). Die Überschrift trägt der Kasten.
  eingebettet?: boolean | undefined

  // aufGeaendert meldet eine geglückte Änderung — die Partnerwahl daneben
  // lädt dann nach, damit ein eben bestätigtes Gegenüber zur Wahl steht.
  aufGeaendert?: (() => void) | undefined

  // nachladen steigt, wenn daneben geschrieben wurde — unter
  // *Verbindungen* nimmt die Suche im Netz jemanden auf (§7, Regel 3).
  //
  // **Ein Zeichen und keine Abhängigkeit des Ladehakens**: `useLaden` hielte
  // jede Änderung für ein neues Bezugsobjekt und räumte die Liste samt allem,
  // was unter ihr hängt (siehe unten, NET-01).
  nachladen?: number | undefined
}

export function Peerkontakte({
  eingebettet = false,
  aufGeaendert,
  nachladen = 0,
}: Peerkontakteeigenschaften = {}) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [gefragt, setzeGefragt] = useState(eingebettet)
  const [fehler, setzeFehler] = useState<string>()

  // gewaehlt ist die Zeile, die jemand geöffnet hat — `null` heißt: keine;
  // `undefined` heißt: niemand hat gewählt, und die Vorgabe gilt.
  const [gewaehlt, setzeGewaehlt] = useState<string | null>()

  // **Über den Ladehaken** (§7): er führt den Lauf mit und trägt eine Antwort
  // nur ein, wenn sie noch zum aktuellen Stand gehört.
  //
  // **Nachgeladen wird über `nachladen` und nicht über eine Abhängigkeit.**
  // Bis NET-01 (#1705) stand hier ein Zähler `geaendert` darin; `useLaden`
  // hielt jede Änderung damit für ein neues Bezugsobjekt und räumte die Liste
  // samt allem, was unter ihr hing — die Meldung einer eben angenommenen
  // Einladung verschwand, gefunden an der Hüllenprobe `net01`.
  const liste = useLaden<Peerkontakteantwort | undefined>(async () => {
    if (!gefragt) {
      return undefined
    }

    const antwort = await api().Peerkontakte()

    return auspacken(antwort, () => antwort)
  }, [gefragt])

  const listeNachladen = liste.nachladen
  const gesehen = useRef(nachladen)

  useEffect(() => {
    if (gesehen.current === nachladen) {
      return
    }

    gesehen.current = nachladen
    listeNachladen()
  }, [nachladen, listeNachladen])

  // handeln fährt einen der acht schreibenden Wege.
  //
  // **Ein Vorgang für alle acht** (§7, Regel 4): sie gehen an dieselbe
  // Liste, und zwei Riegel ließen zwei zugleich laufen — gegen einen Stand,
  // den das Backend noch nicht kennt.
  const handeln = (weg: () => Promise<Peerkontakteantwort>): Promise<void> =>
    vorgang.starte('peerkontakte', async () => {
      setzeFehler(undefined)

      try {
        auspacken(await weg(), () => undefined)

        // **Geschrieben ist geschrieben** (§7, Regel 3): nachgeladen wird
        // erst nach dem geglückten Schreiben, und ein Fehler dabei nimmt die
        // Änderung nicht zurück.
        liste.nachladen()
        aufGeaendert?.()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const laeuft = vorgang.laeuft('peerkontakte')

  // amKontakt fährt einen Weg an einer Zeile und hält sie dabei offen.
  const amKontakt = (
    kennung: string,
    weg: () => Promise<Peerkontakteantwort>,
  ) => {
    setzeGewaehlt(kennung)
    void handeln(weg)
  }

  // **Einmal festgehalten und nicht zweimal gefragt.** Ein zweiter Griff
  // (`liste.daten?.zwecke ?? []`) trüge einen Rückfall, den es nicht gibt —
  // und ein Zweig, den kein Lauf betreten kann, ist ein Zweig, den niemand
  // prüft (§10).
  const daten = liste.daten
  const offen =
    gewaehlt === undefined ? vorgabe(daten?.kontakte ?? []) : gewaehlt

  return (
    <section
      aria-labelledby={eingebettet ? undefined : ueberschrift}
      aria-label={eingebettet ? eigene.kontakteTitel : undefined}
      className="flex flex-col gap-3"
    >
      {!eingebettet && (
        <h3 id={ueberschrift} className="text-lg font-semibold">
          {eigene.kontakteTitel}
        </h3>
      )}

      {!eingebettet && (
        <Erklaerungsabsatz
          text={eigene.kontakteErklaerung}
          klasse="max-w-prose text-sm text-leise"
        />
      )}

      {!gefragt && (
        <Knopf
          gewicht="leise"
          aufKlick={() => {
            setzeGefragt(true)
          }}
        >
          {eigene.kontakteZeigen}
        </Knopf>
      )}

      {liste.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {liste.fehler}
        </p>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {/* **Wer in der Nähe ist, steht im Kasten vorn** — UXR-16 (#1924),
          nach dem Artboard *Gegenüber*: „In deiner Nähe" führt, und „Weiter
          weg? Dann tauscht ihr einen Einladungscode aus" steht darunter. Bis
          dahin gab es die Suche im Netz nur unter *Einstellungen ›
          Verbindungen*; wer beim Handeln jemanden am selben Tisch aufnehmen
          wollte, musste dorthin und zurück. **Gefragt wird weiter erst auf
          Knopfdruck** — ein Ruf ins Netz, den niemand angestoßen hat, sagte
          jedem dort, dass hier jemand sucht. In den Einstellungen steht die
          Suche weiter als eigener Abschnitt (`Nachbarschaft`). */}
      {eingebettet && (
        <section aria-label={eigene.nachbarnTitel} className="flex flex-col">
          <Nachbarn
            aufGeaendert={(kennung) => {
              setzeGewaehlt(kennung)
              liste.nachladen()
              aufGeaendert?.()
            }}
          />
        </section>
      )}

      {/* **Einladen zuerst, die Kennung von Hand danach** (NET-01.AK-01,
          #1705): der Austausch zweier Codes ist die direkte Handlung, die
          rohe Kennung eine technische Angabe — sie steht im Kasten darunter
          und nicht mehr vorn. */}
      {daten !== undefined && (
        <Einladungstausch
          zwecke={daten.zwecke}
          nachAufnahme={() => {
            liste.nachladen()
            aufGeaendert?.()
          }}
        />
      )}

      {daten !== undefined && (
        <Aufklappkasten titel={eigene.einladung.vonHand}>
          <Kontaktaufnahme
            laeuft={laeuft}
            aufnehmen={(kennung, name) => {
              void handeln(() => api().PeerkontaktAufnehmen(kennung, name))
            }}
          />
        </Aufklappkasten>
      )}

      {/* **Im Kasten steht die Erklärung an der Liste** (UXR-16, #1924):
          obenan erklärte sie etwas, das erst unter Suche und Einladung kam.
          In den Einstellungen steht sie unter der Überschrift. */}
      {eingebettet && daten !== undefined && (
        <Erklaerungsabsatz
          text={eigene.kontakteErklaerung}
          klasse="max-w-prose text-sm text-leise"
        />
      )}

      {daten?.kontakte.length === 0 && (
        <p className="max-w-prose text-sm text-leise">{eigene.kontakteLeer}</p>
      )}

      {daten !== undefined && daten.kontakte.length > 0 && (
        // **Vorn fünf, die offene Zeile immer** (#2035): mit acht Gegenübern
        // riss *Einstellungen › Verbindungen* am Fenster R2 und R4. Ein eben
        // aufgenommener Fund steht hinten und schlägt offen auf — er bleibt
        // im Bild.
        <Faltliste
          eintraege={daten.kontakte}
          vorn={5}
          schritt={10}
          immer={(kontakt) => kontakt.kennung === offen}
          klasse="flex flex-col gap-2"
          zeile={(kontakt) => (
            <Peerkontaktzeile
              key={kontakt.kennung}
              kontakt={kontakt}
              zwecke={daten.zwecke}
              laeuft={laeuft}
              offen={offen === kontakt.kennung}
              umschalten={() => {
                setzeGewaehlt(
                  offen === kontakt.kennung ? null : kontakt.kennung,
                )
              }}
              bestaetigen={(kennung, verglichen) => {
                amKontakt(kennung, () =>
                  api().PeerkontaktBestaetigen(kennung, verglichen),
                )
              }}
              sperren={(kennung) => {
                amKontakt(kennung, () => api().PeerkontaktSperren(kennung))
              }}
              entsperren={(kennung) => {
                amKontakt(kennung, () => api().PeerkontaktEntsperren(kennung))
              }}
              erlauben={(kennung, zweck) => {
                amKontakt(kennung, () =>
                  api().PeerzweckErlauben(kennung, zweck),
                )
              }}
              entziehen={(kennung, zweck) => {
                amKontakt(kennung, () =>
                  api().PeerzweckEntziehen(kennung, zweck),
                )
              }}
              adresseSetzen={(kennung, adresse) => {
                amKontakt(kennung, () =>
                  api().PeerkontaktAdresse(kennung, adresse),
                )
              }}
              notizSetzen={(kennung, notiz) => {
                amKontakt(kennung, () => api().PeerkontaktNotiz(kennung, notiz))
              }}
            />
          )}
        />
      )}
    </section>
  )
}

// vorgabe ist die Zeile, die offen steht, solange niemand gewählt hat: die
// erste, die auf einen Vergleich wartet — gibt es keine und nur ein
// Gegenüber, dieses. Eine einzelne Zeile zuzuklappen hieße, den einzigen
// Inhalt hinter einen Klick zu legen.
function vorgabe(kontakte: readonly Peerkontakt[]): string | undefined {
  return kontakte.find(
    (kontakt) => kontakt.stand === 'unbestaetigt' || kontakte.length === 1,
  )?.kennung
}
