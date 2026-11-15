// Die Handgriffe des Arbeitswegslaufs — warten, klicken, wählen, tippen,
// Kürzel drücken, Tabstelle zählen, Bindung rufen.
//
// **Eigene Datei, weil sie eine eigene Frage beantwortet** (§2):
// `arbeitswege.mjs` sagt, **welcher Weg** gemessen wird, hier steht, **womit**.
// Werkstück U1 von Phase 5.1
// (#1412).
//
// # Warum nicht `kernweggriffe.mjs`
//
// Dessen Griffe kennen **Kartenfelder** — `insFeld` sucht ein `input` mit
// `aria-label`, und das ist auf der Kartenlasche richtig. Die Wege hier führen
// über Auswahlfelder, Reiter, Textflächen und Tastenkürzel; ein Griff, der
// jedes Element über einen **Ausdruck** findet, deckt alle. Die Zählung ist
// dieselbe: **ein Klick, ein Tastendruck, ein Tippen, gleich wie lang.**
//
// # Zeit wird gemessen, und das ist eine Entscheidung
//
// `kernwegmass.mjs` misst keine Zeit — „eine Dauer wäre eine Aussage über
// diesen Rechner". Das stimmt, und #1412 verlangt sie trotzdem: *„Ladezeiten
// … festlegen."* Also steht sie hier **mit ihrer Reichweite** (§10): dieser
// Rechner, dieser Tag, Auflösung 50 ms, und die Zahl im Dokument trägt beides.

import { schicke, warte, werte } from "./abzuegecdp.mjs";
import { tabulatorerreichbar } from "./tabstellen.mjs";

// knopfMit gibt den Ausdruck für einen Knopf mit genau dieser Aufschrift.
export function knopfMit(aufschrift) {
  return `[...document.querySelectorAll('button')].find((b) => b.textContent.trim() === ${JSON.stringify(aufschrift)})`;
}

// reiterMit gibt den Ausdruck für einen Reiter (role="tab") mit dieser
// Aufschrift.
export function reiterMit(aufschrift) {
  return `[...document.querySelectorAll('[role="tab"]')].find((b) => b.textContent.trim() === ${JSON.stringify(aufschrift)})`;
}

// feldUnter gibt den Ausdruck für das Eingabefeld unter einer Beschriftung —
// das erste `input`, `select` oder `textarea` in dem `label`, dessen Text so
// beginnt.
//
// **`label.control` zuerst**: ein `Pflichtfeld` beschriftet über `htmlFor`,
// und dann liegt das Feld **neben** dem `label`, nicht darin. Gefunden am
// 2026-09-08, als der Deckweg sein Namensfeld nicht fand.
export function feldUnter(beschriftung) {
  // **Genau vor ungefähr**: steht der Navigator (eingeklappte Seitenleiste,
  // #1490), gibt es mehr Beschriftungen, und „Set" träfe sonst „Setname".
  return `(() => { const alle = [...document.querySelectorAll('label')]; const l = alle.find((x) => x.textContent.trim() === ${JSON.stringify(beschriftung)}) ?? alle.find((x) => x.textContent.trim().startsWith(${JSON.stringify(beschriftung)})); return l ? (l.control ?? l.querySelector('input, select, textarea')) : null; })()`;
}

// knopfBenannt gibt den Ausdruck für einen Knopf mit dieser Bezeichnung für
// das Vorleseprogramm — `aria-label`. Ein `Knopf` mit `bezeichnung` trägt
// sichtbar etwas Kürzeres („Hierhin wechseln") als das, was ihn eindeutig
// macht („Zu Ich · Turnier wechseln").
export function knopfBenannt(bezeichnung) {
  return `[...document.querySelectorAll('button')].find((b) => b.getAttribute('aria-label') === ${JSON.stringify(bezeichnung)})`;
}

