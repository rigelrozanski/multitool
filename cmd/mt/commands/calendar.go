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
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/rigelrozanski/common"
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

var (
	credentialsFileName = "calendar_credentials.json"
	tokenFileName       = "calendar_token.json"
)

func monthNumber(month string) string {
	lc := strings.ToLower(month)
	switch lc {
	case "jan":
		return "01"
	case "feb":
		return "02"
	case "mar":
		return "03"
	case "apr":
		return "04"
	case "may":
		return "05"
	case "jun":
		return "06"
	case "jul":
		return "07"
	case "aug":
		return "08"
	case "sep":
		return "09"
	case "oct":
		return "10"
	case "nov":
		return "11"
	case "dec":
		return "12"
	default:
		return "" // invalid
	}
}

// basically a validity check for the days. "01", "31" passes, " 1", "32" fails
func dayNumber(day string) string {
	if len(day) != 2 {
		return ""
	}
	_, err := strconv.Atoi(string(day[0]))
	if err != nil {
		return ""
	}
	_, err = strconv.Atoi(string(day[1]))
	if err != nil {
		return ""
	}
	conv3, err := strconv.Atoi(day)
	if err != nil {
		return ""
	}
	if conv3 < 31 {
		return day
	}
	return ""
}

// add an entry to the calendar
var AddCalEntryCmd = &cobra.Command{
	Use:  "add [source-file] [lineno]",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		// get the calendar file
		sourceFile := args[0]
		entryLineNo, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		if !common.FileExists(sourceFile) {
			return errors.New("file don't exist")
		}
		lines, err := common.ReadLines(sourceFile)
		if err != nil {
			return err
		}

		/////////////////////////////////////////////
		// construct the event, general form:
		//Mar 01 - Fri - LIP at cyberia

		name := lines[entryLineNo][15:]
		year := strconv.Itoa(time.Now().Year()) // use the current year
		month, day := "", ""

		// get the month and day
		for i := entryLineNo; i > 0; i-- {
			line := lines[i]
			if len(line) == 0 {
				return fmt.Errorf("empty line: %v", i)
			}
			monthI := monthNumber(line[0:3])
			dayI := dayNumber(line[4:6])
			if len(month) == 0 && len(monthI) == 2 {
				month = monthI
			}
			if len(day) == 0 && len(dayI) == 2 {
				day = dayI
			}
			if len(month) != 0 && len(day) != 0 {
				break
			}
		}

		// construct the date
		date := fmt.Sprintf("%v-%v-%v", year, month, day)

		///////////////////////////////////////////////////////////
		// Get credentials and send the event to google calendar
		credentialsFile, err := getFileInThisPath(credentialsFileName)
		if err != nil {
			return err
		}

		b, err := ioutil.ReadFile(credentialsFile)
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)

		srv, err := calendar.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Calendar client: %v", err)
		}

		event := &calendar.Event{
			Summary:  name,
			Location: "",
			Start: &calendar.EventDateTime{
				Date:     date,
				TimeZone: "America/Toronto",
			},
			End: &calendar.EventDateTime{
				Date:     date,
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
	Use: "remove [source-file] [lineno]",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
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
	tokFile, _ := getFileInThisPath(tokenFileName) // ignore error here, the next line will catch it
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
