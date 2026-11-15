// **Was die vier übrigen Kernwege kosten** — Werkstück U1 von Phase 5.1
// (#1412).
//
// # Wozu
//
// Die Abnahme von #1412 verlangt für **jeden** Kernweg einen reproduzierbaren
// Ausgangsbestand, einen gemessenen Ausgangswert und ein numerisches Ziel,
// **bevor** ein Arbeitsplatz umgebaut wird. `kernwegmass.mjs` und
// `erstlaufmass.mjs` zählen die Karte; dieser Lauf zählt Paket, Sammlung,
// Deck und Profil — und er zählt drei Dinge je Weg:
//
//   - **Handlungen**: ein Klick, ein Tastendruck, ein Tippen, gleich wie lang.
//   - **Tabstellen**: die wievielte Tabstelle das erste Bedienelement des
//     Weges auf seiner Seite ist — was jemand ohne Maus drückt.
//   - **Ladezeit**: vom Kürzel bis die Seite dasteht, auf diesem Rechner, in
//     Schritten von 50 ms.
//
// ```bash
// node scripts/sammlungssaat.mjs /tmp/mse51heim 10000   # einmal
// HOME=/tmp/mse51heim make dev                          # in einem eigenen Fenster
// node scripts/arbeitswegmass.mjs
// ```
//
// # Er weist ab, was nicht der Zusage entspricht
//
// Jeder Weg trägt seine erwartete Handlungszahl. Wer einen Arbeitsplatz
// umbaut, senkt sie dort — und ein Lauf, der mehr zählt, endet rot statt mit
// einer Zahl, die aussieht wie eine Messung (Befund 2 des Nachlaufs über
// Phase 4.3).
//
// # Was er **nicht** misst
//
//   - **Nicht, ob die Seite gut aussieht.** Dafür gibt es die Sichtprüfung.
//   - **Nicht die Zeit auf einer anderen Maschine.** Eine Dauer ist eine
//     Aussage über diesen Rechner (0007); die Zahl im Dokument trägt Datum
//     und Gerät.
//   - **Nicht den Bestand dieses Rechners.** Er fährt gegen das verlegte
//     Zuhause und sagt zu Beginn, wie viele Exemplare dort liegen.

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte } from "./abzuegecdp.mjs";
import { handgriffe } from "./arbeitsweggriffe.mjs";
import { arbeitswege } from "./arbeitswege.mjs";
import { eingeklapptGewuenscht, seitenleisteAusklappen, seitenleisteEinklappen } from "./seitenleistenmass.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";
const nurWeg = process.argv[3] ?? "";

const profil = await mkdtemp(join(tmpdir(), "arbeitswegmass-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

const { griffe, gezaehlt, zuruecksetzen } = handgriffe();

let ws = null;
let fehlgeschlagen = false;
const zeilen = [];

// **Ein Abbruch räumt trotzdem auf** — sonst bliebe der Messbestand stehen.
for (const zeichen of ["SIGINT", "SIGTERM"]) {
  process.on(zeichen, () => {
    process.stdout.write(`\n  abgebrochen — der Messbestand bleibt womöglich stehen\n`);
    process.exitCode = 1;
  });
}

// spur schreibt einen Zwischenstand, wenn `MSE_SPUR` gesetzt ist — für die
// Frage, wo ein Lauf hängen bleibt, ohne dass die Ausgabe der Messung davon
// voll wird.
function spur(was) {
  if (process.env.MSE_SPUR) {
    process.stderr.write(`  [${new Date().toISOString().slice(11, 19)}] ${was}\n`);
  }
}

// frisch lädt die Anwendung neu — jeder Weg beginnt auf der Startseite, mit
// leerem Verlauf, wie ein Mensch, der das Programm gerade geöffnet hat.
async function frisch() {
  await schicke(ws, "Page.navigate", { url: adresse });
  spur("navigiert");
  await griffe.bis(ws, `typeof window.go?.app?.Bindungen === 'object'`, "die Brücke");
  spur("Brücke da");
  await griffe.bis(ws, `document.querySelector('main') !== null`, "die Startseite");
  spur("main da");
  await griffe.bisRuhe(ws);
  spur("Ruhe");
}

try {
  ws = await verbinden(profil);
  spur("verbunden");

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: 1280,
    height: 800,
    deviceScaleFactor: 1,
    mobile: false,
  });

  await frisch();
  spur("Startseite steht");
  // **Mit eingeklappter Leiste**, wenn gewünscht (#1490): die Leiste ist eine
  // Einstellung auf Geräteebene und steht danach bei jedem `frisch()` zu.
  if (eingeklapptGewuenscht) {
    zeilen.push(`  ${await seitenleisteEinklappen(ws)}`);
    await frisch();
  }

  const exemplare = await griffe.zaehleExemplare(ws);
  spur("Bestand gelesen");
  process.stdout.write(
    `Arbeitswege bei 1280 × 800 gegen ${adresse} — ${String(exemplare)} Exemplare im Bestand\n\n`,
  );

  for (const weg of arbeitswege) {
    if (nurWeg !== "" && weg.schluessel !== nurWeg) {
      continue;
    }

    process.stdout.write(`  … ${weg.name}\n`);

    let stand = null;
    let ergebnis = null;
    let fehler = null;

    try {
      stand = await weg.vorbereiten(griffe, ws);
      await frisch();
      zuruecksetzen();
      ergebnis = await weg.fahren(griffe, ws, stand);
    } catch (grund) {
      fehler = grund instanceof Error ? grund.message : String(grund);
    } finally {
      if (stand !== null) {
        try {
          await weg.aufraeumen(griffe, ws, stand);
        } catch (grund) {
          fehler = `${fehler ?? "" } (aufräumen: ${grund instanceof Error ? grund.message : String(grund)})`;
        }
      }
    }

    if (fehler !== null) {
      fehlgeschlagen = true;
      zeilen.push(`  ✗ ${weg.name}: ${fehler}`);

      continue;
    }

    const handlungen = gezaehlt();
    const befund = handlungen === weg.erwartet.handlungen ? "" : ` — erwartet ${String(weg.erwartet.handlungen)}`;

    if (befund !== "") {
      fehlgeschlagen = true;
    }

    zeilen.push(
      `  ${befund === "" ? "✓" : "✗"} ${weg.name.padEnd(18)} ${String(handlungen).padStart(2)} Handlungen  ` +
        `${String(ergebnis.tabstellen).padStart(3)}. Tabstelle  ${String(ergebnis.ladezeit).padStart(5)} ms` +
        `${ergebnis.zusatz ? `  ${ergebnis.zusatz}` : ""}${befund}`,
    );
  }
} catch (grund) {
  fehlgeschlagen = true;
  zeilen.push(`  ✗ ${grund instanceof Error ? grund.message : String(grund)}`);
} finally {
  if (ws !== null && eingeklapptGewuenscht) {
    try {
      await frisch();
      await seitenleisteAusklappen(ws);
    } catch (grund) {
      zeilen.push(`  ✗ die Seitenleiste blieb zu: ${grund instanceof Error ? grund.message : String(grund)}`);
    }
  }
  process.stdout.write(`${zeilen.join("\n")}\n`);

  if (ws !== null) {
    ws.close();
  }

  chrome.kill();
  // **Chrome schreibt nach dem Beenden noch** — ohne die Frist blieb
  // `ENOTEMPTY` übrig, wie schon in `kernwegmass.mjs`.
  await warte(500);
  await rm(profil, { recursive: true, force: true, maxRetries: 5 });
}

process.exit(fehlgeschlagen ? 1 : 0);
