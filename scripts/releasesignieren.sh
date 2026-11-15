#!/usr/bin/env bash
#
# Die **GPG-Signatur** über die Freigabeartefakte
# (#892).
#
# ## Was hier signiert wird — und was das deckt
#
# **Eine Signatur, nicht drei.** Signiert wird `pruefsummen.txt`; die
# Prüfsummen decken das `.dmg` und beide Stücklisten. Wer die Signatur prüft
# und danach `shasum -c` fährt, hat die ganze Freigabe geprüft — und muss
# dafür genau **einen** Schlüssel kennen.
#
# Drei einzelne Signaturen wären nicht mehr wert und drei Dateien mehr: eine
# gültige Signatur über ein `.dmg`, dessen Prüfsumme in einer unsignierten
# Liste steht, sagt nichts über die Stücklisten daneben.
#
# ## Warum überhaupt, wo doch ad hoc signiert wird
#
# Weil beides verschiedene Fragen beantwortet
# ([0003](../docs/entscheidungen/0003-keine-notarisierung.md)):
#
#   * `codesign --sign -` sagt dem **Betriebssystem**, dass das Bündel seit
#     dem Bau nicht verändert wurde. Es sagt **nicht**, wer es gebaut hat —
#     eine Ad-hoc-Signatur trägt keine Identität, und genau deshalb weist
#     Gatekeeper sie ab.
#   * Die GPG-Signatur sagt, **wer** es gebaut hat. Seit 1.0.0 ist es ein
#     eigener Freigabeschlüssel, getrennt vom Commitschlüssel (Vermerke 0194
#     und 0205); die Anwendung trägt ihn eingebettet.
#
# Ohne sie ist ein heruntergeladenes `.dmg` eine Datei aus dem Netz, deren
# Prüfsumme in einer Datei daneben steht — und die kann jeder mitliefern, der
# beides austauscht.

# signaturschluessel gibt den Schlüssel, mit dem signiert wird.
#
# **Der, gegen den die Anwendung prüft** — `Signaturschluessel` in
# `internal/aktualisierung/signatur.go`. Bis 0.9.0 war es `user.signingkey`,
# weil Commit- und Freigabeschlüssel derselbe waren; seit 1.0.0 sind sie
# getrennt (Vermerk 0205), und die Git-Einstellung nennt den Commitschlüssel.
# Ihn hier ein zweites Mal aufzuschreiben hieße, zwei Wahrheiten über
# dieselbe Sache zu führen.
signaturschluessel() {
  anwendungsschluessel
}

# anwendungsschluessel gibt den Schlüssel, den die **Anwendung** erwartet.
#
# Er steht als Konstante in `internal/aktualisierung/signatur.go` und ist der
# Schlüssel, gegen den ein geladenes Abbild geprüft wird.
anwendungsschluessel() { # anwendungsschluessel [quelle]
  local quelle="${1:-$(dirname "${BASH_SOURCE[0]}")/../internal/aktualisierung/signatur.go}"

  sed -n 's/^const Signaturschluessel = "\(.*\)"$/\1/p' "$quelle" | head -1
}

# schluesselstimmtueberein prüft, ob signiert wird, womit geprüft wird.
#
# # Die Naht, die sonst niemand ansieht
#
# `signaturschluessel` nennt die kurze Kennung aus der Quelle; signiert wird
# mit dem, was `gpg` auf **diesem Rechner** unter ihr findet. Die ausgelieferte
# Anwendung prüft gegen den Schlüssel, der **in ihr** eingebettet ist. Laufen
# die auseinander, ist alles grün: signiert wird,
# geprüft wird, der Fingerabdruck steht im Beleg — und **jede installierte
# Anwendung weist die Freigabe ab**.
#
# Der Fehler wäre still bis zum ersten Nutzer. Angestrichen vom Nachlauf des
# Gesamtreviews vor 0.2.0.
#
# **Verglichen wird der Fingerabdruck der wirklich erzeugten Signatur**, nicht
# die Kennung: was am Ende signiert hat, sagt nur die Signatur.
schluesselstimmtueberein() { # schluesselstimmtueberein <datei> [quelle]
  local datei="$1" erwartet
  erwartet="$(anwendungsschluessel "${2:-}")"

  if [ -z "$erwartet" ]; then
    printf 'die Anwendung nennt keinen Signaturschlüssel — internal/aktualisierung/signatur.go\n' >&2

    return 1
  fi

  local abdruck
  abdruck="$(signaturfingerabdruck "$datei")"

  case "$abdruck" in
    *"$erwartet") return 0 ;;
    *)
      printf 'signiert wurde mit %s, die Anwendung prüft gegen %s — jede installierte Fassung wiese diese Freigabe ab\n' \
        "${abdruck:-nichts}" "$erwartet" >&2

      return 1
      ;;
  esac
}

