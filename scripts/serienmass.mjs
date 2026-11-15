// **Was eine Seite kostet — gemessen, nicht geschätzt**
// (#1290).
//
// # Wozu
//
// Die Abnahme von Phase 4.3 verlangt eine Zahl: *„höchstens X Nutzerhandlungen
// … X wird in der Baseline festgelegt und danach unterschritten oder
// begründet."* Eine solche Zahl von Hand zu erheben ist einmal möglich und
// beim zweiten Mal nicht mehr vergleichbar.
//
// Dieser Lauf erhebt sie **an der laufenden Brücke**: `make dev-browser`, ein Chrome
// ohne Fenster bei 1280 × 800, echte Tastendrücke über
// `Input.dispatchKeyEvent`. Was er zählt, steht in `docs/SERIENARBEIT.md`.
//
// ```bash
// node scripts/serienmass.mjs
// ```
//
// # Warum 1280 × 800
//
// Weil es das kleinste Maß ist, das ein aktuelles MacBook Air im
// Standardmaßstab zeigt. Die Abzüge der Dokumentation fahren 1440 × 841
// (Hüllenprobe `readmebilder`) — dort geht es darum, wie etwas **aussieht**,
// hier darum, ob es **hineinpasst**.
//
// # Der Ausdruck für „tabulatorerreichbar" liegt in `tabstellen.mjs`
//
// **Einmal, nicht zweimal**: `arbeitswegmass.mjs` (#1412) zählt mit demselben
// Ausdruck, und zwei Begriffe von „erreichbar" nebeneinander wären zwei
// Messgeräte, deren Zahlen niemand vergleichen dürfte (§10). Warum
// `button:not([disabled])` nicht genügt, steht dort.
//
// **Die Reichweite einer Messung ist die Reichweite ihrer Aussage** (§10). Wer
// zwei Zahlen vergleicht, prüft zuerst, ob dasselbe Gerät sie erhoben hat.
//
// # Was er **nicht** misst
//
//   - **Nicht, ob eine Seite gut aussieht.** Dafür gibt es die Sichtprüfung
//     (§3) und Phase 4.10 (#1325).
//   - **Nicht die Zeit.** Eine Dauer wäre eine Aussage über diesen Rechner.
//   - **Nicht die Vollständigkeit einer Seite.** Er zählt, was dasteht.

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";
import { anwendungsseiten, setlaschen } from "./serienmasslisten.mjs";
import { tabulatorerreichbar } from "./tabstellen.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

// breite und hoehe sind das Maß, an dem gemessen wird.
const breite = 1280;
const hoehe = 800;

// messen liest die Zahlen einer Seite.
async function messen(ws) {
  return JSON.parse(
    await werte(
      ws,
      `JSON.stringify((() => {
        const erreichbar = [...document.querySelectorAll(${JSON.stringify(tabulatorerreichbar)})];
        const felder = [...document.querySelectorAll('input[type="text"], input:not([type]), textarea')];

        // Was das Auge auf einmal sieht — Werkstueck H von Phase 4.10,
        // Issue 1325. Die Zahl daneben zaehlt den BAUM, und das ist fuer die
        // Frage "wirkt die Seite erschlagend?" die falsche Groesse: die
        // Paketliste hatte 411 Knoepfe, davon 331-mal "Entfernen" — und ueber
        // der Falz standen nie mehr als ein Dutzend.
        //
        // Sichtbar heisst hier: im ersten Bildschirm, nicht display:none,
        // nicht durchsichtig, und mit einer Flaeche. Was hinter einem
        // zugeklappten Kasten liegt, zaehlt nicht — genau das ist ja die
        // Abhilfe.
        const sichtbar = (knoten) => {
          const stil = getComputedStyle(knoten);

          if (stil.visibility === 'hidden' || stil.display === 'none') return false;
          if (Number(stil.opacity) < 0.05) return false;

          const kasten = knoten.getBoundingClientRect();

          return kasten.width > 0 && kasten.height > 0 &&
            kasten.top < ${String(hoehe)} && kasten.bottom > 0;
        };

        const bedienelemente = [...document.querySelectorAll(
          'button, a[href], input, textarea, select, [role="tab"], [tabindex]:not([tabindex="-1"])')];

        return {
          tabstellen: erreichbar.length,
          knoepfe: document.querySelectorAll('button').length,
          sichtbar: bedienelemente.filter(sichtbar).length,
          eingaben: document.querySelectorAll('input, textarea, select').length,
          hoehe: document.documentElement.scrollHeight,
          bildschirme: Math.round(document.documentElement.scrollHeight / ${String(hoehe)} * 10) / 10,
          textfelder: felder.length,
          ueberDerFalz: felder.filter((f) => f.getBoundingClientRect().top < ${String(hoehe)}).length,
          ueberschrift: document.querySelector('main h1, main h2')?.textContent?.trim().slice(0, 40) ?? '',
        };
      })())`,
    ),
  );
}

