// **Was an diesem Tisch gerade zu stellen ist** —
// #1623.
//
// # Warum sie sich aus `erlaubt` ergibt
//
// **Weil der Kern es ausrechnet und zwar je Sitz** (`Sandkasten.Erlaubt`,
// samt Priorität nach CR 117.1). Ein Knopf, den der Kern abwiese, wäre eine
// Einladung, die er danach zurücknimmt — und wer nicht an der Reihe ist,
// sieht so **weniger** statt eines Fehlers.
//
// # Und warum nach Gegenstand gruppiert
//
// **Weil niemand am Tisch nach Werkstücken sucht.** „Was kann ich mit dieser
// Karte tun" ist die Frage, nicht „was kam mit M7".
import { useState } from 'react'

import { Befehlsform } from './Befehlsform'
import { formenDerGruppe } from '../spiel/formenwahl'
import { useTexte } from '../sprache/kontext'

import type { Gruppe } from '../spiel/befehlsformen'
import type { Partie, Spielobjekt } from '../apipartie'
import { Unterabschnittstitel } from './Abschnittstitel'

export function Handlungsleiste({
  partie,
  erlaubt,
  objekt,
  amObjekt,
  befehlen,
  angehalten,
}: {
  partie: Partie
  erlaubt: readonly string[]
  objekt: Spielobjekt | undefined

  // amObjekt ist, was der Kern an der gewählten Karte anbietet (P3,
  // `handlungen`) — **die Kartengruppe nimmt nur das** (P4, #2372, C-10):
  // sonst böte die Leiste an einer Handkarte „Tappen“ an, das Klick und Menü
  // dort nicht kennen. Ohne Angabe bleibt sie bei `erlaubt`.
  amObjekt?: readonly string[] | undefined
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  angehalten: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.handlungen

  // **Vier Gruppen und nicht fünf.** Die Gruppe `zug` — passen, abkürzen,
  // Zug beenden, aufgeben — steht in der `Tischsteuerung` **über** dem
  // Tisch, bei der Auskunft, wer am Zug ist. Sie ist die häufigste Handlung
  // überhaupt, und sie gehört dorthin, wo man liest, dass man dran ist.
  //
  // **Hier noch einmal wäre sie zweimal da** (§2) — und ein Lauf über „den"
  // Passen-Knopf fände zwei.
  const gruppen: { gruppe: Gruppe; titel: string }[] = [
    { gruppe: 'karte', titel: eigene.handlungenKarte },
    { gruppe: 'zone', titel: eigene.handlungenZone },
    { gruppe: 'sitz', titel: eigene.handlungenSitz },
    { gruppe: 'kampf', titel: eigene.handlungenKampf },
  ]

  if (erlaubt.length === 0) {
    return <p className="text-sm text-leise">{eigene.handlungenLeer}</p>
  }

  return (
    <div className="@container flex flex-col gap-4">
      {gruppen.map(({ gruppe, titel }) => (
        <Handlungsgruppe
          // **Die Kartengruppe hängt an der Karte** (§7, Regel 1): wer eine
          // andere wählt, findet ihre Felder zu und leer — die Anzahl der
          // Marken von vorhin gehörte der anderen. Die übrigen Gruppen
          // fragen nach keiner Karte und bleiben, wie sie sind.
          key={gruppe === 'karte' ? `karte:${objekt?.kennung ?? ''}` : gruppe}
          gruppe={gruppe}
          titel={titel}
          partie={partie}
          erlaubt={
            gruppe === 'karte' && amObjekt !== undefined
              ? erlaubt.filter((art) => amObjekt.includes(art))
              : erlaubt
          }
          objekt={objekt}
          befehlen={befehlen}
          angehalten={angehalten}
        />
      ))}

      {/* **Warum ein Knopf fehlt** — Artboard *Tisch*: angeboten wird, was
          `Erlaubt` im Kern gerade zulässt. Ohne den Satz sähe ein fehlender
          Knopf aus wie ein vergessener. */}
      <p className="text-sm text-leise">{eigene.handlungenNurErlaubt}</p>
    </div>
  )
}

// Handlungsgruppe zeigt die Befehle eines Gegenstands.
//
// **Leer heißt: gar nicht da.** Eine Überschrift ohne Knöpfe darunter sagt
// „hier wäre etwas" und ist eine Zusage, die niemand einlöst (§10).
function Handlungsgruppe({
  gruppe,
  titel,
  partie,
  erlaubt,
  objekt,
  befehlen,
  angehalten,
}: {
  gruppe: Gruppe
  titel: string
  partie: Partie
  erlaubt: readonly string[]
  objekt: Spielobjekt | undefined
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  angehalten: boolean
}) {
  const texte = useTexte()

  // **Eine Form zur Zeit ist offen** — zwei offene wären wieder die Wand, die
  // das Aufklappen wegnehmen soll.
  const [offen, setzeOffen] = useState<string>()

  const formen = formenDerGruppe(gruppe, erlaubt, objekt)

  if (formen.length === 0) {
    return undefined
  }

  // **Ohne gewählte Karte steht der Satz statt der Knöpfe.** Wer sie zeigte,
  // böte Handlungen für ein Objekt an, das niemand genannt hat — und der
  // Kern antwortete mit „kein Objekt genannt".
  if (gruppe === 'karte' && objekt === undefined) {
    return (
      <section className="flex flex-col gap-2">
        <Unterabschnittstitel className="text-sm font-semibold">
          {titel}
        </Unterabschnittstitel>

        <p className="text-sm text-leise">
          {texte.seiten.handlungen.handlungenKeineKarte}
        </p>
      </section>
    )
  }

  return (
    <section className="flex flex-col gap-2">
      <Unterabschnittstitel className="text-sm font-semibold">
        {titel}
      </Unterabschnittstitel>

      <div className="flex flex-wrap gap-2">
        {formen.map((form) => (
          <Befehlsform
            key={form.art}
            form={form}
            partie={partie}
            objekt={objekt}
            befehlen={befehlen}
            angehalten={angehalten}
            offen={offen === form.art}
            umschalten={() => {
              setzeOffen((war) => (war === form.art ? undefined : form.art))
            }}
          />
        ))}
      </div>
    </section>
  )
}
