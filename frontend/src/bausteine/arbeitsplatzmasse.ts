// **Maße und Klassen des Arbeitsplatzes** — herausgelöst aus
// `Arbeitsplatz.tsx`, als der Anlass für den Inspector (UXR-17, #1925) ihn
// über die Zeilengrenze aus §2 brachte. Dort, was er tut; hier, wie hoch die
// Spalten stehen und wie die über dem Zentrum aussehen.

// **Die Spalten reichen bis an den Fensterrand, und jede rollt für sich** —
// UX-15 (#1771).
// Gemessen am 2026-09-24 in WebKit: der Vorlagenbaum mit 536 Paketen machte
// die Seite 60.317 Punkte hoch, und wer unten eine Datei wählte, sah sie
// vierzig Bildschirme weiter oben aufgehen. Die Artboards geben jeder Spalte
// die Höhe des Fensters.
//
// Spaltenrand ist der Innenrand unter dem Inhalt — `pb-8` an `main` in
// `Rahmen.tsx`, derselbe wie beim Tisch.
export const Spaltenrand = 32

// Spaltenmindesthoehe ist die Höhe, unter die eine Spalte nicht fällt: in
// einem flachen Fenster rollt die Seite, statt eine Spalte zu einem Schlitz
// zu machen. Im Mindestfenster (1.024 × 700) greift sie nicht.
export const Spaltenmindesthoehe = 320

// **Bei Enge geht die weichende Spalte über dem Zentrum auf** — UX-14.AK-03
// (#1686),
// Artboard *Enge*: *„nichts ist gesperrt und nichts abgeschnitten"*. Bis zum
// 2026-09-25 war ihr Knopf dann angehalten, und bei 1.024 × 700 war die
// Kartenliste gar nicht zu erreichen. **Eine Kante statt eines Schattens**,
// wie die Schubladen (Vermerk 0065, Festlegung 3).
export const ueberNavigator =
  'absolute inset-y-0 left-0 z-10 flex min-h-0 w-60 flex-col overflow-y-auto border-r-2 border-kante bg-flaeche'
export const ueberInspector =
  'absolute inset-y-0 right-0 z-10 flex min-h-0 w-80 flex-col overflow-y-auto border-l-2 border-kante bg-flaeche'

// wurzelschrift ist die Schriftgröße des Wurzelelements in Punkten — die
// Größe, auf die sich jedes `rem` bezieht (`stil.css`, `--schriftgroesse`).
// Ohne lesbaren Wert gilt die Vorgabe von 16.
export function wurzelschrift(): number {
  const gelesen = parseFloat(
    getComputedStyle(document.documentElement).fontSize,
  )

  return Number.isFinite(gelesen) && gelesen > 0 ? gelesen : 16
}

// breite gibt die Breite aus dem Gedächtnis als Stil — und ohne Gedächtnis
// nichts, damit die Klasse aus 0076 gilt.
export function breite(
  punkte: number | undefined,
): { width: number } | undefined {
  return punkte === undefined ? undefined : { width: punkte }
}

// zentrumMindestens ist, was dem Zentrum bleibt, solange der Navigator steht
// — die 480 aus Vermerk 0076. Seit AP-12 (#2369) hält sie auch der Trenner:
// breiter zieht ihn niemand, als dass das Zentrum darunter fiele.
export const zentrumMindestens = 480

// engeGrenze ist die Breite **dieses Bausteins**, unter der der Navigator von
// selbst einklappt: Navigator 240 + Zentrum 480 + Inspector 320 + zweimal
// der Abstand von 16.
//
// **Hier stand 1.200 — die „Inhaltsbreite" aus 0076**, und die ist am
// Inhaltsbereich des Rahmens gemessen, nicht an diesem Baustein: Rahmen und
// Setkopf nehmen 114 Punkte Rand, und bei 1.440 Punkten Fenster misst der
// Arbeitsplatz **1.086**. Mit 1.200 stand der Navigator auf keinem Fenster,
// das es gibt — gefunden in der Sichtprüfung von U5 (#1416). Der Baustein
// misst sich selbst, also zählt die Summe seiner Teile.
export const engeGrenze = 240 + zentrumMindestens + 320 + 2 * 16

// navigatorPasst sagt, ob bei dieser Breite Platz für den Navigator ist.
//
// **In der Wurzelschrift gerechnet** — G20 aus dem Zweitblick über Phase
// 6.3 (#1664): die Spalten stehen in `rem` (`w-60`, `w-80`, `gap-4`) und
// wachsen mit „Sehr groß"; die Grenze stand in Punkten, und bei 125 % blieb
// dem Zentrum an ihr 332 statt 600. `engeGrenze` gilt für 16 Punkte.
export function navigatorPasst(breite: number, wurzelschrift = 16): boolean {
  return breite >= (engeGrenze * wurzelschrift) / 16
}
