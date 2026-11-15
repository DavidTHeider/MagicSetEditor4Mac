// **Eine Kartenliste im Fenster einlesen** — Werkstück 3 von Phase 4.5
// (#1292).
//
// # Warum zwei Schritte und nicht einer
//
// **§16, Regel 6: „Nichts wird still übernommen."** Der erste Knopf ordnet zu
// und schreibt nichts; der zweite nimmt auf, was entschieden ist.
//
// Der Grund ist gemessen: „Counterspell“ ohne Setangabe passt bei Scryfall auf
// **72** Drucke (nachgelesen am 2026-08-30). Einen davon still einzutragen
// füllte eine Sammlung mit Papier, das niemand besitzt.
//
// # Warum ein eigener Kasten und keine zweite Seite
//
// Weil es zur Sammlung gehört und dort hin will: wer eine Liste einliest,
// sieht das Ergebnis in derselben Ansicht. Eine eigene Seite verlangte einen
// Wechsel und damit einen Kontext, der beim Zurückkommen weg wäre (§7,
// Regel 1).
import { useId, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { Deutungsbilanz, Deutungszeile, entschieden } from './Deutungszeile'
import { useTexte } from '../sprache/kontext'
import { useLauf } from '../haken/useLauf'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Einfuhrvorschau, braucheVorschau } from './Einfuhrvorschau'

import type {
  Einfuhrbericht,
  Listendeutung,
  Listeneinfuhrzeile,
  SammelvorschauDTO,
  Zusammenfuehrung,
} from '../antwortenlisteneinfuhr'

export interface ListeneinfuhrEigenschaften {
  // aufAufgenommen läuft, nachdem etwas geschrieben wurde.
  //
  // **Damit die Liste darunter nachlädt** — Bestenfalls-Arbeit (§7, Regel 3):
  // ein Fehler dabei macht die Einfuhr nicht rückgängig.
  aufAufgenommen: () => void
}

