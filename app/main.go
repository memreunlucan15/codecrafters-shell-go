package main

import (
	//"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {

	//rl, err := readline.New("$ ")
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: &benimCompleter{},
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	// TODO: Uncomment the code below to pass the first stage

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
		} else if tokens[0] == "type" {

			switch tokens[1] { // type sonrası builtin komut kontrolü
			case "exit", "echo", "type", "pwd", "cd":
				fmt.Fprintln(out, tokens[1]+" is a shell builtin")
			default:

				path, err1 := exec.LookPath(tokens[1])
				if err1 == nil {
					fmt.Fprintln(out, tokens[1]+" is "+path)
				} else {
					fmt.Fprintln(outErr, tokens[1]+": not found")
				}
			}

		} else if tokens[0] == "pwd" { // pwd ile ablosute path alma

			abs_path, _ := os.Getwd()
			fmt.Fprintln(out, abs_path)

		} else if tokens[0] == "cd" { // cd ile directory değişimi

			if tokens[1] != "~" {
				err = os.Chdir(tokens[1])
				if err != nil {
					fmt.Fprintln(out, "cd: "+tokens[1]+": No such file or directory")
				}
			} else {
				home_dir, _ := os.UserHomeDir()
				_ = os.Chdir(home_dir)
			}

		} else if tokens[0] == "exit" {
			break
		} else if tokens[0] == "echo" {
			fmt.Fprintln(out, strings.TrimPrefix(command, "echo "))
		} else {
			fmt.Fprintln(outErr, command+": command not found")
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

		if c == ">" || c == "1>" {
			durum = 1
			return durum
		} else if c == "2>" {
			durum = 2
			return durum
		} else if c == ">>" || c == "1>>" {
			durum = 3
			return durum
		} else if c == "2>>" {
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

	if prefix != b.oncekiPrefix {
		b.tabSayisi = 0 // yeni yazı → sıfırla
		b.oncekiPrefix = prefix
	}

	b.tabSayisi++

	gorulen := map[string]bool{}

	for i := 0; i < len(bizimbuiltinler); i++ {
		if !gorulen[bizimbuiltinler[i]] {
			builtinler = append(builtinler, bizimbuiltinler[i])
			gorulen[bizimbuiltinler[i]] = true
		}
	}

	for i := 0; i < len(klasorler); i++ {
		girdi, _ := os.ReadDir(klasorler[i])
		for j := 0; j < len(girdi); j++ {
			if !gorulen[girdi[j].Name()] {
				builtinler = append(builtinler, girdi[j].Name())
				gorulen[girdi[j].Name()] = true
			}
		}
	}
	sort.Strings(builtinler)
	for i := 0; i < len(builtinler); i++ {
		sonuc = strings.HasPrefix(builtinler[i], prefix) // havuzdaki adaylar prefix ile mi başlıyor
		var siraBuiltin = builtinler[i]
		if sonuc {
			sira = siraBuiltin[len(prefix):]          // adaydaki prefixten fazla olan karakterleri sira ya atadık
			sira = sira + " "                         // boşluk ekledik
			oneriler = append(oneriler, []rune(sira)) // öneriler listesine sira yı ekledik
			eslesenler = append(eslesenler, builtinler[i])
		}

	}

	if len(oneriler) == 0 {
		fmt.Print("\x07")
	} else if len(oneriler) == 1 {

		return oneriler, len(prefix)
	} else if len(oneriler) > 1 && b.tabSayisi == 1 {
		fmt.Print("\x07")
		if strings.HasSuffix(prefix, "_") {
			yazdir := oneriler[0]
			fmt.Print(yazdir)
			oneriler = nil
		} else {

			oneriler = nil
		}
	} else if len(oneriler) > 1 && b.tabSayisi == 2 {
		fmt.Print("\n")
		fmt.Print(strings.Join(eslesenler, "  "))
		fmt.Print("\n")
		fmt.Print("$ " + prefix)
		return nil, len(prefix)
	}

	return oneriler, len(prefix) // önerileri ve prefixin uzunluğunu geri döndürdük

}
