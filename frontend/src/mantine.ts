// **Mantine, an diese Anwendung angeschlossen** — UX-07
// (#1679),
// gemessen und begründet in
// [0125](../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md).
//
// # Was hier steht und was nicht
//
// **Hier steht, wie Mantine an dieses Projekt angeschlossen ist** — die
// Umgebung und das Thema. *Dass* es angeschlossen ist, tut `Anwendung.tsx`;
// *was* damit gebaut wird, steht in den Bauteilen.
//
// **Und es ist bewusst kein `.tsx`**: hier entsteht kein Baum, sondern zwei
// Werte. Beide lassen sich als gewöhnliche Funktionen prüfen, und damit
// entsteht kein Zweig, den nur eine Prüfung betritt.

import { createTheme, useMantineColorScheme } from '@mantine/core'
import { useLayoutEffect } from 'react'

// Mantineumgebung ist, was `MantineProvider` als `env` bekommt.
export type Mantineumgebung = 'default' | 'test'

// mantineumgebung sagt, in welcher Umgebung Mantine läuft.
//
// # Warum das überhaupt eine Frage ist
//
// **Weil Mantine seine Overlays sonst nicht zeigt, wo niemand rechnet.** Jedes
// Overlay — Modal, Menü, Popover, Auswahlliste — geht durch `Transition`, und
// die blendet über eine Dauer ein. Unter jsdom läuft diese Dauer, aber es gibt
// kein Layout; der Inhalt bleibt auf `display: none` stehen und ist damit für
// **jede rollenbasierte Abfrage unsichtbar**. Gemessen am 2026-09-20: ohne
// diese Umgebung fallen Menü, Auswahlliste und Hinweisfähnchen aus, mit ihr
// bestehen alle sieben Proben.
//
// **Der Weg steht in Mantines Quelltext**, nicht in einer Erinnerung:
// `Transition.mjs`, Zeile 22 — `if (env === "test") return mounted ? … : …`.
// Die Animation entfällt dann ganz.
//
// # Warum als Funktion und nicht als Ausdruck am Provider
//
// **Weil ein `import.meta.env.MODE === 'test'` im Baum ein Zweig wäre, den nur
// die Prüfung betritt** — und die andere Hälfte stünde für immer ungemessen
// da. So steht die Entscheidung in einer Funktion, die beide Eingaben
// bekommt, und der Aufruf daneben ist ein Ausdruck ohne Verzweigung.
export function mantineumgebung(modus: string): Mantineumgebung {
  return modus === 'test' ? 'test' : 'default'
}

// mantinethema bindet Mantine an die Gestaltungssprache dieser Anwendung.
//
// # Warum überhaupt ein Thema
//
// **Weil Mantine sonst seine eigene Sprache spräche.** Es bringt eine Palette,
// eine Schrift und eine Rundung mit; diese Anwendung hat beides bereits, und
// zwar gemessen ([0065](../../docs/entscheidungen/0065-wie-diese-anwendung-aussieht.md),
// [`KONTRASTE.md`](../../docs/KONTRASTE.md)). Zwei Gestaltungssprachen
// nebeneinander wären genau der CSS-Konflikt, den UX-07.AK-04 verhindern
// will.
//
// # Was hier gesetzt wird — und was ausdrücklich nicht
//
//   - **Die Schriften** zeigen auf dieselben Merkmale wie `stil.css`. Ein
//     zweiter Schriftstapel daneben liefe auseinander.
//   - **Die Rundung** ist `rounded`/`rounded-md`, wie Festlegung 3 es sagt.
//   - **Die Farben nicht.** Mantines Palette ist ein Zahlenraster von 0 bis 9
//     je Ton; unsere sind acht benannte Merkmale. Eine Abbildung darauf wäre
//     eine Erfindung — welches Grau ist `--mse-leise`? Die Bauteile bekommen
//     ihre Farben deshalb weiter über Tailwind, und Mantine steuert Verhalten
//     bei: Fokusfalle, Rollen, Tastatur.
//
// **Das ist die Grenze dieser Einführung**, und sie steht hier, damit niemand
// sie für eine Auslassung hält.
export const mantinethema = createTheme({
  fontFamily: 'var(--font-text)',
  fontFamilyMonospace: 'ui-monospace, SFMono-Regular, Menlo, monospace',
  headings: { fontFamily: 'var(--font-anzeige)' },
  defaultRadius: 'md',

  // **Kein eigener Fokusring.** `stil.css` setzt ihn für die ganze Anwendung,
  // und `internal/werkzeuge/fokusring_test.go` hält ihn; ein zweiter aus
  // Mantine läge darüber und wäre an keiner Stelle gewacht.
  focusRing: 'auto',
})

// useMantineschema hält Mantines Farbschema beim aufgelösten Erscheinungsbild
// — Gegenprobe A03 im Gast unter macOS 15.2
// (#2217):
// ohne ihn stand Mantine im dunklen Erscheinungsbild auf »light«. Sichtbar
// wurde das nicht — die Überlagerungen sind eigen gestaltet, Text und Grund
// laufen über `mantinemerkmale.css` —, aber alles, was Mantine selbst nach dem
// Schema wählt, nähme die helle Palette.
//
// **Hier und nicht im Gestaltungsraum**, weil Mantine an genau dieser Stelle
// angeschlossen ist (`docs/ABHAENGIGKEITEN.md`, ein Wächter hält es). **Mit
// Vergleich**: `setColorScheme` ist nicht beständig, und ein Effekt, der es
// als Abhängigkeit trägt, liefe ohne ihn im Kreis.
export function useMantineschema(dunkel: boolean): void {
  const { colorScheme, setColorScheme } = useMantineColorScheme()

  useLayoutEffect(() => {
    const gewuenscht = dunkel ? 'dark' : 'light'
    if (colorScheme !== gewuenscht) {
      setColorScheme(gewuenscht)
    }
  }, [dunkel, colorScheme, setColorScheme])
}
