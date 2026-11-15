// **Die Nutzer und ihre Profile laden** — Phase 4.12
// (#1351).
//
// **Über `useLaden`, wie jeder Ladeeffekt dieses Fensters** (§7, Regel 2): er
// zählt den Stand mit, aus dem eine Antwort stammt, und wirft eine überholte
// weg.
import { useCallback, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { AnbieterkontoDTO, NutzerDTO } from '../api'
import type { AnmeldeanbieterDTO } from '../apinutzer'

export interface Nutzerstand {
  nutzer: readonly NutzerDTO[]

  // anbieter sind die Anbieter, über die man sich anmelden kann.
  anbieter: readonly string[]

  // anbieterstaende sagen je Anbieter, ob er eingerichtet ist und wogegen er
  // geprüft ist (ID-05).
  anbieterstaende: readonly AnmeldeanbieterDTO[]

  // konten sind die Anbieterkonten des Nutzers, der das gezeigte Profil hat.
  konten: readonly AnbieterkontoDTO[]

  // eigner ist der Nutzer, dem das gezeigte Profil gehört.
  eigner: string

  // arbeitet ist der Mensch, der gerade arbeitet — Werkstück 17 von Phase 4.12.
  //
  // **Nicht dasselbe wie [eigner].** Wer in ein fremdes Profil gewechselt ist,
  // bleibt, wer er ist; der Eigner ist dann ein anderer.
  arbeitet: string

  // gezeigt ist die Kennung des Profils, in dem gearbeitet wird — leer,
  // solange sie nicht dasteht.
  gezeigt: string

  // bilder sind die Profilbilder als Daten-URI, je Nutzerkennung — wer
  // keines hat, steht nicht darin.
  bilder: Record<string, string>

  // badges sind die Abzeichen, je **Profil**kennung — Werkstück 9 (#1351).
  badges: Record<string, string>

  laedt: boolean
  fehler: string | undefined
  neuLaden: () => void
}

// **`fremdstand` zählt, was ein anderer Prozess geschrieben hat** (#156) —
// Befund 4 des Nachlaufs zum Gesamtreview über 0.6.0 (#1660): legte die
// Kommandozeile einen Nutzer an, stand er erst nach einem Wechsel da.
export function useNutzer(fremdstand = 0): Nutzerstand {
  // **Eine Zahl und kein Rückruf** — dieselbe Form wie bei der Paketliste:
  // wer neu laden will, zählt hoch, und der Effekt sieht es.
  const [runde, setzeRunde] = useState(0)

  // **Fünf Rufe, zusammen angestoßen** — und das ist **kein** gemeinsamer
  // Stand. Was es ist und was nicht, steht unten bei `gelesen`.
  const holen = useCallback(async () => {
    const [liste, profil, arbeitet, anbieter, bilder] = await Promise.all([
      api().Nutzer(),
      api().GezeigtesProfil(),
      // **Wer arbeitet, im selben Zug** — Werkstück 17 von Phase 4.12
      // (#1400).
      //
      // **Nicht aus `GezeigtesProfil` abzulesen**: wer in ein fremdes Profil
      // gewechselt ist, bleibt, wer er ist — und `profil.nutzerID` nennt den
      // **Eigner** des Profils, in dem gearbeitet wird.
      api().WerArbeitet(),
      api().Anbieter(),
      // **Zusammen mit der Liste angestoßen**, damit die Seite nicht zweimal
      // nacheinander wartet. **Ein Bild kann trotzdem zu einem Namen
      // gehören, den es nicht mehr gibt** — dieselbe Einschränkung wie oben,
      // und sie kostet ein Bild, das einen Augenblick alt ist.
      api().Nutzerbilder(),
    ])

    const gelesen = auspacken(liste, () => liste.nutzer)
    const gezeigt = auspacken(profil, () => profil.profil)?.id ?? ''

    // **Die Konten des Nutzers, dem das gezeigte Profil gehört.**
    //
    // **Und ein vierter Ruf, weil sie an einem Nutzer hängen** — die Übersicht
    // führt Profile, nicht Konten.
    //
    // # Ein Zug ist kein Stand
    //
    // `Promise.all` **startet** fünf Rufe zugleich, aber jeder geht als eigene
    // Anfrage über die Wails-Naht und liest in einer eigenen Transaktion. Wer
    // dazwischen wechselt oder schreibt, bekommt Antworten aus zwei Ständen —
    // die Nutzerliste von vorhin und das gezeigte Profil von jetzt.
    //
    // **Was das kostet, ist eine Anzeige, die einen Augenblick alt ist** — und
    // nicht ein falscher Schreibvorgang. Der Grund dafür:
    //
    // Die Umschaltsperre hält nicht jeden Wechsel gegen **jede** Bindung an —
    // `app.imProfil` hält sie nur, **solange es die Kennung liest**, und der
    // Ruf danach läuft ohne sie. Was wirklich gilt: ein Umschalten geschieht
    // ganz oder gar nicht (`App.profilUmstellen`), und ein Ruf, der die
    // Kennung schon hat, führt sie zu Ende. Das ist §7, Regel 2 — eine Antwort
    // gehört zu dem Stand, aus dem sie stammt.
    //
    // **Ein Backend-Verbund wäre die vollständige Antwort** und eine neue
    // Fähigkeit an vier Flächen (§16) — großer Eingriff für eine Zahl, die
    // beim nächsten Laden ohnehin stimmt. Der Preis steht hier, damit ihn
    // niemand für gelöst hält.
    const eigner = gelesen.find((einer) =>
      einer.profile.some((eines) => eines.id === gezeigt),
    )

    const konten =
      eigner === undefined
        ? { konten: [] }
        : await api().Anbieterkonten(eigner.id)

    return {
      nutzer: gelesen,
      anbieter: auspacken(anbieter, () => anbieter.anbieter),
      anbieterstaende: auspacken(anbieter, () => anbieter.staende ?? []),
      konten: auspacken(konten, () => konten.konten),
      eigner: eigner?.id ?? '',
      arbeitet: auspacken(arbeitet, () => arbeitet.nutzerID),
      // **`auspacken` wirft, wenn der Umschlag einen Fehler trägt** — und
      // gibt sonst, was der Rückruf liefert. Dass die Bindung `null` geben
      // *könnte*, sagt der Typ; im Betrieb kommt es nicht vor: es gibt immer
      // ein gezeigtes Profil (Wanderung 0034). Ein leerer Wert wäre die
      // ehrlichere Antwort als ein Absturz.
      gezeigt,
      bilder: auspacken(bilder, () => bilder.bilder),
      badges: auspacken(bilder, () => bilder.badges),
    }
  }, [])

  // **`behalten`, weil ein Nachladen kein Wechsel des Bezugsobjekts ist.**
  //
  // §7, Regel 1 räumt den Kontext, wenn das **Bezugsobjekt** wechselt — hier
  // wechselt keines, es kommt dieselbe Liste noch einmal. Ohne diese Zeile
  // klappte die Seite nach jedem Anlegen auf „lädt …" um und riss das
  // Formular samt seiner Meldung mit weg.
  const { daten, laedt, fehler } = useLaden(holen, [runde, fremdstand], true)

  const neuLaden = useCallback(() => {
    setzeRunde((zahl) => zahl + 1)
  }, [])

  return {
    nutzer: daten?.nutzer ?? [],
    anbieter: daten?.anbieter ?? [],
    anbieterstaende: daten?.anbieterstaende ?? [],
    konten: daten?.konten ?? [],
    eigner: daten?.eigner ?? '',
    arbeitet: daten?.arbeitet ?? '',
    gezeigt: daten?.gezeigt ?? '',
    bilder: daten?.bilder ?? {},
    badges: daten?.badges ?? {},
    laedt,
    fehler,
    neuLaden,
  }
}
