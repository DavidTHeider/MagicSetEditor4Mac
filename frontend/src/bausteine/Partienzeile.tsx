// **Eine Partie in der Liste** — UXR-13
// (#1921),
// nach der Spalte *Deine Partien* des Artboards *Eine Partie anfangen*:
// oben, wer spielt, darunter Spiel, Stand und Zug.
//
// # Zwei Menüs und kein Knopf je Sitz
//
// **Bis UXR-13 trug eine Zeile je Sitz einen Knopf** und daneben Duplizieren,
// Ausführen, Löschen und „+ Marke" — eine Partie zu viert neun
// Bedienelemente, neunzehn Partien 122 (gezählt unter jsdom mit
// `test/spielbestand.ts`). In der schmalen Spalte des Artboards bräche jede
// Zeile mehrfach um, und die Liste stünde lauter da als der Anfang daneben.
//
// **Jetzt sind es zwei, gleich wie viele Sitze**: *Weiterspielen* öffnet die
// Sitze und das Zuschauen, das Menü der Partie die Wege mit ihr als Ganzem.
// Der Sitz wird weiter **gewählt und nicht geraten** — der Grund steht in
// `Partienliste.tsx`.
//
// # Die Marke „dran“
//
// **Das Artboard zeigt an jeder Partie, ob man dran ist** — UXR-F16
// (#2031),
// Folgebefund aus UXR-13. Wer mehrere Partien offen hat, musste jede öffnen,
// um zu sehen, wo eine Entscheidung wartet.
//
// **Gerechnet wird nicht hier.** `partie.dran` kommt vom Dienst (§2): wer
// Priorität hat, wer abwerfen oder eine Legende wählen muss, steht im
// Spielstand der Regelmaschine, und die drei Flächen ohne Fenster nennen
// dasselbe Feld (§16). Die Zeile zeigt es nur — und zeigt nichts, wo es nicht
// gilt: **Zuschauerpartien** (kein Sitz an dieses Profil gebunden), **angehaltene
// und beendete** sowie der **freie Modus** tragen `dran = false`.
//
// **Gefüllt und nicht umrandet**, damit es sich von den Marken darunter
// unterscheidet: die sind eine Ordnung, die man selbst vergibt, diese ist eine
// Auskunft über die Partie. Dieselbe Farbpaarung wie die Zahl über der
// Handelsliste, und damit schon gemessen (`docs/KONTRASTE.md`).
//
// # Warum Zuschauen hier steht
//
// **Weil es eine Art ist, an einen Tisch zu kommen.** Die Domäne kennt den
// Zuschauer seit G1 (`Partie.Sicht("")`), und am Tisch steht er seit
// PLAY-05.AK-02 in der Sichtwahl; aus der Liste ging es bis UXR-13 nur über
// einen Sitz und dann den Wechsel.
import { useState } from 'react'

import { Kontextmenue } from './Kontextmenue'
import { Menue, type Menueeintrag } from './Menue'
import { Objektmarken } from './Objektmarken'
import { Partienverwaltung, verwaltungseintraege } from './Partienverwaltung'
import { einsetzen } from '../sprache/einsetzen'
import { partienname } from './partienauswahl'
import { standwort } from '../seiten/partiewoerter'
import { useTexte } from '../sprache/kontext'

import type { Partiezeile } from '../apipartie'
import type { Verwaltungsstand } from '../haken/usePartieverwaltung'

interface Zeileneigenschaften {
  partie: Partiezeile

  // einsteigen nimmt Partie und Sitz — die leere Sitzkennung ist der
  // Zuschauer.
  einsteigen: (partieID: string, sitz: string) => void

  verwaltung: Verwaltungsstand

  // marken sind die Marken dieser Partie; nachladen liest sie nach einem
  // Schreiben neu.
  marken: readonly string[]
  nachladen: () => void
}

