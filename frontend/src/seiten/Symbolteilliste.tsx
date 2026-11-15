// **Die Teilliste des Symboleditors** —
// #890, um
// Mehrfachwahl, Gruppen und Reihenfolge erweitert in Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// MSE hat sie links neben der Zeichenfläche (`src/gui/symbol/part_list.cpp`,
// Fassung f9d9356 vom 2021-07-06): jede Form mit ihrem Namen und einem
// Sinnbild für ihre Verrechnung.
//
// **Hier steht die Zahl der Punkte dabei.** MSE zeigt sie nicht; sie ist die
// eine Auskunft, die sagt, ob ein Teil überhaupt etwas umschließt — unter drei
// Punkten tut es das nicht.
//
// **Eine Symmetrie hat keine eigenen Punkte**, und „0 Punkte" wäre für sie
// wahr und irreführend zugleich — es liest sich, als habe das Teil darin seine
// verloren. Sie nennt deshalb ihre Kopienzahl
// (#1092).
//
// # Jede Zeile nennt ihr Teil
//
// **Drei Knöpfe je Zeile heißen in jeder Zeile gleich**: „Nach oben", „Nach
// unten", „Teil löschen". Wer nur die Beschriftung hört, weiß nicht, welches
// Teil gemeint ist — deshalb tragen sie zusätzlich eine `bezeichnung` mit dem
// Namen. Das ist genau der Fall, für den `Knopf` sie hat.
//
// # Was hier gewählt werden kann und was nicht
//
// **Nur die oberste Ebene.** Die Kinder einer Gruppe stehen eingerückt
// darunter, sind aber nicht anklickbar. Das ist keine Sparmaßnahme, sondern
// MSEs Modell: `GroupSymbolPartsAction` läuft über `root.parts` (die oberste
// Ebene des Symbols), und `SymbolPartsSelection::select` weist eine Wahl ab,
// deren Vorfahr schon gewählt ist (`src/gui/symbol/selection.cpp`,
// Zeilen 28–32, Fassung `f9d9356` vom 2021-07-06). Eine Wahl über zwei Ebenen
// hinweg hätte in beiden Handlungen keinen Ort.

import { Wirdgeladen } from '../bausteine/Ladehinweise'
import type { SymbolteilDTO } from '../antworten'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface SymbolteillisteEigenschaften {
  teile: SymbolteilDTO[]

  // geladen sagt, dass die Zeichnung aus dem Set angekommen ist — vorher ist
  // sie unbekannt und nicht leer (UX-08.NAT-01); nach einem Fehlschlag
  // (`fehlschlag`) behauptet die Liste gar nichts, der Fehler steht mit
  // seinem nächsten Versuch unter der Seite (S15.NAT-02). **Wahlfrei**, weil
  // eine neue Zeichnung ohne Set nichts zu laden hat.
  geladen?: boolean
  fehlschlag?: boolean
  wahl: readonly number[]
  waehlen: (stelle: number, hinzu: boolean) => void
  loeschen: (stelle: number) => void
  gruppieren: () => void
  aufloesen: () => void
  verschieben: (von: number, nach: number) => void
  kannGruppieren: boolean
  kannAufloesen: boolean
}

