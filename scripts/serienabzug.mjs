// **Jede Seite abgelichtet** — Werkstück A von Phase 4.10
// (#1325).
//
// # Wozu, und was ihn von den README-Bildern unterscheidet
//
// Die Hüllenprobe `readmebilder` macht die **Bilder des README**: ausgewählte
// Wege am nativen Fenster, in einem Zustand, der etwas erklärt. Dieser Lauf macht etwas
// anderes — er legt **alle zweiundzwanzig** Seiten nebeneinander, hell und
// dunkel, bei 1280 × 800, damit ein Mensch sie **vergleichen** kann.
//
// Die Abnahme von Phase 4.10 verlangt genau das: *„Jede Seite ist abgelichtet,
// vermessen und schriftlich beurteilt."* Vermessen wird in `serienmass.mjs`,
// und beide fahren dieselbe Liste (`serienmasslisten.mjs`) — eine Seite, die
// gemessen und nicht abgelichtet wird, gäbe es sonst.
//
// ```bash
// node scripts/serienabzug.mjs [zielverzeichnis]
// ```
//
// # Was er **nicht** tut
//
//   - **Er urteilt nicht.** Das ist Werkstück B, und es geht nur von Hand.
//   - **Er scrollt nicht.** Was über der Falz steht, ist die Frage; ein Abzug
//     der ganzen Seitenhöhe beantwortete sie nicht.

import { spawn } from "node:child_process";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";
import { anwendungsseiten, setlaschen } from "./serienmasslisten.mjs";

const adresse = "http://localhost:34115";
const ziel = process.argv[2] ?? join(process.cwd(), "abzuege");

const breite = 1280;
const hoehe = 800;

// ruhe wartet, bis die Seite nichts mehr an sich ändert.
async function ruhe(ws) {
  const zaehlen = async () =>
    Number(await werte(ws, `document.querySelectorAll('*').length`));

  let vorher = await zaehlen();

  for (let versuch = 0; versuch < 20; versuch += 1) {
    await warte(500);

    const jetzt = await zaehlen();
    if (jetzt === vorher) {
      return;
    }

    vorher = jetzt;
  }
}

// klicke ruft einen Knopf über einen Ausdruck und sagt, ob er dastand.
async function klicke(ws, ausdruck) {
  return (
    (await werte(
      ws,
      `(() => { const e = ${ausdruck}; if (!e) return false; e.click(); return true; })()`,
    )) === true
  );
}

// ablichten schreibt einen Abzug.
async function ablichten(ws, name, anhang) {
  const { data } = await schicke(ws, "Page.captureScreenshot", { format: "png" });
  const datei = join(ziel, `${name}${anhang}.png`);

  await writeFile(datei, Buffer.from(data, "base64"));
  process.stdout.write(`  ✓ ${datei}\n`);
}

const profil = await mkdtemp(join(tmpdir(), "serienabzug-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

try {
  await mkdir(ziel, { recursive: true });

  const ws = await verbinden(profil);

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: breite,
    height: hoehe,
    deviceScaleFactor: 1,
    mobile: false,
  });

  await schicke(ws, "Page.navigate", { url: adresse });
  await warte(4000);

  for (const erscheinung of ["light", "dark"]) {
    const anhang = erscheinung === "dark" ? "-dunkel" : "";

    await schicke(ws, "Emulation.setEmulatedMedia", {
      features: [{ name: "prefers-color-scheme", value: erscheinung }],
    });

    for (const seite of anwendungsseiten) {
      for (const beschriftung of seite.weg) {
        if (
          !(await klicke(
            ws,
            `[...document.querySelectorAll('button')]
               .find((b) => b.textContent.trim().startsWith(${JSON.stringify(beschriftung)}))`,
          ))
        ) {
          throw new Error(`kein Knopf, der mit ${beschriftung} beginnt`);
        }

        await warte(1200);
      }

      await ruhe(ws);
      await ablichten(ws, `seite-${seite.schluessel}`, anhang);
    }

    if (
      !(await klicke(
        ws,
        `[...document.querySelectorAll('button')]
           .find((b) => b.textContent.trim().startsWith("Sets"))`,
      ))
    ) {
      throw new Error("die Seitenleiste führt keine Sets");
    }

    await ruhe(ws);

    if (
      !(await klicke(
        ws,
        `[...document.querySelectorAll('button[aria-label]')]
           .find((b) => /Karten dieses Sets zeigen$/.test(b.getAttribute('aria-label')))`,
      ))
    ) {
      throw new Error("kein Set zum Öffnen");
    }

    await warte(1500);

    for (const lasche of setlaschen) {
      // **Sieben Laschen stehen in einer Schublade** (UX-08, #1680) — erst
      // ihr Knopf im Kopf des Sets, dann der Reiter darin.
      if (
        lasche.schublade !== undefined &&
        !(await klicke(
          ws,
          `[...document.querySelectorAll('button')]
             .find((b) => b.textContent.trim() === ${JSON.stringify(lasche.schublade)})`,
        ))
      ) {
        throw new Error(`kein Knopf ${lasche.schublade}`);
      }

      // **Die Schublade steht erst nach einem Effekt** — Mantine hängt ihr
      // Portal in `useEffect` ein; der Reiter darin ist vorher nicht da.
      if (lasche.schublade !== undefined) {
        await warte(600);
      }

      if (
        !(await klicke(
          ws,
          `[...document.querySelectorAll('[role="tab"], label')]
             .find((b) => b.textContent.trim() === ${JSON.stringify(lasche.reiter)})`,
        ))
      ) {
        throw new Error(`keine Lasche ${lasche.reiter}`);
      }

      await warte(1200);
      await ruhe(ws);
      await ablichten(ws, `lasche-${lasche.schluessel}`, anhang);
    }
  }

  ws.close();
} finally {
  chrome.kill();
  await warte(400);
  await rm(profil, { recursive: true, force: true }).catch(() => undefined);
}
