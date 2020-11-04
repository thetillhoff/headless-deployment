package main

// used go version: 1.14

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

var fileMutex sync.Mutex

func getEnv(name string) string {
	return os.Getenv(name)
}

// error checking is often needed, so simplified with
//   check(err)
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Load file and return it's content.
func loadFile(filepath string) string {
	fileContent, err := ioutil.ReadFile(filepath) // read contents from path into variable
	check(err)
	return string(fileContent)
}

// Open file and overwrite it's content.
func writeFile(filepath string, content string) {
	// write the whole body at once
	err := ioutil.WriteFile(filepath, []byte(content), 0644)
	check(err)
}

func readKeyValue(filepath string, key string) string {
	content := loadFile(filepath)         // load file contents
	lines := strings.Split(content, "\n") // split file contents into lines
	linenumber := -1                      // initialize search parameter
	for index, line := range lines {      // iterate over lines
		if strings.Split(line, " ")[0] == key { // seperate first word and compare it against mac // if equal
			linenumber = index // set search result
			break              // stop searching
		}
	}
	return lines[linenumber]
}

func writeKeyValue(filepath string, key string, value string) {
	content := loadFile(filepath)         // load file contents
	lines := strings.Split(content, "\n") // split file contents into lines
	linenumber := -1                      // initialize search parameter
	for index, line := range lines {      // iterate over lines
		if strings.Split(line, " ")[0] == key { // seperate first word and compare it against mac // if equal
			linenumber = index // set search result
			break              // stop searching
		}
	}

	line := key
	if value != "" {
		line = strings.Join([]string{key, value}, " ")
	}
	if linenumber == -1 { // doesn't exist yet -> append
		lines = append(lines, line)
	} else { // does exist -> replace
		lines[linenumber] = line
	}

	content = strings.Join(lines, "\n")      // merge lines into one string
	if !(strings.HasSuffix(content, "\n")) { // if content doesn't end with newline
		content = content + "\n" // append newline
	}
	writeFile(filepath, content) // write content back to file
}

// Extract IP out of http.Request and return it.
func getUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	u, err := url.Parse("//" + IPAddress)
	check(err)

	return u.Hostname()
}

// GETsingle checks if request contains get parameter exactly once and return it (and true if success | false for error)
func GETsingle(r *http.Request, get string) (string, bool) {
	gets, ok := r.URL.Query()["mac"] // could be multiple 'mac' parameters, thus error checking follows
	if !ok || len(gets) == 0 {       // if parameter 'mac' was not found
		return "No " + get + " provided.", false
	} else if ok && len(gets) > 1 { // if more than one parameter 'mac' was found
		return "More than one " + get + " provided.", false
	}

	// exactly one mac was provided
	return gets[0], true // extract the single mac
}

// validate 'mac' for being a valid MAC
func validMAC(mac string) bool {
	re := regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$")
	return re.Match([]byte(mac))
}

// generatePassword generates a plain password with specified length
func generatePassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	plain := ""
	for i := 1; i < length; i++ {
		plain = plain + string(letters[rand.Intn(len(letters))])
	}
	return plain
}

// generateSecurePassword generates a salted & encrypted password out of the specified plain password
func generateSecurePassword(plain string) string {
	out, err := exec.Command("mkpasswd -m sha-512 -S $(pwgen -ns 16 1) " + plain).Output()
	check(err)
	return string(out)
}

// Handler for '<SERVERADDR>/default'
func defaultHandler(w http.ResponseWriter, r *http.Request) {

	mac, ok := GETsingle(r, "mac")
	if !ok {
		fmt.Fprintf(w, mac)
	} else if !validMAC(mac) {
		fmt.Fprintf(w, "Provided MAC is invalid.")
	} else { // mac was provided correctly
		t, err := template.ParseFiles("provisioning_debian" + ".tmpl")
		check(err)
		data := make(map[string]string)

		// data
		data["mac"] = mac       // set mac for further communication
		data["server"] = r.Host // set own hostname for further data requests
		data["hostname"] = mac  // name machines after their mac

		// execute template
		err = t.Execute(w, data)
		check(err)

		// save data to file
		fileMutex.Lock()
		writeKeyValue("/hosts", data["mac"], "") // either append to or replace in file (mac is key)
		fileMutex.Unlock()
	}
}

// Handler for '<SERVERADDR>/preseed'
func preseedHandler(w http.ResponseWriter, r *http.Request) {

	mac, ok := GETsingle(r, "mac")
	if !ok {
		fmt.Fprintf(w, mac)
	} else if !validMAC(mac) {
		fmt.Fprintf(w, "Provided MAC is invalid.")
	} else { // mac was provided correctly

		t, err := template.ParseFiles("preseed" + ".tmpl")
		check(err)
		data := make(map[string]string)

		// password generation & encryption
		plainPassword := generatePassword(16)                   // generate plain password
		securePassword := generateSecurePassword(plainPassword) // generate salted & encrypted password

		// data
		data["mac"] = mac                     // set mac for further communication
		data["server"] = r.Host               // set own hostname for further data requests
		data["hostname"] = mac                // name machines after their mac
		data["username"] = getEnv("username") // set username
		data["passcrypt"] = securePassword    // set secure password string

		// execute template
		err = t.Execute(w, data)
		check(err)

		// save data to file
		fileMutex.Lock()
		writeKeyValue("/hosts", data["mac"], data["passcrypt"]) // either append to or replace in file (mac is key)
		fileMutex.Unlock()
	}
}

// Handler for '<SERVERADDR>/preseedlate'
func preseedlateHandler(w http.ResponseWriter, r *http.Request) {

	mac, ok := GETsingle(r, "mac")
	if !ok {
		fmt.Fprintf(w, mac)
	} else if !validMAC(mac) {
		fmt.Fprintf(w, "Provided MAC is invalid.")
	} else { // mac was provided correctly
		t, err := template.ParseFiles("preseedlate" + ".tmpl")
		check(err)
		data := make(map[string]string)

		fileMutex.Lock()

		plainPassword := readKeyValue("/hosts", mac)

		// data
		data["server"] = r.Host // get own hostname
		data["mac"] = mac

		// execute template
		err = t.Execute(w, data)
		check(err)

		// save data to file
		writeKeyValue("/hosts", data["mac"], plainPassword+" "+getUserIP(r)) // either append to or replace in file (mac is key)
		fileMutex.Unlock()
	}
}

func main() {

	fmt.Print("To access this webserver access localhost:8080\n")
	http.HandleFunc("/default", defaultHandler)
	http.HandleFunc("/preseed", preseedHandler)
	http.HandleFunc("/preseedlate", preseedlateHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
