// Package ore interpreta e calcola espressioni di ore come "8:30 + 9 - 45m".
package ore

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Termine è un operando dell'espressione con il suo segno.
type Termine struct {
	Segno  int // +1 o -1
	Minuti int
	Testo  string // testo originale digitato
}

// Risultato di un calcolo.
type Risultato struct {
	Termini []Termine
	Minuti  int
}

// Interpreta legge un'espressione come "1430 - 8,30 + 9" e calcola il totale.
// Se l'espressione inizia con + o -, il calcolo parte da `precedente`
// (se non nil), così si può continuare dall'ultimo risultato.
func Interpreta(espr string, precedente *int) (Risultato, error) {
	s := strings.TrimSpace(espr)
	if s == "" {
		return Risultato{}, errors.New("inserisci un orario")
	}

	var res Risultato
	if (s[0] == '+' || s[0] == '-') && precedente != nil {
		res.Termini = append(res.Termini, Termine{Segno: 1, Minuti: *precedente, Testo: Formatta(*precedente)})
	}

	segno := 1
	attesoOperando := true
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '+' || c == '-':
			if attesoOperando {
				if c == '-' {
					segno = -segno
				}
			} else {
				segno = 1
				if c == '-' {
					segno = -1
				}
				attesoOperando = true
			}
			i++
		default:
			j := i
			for j < len(s) && !strings.ContainsRune("+- \t", rune(s[j])) {
				j++
			}
			tok := s[i:j]
			if !attesoOperando {
				return Risultato{}, fmt.Errorf("manca + o - prima di %q", tok)
			}
			m, err := Operando(tok)
			if err != nil {
				return Risultato{}, err
			}
			res.Termini = append(res.Termini, Termine{Segno: segno, Minuti: m, Testo: tok})
			attesoOperando = false
			i = j
		}
	}
	if len(res.Termini) == 0 {
		return Risultato{}, errors.New("inserisci un orario")
	}
	if attesoOperando {
		return Risultato{}, errors.New("l'espressione finisce con un operatore")
	}
	for _, t := range res.Termini {
		res.Minuti += t.Segno * t.Minuti
	}
	return res, nil
}

// MaxMinuti è il valore massimo di un singolo orario: 24:00.
const MaxMinuti = 24 * 60

// OraNonValida è l'errore per un valore che non può essere un orario,
// per esempio "28" o "14:75".
type OraNonValida struct{ Testo string }

func (e *OraNonValida) Error() string { return fmt.Sprintf("ora non valida: %s", e.Testo) }

// Operando converte un singolo valore in minuti. Formati accettati:
//
//	9      → 9:00     (1-2 cifre = ore)
//	930    → 9:30     (3-4 cifre: le ultime due sono i minuti)
//	1430   → 14:30
//	14:30, 14.30, 14,30 → 14:30
//	:45    → 0:45
//	45m    → 0:45
//
// Ogni valore deve stare fra 0:00 e 24:00, altrimenti è *OraNonValida.
func Operando(tok string) (int, error) {
	t := strings.ToLower(strings.TrimSpace(tok))
	sbagliato := fmt.Errorf("non capisco %q", tok)
	nonValida := &OraNonValida{tok}
	if t == "" {
		return 0, sbagliato
	}

	if strings.HasSuffix(t, "m") || strings.HasSuffix(t, "min") || strings.HasSuffix(t, "'") {
		n := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(t, "min"), "m"), "'")
		v, err := cifre(n)
		if err != nil {
			return 0, sbagliato
		}
		if v > MaxMinuti {
			return 0, nonValida
		}
		return v, nil
	}

	var h, m int
	if k := strings.IndexAny(t, ":.,"); k >= 0 {
		hs, ms := t[:k], t[k+1:]
		if strings.ContainsAny(ms, ":.,") || len(ms) == 0 {
			return 0, sbagliato
		}
		if hs != "" {
			var err error
			if h, err = cifre(hs); err != nil {
				return 0, sbagliato
			}
		}
		var err error
		if m, err = cifre(ms); err != nil {
			return 0, sbagliato
		}
		if len(ms) > 2 {
			return 0, nonValida
		}
		if len(ms) == 1 { // 14,3 → 14:30, come si legge a voce
			m *= 10
		}
	} else {
		v, err := cifre(t)
		if err != nil {
			return 0, sbagliato
		}
		switch {
		case len(t) <= 2:
			h = v
		case len(t) <= 4:
			h, m = v/100, v%100
		default:
			return 0, nonValida
		}
	}
	if m > 59 || h*60+m > MaxMinuti {
		return 0, nonValida
	}
	return h*60 + m, nil
}

