#!/usr/bin/env bash
# Aktiviert die projektlokale Go-Umgebung — das Gegenstück zu `source venv/bin/activate`.
#
# Verwendung:
#   source scripts/activate.sh     # aktivieren
#   go_deactivate                  # verlassen
#
# Wirkung: sämtliche Go-Artefakte (Modul-Cache, Build-Cache, Werkzeug-Binärdateien)
# landen unter .go/ im Projekt statt in ~/go und ~/Library/Caches/go-build.
# Die Compiler-Version bestimmt die `toolchain`-Zeile in go.mod (AGENTS.md §1).

if [ -n "${BASH_SOURCE[0]:-}" ]; then
  _mse_script="${BASH_SOURCE[0]}"
elif [ -n "${(%):-%x}" ] 2>/dev/null; then
  _mse_script="${(%):-%x}"
else
  echo "activate.sh: bitte mit 'source' in bash oder zsh aufrufen" >&2
  return 1 2>/dev/null || exit 1
fi

MSE_ROOT="$(cd "$(dirname "$_mse_script")/.." && pwd)"
unset _mse_script

if [ -n "${MSE_ACTIVE:-}" ]; then
  echo "mse4mac: Umgebung ist bereits aktiv ($MSE_ROOT)"
  return 0 2>/dev/null || exit 0
fi

# Vorherigen Zustand sichern, damit go_deactivate ihn wiederherstellen kann.
MSE_ALT_PATH="$PATH"
MSE_ALT_GOPATH="${GOPATH-}"
MSE_ALT_GOBIN="${GOBIN-}"
MSE_ALT_GOMODCACHE="${GOMODCACHE-}"
MSE_ALT_GOCACHE="${GOCACHE-}"
MSE_ALT_GOLANGCI="${GOLANGCI_LINT_CACHE-}"
MSE_ALT_GOFLAGS="${GOFLAGS-}"
MSE_ALT_GOTOOLCHAIN="${GOTOOLCHAIN-}"
MSE_ALT_ZIELSYSTEM="${MACOSX_DEPLOYMENT_TARGET-}"
MSE_ALT_CGO_CFLAGS="${CGO_CFLAGS-}"
MSE_ALT_CGO_LDFLAGS="${CGO_LDFLAGS-}"
MSE_ALT_PS1="${PS1-}"
export MSE_ALT_PATH MSE_ALT_GOPATH MSE_ALT_GOBIN MSE_ALT_GOMODCACHE \
       MSE_ALT_GOCACHE MSE_ALT_GOLANGCI MSE_ALT_GOFLAGS MSE_ALT_GOTOOLCHAIN \
       MSE_ALT_ZIELSYSTEM MSE_ALT_CGO_CFLAGS MSE_ALT_CGO_LDFLAGS MSE_ALT_PS1

export GOPATH="$MSE_ROOT/.go"
export GOBIN="$MSE_ROOT/.go/bin"
export GOMODCACHE="$MSE_ROOT/.go/mod"
export GOCACHE="$MSE_ROOT/.go/build"

# golangci-lint legt seinen Zwischenspeicher sonst unter ~/Library/Caches ab —
# außerhalb des Projekts, entgegen §1, und über Projekte hinweg geteilt.
export GOLANGCI_LINT_CACHE="$MSE_ROOT/.go/golangci-lint"
export GOFLAGS="-mod=mod"

# **Die in go.mod gepinnte Toolchain gilt verbindlich — und `auto` setzt das
# nicht durch.**
#
# Die `go`-Direktive nennt eine **Untergrenze**: ist die Toolchain des
# Rechners neuer, nimmt Go sie. Am 2026-09-01 hat Homebrew mitten in einer
# Sitzung von 1.26.6 auf 1.27.0 gehoben, und die Stufe `lint` wurde rot —
# golangci-lint v2.12.2 kann die Standardbibliothek von 1.27 nicht lesen
# (`math/rand/v2`: „method must have no type parameters"). Der Quelltext war
# unverändert.
#
# **Dieselbe Falle wie bei `packageManager`** (AGENTS.md §1): eine Angabe ist
# eine Erklärung, wirksam wird sie erst durch etwas, das sie durchsetzt. Die
# Zeile hier ist dieses Etwas — sie nennt die Fassung ausdrücklich, und Go
# lädt sie bei Bedarf nach.
#
# **Wer Go hebt, hebt es hier und in `go.mod`** — und sieht nach, ob die
# gepinnten Werkzeuge die neue Standardbibliothek lesen können.
export GOTOOLCHAIN=go1.27.4

