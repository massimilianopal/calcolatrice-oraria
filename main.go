//go:build windows

// Calcolatrice ore: somma e sottrae orari digitati in modo veloce
// (1430, 14,30, 14.30 → 14:30) e tiene uno storico dei calcoli.
package main

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"calcolatrice-oraria/ore"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type voce struct {
	testo  string
	minuti int
}

type app struct {
	mw        *walk.MainWindow
	input     *walk.LineEdit
	anteprima *walk.Label
	risultato *walk.Label
	orologio  *walk.Label
	lista     *walk.ListBox

	storico     []voce
	formattando bool
}

const aiuto = "Esempi:  8:30 + 9   ·   1430 - 830   ·   14,30 + 1.15\n" +
	"9 = 9 ore   ·   :45 o 45m = 45 minuti   ·   ogni orario va da 0:00 a 24:00\n" +
	"Invio = calcola   ·   Esc = azzera   ·   doppio clic sullo storico = riprendi il calcolo"

// vuoto tiene occupata la riga di un'etichetta senza testo, così l'interfaccia non salta.
const vuoto = "\u00a0"

var (
	grigio = walk.RGB(110, 110, 110)
	rosso  = walk.RGB(200, 30, 30)
	nero   = walk.RGB(0, 0, 0)
)

func main() {
	a := &app{}
	a.caricaStorico()

	err := MainWindow{
		AssignTo: &a.mw,
		Title:    "Calcolatrice ore",
		Icon:     2, // icona incorporata da rsrc (il manifest ha ID 1)
		Size:     Size{Width: 860, Height: 440},
		MinSize:  Size{Width: 560, Height: 340},
		Font:     Font{Family: "Segoe UI", PointSize: 10},
		Layout:   HBox{},
		Children: []Widget{
			Composite{
				Layout:        VBox{MarginsZero: true},
				StretchFactor: 5,
				Children: []Widget{
					Label{Text: "Calcolo:"},
					LineEdit{
						AssignTo:      &a.input,
						Font:          Font{Family: "Segoe UI", PointSize: 18},
						OnKeyDown:     a.tasto,
						OnTextChanged: a.testoCambiato,
					},
					Label{AssignTo: &a.anteprima, Text: vuoto, TextColor: walk.RGB(110, 110, 110)},
					VSpacer{Size: 6},
					Label{
						AssignTo: &a.risultato,
						Text:     "0:00",
						Font:     Font{Family: "Segoe UI", PointSize: 32, Bold: true},
					},
					Label{AssignTo: &a.orologio, Text: vuoto, TextColor: walk.RGB(110, 110, 110)},
					Composite{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							PushButton{Text: "Calcola (Invio)", OnClicked: a.calcola},
							PushButton{Text: "Copia risultato", OnClicked: a.copia},
							PushButton{Text: "Azzera (Esc)", OnClicked: a.azzera},
							HSpacer{},
						},
					},
					VSpacer{},
					Label{Text: aiuto, TextColor: walk.RGB(110, 110, 110)},
				},
			},
			Composite{
				Layout:        VBox{MarginsZero: true},
				StretchFactor: 4,
				Children: []Widget{
					Label{Text: "Storico:"},
					ListBox{
						AssignTo:        &a.lista,
						Font:            Font{Family: "Consolas", PointSize: 10},
						OnItemActivated: a.riusa,
					},
					PushButton{Text: "Svuota storico", OnClicked: a.svuota},
				},
			},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(nil, "Errore", err.Error(), walk.MsgBoxIconError)
		return
	}
	// Esc va registrato come scorciatoia: il ciclo dei messaggi lo intercetta
	// prima che arrivi al campo di testo.
	esc := walk.NewAction()
	esc.SetShortcut(walk.Shortcut{Key: walk.KeyEscape})
	esc.Triggered().Attach(a.azzera)
	a.mw.ShortcutActions().Add(esc)

	a.aggiornaLista()
	a.input.SetFocus()
	a.mw.Run()
}

func (a *app) tasto(k walk.Key) {
	if k == walk.KeyReturn {
		a.calcola()
	}
}

// testoCambiato inserisce i due punti mentre si scrive e aggiorna l'anteprima.
func (a *app) testoCambiato() {
	if a.formattando {
		return
	}
	testo := a.input.Text()
	_, fine := a.input.TextSelection()
	if nuovo, cur := ore.FormattaDigitazione(testo, fine); nuovo != testo {
		a.formattando = true
		a.input.SetText(nuovo)
		a.input.SetTextSelection(cur, cur)
		a.formattando = false
	}
	a.aggiornaAnteprima()
}

