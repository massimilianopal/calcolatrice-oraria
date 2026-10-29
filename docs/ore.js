// Lettura e calcolo di espressioni di ore come "8:30 + 9 - 45m".
// Stessa logica di ore/ore.go dell'app Windows: se cambi una, cambia anche l'altra.

// Valore massimo di un singolo orario: 24:00.
export const MAX_MINUTI = 24 * 60;

// Errore per un valore che non può essere un orario, per esempio "28" o "14:75".
export class OraNonValida extends Error {
  constructor(testo) {
    super(`ora non valida: ${testo}`);
    this.testo = testo;
  }
}

const SEPARATORI = "+- \t";

// interpreta legge un'espressione come "1430 - 8,30 + 9" e calcola il totale.
// Ogni calcolo parte da zero: "+3" vale 3:00.
// Restituisce { termini: [{segno, minuti, testo}], minuti }.
export function interpreta(espr) {
  const s = espr.trim();
  if (s === "") throw new Error("inserisci un orario");

  const termini = [];
  let segno = 1;
  let attesoOperando = true;
  let i = 0;
  while (i < s.length) {
    const c = s[i];
    if (c === " " || c === "\t") {
      i++;
    } else if (c === "+" || c === "-") {
      if (attesoOperando) {
        if (c === "-") segno = -segno;
      } else {
        segno = c === "-" ? -1 : 1;
        attesoOperando = true;
      }
      i++;
    } else {
      let j = i;
      while (j < s.length && !SEPARATORI.includes(s[j])) j++;
      const tok = s.slice(i, j);
      if (!attesoOperando) throw new Error(`manca + o - prima di "${tok}"`);
      termini.push({ segno, minuti: operando(tok), testo: tok });
      attesoOperando = false;
      i = j;
    }
  }
  if (termini.length === 0) throw new Error("inserisci un orario");
  if (attesoOperando) throw new Error("l'espressione finisce con un operatore");
  return { termini, minuti: termini.reduce((tot, t) => tot + t.segno * t.minuti, 0) };
}