// hinKlicken klickt einen Weg über die Beschriftungen der Seitenleiste.
//
// **Sie bricht ab, wenn ein Knopf fehlt** — Befund 3 des Zweitblicks über
// Phase 4.3. Der erste Anlauf warf das `false` weg und wartete stur weiter;
// eine fehlgeschlagene Navigation gab damit plausible Zahlen **für die falsche
// Seite**. Ein Messgerät, das bei einem Fehler weitermisst, ist schlimmer als
// keines.
async function hinKlicken(ws, weg) {
  for (const beschriftung of weg) {
    const getroffen = await werte(
      ws,
      `(() => {
        const e = [...document.querySelectorAll('button')]
          .find((b) => b.textContent.trim().startsWith(${JSON.stringify(beschriftung)}));
        if (!e) return false;
        e.click();
        return true;
      })()`,
    );

    if (getroffen !== true) {
      // **Die Meldung sagt, was dastand.** Ohne die Aufzählung schickt sie
      // jemanden suchen: „kein Knopf, der mit Turniere beginnt" lässt offen,
      // ob die Seite fehlt, anders heißt oder gar nicht in der Leiste steht.
      const dawaren = await werte(
        ws,
        `JSON.stringify([...document.querySelectorAll('button')]
           .map((b) => b.textContent.trim()).filter(Boolean).slice(0, 40))`,
      );

      throw new Error(
        `kein Knopf, der mit ${beschriftung} beginnt — dastanden: ${dawaren}`,
      );
    }

    await warte(1500);
  }
}

// zeileSchreiben setzt eine Zeile der Tabelle.
function zeileSchreiben(name, zahlen) {
  process.stdout.write(
    `| ${name} | ${String(zahlen.tabstellen)} | ${String(zahlen.knoepfe)} ` +
    `| ${String(zahlen.sichtbar)} ` +
      `| ${String(zahlen.hoehe)} px | ${String(zahlen.bildschirme)} ` +
      `| ${String(zahlen.textfelder)} | ${String(zahlen.ueberDerFalz)} |\n`,
  );
}

// setOeffnen schlägt das erste Set der Liste auf.
//
// **Über das erste und nicht über ein benanntes.** Die README-Bilder nehmen
// ein festes Set, weil ein Abzug immer dasselbe zeigen soll; hier geht es um die
// **Kosten** einer Lasche, und die hängen am Set nur über seine Größe. Wer
// ein bestimmtes Set verlangt, macht den Lauf von einem Bestand abhängig.
async function setOeffnen(ws) {
  const getroffen = await werte(
    ws,
    `(() => {
      const e = [...document.querySelectorAll('button[aria-label]')]
        .find((b) => /Karten dieses Sets zeigen$/.test(b.getAttribute('aria-label')));
      if (!e) return false;
      e.click();
      return true;
    })()`,
  );

  if (getroffen !== true) {
    throw new Error(
      "kein Set zum Öffnen — der Bestand ist leer, und ohne offenes Set " +
        "lassen sich die Laschen nicht messen",
    );
  }

  await warte(1500);
}

// schubladeOeffnen drückt den Knopf im Kopf des Sets, der eine Schublade
// öffnet (UX-08, #1680). **Auch wenn schon eine offen ist**: der Knopf ist
// ein Seitenwechsel, und die Schublade der neuen Seite ersetzt die alte.
async function schubladeOeffnen(ws, knopf) {
  const gedrueckt = await werte(
    ws,
    `(() => {
      const e = [...document.querySelectorAll('button')]
        .find((b) => b.textContent.trim() === ${JSON.stringify(knopf)});
      if (!e) return "fehlt";
      e.click();
      return "geklickt";
    })()`,
  );

  if (gedrueckt !== "geklickt") {
    throw new Error(`kein Knopf ${knopf} im Kopf des Sets`);
  }

  await warte(600);
}

// laschenWechsel klickt einen Reiter des offenen Sets.
//
// **Geprüft wird `aria-selected`, nicht eine Überschrift.** Eine Lasche trägt
// ihren Namen im Reiter; eine `h1` mit demselben Wort hat sie nicht. Wer hier
// eine Überschrift verlangte, prüfte etwas, das es nicht gibt — und ein
// Messgerät, das bei einem Fehler weitermisst, ist schlimmer als keines.
async function laschenWechsel(ws, reiter) {
  const gewaehlt = await werte(
    ws,
    `(() => {
      const e = [...document.querySelectorAll('[role="tab"], label')]
        .find((b) => b.textContent.trim() === ${JSON.stringify(reiter)});
      if (!e) return "fehlt";
      e.click();
      return "geklickt";
    })()`,
  );

  if (gewaehlt !== "geklickt") {
    const dawaren = await werte(
      ws,
      `JSON.stringify([...document.querySelectorAll('[role="tab"]')]
         .map((b) => b.textContent.trim()).filter(Boolean))`,
    );

    throw new Error(`keine Lasche ${reiter} — dastanden: ${dawaren}`);
  }

  await warte(1200);

  const steht = await werte(
    ws,
    // **Oder das angehakte Ziel des Ausgabeassistenten** (UXR-09, #1917).
    `[...document.querySelectorAll('[role="tab"][aria-selected="true"]')]
       .some((b) => b.textContent.trim() === ${JSON.stringify(reiter)})
     || [...document.querySelectorAll('label')]
       .some((b) => b.textContent.trim() === ${JSON.stringify(reiter)}
         && document.getElementById(b.htmlFor)?.checked === true)`,
  );

  if (steht !== true) {
    throw new Error(`die Lasche ${reiter} ist nach dem Klick nicht gewählt`);
  }
}

