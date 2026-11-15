// Eine **Zeile der Setangaben** — Aufschrift, Eingabe, Konflikt, Fehler.
//
// # Warum als eigene Datei
//
// Seit U03 (#2120)
// hält jede Zeile ihren Entwurf selbst (`useEntwurf`), und ein Haken gehört
// in ein Bauteil, nicht in eine Schleife. `Setangaben.tsx` beantwortet, was
// die Lasche tut; wie **ein** Feld aussieht, steht hier — derselbe Schnitt
// wie `Feldzeile` neben dem Karteneditor.
import type { FeldbeschreibungDTO } from '../api'
import { Entwurfskonflikt } from '../bausteine/Entwurfskonflikt'
import { useEntwurf } from '../haken/useEntwurf'
import { Feldeingabe } from './Feldeingabe'
import { fehlerkennung } from './feldkennung'
import { ohneAuszeichnung } from './ohneauszeichnung'

interface SetangabezeileEigenschaften {
  form: FeldbeschreibungDTO
  wert: string
  fehler: string | undefined
  laeuft: boolean
  autoErsetzen: (
    text: string,
    stelle: number,
  ) => Promise<{ text: string; stelle: number }>
  schreiben: (wert: string, bekannt: readonly string[]) => void
}

export function Setangabezeile({
  form,
  wert,
  fehler,
  laeuft,
  autoErsetzen,
  schreiben,
}: SetangabezeileEigenschaften) {
  // **Ein Entwurf übersteht fremde Änderungen** (U03): die Regel steht in
  // `useEntwurf`, und ohne Vorschau gibt es nichts, dem er gemeldet würde.
  const entwurf = useEntwurf(wert, schreiben)

  return (
    // **Ein mehrzeiliges Feld nimmt beide Spalten.** `Description`
    // ist ein Textbereich; in einer halben Spalte wäre er ein
    // Kasten, in den man nicht schreiben mag.
    <label
      className={
        form.mehrzeilig
          ? 'flex flex-col gap-1 @2xl:col-span-2'
          : 'flex flex-col gap-1'
      }
    >
      <span className="text-sm font-medium">{form.aufschrift}</span>

      {/* **Der gespeicherte Wert ist der Schlüssel, solange niemand
          tippt** — eine Rücknahme erreicht das Feld ohne Neuaufbau der
          Lasche; mit Entwurf bleibt der Schlüssel der von damals. */}
      <Feldeingabe
        key={entwurf.schluessel}
        feld={form.name}
        wert={wert}
        form={form}
        angehalten={laeuft}
        fehlerhaft={fehler !== undefined}
        beschriebenVon={
          fehler === undefined ? undefined : fehlerkennung(form.name)
        }
        autoErsetzen={autoErsetzen}
        aendern={entwurf.aendern}
        verlassen={entwurf.verlassen}
      />

      <Entwurfskonflikt
        griff={entwurf}
        fremd={ohneAuszeichnung(wert)}
        angehalten={laeuft}
      />

      {fehler !== undefined && (
        <span role="alert">
          <span id={fehlerkennung(form.name)} className="text-sm text-fehler">
            {fehler}
          </span>
        </span>
      )}
    </label>
  )
}