export function Listeneinfuhr({ aufAufgenommen }: ListeneinfuhrEigenschaften) {
  const eigene = useTexte().seiten.sammlung

  const feldID = useId()

  const [text, setzeText] = useState('')
  // **`null` und nicht `undefined`.** Die Naht gibt `Listendeutung | null`;
  // ein zweites Nichts daneben wäre eine Unterscheidung ohne Fall — und ein
  // Zweig ohne Fall gehört weg (§3).
  const [deutung, setzeDeutung] = useState<Listendeutung | null>(null)
  const [bericht, setzeBericht] = useState<Einfuhrbericht | null>(null)
  const [fehler, setzeFehler] = useState<string>()

  // wahl hält je Zeilennummer den gewählten Druck.
  //
  // **An der Zeilennummer und nicht am Stellenindex**: die Nummer ist die der
  // **Eingabe** und bleibt, auch wenn eine Zeile wegfällt.
  const [wahl, setzeWahl] = useState<Record<number, string>>({})

  // vorgemerkt ist eine Vorschau, die auf die Wahl wartet — samt den Zeilen,
  // für die sie gilt (DOM-03.AK-01, #1698). **Die Zeilen reisen mit**: wer
  // danach eine andere Wahl träfe, schriebe sonst Zeilen, die niemand
  // vorab gesehen hat.
  const [vorgemerkt, setzeVorgemerkt] = useState<{
    vorschau: SammelvorschauDTO
    zeilen: Listeneinfuhrzeile[]
  } | null>(null)

  const vorgang = useVorgang()

  // lauf zählt die Wechsel mit — §7, Regel 2.
  //
  // **`useVorgang` verhindert Parallelstarts, keine überholten Antworten.**
  // Wer Liste A deutet und den Text auf B ändert, während die Antwort
  // unterwegs ist, bekam sie eingetragen: A stand wieder da, während sichtbar
  // B im Feld steht. Angestrichen vom Nachlauf über Phase 4.5, Befund 2.
  const lauf = useLauf()

  const texte = useTexte()

  // deuten ordnet die Zeilen zu — und schreibt nichts.
  const deuten = () => {
    const meiner = lauf.naechster()

    void vorgang.starte('deuten', async () => {
      try {
        const antwort = await api().ListeDeuten(text)
        const gedeutet = auspacken(antwort, () => antwort.deutung)

        // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
        if (!lauf.istAktuell(meiner)) {
          return
        }

        setzeFehler(undefined)
        setzeBericht(null)
        setzeVorgemerkt(null)

        // **Eine neue Deutung räumt die alte Wahl mit** (§7, Regel 1): eine
        // Zeilennummer aus dem vorigen Lauf zeigte sonst auf eine andere Zeile.
        setzeWahl({})
        setzeDeutung(gedeutet)
      } catch (ursache: unknown) {
        // **Für den Fehler gilt dasselbe**, und er ist die unauffälligere
        // Hälfte: eine Meldung zu einer Liste, die niemand mehr im Feld hat,
        // sieht aus wie ein Fehler der jetzigen.
        if (!lauf.istAktuell(meiner)) {
          return
        }

        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  // einlesen fragt erst, was die Einfuhr zusammenführte — DOM-03.AK-01
  // (#1698). **Trifft keine Zeile eine vorhandene, geht es ohne
  // Zwischenschritt weiter**; sonst steht die Vorschau da, und geschrieben
  // wird erst mit der Wahl.
  //
  // **Derselbe Vorgangsschlüssel wie beim Deuten** (§7, Regel 4, Befund 2):
  // die beiden schließen einander aus — wer während einer laufenden Deutung
  // schriebe, schriebe gegen einen Stand, den es gleich nicht mehr gibt.
  const einlesen = () => {
    const meiner = lauf.naechster()

    void vorgang.starte('deuten', async () => {
      const zeilen = entschieden(deutung, wahl)

      if (zeilen.length === 0) {
        setzeFehler(eigene.listeNichtsGewaehlt)

        return
      }

      try {
        const antwort = await api().ListeVorschauen(zeilen)
        const vorschau = auspacken(antwort, () => antwort.vorschau)

        if (!lauf.istAktuell(meiner)) {
          return
        }

        if (braucheVorschau(vorschau)) {
          setzeFehler(undefined)
          setzeVorgemerkt({ vorschau, zeilen })

          return
        }
      } catch (ursache: unknown) {
        if (lauf.istAktuell(meiner)) {
          setzeFehler(meldungAus(ursache, texte))
        }

        return
      }

      await schreiben(zeilen, 'addieren', meiner)
    })
  }

  // mitWahl liest ein, was die Vorschau zeigte — so zusammengeführt, wie
  // gewählt.
  const mitWahl = (zeilen: Listeneinfuhrzeile[], modus: Zusammenfuehrung) => {
    const meiner = lauf.naechster()

    setzeVorgemerkt(null)

    void vorgang.starte('deuten', () => schreiben(zeilen, modus, meiner))
  }

  // schreiben nimmt auf, was entschieden ist.
  const schreiben = async (
    zeilen: Listeneinfuhrzeile[],
    modus: Zusammenfuehrung,
    meiner: number,
  ) => {
    try {
      const antwort = await api().ListeEinlesen(zeilen, modus)

      // **Geschrieben ist geschrieben** (§7, Regel 3). Ein Bericht neben
      // einem Fehler heißt: ein Teil ist durchgekommen — die Sammlung
      // daneben ist veraltet, und ein zweiter Versuch schriebe ihn noch
      // einmal. Ohne Bericht ist gar nichts passiert, und dann wird auch
      // **nicht** nachgeladen: ein Lesefehler überschriebe die Meldung, die
      // sagt, was zu tun ist.
      const teilstand = antwort.bericht

      if (teilstand !== null) {
        aufAufgenommen()
      }

      if (!lauf.istAktuell(meiner)) {
        return
      }

      // **Der Teilstand wird gesetzt, bevor `auspacken` wirft.**
      // Andersherum sah das Fenster von vierzig Zeilen, von denen dreißig
      // durchgingen, nur „es ging schief" — angestrichen vom Nachlauf über
      // Phase 4.5, Befund 3.
      setzeBericht(teilstand)

      auspacken(antwort, () => teilstand)

      setzeFehler(undefined)
    } catch (ursache: unknown) {
      if (!lauf.istAktuell(meiner)) {
        return
      }

      setzeFehler(meldungAus(ursache, texte))
    }
  }

  const zeilen = deutung?.zeilen ?? []

  return (
    <section
      aria-label={eigene.listeTitel}
      className="flex flex-col gap-3 rounded-sm border border-kante p-3"
    >
      <h3 className="text-sm font-medium">{eigene.listeTitel}</h3>

      <Erklaerungsabsatz
        text={eigene.listeErklaerung}
        klasse="max-w-3xl text-xs text-leise"
      />

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium" id={feldID}>
          {eigene.listeFeld}
        </span>
        <textarea
          aria-labelledby={feldID}
          value={text}
          rows={4}
          placeholder={eigene.listePlatzhalter}
          onChange={(ereignis) => {
            setzeText(ereignis.target.value)

            // **Eine Textänderung verwirft die Deutung** — Befund 2 des
            // Zweitblicks über Phase 4.5.
            //
            // Ohne das könnte jemand Liste A deuten, den Text durch Liste B
            // ersetzen und dann »Aufnehmen« drücken: geschrieben würde A,
            // obwohl sichtbar B im Feld steht. Das ist §7 Regel 1 — ein
            // Wechsel des Bezugsobjekts räumt den offenen Kontext.
            setzeDeutung(null)
            setzeWahl({})
            setzeBericht(null)
            setzeVorgemerkt(null)

            // **Und eine noch laufende Deutung wird ungültig** (§7, Regel 2).
            // Ohne diese Zeile käme sie gleich an und setzte die Zuordnung
            // der alten Liste ein.
            lauf.naechster()
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>

      <div className="flex flex-wrap items-center gap-3">
        <button
          type="button"
          onClick={deuten}
          disabled={vorgang.laeuft('deuten')}
          className={knopf}
        >
          {eigene.listeDeuten}
        </button>

        {zeilen.length > 0 && (
          <button
            type="button"
            onClick={einlesen}
            disabled={vorgang.laeuft('deuten')}
            className={knopf}
          >
            {eigene.listeEinlesen}
          </button>
        )}

        {zeilen.length > 0 && <Deutungsbilanz zeilen={zeilen} />}
      </div>

      {/* **Die Vorschau steht vor dem Schreiben** — DOM-03.AK-01 (#1698),
          und nur, wenn eine Zeile eine vorhandene träfe. */}
      {vorgemerkt !== null && (
        <Einfuhrvorschau
          vorschau={vorgemerkt.vorschau}
          laeuft={vorgang.laeuft('deuten')}
          aufEinlesen={(modus) => {
            mitWahl(vorgemerkt.zeilen, modus)
          }}
          aufAbbrechen={() => {
            setzeVorgemerkt(null)
          }}
        />
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {/* **Fließtextfarbe und keine eigene** — `text-text` auf Weiß
          ist mit 17,83 gemessen, `text-slate-100` auf `slate-900` mit 16,28
          (docs/KONTRASTE.md). Der erste Anlauf nahm `slate-700`/`slate-300`,
          und die stehen dort nicht: §3 verlangt **gemessene** Kontraste, und
          ein Wächter hält die Datei gegen den Quelltext. */}
      {bericht !== null && (
        <p className="text-sm text-text">
          {einsetzen(eigene.listeBericht, {
            angelegt: String(bericht.angelegt.length),
            erhoeht: String(bericht.erhoeht),
            stueck: String(bericht.stueck),
          })}
        </p>
      )}

      {bericht !== null && bericht.ausgelassen > 0 && (
        <p className="text-sm text-text">
          {einsetzen(eigene.listeAusgelassen, {
            ausgelassen: String(bericht.ausgelassen),
          })}
        </p>
      )}

      {/* **Was nicht durchkam, steht da** — Befund 4 des Zweitblicks über
          Phase 4.5. Die DTOs trugen `befunde`, und niemand las sie: bei einem
          Fassungskonflikt stand nur »N geändert«, und die übergangene Zeile
          war unsichtbar. */}
      {(bericht?.befunde ?? []).map((befund) => (
        <p key={befund} className="text-sm text-hinweis">
          {befund}
        </p>
      ))}

      {(deutung?.befunde ?? []).map((befund) => (
        <p key={befund} className="text-sm text-hinweis">
          {befund}
        </p>
      ))}

      {zeilen.length > 0 && (
        <ul className="flex flex-col gap-2">
          {zeilen.map((zeile) => (
            <li key={zeile.zeilennummer}>
              <Deutungszeile
                zeile={zeile}
                gewaehlt={wahl[zeile.zeilennummer] ?? ''}
                setze={(printing) => {
                  setzeWahl((vorher) => ({
                    ...vorher,
                    [zeile.zeilennummer]: printing,
                  }))
                }}
              />
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}

// knopf ist die Gestalt der beiden Knöpfe.
const knopf =
  'rounded-sm border border-kante px-3 py-1 text-sm focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-50 '
