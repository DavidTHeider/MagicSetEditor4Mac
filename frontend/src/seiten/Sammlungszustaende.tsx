// **Was die Mitte der Sammlung sagt, wenn keine Liste dasteht** — aus
// `Sammlung.tsx` herausgelöst, als G11 sie über 400 Zeilen trug (§2). Der
// Schnitt folgt der Frage: dort steht die Seite, hier, **welcher Zustand**
// gilt — lädt, leer, gefiltert leer.
//
// # Leer heißt: es kam ein Stand, und er war leer
//
// **Befund 5 des Zweitblicks über Phase G**: nach einem gescheiterten ersten
// Laden stand neben der Meldung „Noch nichts eingetragen" samt Handlungen —
// eine Aussage über Daten, die nie kamen. Der Leerzustand verlangt jetzt
// einen geladenen Stand.
//
// # Aufnehmen schreibt, Einlesen nicht
//
// **Befund 4** desselben Zweitblicks: G10 sperrte *+ Aufnehmen* im Kopf, und
// derselbe Weg stand hier offen. Das Einlesen bleibt offen wie im Kopf — die
// Schublade gibt auch das Ausgeben her, und das darf ein Leser
// (`schreibknoepfe.test.ts`).
import { Knopf } from '../bausteine/Knopf'
import { Wirdgeladen } from '../bausteine/Ladehinweise'
import { Leerzustand } from '../bausteine/Leerzustand'
import { useTexte } from '../sprache/kontext'

export function Sammlungszustaende({
  insgesamt,
  sichtbar,
  laedt,
  geladen,
  aufnehmen,
  einlesen,
  zuruecksetzen,
}: {
  insgesamt: number
  sichtbar: number
  laedt: boolean
  geladen: boolean
  aufnehmen: () => void
  einlesen: () => void
  zuruecksetzen: () => void
}) {
  const eigene = useTexte().seiten.sammlung

  return (
    <>
      {/* **Beim ersten Laden ein Hinweis, kein Leerzustand** — G11 (#2277):
          die Mitte stand sonst leer da. */}
      {insgesamt === 0 && laedt && <Wirdgeladen />}
      {/* **Leer heißt: was fehlt, und der nächste Schritt als Knopf**
          (`docs/UX-ARCHITEKTUR.md`, Zustände). */}
      {insgesamt === 0 && !laedt && geladen && (
        <Leerzustand titel={eigene.leerTitel} erklaerung={eigene.leer}>
          <Knopf schreibt aufKlick={aufnehmen}>
            {eigene.neuesTitel}
          </Knopf>
          <Knopf aufKlick={einlesen}>{eigene.listeTitel}</Knopf>
        </Leerzustand>
      )}
      {/* **Gefiltert leer ist nicht leer** (Artboard *BelegZustaende*,
          UX-08.AK-02): etwas ist da, nur nicht in dieser Sicht — und der
          Knopf nimmt Suche und Filter zurück, statt es dem Menschen zu
          überlassen, welches Feld er leeren muss. */}
      {insgesamt > 0 && sichtbar === 0 && !laedt && (
        <div className="flex flex-wrap items-center gap-3">
          <p className="max-w-3xl text-sm text-leise">{eigene.ohneTreffer}</p>
          <Knopf gewicht="leise" aufKlick={zuruecksetzen}>
            {eigene.ohneTrefferZuruecksetzen}
          </Knopf>
        </div>
      )}
    </>
  )
}
