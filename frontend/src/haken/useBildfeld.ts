// **Ein Bildfeld füllen** — über den Dialog oder durch Ablegen
// (#246).
//
// Eigene Datei, weil `Karteneditor.tsx` mit ihr über 400 Zeilen ginge (§2) —
// und weil der Schnitt einer Frage folgt: dort steht, was das Fenster zeigt,
// hier, wie ein Bild hineinkommt.
import { useEffect, useState } from 'react'

import {
  api,
  auspacken,
  type AuswahlDTO,
  type Umschlag,
  type ZuschnittAntwort,
} from '../api'
import type { ScryfallbildAntwort } from '../apiscryfall'
import { meldungAus, verursachendesFeld } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useAblage } from '../ablage'

// Feldmeldung ordnet eine Meldung dem Feld zu, das sie verursacht hat.
export interface Feldmeldung {
  feld: string
  meldung: string
}

// Vorgangsnaht ist der Teil von `useVorgang`, den dieser Haken braucht.
//
// **Eine eigene Beschreibung und kein Import des Hakentyps**: er nimmt den
// Vorgang von außen, damit Schreiben und Bildsetzen sich **denselben**
// Schlüssel teilen (§7 Regel 4) — zwei Vorgänge auf ein Feld wären zwei
// gleichzeitige Änderungen an derselben Zeile.
interface Vorgangsnaht {
  // **`undefined` heißt: es lief schon einer.** `useVorgang` gibt dann gar
  // nichts zurück, und das ist etwas anderes als „gescheitert" — nachgeladen
  // wird nur nach einem geglückten Schreiben (§7 Regel 3).
  starte: (
    schluessel: string,
    tun: () => Promise<boolean>,
  ) => Promise<boolean | undefined>
}

interface BildfeldEigenschaften {
  // setID ist das Set, gegen das der Dienst jede Kartenänderung hält
  // (#808).
  setID: string

  karteID: string
  vorgang: Vorgangsnaht
  melden: (fehler: Feldmeldung | undefined) => void

  // geglueckt läuft nach einem geglückten Einsetzen — Regel 3: nachgeladen
  // wird **nur** dann.
  geglueckt: () => void
}

// Offener Zuschnitt ist eine Wahl, die auf ihren Ausschnitt wartet.
export interface OffenerZuschnitt {
  feld: string
  antwort: ZuschnittAntwort

  // **ki steht nur bei einem Bild, das ein Modell erzeugt hat** (A12,
  // #1451).
  //
  // **Neben `scryfall` und nicht statt seiner**: die drei Wege — Platte,
  // Scryfall, Modell — enden in **einem** Zuschnittfenster und setzen über
  // **drei** Bindungen ein, weil ihre Pflichten verschieden sind. Ein Bild
  // von der Platte hat keine Herkunft, eines von Scryfall eine Nennung,
  // eines aus einem Modell einen Vorschlag im Vorrat, der danach fällt.
  ki?: { kennung: string }

  // **scryfall steht nur bei einem Bild von diesem Dienst.**
  //
  // Es ist der Grund, aus dem das Einsetzen sich verzweigt: die Bedingung von
  // Scryfall verlangt Künstler **und** Copyright in derselben Oberfläche wie
  // das Bild, und die `.mse-set`-Ausfuhr ist auch eine. Ein Bild von der
  // Platte hat diese Pflicht nicht.
  //
  // **Die drei stehen zusammen in einem Feld und nicht einzeln daneben.** Ein
  // Künstler ohne Nennung wäre ein Zustand, den es nicht geben darf — und
  // drei einzelne Felder ließen ihn zu, samt einer Verzweigung dafür, die
  // kein Aufruf betritt.
  scryfall?: {
    kuenstler: string
    nennung: string
    adresse: string

    // befunde ist, was die Kartenauskunft tolerant hingenommen hat — eine
    // Drift beim Anbieter, die niemanden aufhält und die niemand
    // verschweigen soll (§16). Der Weg dahinter warf sie weg (Nachlauf über
    // Phase 4.5, Befund 10).
    befunde: string[]
  }
}