func (a *app) aggiornaAnteprima() {
	t := strings.TrimSpace(a.input.Text())
	if t == "" {
		a.mostraAnteprima(vuoto, false)
		return
	}
	r, err := ore.Interpreta(t)
	if err != nil {
		var nv *ore.OraNonValida
		switch {
		case errors.As(err, &nv):
			a.mostraAnteprima("⚠ Ora non valida: "+nv.Testo, true)
		case strings.HasSuffix(t, "+") || strings.HasSuffix(t, "-"):
			// sta ancora scrivendo il prossimo orario
			a.mostraAnteprima(vuoto, false)
		default:
			a.mostraAnteprima("⚠ "+err.Error(), true)
		}
		return
	}
	a.mostraAnteprima("Letto come:  "+ore.Descrivi(r), false)
}

func (a *app) mostraAnteprima(testo string, errore bool) {
	a.anteprima.SetText(testo)
	if errore {
		a.anteprima.SetTextColor(rosso)
		a.input.SetTextColor(rosso)
	} else {
		a.anteprima.SetTextColor(grigio)
		a.input.SetTextColor(nero)
	}
}

func (a *app) calcola() {
	r, err := ore.Interpreta(a.input.Text())
	if err != nil {
		var nv *ore.OraNonValida
		if errors.As(err, &nv) {
			a.mostraAnteprima("⚠ Ora non valida: "+nv.Testo, true)
		} else {
			a.mostraAnteprima("⚠ "+err.Error(), true)
		}
		return
	}
	ris := ore.Formatta(r.Minuti)
	a.mostraRisultato(r.Minuti)

	testo := ore.Descrivi(r) + " = " + ris
	if o := ore.OrologioBreve(r.Minuti); o != "" {
		testo += "  (" + o + ")"
	}
	a.storico = append([]voce{{testo, r.Minuti}}, a.storico...)
	if len(a.storico) > 500 {
		a.storico = a.storico[:500]
	}
	a.aggiornaLista()
	a.salvaStorico()

	// il testo resta selezionato: si può correggerlo o digitare il prossimo calcolo
	a.input.SetTextSelection(0, -1)
	a.input.SetFocus()
}

func (a *app) copia() {
	walk.Clipboard().SetText(a.risultato.Text())
}

func (a *app) mostraRisultato(minuti int) {
	a.risultato.SetText(ore.Formatta(minuti))
	a.orologio.SetText(ore.Orologio(minuti) + vuoto)
}

// azzera svuota il campo e riporta il risultato a 0:00.
func (a *app) azzera() {
	a.input.SetText("")
	a.mostraRisultato(0)
	a.input.SetFocus()
}

// riusa rimette nel campo il calcolo di una riga dello storico, es. "18:00 + 3:00".
func (a *app) riusa() {
	i := a.lista.CurrentIndex()
	if i < 0 || i >= len(a.storico) {
		return
	}
	espr, _, _ := strings.Cut(a.storico[i].testo, " = ")
	a.input.SetText(espr)
	a.input.SetFocus()
	a.input.SetTextSelection(len(espr), len(espr))
}

func (a *app) svuota() {
	if len(a.storico) == 0 {
		return
	}
	if walk.MsgBox(a.mw, "Svuota storico", "Cancellare tutto lo storico?",
		walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
		return
	}
	a.storico = nil
	a.aggiornaLista()
	a.salvaStorico()
	a.input.SetFocus()
}

func (a *app) aggiornaLista() {
	righe := make([]string, len(a.storico))
	for i, v := range a.storico {
		righe[i] = v.testo
	}
	a.lista.SetModel(righe)
}

func fileStorico() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "CalcolatriceOre", "storico.txt")
}

// Il file ha una riga per calcolo: "minuti<TAB>testo".
func (a *app) caricaStorico() {
	f, err := os.Open(fileStorico())
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		min, testo, ok := strings.Cut(sc.Text(), "\t")
		if !ok {
			continue
		}
		m, ok := leggiMinuti(min)
		if !ok {
			continue
		}
		a.storico = append(a.storico, voce{testo, m})
	}
}

func (a *app) salvaStorico() {
	p := fileStorico()
	if p == "" {
		return
	}
	os.MkdirAll(filepath.Dir(p), 0o755)
	var b strings.Builder
	for _, v := range a.storico {
		b.WriteString(ore.Formatta(v.minuti) + "\t" + v.testo + "\n")
	}
	os.WriteFile(p, []byte(b.String()), 0o644)
}

// leggiMinuti legge un valore "h:mm" (anche oltre le 24 ore o negativo) dal file dello storico.
func leggiMinuti(s string) (int, bool) {
	neg := strings.HasPrefix(s, "-")
	h, m, ok := strings.Cut(strings.TrimPrefix(s, "-"), ":")
	if !ok {
		return 0, false
	}
	hh, err1 := strconv.Atoi(h)
	mm, err2 := strconv.Atoi(m)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	v := hh*60 + mm
	if neg {
		v = -v
	}
	return v, true
}
