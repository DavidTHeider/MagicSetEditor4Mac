# Freigaben für die Arbeit an 0.8.0

**Erteilt vom Auftraggeber am 2026-10-09** mit dem Auftrag, 0.8.0 über die
Phasen A bis J des Dach-Issues
#2204 zu
verwirklichen. Die Leitlinien nennen sie in §17; hier steht, **wofür** jede
gilt und **was keine davon deckt**. Eine Freigabe, die nur im Gesprächsverlauf
steht, kennt die nächste Sitzung nicht.

| Freigabe | Wofür ausdrücklich |
| --- | --- |
| **Shell-Befehle** | Bauen, Testen, Messen, Proben, `make ci` und alles, was ein Paket an Werkzeugen braucht — **gepinnt** und unter `.go/` wie in §1, nie global |
| **VM-Steuerung** | Lume-Gäste anlegen, starten, über VNC fahren, Netze dazwischen schalten und wieder löschen — für Plattform-, Quarantäne-, Netz- und Schlüsselbundproben (A03, D03/D04, I01–I03) |
| **Computer Use** | das native Fenster der **Abnahme-** oder **Prüfkennung** bedienen: Zeiger, Kontextmenü, Dateidialoge, Drucken und Teilen, VoiceOver (A04, H01–H03) |
| **Netzwerk** | Abrufe für Quellen, Fassungen und Dokumentation; Probendienste auf diesem Rechner und zwischen Gästen |
| **Schlüsselbund** | die bis dahin offene Freigabe: Einträge **dieses Programms** in einem **eigenen Prüfbund** und im Wegwerf-Gast (I01, [0181](entscheidungen/0181-der-schluesselbund-bleibt-ungebunden-solange-ad-hoc-signiert-wird.md)) |
| **Ein-Klick-Einrichtung** | `mse`, `mse-mcp` und `mse-rest` aus dem Fenster heraus einrichten und starten — samt der Rechte, die das auf diesem Rechner braucht, und der Wahl, **auf welchen Schnittstellen** ein Dienst lauscht |

## Was keine dieser Freigaben deckt

- **Echte Bestände, Konten und Geheimnisse** bleiben unberührt — Proben laufen
  mit eigenem `HOME`, eigenem Bestand und eigenem Bund, nie im echten
  `~/Library` (§3). **Die sichere Prüfgrenze für das Fenster ist eine eigene
  Kennung oder ein Gast**: der native Dateidialog schreibt über `cfprefsd`
  unter der Kennung, `HOME` und `CFFIXED_USER_HOME` halten ihn nicht auf
  (Befund F-02 des Audits, A01).
- **Systemeinstellungen** — Sicherheit, Datenschutz, Gatekeeper, Bedienungshilfen
  — legt ein Mensch um. Eine Probe, die eine davon braucht, sagt das, statt sie
  still zu ändern (H02).
- **Kennwörter werden nicht eingegeben**, auch nicht im Gast; ein Dialog, der
  eines verlangt, ist ein Halt und kein Hindernis.
- **Downloads großer Abbilder** — ein macOS-Wiederherstellungsabbild, ein
  Modell — bleiben angekündigt: Name, Quelle und Größe stehen im Issue, bevor
  sie geladen werden.

## Was sie nicht ändern

Testgetrieben bleibt testgetrieben, ein Push braucht seinen grünen Lauf, und
eine Messung trägt Commit, Maschine und Grenzen. Die Freigaben nehmen nur die
Rückfrage weg, **ob** ein Werkzeug benutzt werden darf — nicht die Frage, **ob
es das Richtige misst**.
