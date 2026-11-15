// **Der CDP-Draht** — verbinden, Befehle schicken, Ausdrücke auswerten
// (#892,
// #1100).
//
// **Eine eigene Datei**, seit `abzuege.mjs` über die Zeilengrenze aus §2 ging.
// Der Schnitt folgt der Frage: dort steht, **welche** Seiten abgelichtet
// werden und wie man zu ihnen klickt — hier, **wie** man mit Chrome redet.
// Von der Anwendung weiß diese Datei nichts.

import { readFile } from "node:fs/promises";
import { join } from "node:path";

// warte hält an — als Funktionsdeklaration, damit sie vor ihrer Zeile im
// Quelltext benutzbar ist. Ein `const` wäre es nicht, und `verbinden` läuft
// zuerst.
export function warte(ms) {
  return new Promise((fertig) => setTimeout(fertig, ms));
}

// laufendeNummer zählt die CDP-Befehle.
//
// **Sie steht hier oben und nicht bei `schicke`.** Eine `let`-Bindung weiter
// unten ist im toten Bereich, solange die Zeile nicht gelaufen ist — und
// `verbinden` schickt den ersten Befehl vorher. Der Fehler stand zehn
// Sekunden lang als „noch nicht da" in einer Warteschleife.
let laufendeNummer = 0;

// verbinden wartet, bis Chrome seine Debugadresse hergibt.
//
// **Mit einer Frist.** Ein Warten ohne Ende macht aus einem gescheiterten
// Start ein hängendes Skript, und das ist der Fehler, den man nicht sieht.
export async function verbinden(profil) {
  const port = await portLesen(profil);

  let letzterGrund;

  for (let versuch = 0; versuch < 50; versuch += 1) {
    try {
      const antwort = await fetch(`http://127.0.0.1:${String(port)}/json/list`);
      const seiten = await antwort.json();
      const seite = seiten.find((s) => s.type === "page");

      if (seite) {
        const ws = new WebSocket(seite.webSocketDebuggerUrl);
        await new Promise((fertig) =>
          ws.addEventListener("open", fertig, { once: true }),
        );

        await schicke(ws, "Page.enable", {});
        await schicke(ws, "Runtime.enable", {});

        return ws;
      }
    } catch (grund) {
      // **Der letzte Grund wird behalten, nicht verschluckt.**
      //
      // Ein `catch {}` in einer Warteschleife macht aus jedem Fehler
      // „noch nicht da" — auch aus einem, der nie vergeht. Der erste Anlauf
      // lief so zehn Sekunden und meldete dann etwas Falsches.
      letzterGrund = grund;
    }

    await warte(200);
  }

  throw new Error(
    `Chrome gab seine Debugadresse nicht her — zuletzt: ${String(letzterGrund)}`,
  );
}

// portLesen wartet auf `DevToolsActivePort` im Profilverzeichnis.
//
// **Chrome schreibt ihn dorthin, sobald er lauscht** — die erste Zeile ist die
// Nummer. Das ist der einzige Weg, ihn bei `--remote-debugging-port=0` zu
// erfahren.
async function portLesen(profil) {
  for (let versuch = 0; versuch < 100; versuch += 1) {
    try {
      const inhalt = await readFile(join(profil, "DevToolsActivePort"), "utf8");
      const nummer = Number.parseInt(inhalt.split("\n")[0] ?? "", 10);

      if (Number.isInteger(nummer) && nummer > 0) {
        return nummer;
      }
    } catch {
      // noch nicht geschrieben
    }

    await warte(100);
  }

  throw new Error("Chrome schrieb keinen DevToolsActivePort");
}

// schicke sendet einen CDP-Befehl und wartet auf seine Antwort.
export function schicke(ws, methode, params) {
  laufendeNummer += 1;

  const id = laufendeNummer;

  return new Promise((fertig, scheitern) => {
    const horcher = (ereignis) => {
      const nachricht = JSON.parse(ereignis.data);

      if (nachricht.id !== id) {
        return;
      }

      ws.removeEventListener("message", horcher);

      if (nachricht.error) {
        scheitern(new Error(`${methode}: ${nachricht.error.message}`));

        return;
      }

      fertig(nachricht.result);
    };

    ws.addEventListener("message", horcher);
    ws.send(JSON.stringify({ id, method: methode, params }));
  });
}

// werte führt einen Ausdruck in der Seite aus und gibt sein Ergebnis.
export async function werte(ws, ausdruck) {
  const { result } = await schicke(ws, "Runtime.evaluate", {
    expression: ausdruck,
    awaitPromise: true,
    returnByValue: true,
  });

  return result.value;
}
