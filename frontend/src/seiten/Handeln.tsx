// **Der Bereich Handeln** — Werkstück T9 von Phase 5.6
// (#1474).
//
// # Warum es diesen Bereich gibt
//
// **Weil der Handel bis heute in einer Kontaktzeile unter „Technik" steckte.**
// Abgleich, Post und Verbuchung standen dort mit dem Satz „bis der Bereich
// kommt"; sie stehen jetzt hier. Wer handeln will, sucht nicht unter Technik.
//
// # Zwei Gebiete und nicht sieben
//
// **Weil es zwei Fragen sind**: *mit wem* und *was ist es wert*. Übersicht,
// Kandidaten, Angebote, laufende Tausche und Verlauf sind alle Fragen an
// **ein** Gegenüber — sie in eigene Laschen zu legen hieße, zwischen ihnen
// hin und her zu springen, während man eines ansieht.
//
// # Und warum das Gegenüber in einer Schublade gewählt wird
//
// **Weil die Spalte links den laufenden Handeln gehört** (UXR-15,
// #1923).
// Wählen, Aufnehmen und Bestätigen stehen hinter „Neuer Tausch …" — der
// Grund steht in `Handelsschublade.tsx`.
import { useRef, useState, type RefObject } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { Handelsliste } from '../bausteine/Handelsliste'
import { Handelsraum } from '../bausteine/Handelsraum'
import { Knopf } from '../bausteine/Knopf'
import { Tauschpartnerleiste } from '../bausteine/Tauschpartnerleiste'
import { Laschen } from '../bausteine/Laschen'
import { Marktkonto } from '../bausteine/Marktkonto'
import { Portfolio } from '../bausteine/Portfolio'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Verdeckung, useVerdeckt } from '../verdeckung'
import { Handelsschublade } from './Handelsschublade'

import type { Handelspartner } from '../apihandelsverzeichnis'

interface Handelneigenschaften {
  // bestandsstand zählt, was ein **anderer** Prozess geschrieben hat — die
  // Liste der Handel lädt dann nach (TRADE-01.AK-04, #1715).
  bestandsstand?: number
}

// Aufschlag ist, was die Seite aufgeschlagen hat: ein Gegenüber **mit
// Namen** (UXR-15, #1923) — der Raum sagt „Wartet auf Turfer" und nicht „auf
// dein Gegenüber" — und der Handel, mit dem er aufging.
interface Aufschlag {
  partner: Handelspartner
  vorgang: string
  archiviert: boolean

  // neu zählt die Wahlen in der Schublade — **jede beginnt einen neuen
  // Entwurf** (#2391). Ohne sie sah eine zweite Wahl desselben Gegenübers
  // aus wie die erste, der Raum blieb stehen, und in ihm der angebotene
  // Handel, den er inzwischen gewählt hatte.
  neu?: number
}

