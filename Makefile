# MagicSetEditor4Mac — Einstiegspunkte
#
# Verbindlich ist AGENTS.md. Dieses Makefile ruft nur auf, was dort steht.

SHELL := /usr/bin/env bash
.DEFAULT_GOAL := hilfe

.PHONY: hilfe ci format types lint test security tools hooks leitlinien kennzahlen dev dev-browser bau sbom sauber platz fremdpakete belege webkitprobe webkitabnahme huellenprobe fuzzkampagne mutationslauf lastlauf

hilfe: ## diese Übersicht
	@printf '\nMagicSetEditor4Mac\n\n'
	@grep -E '^[a-zA-ZäöüÄÖÜ_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'
	@printf '\nVor jeder Arbeit an Go-Code:  source scripts/activate.sh\n\n'

## ---------------------------------------------------------------- Pipeline

ci: ## die vollständige Pipeline
	@./scripts/ci.sh all

vorpruefung: ## der schnelle Riegel davor — fünf Minuten statt fünfzig
	@./scripts/vorpruefung.sh

format: ## nur die Stufe format
	@./scripts/ci.sh format

types: ## nur die Stufe types
	@./scripts/ci.sh types

lint: ## nur die Stufe lint
	@./scripts/ci.sh lint

test: ## nur die Stufe test
	@./scripts/ci.sh test

security: ## nur die Stufe security
	@./scripts/ci.sh security

## ------------------------------------------------------------- Einrichtung

tools: ## die gepinnten Werkzeuge nach .go/bin bauen
	@if [ ! -f tools/go.mod ]; then \
	  printf '\033[33mtools/go.mod fehlt — siehe Phase 1, Werkzeugpinnung\033[0m\n'; \
	  exit 1; \
	fi
	@cd tools && go install tool
	@printf '\033[32mWerkzeuge in .go/bin\033[0m\n'

hooks: ## den pre-push-Hook einhängen
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-push
	@printf '\033[32mpre-push-Hook aktiv\033[0m\n'

leitlinien: ## CLAUDE.md aus AGENTS.md schreiben
	@./scripts/leitlinien.sh

kennzahlen: ## den Umfang erheben — für den Abschnitt im README
	@go test ./internal/werkzeuge -run TestDerUmfangImReadme -v -count=1 \
		2>&1 | grep -E "Quelltext:|steht|--- " || true

## ------------------------------------------------- Die Läufe gegen fremde Pakete

# **Sie laufen nicht in `ci`** — sie brauchen Pakete, Sets und Referenzbilder,
# die nicht im Repositorium liegen (§11). Damit sind sie die einzige Prüfebene,
# die niemand regelmäßig fährt, und sie sammeln veraltete Zusagen: am
# 2026-08-22 kamen drei auf einmal heraus.
#
# **Die Belegung stand nirgends**, und wer sie rät, bekommt rote Läufe, die
# nichts über den Stand sagen. Hier steht sie.
PACKWURZEL   ?= /Users/david/full-magic-pack/data
SETPAKET     ?= $(HOME)/Library/Application Support/MagicSetEditor4Mac/kanonische-sets/Archenemy.mse-set
REFERENZEN   ?= $(HOME)/Library/Application Support/MagicSetEditor4Mac/referenzbilder

# **Und der Ort der kanonischen Sets** — seit dem 2026-09-14 muss er hier
# stehen. Bis dahin fand ihn `kanonort()` über `$$HOME`; seit die Läufe ein
# eigenes Zuhause bekommen (§3, der Anmeldebund), ist das ein leeres
# Verzeichnis. **Gemerkt am 2026-09-20**: `TestDieZeilenlageGegenMSE2` war
# rot mit „das Paket … ist nicht zu finden" — kein Fund über die Zeilenlage,
# sondern einer über die Umgebung.
KANON        ?= $(HOME)/Library/Application Support/MagicSetEditor4Mac/kanonische-sets

