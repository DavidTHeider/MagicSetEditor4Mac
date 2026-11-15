#!/usr/bin/env bash
#
# Prüfungen der Einordnung aus scripts/sicherheitslage.sh — ENG-04.AK-03
# (#1729).
#
# Gegen **nachgestellte** Ausgaben der beiden Scanner: einen echten Ausfall der
# Schwachstellendatenbank herbeizuführen hieße, das Netz abzuklemmen, und eine
# echte Schwachstelle einzubauen, eine verwundbare Fassung zu pinnen. Die
# Gestalt der Ausgaben ist nachgelesen — govulncheck v1.6.0
# (`internal/scan/errors.go`, `text.go`: 0 ohne Befund, 3 mit, 1 Fehler) und
# npm 11.12.1 (`lib/utils/audit-error.js`).

set -uo pipefail

WURZEL="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=scripts/sicherheitslage.sh
source "$WURZEL/scripts/sicherheitslage.sh"

# shellcheck source=scripts/releasebeleg.sh
source "$WURZEL/scripts/releasebeleg.sh"

GRUEN=$'\033[32m'; ROT=$'\033[31m'; AUS=$'\033[0m'
[ -t 1 ] || { GRUEN=''; ROT=''; AUS=''; }

FEHLER=0
RAUM="$(mktemp -d)"
trap 'rm -rf "$RAUM"' EXIT

gleich() { # gleich <beschreibung> <erwartet> <bekommen>
  if [ "$2" = "$3" ]; then
    printf '%s  ✓ %s%s\n' "$GRUEN" "$1" "$AUS"
  else
    printf '%s  ✗ %s — erwartet „%s“, bekommen „%s“%s\n' "$ROT" "$1" "$2" "$3" "$AUS"
    FEHLER=1
  fi
}

# ------------------------------------------------------------- govulncheck

printf 'No vulnerabilities found.\n' > "$RAUM/govuln-leer"
gleich "govulncheck 0 ist ohne Befund" "ohne-befund" \
  "$(govuln_zustand 0 "$RAUM/govuln-leer")"

printf 'Vulnerability #1: GO-2026-0001\n\nYour code is affected by 1 vulnerability from 1 module.\n' \
  > "$RAUM/govuln-befund"
gleich "govulncheck 3 ist ein Befund" "befund" \
  "$(govuln_zustand 3 "$RAUM/govuln-befund")"

printf 'govulncheck: fetching vulnerabilities: Get "https://vuln.go.dev/index/modules.json.gz": dial tcp: lookup vuln.go.dev: no such host\n' \
  > "$RAUM/govuln-netz"
gleich "govulncheck ohne Datenbank ist nicht erreichbar" "nicht-erreichbar" \
  "$(govuln_zustand 1 "$RAUM/govuln-netz")"

printf 'govulncheck: loading packages: there are errors in your code\n' > "$RAUM/govuln-kaputt"
gleich "ein anderer Fehler ist ein gescheiterter Lauf, kein Ausfall des Netzes" "gescheitert" \
  "$(govuln_zustand 1 "$RAUM/govuln-kaputt")"

# ---------------------------------------------------------------- npm audit

cat > "$RAUM/npm-leer" <<'JSON'
{"auditReportVersion":2,"vulnerabilities":{},"metadata":{"vulnerabilities":{"info":0,"low":0,"moderate":0,"high":0,"critical":0,"total":0}}}
JSON
gleich "npm audit ohne Schwachstelle ist ohne Befund" "ohne-befund" \
  "$(npmaudit_zustand 0 "$RAUM/npm-leer")"

cat > "$RAUM/npm-maessig" <<'JSON'
{"auditReportVersion":2,"vulnerabilities":{"x":{}},"metadata":{"vulnerabilities":{"info":0,"low":1,"moderate":2,"high":0,"critical":0,"total":3}}}
JSON
gleich "mittlere Funde unter der Schwelle sind ohne Befund — die Schwelle ist high" "ohne-befund" \
  "$(npmaudit_zustand 0 "$RAUM/npm-maessig")"

