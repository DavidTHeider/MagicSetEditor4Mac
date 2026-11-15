// **Die Vorschau einer ungespeicherten Vorlage** — Werkstück 3 von Phase 3.3
// (#1164).
//
// # Warum sie ohne Bestätigung zeichnet
//
// §6: „Eine Änderung an einem Feld ist in der Vorschau zu sehen, ohne dass
// jemand etwas bestätigt." Für eine Vorlage wiegt das schwerer als für ein
// Kartenfeld — Speichern ändert sie **für jedes Set**, das dieses Paket
// benutzt. Wer erst speichern müsste, probierte an allen Sets zugleich.
//
// # Warum sie nicht wartet, sondern abbricht
//
// §7: ein Lauf, den ein neuer abgelöst hat, ist kein Ergebnis. `useLaden` zählt
// die Läufe mit, und Go bricht den vorigen am `context` ab — sonst zeigte das
// Fenster nach vier Anschlägen den drittletzten Stand.
//
// **`behalten`**: das alte Bild bleibt stehen, während das neue entsteht. Ein
// Bild, das bei jedem Tastendruck verschwindet und wiederkommt, ist schlechter
// als ein einen Augenblick altes — dieselbe Wahl wie bei der Kartenvorschau.
import { Bindungsfehler } from '../bindungsfehler'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

import type { Skriptstelle, VorschauDTO } from '../api'

interface VorlagenvorschauEigenschaften {
  paket: string
  datei: string
  inhalt: string

  // karteID nennt die Testkarte — leer heißt: die Probekarte des Pakets
  // (Werkstück U7, #1418).
  karteID: string

  // veraltet sagt, dass der Puffer inzwischen einen Befund hat und `inhalt`
  // der letzte gültige Stand ist — die Marke steht über dem Bild.
  veraltet?: Skriptstelle | null
}

export function Vorlagenvorschau({
  paket,
  datei,
  inhalt,
  karteID,
  veraltet = null,
}: VorlagenvorschauEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.vorlagen

  // **Kein passendes Set ist kein Fehler, sondern eine Auskunft.**
  //
  // Sie wird hier von den echten Fehlern getrennt und als `null` weitergegeben;
  // alles andere fliegt weiter und landet in `stand.fehler`. Ohne diese
  // Trennung stand „Kein offenes Set benutzt diese Vorlage" auch bei einem
  // geschlossenen Bestand — und schickte jemanden ein Set öffnen, das offen
  // ist. Angestrichen vom Zweitblick durch Codex über Phase 3.3.
  const stand = useLaden<VorschauDTO | null>(
    async () => {
      const antwort = await api().Vorlagenvorschau(
        paket,
        datei,
        inhalt,
        1,
        karteID,
      )

      try {
        return auspacken(antwort, () => antwort.vorschau)
      } catch (ursache) {
        // **Am Schlüssel und nicht am Code.** `nichtGefunden` trägt jedes
        // `ErrNotFound`; der leere Zustand gilt für genau einen davon.
        if (
          ursache instanceof Bindungsfehler &&
          ursache.schluessel === 'keinSetMitStil'
        ) {
          return null
        }

        throw ursache
      }
    },
    [paket, datei, inhalt, karteID],
    true,
  )

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{eigene.vorschau}</h3>
      <Erklaerungsabsatz
        text={eigene.vorschauErklaerung}
        klasse="text-sm text-leise"
      />

      {/* **Die Auskunft steht hier und nicht als Meldung des Fensters.** „Kein
          offenes Set benutzt diese Vorlage" ist kein Fehlschlag, sondern eine
          Auskunft über den Bestand — sie gehört neben die Stelle, an der sie
          gilt (§6). */}
      {stand.daten === null && (
        <p className="text-sm text-leise">{eigene.vorschauKeineKarte}</p>
      )}

      {/* **Ein echter Fehler sagt, was er ist.** */}
      {stand.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {stand.fehler}
        </p>
      )}

      {stand.laedt && stand.daten === undefined && (
        <p className="text-sm text-leise">{eigene.vorschauLaeuft}</p>
      )}

      {veraltet !== null && stand.daten != null && (
        <p role="status" className="text-sm text-hinweis">
          {einsetzen(eigene.veraltet, { zeile: String(veraltet.zeile) })}
        </p>
      )}

      {stand.laedt && stand.daten != null && (
        <p className="text-xs text-leise">{eigene.wirdNeuGezeichnet}</p>
      )}

      {stand.daten != null && (
        <>
          <img
            src={stand.daten.bild}
            alt={eigene.vorschau}
            width={stand.daten.breite}
            height={stand.daten.hoehe}
            className="block max-w-full rounded-sm border border-kante"
          />

          {stand.daten.befunde.length > 0 && (
            <ul className="flex flex-col gap-1 text-sm text-hinweis">
              {stand.daten.befunde.map((befund) => (
                <li key={befund}>{befund}</li>
              ))}
            </ul>
          )}
        </>
      )}
    </section>
  )
}
