// **Teilen über die Ziele des Systems** — NET-04
// (#1708),
// entschieden in 0155.
//
// # Drei Schritte, und nichts geht von selbst hinaus
//
// Erst **vorbereiten** — die Kopie liegt, und hier steht, was sie trägt
// und was nicht (AK-02). Dann **anbieten** — der Dialog von macOS; wer das
// Ziel wählt und den Versand bestätigt, tut es dort. Zuletzt **Fertig**,
// und die Kopie ist weg (AK-04). Wer die Seite vorher verlässt, räumt sie
// auch weg.
//
// # Warum der Aufrufer einen `key` setzt
//
// **Ein Wechsel des Bezugs räumt den offenen Kontext** (§7, Regel 1): wer
// auf eine andere Karte wechselt, soll nicht die Kopie der vorigen teilen.
// Mit `key={bezug}` baut React den Baustein neu, und das Verlassen räumt die
// alte Kopie weg — Regel 1 und 2 in einer Gestalt (`docs/SCHREIBWEGE.md`).
import { useEffect, useEffectEvent, useId, useRef, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useVorgang } from '../haken/useVorgang'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { alsGroesse } from '../seiten/Diagnosepaket'
import { inDieZwischenablage } from '../zwischenablage'

import type { Teilart, TeilgutDTO } from '../apiteilen'

interface Eigenschaften {
  art: Teilart
  bezug: string

  // anlass stößt das Vorbereiten **von außen** an — UXR-06 (#1914). Auf der
  // Kartenseite steht „Teilen …" als Eintrag im Menü „Diese Karte" und nicht
  // als eigener Knopf; das Menü zählt die Zahl hoch, und **jede neue**
  // bereitet vor. Die beim Aufbau gilt als gesehen: ein neuer `key` je Karte
  // baut den Baustein neu, und er soll dann nicht von selbst teilen.
  //
  // **Mit Anlass steht kein eigener Knopf da** — der Anstoß ist außen.
  anlass?: number
}

