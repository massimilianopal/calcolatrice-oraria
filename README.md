# Calcolatrice ore

Piccola app per Windows che somma e sottrae orari. Un solo file `CalcolatriceOre.exe`, nessuna installazione.

**Versione web (senza installare niente):** <https://massimilianopal.github.io/calcolatrice-oraria/>
Utile dove l'antivirus blocca l'exe. Funziona anche offline e con Chrome o Edge si può installare come app (icona "Installa" nella barra degli indirizzi). Lo storico resta nel browser di quel computer.

**Scaricare l'app Windows:** apri la pagina [Releases](../../releases), scarica `CalcolatriceOre.zip`, estrai e avvia `CalcolatriceOre.exe`.

## Come si scrive

I due punti vengono inseriti da soli mentre scrivi.

| Digiti                    | Diventa / significa |
|---------------------------|---------------------|
| `9`                       | 9 ore               |
| `830`                     | 8:30                |
| `1430`                    | 14:30               |
| `14,30` o `14.30`         | 14:30               |
| `:45` o `45m`             | 45 minuti           |
| `28`, `25:00`, `14:75`    | **ora non valida**  |

Ogni orario deve stare fra 0:00 e 24:00. Il risultato invece può superare le 24 ore:
`20:30 + 9` → 29:30 (05:30 del giorno dopo).

- **Invio** calcola, **Esc** (o il tasto **Azzera**) svuota il campo e riporta il risultato a 0:00.
- Ogni calcolo parte da zero: `+3` vale semplicemente 3:00.
- Dopo il calcolo il testo resta selezionato: scrivi subito il prossimo.
- Doppio clic su una riga dello storico rimette quel calcolo nel campo (es. `18:00 + 3:00`).
- Lo storico è salvato in `%APPDATA%\CalcolatriceOre\storico.txt` (app Windows) o nel browser (versione web).

## Sviluppo

Scritta in Go con [walk](https://github.com/lxn/walk) (interfaccia nativa Windows). Si compila anche da Linux o macOS: serve Go, poi `./build.sh`.
La logica (lettura degli orari e calcoli) è in `ore/`, con i test: `go test ./ore`.

La versione web è in `docs/`: HTML e JavaScript senza passaggi di build, pubblicata da GitHub Pages (Settings → Pages → Deploy from a branch → `main`, cartella `/docs`). La logica è in `docs/ore.js`, copia di `ore/ore.go`: se ne cambi una, cambia anche l'altra. Test: `node --test docs/ore.test.mjs`. Per provarla in locale: `python3 -m http.server --directory docs`.

Ad ogni push GitHub Actions compila l'exe; creando un tag `vX.Y.Z` lo pubblica nelle Releases.
