#!/usr/bin/env bash
#
# **Der Freigabeweg ab der Sicherheitslage** — aus `release.sh` gezogen
# (#1830),
# damit ein Lauf ihn **ausführen** kann, statt seinen Text zu lesen
# (`releaseschluss.test.sh`).
#
# Der Aufrufer stellt bereit: `kopf`, `ok`, `ende` (bricht mit `exit 1` ab),
# `GELB`, `AUS` und `WURZEL`; dazu die Funktionen aus `releasebeleg.sh` und
# `sicherheitslage.sh`.

# freigabeschluss schreibt den Beleg, fragt davor das Urteil über die Lage —
# und endet bei „Fertig" nur, wenn es die Freigabe erlaubt.
#
# **`BELEGHASH` bleibt gesetzt**: `release.sh` nennt ihn danach im Befehl, mit
# dem die Handprüfungen eingetragen werden.
freigabeschluss() { # freigabeschluss <ausgabe> <fassung> <fingerabdruck>
  local ausgabe="$1" fassung="$2" fingerabdruck="$3" meldung

  kopf "Freigabebeleg"

  # **Der Beleg entsteht hier und nicht später**, denn hier liegen die Zahlen:
  # der Commit, der gerade gebaut wurde, und die Prüfsumme des Abbilds, das
  # gerade entstanden ist. Wer ihn danach von Hand schriebe, schriebe ab.
  BELEGHASH="$(artefakthash "$ausgabe/pruefsummen.txt")" ||
    ende "der Artefakthash ließ sich nicht lesen — ohne ihn ist der Beleg wertlos"

  # **Die Sicherheitslage gehört in den Beleg** (ENG-04.AK-03): mit ihrem
  # Wort, denn ein Scanner, der nicht antwortete, ist nicht grün. **Ob es
  # weitergeht, sagt `freigabeurteil`** — ein Befund und ein gescheiterter
  # Lauf halten an, eine fehlende Antwort nur ohne ausdrückliche Entscheidung.
  sicherheitslage ./internal/... ./cmd/... . > "$ausgabe/sicherheitslage.txt"
  sed 's/^/      /' "$ausgabe/sicherheitslage.txt"
  meldung="$(freigabeurteil "$ausgabe/sicherheitslage.txt")" || ende "$meldung"
  [ -z "$meldung" ] || printf '      %s%s%s\n' "$GELB" "$meldung" "$AUS"

  belegschreiben "$ausgabe/$(belegname "$fassung")" "$fassung" \
    "$(git -C "$WURZEL" rev-parse HEAD)" "$BELEGHASH" "$ausgabe/pruefsummen.txt" \
    "$fingerabdruck" "$ausgabe/sicherheitslage.txt"

  ok "$(belegname "$fassung") — sieben Handprüfungen stehen offen darin"

  kopf "Fertig"
  printf '  %s\n' "$ausgabe"/*
}
