// **Das eigene Konto bei einem Marktplatz** — Werkstück T11 von Phase 5.6
// (#1476).
//
// # Warum es erst auf Nachfrage kommt
//
// **Weil es eine Frage nach draußen ist.** Wer den Bereich *Handeln*
// aufschlägt, will meist etwas anderes; ein Abruf beim Zeichnen ginge zu
// einem fremden Dienst, ohne dass jemand danach gefragt hat (§16, Regel 5).
//
// # Und warum hier nichts angeboten und nichts gekauft wird
//
// **Weil dieses Programm dort nichts schreibt.** Der Anbieter kennt die
// Wege; gebaut ist keiner
// ([0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md)).
// Was hier steht, ist eine Übersicht — und an jeder Bestellung, ob sie nach
// der Regel des Anbieters ein Stück aus dem eigenen Lager genommen hat.
// **Gebucht wird daraus nichts.**
//
// # Und warum je Liste zehn Zeilen stehen
//
// **Weil ein Händlerkonto Hunderte Angebote hat** (UXR-15,
// #1923;
// R4 in `docs/UIREVIEW.md`): jedes stand als eigene Zeile untereinander.
// Jetzt zehn je Liste und der Rest stückweise (`Faltliste`), und jede Liste
// trägt ihren Namen — vorher standen zwei unbenannte Listen nebeneinander.
import { useState } from 'react'

import { Faltliste } from './Faltliste'
import { Knopf } from './Knopf'
import { Satz } from './Satz'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { derMarktplatz } from '../apimarktkonto'
import { meldungAus } from '../meldung'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { alsGeld } from '../geld'
import { useVorgang } from '../haken/useVorgang'

import type { Kontostand } from '../apimarktkonto'

export function Marktkonto() {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const lauf = useVorgang()

  const [fehler, setzeFehler] = useState<string>()
  const [stand, setzeStand] = useState<Kontostand | null>(null)

  const kennzeichen = 'marktkonto'

  const nachsehen = (): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().Marktkonto(derMarktplatz, 0, 0)

        setzeStand(auspacken(antwort, () => antwort.stand))
      } catch (ursache) {
        setzeStand(null)
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-base font-semibold">{eigene.kontoTitel}</h2>

      <p className="max-w-prose text-sm text-leise">
        <Satz text={eigene.kontoErklaerung} />
      </p>

      <div>
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void nachsehen()
          }}
        >
          {eigene.kontoNachsehen}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {stand !== null && <Kontosicht stand={stand} />}
    </section>
  )
}

// dieZeilenVorn und derSchritt sind, wie viele Zeilen je Liste zu Anfang
// dastehen und wie viele ein Klick dazu zeigt.
const dieZeilenVorn = 10
const derSchritt = 20

// Kontosicht zeigt, was zurückkam.
function Kontosicht({ stand }: { stand: Kontostand }) {
  const eigene = useTexte().seiten.handeln
  const darstellung = useDarstellung()

  return (
    <div className="flex flex-col gap-2 text-sm">
      <p role="status" className="text-leise">
        {stand.name} · {eigene.kontoAngebote}: {stand.angebote.length} ·{' '}
        {eigene.kontoBestellungen}: {stand.bestellungen.length}
      </p>

      {stand.befunde.map((befund) => (
        <p key={befund} className="text-xs text-fehler">
          {befund}
        </p>
      ))}

      <Faltliste
        eintraege={stand.angebote}
        vorn={dieZeilenVorn}
        schritt={derSchritt}
        bezeichnung={eigene.kontoAngebote}
        klasse="flex flex-col gap-0.5 font-mono text-xs"
        zeile={(angebot) => (
          <li key={angebot.id}>
            {angebot.menge} × {angebot.name} ·{' '}
            {alsGeld(angebot.preisCent, angebot.waehrung, darstellung)}
          </li>
        )}
      />

      <Faltliste
        eintraege={stand.bestellungen}
        vorn={dieZeilenVorn}
        schritt={derSchritt}
        bezeichnung={eigene.kontoBestellungen}
        klasse="flex flex-col gap-0.5 font-mono text-xs"
        zeile={(bestellung) => (
          <li key={bestellung.id}>
            {bestellung.code} · {bestellung.zustand} ·{' '}
            {bestellung.bewegtBestand
              ? eigene.kontoBewegt
              : eigene.kontoBewegtNicht}
          </li>
        )}
      />
    </div>
  )
}
