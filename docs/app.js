// Interfaccia della versione web: stesso comportamento di main.go dell'app Windows.
import { interpreta, formatta, orologio, orologioBreve, descrivi, formattaDigitazione, OraNonValida } from "./ore.js";

const $ = (id) => document.getElementById(id);
const input = $("calcolo");
const anteprima = $("anteprima");
const risultato = $("risultato");
const orologioEl = $("orologio");
const lista = $("storico");

// vuoto tiene occupata la riga di un'etichetta senza testo, così l'interfaccia non salta.
const VUOTO = " ";
const CHIAVE = "calcolatrice-ore.storico";
const MAX_STORICO = 500;

let storico = caricaStorico(); // [{ testo, minuti }], il più recente per primo

function messaggioErrore(err) {
  return err instanceof OraNonValida ? "⚠ Ora non valida: " + err.testo : "⚠ " + err.message;
}

function mostraAnteprima(testo, errore) {
  anteprima.textContent = testo;
  anteprima.classList.toggle("errore", errore);
  input.classList.toggle("errore", errore);
}

// inserisce i due punti mentre si scrive e aggiorna l'anteprima.
input.addEventListener("input", () => {
  const testo = input.value;
  const [nuovo, cur] = formattaDigitazione(testo, input.selectionEnd ?? testo.length);
  if (nuovo !== testo) {
    input.value = nuovo;
    input.setSelectionRange(cur, cur);
  }
  aggiornaAnteprima();
});

function aggiornaAnteprima() {
  const t = input.value.trim();
  if (t === "") return mostraAnteprima(VUOTO, false);
  try {
    mostraAnteprima("Letto come:  " + descrivi(interpreta(t)), false);
  } catch (err) {
    if (!(err instanceof OraNonValida) && (t.endsWith("+") || t.endsWith("-"))) {
      mostraAnteprima(VUOTO, false); // sta ancora scrivendo il prossimo orario
    } else {
      mostraAnteprima(messaggioErrore(err), true);
    }
  }
}

function mostraRisultato(minuti) {
  risultato.textContent = formatta(minuti);
  orologioEl.textContent = orologio(minuti) || VUOTO;
}

function calcola() {
  let r;
  try {
    r = interpreta(input.value);
  } catch (err) {
    mostraAnteprima(messaggioErrore(err), true);
    return;
  }
  mostraRisultato(r.minuti);

  let testo = descrivi(r) + " = " + formatta(r.minuti);
  const o = orologioBreve(r.minuti);
  if (o) testo += "  (" + o + ")";
  storico.unshift({ testo, minuti: r.minuti });
  storico = storico.slice(0, MAX_STORICO);
  aggiornaLista();
  salvaStorico();

  // il testo resta selezionato: si può correggerlo o digitare il prossimo calcolo
  input.focus();
  input.select();
}

// azzera svuota il campo e riporta il risultato a 0:00.
function azzera() {
  input.value = "";
  mostraAnteprima(VUOTO, false);
  mostraRisultato(0);
  input.focus();
}

async function copia() {
  const testo = risultato.textContent;
  try {
    await navigator.clipboard.writeText(testo);
  } catch {
    // senza permesso agli appunti (o pagina non sicura) si copia dal campo
    const prima = input.value;
    input.value = testo;
    input.select();
    document.execCommand("copy");
    input.value = prima;
  }
  input.focus();
}

// riusa rimette nel campo il calcolo di una riga dello storico, es. "18:00 + 3:00".
function riusa(i) {
  const v = storico[i];
  if (!v) return;
  const espr = v.testo.split(" = ")[0];
  input.value = espr;
  input.focus();
  input.setSelectionRange(espr.length, espr.length);
  aggiornaAnteprima();
}

function svuota() {
  if (storico.length === 0) return;
  if (!confirm("Cancellare tutto lo storico?")) return;
  storico = [];
  aggiornaLista();
  salvaStorico();
  input.focus();
}

function aggiornaLista() {
  lista.replaceChildren();
  if (storico.length === 0) {
    const li = document.createElement("li");
    li.className = "vuoto";
    li.textContent = "Nessun calcolo";
    lista.append(li);
    return;
  }
  storico.forEach((v, i) => {
    const li = document.createElement("li");
    li.textContent = v.testo;
    li.title = v.testo;
    li.tabIndex = 0;
    li.dataset.i = i;
    lista.append(li);
  });
}

lista.addEventListener("dblclick", (e) => {
  const li = e.target.closest("li[data-i]");
  if (li) riusa(Number(li.dataset.i));
});
lista.addEventListener("keydown", (e) => {
  const li = e.target.closest("li[data-i]");
  if (li && e.key === "Enter") riusa(Number(li.dataset.i));
});

// Lo storico resta nel browser (localStorage) di questo computer.
function caricaStorico() {
  try {
    const dati = JSON.parse(localStorage.getItem(CHIAVE) || "[]");
    return Array.isArray(dati) ? dati.filter((v) => typeof v?.testo === "string") : [];
  } catch {
    return [];
  }
}

function salvaStorico() {
  try {
    localStorage.setItem(CHIAVE, JSON.stringify(storico));
  } catch {
    // navigazione privata o spazio pieno: lo storico vale solo per questa sessione
  }
}

input.addEventListener("keydown", (e) => {
  if (e.key === "Enter") {
    e.preventDefault();
    calcola();
  }
});
document.addEventListener("keydown", (e) => {
  if (e.key === "Escape") azzera();
});
$("calcola").addEventListener("click", calcola);
$("copia").addEventListener("click", copia);
$("azzera").addEventListener("click", azzera);
$("svuota").addEventListener("click", svuota);

aggiornaLista();
input.focus();

// Funziona anche offline e si può installare come app (Chrome/Edge: icona nella barra degli indirizzi).
if ("serviceWorker" in navigator && location.protocol !== "file:") {
  navigator.serviceWorker.register("sw.js").catch(() => {});
}
