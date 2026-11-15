// **Die Kernwege mit eingeklappter Seitenleiste** — Werkstück U14 von Phase 5.1
// (#1490).
//
// `kernwegmass.mjs` und `arbeitswegmass.mjs` fahren ihre Wege gewöhnlich mit
// stehender Leiste. Mit `MSE_SEITENLEISTE=zu` klappen sie sie vorher über
// ⌥⌘S ein — so wie ein Mensch es täte —, nennen die Breite des Hauptbereichs
// vorher und nachher, und klappen sie am Ende wieder aus: die Leiste ist eine
// Einstellung auf Geräteebene und bliebe sonst im Messzuhause zu.
import { schicke, warte, werte } from "./abzuegecdp.mjs";

export const eingeklapptGewuenscht = process.env.MSE_SEITENLEISTE === "zu";

const hauptbreite = `String(Math.round(document.querySelector('main')?.getBoundingClientRect().width ?? 0))`;
// **`hidden` und nicht „im Baum“**: die Leiste bleibt eingeklappt im DOM, nur
// versteckt — wer nach `<nav>` fragt, sieht sie immer.
// **Als Zeichenkette**: `werte` gibt den Wert zurück, wie er ist, und ein
// `false === "false"` ist nie wahr — der erste Anlauf wartete deshalb auf
// einen Zustand, der längst da war.
const leisteSteht = `String(!document.getElementById('seitenleiste')?.hidden)`;

// wahlBefehlS schickt ⌥⌘S — Alt ist Bit 1, Meta Bit 4 der CDP-Modifikatoren.
async function wahlBefehlS(ws) {
  for (const type of ["keyDown", "keyUp"]) {
    await schicke(ws, "Input.dispatchKeyEvent", {
      type, key: "s", code: "KeyS", windowsVirtualKeyCode: 83, nativeVirtualKeyCode: 83, modifiers: 5,
    });
  }
}

// schalten schickt ⌥⌘S und wartet, bis die Leiste steht oder weg ist.
//
// **Mit Wiederholung.** Direkt nach dem Laden hängt der Tastenhorcher des
// Rahmens noch nicht — ein `useEffect` läuft nach dem Zeichnen —, und ein
// Tastendruck davor geht verloren (§3: „ein verlorener Tastendruck kommt
// nicht wieder"). Gefunden am 2026-09-09: der erste Anlauf schickte einmal
// und wartete zehn Sekunden auf nichts.
async function schalten(ws, steht) {
  for (let versuch = 0; versuch < 5; versuch += 1) {
    await warte(300);
    await wahlBefehlS(ws);
    const frist = Date.now() + 2000;
    while (Date.now() < frist) {
      if ((await werte(ws, leisteSteht)) === (steht ? "true" : "false")) {
        return;
      }
      await warte(50);
    }
  }
  throw new Error(`die Seitenleiste ist nach ⌥⌘S nicht ${steht ? "auf" : "zu"}`);
}

// seitenleisteEinklappen klappt ein und gibt die Zeile für den Bericht.
export async function seitenleisteEinklappen(ws) {
  // **Erst stehend messen.** Die Leiste ist eine Geräteeinstellung; ein
  // früherer Lauf kann sie zu hinterlassen haben.
  if ((await werte(ws, leisteSteht)) === "false") {
    await schalten(ws, true);
    await warte(200);
  }
  const vorher = await werte(ws, hauptbreite);
  await schalten(ws, false);
  await warte(200);
  const nachher = await werte(ws, hauptbreite);

  return `Seitenleiste eingeklappt (⌥⌘S): Hauptbereich ${vorher} → ${nachher} Punkte breit`;
}

// seitenleisteAusklappen stellt den Zustand wieder her — Bestenfalls-Arbeit.
export async function seitenleisteAusklappen(ws) {
  if ((await werte(ws, leisteSteht)) === "true") {
    return;
  }
  await schalten(ws, true);
}
