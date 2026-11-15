#!/usr/bin/env bash
#
# Prüfungen des Leistungsvergleichs mit MSE2 — QA-04
# (#1735).
#
# **Sie fahren keines der beiden Programme.** Geprüft werden die Rechnung
# (`scripts/leistungsstatistik.mjs`) und die zwei Sitzungen
# (`scripts/leistungssitzung.mjs`) an Zahlen und Antworten, die der Test
# vorgibt, und die Kartenauswahl (`scripts/leistungsauswahl.mjs`) an einem
# Set aus drei Karten, der Bericht (`scripts/leistungsbericht.mjs`) und das
# Messen eines Prozesses (`scripts/leistungsprozess.mjs`) an einem
# nachgebauten, die Prüfungen beim Messen (`scripts/leistungsphasen.mjs`) und
# welcher Prozess mit welcher Eingabe startet (`scripts/leistungsteilnehmer.mjs`).
# Dazu seit dem Zweitblick über Phase 6.11 (Z1, Z2): die Teile selbst
# (`scripts/leistungsteile.mjs`) an Teilnehmern aus Doubles, der Vergleich der
# Bildmaße (`scripts/leistungsmasse.mjs`), die vollständige Prüfung eines PNG
# (`scripts/pngpruefung.mjs`) und das Abschalten des Guthabens, das ein
# Vergleich nicht messen soll (`scripts/leistungskontingent.mjs`).
# Der Lauf selbst, `node scripts/leistungsvergleich.mjs`, braucht ein gebautes
# MSE2, die Pakete und die kanonischen Sets und steht deshalb in
# `docs/RELEASE.md` und nicht in der Pipeline.
set -euo pipefail

wurzel="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

node --test \
	"$wurzel/scripts/leistungsstatistik.test.mjs" \
	"$wurzel/scripts/leistungsdauer.test.mjs" \
	"$wurzel/scripts/leistungssitzung.test.mjs" \
	"$wurzel/scripts/leistungsauswahl.test.mjs" \
	"$wurzel/scripts/leistungsbericht.test.mjs" \
	"$wurzel/scripts/leistungsprozess.test.mjs" \
	"$wurzel/scripts/leistungsphasen.test.mjs" \
	"$wurzel/scripts/leistungsteilnehmer.test.mjs" \
	"$wurzel/scripts/leistungsteile.test.mjs" \
	"$wurzel/scripts/leistungsmasse.test.mjs" \
	"$wurzel/scripts/leistungskontingent.test.mjs" \
	"$wurzel/scripts/leistungsschnittteile.test.mjs" \
	"$wurzel/scripts/pngpruefung.test.mjs"
