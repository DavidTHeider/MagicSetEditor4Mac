// **Eine Zeile je Spielbefehl** —
// #1623.
//
// # Was hier steht und was nicht
//
// **Hier steht, wie ein Befehl bedient wird**; *was* er braucht, steht in
// `spiel/befehlsformen.ts`, und *wie daraus eine Nutzlast wird*, in
// `spiel/nutzlast.ts`. Drei Fragen, drei Dateien (§2) — und die mittlere ist
// die, die ein Wächter gegen den Kern hält.
//
// # Und warum jede Zeile ihren eigenen Zustand hat
//
// **Weil zwei Befehle nichts miteinander zu tun haben.** Wer in `Marken` die
// Art tippt und dann `Schaden` stellt, hat nichts an den Marken geändert —
// ein gemeinsamer Zustand ließe die eine Eingabe in der anderen Zeile
// auftauchen.
//
// # Und warum die Felder erst auf Zuruf dastehen
//
// **Weil die Spalte neben dem Tisch sonst eine Formularwand ist** — UX-08
// (#1680)
// und das Artboard *Tisch*, in dem „Mit dieser Karte" drei Knöpfe zeigt und
// kein Formular. Im Sandkasten ist fast jeder Befehl erlaubt, und jeder
// brächte seine Felder mit.
//
// **Aufgeklappt wird in der Spalte und nicht in einem Popover** über dem
// Tisch. Die Felder fragen nach Dingen, die dort liegen — eine Karte zum
// Anhängen, eine Zone, ein Ziel —, und ein Popover deckte genau das zu,
// wonach es fragt.
import { useId, useRef, useState } from 'react'

import { Befehlsfeld } from './Befehlsfeld'
import { Kampffeld } from './Kampffeld'
import { Knopf } from './Knopf'
import { Spielsteinfeld } from './Spielsteinfeld'
import { alsListe, alsText } from '../spiel/feldwerte'
import { handlungswort } from '../seiten/partiewoerter'
import { nutzlastAus, sichtbareFelder, startwert } from '../spiel/nutzlast'
import { useTexte } from '../sprache/kontext'

import type { Befehlsform as Form } from '../spiel/befehlsformen'
import type { Partie, Spielobjekt } from '../apipartie'
import type { Werte } from '../spiel/nutzlast'

export function Befehlsform({
  form,
  partie,
  objekt,
  befehlen,
  angehalten,
  offen,
  umschalten,
}: {
  form: Form
  partie: Partie
  objekt: Spielobjekt | undefined
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  angehalten: boolean

  // offen sagt, ob die Felder dastehen; `umschalten` klappt sie auf und zu.
  // **Beides gehört der Gruppe**: sie hält einen zur Zeit offen.
  offen: boolean
  umschalten: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.handlungen

  const [werte, setzeWerte] = useState<Werte>({})
  const felderID = useId()
  const ausloeser = useRef<HTMLButtonElement>(null)

  const name = handlungswort(eigene.handlungsname, form.art)

  const stellen = () => {
    befehlen(form.art, nutzlastAus(form, werte, objekt, partie))
    setzeWerte({})
  }

  // **Ohne Felder ist der Name der Knopf.** Ein „Stellen" daneben fragte,
  // was gestellt werden soll — und die Antwort stünde auf dem Knopf davor.
  if (form.felder.length === 0) {
    return (
      <Knopf klein angehalten={angehalten} aufKlick={stellen}>
        {name}
      </Knopf>
    )
  }

  // **Mit Feldern klappt der Name sie auf**, und die Auslassungspunkte sagen
  // es vorher — die Regel des Macs für einen Befehl, der noch etwas wissen
  // will. Für ein Vorleseprogramm sagt es `aria-expanded`; die Punkte
  // gehören deshalb nicht zum Namen.
  //
  // **Der Knopf ist nie gesperrt**: Aufklappen schickt nichts. Gesperrt ist
  // `Stellen` (§7, Regel 4).
  //
  // **Die Felder stehen gleich unter ihrem Knopf**, auf einer eigenen Zeile
  // (`w-full`); die Knöpfe danach rücken darunter. Die Tabulatortaste führt
  // vom Knopf in seine Felder, und das Auge findet sie dort, wo es geklickt
  // hat.
  //
  // **Der erste Anlauf stellte sie unter alle Knöpfe der Gruppe**
  // (`order-last`), damit die Zeilen beim Aufklappen stehen blieben. Die
  // WebKit-Probe hat es am 2026-09-24 gemessen: im Sandkasten hat eine Karte
  // zwanzig Handlungen, und die Felder von „Marken…" standen 350 Punkte
  // unter ihrem Knopf — außerhalb der Ansicht.
  return (
    <>
      <Knopf
        klein
        griff={ausloeser}
        ausgeklappt={offen}
        steuert={felderID}
        aufKlick={umschalten}
      >
        {/* **Ein Kind und nicht zwei**: `Knopf` setzt zwischen seine Kinder
            einen Abstand, und die Punkte gehören an das Wort. */}
        <span>
          {name}
          <span aria-hidden="true">…</span>
        </span>
      </Knopf>

      {offen && (
        <div
          id={felderID}
          role="group"
          aria-label={name}
          className="flex w-full flex-col items-start gap-3 rounded-sm border border-kante p-3"
        >
          <div className="flex flex-wrap items-end gap-2">
            {sichtbareFelder(form, werte).map((feld) => (
              <Feldeingabe
                key={feld.name}
                feld={feld}
                partie={partie}
                wert={werte[feld.name] ?? startwert(feld, partie)}
                setze={(wert) => {
                  setzeWerte((bisher) => ({ ...bisher, [feld.name]: wert }))
                }}
              />
            ))}
          </div>

          {/* **Nach dem Stellen klappt es zu, und der Fokus kehrt zum Knopf
              zurück** (UX-08.AK-03): der gedrückte Knopf verschwindet mit
              den Feldern, und ohne Rückkehr stünde der Fokus im Nichts. */}
          <Knopf
            klein
            angehalten={angehalten}
            aufKlick={() => {
              stellen()
              umschalten()
              ausloeser.current?.focus()
            }}
          >
            {eigene.handlungStellen}
          </Knopf>
        </div>
      )}
    </>
  )
}

// Feldeingabe wählt das Bedienelement zu einer Feldart.
//
// **Drei Bauteile und kein `switch` über zwölf Arten**: die einfachen
// unterscheiden sich nur in ihrer Auswahlliste (`Befehlsfeld`), der
// Spielstein bringt vier Angaben mit, und ein Kampf besteht aus Paaren.
function Feldeingabe({
  feld,
  partie,
  wert,
  setze,
}: {
  feld: Form['felder'][number]
  partie: Partie
  wert: string | string[]
  setze: (wert: string | string[]) => void
}) {
  if (feld.art === 'spielstein') {
    return <Spielsteinfeld wert={alsText(wert)} setze={setze} />
  }

  if (feld.art === 'angriffe' || feld.art === 'bloecke') {
    return (
      <Kampffeld
        art={feld.art}
        wert={alsListe(wert)}
        setze={setze}
        partie={partie}
      />
    )
  }

  return <Befehlsfeld feld={feld} wert={wert} setze={setze} partie={partie} />
}