export function Partienzeile({
  partie,
  einsteigen,
  verwaltung,
  marken,
  nachladen,
}: Zeileneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  // **Die Marken stehen auf Wunsch offen** — wie bei den Paketen
  // (`Paketzeilen.tsx`): ein Feld je Zeile wäre bei neunzehn Partien eine
  // Formularwand (R3).
  const [markenOffen, setzeMarkenOffen] = useState(false)

  const name = partienname(partie)

  const weiter: Menueeintrag[] = [
    ...partie.sitze.map((sitz) => ({
      schluessel: sitz.kennung,
      // **Der Eintrag heißt, was er tut, und nicht nur, wen er meint** —
      // WCAG 2.4.6: bei neunzehn Partien hörte ein Vorleseprogramm sonst
      // vierzig Namen ohne Aufgabe.
      name: einsetzen(eigene.partienEinsteigenAls, {
        name: sitz.name === '' ? sitz.kennung : sitz.name,
      }),
      aufWahl: () => {
        einsteigen(partie.id, sitz.kennung)
      },
    })),
    {
      schluessel: 'zuschauen',
      name: eigene.partienZuschauen,
      aufWahl: () => {
        einsteigen(partie.id, '')
      },
    },
  ]
  const umschalten = () => {
    setzeMarkenOffen((war) => !war)
  }

  return (
    // **Der Rechtsklick an der Zeile zeigt alle ihre Handlungen** — AP-5
    // (#2362), Befund B-01: die beiden Menüs darin zusammen, in derselben
    // Reihenfolge. Löschen fragt hier so wenig wie dort; warum, steht in
    // `Partienverwaltung.tsx`.
    <Kontextmenue
      element="li"
      klasse="flex flex-col gap-1 rounded-md border border-kante px-3 py-2"
      bezeichnung={einsetzen(eigene.partienHandlungen, { name })}
      eintraege={[
        ...weiter,
        ...verwaltungseintraege(partie.id, verwaltung, umschalten, texte),
      ]}
      objekt={`partie/${partie.id}`}
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <strong className="font-semibold">{name}</strong>

        {partie.dran && (
          <span className="rounded-full bg-marke px-2 py-0.5 text-xs text-marketext">
            {eigene.partienDran}
          </span>
        )}
      </div>

      {/*
        **Das Spiel steht da, seit man danach suchen kann.** Ein Treffer,
        dessen Grund nicht dasteht, sieht aus wie ein Fehler — und wer zwei
        Spiele führt, unterscheidet seine Tische ohnehin zuerst daran.
      */}
      <p className="flex flex-wrap gap-x-2 text-xs text-leise">
        <span>
          {partie.spiel} · {zeilenstand(partie, eigene)}
        </span>
        <span>
          {eigene.partienObjekte}: {partie.objekte}
        </span>
      </p>

      {marken.length > 0 && (
        <ul className="flex flex-wrap gap-1">
          {marken.map((marke) => (
            <li
              key={marke}
              className="rounded-full border border-kante px-2 py-0.5 text-xs"
            >
              {marke}
            </li>
          ))}
        </ul>
      )}

      <div className="flex flex-wrap items-center gap-2">
        <Menue
          bezeichnung={einsetzen(eigene.partienWeiterspielenIn, { name })}
          eintraege={weiter}
        >
          {eigene.partienWeiterspielen}
        </Menue>

        <Partienverwaltung
          kennung={partie.id}
          name={name}
          lauf={verwaltung}
          marken={umschalten}
        />
      </div>

      {markenOffen && (
        <Objektmarken
          art="partie"
          kennung={partie.id}
          marken={marken}
          geaendert={nachladen}
        />
      )}
    </Kontextmenue>
  )
}

// zeilenstand gibt den Stand in Worten — und den Zug, wo es einen gibt.
//
// **Null heißt: es geht nicht zugweise zu** (G9, #1592), und dann steht keine
// Zugzahl da. Eine Null wäre eine Auskunft, die es nicht gibt.
//
// # Warum hier `standwort` aus `partiewoerter` steht
//
// **Weil hier bis zum 2026-09-17 der rohe Schlüssel stand.** Die Liste trug
// eine eigene Funktion desselben Namens, die `partie.stand` unübersetzt
// durchreichte — in der Liste las sich das als „laeuft", eine Zeile tiefer am
// Tisch als „läuft". Gesehen bei der ersten Sichtprüfung mit einem echten
// Deck; zwei Funktionen mit einem Namen sind eine Falle, und eine davon war
// die falsche.
function zeilenstand(
  partie: Partiezeile,
  eigene: Record<string, unknown> & { partienZug: string },
): string {
  const wort = standwort(eigene, partie.stand)

  return partie.zug === 0
    ? wort
    : `${wort} · ${eigene.partienZug} ${String(partie.zug)}`
}