export function Handeln({ bestandsstand = 0 }: Handelneigenschaften) {
  const eigene = useTexte().seiten.handeln

  // **Hier und nicht im Gebiet**, weil die Schublade über der ganzen Seite
  // steht: sie sperrt Kopf und Laschen mit (`inert`, `Verdeckung`), und sie
  // selbst steht **außerhalb** der Verdeckung — ein Hinweis in ihr soll sein
  // Escape behalten (`Tooltip`). Dieselbe Gestalt wie bei Paketen.
  const [schublade, setzeSchublade] = useState(false)
  const [aufschlag, setzeAufschlag] = useState<Aufschlag>()
  const ausloeser = useRef<HTMLButtonElement>(null)
  const verdeckt = useVerdeckt() || schublade

  return (
    <>
      <Verdeckung value={verdeckt}>
        <div className="flex flex-col gap-4" inert={schublade}>
          <Seitenkopf titel={eigene.titel} erklaerung={eigene.erklaerung} />

          <Laschen
            bezeichnung={eigene.gebiete}
            gebiete={[
              {
                schluessel: 'partner',
                name: eigene.gebietPartner,
                inhalt: (
                  <Partnergebiet
                    bestandsstand={bestandsstand}
                    aufschlag={aufschlag}
                    schlageAuf={setzeAufschlag}
                    ausloeser={ausloeser}
                    oeffneSchublade={() => {
                      // **Der Fokus geht vorher auf den Knopf**, damit er
                      // nach dem Schließen dorthin zurückkehrt — in WebKit
                      // gibt ein Klick keinen Fokus.
                      ausloeser.current?.focus()
                      setzeSchublade(true)
                    }}
                  />
                ),
              },
              {
                schluessel: 'markt',
                name: eigene.gebietMarkt,
                inhalt: <Marktgebiet />,
              },
            ]}
          />
        </div>
      </Verdeckung>

      <Handelsschublade
        offen={schublade}
        schliessen={() => {
          setzeSchublade(false)
        }}
        waehlen={(kennung, name) => {
          setzeAufschlag((bisher) => ({
            partner: { kennung, name },
            vorgang: '',
            archiviert: false,
            neu: (bisher?.neu ?? 0) + 1,
          }))
        }}
      />
    </>
  )
}

interface Partnergebietseigenschaften {
  bestandsstand: number
  aufschlag: Aufschlag | undefined
  schlageAuf: (aufschlag: Aufschlag) => void

  // ausloeser ist „Neuer Tausch …" — die Seite gibt ihm den Fokus zurück.
  ausloeser: RefObject<HTMLButtonElement | null>
  oeffneSchublade: () => void
}

