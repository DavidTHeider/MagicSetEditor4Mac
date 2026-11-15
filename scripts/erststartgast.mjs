#!/usr/bin/env node
// **Der Erststart aus dem Abbild, im Wegwerf-Gast** — HAND6
// (#2384).
//
//     MSE_FREIGABE_NUR_BAU=1 ./scripts/release.sh <fassung>
//     node scripts/erststartgast.mjs build/freigabe/MagicSetEditor4Mac-<fassung>.dmg [<vm>]
//
// **Nie am Wirt**: der Lauf legt eine Quarantäne an und lässt Gatekeeper
// abweisen — am Wirt stünde die Abweisung auf dem Bildschirm dessen, der
// arbeitet. Der Treiber ist `scripts/huellenprobe/gast/erststart.pl`; was er
// misst und was nicht, steht in seinem Kopf. Der Beleg geht nach
// `docs/belege/huellenprobe/erststart-<system>.md`.
//
// **Was dieser Lauf nicht ist**: der Bezug über den Freigabekanal ohne Zugang
// zum Repositorium (das Abbild kommt über das geteilte Verzeichnis) und die
// Freigabe über „Dennoch öffnen“ — beides gehört zum Freigabegate (#2385).
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { basename, join } from "node:path";

import {
  fahre,
  gastziel,
  imGeteilten,
  schrittzeilen,
  wurzel,
} from "./huellenprobe/gastlauf.mjs";

const [abbild, ziel] = process.argv.slice(2);
if (!abbild || !existsSync(abbild)) {
  console.error(
    "Aufruf: node scripts/erststartgast.mjs <MagicSetEditor4Mac-<fassung>.dmg> [<vm>]",
  );
  process.exit(2);
}
const { vm } = gastziel(ziel, { wirtErlaubt: false });
const pruefsumme = execFileSync("shasum", ["-a", "256", abbild], {
  encoding: "utf8",
}).split(/\s+/)[0];
const commit = execFileSync(
  "git",
  ["-C", wurzel, "rev-parse", "--short=9", "HEAD"],
  { encoding: "utf8" },
).trim();

const beginn = new Date();
const { lauf, ergebnis } = await imGeteilten("hand6", async (geteilt) => {
  const ordner = join(geteilt, "hand6");
  mkdirSync(ordner, { recursive: true });
  execFileSync("cp", ["-c", abbild, join(ordner, "abbild.dmg")]);
  writeFileSync(
    join(ordner, "pruefsumme.txt"),
    `${pruefsumme}  ${basename(abbild)}\n`,
  );
  for (const datei of ["erststart.pl", "Gast.pm"])
    execFileSync("cp", [
      "-c",
      join(wurzel, "scripts/huellenprobe/gast", datei),
      join(ordner, datei),
    ]);
  const lauf = await fahre({
    vm,
    geteilt,
    name: "hand6",
    treiber: "erststart.pl",
    amWirt: false,
  });
  const datei = join(geteilt, "hand6-ergebnis.json");

  return {
    lauf,
    ergebnis: existsSync(datei)
      ? JSON.parse(readFileSync(datei, "utf8"))
      : null,
  };
});
if (!ergebnis) {
  console.error(
    `Kein Ergebnis (Rückgabe ${String(lauf.rueckgabe)}):\n${lauf.ausgabe}`,
  );
  process.exit(1);
}

const rot = ergebnis.schritte.filter((s) => !s.ok);
const g = ergebnis.maschine;
const text = [
  `# Beleg: HAND6 — der Erststart aus dem Abbild auf macOS ${g.system}`,
  "",
  // **Ein eigener Kopf**: `kopfzeilen` nennt den Abnahmebau, und hier läuft
  // das Freigabeabbild.
  `**Stand:** Commit \`${commit}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace("T", " ")} UTC · **Befehl:** \`node scripts/erststartgast.mjs ${basename(abbild)} ${vm}\` · **Rückgabe:** ${String(lauf.rueckgabe)}`,
  "",
  `**Maschine:** in der Lume-VM \`${vm}\` — macOS ${g.system} (${g.bau}), ${g.modell}, ${g.kerne} Kerne.`,
  "",
  `- **Abbild**: \`${basename(abbild)}\`, SHA-256 \`${pruefsumme}\`, gebaut am Stand \`${commit}\` mit \`MSE_FREIGABE_NUR_BAU=1 ./scripts/release.sh\` — **kein Freigabebau**: ohne Pipeline, Prüfsummenliste und Signatur.`,
  `- **Bündel**: ${Object.entries(ergebnis.info)
    .map(([k, v]) => `\`${k}\` = \`${v}\``)
    .join(", ")}.`,
  `- **Quarantäne**: kenntlich synthetisch, Absender \`mse4mac-hand6\`; am herauskopierten Bündel \`${ergebnis.erbt ?? "?"}\`. **Kein \`xattr -d\`** in diesem Lauf.`,
  "",
  ...schrittzeilen(ergebnis.schritte, "Teil"),
  "",
  "## Was entstand",
  "",
  `Schema: ${ergebnis.schema || "—"} Wanderungen.`,
  "",
  "| Pfad | Art | Rechte | Größe |",
  "| --- | --- | --- | --- |",
  ...ergebnis.dateien.map(
    (d) =>
      `| \`${d.pfad}\` | ${d.art} | \`${d.rechte}\` | ${d.art === "Datei" ? String(d.groesse) : ""} |`,
  ),
  "",
  "## Was syspolicyd sagte",
  "",
  "```",
  (ergebnis.syspolicyd ?? "").trim() || "(nichts in den letzten drei Minuten)",
  "```",
  "",
  "## Grenzen",
  "",
  ...ergebnis.grenzen.map((x) => `- ${x}`),
  "- **Der Bezug über den Freigabekanal ohne Zugang zum Repositorium** ist hier nicht gelaufen: das Abbild kam über das geteilte Verzeichnis. Er braucht ein öffentliches Repositorium (Vermerk 0190) und gehört zum Freigabegate (#2385).",
  "",
].join("\n");

const beleg = join(wurzel, `docs/belege/huellenprobe/erststart-${g.system}.md`);
writeFileSync(beleg, text);
console.log(`${rot.length ? "✗" : "✓"} erststart → ${beleg}`);
process.exit(rot.length ? 1 : 0);