// handgriffe baut die Griffe samt ihrem Zähler.
export function handgriffe() {
  let handlungen = 0;

  // bis wartet auf eine Bedingung in der Seite — und gibt zurück, wie lange
  // sie gebraucht hat. **Alle 50 ms**, das ist die Auflösung jeder Dauer.
  async function bis(ws, ausdruck, was, frist = 30000) {
    const beginn = Date.now();

    while (Date.now() - beginn < frist) {
      if ((await werte(ws, ausdruck)) === true) {
        return Date.now() - beginn;
      }

      await warte(50);
    }

    // **Mit dem, was stattdessen dasteht.** Ein „kam nicht" ohne den Text
    // der Seite ist ein Rätsel; mit ihm ist es ein Befund.
    const stattdessen = String(
      await werte(ws, `(document.querySelector('main')?.textContent ?? '').replace(/\\s+/g, ' ').slice(0, 1500)`),
    );

    throw new Error(`${was} kam in ${String(frist / 1000)} Sekunden nicht — die Seite zeigt: „${stattdessen}"`);
  }

  // bisRuhe wartet, bis die Seite sich nicht mehr ändert — Zusehen zählt
  // nicht als Handlung.
  async function bisRuhe(ws, versuche = 30) {
    const zaehlen = async () =>
      Number(await werte(ws, `document.querySelectorAll('*').length`));

    let vorher = await zaehlen();

    for (let versuch = 0; versuch < versuche; versuch += 1) {
      await warte(300);

      const jetzt = await zaehlen();
      if (jetzt === vorher) {
        return;
      }

      vorher = jetzt;
    }

    throw new Error("die Seite kam nicht zur Ruhe");
  }

  // klick klickt das Element, das der Ausdruck gibt — und zählt.
  //
  // **Mit Fokus davor**, weil ein Mausklick das auch tut; ein Eingabefeld
  // daneben schreibt seinen Wert erst beim Verlassen.
  async function klick(ws, ausdruck, was) {
    handlungen += 1;

    const getroffen = await werte(
      ws,
      `(() => { const e = ${ausdruck}; if (!e) return false; e.focus(); e.click(); return true; })()`,
    );

    if (getroffen !== true) {
      throw new Error(`nichts zu klicken: ${was}`);
    }
  }

  // insFeld setzt den Fokus in ein Feld — so lange, bis er hält — und zählt.
  async function insFeld(ws, ausdruck, was) {
    handlungen += 1;

    for (let versuch = 0; versuch < 40; versuch += 1) {
      const sitzt = await werte(
        ws,
        `(() => { const e = ${ausdruck}; if (!e) return false; e.focus(); return document.activeElement === e; })()`,
      );

      if (sitzt === true) {
        return;
      }

      await warte(250);
    }

    throw new Error(`der Fokus blieb nicht im Feld: ${was}`);
  }

  // tippe schreibt in das fokussierte Feld — **eine** Handlung, über
  // `Input.insertText`, weil ein Chrome ohne Fenster keine Zeichen über
  // `dispatchKeyEvent` annimmt (gemessen am 2026-08-30, `kernweggriffe.mjs`).
  async function tippe(ws, text) {
    handlungen += 1;

    await schicke(ws, "Input.insertText", { text });
  }

  // alles wählt den Inhalt des fokussierten Felds aus — ⌘A, **eine**
  // Handlung. Wer ein vorbelegtes Feld überschreibt, drückt sie; ein
  // `insertText` ohne sie hängt an, was schon dasteht — so entstand am
  // 2026-09-08 ein Deck des Spiels „magicprobe".
  async function alles(ws) {
    handlungen += 1;

    await werte(ws, `(() => { const e = document.activeElement; if (e && typeof e.select === 'function') e.select(); return true; })()`);
  }

  // waehle setzt ein Auswahlfeld auf einen Wert — und zählt.
  //
  // **Über den nativen Setter und ein `change`-Ereignis**, damit React die
  // Änderung sieht: ein direktes `.value =` an einem gesteuerten Feld
  // schreibt am Zustand vorbei.
  async function waehle(ws, ausdruck, wert, was) {
    handlungen += 1;

    const getroffen = await werte(
      ws,
      `(() => {
        const e = ${ausdruck};
        if (!e) return false;
        const setter = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set;
        setter.call(e, ${JSON.stringify(wert)});
        e.dispatchEvent(new Event('change', { bubbles: true }));
        return e.value === ${JSON.stringify(wert)};
      })()`,
    );

    if (getroffen !== true) {
      throw new Error(`nichts zu wählen: ${was}`);
    }
  }

  // taste drückt eine Taste **in** einem Element — und zählt.
  //
  // **Für ⏎ in der Kartenwahl** (U10): das Feld bekommt zuerst den Fokus,
  // sonst geht die Taste an den Rumpf. Der Fokus ist keine eigene Handlung —
  // er folgt der Wahl, die eine Zeile darüber gezählt wurde.
  async function taste(ws, ausdruck, zeichen, was) {
    handlungen += 1;

    const getroffen = await werte(
      ws,
      `(() => { const e = ${ausdruck}; if (!e) return false; e.focus(); return true; })()`,
    );

    if (getroffen !== true) {
      throw new Error(`nichts zu drücken: ${was}`);
    }

    for (const type of ["keyDown", "keyUp"]) {
      await schicke(ws, "Input.dispatchKeyEvent", {
        type,
        key: zeichen,
        code: zeichen === "Enter" ? "Enter" : zeichen,
        windowsVirtualKeyCode: zeichen === "Enter" ? 13 : 0,
        nativeVirtualKeyCode: zeichen === "Enter" ? 13 : 0,
      });
    }
  }

  // kommando drückt ⌘ mit einer Taste — und zählt. **modifiers 4 = Meta.**
  async function kommando(ws, zeichen, code, vk) {
    handlungen += 1;

    for (const type of ["keyDown", "keyUp"]) {
      await schicke(ws, "Input.dispatchKeyEvent", {
        type,
        modifiers: 4,
        key: zeichen,
        code,
        windowsVirtualKeyCode: vk,
        nativeVirtualKeyCode: vk,
      });
    }
  }

  // tabstelle sagt, die wievielte Tabstelle ein Element ist — von der ersten
  // erreichbaren der Seite an gezählt, 1-basiert; -1, wenn es nicht
  // erreichbar ist.
  //
  // **Das ist die Zahl, die das Issue meint**: wie oft jemand ⇥ drückt, der
  // ohne Maus zum ersten Bedienelement des Weges will.
  // **Verstecktes zählt nicht.** Die eingeklappte Seitenleiste bleibt mit
  // `hidden` im Baum (#1490); ihre Knöpfe erreicht niemand mit Tab.
  async function tabstelle(ws, ausdruck) {
    return Number(
      await werte(
        ws,
        `(() => { const e = ${ausdruck}; const alle = [...document.querySelectorAll(${JSON.stringify(tabulatorerreichbar)})].filter((x) => x.closest('[hidden]') === null); const i = alle.indexOf(e); return i < 0 ? -1 : i + 1; })()`,
      ),
    );
  }

  // rufe ruft eine Bindung über die laufende Brücke — **kein** Klick, keine
  // Handlung: so legt ein Weg an, was er braucht, und räumt es weg.
  async function rufe(ws, name, ...args) {
    return JSON.parse(
      await werte(
        ws,
        `(async () => JSON.stringify(await window.go.app.Bindungen.${name}(${args.map((a) => JSON.stringify(a)).join(", ")})))()`,
      ),
    );
  }

  // zaehleExemplare sagt, wie viele Exemplare die Sammlung hat — **nur die
  // Zahl**. Die ganze Antwort sind bei 10.000 Exemplaren 4 MB, und die kommen
  // über `Runtime.evaluate` mit `returnByValue` nicht an: der Lauf stand am
  // 2026-09-08 85 Sekunden ohne Fehler, während die Bindung selbst 150 ms
  // braucht. Was gezählt wird, wird in der Seite gezählt.
  async function zaehleExemplare(ws) {
    return Number(
      await werte(
        ws,
        `(async () => (await window.go.app.Bindungen.Sammlung()).exemplare?.length ?? 0)()`,
      ),
    );
  }

  return {
    griffe: { bis, bisRuhe, klick, insFeld, tippe, alles, waehle, taste, kommando, tabstelle, rufe, zaehleExemplare },
    gezaehlt: () => handlungen,
    zuruecksetzen: () => {
      handlungen = 0;
    },
  };
}
