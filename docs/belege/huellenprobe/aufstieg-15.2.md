# Beleg: I03 — Quarantäne und Aufstieg auf macOS 15.2

**Stand:** Commit `2ca9a871c` · **Datum:** 2026-10-11 06:29 UTC · **Befehl:** `node scripts/aufstieggast.mjs MagicSetEditor4Mac-0.7.0.dmg mse080-152` · **Rückgabe:** 0

**Maschine:** in der Lume-VM `mse080-152` — macOS 15.2 (24C101), VirtualMac2,1, 4 Kerne.

**Bündel:** Abnahmebau, Baustand `1965cca4f47c`, Programm `1318105eb7adfd33…`, gebaut mit `minos 15.2`.

**Altes Bündel:** das Freigabeabbild von 0.7.0, `577d25f752988ec3…` laut der mit `gpg` geprüften `pruefsummen.txt`.

**Ergebnis:** alle 26 Schritte gelungen.

| Teil | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Gatekeeper | 0.7.0-Abbild prüfen | gelungen | 2598 | 577d25f752988ec3 |
| Gatekeeper | Abbild einhängen | gelungen | 9265 | 1 |
| Gatekeeper | spctl: 0.7.0 aus dem Abbild | gelungen | 2459 | Rückgabe 3: /tmp/i03-Nt7ddZ/quarantaene.app: rejected  |
| Gatekeeper | open: 0.7.0 aus dem Abbild | gelungen | 8275 | nicht gestartet |
| Gatekeeper | spctl: Kandidat, Quarantäne am Bündel | gelungen | 1567 | Rückgabe 3: /tmp/i03-Nt7ddZ/kandidat-q.app: rejected  |
| Gatekeeper | open: Kandidat, Quarantäne am Bündel | gelungen | 8099 | nicht gestartet |
| 0.7.0 | Fassung | gelungen | 1861 | 0.7.0 |
| 0.7.0 | Set einfahren | gelungen | 167 | 1 |
| 0.7.0 | Karte lesen | gelungen | 59 | 1 |
| 0.7.0 | Kartenbild einsetzen | gelungen | 1715 | 1 |
| 0.7.0 | Deck aus dem Set | gelungen | 67 | 1 |
| 0.7.0 | zweiter Nutzer samt Profil | gelungen | 177 | 2 |
| 0.7.0 | Inhalt | gelungen | 304 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| 0.7.0 | Schema | gelungen | 27 | 97 |
| Kandidat | Aufstieg | gelungen | 1098 | 1 |
| Kandidat | Schema | gelungen | 13 | 97 |
| Kandidat | Inhalt | gelungen | 275 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| Kandidat | derselbe Inhalt | gelungen | 0 | 1 |
| Kandidat | die Karte mit dem Bild zeichnen | gelungen | 264 | 38050 |
| Kandidat | das Fenster über dem Bestand | gelungen | 2740 | 2631 |
| Kandidat | keine Wanderung, keine Sicherung davor | gelungen | 0 | 1 |
| Kandidat | ein Wert nach dem Aufstieg | gelungen | 76 | 1 |
| Kandidat | Inhalt nach dem Schreiben | gelungen | 297 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| Rückweg | 0.7.0 liest den Bestand des Kandidaten | gelungen | 76 | 1 |
| Rückweg | Inhalt mit 0.7.0 | gelungen | 321 | {"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1} |
| Rückweg | derselbe Inhalt wie nach dem Schreiben | gelungen | 0 | 1 |

- **Quarantäne**: kenntlich synthetisch, `com.apple.quarantine` mit dem Absender `mse4mac-i03`; das Bündel aus dem eingehängten Abbild erbt `0083;6acb2ce1;mse4mac-i03;`.
- **Der Nutzerweg** — *Datenschutz & Sicherheit → Dennoch öffnen* — ist eine Einstellung und wird von einem Menschen gegeben, nicht von diesem Lauf (`AGENTS.md` §17); **abgenommen wird er im Freigabebeleg** als Handprüfung („Die Schritte 5 und 6 des Erststart-Wegs", `docs/RELEASE.md`). Die Programme der folgenden Teile sind deshalb ohne Quarantäne ausgepackt (`ditto --noqtn`), wie nach dieser Freigabe.
- **Schema** (verbuchte Wanderungen): vor dem Aufstieg 97, danach 97 — keine Wanderung; der Rückweg ist das alte Bündel selbst.
- **Sicherung vor der Wanderung**: keine.
- **Inhalt** — **Inhalte, nicht Anzahlen**: je Karte ihre Werte, je Deck seine Plätze, je Bild seine Bytes (als Abdruck), dazu Nutzer samt Profilen:
  vorher `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`; nach dem Aufstieg `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 2c87a59b66ee","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`; nachdem der Kandidat einen Wert schrieb `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`; nach dem Rückweg `{"bilder":"1 9f98c5f5879e","decks":"Aufstieg=c37fa6b9c9e3","karten":"5 b225d9a55e5a","nutzer":"Ich:Mein Profil;Zweite:Mein Profil,Turnier","sets":1}`.
- **Nicht gebraucht**: die Sicherung vor einer Wanderung — es gab keine. Der Rückweg ist das alte Bündel, wie `docs/ADMINISTRATION.md` („Die Appfassung tauschen") ihn beschreibt.