# **Die macOS-Untergrenze für cgo** — dieselbe Datei wie in Pipeline und Bau
# (ENG-02); warum, steht dort.
# shellcheck source=scripts/macos.sh
. "$MSE_ROOT/scripts/macos.sh"

export PATH="$GOBIN:$PATH"
export MSE_ACTIVE=1
export MSE_ROOT

mkdir -p "$GOBIN" "$GOMODCACHE" "$GOCACHE"

# shellcheck disable=SC2016
PS1="(mse4mac) ${PS1-}"
export PS1

go_deactivate() {
  [ -z "${MSE_ACTIVE:-}" ] && { echo "mse4mac: keine aktive Umgebung"; return 1; }

  PATH="$MSE_ALT_PATH"; export PATH
  PS1="$MSE_ALT_PS1"; export PS1

  _mse_restore() {
    # $1 = Variablenname, $2 = gesicherter Wert, $3 = "1" wenn ursprünglich gesetzt
    if [ "$3" = "1" ]; then export "$1=$2"; else unset "$1"; fi
  }
  _mse_restore GOPATH             "$MSE_ALT_GOPATH"     "$([ -n "$MSE_ALT_GOPATH" ] && echo 1)"
  _mse_restore GOBIN              "$MSE_ALT_GOBIN"      "$([ -n "$MSE_ALT_GOBIN" ] && echo 1)"
  _mse_restore GOMODCACHE         "$MSE_ALT_GOMODCACHE" "$([ -n "$MSE_ALT_GOMODCACHE" ] && echo 1)"
  _mse_restore GOCACHE            "$MSE_ALT_GOCACHE"    "$([ -n "$MSE_ALT_GOCACHE" ] && echo 1)"
  _mse_restore GOLANGCI_LINT_CACHE "$MSE_ALT_GOLANGCI"  "$([ -n "$MSE_ALT_GOLANGCI" ] && echo 1)"
  _mse_restore GOFLAGS            "$MSE_ALT_GOFLAGS"    "$([ -n "$MSE_ALT_GOFLAGS" ] && echo 1)"
  _mse_restore GOTOOLCHAIN        "$MSE_ALT_GOTOOLCHAIN" "$([ -n "$MSE_ALT_GOTOOLCHAIN" ] && echo 1)"
  _mse_restore MACOSX_DEPLOYMENT_TARGET "$MSE_ALT_ZIELSYSTEM" "$([ -n "$MSE_ALT_ZIELSYSTEM" ] && echo 1)"
  _mse_restore CGO_CFLAGS         "$MSE_ALT_CGO_CFLAGS" "$([ -n "$MSE_ALT_CGO_CFLAGS" ] && echo 1)"
  _mse_restore CGO_LDFLAGS        "$MSE_ALT_CGO_LDFLAGS" "$([ -n "$MSE_ALT_CGO_LDFLAGS" ] && echo 1)"
  unset -f _mse_restore

  unset MSE_ACTIVE MSE_ROOT
  unset MSE_ALT_PATH MSE_ALT_GOPATH MSE_ALT_GOBIN MSE_ALT_GOMODCACHE
  unset MSE_ALT_GOCACHE MSE_ALT_GOLANGCI MSE_ALT_GOFLAGS MSE_ALT_GOTOOLCHAIN MSE_ALT_PS1
  unset MSE_ALT_ZIELSYSTEM MSE_ALT_CGO_CFLAGS MSE_ALT_CGO_LDFLAGS
  unset -f go_deactivate

  echo "mse4mac: Umgebung verlassen"
}

echo "mse4mac: Umgebung aktiv"
echo "  Wurzel      $MSE_ROOT"
echo "  GOPATH      $GOPATH"
echo "  Go          $(go version 2>/dev/null || echo 'nicht gefunden')"
echo "  Verlassen   go_deactivate"
