#!/usr/bin/env bash
# **Das Kompatibilitätsmanifest einer Freigabe** — Werkstück O6 von Phase 5.2
# (#1430).
#
# **Eine eigene Datei**, weil `release.sh` mit ihr über die Zeilengrenze aus §2
# ginge — dieselbe Trennung wie bei `ci-reichweite.sh`. Der Schnitt folgt der
# Frage: dort steht der Freigabeweg, hier, **was diese Fassung öffnen kann**.
#
# # Warum es vor den Prüfsummen geschrieben wird
#
# **Damit es in `pruefsummen.txt` steht** — und die wird mit OpenPGP signiert.
# Damit ist das Manifest genauso authentisch wie das Abbild daneben, ohne einen
# zweiten Schlüssel, ein zweites Format und eine zweite Prüfung.
#
# # Und was hier **nicht** entschieden wird
#
# Ob eine Fassung passt. Das Manifest sagt nur, was sie kennt; verglichen wird
# es gegen den echten Ledger eines Bestands (`domain.Vertraeglichkeit`).

# manifest_schreiben legt `manifest.json` ins Ausgabeverzeichnis.
#
#   $1  Ausgabeverzeichnis   $2  Fassung   $3  Repositoriumswurzel
manifest_schreiben() {
  local AUSGABE="$1" VERSION="$2" WURZEL="$3"
  local ABBILD ARTEFAKT WANDERUNGEN GRUNDLINIE

  kopf "Kompatibilitätsmanifest"

  ABBILD="$(ls "$AUSGABE"/*.dmg 2>/dev/null | head -1)"
  [ -n "$ABBILD" ] || ende "kein Abbild — das Manifest hätte nichts zu beschreiben"

  ARTEFAKT="$(shasum -a 256 "$ABBILD" | cut -d' ' -f1)"

  # **Die Wanderungen aus dem Quelltext**, nicht aus einem Bestand: das
  # Manifest sagt, was diese **Fassung** kennt — nicht, was irgendein Bestand
  # verbucht hat.
  #
  # **Mit Prüfsumme je Wanderung**, und das ist der Punkt des Werkstücks: eine
  # nachträglich geänderte Migration trägt dieselbe Nummer und einen anderen
  # Inhalt, und eine Schemanummer sieht das nicht.
  #
  # **Die Schritte der Grundlinie aus ihrem Umfang** (Vermerk 0189): sie
  # liegen nicht mehr als Datei vor, und `umfang.txt` trägt Name und
  # Prüfsumme, wie sie 0.7.0 ausgeliefert hat. Was danach kommt, ist wieder
  # eine Datei unter `migrationen/`.
  WANDERUNGEN="$(
    {
      grep -v -e '^#' -e '^$' "$WURZEL/internal/storage/grundlinie/umfang.txt"
      for DATEI in "$WURZEL"/internal/storage/migrationen/*.sql; do
        [ -e "$DATEI" ] || continue
        printf '%s %s\n' "$(basename "$DATEI")" "$(shasum -a 256 "$DATEI" | cut -d' ' -f1)"
      done
    } | while read -r NAME SUMME; do
      printf '    {"name": "%s", "sha256": "%s"},\n' "$NAME" "$SUMME"
    done | sed '$ s/,$//'
  )"

  [ -n "$WANDERUNGEN" ] || ende "keine Wanderungen gefunden — das Manifest wäre eine Lücke"

  # **Und bis wohin sie nur noch als Ganzes reichen** — eine Fassung, die
  # später ihre Grundlinie verschiebt, sagt so, welcher Bestand sie nicht
  # öffnet (`domain.HindernisVorDerGrundlinie`).
  GRUNDLINIE="$(grep -v -e '^#' -e '^$' "$WURZEL/internal/storage/grundlinie/umfang.txt" | tail -1 | cut -d' ' -f1)"
  [ -n "$GRUNDLINIE" ] || ende "der Umfang der Grundlinie ist leer — das Manifest wäre eine Lücke"

  cat > "$AUSGABE/manifest.json" <<MANIFEST
{
  "fassung": "$VERSION",
  "plattform": "darwin/arm64",
  "artefakt": "$ARTEFAKT",
  "einstellungen": 1,
  "protokoll": 1,
  "grundlinie": "$GRUNDLINIE",
  "wanderungen": [
$WANDERUNGEN
  ]
}
MANIFEST

  ok "manifest.json ($(printf '%s' "$WANDERUNGEN" | grep -c '"name"') Wanderungen)"
}