// FormattaDigitazione mette i due punti mentre l'utente scrive:
// "830" → "8:30", "1430" → "14:30", "12,59" → "12:59".
// cursore è la posizione del cursore nel testo; viene restituita
// la posizione corrispondente nel testo formattato.
func FormattaDigitazione(s string, cursore int) (string, int) {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return s, cursore // solo testo ASCII
		}
	}
	if cursore < 0 || cursore > len(s) {
		cursore = len(s)
	}
	s = strings.NewReplacer(",", ":", ".", ":").Replace(s)

	var b strings.Builder
	for i := 0; i < len(s); {
		if strings.IndexByte("+- \t", s[i]) >= 0 {
			b.WriteByte(s[i])
			i++
			continue
		}
		j := i
		for j < len(s) && strings.IndexByte("+- \t", s[j]) < 0 {
			j++
		}
		b.WriteString(formattaToken(s[i:j]))
		i = j
	}
	nuovo := b.String()

	// il cursore resta dopo lo stesso numero di caratteri (esclusi i due punti)
	k := strings.Count(s[:cursore], ":")
	k = cursore - k
	pos := 0
	for pos < len(nuovo) && k > 0 {
		if nuovo[pos] != ':' {
			k--
		}
		pos++
	}
	for pos < len(nuovo) && nuovo[pos] == ':' {
		pos++
	}
	return nuovo, pos
}

func formattaToken(t string) string {
	if h, m, ok := strings.Cut(t, ":"); ok {
		// "1:43" + "0" → "1:430": le cifre vanno ridistribuite → "14:30"
		if len(m) == 3 && len(h)+len(m) <= 4 && soloCifre(h) && soloCifre(m) {
			return formattaToken(h + m)
		}
		return t
	}
	if !soloCifre(t) {
		return t
	}
	switch len(t) {
	case 3:
		return t[:1] + ":" + t[1:]
	case 4:
		return t[:2] + ":" + t[2:]
	}
	return t
}

func soloCifre(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func cifre(s string) (int, error) {
	if !soloCifre(s) || len(s) > 6 {
		return 0, errors.New("non numerico")
	}
	return strconv.Atoi(s)
}

// Formatta restituisce i minuti come "h:mm" (con segno se negativi).
func Formatta(minuti int) string {
	segno := ""
	if minuti < 0 {
		segno = "-"
		minuti = -minuti
	}
	return fmt.Sprintf("%s%d:%02d", segno, minuti/60, minuti%60)
}

// Orologio indica a che ora del giorno corrisponde il risultato quando
// esce dalle 24 ore, es. "05:00 del giorno dopo". Vuoto se 0:00-23:59.
func Orologio(minuti int) string {
	const giorno = 24 * 60
	if minuti >= 0 && minuti < giorno {
		return ""
	}
	giorni := minuti / giorno
	resto := minuti % giorno
	if resto < 0 {
		resto += giorno
		giorni--
	}
	ora := fmt.Sprintf("%02d:%02d", resto/60, resto%60)
	switch giorni {
	case 1:
		return ora + " del giorno dopo"
	case -1:
		return ora + " del giorno prima"
	}
	if giorni > 0 {
		return fmt.Sprintf("%s, +%d giorni", ora, giorni)
	}
	return fmt.Sprintf("%s, %d giorni", ora, giorni)
}

// Descrivi mostra come è stata interpretata l'espressione, es. "8:30 + 9:00".
func Descrivi(r Risultato) string {
	var b strings.Builder
	for i, t := range r.Termini {
		switch {
		case i == 0 && t.Segno < 0:
			b.WriteString("-")
		case i > 0 && t.Segno < 0:
			b.WriteString(" - ")
		case i > 0:
			b.WriteString(" + ")
		}
		b.WriteString(Formatta(t.Minuti))
	}
	return b.String()
}

// OrologioBreve è come Orologio ma compatto, per lo storico: "05:00 +1g".
func OrologioBreve(minuti int) string {
	const giorno = 24 * 60
	if minuti >= 0 && minuti < giorno {
		return ""
	}
	giorni := minuti / giorno
	resto := minuti % giorno
	if resto < 0 {
		resto += giorno
		giorni--
	}
	return fmt.Sprintf("%02d:%02d %+dg", resto/60, resto%60, giorni)
}