export function Teilen({ art, bezug, anlass }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.teilen
  const darstellung = useDarstellung()
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [gut, setzeGut] = useState<TeilgutDTO>()
  const [fehler, setzeFehler] = useState<string>()
  const [kopiert, setzeKopiert] = useState(false)

  // **Was offen ist, räumt das Verlassen weg** — die Kennung steht in einer
  // Referenz, weil das Aufräumen den letzten Stand braucht und nicht den
  // beim Aufbau.
  const offen = useRef<string | undefined>(undefined)

  // **Und was nach dem Verlassen ankommt, ebenso** (Zweitblick über Phase
  // 6.7, Z7): die Kennung entsteht, nachdem der Knopf gedrückt wurde, und
  // wer die Seite vorher verlässt, hätte sonst nichts zum Aufräumen.
  const lebt = useRef(true)

  // **Aufbau und Abbau symmetrisch** (Nachlauf über Phase 6.7, N5): unter
  // StrictMode baut React jeden Effekt einmal auf, ab und wieder auf — ein
  // Abbau, der für immer gälte, verwürfe danach jede Teilung.
  useEffect(() => {
    lebt.current = true

    return () => {
      lebt.current = false

      if (offen.current !== undefined) {
        void api().TeilenVerwerfen(offen.current)
      }
    }
  }, [])

  const laeuft = vorgang.laeuft('teilen')
  const teile: Record<string, string | undefined> = eigene.teile

  const fahre = (weg: () => Promise<void>): Promise<void> =>
    vorgang.starte('teilen', async () => {
      setzeFehler(undefined)

      try {
        await weg()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const vorbereiten = () =>
    fahre(async () => {
      const antwort = await api().TeilenVorbereiten(art, bezug)
      const neu = auspacken(antwort, () => antwort.teilgut)

      if (neu !== null && !lebt.current) {
        void api().TeilenVerwerfen(neu.kennung)

        return
      }

      if (neu !== null) {
        offen.current = neu.kennung
        setzeGut(neu)
      }
    })

  const anbieten = (kennung: string) =>
    fahre(async () => {
      const antwort = await api().TeilenAnbieten(kennung)
      auspacken(antwort, () => undefined)
    })

  // **Ein neuer Anlass bereitet vor** — einer, während eine Teilung offen
  // steht, nicht: die erste läge sonst ungeräumt im Teilordner. Läuft das
  // Vorbereiten noch, hält `useVorgang` den zweiten Ruf an.
  const gesehen = useRef(anlass)
  const aufAnlass = useEffectEvent(() => {
    if (offen.current === undefined) {
      void vorbereiten()
    }
  })

  useEffect(() => {
    if (anlass === gesehen.current) {
      return
    }

    gesehen.current = anlass
    aufAnlass()
  }, [anlass])

  // **Die Liste, die ein Deckbaukasten versteht** — dieselbe, die geteilt
  // wird, aber in die Zwischenablage.
  const listeKopieren = () =>
    fahre(async () => {
      const antwort = await api().DecklisteAusgeben(bezug)
      const liste = auspacken(antwort, () => antwort.liste)

      setzeKopiert(await inDieZwischenablage(liste))
    })

  const fertig = (kennung: string) =>
    fahre(async () => {
      // **Erst nach dem geglückten Verwerfen vergessen** — scheitert es,
      // bleibt die Teilung stehen, und „Fertig" räumt noch einmal (Z7, N4).
      const antwort = await api().TeilenVerwerfen(kennung)
      auspacken(antwort, () => undefined)

      offen.current = undefined
      setzeGut(undefined)
      setzeKopiert(false)
    })

  // **Ohne Inhalt ohne Platz** (`empty:hidden`): mit einem Anlass von außen
  // steht ohne Teilung nichts darin, und ein leerer Kasten nähme in einer
  // Spalte mit Abstand eine Lücke ein.
  return (
    <div className="flex flex-col gap-2 empty:hidden">
      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {gut === undefined ? (
        anlass === undefined && (
          <div>
            <Knopf
              gewicht="leise"
              angehalten={laeuft}
              aufKlick={() => void vorbereiten()}
            >
              {eigene.knopf}
            </Knopf>
          </div>
        )
      ) : (
        <section
          aria-labelledby={ueberschrift}
          className="flex max-w-prose flex-col gap-2 rounded-lg border border-kante p-3 text-sm"
        >
          <h4 id={ueberschrift} className="font-semibold">
            {eigene.ueberschrift}
          </h4>

          <p className="break-all font-mono">
            {einsetzen(eigene.datei, {
              dateiname: gut.dateiname,
              groesse: alsGroesse(gut.groesse, darstellung),
            })}
          </p>

          <Teilliste
            titel={eigene.mitgehend}
            schluessel={gut.mitgehendes}
            teile={teile}
          />
          <Teilliste
            titel={eigene.fehlend}
            schluessel={gut.verluste}
            teile={teile}
          />

          {gut.zuGrossFuerMail && (
            <p className="text-fehler">{eigene.zuGross}</p>
          )}

          <p className="text-leise">{eigene.kopie}</p>
          <p className="text-leise">{eigene.empfaenger}</p>

          {/* **Das Anbieten führt** (UXR-16, #1924): vier gleichrangige
              Knöpfe standen hier, und keiner sagte, welcher der Schritt ist,
              für den der Kasten aufging. */}
          <div className="flex flex-wrap items-center gap-2">
            <Knopf
              gewicht="haupt"
              angehalten={laeuft}
              aufKlick={() => void anbieten(gut.kennung)}
            >
              {eigene.anbieten}
            </Knopf>

            {art === 'deck' && (
              <Aktionsknopf
                aktion="kopieren"
                gewicht="leise"
                angehalten={laeuft}
                aufKlick={() => void listeKopieren()}
              >
                {eigene.listeKopieren}
              </Aktionsknopf>
            )}

            {gut.text !== '' && (
              <Aktionsknopf
                aktion="kopieren"
                gewicht="leise"
                aufKlick={() => {
                  void inDieZwischenablage(gut.text).then(setzeKopiert)
                }}
              >
                {eigene.kopieren}
              </Aktionsknopf>
            )}

            <Knopf
              gewicht="leise"
              angehalten={laeuft}
              aufKlick={() => void fertig(gut.kennung)}
            >
              {eigene.fertig}
            </Knopf>

            {kopiert && (
              <span role="status" className="text-leise">
                {eigene.kopiert}
              </span>
            )}
          </div>
        </section>
      )}
    </div>
  )
}

// Teilliste nennt, was mitgeht oder fehlt — ohne Liste, wenn es nichts gibt.
function Teilliste({
  titel,
  schluessel,
  teile,
}: {
  titel: string
  schluessel: string[]
  teile: Record<string, string | undefined>
}) {
  if (schluessel.length === 0) {
    return null
  }

  return (
    <div>
      <p className="font-medium">{titel}</p>
      <ul className="ml-4 list-disc">
        {schluessel.map((eins) => (
          <li key={eins}>{teile[eins] ?? eins}</li>
        ))}
      </ul>
    </div>
  )
}