cat > "$RAUM/npm-hoch" <<'JSON'
{"auditReportVersion":2,"vulnerabilities":{"x":{}},"metadata":{"vulnerabilities":{"info":0,"low":0,"moderate":0,"high":1,"critical":0,"total":1}}}
JSON
gleich "ein hoher Fund ist ein Befund" "befund" \
  "$(npmaudit_zustand 1 "$RAUM/npm-hoch")"

cat > "$RAUM/npm-503" <<'JSON'
{"message":"503 Service Unavailable - POST https://registry.npmjs.org/-/npm/v1/security/advisories/bulk","method":"POST","uri":"https://registry.npmjs.org/-/npm/v1/security/advisories/bulk","statusCode":503,"body":""}
JSON
gleich "ein Audit-Endpunkt mit 503 ist nicht erreichbar" "nicht-erreichbar" \
  "$(npmaudit_zustand 1 "$RAUM/npm-503")"

printf 'npm error code ENOTFOUND\nnpm error request to https://registry.npmjs.org failed, reason: getaddrinfo ENOTFOUND registry.npmjs.org\n' \
  > "$RAUM/npm-netz"
gleich "kein Netz ist nicht erreichbar, auch ohne JSON" "nicht-erreichbar" \
  "$(npmaudit_zustand 1 "$RAUM/npm-netz")"

printf 'npm error code ELOCKVERIFY\nnpm error the lockfile is broken\n' > "$RAUM/npm-kaputt"
gleich "ein anderer Fehler ist ein gescheiterter Lauf" "gescheitert" \
  "$(npmaudit_zustand 1 "$RAUM/npm-kaputt")"

# **Ein Bericht, dem sein Rückgabewert widerspricht** — Befund 4 des
# Zweitblicks über Phase 6.2. npm 11.12.1 endet mit 1 genau dann, wenn eine
# Zahl ab der Schwelle im Bericht steht (`npm-audit-report` 7.0.0,
# `lib/exit-code.js`, über `process.exitCode || result.exitCode` in
# `lib/commands/audit.js`). Eine 1 zu einem Bericht ohne solchen Fund heißt,
# dass vorher etwas anderes scheiterte — und das ist kein grüner Lauf.
gleich "ein Bericht ohne hohen Fund bei Rückgabe 1 ist gescheitert" "gescheitert" \
  "$(npmaudit_zustand 1 "$RAUM/npm-maessig")"

# **Umgekehrt bleibt ein genannter Fund ein Befund**, auch wenn der
# Rückgabewert schweigt: eine Schwachstelle, die im Bericht steht, wird nicht
# dadurch kleiner, dass niemand rot wurde.
gleich "ein hoher Fund bleibt ein Befund, auch bei Rückgabe 0" "befund" \
  "$(npmaudit_zustand 0 "$RAUM/npm-hoch")"

# **Und ein Bericht mit Fehlerangabe ist keiner.** npm 11.12.1 schreibt beides
# nicht zusammen (`lib/utils/audit-error.js` wirft vor dem Bericht); steht es
# doch da, entscheidet die Meldung und nicht die Zahl.
cat > "$RAUM/npm-mitfehler" <<'JSON'
{"auditReportVersion":2,"vulnerabilities":{},"metadata":{"vulnerabilities":{"info":0,"low":0,"moderate":0,"high":0,"critical":0,"total":0}},"error":{"code":"ELOCKVERIFY","summary":"the lockfile is broken"}}
JSON
gleich "ein Bericht mit Fehlerangabe ist gescheitert" "gescheitert" \
  "$(npmaudit_zustand 0 "$RAUM/npm-mitfehler")"
gleich "ein Bericht mit Fehlerangabe gibt keine Zahlen" "" \
  "$(npmaudit_zahlen "$RAUM/npm-mitfehler")"

# ------------------------------------------------------- Was der Beleg nennt