# **Und die zwei, ohne die der Kanonvergleich nicht läuft** — seit dem
# 2026-09-22 hier (QA-01, #1733). `MESSMETHODIK.md` §7 nannte sie, das
# Makefile nicht: `TestDerAbstandZuMSE2AmKanon` sprang deshalb in jedem Lauf
# über dieses Ziel und meldete `ok`, und die Zahlen im README stammten aus
# Läufen von Hand. **`SCHRIFTWURZEL` ist die Wurzel des Packs, nicht sein
# `data`** — die Schriften liegen daneben.
SCHRIFTWURZEL ?= /Users/david/full-magic-pack
MSE2AUSGABE   ?= $(HOME)/Library/Application Support/MagicSetEditor4Mac/mse2-ausgabe

# **Und der Quelltext von MSE2** — drei Wächter hinter `fremdpakete` lesen ihn
# (`MSE2BESTAND.md`, `PARITAET.md`, die Skriptnamen). Er liegt neben dem
# Repositorium (§11) und hieß an einer Stelle `MSE_MSE2QUELLE`, an drei
# anderen `MSE2_QUELLE`; seit dem 2026-09-22 überall das zweite.
MSE2QUELLE    ?= /Users/david/mse2-quelle

# **Und MSE2 selbst, gebaut** — für die Abnahme von Phase X: zeichnet MSE2
# unsere Symboldatei (`symbolorakel_test.go`)? Bis zum 2026-09-22 sprang sie
# in jedem Lauf über dieses Ziel.
MSE2BINAER    ?= /Users/david/mse2-quelle/build/magicseteditor
MSE2DATEN     ?= $(HOME)/Library/Application Support/magicseteditor/data

# **Und die Schriften, mit denen MSE2 die Referenz zeichnet** — QA-03 (#1656).
# MSE2 fragt die installierten Schriften nach dem Namen; fehlt eine, weicht es
# still aus. `TestDieReferenzschriftenReichen` hält jede Familie, die wir aus
# dem Pack nehmen, gegen dieses Verzeichnis — und `scripts/mse2-referenz.sh`
# fragt denselben Lauf, bevor es zeichnet.
MSE2SCHRIFTEN ?= $(HOME)/Library/Fonts/mse2-referenz

# **Der Lauf hinterlässt einen Beleg** — QA-01
# (#1733).
# `internal/laufbeleg/lauf` fährt `go test -json` über die markierten Fälle,
# trennt Ausführung und Ergebnis und schreibt `docs/belege/fremdpakete.json`.
# **Fehlt eine der zehn Eingaben, wird nicht gemessen**, und der Beleg sagt
# BLOCKED — bis hierhin sprang ein solcher Lauf und hieß grün (#1656).
#
# **Der absichtlich rote Lauf ist ein Trend im Plan** und kein `|| true`
# mehr: bis hierhin schluckte das auch einen Übersetzungsfehler.
#
# **Dieselben zehn Eingaben beim Messen und beim Prüfen**: `make belege`
# rechnet ihre Hashes neu und hält sie gegen den Beleg. Ohne sie hieße ein
# Beleg »aktuell«, nachdem MSE2 die Referenzbilder neu gerastert hat —
# angestrichen vom Zweitblick über Phase 6.0.
EINGABEN = MSE_FREMDPAKETE="$(PACKWURZEL)" MSE_SCHRIFTWURZEL="$(SCHRIFTWURZEL)" \
	MSE_SETPAKET="$(SETPAKET)" MSE_REFERENZBILDER="$(REFERENZEN)" \
	MSE_KANON="$(KANON)" MSE_MSE2AUSGABE="$(MSE2AUSGABE)" \
	MSE2_QUELLE="$(MSE2QUELLE)" MSE2_BINAER="$(MSE2BINAER)" MSE2_DATEN="$(MSE2DATEN)" \
	MSE2_SCHRIFTEN="$(MSE2SCHRIFTEN)"

fremdpakete: ## die Läufe gegen fremde Pakete, mit Beleg in docs/belege (nicht Teil von `ci`)
	@$(EINGABEN) go run ./internal/laufbeleg/lauf -plan docs/belege/fremdpakete.plan.json

dienstdrift: ## die Läufe gegen die Kartendienste ohne Konto, mit Beleg in docs/belege (nicht Teil von `ci`) — älter als 30 Tage macht `make belege` rot
	@$(EINGABEN) go run ./internal/laufbeleg/lauf -plan docs/belege/dienstdrift.plan.json