export function Symbolteilliste({
  teile,
  geladen = true,
  fehlschlag = false,
  wahl,
  waehlen,
  loeschen,
  gruppieren,
  aufloesen,
  verschieben,
  kannGruppieren,
  kannAufloesen,
}: SymbolteillisteEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol

  if (teile.length === 0) {
    return (
      <section className="flex flex-col gap-2">
        <h3 className="text-lg font-medium">{eigene.teile}</h3>
        {geladen && <p className="text-sm text-leise">{eigene.keineTeile}</p>}
        {!geladen && !fehlschlag && <Wirdgeladen />}
      </section>
    )
  }

  return (
    <section className="flex min-w-72 flex-col gap-2">
      <h3 className="text-lg font-medium">{eigene.teile}</h3>

      <p className="text-sm text-leise">{eigene.wahlHilfe}</p>

      <ul className="flex flex-col gap-1">
        {teile.map((teil, stelle) => (
          <li
            key={teil.name + String(stelle)}
            className="flex flex-col gap-1 rounded-sm border border-kante p-2"
          >
            <div className="flex flex-col gap-1">
              <button
                type="button"
                onClick={(ereignis) => {
                  waehlen(stelle, ereignis.metaKey || ereignis.ctrlKey)
                }}
                // **`aria-pressed` und nicht `aria-current`** (#1168).
                //
                // „Current" ist genau **eines** aus einer Menge — das war
                // richtig, solange die Liste eine Einzelwahl hatte. Seit es
                // eine Mehrfachwahl gibt, ist die Zusage falsch: ein Vorleser
                // nennte drei Zeilen zugleich „aktuell". Der Knopf schaltet
                // die Zugehörigkeit zur Wahl um, und genau das sagt
                // `aria-pressed`.
                aria-pressed={wahl.includes(stelle)}
                className={
                  wahl.includes(stelle)
                    ? 'rounded-sm px-1 py-1 text-left text-sm font-medium text-marke focus:outline-2 focus:outline-offset-2 focus:outline-ring'
                    : 'rounded-sm px-1 py-1 text-left text-sm text-text focus:outline-2 focus:outline-offset-2 focus:outline-ring'
                }
              >
                {aufschrift(teil, eigene)}
              </button>

              <div className="flex flex-wrap items-center gap-1">
                <Knopf
                  gewicht="leise"
                  angehalten={stelle === 0}
                  bezeichnung={einsetzen(eigene.nachObenFuer, {
                    name: teil.name,
                  })}
                  aufKlick={() => {
                    verschieben(stelle, stelle - 1)
                  }}
                >
                  {eigene.nachOben}
                </Knopf>

                <Knopf
                  gewicht="leise"
                  angehalten={stelle === teile.length - 1}
                  bezeichnung={einsetzen(eigene.nachUntenFuer, {
                    name: teil.name,
                  })}
                  aufKlick={() => {
                    verschieben(stelle, stelle + 1)
                  }}
                >
                  {eigene.nachUnten}
                </Knopf>

                <Aktionsknopf
                  aktion="loeschen"
                  gewicht="leise"
                  bezeichnung={einsetzen(eigene.teilLoeschenFuer, {
                    name: teil.name,
                  })}
                  aufKlick={() => {
                    loeschen(stelle)
                  }}
                >
                  {eigene.teilLoeschen}
                </Aktionsknopf>
              </div>
            </div>

            {teil.teile.length > 0 && (
              <ul className="ml-4 flex flex-col border-l border-kante pl-3">
                {teil.teile.map((kind, wo) => (
                  <li
                    key={kind.name + String(wo)}
                    className="px-2 py-0.5 text-sm text-leise"
                  >
                    {einsetzen(eigene.teilKind, { name: kind.name })}
                  </li>
                ))}
              </ul>
            )}
          </li>
        ))}
      </ul>

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          angehalten={!kannGruppieren}
          aufKlick={gruppieren}
        >
          {eigene.gruppieren}
        </Knopf>

        <Knopf gewicht="leise" angehalten={!kannAufloesen} aufKlick={aufloesen}>
          {eigene.aufloesen}
        </Knopf>
      </div>

      <p className="text-sm text-leise">{eigene.gruppierenHilfe}</p>
    </section>
  )
}

// aufschrift nennt ein Teil so, wie es in der Liste steht.
//
// **Drei Arten, drei Sätze.** Eine Gruppe hat keine eigenen Punkte, so wenig
// wie eine Symmetrie — „0 Punkte" läse sich, als habe sie welche verloren.
function aufschrift(
  teil: SymbolteilDTO,
  eigene: ReturnType<typeof useTexte>['seiten']['karten']['symbol'],
): string {
  if (teil.art === 'Symmetrie') {
    return einsetzen(eigene.teilMitKopien, {
      name: teil.name,
      zahl: String(teil.kopien),
    })
  }

  if (teil.art === 'Gruppe') {
    return einsetzen(eigene.teilMitKindern, {
      name: teil.name,
      zahl: String(teil.teile.length),
    })
  }

  return einsetzen(eigene.teilMitPunkten, {
    name: teil.name,
    zahl: String(teil.punkte.length),
  })
}
