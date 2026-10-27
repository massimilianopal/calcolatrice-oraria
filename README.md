# Calcolatrice ore

Piccola app per Windows che somma e sottrae orari. Un solo file `CalcolatriceOre.exe`, nessuna installazione.

**Scaricare:** apri la pagina [Releases](../../releases), scarica `CalcolatriceOre.zip`, estrai e avvia `CalcolatriceOre.exe`.

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

- **Invio** calcola, **Esc** cancella il campo.
- Dopo il calcolo il testo resta selezionato: scrivi subito il prossimo.
- Se inizi con `+` o `-` il calcolo parte dall'ultimo risultato (es. `- 45m`).
- Doppio clic su una riga dello storico la riprende come risultato da cui continuare.
- Lo storico è salvato in `%APPDATA%\CalcolatriceOre\storico.txt`.

## Sviluppo

Scritta in Go con [walk](https://github.com/lxn/walk) (interfaccia nativa Windows). Si compila anche da Linux o macOS: serve Go, poi `./build.sh`.
La logica (lettura degli orari e calcoli) è in `ore/`, con i test: `go test ./ore`.

Ad ogni push GitHub Actions compila l'exe; creando un tag `vX.Y.Z` lo pubblica nelle Releases.