export function useBildfeld({
  setID,
  karteID,
  vorgang,
  melden,
  geglueckt,
}: BildfeldEigenschaften): {
  bildSetzen: (feld: string) => Promise<boolean | undefined>
  vonScryfall: (feld: string, name: string, art: string) => Promise<boolean>

  // vonKI öffnet den Zuschnitt für einen wartenden Bildvorschlag (A12).
  vonKI: (feld: string, kennung: string) => Promise<boolean>
  // merkeFeld hält fest, über welchem Feld der Zeiger steht (#891).
  //
  // **Eine Methode und kein Verweis**: was hier geteilt wird, ist Verhalten
  // und kein Zwischenergebnis des Zeichnens ([Ablagestelle]).
  merkeFeld: (feld: string | undefined) => void
  zuschnitt: OffenerZuschnitt | undefined
  einsetzen: (auswahl: AuswahlDTO) => Promise<boolean | undefined>
  abbrechen: () => void
} {
  const texte = useTexte()

  // **Der offene Zuschnitt ist Kontext, und ein Kartenwechsel räumt ihn**
  // (§7, Regel 1). Ohne das stünde nach dem Wechsel ein Fenster da, dessen
  // „Einsetzen" ein Feld der **vorigen** Karte träfe.
  const [zuschnitt, setzeZuschnitt] = useState<OffenerZuschnitt | undefined>(
    undefined,
  )

  // **Beim Zeichnen zurückgesetzt, nicht in einem Effekt.** Ein Effekt liefe
  // **nach** dem Zeichnen: einen Herzschlag lang stünde das Fenster der
  // vorigen Karte da, und React meldet den Fall selbst als „cascading
  // renders". Der Weg, den React dafür nennt, ist ein Vergleich mit dem
  // vorigen Wert und ein Setzen **während** des Zeichnens.
  const [vorigeKarte, setzeVorigeKarte] = useState(karteID)

  if (karteID !== vorigeKarte) {
    setzeVorigeKarte(karteID)
    setzeZuschnitt(undefined)
  }

  // bildvorgang führt beide Wege — Dialog und Ablegen — durch dieselben Regeln.
  const bildvorgang = async (feld: string, rufen: () => Promise<Umschlag>) => {
    // Regel 4: ein Vorgang je Datensatz, und der Schlüssel ist Karte **und**
    // Feld — derselbe wie beim Schreiben, denn beides schreibt dasselbe Feld.
    const ergebnis = await vorgang.starte(`${karteID}/${feld}`, async () => {
      try {
        auspacken(await rufen(), () => undefined)

        return true
      } catch (ursache: unknown) {
        melden({
          feld: verursachendesFeld(ursache, feld),
          meldung: meldungAus(ursache, texte),
        })

        return false
      }
    })

    if (ergebnis === true) {
      melden(undefined)

      // **Das Feld wird neu gelesen.** Der Dateiname entsteht beim Ablegen;
      // ohne diesen Ruf stünde im Eingabefeld weiter der alte.
      geglueckt()
    }

    return ergebnis
  }

  // vorbereiten führt beide Wege — Dialog und Ablegen — zum Zuschnittfenster.
  //
  // **Vorbereiten ist kein Vorgang** (§7, Regel 4): es schreibt nichts. Der
  // Vorgang beginnt beim Einsetzen, und dort trägt er denselben Schlüssel wie
  // jedes andere Schreiben auf dieses Feld.
  const vorbereiten = async (
    feld: string,
    rufen: () => Promise<ZuschnittAntwort>,
    // dazu sind die Felder, die nur ein Bild von Scryfall hat.
    //
    // **Als Funktion und nicht als Wert**: sie hängen an der Antwort, und die
    // gibt es erst nach dem Ruf.
    dazu: (antwort: ZuschnittAntwort) => Partial<OffenerZuschnitt> = () => ({}),
  ) => {
    try {
      const antwort = await rufen()

      auspacken(antwort, () => undefined)
      melden(undefined)

      // **Ein leerer Pfad heißt „abgebrochen"** — nichts gewählt, nichts offen.
      setzeZuschnitt(
        antwort.pfad === '' ? undefined : { feld, antwort, ...dazu(antwort) },
      )

      return antwort.pfad !== ''
    } catch (ursache: unknown) {
      melden({
        feld: verursachendesFeld(ursache, feld),
        meldung: meldungAus(ursache, texte),
      })

      return false
    }
  }

  // **bildSetzen öffnet den Dateidialog.** Er steht in Go ([app.BildSetzen]):
  // ein Fenster bekäme einen Namen und keinen Pfad.
  const bildSetzen = (feld: string) =>
    vorbereiten(feld, () => api().BildSetzen(karteID, feld))

  // **Und eine gezogene Datei landet im Feld darunter** (§6). Der Pfad kommt
  // aus Wails, das Feld aus dem Überziehen — [useAblage].
  //
  // **Angemeldet wird eine Ebene höher** (#891): Wails meldet ein Ablegen
  // einmal fürs ganze Fenster, und ein Horcher auf der Kartenseite bekam
  // nichts mit, wenn keine offen war — und machte aus einer gezogenen
  // `.mse-set` ein Bild. Hier steht nur noch, **was mit einem Bild geschieht**.
  const ablage = useAblage()

  useEffect(() => {
    ablage.setzeBildweg((feld, pfad) => {
      void vorbereiten(feld, () => api().BildAusDatei(karteID, feld, pfad))
    })

    // **Und beim Abräumen wieder heraus.** Eine geschlossene Kartenseite
    // hinterließe sonst einen Weg, der auf ihre Karte zeigt.
    //
    // **Das gemerkte Feld geht mit** (§7, Regel 1) — angestrichen vom
    // Zweitblick über Phase Y als Befund 2. Wer ein Bild über `kunst` zieht,
    // den Zug abbricht, die Karte wechselt und dann **neben** die Karte
    // ablegt, setzte es sonst in das `kunst` der **neuen** Karte: `nimmFeld`
    // gab noch das Feld von vorhin, und `bildweg` zeigte schon auf die neue.
    return () => {
      ablage.setzeBildweg(undefined)
      ablage.merkeFeld(undefined)
    }
  })

  // **vonScryfall holt das Bild zum Kartennamen** (#1005).
  //
  // **Kein Dialog und kein Ziehen** — der Name bestimmt das Bild. Es endet wie
  // die anderen beiden im Zuschnittfenster, und geschrieben wird erst dort.
  const vonScryfall = (feld: string, name: string, art: string) =>
    vorbereiten(
      feld,
      async () =>
        (await api().ScryfallBild(
          karteID,
          feld,
          name,
          art,
        )) as unknown as ZuschnittAntwort,
      (antwort) => {
        // **Die Nennung reist mit dem Bild** und wird nicht daneben gebaut:
        // wer das Bild hat, hat damit auch, was neben ihm stehen muss.
        const bild = antwort as unknown as ScryfallbildAntwort

        return {
          scryfall: {
            kuenstler: bild.kuenstler,
            nennung: bild.nennung,
            adresse: bild.adresse,
            befunde: bild.befunde,
          },
        }
      },
    )

  // **vonKI holt das wartende Bild aus dem Vorrat** (A12, #1451).
  //
  // **Derselbe Weg wie die anderen beiden** — und das ist der Punkt: das
  // Issue verlangt „Übernahme nutzt vorhandene Crop-/Ablage-/Renderfunktion".
  // Ein zweites Zuschnittfenster daneben wäre eine zweite Antwort auf
  // dieselbe Frage.
  const vonKI = (feld: string, kennung: string) =>
    vorbereiten(
      feld,
      () => api().KIBildZuschnitt(karteID, feld, kennung),
      () => ({ ki: { kennung } }),
    )

  // **einsetzen legt ab, was das Fenster gewählt hat.**
  const einsetzen = async (auswahl: AuswahlDTO) => {
    /* v8 ignore next 3 -- unerreichbar: ohne offenen Zuschnitt steht kein Fenster */
    if (zuschnitt === undefined) {
      return undefined
    }

    // **Zwei Wege, weil die Pflicht verschieden ist.** Ein Bild von Scryfall
    // nimmt seinen Künstler mit in die Karte; eines von der Platte hat keinen
    // zu nennen. Beide gehen durch denselben Vorgang (§7, Regel 4).
    const ergebnis = await bildvorgang(zuschnitt.feld, () =>
      bindungFuer(zuschnitt, { setID, karteID, auswahl }),
    )

    // **Nur nach einem geglückten Einsetzen schließt es** (§7, Regel 3): wer
    // die Meldung liest, soll die Wahl noch dastehen sehen.
    if (ergebnis === true) {
      setzeZuschnitt(undefined)
    }

    return ergebnis
  }

  const abbrechen = () => {
    setzeZuschnitt(undefined)
    melden(undefined)
  }

  return {
    bildSetzen,
    vonScryfall,
    vonKI,
    merkeFeld: ablage.merkeFeld,
    zuschnitt,
    einsetzen,
    abbrechen,
  }
}

