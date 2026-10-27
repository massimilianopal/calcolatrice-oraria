package ore

import "testing"

func TestOperando(t *testing.T) {
	casi := map[string]int{
		"9": 540, "09": 540, "0": 0, "930": 570, "1430": 870, "0830": 510,
		"14:30": 870, "14.30": 870, "14,30": 870, "14,3": 870, "8:05": 485,
		":45": 45, "45m": 45, "90min": 90, "24": 1440, "24:00": 1440, "2400": 1440, "23:59": 1439,
	}
	for in, want := range casi {
		got, err := Operando(in)
		if err != nil || got != want {
			t.Errorf("Operando(%q) = %d, %v; voglio %d", in, got, err, want)
		}
	}
	for _, in := range []string{"abc", "1:2:3", "14:", "-", "7,5h"} {
		if _, err := Operando(in); err == nil {
			t.Errorf("Operando(%q) doveva dare errore", in)
		}
	}
	for _, in := range []string{"28", "25", "1460", "14:60", "24:01", "2401", "2830", "10030", "1500m", "14:305"} {
		_, err := Operando(in)
		if _, ok := err.(*OraNonValida); !ok {
			t.Errorf("Operando(%q) = %v; voglio ora non valida", in, err)
		}
	}
}

func TestInterpreta(t *testing.T) {
	casi := map[string]string{
		"8:30 + 9":         "17:30",
		"830+9":            "17:30",
		"1430 - 8,30":      "6:00",
		"8 - 9":            "-1:00",
		"-2 + 5":           "3:00",
		"20 + 9":           "29:00",
		"8:30 + 9 - 45m":   "16:45",
		"1745-0830-:30":    "8:45",
		"  7,30  +  0,30 ": "8:00",
	}
	for in, want := range casi {
		r, err := Interpreta(in)
		if err != nil || Formatta(r.Minuti) != want {
			t.Errorf("Interpreta(%q) = %s, %v; voglio %s", in, Formatta(r.Minuti), err, want)
		}
	}
	for _, in := range []string{"", "8 9", "8 +", "+", "8 ++"} {
		if _, err := Interpreta(in); err == nil {
			t.Errorf("Interpreta(%q) doveva dare errore", in)
		}
	}
	// nessuna continuazione dal calcolo precedente: "+3" è solo 3:00
	r, err := Interpreta("+3")
	if err != nil || r.Minuti != 180 || Descrivi(r) != "3:00" {
		t.Errorf("+3: %d %v %q", r.Minuti, err, Descrivi(r))
	}
}

func TestOrologio(t *testing.T) {
	casi := map[int]string{
		600: "", 29 * 60: "05:00 del giorno dopo", -60: "23:00 del giorno prima",
		50 * 60: "02:00, +2 giorni", 24 * 60: "00:00 del giorno dopo",
	}
	for in, want := range casi {
		if got := Orologio(in); got != want {
			t.Errorf("Orologio(%d) = %q; voglio %q", in, got, want)
		}
	}
}

func TestOrologioBreve(t *testing.T) {
	if got := OrologioBreve(29 * 60); got != "05:00 +1g" {
		t.Error(got)
	}
	if got := OrologioBreve(-60); got != "23:00 -1g" {
		t.Error(got)
	}
	if got := OrologioBreve(60); got != "" {
		t.Error(got)
	}
}

func TestFormattaDigitazione(t *testing.T) {
	// ogni passo simula la digitazione di un carattere alla fine
	sequenze := map[string]string{
		"830":            "8:30",
		"1430":           "14:30",
		"12,59":          "12:59",
		"14.30":          "14:30",
		"28":             "28",
		"830+9":          "8:30+9",
		"1745 - 0830":    "17:45 - 08:30",
		"45m":            "45m",
		":45":            ":45",
		"8:30 + 1430-:5": "8:30 + 14:30-:5",
	}
	for digitato, want := range sequenze {
		testo := ""
		for _, c := range digitato {
			testo, _ = FormattaDigitazione(testo+string(c), len(testo)+1)
		}
		if testo != want {
			t.Errorf("digitando %q ottengo %q; voglio %q", digitato, testo, want)
		}
	}
	if s, c := FormattaDigitazione("1:430", 5); s != "14:30" || c != 5 {
		t.Errorf("1:430 → %q %d", s, c)
	}
	if s, c := FormattaDigitazione("12,", 3); s != "12:" || c != 3 {
		t.Errorf("12, → %q %d", s, c)
	}
	// cursore in mezzo: "8:30 + 9" con "1" inserito prima del 9
	if s, c := FormattaDigitazione("830 + 19", 7); s != "8:30 + 19" || c != 8 {
		t.Errorf("cursore in mezzo → %q %d", s, c)
	}
}
