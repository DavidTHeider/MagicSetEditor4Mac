// **Was der Kernweg „anlegen, füllen, weiter" kostet**
// (#1290).
//
// # Wozu
//
// Die Abnahme von Phase 4.3 verlangt eine Zahl für den **ganzen** Weg:
// *„Eine neue Karte mit denselben Grundangaben wie die vorige braucht höchstens
// X Nutzerhandlungen und keinen Seitenwechsel."*
//
// `serienmass.mjs` zählt, was auf einer Seite **steht**; dieser Lauf zählt, was
// jemand **tut**.
//
// ```bash
// node scripts/kernwegmass.mjs
// ```
//
// # Er legt sich sein eigenes Set an — und räumt es weg
//
// **Weil er schreibt.** Der erste Anlauf lief gegen den echten Bestand und
// hinterließ zwei leere Karten in `Vanguard.mse-set` (2026-08-30, gelöscht).
// Ein Messlauf, der schreibt, fasst keinen fremden Bestand an.
//
// # Er wartet auf Elemente, nicht auf Fristen
//
// **Und das ist der Grund, warum der erste Anlauf nichts maß.** Er fokussierte
// das Namensfeld nach einer festen Wartezeit — das Feld war noch nicht da, das
// Tippen ging ins Leere, und die Karte entstand leer. Gezählt wurden trotzdem
// vier Handlungen, und die Zahl sah aus wie eine Messung.
//
// # Was eine Nutzerhandlung ist
//
// Ein Klick oder ein Tastendruck. **Tippen zählt als eine**, gleich wie lang
// der Text ist: wer einen Namen eingibt, tut eine Sache.
//
// # Was er **nicht** misst
//
//   - **Nicht die Zeit.** Eine Dauer wäre eine Aussage über diesen Rechner.
//   - **Nicht die übrigen Kernwege** — Suchen/Ersetzen, Massenänderung, Bild
//     setzen, Ausgeben, Sammlung, Preis. Sie stehen im Issue und fehlen hier.

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";
import { handgriffe } from "./kernweggriffe.mjs";
import { eingeklapptGewuenscht, seitenleisteAusklappen, seitenleisteEinklappen } from "./seitenleistenmass.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

// **Spiel und Stil kommen von der Kommandozeile**
// (#1412),
// damit der Lauf auch gegen die Saat (`probe`, `probe-schlicht`) geht und
// nicht den bezogenen M15-Pack voraussetzt. Die Vorgabe `magic`/`magic-m15`
// bleibt, damit jede frühere Zahl ihren Weg behält.
const spiel = process.argv[3] ?? "magic";
const stil = process.argv[4] ?? "magic-m15";

// **Der Setname trägt eine Kennung dieses Laufs** — Befund 3 des Nachlaufs
// über Phase 4.3. Zwei Läufe zugleich wählten sonst denselben Namen aus der
// Liste, und einer vermäße das Set des anderen: Setnamen sind im Bestand nicht
// eindeutig (`0002_setbestand.sql`).
const setname = `Messlauf Kernweg ${String(process.pid)}`;

// erwartet ist die Abnahme, gegen die dieser Lauf prüft.
//
// **Er gibt sie nicht nur aus, er weist ab** (Befund 2). Ein Messlauf, der
// eine falsche Zahl druckt und mit 0 endet, ist eine Anzeige und keine
// Abnahme.
const erwartet = {
  handlungen: 4,
  karten: 2,
  namen: ["Blitzschlag", "Feuerball"],
};

