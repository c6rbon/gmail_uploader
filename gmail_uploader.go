package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sam-falvo/mbox"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var no_upload = flag.Bool("n", false, "Do not actually upload. Print messages instead.")
var only_msgno = flag.String("only_msgno", "", "Comma separated list of message number to constrain uploads to.")

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-uploader.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	flag.Parse()
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailInsertScope, gmail.GmailModifyScope, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	if len(flag.Args()) != 1 {
		log.Fatalf("Arg 1 should be an mbox")
	}

	fn := flag.Arg(0)

	msgno := make(map[int]int)

	if *only_msgno != "" {
		for _, n := range strings.Split(*only_msgno, ",") {
			i, err := strconv.Atoi(n)
			if err != nil {
				log.Fatalf("Unable to parse message numbers: %v", err)
			}
			msgno[i] = 1
		}
	}

	// Open mailbox
	f, err := os.Open(fn)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()

	{
		limit_s := ""
		if *only_msgno != "" {
			limit_s = fmt.Sprintf("(%d to import: %s)", len(msgno), *only_msgno)
		}
		fmt.Printf("%s Starting import %s\n", fn, limit_s)
	}

	ms, err := mbox.CreateMboxStream(f)
	if err != nil {
		log.Fatalf("Could not parse MBOX: %v", err)
	}

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"

	cnt := 0
	upld := 0

	buffer := make([]byte, 1024)

	for {
		msg, err := ms.ReadMessage()
		if err != nil {
			if err == io.EOF {
				// We're done.
				break
			}
			log.Printf("%s:%d Error parsing message: %v", fn, cnt, err)
		}

		var email bytes.Buffer
		email.WriteString(fmt.Sprintf("From %s\n", msg.Sender()))
		for _, hdr := range msg.AllHeaders() {
			email.WriteString(hdr)
			email.WriteString("\n")
		}

		email.WriteString("\n")
		bodyReader := msg.BodyReader()
		for err == nil {
			n, err := bodyReader.Read(buffer)
			if err != nil {
				break
			}
			email.Write(buffer[0:n])
		}

		if err == io.EOF {
			// lines now contains the collected body of the most recently
			// read message.
		}

		encoded := base64.URLEncoding.EncodeToString(email.Bytes())
		// Upload the message

		// Horrible hack
		mid := msg.Headers()["Message-ID"]
		if mid == nil {
			mid = msg.Headers()["Message-Id"]
		}
		if mid == nil {
			mid = msg.Headers()["Message-id"]
		}

		if *no_upload != true {
			if *only_msgno != "" && msgno[cnt] == 1 {
				fmt.Printf("%s:%d Uploading %d bytes\n", fn, cnt, len(encoded))

				r, err := srv.Users.Messages.Import(user, &gmail.Message{Raw: encoded}).Do()
				if err != nil {
					fmt.Printf("%s:%d FAILED to import message ID %s: %v\n",
						fn, cnt, mid, err)
				} else {
					fmt.Printf("%s:%d ID %s successful %s\n", fn, cnt, r.Id, mid)
					upld++
				}
			}
		} else {
			fmt.Printf(email.String())
			fmt.Println("-----------------------------------------------")
			fmt.Println(encoded)
			fmt.Println("-----------------------------------------------")
		}
		cnt++
	}
	fmt.Printf("%s %d messages uploaded out of %d processed.\n", fn, upld, cnt)

}
