#!/usr/bin/env bash
#
# Die **Herkunft einer Freigabe** — I06
# (#2264).
#
# # Wofür
#
# Der Freigabebeleg bindet das Abbild an einen Commit; was **darin** liegt und
# womit es gebaut wurde, sagte bis hierher keine Datei. `herkunft.json` nennt
# Commit, ob der Baum sauber war, die gepinnten Werkzeuge, die SHA-256 jedes
# Programms im Bündel, beider Stücklisten und des Abbilds. **Sie reist in
# `pruefsummen.txt` mit und ist damit signiert** — deshalb läuft dieser
# Schritt vor dem `shasum`-Aufruf, wie `releasemanifest.sh`.
#
# # Was sie nicht sagt
#
# **Dass ein zweiter Bau dieselben Bytes gäbe.** Das sagt nur ein Vergleich
# zweier Bauläufe (`scripts/baugleich.sh`, Beleg `docs/belege/baugleich.md`);
# diese Datei hält fest, **was** gebaut wurde, nicht dass es sich wiederholen
# ließe.

# herkunft_schreiben legt `herkunft.json` neben die übrigen Artefakte.
herkunft_schreiben() { # herkunft_schreiben <ausgabe> <app> <wurzel>
  local AUSGABE="$1" APP="$2" WURZEL="$3"
  local COMMIT SAUBER ABBILD PROGRAMME STUECKLISTEN

  kopf "Herkunft"

  [ -d "$APP/Contents/MacOS" ] || ende "kein Bündel unter $APP — die Herkunft hätte keine Programme"

  ABBILD="$(ls "$AUSGABE"/*.dmg 2>/dev/null | head -1)"
  [ -n "$ABBILD" ] || ende "kein Abbild — die Herkunft hätte nichts zu beschreiben"

  COMMIT="$(git -C "$WURZEL" rev-parse HEAD)"
  SAUBER=false
  [ -z "$(git -C "$WURZEL" status --porcelain)" ] && SAUBER=true

  PROGRAMME="$(
    for DATEI in "$APP"/Contents/MacOS/*; do
      [ -f "$DATEI" ] || continue
      printf '    {"name": "%s", "sha256": "%s"},\n' \
        "$(basename "$DATEI")" "$(shasum -a 256 "$DATEI" | cut -d' ' -f1)"
    done | sed '$ s/,$//'
  )"
  [ -n "$PROGRAMME" ] || ende "das Bündel trägt keine Programme"

  STUECKLISTEN="$(
    for DATEI in "$AUSGABE"/stueckliste-*.json; do
      [ -f "$DATEI" ] || continue
      printf '    {"name": "%s", "sha256": "%s"},\n' \
        "$(basename "$DATEI")" "$(shasum -a 256 "$DATEI" | cut -d' ' -f1)"
    done | sed '$ s/,$//'
  )"

  # **Die Fassung jedes Werkzeugs aus der Pinnung**, nicht aus dem PATH
  # geraten — und die Toolchain aus `go env`, denn `GOTOOLCHAIN` entscheidet,
  # welche wirklich gebaut hat (AGENTS.md §1).
  cat > "$AUSGABE/herkunft.json" <<HERKUNFT
{
  "commit": "$COMMIT",
  "sauber": $SAUBER,
  "werkzeuge": {
    "go": "$(go env GOVERSION)",
    "wails": "$(awk '$1=="github.com/wailsapp/wails/v3"{print $2}' "$WURZEL/go.mod")",
    "node": "$(node --version 2>/dev/null)",
    "npm": "$(npm --version 2>/dev/null)",
    "cyclonedx-gomod": "$(awk '$1=="github.com/CycloneDX/cyclonedx-gomod" && $2 ~ /^v/ {print $2}' "$WURZEL/tools/go.mod")",
    "macos": "$(sw_vers -productVersion 2>/dev/null)",
    "clang": "$(xcrun clang --version 2>/dev/null | head -1)"
  },
  "abbild": {"name": "$(basename "$ABBILD")", "sha256": "$(shasum -a 256 "$ABBILD" | cut -d' ' -f1)"},
  "programme": [
$PROGRAMME
  ],
  "stuecklisten": [
$STUECKLISTEN
  ]
}
HERKUNFT
  ok "herkunft.json ($(printf '%s' "$PROGRAMME" | grep -c '"name"') Programme)"
}
