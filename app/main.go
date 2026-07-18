package main

import (
	//"bufio"
	"fmt"
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

		if tokens[len(tokens)-1] == "&" {
			bg_job_cmd := tokens
			tokens = tokens[:len(tokens)-1]
			prog := exec.Command(tokens[0], tokens[1:]...)
			prog.Stdout = out
			prog.Stderr = outErr
			prog.Start()
			if err != nil {
			}
			job_no++
			bg_job_no_and_cmd[job_no] = strings.Join(bg_job_cmd, " ")
			job_pid := strconv.Itoa(prog.Process.Pid)
			fmt.Println("[" + strconv.Itoa(job_no) + "]" + " " + job_pid)

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
			switch tokens[0] {
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
						err = os.Chdir(tokens[1])
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
					return
				}
			case "echo":
				{
					fmt.Fprintln(out, strings.TrimPrefix(command, "echo "))
				}
			case "complete":
				{
					if len(tokens) > 1 {
						switch tokens[1] {
						case "-C":
							if len(tokens) > 3 {
								kayitlar[tokens[3]] = tokens[2]
							} else {
								continue
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
								continue
							}
						case "-r":
							if len(tokens) > 2 {
								delete(kayitlar, tokens[2])
							} else {
								continue
							}
						default:
							if len(tokens) > 2 {
								fmt.Fprintln(outErr, tokens[0]+": "+tokens[2]+": "+"no completion specification")
							} else {
								continue
							}
						}

					} else {
						continue
					}
				}
			case "jobs":
				{
					if job_no == 0 {
						fmt.Fprint(out, "$ ")
					} else {

						job_marker := []string{" ", "-", "+"}
						jm_no := 0
						for i := 1; i < job_no; i++ {
							switch i {
							case job_no:
								jm_no = 2
							case (job_no - 1):
								jm_no = 1
							default:
								jm_no = 0
							}
							fmt.Println("[" + strconv.Itoa(job_no) + "]" + job_marker[jm_no] + "  " + "Running                 " + bg_job_no_and_cmd[i])
						}
					}
				}
			default:
				{
					fmt.Fprintln(outErr, command+": command not found")
				}

			}
		}
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
