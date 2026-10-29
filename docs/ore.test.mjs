// Test della logica della versione web: node --test docs/
// Sono gli stessi casi di ore/ore_test.go.
import { test } from "node:test";
import assert from "node:assert/strict";
import {
  operando, interpreta, formatta, orologio, orologioBreve, descrivi,
  formattaDigitazione, OraNonValida,
} from "./ore.js";

test("operando", () => {
  const casi = {
    "9": 540, "09": 540, "0": 0, "930": 570, "1430": 870, "0830": 510,
    "14:30": 870, "14.30": 870, "14,30": 870, "14,3": 870, "8:05": 485,
    ":45": 45, "45m": 45, "90min": 90, "24": 1440, "24:00": 1440, "2400": 1440, "23:59": 1439,
  };
  for (const [inp, voglio] of Object.entries(casi)) assert.equal(operando(inp), voglio, inp);
  for (const inp of ["abc", "1:2:3", "14:", "-", "7,5h"]) assert.throws(() => operando(inp), inp);
  for (const inp of ["28", "25", "1460", "14:60", "24:01", "2401", "2830", "10030", "1500m", "14:305"]) {
    assert.throws(() => operando(inp), OraNonValida, inp);
  }
});

test("interpreta", () => {
  const casi = {
    "8:30 + 9": "17:30", "830+9": "17:30", "1430 - 8,30": "6:00", "8 - 9": "-1:00",
    "-2 + 5": "3:00", "20 + 9": "29:00", "8:30 + 9 - 45m": "16:45",
    "1745-0830-:30": "8:45", "  7,30  +  0,30 ": "8:00",
  };
  for (const [inp, voglio] of Object.entries(casi)) assert.equal(formatta(interpreta(inp).minuti), voglio, inp);
  for (const inp of ["", "8 9", "8 +", "+", "8 ++"]) assert.throws(() => interpreta(inp), inp);
  // nessuna continuazione dal calcolo precedente: "+3" è solo 3:00
  const r = interpreta("+3");
  assert.equal(r.minuti, 180);
  assert.equal(descrivi(r), "3:00");
});

test("orologio", () => {
  assert.equal(orologio(600), "");
  assert.equal(orologio(29 * 60), "05:00 del giorno dopo");
  assert.equal(orologio(-60), "23:00 del giorno prima");
  assert.equal(orologio(50 * 60), "02:00, +2 giorni");
  assert.equal(orologio(24 * 60), "00:00 del giorno dopo");
  assert.equal(orologioBreve(29 * 60), "05:00 +1g");
  assert.equal(orologioBreve(-60), "23:00 -1g");
  assert.equal(orologioBreve(60), "");
});

test("formattaDigitazione", () => {
  // ogni passo simula la digitazione di un carattere alla fine
  const sequenze = {
    "830": "8:30", "1430": "14:30", "12,59": "12:59", "14.30": "14:30", "28": "28",
    "830+9": "8:30+9", "1745 - 0830": "17:45 - 08:30", "45m": "45m", ":45": ":45",
    "8:30 + 1430-:5": "8:30 + 14:30-:5",
  };
  for (const [digitato, voglio] of Object.entries(sequenze)) {
    let testo = "";
    for (const c of digitato) [testo] = formattaDigitazione(testo + c, testo.length + 1);
    assert.equal(testo, voglio, digitato);
  }
  assert.deepEqual(formattaDigitazione("1:430", 5), ["14:30", 5]);
  assert.deepEqual(formattaDigitazione("12,", 3), ["12:", 3]);
  // cursore in mezzo: "8:30 + 9" con "1" inserito prima del 9
  assert.deepEqual(formattaDigitazione("830 + 19", 7), ["8:30 + 19", 8]);
});
