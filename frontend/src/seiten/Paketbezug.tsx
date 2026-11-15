// Eine Vorlage beziehen — der Inhalt der Schublade *Paket beziehen*.
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
// Herausgelöst aus `Pakete.tsx`, als der Bezug mit UXR-17
// (#1925)
// von der Seite in die Schublade wanderte; das Verhalten ist dasselbe.
//
// # Warum zwei Schritte
//
// Ein Programm, das ungefragt über ein Gigabyte lädt, überrascht seine
// Nutzenden (#56) — und wer nicht weiß, woher eine Vorlage kommt, kann ihre
// Rechtslage nicht einschätzen (#139, §11). Deshalb zeigt die Schublade
// **zuerst** die Herkunft und lädt erst nach einer Zustimmung.
import { useState } from 'react'

import {
  api,
  auspacken,
  type BezugAntwort,
  type HerkunftAntwort,
  type QuelleDTO,
} from '../api'
import { useLauf } from '../haken/useLauf'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Bezugsquellen } from './Bezugsquellen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { kennungseingabe } from '../bausteine/kennungseingabe'

interface Eingaben {
  adresse: string
  stand: string
  unterverzeichnis: string
  name: string
}

// **Das Formular beginnt leer** —
// #263.
//
// Eine Vorgabe — etwa Full-Magic-Pack, ein Commit und
// `unterverzeichnis: 'data'` — hätte zwei Mängel, und der zweite wäre still:
//
//  1. **Sie erklärte sich nicht.** Wer die Anwendung zum ersten Mal öffnet,
//     sieht eine Adresse, weiß aber nicht, was dahinter liegt, wie groß es ist
//     und ob es das ist, was er braucht.
//  2. **`data` ließe die Schriften zurück.** Die Packs tragen ihre Vorlagen
//     unter `data` und ihre Schriften daneben (`Magic - Fonts`, im
//     installierten M15-Pack 131 Dateien, 26 MB). Wer nur `data` holt,
//     bekommt Vorlagen ohne Schriften — und jede Karte zeigt „Schrift fehlt",
//     ohne dass jemand sähe, warum.
//
// Beides beantwortet die Quellenliste darüber
// (`internal/bezug/quellen.go`): vier Sammlungen mit Namen, Größe und einem
// Satz dazu, und ein Klick trägt sie hier ein.
const leer: Eingaben = {
  adresse: '',
  stand: '',
  unterverzeichnis: '',
  name: '',
}

const knopfklassen =
  'rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60'
const pruefknopf = 'mt-3 ' + knopfklassen