// operando converte un singolo valore in minuti. Formati accettati:
//
//   9      → 9:00     (1-2 cifre = ore)
//   930    → 9:30     (3-4 cifre: le ultime due sono i minuti)
//   1430   → 14:30
//   14:30, 14.30, 14,30 → 14:30
//   :45    → 0:45
//   45m    → 0:45
//
// Ogni valore deve stare fra 0:00 e 24:00, altrimenti lancia OraNonValida.
export function operando(tok) {
  const t = tok.trim().toLowerCase();
  const sbagliato = () => new Error(`non capisco "${tok}"`);
  const nonValida = () => new OraNonValida(tok);
  if (t === "") throw sbagliato();

  if (t.endsWith("m") || t.endsWith("min") || t.endsWith("'")) {
    const n = t.replace(/min$/, "").replace(/m$/, "").replace(/'$/, "");
    const v = cifre(n);
    if (v === null) throw sbagliato();
    if (v > MAX_MINUTI) throw nonValida();
    return v;
  }

  let h = 0;
  let m = 0;
  const k = t.search(/[:.,]/);
  if (k >= 0) {
    const hs = t.slice(0, k);
    const ms = t.slice(k + 1);
    if (/[:.,]/.test(ms) || ms.length === 0) throw sbagliato();
    if (hs !== "") {
      h = cifre(hs);
      if (h === null) throw sbagliato();
    }
    m = cifre(ms);
    if (m === null) throw sbagliato();
    if (ms.length > 2) throw nonValida();
    if (ms.length === 1) m *= 10; // 14,3 → 14:30, come si legge a voce
  } else {
    const v = cifre(t);
    if (v === null) throw sbagliato();
    if (t.length <= 2) h = v;
    else if (t.length <= 4) [h, m] = [Math.floor(v / 100), v % 100];
    else throw nonValida();
  }
  if (m > 59 || h * 60 + m > MAX_MINUTI) throw nonValida();
  return h * 60 + m;
}

// formattaDigitazione mette i due punti mentre l'utente scrive:
// "830" → "8:30", "1430" → "14:30", "12,59" → "12:59".
// cursore è la posizione del cursore nel testo; viene restituita
// la posizione corrispondente nel testo formattato.
export function formattaDigitazione(s, cursore) {
  if (/[^\x00-\x7f]/.test(s)) return [s, cursore]; // solo testo ASCII
  if (cursore < 0 || cursore > s.length) cursore = s.length;
  s = s.replace(/[,.]/g, ":");

  let nuovo = "";
  for (let i = 0; i < s.length; ) {
    if (SEPARATORI.includes(s[i])) {
      nuovo += s[i++];
      continue;
    }
    let j = i;
    while (j < s.length && !SEPARATORI.includes(s[j])) j++;
    nuovo += formattaToken(s.slice(i, j));
    i = j;
  }

  // il cursore resta dopo lo stesso numero di caratteri (esclusi i due punti)
  let k = cursore - (s.slice(0, cursore).match(/:/g) || []).length;
  let pos = 0;
  while (pos < nuovo.length && k > 0) {
    if (nuovo[pos] !== ":") k--;
    pos++;
  }
  while (pos < nuovo.length && nuovo[pos] === ":") pos++;
  return [nuovo, pos];
}

function formattaToken(t) {
  const k = t.indexOf(":");
  if (k >= 0) {
    const h = t.slice(0, k);
    const m = t.slice(k + 1);
    // "1:43" + "0" → "1:430": le cifre vanno ridistribuite → "14:30"
    if (m.length === 3 && h.length + m.length <= 4 && soloCifre(h) && soloCifre(m)) {
      return formattaToken(h + m);
    }
    return t;
  }
  if (!soloCifre(t)) return t;
  if (t.length === 3) return t.slice(0, 1) + ":" + t.slice(1);
  if (t.length === 4) return t.slice(0, 2) + ":" + t.slice(2);
  return t;
}

function soloCifre(s) {
  return /^[0-9]+$/.test(s);
}

function cifre(s) {
  if (!soloCifre(s) || s.length > 6) return null;
  return parseInt(s, 10);
}

const due = (n) => String(n).padStart(2, "0");

// formatta restituisce i minuti come "h:mm" (con segno se negativi).
export function formatta(minuti) {
  const segno = minuti < 0 ? "-" : "";
  const a = Math.abs(minuti);
  return `${segno}${Math.floor(a / 60)}:${due(a % 60)}`;
}

function giorniEResto(minuti) {
  const giorno = 24 * 60;
  const giorni = Math.floor(minuti / giorno);
  const resto = minuti - giorni * giorno;
  return [giorni, `${due(Math.floor(resto / 60))}:${due(resto % 60)}`];
}

// orologio indica a che ora del giorno corrisponde il risultato quando
// esce dalle 24 ore, es. "05:00 del giorno dopo". Vuoto se 0:00-23:59.
export function orologio(minuti) {
  if (minuti >= 0 && minuti < MAX_MINUTI) return "";
  const [giorni, ora] = giorniEResto(minuti);
  if (giorni === 1) return `${ora} del giorno dopo`;
  if (giorni === -1) return `${ora} del giorno prima`;
  return giorni > 0 ? `${ora}, +${giorni} giorni` : `${ora}, ${giorni} giorni`;
}

// orologioBreve è come orologio ma compatto, per lo storico: "05:00 +1g".
export function orologioBreve(minuti) {
  if (minuti >= 0 && minuti < MAX_MINUTI) return "";
  const [giorni, ora] = giorniEResto(minuti);
  return `${ora} ${giorni > 0 ? "+" : ""}${giorni}g`;
}

// descrivi mostra come è stata interpretata l'espressione, es. "8:30 + 9:00".
export function descrivi(r) {
  return r.termini
    .map((t, i) => {
      let p = "";
      if (i === 0 && t.segno < 0) p = "-";
      else if (i > 0) p = t.segno < 0 ? " - " : " + ";
      return p + formatta(t.minuti);
    })
    .join("");
}