const profil = await mkdtemp(join(tmpdir(), "kernwegmass-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

const { griffe, gezaehlt, zuruecksetzen } = handgriffe();
const { bis, bisRuhe, klick, insFeld, tippe, taste } = griffe;

let ws = null;
let setID = null;
let fehlgeschlagen = false;

// **Ein Abbruch räumt trotzdem auf** — Befund 3 des Nachlaufs über Phase 4.3.
// Ohne diese beiden Zeilen liefe das `finally` bei ⌃C nicht, und das Messset
// bliebe im Bestand stehen.
for (const zeichen of ["SIGINT", "SIGTERM"]) {
  process.on(zeichen, () => {
    process.stdout.write(`\n  abgebrochen — das Messset bleibt womöglich stehen\n`);
    process.exitCode = 1;
  });
}

try {
  ws = await verbinden(profil);

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: 1280,
    height: 800,
    deviceScaleFactor: 1,
    mobile: false,
  });

  await schicke(ws, "Page.navigate", { url: adresse });
  await bis(ws, `typeof window.go?.app?.Bindungen === 'object'`, "die Brücke");

  const angelegt = JSON.parse(
    await werte(
      ws,
      `(async () => JSON.stringify(await window.go.app.Bindungen.SetAnlegen(${JSON.stringify(setname)}, ${JSON.stringify(spiel)}, ${JSON.stringify(stil)})))()`,
    ),
  );

  setID = String(angelegt.setID ?? angelegt.id ?? "");
  if (setID === "") {
    throw new Error(`kein Set angelegt: ${JSON.stringify(angelegt)}`);
  }

  // **Aufbau zählt nicht.** Wer in Serie arbeitet, hat sein Set schon offen.
  await schicke(ws, "Page.navigate", { url: adresse });
  await bis(ws, `typeof window.go?.app?.Bindungen === 'object'`, "die Brücke");
  // **Mit eingeklappter Leiste**, wenn gewünscht (#1490).
  if (eingeklapptGewuenscht) {
    await bis(ws, `document.querySelector('main') !== null`, "die Startseite");
    process.stdout.write(`  ${await seitenleisteEinklappen(ws)}\n`);
  }
  await werte(
    ws,
    `(() => { const e=[...document.querySelectorAll('button')].find((b)=>b.textContent.trim().startsWith('Sets')); e.click(); return true; })()`,
  );
  await bis(
    ws,
    `[...document.querySelectorAll('button')].some((b)=>b.textContent.trim().startsWith(${JSON.stringify(setname)}))`,
    "das Set in der Liste",
  );
  await werte(
    ws,
    `(() => { const e=[...document.querySelectorAll('button')].find((b)=>b.textContent.trim().startsWith(${JSON.stringify(setname)})); e.click(); return true; })()`,
  );
  await bis(
    ws,
    `[...document.querySelectorAll('button')].some((b)=>b.textContent.trim()==='Karten')`,
    "die Kartenlasche",
  );
  await werte(
    ws,
    `(() => { const e=[...document.querySelectorAll('button')].find((b)=>b.textContent.trim()==='Karten'); e.click(); return true; })()`,
  );
  await bis(
    ws,
    `[...document.querySelectorAll('button')].some((b)=>b.textContent.trim()==='Karte anlegen')`,
    "der Knopf »Karte anlegen«",
  );

  const lascheVorher = await werte(
    ws,
    `document.querySelector('[role="tab"][aria-selected="true"]')?.textContent?.trim() ?? null`,
  );

  // ── Erste Karte: anlegen und füllen ──────────────────────────────────
  await klick(ws, "Karte anlegen");
  await bis(
    ws,
    `[...document.querySelectorAll('input')].some((x)=>x.getAttribute('aria-label')==='name')`,
    "das Formular der ersten Karte",
  );
  await bisRuhe(ws);

  await insFeld(ws, "name");
  await tippe(ws, "Feuerball");
  await bis(
    ws,
    `[...document.querySelectorAll('input')].find((x)=>x.getAttribute('aria-label')==='name')?.value === 'Feuerball'`,
    "der getippte Name",
  );


  // ── Der gemessene Weg beginnt hier ───────────────────────────────────
  zuruecksetzen();

  await klick(ws, "Karte anlegen");
  await bis(
    ws,
    `[...document.querySelectorAll('input')].some((x)=>x.getAttribute('aria-label')==='name')`,
    "das Formular der zweiten Karte",
  );
  await bisRuhe(ws);

  await insFeld(ws, "name");
  await tippe(ws, "Blitzschlag");
  await bis(
    ws,
    `[...document.querySelectorAll('input')].find((x)=>x.getAttribute('aria-label')==='name')?.value === 'Blitzschlag'`,
    "der zweite getippte Name",
  );

  // **Welche Karte die vorige ist, entscheidet die Sortierung** — nicht die
  // Reihenfolge des Anlegens. `Blitzschlag` steht vor `Feuerball`, also führt
  // ⇞ von der zweiten Karte zur ersten und nicht zur zuletzt angelegten.
  //
  // Zugesichert ist deshalb: **der Name wechselt.** Welcher es ist, sagt die
  // Ordnung des Spiels, und die prüfen die Läufe daneben.
  const vorBlaettern = await werte(
    ws,
    `[...document.querySelectorAll('input')].find((x)=>x.getAttribute('aria-label')==='name')?.value ?? ''`,
  );

  await taste(ws, "PageUp", "PageUp", 33);
  await bis(
    ws,
    `([...document.querySelectorAll('input')].find((x)=>x.getAttribute('aria-label')==='name')?.value ?? '') !== ${JSON.stringify(vorBlaettern)}`,
    "die andere Karte nach ⇞",
  );

  const lascheNachher = await werte(
    ws,
    `document.querySelector('[role="tab"][aria-selected="true"]')?.textContent?.trim() ?? null`,
  );

  // **Dem Schreibweg Zeit lassen.** Der Wert geht beim Verlassen des Feldes
  // hinaus und ist eine Antwort später im Bestand; wer sofort liest, liest den
  // Stand davor.
  await warte(2500);

  const stand = JSON.parse(
    await werte(
      ws,
      `(async () => { const k = await window.go.app.Bindungen.Kartenzeilen(${JSON.stringify(setID)});
        const zeilen = k.zeilen ?? k.karten ?? [];
        return JSON.stringify({ karten: zeilen.length, namen: zeilen.map((z)=>z.name) }); })()`,
    ),
  );

  process.stdout.write(
    `Kernweg „anlegen, füllen, weiter"\n` +
      `  Nutzerhandlungen  ${String(gezaehlt())}\n` +
      `  Seitenwechsel     ${lascheVorher === lascheNachher ? "keiner" : `${String(lascheVorher)} → ${String(lascheNachher)}`}\n` +
      `  Karten im Set     ${String(stand.karten)} — ${stand.namen.join(", ")}\n`,
  );

  // **Und jetzt die Abnahme** (Befund 2 des Nachlaufs über Phase 4.3).
  const befunde = [];

  if (gezaehlt() !== erwartet.handlungen) {
    befunde.push(`${String(gezaehlt())} Handlungen statt ${String(erwartet.handlungen)}`);
  }

  if (lascheVorher !== lascheNachher) {
    befunde.push(`ein Seitenwechsel: ${String(lascheVorher)} → ${String(lascheNachher)}`);
  }

  if (stand.karten !== erwartet.karten) {
    befunde.push(`${String(stand.karten)} Karten statt ${String(erwartet.karten)}`);
  }

  // **Die Namen, sortiert** — welche Karte vorn steht, entscheidet die Ordnung
  // des Spiels, und die prüfen die Läufe daneben. Was hier zählt: **beide
  // tragen ihren Namen**. Ohne diese Zeile ginge ein Datenverlust wie der vom
  // 2026-08-30 unbemerkt durch, und der Lauf druckte trotzdem „4".
  const gefunden = [...stand.namen].sort().join(", ");
  const gewollt = [...erwartet.namen].sort().join(", ");

  if (gefunden !== gewollt) {
    befunde.push(`die Namen lauten „${gefunden}" statt „${gewollt}"`);
  }

  if (befunde.length > 0) {
    fehlgeschlagen = true;

    for (const befund of befunde) {
      process.stdout.write(`  ✗ ${befund}\n`);
    }
  }
} finally {
  if (ws !== null && setID !== null) {
    try {
      await schicke(ws, "Page.navigate", { url: adresse });
      await bis(ws, `typeof window.go?.app?.Bindungen === 'object'`, "die Brücke zum Aufräumen");
      // **Die Antwort wird gelesen** (Befund 3): ein Löschen, das scheitert,
      // gab vorher nur ein `ACHTUNG` und einen grünen Lauf.
      const geloescht = JSON.parse(
        await werte(
          ws,
          `(async () => JSON.stringify(await window.go.app.Bindungen.SetLoeschen(${JSON.stringify(setID)}, ${JSON.stringify(setname)})))()`,
        ),
      );

      if (geloescht.fehler !== null && geloescht.fehler !== undefined) {
        fehlgeschlagen = true;
        process.stdout.write(`  ✗ das Messset ließ sich nicht löschen: ${JSON.stringify(geloescht.fehler)}\n`);
      }

      const uebrig = await werte(
        ws,
        `(async () => { const s = await window.go.app.Bindungen.Sets();
          return String((s.sets ?? []).filter((e)=>e.name===${JSON.stringify(setname)}).length); })()`,
      );

      if (uebrig === "0") {
        process.stdout.write(`  aufgeräumt        das Messset ist weg\n`);
      } else {
        fehlgeschlagen = true;
        process.stdout.write(`  ✗ ${String(uebrig)} Messsets sind übrig\n`);
      }
    } catch (fehler) {
      fehlgeschlagen = true;
      process.stdout.write(`  ✗ Aufräumen fehlgeschlagen: ${String(fehler)}\n`);
    }
  }

  if (ws !== null && eingeklapptGewuenscht) {
    try {
      await seitenleisteAusklappen(ws);
    } catch (grund) {
      process.stdout.write(`  ✗ die Seitenleiste blieb zu: ${String(grund)}\n`);
    }
  }
  chrome.kill();
  await warte(200);
  // **Mit Wiederholung**, wie in `arbeitswegmass.mjs`: Chrome schreibt beim
  // Beenden noch in sein Profil, und ein `rmdir` darauf endet mit ENOTEMPTY.
  await rm(profil, { recursive: true, force: true, maxRetries: 5 });
}

if (fehlgeschlagen) {
  process.exit(1);
}