// stehtDieAnwendung prüft, ob unter der Adresse überhaupt etwas antwortet.
//
// **Weil dieser Lauf keinen Server startet** (Befund 3): er braucht ein
// laufendes `make dev-browser`. Ohne diese Prüfung misst er eine Fehlerseite und gibt
// Zahlen aus, die wie welche aussehen.
async function stehtDieAnwendung(ws) {
  const bruecke = await werte(
    ws,
    `typeof window.go === 'object' && typeof window.go.app === 'object'`,
  );

  if (bruecke !== true) {
    throw new Error(
      `unter ${adresse} antwortet keine Anwendung — l\u00e4uft \`make dev-browser\`?`,
    );
  }
}

// bisRuhe wartet, bis die Seite fertig geladen ist.
//
// **Weil eine halbe Seite plausible Zahlen gibt** (Befund 3 des Zweitblicks
// über Phase 4.3). Die Paketseite zählte je nach Lauf 51 oder 1.833
// Textstellen — dieselbe Seite, einmal vor und einmal nach dem Laden ihrer
// Liste. Beide Läufe waren grün, und einer davon sagte nichts.
//
// **Zweimal dieselbe Zahl heißt fertig.** Ein fester Wartewert müsste raten,
// wie lange — und ein Test darüber prüfte den Rechner.
async function bisRuhe(ws) {
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

  throw new Error("die Seite kam in zehn Sekunden nicht zur Ruhe");
}

const profil = await mkdtemp(join(tmpdir(), "serienmass-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

try {
  const ws = await verbinden(profil);

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: breite,
    height: hoehe,
    deviceScaleFactor: 1,
    mobile: false,
  });

  process.stdout.write(
    `| Seite | Tabstellen | Knöpfe | sichtbar | Höhe | Bildschirme | Textfelder | über der Falz |\n` +
      `| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n`,
  );

  await schicke(ws, "Page.navigate", { url: adresse });
  await warte(3500);
  await stehtDieAnwendung(ws);

  // **Einmal laden, dann durchklicken.** Ein Neuaufbau je Seite misst nicht,
  // was jemand tut: wer die Anwendung benutzt, klickt in der Leiste. Genau das
  // tut dieser Lauf.
  for (const seite of anwendungsseiten) {
    await hinKlicken(ws, seite.weg);
    await bisRuhe(ws);

    const zahlen = await messen(ws);

    // **Die Überschrift wird verglichen und nicht nur erhoben** (Befund 3).
    // Der erste Anlauf las sie und sah sie nie an — eine Seite, die nicht
    // aufging, stand mit den Zahlen der vorigen in der Tabelle.
    if (seite.ueberschrift !== undefined && zahlen.ueberschrift !== seite.ueberschrift) {
      throw new Error(
        `${seite.name}: die Überschrift lautet ${JSON.stringify(zahlen.ueberschrift)}, ` +
          `erwartet war ${JSON.stringify(seite.ueberschrift)}`,
      );
    }

    zeileSchreiben(seite.name, zahlen);
  }

  // **Und die dreizehn Laschen eines offenen Sets** — seit dem 2026-09-03.
  //
  // Bis dahin maß dieser Lauf neun von zweiundzwanzig Seiten, und die
  // Tabelle in `docs/SERIENARBEIT.md` sah trotzdem vollständig aus. Die
  // Laschen sind der größere Teil der Anwendung: dort steht die Karte, und
  // dort verbringt jemand seine Zeit.
  await hinKlicken(ws, ["Sets"]);
  await bisRuhe(ws);
  await setOeffnen(ws);
  await bisRuhe(ws);

  process.stdout.write(`\n`);
  process.stdout.write(
    `| Lasche | Tabstellen | Knöpfe | sichtbar | Höhe | Bildschirme | Textfelder | über der Falz |\n` +
      `| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n`,
  );

  for (const lasche of setlaschen) {
    if (lasche.schublade !== undefined) {
      await schubladeOeffnen(ws, lasche.schublade);
    }

    await laschenWechsel(ws, lasche.reiter);
    await bisRuhe(ws);

    zeileSchreiben(lasche.name, await messen(ws));
  }
} finally {
  chrome.kill();
  await warte(200);
  await rm(profil, { recursive: true, force: true });
}
