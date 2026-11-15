# Beleg: I03 — Quarantäne und Aufstieg am Wirt

**Stand:** Commit `58a13b926` · **Datum:** 2026-10-11 05:18 UTC · **Befehl:** `node scripts/aufstieggast.mjs MagicSetEditor4Mac-0.7.0.dmg --wirt` · **Rückgabe:** 0

**Maschine:** am Wirt (Positivkontrolle, ohne Gatekeeper) — macOS 27.0.1 (26A434), Mac17,3, 10 Kerne.

**Bündel:** Abnahmebau, Baustand `2c7004644582`, Programm `56b90f30d50dda77…`, gebaut mit `minos 15.2`.

**Altes Bündel:** das Freigabeabbild von 0.7.0, `577d25f752988ec3…` laut der mit `gpg` geprüften `pruefsummen.txt`.

**Ergebnis:** alle 22 Schritte gelungen.

| Teil | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Gatekeeper | 0.7.0-Abbild prüfen | gelungen | 563 | 577d25f752988ec3 |
| Gatekeeper | Abbild einhängen | gelungen | 4766 | 1 |
| 0.7.0 | Fassung | gelungen | 765 | 0.7.0 |
| 0.7.0 | Set einfahren | gelungen | 119 | 1 |
| 0.7.0 | Karte lesen | gelungen | 26 | 1 |
| 0.7.0 | Kartenbild einsetzen | gelungen | 372 | 1 |
| 0.7.0 | Deck aus dem Set | gelungen | 28 | 1 |
| 0.7.0 | zweiter Nutzer samt Profil | gelungen | 77 | 2 |
| 0.7.0 | Inhalt | gelungen | 132 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| 0.7.0 | Schema | gelungen | 16 | 97 |
| Kandidat | Aufstieg | gelungen | 934 | 1 |
| Kandidat | Schema | gelungen | 10 | 97 |
| Kandidat | Inhalt | gelungen | 140 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| Kandidat | derselbe Inhalt | gelungen | 0 | 1 |
| Kandidat | die Karte mit dem Bild zeichnen | gelungen | 54 | 38050 |
| Kandidat | das Fenster über dem Bestand | gelungen | 641 | 505 |
| Kandidat | keine Wanderung, keine Sicherung davor | gelungen | 0 | 1 |
| Kandidat | ein Wert nach dem Aufstieg | gelungen | 45 | 1 |
| Kandidat | Inhalt nach dem Schreiben | gelungen | 131 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| Rückweg | 0.7.0 liest den Bestand des Kandidaten | gelungen | 26 | 1 |
| Rückweg | Inhalt mit 0.7.0 | gelungen | 138 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| Rückweg | derselbe Inhalt wie nach dem Schreiben | gelungen | 0 | 1 |

- **Schema** (verbuchte Wanderungen): vor dem Aufstieg 97, danach 97 — keine Wanderung; der Rückweg ist das alte Bündel selbst.
- **Sicherung vor der Wanderung**: keine.
- **Inhalt** — **Inhalte, nicht Anzahlen**: je Karte ihre Werte, je Deck seine Plätze, je Bild seine Bytes (als Abdruck), dazu Nutzer samt Profilen:
  vorher `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`; nach dem Aufstieg `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`; nachdem der Kandidat einen Wert schrieb `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`; nach dem Rückweg `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`.
- **Nicht gebraucht**: die Sicherung vor einer Wanderung — es gab keine. Der Rückweg ist das alte Bündel, wie `docs/ADMINISTRATION.md` („Die Appfassung tauschen") ihn beschreibt.
