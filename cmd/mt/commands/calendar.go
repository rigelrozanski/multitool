package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/spf13/cobra"
)

// root command for calendar commands
var CalCmd = &cobra.Command{
	Use:   "cal",
	Short: "vim calandar managment",
}

func init() {
	CalCmd.AddCommand(AddCalEntryCmd)
	CalCmd.AddCommand(RemoveCalEntryCmd)
	RootCmd.AddCommand(CalCmd)
}

// add an entry to the calendar
var AddCalEntryCmd = &cobra.Command{
	Use: "add [month] [day] [name]",
	RunE: func(cmd *cobra.Command, args []string) error {

		credentialsFile, err := getFileInThisPath("calendar_credentials.json")
		if err != nil {
			return err
		}

		b, err := ioutil.ReadFile(credentialsFile)
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)

		srv, err := calendar.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Calendar client: %v", err)
		}

		event := &calendar.Event{
			Summary:  "Rige Test2",
			Location: "",
			Start: &calendar.EventDateTime{
				Date:     "2019-03-11",
				TimeZone: "America/Toronto",
			},
			End: &calendar.EventDateTime{
				Date:     "2019-03-11",
				TimeZone: "America/Toronto",
			},
		}

		calendarID := "primary"
		event, err = srv.Events.Insert(calendarID, event).Do()
		if err != nil {
			log.Fatalf("Unable to create event. %v\n", err)
		}
		fmt.Printf("Event created: %s\n", event.HtmlLink)

		return nil
	},
}

// remove an entry to the calendar
var RemoveCalEntryCmd = &cobra.Command{
	Use: "remove [month] [day] [name]",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

//__________________________________________________________________________

// getFileInThisPath - get a file in the path of THIS golang file
func getFileInThisPath(filename string) (string, error) {
	// get the config file
	_, thisfilename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("No caller information")
	}
	filepath := path.Join(path.Dir(thisfilename), filename)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return filepath, fmt.Errorf("expected file at:\n\t%v", filepath)
	}
	return filepath, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile, _ := getFileInThisPath("calendar_token.json") // ignore error here
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
