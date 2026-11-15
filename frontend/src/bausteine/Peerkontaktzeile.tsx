// **Ein bekanntes Gegenüber** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Eine Zeile mit Namen und Stand, das Übrige auf Wunsch
//
// **UXR-16** (#1924).
// Jede Zeile stand als offenes Formular da: Vergleich, Adresse, Notiz, je
// Zweck ein Knopf, die Sichten, Sperren — mit acht Gegenübern 85
// Bedienelemente, 23 Felder und vier Hauptknöpfe in einer Region (R1–R3 in
// `docs/UIREVIEW.md`). Und vorn stand „Unbestätigt", nicht, **wer**. Jetzt
// trägt die Zeile den Namen, den Stand und die letzte Begegnung, und ihr
// Knopf öffnet, was zu ihr gehört — **eine Zeile zugleich**; welche, hält
// `Peerkontakte`. Der Knopf heißt nach dem, was dort als Erstes steht:
// „Vergleichen", solange das Gegenüber wartet, sonst „Verwalten".
//
// # Warum das Bestätigen ein Feld hat und kein Häkchen
//
// **Weil eine Bestätigung ein Vergleich ist.** Wer bestätigt, hat einen
// Fingerabdruck vor sich gehabt und tippt ihn ab; stimmt er nicht, sagt es die
// Anwendung. Ein „ja, ist er" wäre ein Klick, und ein Klick belegt nichts —
// das Issue verlangt ausdrücklich den Vergleich.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { Peerkontakteinzelheiten } from './Peerkontakteinzelheiten'
import { alsZeitpunkt } from '../zeitpunkt'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'

import type { Peerkontakt } from '../apipeerkontakt'
import type { Kontaktwege } from './Peerkontakteinzelheiten'

interface Peerkontaktzeileneigenschaften extends Kontaktwege {
  kontakt: Peerkontakt
  zwecke: string[]
  laeuft: boolean

  // offen sagt, ob das, was zu diesem Gegenüber gehört, dasteht.
  offen: boolean
  umschalten: () => void
}

// markenklasse ist die Gestalt der Marke neben dem Namen — dieselbe wie an
// einer Fläche (`Flaeche.tsx`), mit Farben, die gemessen sind.
const markenklasse = 'rounded-full border border-kante px-2 text-xs text-leise'

export function Peerkontaktzeile({
  kontakt,
  zwecke,
  laeuft,
  offen,
  umschalten,
  ...wege
}: Peerkontaktzeileneigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const kasten = useId()

  const name = kontaktname(kontakt)
  const aktion =
    kontakt.stand === 'unbestaetigt'
      ? eigene.gegenueber.vergleichen
      : eigene.gegenueber.verwalten

  // **Einmal offen, bleibt gezeichnet** — wie im `Aufklappkasten` und aus
  // demselben Grund (S11.NAT-03): eine halb getippte Notiz übersteht, dass
  // jemand eine andere Zeile öffnet.
  const [gezeichnet, setzeGezeichnet] = useState(offen)
  if (offen && !gezeichnet) {
    setzeGezeichnet(true)
  }

  return (
    <li className="flex flex-col gap-2 rounded-sm border border-kante p-3">
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        <span className="font-medium break-all">{name}</span>

        <span className={markenklasse}>
          {standsatzKontakt(eigene, kontakt)}
        </span>

        {kontakt.gesehen !== '' && (
          <span className="text-sm text-leise">
            {einsetzen(eigene.kontakteGesehen, {
              zeitpunkt: alsZeitpunkt(kontakt.gesehen, darstellung, zeitzone),
            })}
          </span>
        )}

        <span className="ml-auto">
          <Knopf
            klein
            gewicht="leise"
            bezeichnung={`${aktion}: ${name}`}
            ausgeklappt={offen}
            steuert={kasten}
            aufKlick={umschalten}
          >
            {aktion}
          </Knopf>
        </span>
      </div>

      {/* **Ein neuer Schlüssel für jemanden, den dieses Profil kannte**
          (NET-01.AK-03, #1705) — ein eigener Zustand, und nichts von vorher
          gilt für ihn. **Er steht auch an einer geschlossenen Zeile**: wer
          ihn übersieht, hält das Gegenüber für das alte. */}
      {kontakt.lage === 'schluesselGeaendert' && (
        <p className="text-sm text-fehler">
          {eigene.einladung.lageSchluesselGeaendert}
        </p>
      )}

      {gezeichnet && (
        <div id={kasten} hidden={!offen} className="flex flex-col gap-2">
          <Peerkontakteinzelheiten
            kontakt={kontakt}
            name={name}
            zwecke={zwecke}
            laeuft={laeuft}
            {...wege}
          />
        </div>
      )}
    </li>
  )
}

// kontaktname ist, wie die Zeile ein Gegenüber nennt — **der eigene Name**,
// sonst der, den es sich gibt, wie in der Partnerwahl; und hat es keinen,
// der Fingerabdruck: eine Zeile ohne Namen ließe sich von der nächsten nicht
// unterscheiden.
function kontaktname(kontakt: Peerkontakt): string {
  return (
    kontakt.eigenerName || kontakt.selbstgewaehlterName || kontakt.fingerabdruck
  )
}

// standsatzKontakt gibt den Satz zu einem der drei Stände.
//
// **Der Rückfall ist der Stand selbst** — dieselbe Überlegung wie bei
// `standsatz` in `Peeridentitaet.tsx`: ein vierter, den dieses Fenster noch
// nicht kennt, ist immer noch eine Auskunft.
function standsatzKontakt(
  eigene: {
    kontakteStandUnbestaetigt: string
    kontakteStandBestaetigt: string
    kontakteStandGesperrt: string
  },
  kontakt: Peerkontakt,
): string {
  const saetze: Record<string, string> = {
    unbestaetigt: eigene.kontakteStandUnbestaetigt,
    bestaetigt: eigene.kontakteStandBestaetigt,
    gesperrt: eigene.kontakteStandGesperrt,
  }

  return saetze[kontakt.stand] ?? kontakt.stand
}