export function Paketbezug() {
  const texte = useTexte()
  const vorgang = useVorgang()

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2).
  // Wer die Adresse ändert, während die Herkunft unterwegs ist, bekäme sie
  // sonst nachträglich eingetragen — zu Feldern, die sie nie beschrieben hat.
  const lauf = useLauf()

  const [eingaben, setzeEingaben] = useState<Eingaben>(leer)

  // **Was bestätigt wurde, steht hier — nicht in den Eingabefeldern.**
  //
  // Der Zweitblick über Phase 9 hat es gefunden: nach dem Hinweis blieben die
  // Felder änderbar, und »Beziehen« nahm die aktuellen Werte. Wer die Herkunft
  // eines Repositoriums liest und danach eine andere Adresse eintippt, bezöge
  // etwas, dem er nie zugestimmt hat (§7, Regel 2: eine Antwort gehört zu dem
  // Stand, aus dem sie stammt).
  //
  // **Ein Wechsel räumt den offenen Kontext** (§7, Regel 1): sobald jemand ein
  // Feld ändert, ist die Zustimmung hinfällig und der Hinweis verschwindet.
  const [bestaetigt, setzeBestaetigt] = useState<Eingaben>()

  // **Der Satz entsteht hier, aus Angaben von dort** (§9). Das Backend gibt
  // Quelle und Stand; der Text dazu steht in den Sprachkatalogen — sonst stünde
  // in der englischen Oberfläche ein deutscher Absatz.
  const [herkunft, setzeHerkunft] = useState<HerkunftAntwort>()
  const [ergebnis, setzeErgebnis] = useState<BezugAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const vollstaendig = eingaben.adresse !== '' && eingaben.stand !== ''

  // **Nicht abgeschaltet, sondern geprüft.** Eine ausgegraute Schaltfläche ist
  // per Tastatur nicht erreichbar — sie fällt aus der Reihenfolge, und wer sie
  // ansteuert, erfährt nicht, warum nichts geschieht. Sie bleibt bedienbar und
  // sagt beim Klick, was fehlt (§6).
  const herkunftZeigen = async () => {
    if (!vollstaendig) {
      setzeFehler(texte.seiten.pakete.bezugUnvollstaendig)

      return
    }

    const stand = lauf.naechster()
    const gefragt = eingaben

    await vorgang.starte('hinweis', async () => {
      setzeFehler(undefined)
      setzeErgebnis(undefined)

      try {
        const gezeigt = await api().PaketHerkunft(
          gefragt.adresse,
          gefragt.stand,
        )

        auspacken(gezeigt, () => undefined)

        if (!lauf.istAktuell(stand)) {
          return
        }

        setzeHerkunft(gezeigt)
        setzeBestaetigt(gefragt)
      } catch (ursache: unknown) {
        if (lauf.istAktuell(stand)) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  const beziehen = async (gewaehlt: Eingaben) => {
    await vorgang.starte('bezug', async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().PaketBeziehen(
          gewaehlt.adresse,
          gewaehlt.stand,
          gewaehlt.unterverzeichnis,
          gewaehlt.name,
        )

        auspacken(antwort, () => undefined)

        setzeErgebnis(antwort)
        setzeHerkunft(undefined)
        setzeBestaetigt(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const feld = (schluessel: keyof Eingaben, beschriftung: string) => (
    <label className="flex flex-col gap-1">
      <span className="text-sm">{beschriftung}</span>
      <input
        {...kennungseingabe}
        type="text"
        // **Alle vier sind Kennungen** — Adresse, Stand, Verzeichnis, Name;
        // die Rechtschreibkorrektur von macOS machte daraus Wörter (#2389).
        value={eingaben[schluessel]}
        onChange={(ereignis) => {
          // **Über den vorigen Stand, nicht über den der Auszeichnung**: zwei
          // Änderungen kurz hintereinander gingen sonst gegen denselben alten
          // Wert, und die erste wäre verloren (§7, Regel 2 in klein).
          const wert = ereignis.target.value
          setzeEingaben((alt) => ({ ...alt, [schluessel]: wert }))

          // Der Wechsel räumt Auskunft und Zustimmung mit (§7, Regel 1) — und
          // zählt den Lauf weiter, damit eine unterwegs befindliche Antwort
          // sie nicht wieder einträgt (Regel 2).
          lauf.naechster()
          setzeHerkunft(undefined)
          setzeBestaetigt(undefined)
        }}
        className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      />
    </label>
  )

  return (
    <div className="flex flex-col">
      <Erklaerungsabsatz
        text={texte.seiten.pakete.bezugErklaerung}
        klasse="max-w-prose text-leise"
      />

      {/* **Die Quellen stehen über dem Formular**, nicht daneben: sie sind
          der erste Weg, das Formular der zweite. Wer eine eigene Adresse
          hat, tippt sie darunter. */}
      <Bezugsquellen
        uebernehmen={(quelle: QuelleDTO) => {
          setzeEingaben({
            adresse: quelle.adresse,
            stand: quelle.stand,
            unterverzeichnis: quelle.unterverzeichnis,
            name: quelle.name,
          })

          // **Der Wechsel räumt Auskunft und Zustimmung mit** (§7, Regel 1)
          // und zählt den Lauf weiter (Regel 2) — dieselbe Behandlung wie
          // eine Änderung von Hand. Eine Zustimmung, die zu einer anderen
          // Adresse gehört, wäre die schlimmste Art, sie zu übergehen.
          lauf.naechster()
          setzeHerkunft(undefined)
          setzeBestaetigt(undefined)
          setzeFehler(undefined)
          setzeErgebnis(undefined)
        }}
      />

      <div className="mt-4 grid max-w-xl gap-3">
        {feld('adresse', texte.seiten.pakete.bezugAdresse)}
        {feld('stand', texte.seiten.pakete.bezugStand)}
        {feld('unterverzeichnis', texte.seiten.pakete.bezugUnterverzeichnis)}
        {feld('name', texte.seiten.pakete.bezugName)}
      </div>

      <div>
        <button
          type="button"
          disabled={vorgang.laeuft('hinweis')}
          onClick={() => {
            void herkunftZeigen()
          }}
          className={pruefknopf}
        >
          {texte.seiten.pakete.bezugPruefen}
        </button>
      </div>

      {herkunft !== undefined && bestaetigt !== undefined && (
        <div className="mt-4 max-w-prose rounded-sm border border-kante p-3">
          <p className="text-text">
            {einsetzen(texte.seiten.pakete.bezugHerkunft, {
              quelle: herkunft.quelle,
              stand: herkunft.stand,
            })}
          </p>

          <Erklaerungsabsatz
            text={texte.seiten.pakete.bezugLizenz}
            klasse="mt-2 text-leise"
          />

          <div className="mt-3 flex gap-3">
            <button
              type="button"
              disabled={vorgang.laeuft('bezug')}
              onClick={() => {
                void beziehen(bestaetigt)
              }}
              className={knopfklassen}
            >
              {texte.seiten.pakete.bezugStarten}
            </button>

            <button
              type="button"
              onClick={() => {
                setzeHerkunft(undefined)
                setzeBestaetigt(undefined)
              }}
              className={knopfklassen}
            >
              {texte.seiten.pakete.bezugAbbrechen}
            </button>
          </div>
        </div>
      )}

      {fehler !== undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {ergebnis && (
        <div role="status" className="mt-4">
          <p className="text-leise">
            {`${texte.seiten.pakete.bezugGeholt} ${ergebnis.ort} — ${texte.seiten.pakete.bezugPruefsumme} ${ergebnis.pruefsumme}`}
          </p>

          {ergebnis.uebergangen.length > 0 && (
            <>
              <p className="mt-2 max-w-prose text-leise">
                {texte.seiten.pakete.bezugUebergangen}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {ergebnis.uebergangen.map((datei) => (
                  <li key={datei}>{datei}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </div>
  )
}
