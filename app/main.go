package main

import (
	//"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

var _ = fmt.Print

var kayitlar = map[string]string{}
var job_no = 0
var bg_job_no_and_cmd = map[int]string{}

func main() {

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: &benimCompleter{},
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		command, err := rl.Readline()
		command = strings.TrimSpace(command)
		tokens := tokenci(command) // Tokenlere ayırma
		redir := isRedir(tokens)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)

		}

		var out = os.Stdout
		var outErr = os.Stderr

		cmdpieces, pipeok := isPipeline(tokens)
		blttablo := make([]bool, len(cmdpieces))
		mevcut_program := make([]*exec.Cmd, len(cmdpieces))

		if pipeok {
			for i := 0; i < len(cmdpieces); i++ {
				_, blttablo[i] = isBuiltin(cmdpieces[i])
			}

			//var mevcut_program []*exec.Cmd
			var oncekiCikti io.Reader

			for i := 0; i < len(cmdpieces); i++ {
				mevcut_program[i] = exec.Command(cmdpieces[i][0], cmdpieces[i][1:]...)
				mevcut_program[i].Stderr = outErr
				mevcut_program[i].Stdin = oncekiCikti
				boru, _ := mevcut_program[i].StdoutPipe()
				oncekiCikti = boru
				if len(cmdpieces)-1 == i {
					mevcut_program[i].Stdout = out
				}
			}

			for i := 0; i < len(cmdpieces); i++ {
				mevcut_program[i].Start()
			}
			for i := len(cmdpieces) - 1; i == 0; i-- {
				mevcut_program[i].Wait()
			}
			continue
		}

		if p := isPipe(tokens); p >= 0 {
			first_piece := tokens[:p]
			sec_piece := tokens[p+1:]

			buffer := &bytes.Buffer{}

			_, fp := isBuiltin(first_piece)
			_, sp := isBuiltin(sec_piece)

			if !fp && !sp { // DIŞ - DIŞ
				prog1 := exec.Command(first_piece[0], first_piece[1:]...)
				prog2 := exec.Command(sec_piece[0], sec_piece[1:]...)

				prog1.Stderr = outErr
				prog2.Stdout = out
				prog2.Stderr = outErr

				boru, _ := prog1.StdoutPipe()
				prog2.Stdin = boru // boruyu bağladık

				prog1.Start()
				prog2.Start()

				prog2.Wait()
				prog1.Wait()
			} else if fp && !sp { // BUILT - DIŞ
				runBuiltin(first_piece, buffer, outErr)
				prog2 := exec.Command(sec_piece[0], sec_piece[1:]...)

				prog2.Stdout = out
				prog2.Stderr = outErr

				prog2.Stdin = strings.NewReader(buffer.String())

				prog2.Start()

				prog2.Wait()
			} else if !fp && sp { // DIŞ - BUILT
				prog1 := exec.Command(first_piece[0], first_piece[1:]...)

				prog1.Stdout = io.Discard
				prog1.Stderr = outErr

				prog1.Start()
				runBuiltin(sec_piece, out, outErr)

				prog1.Wait()
			} else if fp && sp { // BUILT - BUILT
				runBuiltin(first_piece, buffer, outErr)
				runBuiltin(sec_piece, out, outErr)

			}
			continue
		}

		if tokens[len(tokens)-1] == "&" {
			bg_job_cmd := tokens
			tokens = tokens[:len(tokens)-1]
			prog := exec.Command(tokens[0], tokens[1:]...)
			prog.Stdout = out
			prog.Stderr = outErr
			prog.Start()
			job_no++

			closest_available := 0
			for i := 1; i <= job_no; i++ {
				v := bg_job_no_and_cmd[i]
				if v == "" {
					closest_available = i
					break
				}
			}

			bg_job_cmd = append(bg_job_cmd, "Running")
			bg_job_no_and_cmd[closest_available] = strings.Join(bg_job_cmd, " ")
			this_process := closest_available
			go func() {
				_ = prog.Wait()
				bg_job_no_and_cmd[this_process] = strings.TrimSuffix(bg_job_no_and_cmd[this_process], "Running") + "Done"
			}()

			job_pid := strconv.Itoa(prog.Process.Pid)
			fmt.Println("[" + strconv.Itoa(closest_available) + "]" + " " + job_pid)
			process_check()
			continue
		}

		switch redir {
		case 1:
			f, _ := os.Create(tokens[len(tokens)-1])
			out = f
			tokens = tokens[:len(tokens)-2]

		case 2:
			e, _ := os.Create(tokens[len(tokens)-1])
			outErr = e
			tokens = tokens[:len(tokens)-2]
		case 3:
			fa, _ := os.OpenFile(tokens[len(tokens)-1], os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
			out = fa
			tokens = tokens[:len(tokens)-2]
		case 4:
			ea, _ := os.OpenFile(tokens[len(tokens)-1], os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
			outErr = ea
			tokens = tokens[:len(tokens)-2]
		}

		if _, err := exec.LookPath(tokens[0]); err == nil { // Path kontrolü

			var prog = exec.Command(tokens[0], tokens[1:]...)
			prog.Stdout = out
			prog.Stderr = outErr
			prog.Run()
		} else {
			ran, quit := runBuiltin(tokens, out, outErr)
			if quit {
				break
			}
			if !ran {
				fmt.Fprintln(outErr, command+": command not found")
			}
		}
		process_check()
	}
}