fuzzkampagne: ## fuzzt jedes Ziel aus docs/RELEASE.md mit Zeit-, Kern- und Speicherbudget, Bericht in docs/belege (nicht Teil von `ci`)
	@node scripts/fuzzkampagne.mjs

mutationslauf: ## schaltet je eine kritische Wache ab und fährt die Tests dagegen, Beleg in docs/belege (nicht Teil von `ci`)
	@node scripts/mutationslauf.mjs

lastlauf: ## Lastprofile gegen REST und MCP, mit und ohne Kontingente, Beleg in docs/belege (nicht Teil von `ci`)
	@node scripts/lastlauf.mjs

webkitprobe: ## die Selbstprobe der WebKit-Probe — braucht ein Fenstersystem (nicht Teil von `ci`)
	@./scripts/webkitprobe-selbst.sh

webkitabnahme: ## die WebKit-Abnahmen gegen die gebaute Anwendung, mit Beleg in docs/belege (nicht Teil von `ci`)
	@./scripts/webkitabnahme.sh

huellenprobe: ## die Abnahmen am nativen Fenster — Leiste, Dateidialoge, Tisch —, mit Beleg in docs/belege (nicht Teil von `ci`)
	@./scripts/huellenprobe.sh

belege: ## die Belege gegen den heutigen Stand halten — überfällig, anderer Stand, rot
	@rueckgabe=0; for plan in docs/belege/*.plan.json; do \
		$(EINGABEN) go run ./internal/laufbeleg/lauf -plan "$$plan" -pruefen || rueckgabe=1; \
	done; exit $$rueckgabe

## ------------------------------------------------------------------- Bauen

dev: ## das Fenster im Entwicklungsmodus, die Oberfläche von Vite
	@./scripts/entwickeln.sh fenster

dev-browser: ## dieselbe Anwendung im Browser unter :34115 — für die Sichtprüfung
	@./scripts/entwickeln.sh browser

bau: ## das .app bauen (ad hoc signiert, mit Hardened Runtime, nicht notarisiert)
	@./scripts/bauen.sh freigabe
	@printf '\033[32mbuild/bin/MagicSetEditor4Mac.app\033[0m\n'
	@printf 'Nicht notarisiert — Erststart siehe docs/RELEASE.md\n'

sbom: ## Stücklisten erzeugen, auf dem Weg der Freigabe (CycloneDX in JSON)
	@./scripts/stueckliste.sh

sauber: ## Bauartefakte entfernen
	@rm -rf build/bin .abdeckung.out frontend/dist frontend/coverage
	@printf '\033[32maufgeräumt\033[0m\n'

# **Die Caches liegen im Projekt** (`scripts/activate.sh`, §1) — das hält sie
# aus `~/go` heraus und macht sie **sichtbar**. Sichtbar heißt aber nicht
# begrenzt: Go räumt seinen Build-Cache nur nach Alter auf, und auf einem
# Dateisystem ohne Zugriffszeiten greift diese Heuristik nicht.
#
# **Am 2026-08-28 nachgemessen: 172 GB.** Über Monate gewachsen, ohne dass
# jemand es sah — `du -sh .go` fährt niemand von sich aus. Danach: 2 GB.
#
# **Was hier wegfällt, entsteht beim nächsten Lauf neu.** Der Build-Cache
# kostet einen langsamen `make ci`, der Lint-Cache einen langsamen `make lint`.
# Der **Modul-Cache bleibt** — er zu holen braucht das Netz, und die
# Versionen sind gepinnt.
platz: ## Caches leeren, die still wachsen (Build und Lint)
	@printf '%s\n' "── Vorher"
	@du -sh .go 2>/dev/null || true
	@go clean -cache
	@rm -rf .go/golangci-lint .go/cache
	@printf '\n%s\n' "── Nachher"
	@du -sh .go 2>/dev/null || true
	@printf '\n\033[32mBuild- und Lint-Cache geleert — sie entstehen beim nächsten Lauf neu\033[0m\n'
	@printf 'Der Modul-Cache (.go/mod) bleibt: ihn zu holen braucht das Netz.\n'