# **Der Beleg nennt die Befunde, nicht nur den Zustand** (AK-03): bei
# govulncheck die Kennungen, bei npm audit die Zahlen je Schwere — auch unter
# der Schwelle, denn ein mittlerer Fund ist kein Befund, aber eine Auskunft.
printf 'Vulnerability #1: GO-2026-0002\nVulnerability #2: GO-2026-0001\nMore info: https://pkg.go.dev/vuln/GO-2026-0002\n' \
  > "$RAUM/govuln-zwei"
gleich "govulncheck nennt jede Kennung einmal, geordnet" "GO-2026-0001, GO-2026-0002" \
  "$(govuln_kennungen "$RAUM/govuln-zwei")"

gleich "npm audit nennt die Zahlen je Schwere" "kritisch 0 · hoch 0 · mittel 2 · niedrig 1" \
  "$(npmaudit_zahlen "$RAUM/npm-maessig")"
gleich "ohne Bericht gibt es keine Zahlen" "" \
  "$(npmaudit_zahlen "$RAUM/npm-503")"

# **Und die Lage fährt beide Scanner** — mit nachgestellten, damit der Lauf
# ohne Netz und ohne Schwachstelle prüfbar ist.
mkdir -p "$RAUM/bin" "$RAUM/lauf/frontend"
cat > "$RAUM/bin/govulncheck" <<SKRIPT
#!/usr/bin/env bash
cat "$RAUM/govuln-zwei"
exit 3
SKRIPT
cat > "$RAUM/bin/npm" <<SKRIPT
#!/usr/bin/env bash
cat "$RAUM/npm-maessig"
SKRIPT
chmod +x "$RAUM/bin/govulncheck" "$RAUM/bin/npm"

lage="$(cd "$RAUM/lauf" && PATH="$RAUM/bin:$PATH" sicherheitslage ./...)"
gleich "die Lage nennt den Befund mit seinen Kennungen" \
  "govulncheck: befund — GO-2026-0001, GO-2026-0002" "$(printf '%s\n' "$lage" | sed -n 1p)"
gleich "die Lage nennt die Zahlen auch ohne Befund" \
  "npm audit: ohne-befund — kritisch 0 · hoch 0 · mittel 2 · niedrig 1" "$(printf '%s\n' "$lage" | sed -n 2p)"

# ----------------------------------------------------------- Der Beleg

# **Die Lage gehört in den Freigabebeleg** (AK-03): drei Zustände, und jeder
# steht dort mit seinem Wort — ein nicht erreichbarer Scanner ist kein grüner.
printf 'govulncheck: ohne-befund\nnpm audit: nicht-erreichbar — 503\n' > "$RAUM/lage"
printf 'abc  MagicSetEditor4Mac-9.9.9.dmg\n' > "$RAUM/summen"
belegschreiben "$RAUM/beleg.md" "9.9.9" "deadbeef" "abc" "$RAUM/summen" "" "$RAUM/lage"

if grep -q '^## Sicherheitslage' "$RAUM/beleg.md" \
  && grep -qF -- '- govulncheck: ohne-befund' "$RAUM/beleg.md" \
  && grep -qF -- '- npm audit: nicht-erreichbar — 503' "$RAUM/beleg.md"; then
  gleich "der Beleg nennt die Lage beider Scanner" "ja" "ja"
else
  gleich "der Beleg nennt die Lage beider Scanner" "ja" "nein"
fi

belegschreiben "$RAUM/ohne.md" "9.9.9" "deadbeef" "abc" "$RAUM/summen"
if grep -q 'Sicherheitslage: nicht erhoben' "$RAUM/ohne.md"; then
  gleich "ohne Lage sagt der Beleg, dass sie fehlt" "ja" "ja"
else
  gleich "ohne Lage sagt der Beleg, dass sie fehlt" "ja" "nein"
fi

# ------------------------------------------------------ Das Urteil der Freigabe