// bindungFuer wählt die Bindung, über die dieser Zuschnitt eingesetzt wird.
//
// **Drei Wege, weil die Pflicht verschieden ist.** Ein Bild von Scryfall
// nimmt seinen Künstler mit in die Karte; eines von der Platte hat keinen zu
// nennen; eines aus einem Modell trägt seine Herkunft und räumt danach den
// Vorrat. Der Zuschnitt davor ist für alle drei derselbe.
function bindungFuer(
  zuschnitt: OffenerZuschnitt,
  wohin: { setID: string; karteID: string; auswahl: AuswahlDTO },
): Promise<Umschlag> {
  const { setID, karteID, auswahl } = wohin

  if (zuschnitt.ki !== undefined) {
    return api().KIBildEinsetzen(
      setID,
      karteID,
      zuschnitt.feld,
      zuschnitt.ki.kennung,
      auswahl,
    )
  }

  if (zuschnitt.scryfall !== undefined) {
    return api().ScryfallBildEinsetzen(
      setID,
      karteID,
      zuschnitt.feld,
      zuschnitt.antwort.pfad,
      zuschnitt.scryfall.kuenstler,
      zuschnitt.scryfall.adresse,
      auswahl,
    )
  }

  return api().BildEinsetzen(
    setID,
    karteID,
    zuschnitt.feld,
    zuschnitt.antwort.pfad,
    auswahl,
  )
}
