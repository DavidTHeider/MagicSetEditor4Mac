# Die fünf kleinen Modelle an der Laufzeit

**2026-10-04 09:09** — AI-02 (#1743), Commit 880b04a78,
Apple M5, 32 GB, macOS 27.0.1, Ollama 0.35.1.

    ollama serve
    node scripts/modellmessung.mjs

## Gepinnt

| Modell | Zweck | Digest | Größe | Format | Quantisierung | Parameter | Kontext | Stellen | Lizenz | Fähigkeiten |
| --- | --- | --- | ---: | --- | --- | ---: | ---: | ---: | --- | --- |
| qwen3:0.6b | Text | `7df6b6e09427` | 523 MB | gguf | Q4_K_M | 751,6 Mio. | 40960 | 1024 | Apache License | completion, tools, thinking |
| gemma3:1b | Text | `8648f39daa8f` | 815 MB | gguf | Q4_K_M | 999,9 Mio. | 32768 | 1152 | Gemma Terms of Use | completion |
| all-minilm:22m | Einbettung | `1b226e2802db` | 46 MB | gguf | F16 | 22,6 Mio. | 512 | 384 | Apache License | embedding |
| embeddinggemma:300m | Einbettung | `85462619ee72` | 622 MB | gguf | BF16 | 307,6 Mio. | 2048 | 768 | Gemma Terms of Use | embedding |
| x/flux2-klein:4b | Bild | `50a0c0ab15ac` | 5.727 MB | safetensors | — | — | — | — | Apache License | image |

## Laden, Speicher, Stabilität

Je Modell 3 Kaltstarts aus dem entladenen Zustand; **Ladezeit** ist `load_duration` der Laufzeit, **Wanduhr** die ganze erste Anfrage. Der Speicher stammt aus `vmmap --summary` des rechnenden Prozesses nach Laden und Arbeit: **eigen** ist »Physical footprint (peak)« samt GPU-Puffern, **Gewichte** die residente Größe der eingeblendeten Modelldatei, die nicht zum Fußabdruck zählt. **Ollama plant** nennt, was `/api/ps` angibt — eine Schätzung der Laufzeit, keine Messung.

| Modell | Ladezeit (Median) | Wanduhr (Median) | eigen (Spitze) | Gewichte | zusammen | Ollama plant | Fenster | Arbeit | Ergebnis |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| qwen3:0.6b | 0,77 s | 0,81 s | 3.865 MB | 542 MB | 4.408 MB | 4.529 MB | 32768 | 10 Antworten, Startwert 7 | alle gleich, 216 Token/s |
| gemma3:1b | 0,54 s | 0,57 s | 333 MB | 825 MB | 1.158 MB | 999 MB | 32768 | 10 Antworten, Startwert 7 | alle gleich, 111 Token/s |
| all-minilm:22m | 0,26 s | 0,27 s | 93 MB | 71 MB | 164 MB | 26 MB | 256 | 120 Texte eingebettet | 0,52 s, 384 Stellen |
| embeddinggemma:300m | 0,52 s | 0,54 s | 189 MB | 641 MB | 830 MB | 679 MB | 2048 | 120 Texte eingebettet | 1,46 s, 768 Stellen |

## Ein Modell, das denken kann

Die 32 Fragen aus `testdaten/ki/testsatz.json`, gestellt wie `internal/ollama` sie stellt — Systemzeile `KIKartenanweisung` aus dem Quelltext, Schema, Temperatur 0 —, einmal **ohne** das Feld `think` (bis AI-02) und einmal mit `false` (seitdem). Je Frage höchstens 30 s.

| Modell | Frage | angekommen | Dauer (Median) | langsamste | Token (Median) |
| --- | --- | ---: | ---: | ---: | ---: |
| qwen3:0.6b | ohne think | 31 von 32 | 1,83 s | 30,00 s | 267 |
| qwen3:0.6b | think: false | 32 von 32 | 0,16 s | 0,49 s | 14 |

**Und der Speicher wächst mit**: qwen3:0.6b nach der festen Arbeit 3.865 MB eigen, nach der Denkprobe 5.691 MB — die Spitze ist eine des Prozesses.

## Das Bildmodell

**x/flux2-klein:4b**: **nicht ausgeführt** — die Antworten der Laufzeit, wörtlich:

| Weg | Antwort |
| --- | --- |
| `/api/images` | 404 404 page not found |
| `/v1/images/generations` | 404 404 page not found |
| `/api/generate` | 400 {"error":"image generation models are not currently supported"} |
