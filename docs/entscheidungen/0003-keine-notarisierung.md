# 0003 — Keine Notarisierung, Freigabe über die Systemeinstellungen

**Stand:** angenommen
**Datum:** 2026-08-08

## Lage

macOS prüft heruntergeladene Programme über Gatekeeper. Ohne
Developer-ID-Zertifikat und ohne Notarisierung wird ein Programm beim ersten
Start abgewiesen.

Beides setzt ein **Apple Developer Program** voraus: 99 € im Jahr, ein
Apple-Konto mit hinterlegten Daten, und eine Bindung, die jedes Jahr erneuert
werden will. Ein solches Konto besteht nicht und soll nicht bestehen.

Hinzu kommt, dass auf Apple Silicon jede ausführbare Datei **irgendeine**
Signatur braucht — auch eine ad-hoc erzeugte. Ohne sie startet sie überhaupt
nicht. „Gar nicht signieren" ist also keine Option.

## Entscheidung

Die Anwendung wird **ad-hoc signiert** und **nicht notarisiert**.

```bash
codesign --force --deep --sign - MagicSetEditor4Mac.app
```

Nutzende geben sie beim ersten Start **einmalig** frei:
*Systemeinstellungen* → *Datenschutz & Sicherheit* → Abschnitt *Sicherheit* →
**„Dennoch öffnen"**.

Dieser Weg wird **erklärt, nicht verschwiegen**: in `README.md`, in
`docs/RELEASE.md`, in jeder Release-Notiz und als Textdatei im `.dmg`. Er ist
eine Eigenschaft dieses Projekts, kein Mangel, den man kleinschreibt.

## Verworfen

**Apple Developer Program abschließen.** Verworfen aus einem ausdrücklichen
Wunsch heraus: kein Konto, keine jährliche Bindung. Das ist eine Entscheidung
über das Projekt, keine technische Einschränkung — und sie darf nicht
stillschweigend rückgängig gemacht werden, wenn die Gatekeeper-Meldung lästig
wird.

**Rechtsklick → Öffnen als Weg dokumentieren.** Verworfen, weil es ihn nicht
mehr gibt: **ab macOS 15** hat Apple den Kontextmenü-Bypass entfernt. Eine
Anleitung, die ihn nennt, ist veraltet und schickt Nutzende in eine Sackgasse.

**`xattr -d com.apple.quarantine` als Hauptweg empfehlen.** Verworfen. Das
Kommando funktioniert, schaltet aber eine Sicherheitsprüfung ab. Es steht in
`docs/RELEASE.md` als Erwähnung für Fortgeschrittene, nicht als Schritt 3 einer
Anleitung für alle.

**Über den Mac App Store ausliefern.** Verworfen — setzt dasselbe Konto voraus,
dazu Sandboxing. Eine Anwendung, die beliebige Vorlagenverzeichnisse auf dem
Rechner liest, passt schlecht in eine Sandbox.

## Folgen

- `spctl --assess` **lehnt ab**. Das ist der erwartete Zustand. Wer diesen Befund
  „behebt", hat diese Entscheidung rückgängig gemacht, ohne sie zu ändern —
  `docs/RELEASE.md` sagt das an der Stelle, an der der Befund auftaucht.
- Das Release trägt kein `.pkg` und keinen Installer, nur ein `.dmg`.
- Automatische Aktualisierung über Sparkle o. Ä. wäre möglich, aber ohne
  Signaturkette wenig wert. Nicht vorgesehen.
- Sollte sich die Lage ändern — ein Konto entsteht doch —, ändert sich **zuerst
  diese Datei**, dann `scripts/release.sh`.
