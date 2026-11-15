#!/usr/bin/env bash
#
# Schreibt den Beleg eines grünen Pipelinelaufs — oder verweigert ihn.
#
# ## Wofür
#
# Der `pre-push`-Hook prüft, statt zu rechnen
# (#1389).
# Er liest `.git/mse-pipeline`, und dieser Beleg sagt: **für diesen Commit,
# bei sauberem Baum, war die Pipeline grün.**
#
# ## Warum der Stand am Anfang mitkommt
#
# **Befund 4 des Zweitblicks über die Phasen 4.0 bis 4.9.** Die erste Fassung
# las Commit und Sauberkeit erst **am Ende** des Laufs. Wer während der zehn
# Minuten committet, bekam damit einen Beleg über einen Stand, den keine
# einzige Stufe gesehen hat — und der Hook ließ ihn durch.
#
# Das ist kein erfundener Fall: genau so wird in diesem Repositorium
# gearbeitet, und am 2026-09-02 ist es beinahe passiert.
#
# Der Aufrufer nimmt den Stand **vor** der ersten Stufe auf und gibt ihn
# hier hinein. Weicht er ab, gibt es keinen Beleg.
#
# ## Aufruf
#
#     scripts/stempel.sh <stand-am-anfang>
#     scripts/stempel.sh --stand            # gibt den Stand aus

set -euo pipefail

# stand nennt Commit und Arbeitsbaum in einer Zeile.
#
# **Der Baum geht als Prüfsumme über `git status --porcelain` ein**, nicht als
# bloßes „sauber/schmutzig": zwei verschiedene schmutzige Bäume sind nicht
# derselbe Stand, und ein Lauf über den einen belegt nichts über den anderen.
stand() {
  printf '%s %s' \
    "$(git rev-parse HEAD 2>/dev/null || echo keincommit)" \
    "$(git status --porcelain 2>/dev/null | shasum -a 256 | cut -d' ' -f1)"
}

if [ "${1:-}" = "--stand" ]; then
  stand
  printf '\n'
  exit 0
fi

# **Ob der Stand noch der vom Anfang ist** — QA-07.AK-03
# (#1968).
# Ein Lauf, dessen Baum sich geändert hat, gibt keinen Beleg mehr; jede
# weitere Stufe wäre Zeit für nichts.
if [ "${1:-}" = "--aktuell" ]; then
  [ "${2:-}" = "$(stand)" ]
  exit $?
fi

anfang="${1:-}"
if [ -z "$anfang" ]; then
  printf 'stempel.sh braucht den Stand vom Anfang des Laufs.\n' >&2
  exit 2
fi

ende="$(stand)"

if [ "$anfang" != "$ende" ]; then
  printf '\033[33m  · kein Beleg: der Stand hat sich während des Laufs geändert\033[0m\n'
  printf '    vorher:  %s\n' "$anfang"
  printf '    nachher: %s\n' "$ende"

  # **Und der alte Beleg wird weggeräumt.** Ein Beleg von vorhin, der stehen
  # bleibt, ist schlimmer als keiner: der Hook liest ihn und lässt durch.
  verzeichnis="$(git rev-parse --git-dir 2>/dev/null)" || exit 0
  rm -f "$verzeichnis/mse-pipeline"

  exit 0
fi

verzeichnis="$(git rev-parse --git-dir)"

if [ -n "$(git status --porcelain)" ]; then
  baumstand="schmutzig"
else
  baumstand="sauber"
fi

printf '%s %s\n' "$(git rev-parse HEAD)" "$baumstand" > "$verzeichnis/mse-pipeline"