func tokenci(line string) []string {
	var sonuc []string
	var inQuotes bool
	var inDQuotes bool
	var bslash bool
	current := ""
	for i := 0; i < len(line); i++ {
		c := line[i]

		if bslash && !inQuotes {
			current += string(c)
			bslash = false
		} else if c == '\\' && !inQuotes {
			bslash = true
		} else if c == '"' && !inQuotes {
			if inDQuotes == false {
				inDQuotes = true
			} else {
				inDQuotes = false
			}
		} else if c == '\'' && inDQuotes == false {
			if inQuotes == false {
				inQuotes = true
			} else {
				inQuotes = false
			}
		} else if c == ' ' && inQuotes == false && inDQuotes == false {
			if current != "" {
				sonuc = append(sonuc, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	sonuc = append(sonuc, current)
	return sonuc
}

func isPipeline(tokenized []string) (commands [][]string, is bool) {
	// var cmdcount = 0
	var cmands [][]string
	var tokens []string
	var pipeMi bool
	tokens = tokenized

	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "|" {
			pipeMi = true
			cmands = append(cmands, tokens[:i])
			tokens = tokens[i+1:]

		}

	}
	if pipeMi {
		cmands = append(cmands, tokens[0:])
	}

	return cmands, pipeMi
}

// "|" var mı, varsa oradan bölüp tokenler yapma fonksiyonu

// çıktı uzunluğuna göre pipe sayısı belirlenir
// fonksiyon çıktısını alıp tokenlerin blt durumuna bakılır
// blt durumuna göre pipelar bağlanır
// komutlar argümanlarıyla birlikte blt durumuna uygun şekilde çalıştırılır

func isBuiltin(tokenized []string) (blt string, durum bool) {

	durum = true
	switch tokenized[0] {

	case "type":
		blt = "type"
	case "pwd":
		blt = "pwd"
	case "cd":
		blt = "cd"
	case "exit":
		blt = "exit"
	case "echo":
		blt = "echo"
	case "complete":
		blt = "complete"
	case "jobs":
		blt = "jobs"
	default:
		blt = ""
		durum = false
	}

	return blt, durum
}

// builtin durumu kontrol fonkisyonu eklenecek
// o zaman pipeline kısmında fp vs sp if kontrollerinin öncesinde runBuiltin çalıştırmaya gerek kalmaz
// bu mantık kurulduktan sonra pipeline kurma ve builtin-dış komut mekanizmaları n tane pipe durumuna göre uyarlanacak.

func isRedir(tokenized []string) int {
	var durum int
	for i := 0; i < len(tokenized); i++ {
		c := tokenized[i]

		switch c {
		case ">", "1>":
			durum = 1
			return durum
		case "2>":
			durum = 2
			return durum
		case ">>", "1>>":
			durum = 3
			return durum
		case "2>>":
			durum = 4
			return durum
		}
	}
	return durum
}

func isPipe(tokenized []string) int {
	var durum int
	for s, t := range tokenized {
		if t == "|" {
			durum = s
			return durum
		} else {
			durum = -1
		}
	}
	return durum
}

func runBuiltin(tokens []string, out, outErr io.Writer) (ran bool, quit bool) {

	ran = true
	typeblt, isblt := isBuiltin(tokens)
	if isblt {
		switch typeblt {
		case "type":
			{

				switch tokens[1] { // type sonrası builtin komut kontrolü
				case "exit", "echo", "type", "pwd", "cd", "complete", "jobs":
					fmt.Fprintln(out, tokens[1]+" is a shell builtin")
				default:

					path, err1 := exec.LookPath(tokens[1])
					if err1 == nil {
						fmt.Fprintln(out, tokens[1]+" is "+path)
					} else {
						fmt.Fprintln(outErr, tokens[1]+": not found")
					}
				}

			}
		case "pwd":
			{ // pwd ile ablosute path alma

				abs_path, _ := os.Getwd()
				fmt.Fprintln(out, abs_path)

			}
		case "cd":
			{ // cd ile directory değişimi

				if tokens[1] != "~" {
					err := os.Chdir(tokens[1])
					if err != nil {
						fmt.Fprintln(out, "cd: "+tokens[1]+": No such file or directory")
					}
				} else {
					home_dir, _ := os.UserHomeDir()
					_ = os.Chdir(home_dir)
				}

			}

		case "exit":
			{
				quit = true
			}
		case "echo":
			{
				fmt.Fprintln(out, strings.Join(tokens[1:], " "))
			}
		case "complete":
			{
				if len(tokens) > 1 {
					switch tokens[1] {
					case "-C":
						if len(tokens) > 3 {
							kayitlar[tokens[3]] = tokens[2]
						} else {
							process_check()
							//continue
						}
					case "-p":
						if len(tokens) > 2 {
							script, var_mi := kayitlar[tokens[2]]

							if !var_mi {
								fmt.Fprintln(outErr, tokens[0]+": "+tokens[2]+": "+"no completion specification")
							} else {
								fmt.Fprintln(out, tokens[0]+" "+"-C"+" "+"'"+script+"' "+tokens[2])
							}
						} else {
							process_check()
							//continue
						}
					case "-r":
						if len(tokens) > 2 {
							delete(kayitlar, tokens[2])
						} else {
							//continue
						}
					default:
						if len(tokens) > 2 {
							fmt.Fprintln(outErr, tokens[0]+": "+tokens[2]+": "+"no completion specification")
						} else {
							process_check()
							//continue
						}
					}

				} else {
					//continue
				}
			}
		case "jobs":
			{

				if len(bg_job_no_and_cmd) == 0 {
					process_check()
					//continue
					//fmt.Fprint(out, "$ ")
				} else {
					biggest := 0
					sec_biggest := 0
					for i := 1; i <= job_no; i++ {
						v := bg_job_no_and_cmd[i]
						if strings.HasSuffix(v, "Running") || strings.HasSuffix(v, "Done") {
							sec_biggest = biggest
							biggest = i
						}
					}
					for i := 1; i < (job_no + 1); i++ {
						job_marker := " "
						if i == biggest {
							job_marker = "+"
						} else if i == sec_biggest {
							job_marker = "-"
						}
						if strings.HasSuffix(bg_job_no_and_cmd[i], "Running") {
							fmt.Println("[" + strconv.Itoa(i) + "]" + job_marker + "  " + "Running                 " + strings.TrimSuffix(bg_job_no_and_cmd[i], " Running"))
						} else if strings.HasSuffix(bg_job_no_and_cmd[i], "Done") {
							fmt.Println("[" + strconv.Itoa(i) + "]" + job_marker + "  " + "Done                 " + strings.TrimSuffix(bg_job_no_and_cmd[i], " & Done"))
							bg_job_no_and_cmd[i] = bg_job_no_and_cmd[i] + "-delete"
						} else {

						}
					}
					for i := 1; i < (len(bg_job_no_and_cmd) + 1); i++ {
						if strings.HasSuffix(bg_job_no_and_cmd[i], "-delete") {
							delete(bg_job_no_and_cmd, i)
						}
					}
				}
			}
		default:
			{
				ran = false
			}
		}
	} else {
		ran = false
	}
	return ran, quit
}

func process_check() {
	biggest := 0
	sec_biggest := 0
	for i := 1; i <= job_no; i++ {
		v := bg_job_no_and_cmd[i]
		if strings.HasSuffix(v, "Running") || strings.HasSuffix(v, "Done") {
			sec_biggest = biggest
			biggest = i
		}
	}
	for i := 1; i < (job_no + 1); i++ {
		job_marker := " "
		if i == biggest {
			job_marker = "+"
		} else if i == sec_biggest {
			job_marker = "-"
		}

		if strings.HasSuffix(bg_job_no_and_cmd[i], "Done") {
			fmt.Println("[" + strconv.Itoa(i) + "]" + job_marker + "  " + "Done                 " + strings.TrimSuffix(bg_job_no_and_cmd[i], " & Done"))
			bg_job_no_and_cmd[i] = bg_job_no_and_cmd[i] + "-delete"
		} else {

		}
	}
	for i := 1; i < (len(bg_job_no_and_cmd) + 1); i++ {
		if strings.HasSuffix(bg_job_no_and_cmd[i], "-delete") {
			delete(bg_job_no_and_cmd, i)
		}
	}
}

type benimCompleter struct {
	tabSayisi    int
	oncekiPrefix string
}

func (b *benimCompleter) Do(line []rune, pos int) ([][]rune, int) {

	prefix := string(line[:pos])                // burada prefix dediğimiz şey, terminalde şu anda yazılı olan komut
	var builtinler []string                     // boş autocomplete öneri havuzu
	bizimbuiltinler := []string{"echo", "exit"} // mevcut builtinlerimiz
	var oneriler [][]rune
	var sonuc bool
	var sira string
	var eslesenler []string
	klasorler := filepath.SplitList(os.Getenv("PATH"))

	var tokenprefix = tokenci(prefix)
	var bltmiwdmi bool           // builtin mi wd mi
	var completer_script_mc bool // cs'nin multiple candidate'e sahip mi
	var prog_output []byte       // cs outputu
	var prog_output_s []string

	var klasor string
	var kok string

	fullprefix := prefix

	klasor = "."
	if len(tokenprefix) > 1 {
		if tokenprefix[0] != "" {

			if strings.Contains(prefix, "/") {
				kelime := tokenprefix[len(tokenprefix)-1]
				i := strings.LastIndex(kelime, "/")
				if i > (-1) {
					klasor = kelime[:i]
					kok = kelime[(i + 1):]

					tokenprefix[len(tokenprefix)-1] = kok
				}
			}
			prefix = strings.TrimSpace(tokenprefix[len(tokenprefix)-1])
			bltmiwdmi = true
		}
	}

	if prefix != b.oncekiPrefix {
		b.tabSayisi = 0 // yeni yazı → sıfırla
		b.oncekiPrefix = prefix
	}

	b.tabSayisi++

	script, varMi := kayitlar[tokenprefix[0]] // mapteki değere tekrar tekrar bakmamak için değeri scripte atadık
	if varMi {
		if path, err := exec.LookPath(script); err == nil { // Path kontrolü
			var prog *exec.Cmd

			if len(tokenprefix) > 1 {
				argv := []string{
					tokenprefix[0],
					tokenprefix[len(tokenprefix)-1],
					tokenprefix[len(tokenprefix)-2],
				}

				prog = exec.Command(path, argv[0], argv[1], argv[2])
			} else {
				prog = exec.Command(path, tokenprefix[0], "", "")
			}
			prog.Env = append(prog.Environ(), "COMP_LINE="+fullprefix)
			prog.Env = append(prog.Environ(), "COMP_POINT="+strconv.Itoa(len(fullprefix)))
			prog_output, _ = prog.Output()
			prog_output_s = strings.Fields(string(prog_output)) // burada çıktıyı adam ediyoruz
			if len(prog_output_s) > 1 {
				completer_script_mc = true
			} else {

				cikti := strings.TrimSpace(string(prog_output))
				if cikti != "" {
					cikti = cikti[len(prefix):] + " "
					oneriler = append(oneriler, []rune(cikti))
					return oneriler, len(prefix)
				} else {
					fmt.Print("\x07")
					return nil, len(prefix)
				}
			}
		}
	}

	gorulen := map[string]bool{}
	gorulenDir := map[string]bool{}

	var adayhavuzu []string
	if bltmiwdmi {
		adayhavuzu = append(adayhavuzu, klasor)
	} else {
		adayhavuzu = klasorler
	}

	if !completer_script_mc {

		// aday havuzu oluşturma döngüsü
		for i := 0; i < len(adayhavuzu); i++ {
			girdi, _ := os.ReadDir(adayhavuzu[i])
			for j := 0; j < len(girdi); j++ {
				if !gorulen[girdi[j].Name()] {
					builtinler = append(builtinler, girdi[j].Name())
					if girdi[j].IsDir() {
						gorulenDir[girdi[j].Name()] = true
					}
					gorulen[girdi[j].Name()] = true
				}
			}
		}

		// builtinlere echo-exit ekleme karar noktası
		if !bltmiwdmi {
			for i := 0; i < len(bizimbuiltinler); i++ {
				if !gorulen[bizimbuiltinler[i]] {
					builtinler = append(builtinler, bizimbuiltinler[i])
					gorulen[bizimbuiltinler[i]] = true
				}
			}
		}

	} else {
		for i := 0; i < len(prog_output_s); i++ {
			builtinler = append(builtinler, prog_output_s[i])
		}

	}

	//sıralama
	sort.Strings(builtinler)

	//filtreleme döngüsü
	for i := 0; i < len(builtinler); i++ {
		sonuc = strings.HasPrefix(builtinler[i], prefix) // havuzdaki adaylar prefix ile mi başlıyor
		var siraBuiltin = builtinler[i]
		if sonuc {
			sira = siraBuiltin[len(prefix):] // adaydaki prefixten fazla olan karakterleri sira ya atadık
			if !gorulenDir[builtinler[i]] {
				sira = sira + " " // boşluk ekledik
			} else {
				sira = sira + "/"
				builtinler[i] = builtinler[i] + "/"
			}
			oneriler = append(oneriler, []rune(sira)) // öneriler listesine sira yı ekledik
			eslesenler = append(eslesenler, builtinler[i])
		} else if completer_script_mc {
			eslesenler = append(eslesenler, builtinler[i])
		}

	}
	var kuyruk []string
	for i := 0; i < len(eslesenler); i++ {
		kuyruk = append(kuyruk, strings.TrimPrefix(eslesenler[i], prefix))
	}

	lcp := ""
	if len(kuyruk) > 1 {
		lcp = strings.TrimSpace(kuyruk[0])
		for i := 0; i < len(kuyruk[1:]); i++ {

			for j := 0; j < len(kuyruk); j++ {
				for !strings.HasPrefix(strings.TrimSpace(kuyruk[j]), lcp) {
					lcp = lcp[:len(lcp)-1]
					if lcp == "" {
						break
					}
				}
			}
		}
	}

	if len(oneriler) == 0 {
		fmt.Print("\x07")
	} else if len(oneriler) == 1 {

		return oneriler, len(prefix)
	}

	if len(lcp) != 0 {

		oneriler = [][]rune{[]rune(lcp)}
	} else if len(oneriler) > 1 && b.tabSayisi == 1 {

		fmt.Print("\x07")
		oneriler = nil
	} else if len(oneriler) > 1 && b.tabSayisi == 2 {
		fmt.Print("\n")
		fmt.Print(strings.Join(eslesenler, "  "))
		fmt.Print("\n")
		fmt.Print("$ " + fullprefix)
		return nil, len(prefix)
	}

	return oneriler, len(prefix) // önerileri ve prefixin uzunluğunu geri döndürdük

}