// Partnergebiet ist die Liste der Handel und der Raum dazu.
//
// **Der `key` am Raum räumt beim Wechsel alles mit** (§7, Regel 1 und 2 in
// einer Gestalt): der eingesetzte Fremdstand, das Ergebnis des Abgleichs, der
// eingetragene Vorgang und der Buchungsstand gehören zu **einem** Gegenüber.
// Ohne ihn stünde der halb ausgefüllte Stand des vorigen über dem neuen.
function Partnergebiet({
  bestandsstand,
  aufschlag,
  schlageAuf,
  ausloeser,
  oeffneSchublade,
}: Partnergebietseigenschaften) {
  const eigene = useTexte().seiten.handeln

  // **Jede Änderung an einem Vorgang lädt die Liste nach** (TRADE-01, #1715).
  const [listenstand, setzeListenstand] = useState(0)
  const [ohneIdentitaet, setzeOhneIdentitaet] = useState(false)
  const listeNachladen = () => {
    setzeListenstand((bisher) => bisher + 1)
  }

  // **Eine mitgereiste Karte wartet sichtbar** — UX-03.AK-01
  // (#1675).
  //
  // **Ein Angebot braucht ein Gegenüber.** Solange keines gewählt ist, gibt
  // es keine Zeile, in die der Druck fiele — und ein Druck, der irgendwo
  // wartet, ohne dass jemand es sieht, ist ein stiller Zustand (§6).
  const anschluss = useAnschluss()
  // **Ein Wunsch aus einem Deck wartet genauso** (UX-03.AK-01): auch er
  // braucht ein Gegenüber.
  const wartend =
    anschluss.offen?.art === 'handel' || anschluss.offen?.art === 'tauschwunsch'
      ? anschluss.offen
      : undefined

  // **Aus der Befehlspalette: „Karten tauschen mit …"** — das Gegenüber
  // steht aufgeschlagen da (UX-02.AK-02, #1674).
  useAufgenommen(
    anschluss.offen?.art === 'gegenueber' ? anschluss.offen : undefined,
    (gewuenscht) => {
      schlageAuf({
        partner: { kennung: gewuenscht.kennung, name: gewuenscht.name },
        vorgang: '',
        archiviert: false,
      })
    },
  )

  // **Nebeneinander erst, wenn der Platz reicht** — gemessen an der Seite
  // (`@container`) und nicht am Fenster (QA01, #2382): bei sehr großer
  // Schrift nahm die Liste ihre 20rem, und dem Handel blieben bei 1024
  // Punkten 173. 42rem sind die Liste, der Abstand und ein Handel, der
  // nicht überläuft; bei normaler Schrift steht es bei 1024 wie bisher.
  return (
    <div className="@container">
      <div className="flex flex-col gap-4 @2xl:flex-row">
        <div className="flex flex-col gap-4 @2xl:w-80 @2xl:shrink-0">
          {/* **Oben und nicht am Fuß der Spalte** wie im Artboard: die Seite
            rollt als Ganzes, und am Fuß einer langen Liste stünde die
            führende Handlung unter der Falz (dieselbe Abweichung wie
            „Karte anlegen", UXR-05). */}
          <div>
            {/* **Ohne Identität führt sie und nicht der Tausch** (#2027, R1):
              ein Tausch ohne Identität käme nicht weit, und der Leerzustand
              der Liste nennt den Schritt davor. */}
            <Knopf
              gewicht={ohneIdentitaet ? 'neben' : 'haupt'}
              griff={ausloeser}
              aufKlick={oeffneSchublade}
            >
              {eigene.neuerTausch}
            </Knopf>
            {/* **Und daneben jedes Gegenüber mit einem Klick** (P12, #2380,
              H1): jede Wahl beginnt einen neuen Entwurf wie die Schublade. */}
            <Tauschpartnerleiste
              waehlen={(kennung, name) => {
                schlageAuf({
                  partner: { kennung, name },
                  vorgang: '',
                  archiviert: false,
                  neu: (aufschlag?.neu ?? 0) + 1,
                })
              }}
            />
          </div>

          {/* **Von selbst nur, solange nichts aufgeschlagen ist** (P13,
            H4): was ein Mensch geöffnet hat, auch einen neuen Entwurf,
            bleibt. */}
          <Handelsliste
            ohneIdentitaet={setzeOhneIdentitaet}
            gewaehlt={aufschlag?.vorgang ?? ''}
            stand={listenstand + bestandsstand}
            oeffnen={(partner, vorgang, archiviert) => {
              schlageAuf({ partner, vorgang, archiviert })
            }}
            vonSelbst={
              aufschlag === undefined
                ? (partner, vorgang) => {
                    schlageAuf({ partner, vorgang, archiviert: false })
                  }
                : undefined
            }
          />
        </div>

        <div className="min-w-0 grow">
          {aufschlag === undefined ? (
            <div className="flex flex-col gap-2">
              <p className="text-sm text-leise">{eigene.partnerKeiner}</p>

              {wartend !== undefined && (
                <p className="text-sm">
                  {einsetzen(eigene.anschlussOhnePartner, {
                    name: wartend.name,
                  })}
                </p>
              )}
            </div>
          ) : (
            <Handelsraum
              key={`${aufschlag.partner.kennung}/${aufschlag.vorgang}/${String(aufschlag.neu ?? 0)}`}
              kennung={aufschlag.partner.kennung}
              name={aufschlag.partner.name}
              vorgang={aufschlag.vorgang}
              archiviert={aufschlag.archiviert}
              liste={listeNachladen}
              bestandsstand={bestandsstand}
            />
          )}
        </div>
      </div>
    </div>
  )
}

// Marktgebiet zeigt, was die eigene Sammlung wert ist.
//
// **Dieselbe Ansicht wie in der Sammlung und keine zweite** — eine zweite
// Rechnung über dieselben Zahlen liefe irgendwann auseinander. Was hier
// dazukommt, ist der Satz darüber: eine Beobachtung ist keine Zusage.
function Marktgebiet() {
  const eigene = useTexte().seiten.handeln

  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-base font-semibold">{eigene.marktTitel}</h2>

      <Erklaerungsabsatz
        text={eigene.marktErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <Portfolio runde={0} />

      <Marktkonto />
    </section>
  )
}