# signiere legt eine abgetrennte Signatur neben eine Datei.
#
# **Abgetrennt und ASCII-gepanzert** (`--detach-sign --armor`): die Datei
# selbst bleibt, wie sie ist — `shasum -c pruefsummen.txt` läuft unverändert
# —, und `.asc` ist Text, den ein Release-Anhang unbeschadet überträgt.
#
# Gibt zurück:
#   0 — signiert
#   1 — kein Schlüssel, oder gpg fehlt, oder das Signieren scheiterte
signiere() { # signiere <datei> [schluessel]
  local datei="$1"
  # **`${2-…}` und nicht `${2:-…}`.** Jenes behandelt einen leeren Wert wie
  # einen fehlenden und fiele auf den Schlüssel aus der Git-Einstellung
  # zurück — ein Aufruf, der ausdrücklich „ohne Schlüssel" sagt, signierte
  # dann doch. Gefunden vom Testlauf, nicht beim Schreiben.
  local schluessel="${2-$(signaturschluessel)}"

  if [ ! -f "$datei" ]; then
    printf 'die Datei %s gibt es nicht\n' "$datei"

    return 1
  fi

  if ! command -v gpg >/dev/null 2>&1; then
    printf 'gpg fehlt — ohne es gibt es keine Signatur\n'

    return 1
  fi

  if [ -z "$schluessel" ]; then
    printf 'kein Signaturschlüssel — internal/aktualisierung/signatur.go nennt keinen\n'

    return 1
  fi

  # **`--batch --yes`**, damit ein zweiter Lauf die alte Signatur ersetzt,
  # statt auf eine Rückfrage zu warten, die in einem Skript niemand
  # beantwortet.
  if ! gpg --batch --yes --local-user "$schluessel" \
    --detach-sign --armor --output "$datei.asc" "$datei" 2>&1; then
    printf 'das Signieren von %s scheiterte\n' "$datei"

    return 1
  fi

  return 0
}

# signaturpruefen prüft eine abgetrennte Signatur.
#
# **Der Gegentest zum Signieren, im selben Lauf.** Eine Signatur, die niemand
# prüft, ist eine Datei — und ob sie zu ihrem Inhalt passt, sagt nur die
# Prüfung. Derselbe Satz wie bei jedem Wächter (§10).
signaturpruefen() { # signaturpruefen <datei>
  local datei="$1"

  if [ ! -f "$datei.asc" ]; then
    printf 'zu %s gibt es keine Signatur\n' "$datei"

    return 1
  fi

  if ! gpg --batch --verify "$datei.asc" "$datei" >/dev/null 2>&1; then
    printf 'die Signatur zu %s passt nicht\n' "$datei"

    return 1
  fi

  return 0
}

# signaturfingerabdruck gibt den Fingerabdruck, der zu einer Signatur gehört.
#
# **Für den Freigabebeleg.** Ein Beleg, der „signiert" sagt und nicht **womit**,
# lässt die eine Frage offen, für die es ihn gibt.
#
# # Über `--status-fd` und nicht über die Meldung
#
# **Die Meldung von `gpg` steht in der Sprache des Rechners.** Auf Deutsch
# heißt sie „mittels EDDSA-Schlüssel 4D31…", auf Englisch „using EDDSA key
# 4D31…" — und die erste Fassung suchte darin „Fingerabdruck … = …", also eine
# Zeile, die `gpg` nur mit `--with-fingerprint` und nur beim Anzeigen eines
# Schlüssels schreibt. Sie fand nie etwas.
#
# **Aufgefallen ist es am Freigabelauf für 0.2.0**, der „Signatur nachgesehen:
# unbekannt" sagte und trotzdem weiterlief: die Prüfung selbst war in Ordnung,
# nur ihr Ergebnis kam nicht heraus. Ein Wächter, der die halbe Zusage hält,
# macht die andere unsichtbar (§10).
#
# `--status-fd 1` schreibt maschinenlesbare Zeilen, die von der Sprache
# unabhängig sind:
#
#	[GNUPG:] VALIDSIG <fingerabdruck> <datum> …
#
# **`VALIDSIG` und nicht `GOODSIG`**: jenes trägt den Fingerabdruck des
# Schlüssels, dieses nur seine kurze Kennung und den Namen.
signaturfingerabdruck() { # signaturfingerabdruck <datei>
  local datei="$1"

  gpg --batch --status-fd 1 --verify "$datei.asc" "$datei" 2>/dev/null |
    awk '$1 == "[GNUPG:]" && $2 == "VALIDSIG" { print $3; exit }'
}