# **Was eine Freigabe mit der Lage tut** — Befund 1 des Zweitblicks über Phase
# 6.2. Hier brach sie nur bei `befund` ab; ein erneuter Scan, der scheiterte,
# endete trotzdem bei „Fertig". In der Pipeline ist derselbe Zustand rot.
urteil() { # urteil <lagedatei> — gibt Rückgabe und Meldung als eine Zeile
  local meldung rueckgabe=0
  meldung="$(freigabeurteil "$1")" || rueckgabe=$?
  printf '%s|%s' "$rueckgabe" "$meldung"
}

printf 'govulncheck: ohne-befund\nnpm audit: ohne-befund — kritisch 0 · hoch 0 · mittel 0 · niedrig 0\n' > "$RAUM/lage-gruen"
gleich "ohne Befund geht die Freigabe weiter, ohne ein Wort" "0|" \
  "$(urteil "$RAUM/lage-gruen")"

printf 'govulncheck: befund — GO-2026-0001\nnpm audit: ohne-befund\n' > "$RAUM/lage-befund"
gleich "ein Befund hält die Freigabe an" \
  "1|eine bekannte Schwachstelle — sie gehört behoben oder begründet, bevor freigegeben wird" \
  "$(urteil "$RAUM/lage-befund")"

printf 'govulncheck: ohne-befund\nnpm audit: gescheitert — npm error code ELOCKVERIFY\n' > "$RAUM/lage-gescheitert"
gleich "ein gescheiterter Scanner hält die Freigabe an" \
  "1|ein Scanner ist gescheitert — ohne seine Antwort ist die Lage unbekannt" \
  "$(urteil "$RAUM/lage-gescheitert")"

printf 'govulncheck: nicht-erreichbar — dial tcp\nnpm audit: ohne-befund\n' > "$RAUM/lage-fern"
gleich "ein Scanner ohne Antwort hält die Freigabe an, solange niemand entscheidet" \
  "1|ein Scanner war nicht erreichbar — später erneut, oder ausdrücklich: MSE_FREIGABE_OHNE_SCANNER=1" \
  "$(unset MSE_FREIGABE_OHNE_SCANNER; urteil "$RAUM/lage-fern")"

# **Mit der Entscheidung geht sie weiter — und die steht im Beleg.** Eine
# Lücke, die jemand bewusst hinnimmt, ist eine andere Aussage als eine, die
# niemand bemerkt hat.
cp "$RAUM/lage-fern" "$RAUM/lage-entschieden"
gleich "wer ausdrücklich entscheidet, gibt frei und hört es" \
  "0|Entscheidung: freigegeben, obwohl ein Scanner nicht antwortete (MSE_FREIGABE_OHNE_SCANNER=1)" \
  "$(MSE_FREIGABE_OHNE_SCANNER=1 urteil "$RAUM/lage-entschieden")"
belegschreiben "$RAUM/entschieden.md" "9.9.9" "deadbeef" "abc" "$RAUM/summen" "" "$RAUM/lage-entschieden"
if grep -qF -- '- Entscheidung: freigegeben, obwohl ein Scanner nicht antwortete' "$RAUM/entschieden.md"; then
  gleich "die Entscheidung steht im Beleg" "ja" "ja"
else
  gleich "die Entscheidung steht im Beleg" "ja" "nein"
fi

# **Ein Befund lässt sich nicht wegentscheiden** — die Variable gilt der
# Lücke, nicht der Schwachstelle.
printf 'govulncheck: nicht-erreichbar — dial tcp\nnpm audit: befund — kritisch 0 · hoch 1 · mittel 0 · niedrig 0\n' > "$RAUM/lage-beides"
gleich "die Entscheidung übergeht keinen Befund" \
  "1|eine bekannte Schwachstelle — sie gehört behoben oder begründet, bevor freigegeben wird" \
  "$(MSE_FREIGABE_OHNE_SCANNER=1 urteil "$RAUM/lage-beides")"

