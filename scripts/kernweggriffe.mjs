// Die Handgriffe des Kernwegslaufs — warten, klicken, tippen, drücken.
//
// **Eigene Datei, weil sie eine eigene Frage beantwortet** (§2):
// `kernwegmass.mjs` sagt, **welcher Weg** gemessen wird, hier steht, **womit**.
// Zusammen gingen sie über die Zeilengrenze, als der Nachlauf über Phase 4.3
// die Abnahme verlangte
// (#1290).
//
// **Sie zählen selbst.** Was eine Nutzerhandlung ist, entscheidet der Griff —
// ein Klick, ein Tastendruck, ein Tippen, gleich wie lang.

import { writeFile } from "node:fs/promises";
import { schicke, warte, werte } from "./abzuegecdp.mjs";

// handgriffe baut die Griffe samt ihrem Zähler.
export function handgriffe() {
  let handlungen = 0;

  // bis wartet auf eine Bedingung in der Seite — mit einer Frist.
  async function bis(ws, ausdruck, was, versuche = 40) {
    for (let versuch = 0; versuch < versuche; versuch += 1) {
      if ((await werte(ws, ausdruck)) === true) {
        return;
      }

      await warte(500);
    }

    // **Und ein Bild davon**, wenn `MSE_ABZUG_BEI_FEHLER` einen Pfad nennt —
    // ein Text von 600 Zeichen sagt nicht, ob der Navigator steht.
    if (process.env.MSE_ABZUG_BEI_FEHLER) {
      const { data } = await schicke(ws, "Page.captureScreenshot", { format: "png" });
      await writeFile(process.env.MSE_ABZUG_BEI_FEHLER, Buffer.from(data, "base64"));
    }
    // **Was stattdessen dastand**, damit ein roter Lauf sagt, wo er hing.
    const stand = await werte(
      ws,
      `(document.querySelector('main')?.textContent ?? '').replace(/\\s+/g, ' ').slice(0, 600)`,
    );
    throw new Error(`${was} kam in ${String(versuche / 2)} Sekunden nicht — es stand da: ${String(stand)}`);
  }

  // bisRuhe wartet, bis die Seite sich nicht mehr ändert.
  //
  // **Weil ein Mensch das auch tut.** Nach »Karte anlegen« baut der Editor sein
  // Formular neu, während die Werte laden; wer in dieser Zeit in ein Feld klickt,
  // verliert den Fokus mit dem nächsten Neuaufbau. Gemessen am 2026-08-30: ohne
  // dieses Warten kam der Fokus **nie** an.
  //
  // Das Warten zählt **nicht** als Handlung — Zusehen ist keine.
  async function bisRuhe(ws, versuche = 30) {
    const zaehlen = async () =>
      Number(await werte(ws, `document.querySelectorAll('*').length`));

    let vorher = await zaehlen();

    for (let versuch = 0; versuch < versuche; versuch += 1) {
      await warte(500);

      const jetzt = await zaehlen();
      if (jetzt === vorher) {
        return;
      }

      vorher = jetzt;
    }

    throw new Error("die Seite kam nicht zur Ruhe");
  }

  // klick drückt einen Knopf über seine Aufschrift — und zählt.
  //
  // **Er setzt vorher den Fokus**, weil ein Mausklick das auch tut. Ohne das
  // verließ das Eingabefeld daneben nie den Fokus, sein Wert wurde nicht
  // geschrieben, und die Messung zählte einen Weg, der nichts speichert. Der
  // erste Anlauf gab deshalb zwei Karten **ohne Namen** aus und nannte den Weg
  // trotzdem „anlegen, füllen, weiter".
  async function klick(ws, aufschrift) {
    handlungen += 1;

    const getroffen = await werte(
      ws,
      `(() => {
        const e = [...document.querySelectorAll('button')]
          .find((b) => b.textContent.trim() === ${JSON.stringify(aufschrift)});
        if (!e) return false;
        e.focus();
        e.click();
        return true;
      })()`,
    );

    if (getroffen !== true) {
      throw new Error(`kein Knopf ${aufschrift}`);
    }
  }

  // insFeld setzt den Fokus in ein Kartenfeld — und zählt.
  //
  // **Der Fokus wird so lange gesetzt, bis er hält.** Der Editor baut sein
  // Formular neu, während die Werte nachladen; ein einmaliges `focus()` mit einer
  // Warteschleife dahinter traf gelegentlich das Feld, das gleich abgeräumt wird.
  // Der Lauf war damit **flatterhaft** — und ein flatterhaftes Messgerät ist
  // keines.
  //
  // **Gezählt wird trotzdem eine Handlung**: ein Mensch klickt einmal, wenn das
  // Formular dasteht.
  async function insFeld(ws, name) {
    handlungen += 1;

    for (let versuch = 0; versuch < 40; versuch += 1) {
      await werte(
        ws,
        `(() => {
          const f = [...document.querySelectorAll('input')]
            .find((x) => x.getAttribute('aria-label') === ${JSON.stringify(name)});
          if (!f) return false;
          f.focus();
          return true;
        })()`,
      );

      const sitzt = await werte(
        ws,
        `document.activeElement?.getAttribute('aria-label') === ${JSON.stringify(name)}`,
      );

      if (sitzt === true) {
        return;
      }

      await warte(500);
    }

    throw new Error(`der Fokus blieb nicht im Feld ${name}`);
  }

  // tippe schreibt in das fokussierte Feld — **eine** Handlung.
  //
  // **Über `Input.insertText` und nicht Zeichen für Zeichen.**
  //
  // Ein Chrome ohne Fenster hat keinen Fokus auf der Seite (`document.hasFocus()`
  // gibt `false`, gemessen am 2026-08-30), und `dispatchKeyEvent` mit
  // `type: "char"` kam damit nicht im Feld an. `insertText` schreibt in das
  // **fokussierte Element**, unabhängig davon, ob das Fenster vorn steht.
  //
  // **Es bleibt eine Handlung**: gemessen wird, was jemand tut, nicht wie viele
  // Tasten er dabei drückt.
  async function tippe(ws, text) {
    handlungen += 1;

    await schicke(ws, "Input.insertText", { text });
  }

  // taste drückt eine Taste — und zählt.
  async function taste(ws, key, code, vk) {
    handlungen += 1;

    for (const type of ["keyDown", "keyUp"]) {
      await schicke(ws, "Input.dispatchKeyEvent", {
        type,
        key,
        code,
        windowsVirtualKeyCode: vk,
        nativeVirtualKeyCode: vk,
      });
    }
  }

  return {
    griffe: { bis, bisRuhe, klick, insFeld, tippe, taste },
    gezaehlt: () => handlungen,
    zuruecksetzen: () => {
      handlungen = 0;
    },
  };
}
