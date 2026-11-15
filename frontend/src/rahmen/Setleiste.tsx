// **Die Kennzahlen des offenen Sets** — AP-9, Teil 2
// (#2366),
// Befund B-15, Entwurf `Bereiche09`: am Setkopf die Karten und die Karten
// mit Befund, jede Zahl mit Ziel.
//
// # Gelesen, nicht geprüft
//
// **Die Zahl kommt aus dem Prüfstand des Sets** — derselbe, den die
// Startseite liest (`Setprobleme`), geprüft wird nur auf Anstoß (PERF02,
// `LEISTUNGSSCHNITTE.md` §11). Gilt er nicht — nie geprüft, veraltet,
// unvollständig —, steht „nicht geprüft“ und keine Null. **Null Befunde
// führen nirgends hin**: ein Filter ohne Treffer wäre eine Sackgasse.
//
// # Die Ziele
//
// **Karten** führt zur Kartenlasche. **Mit Befund** schlägt die erste Karte
// mit Befund auf und filtert die Liste auf alle — die Karte reist mit der
// Wahl (`oeffneKarte`), wie aus der Aufmerksamkeit (AP-8).
//
// # Was die Leiste nicht trägt
//
// **„ohne Bild“ und „fertig“ aus dem Entwurf** — welches Feld das Bild ist,
// sagt das Spiel, und „fertig“ ist hier eine Marke, die jemand vergibt, kein
// Zustand der Karte. Die Abweichung steht in `docs/PROTOTYP-0.9.0.md`.
import { useCallback } from 'react'

import { useAnschluss } from '../anschlussraum'
import { Kennzahlleiste } from '../bausteine/Kennzahlleiste'
import { api } from '../fassade'
import { useLaden } from '../haken/useLaden'
import type { Seitenschluessel } from '../navigation'
import { alterSatz } from '../seiten/altersatz'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { auspacken } from '../umschlag'

export function Setleiste({
  setID,
  karten,
  verlaufsstand,
  wechseln,
  oeffneKarte,
}: {
  setID: string
  karten: number
  // verlaufsstand steigt nach jedem Schreiben — der Stand wird nachgelesen,
  // denn eine Änderung macht ihn veraltet.
  verlaufsstand: number
  wechseln: (seite: Seitenschluessel) => void
  oeffneKarte: (setID: string, karteID: string) => void
}) {
  const texte = useTexte()
  const eigene = texte.allgemein.arbeitsplatz.leiste
  const anschluss = useAnschluss()

  const stand = useLaden(
    useCallback(async () => {
      const antwort = await api().Setprobleme()

      return auspacken(
        antwort,
        () => antwort.staende.find((eines) => eines.setID === setID) ?? null,
      )
    }, [setID]),
    [setID],
    false,
    verlaufsstand,
  )

  const eines = stand.daten
  const gilt = eines?.sicher === true ? eines : undefined
  const [erste] = gilt?.karten ?? []

  const befund =
    stand.fehler !== undefined
      ? eigene.nichtGelesen
      : eines === undefined
        ? undefined
        : gilt === undefined
          ? eigene.nichtGeprueft
          : String(gilt.zahl)

  return (
    <Kennzahlleiste
      bezeichnung={eigene.bezeichnung}
      kennzahlen={[
        {
          schluessel: 'karten',
          was: texte.seiten.sets.spalten.karten,
          wert: String(karten),
          hin: () => {
            wechseln('karten')
          },
        },
        {
          schluessel: 'befund',
          was: eigene.befund,
          wert: befund,
          unter:
            gilt === undefined
              ? undefined
              : einsetzen(eigene.geprueft, {
                  wann: alterSatz(
                    gilt.alter,
                    texte.seiten.start.ueberblick.alter,
                  ),
                }),
          hin:
            gilt === undefined || erste === undefined
              ? undefined
              : () => {
                  oeffneKarte(setID, erste.karteID)
                  anschluss.uebergib({
                    art: 'kartenauswahl',
                    name: einsetzen(
                      texte.seiten.start.aufmerksamkeit.setBefunde,
                      {
                        karten: zaehlwort(
                          gilt.zahl,
                          texte.allgemein.zaehlwoerter.karten,
                        ),
                      },
                    ),
                    setID,
                    karteIDs: gilt.karten.map((karte) => karte.karteID),
                  })
                },
        },
      ]}
    />
  )
}