# **Eine Lage, die einen Scanner nicht nennt, ist nicht erhoben** — und keine
# Antwort ist kein grüner Lauf. `sicherheitslage` schreibt immer zwei Zeilen;
# fehlt eine, ist der Lauf unterwegs gestorben.
: > "$RAUM/lage-leer"
gleich "eine leere Lage hält die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(urteil "$RAUM/lage-leer")"

printf 'govulncheck: ohne-befund\n' > "$RAUM/lage-halb"
gleich "eine halbe Lage hält die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(MSE_FREIGABE_OHNE_SCANNER=1 urteil "$RAUM/lage-halb")"

gleich "eine fehlende Lage hält die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(urteil "$RAUM/gibtsnicht")"

# **Und eine angebrochene Zeile ist keine Antwort** — Befund 1 des Nachlaufs
# über Phase 6.2. Hier genügte der Anfang einer Zeile: `govulncheck: ` ohne
# Zustand traf keines der Sperrmuster und ging durch.
printf 'govulncheck: \nnpm audit: ohne-befund\n' > "$RAUM/lage-angebrochen"
gleich "eine Zeile ohne Zustand hält die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(urteil "$RAUM/lage-angebrochen")"

printf 'govulncheck: vielleicht\nnpm audit: ohne-befund\n' > "$RAUM/lage-unbekannt"
gleich "ein unbekannter Zustand hält die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(urteil "$RAUM/lage-unbekannt")"

printf 'govulncheck: ohne-befund\ngovulncheck: befund — GO-2026-0001\nnpm audit: ohne-befund\n' > "$RAUM/lage-doppelt"
gleich "zwei Zeilen für einen Scanner halten die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(urteil "$RAUM/lage-doppelt")"

printf 'govulncheck: ohne-befund\nnpm audit: ohne-befund\nirgendwas\n' > "$RAUM/lage-fremd"
gleich "eine fremde Zeile hält die Freigabe an" \
  "1|die Lage nennt nicht beide Scanner — sie ist nicht erhoben" \
  "$(urteil "$RAUM/lage-fremd")"

printf 'govulncheck: ohne-befund — alles gut\nnpm audit: ohne-befund — kritisch 0 · hoch 0 · mittel 0 · niedrig 0\n' > "$RAUM/lage-mitzusatz"
gleich "ein Zusatz hinter dem Zustand bleibt erlaubt" "0|" "$(urteil "$RAUM/lage-mitzusatz")"

# **Und der Freigabeweg fragt das Urteil** — nicht eine eigene Abfrage daneben,
# die nur einen der vier Zustände kennt. **Ob das Urteil wirkt**, hält seit
# #1830 `releaseschluss.test.sh`: er fährt den Schluss, statt ihn zu lesen.
# shellcheck disable=SC2016 # `$ausgabe` ist Text im Skript, nicht diese Shell
if grep -Eq '^[^#]*freigabeurteil "\$ausgabe/sicherheitslage.txt"' "$WURZEL/scripts/releaseschluss.sh" \
  && ! grep -Eq "^[^#]*grep -q ': befund'" "$WURZEL/scripts/release.sh" "$WURZEL/scripts/releaseschluss.sh"; then
  gleich "der Freigabeweg fragt das Urteil und keine eigene Abfrage" "ja" "ja"
else
  gleich "der Freigabeweg fragt das Urteil und keine eigene Abfrage" "ja" "nein"
fi

# **Und der Freigabeweg erhebt sie** — sonst hielte die Prüfung nur sich selbst.
if grep -Eq '^[^#]*sicherheitslage[^#]*>[^#]*sicherheitslage' "$WURZEL/scripts/releaseschluss.sh"; then
  gleich "der Freigabeweg erhebt die Lage" "ja" "ja"
else
  gleich "der Freigabeweg erhebt die Lage" "ja" "nein"
fi

printf '\n'
if [ "$FEHLER" -eq 0 ]; then
  printf '%s✓ Sicherheitslage in Ordnung%s\n' "$GRUEN" "$AUS"
else
  printf '%s✗ Sicherheitslage fehlerhaft%s\n' "$ROT" "$AUS"
fi

exit "$FEHLER"
