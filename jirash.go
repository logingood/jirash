package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/andygrunwald/go-jira"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("jirashell", "A command-line jira tool")
	jiraEnd = app.Flag("server", "End-point address.").Default("zendesk.atlassian.net").String()

	weekly     = app.Command("weekly", "Weekly report")
	allTickets = app.Command("all", "All tickets")
	todoLater  = app.Command("todo", "To Do later tickets")
	backLog    = app.Command("backlog", "Backlogs")
	inProgress = app.Command("inprogress", "In progress tickets")

	ticket     = app.Command("create", "Create a ticket")
	ticketSum  = ticket.Arg("summary", "Summary of the ticket").Required().String()
	ticketDesc = ticket.Arg("desc", "Decription of the ticket").Required().String()
)

type Creds struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func readCreds() (login string, pass string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		panic("Can not read password")
	}
	login = strings.TrimSpace(username)
	pass = strings.TrimSpace(string(bytePassword))
	return login, pass
}

func handle_error(err error) {
	if err != nil {
		output := fmt.Sprintf("Can't write credentials: %v", err)
		panic(output)
	}
}

func writeCreds() (login string, pass string) {
	login, pass = readCreds()
	creds := Creds{login, pass}
	b, err := json.Marshal(creds)
	handle_error(err)

	usr, err := user.Current()
	handle_error(err)

	config := string(usr.HomeDir) + "/.jirashell.json"
	err = ioutil.WriteFile(config, b, 0644)
	handle_error(err)

	return login, pass
}

func getCreds() (c Creds) {
	usr, err := user.Current()
	config := string(usr.HomeDir) + "/.jirashell.json"
	raw, err := ioutil.ReadFile(config)
	handle_error(err)
	json.Unmarshal(raw, &c)
	return c
}

func jiraAuth() (jiraClient *jira.Client, login string) {
	usr, err := user.Current()
	config := string(usr.HomeDir) + "/.jirashell.json"
	var pass string
	if _, err := os.Stat(config); os.IsNotExist(err) {
		login, pass = writeCreds()
	} else {
		lpass := getCreds()
		login, pass = lpass.Login, lpass.Password
	}

	jiraClient, err = jira.NewClient(nil, "https://zendesk.atlassian.net")
	handle_error(err)

	res, err := jiraClient.Authentication.AcquireSessionCookie(login, pass)
	if err != nil || res == false {
		panic(err)
	}
	return jiraClient, login
}

func createTicket(jiraClient *jira.Client, login string, desc string, summary string) string {
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				Name: login,
			},
			Description: desc,
			Type: jira.IssueType{
				ID: "6",
			},
			Project: jira.Project{
				ID: "11000",
			},
			Summary: summary,
		},
	}
	issue, _, err := jiraClient.Issue.Create(&i)
	fmt.Printf("issue = %+v\n", issue)
	handle_error(err)
	return issue.Key
}

func jiraSearch(jiraClient *jira.Client, jql string) {
	search_opts := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
	}

	issues, _, _ := jiraClient.Issue.Search(jql, search_opts)
	//"(assignee = mmukhtarov) AND updatedDate > startOfWeek() ORDER BY updatedDate ASC", search_opts)
	for _, issue := range issues {
		fmt.Printf("%s: %+v\n", issue.Key, issue.Fields.Summary)
	}
}

func main() {
	jiraClient, login := jiraAuth()
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case weekly.FullCommand():
		search_string := "(assignee = mmukhtarov) AND updatedDate > startOfWeek() ORDER BY updatedDate ASC"
		jiraSearch(jiraClient, search_string)
	case allTickets.FullCommand():
		search_string := "(assignee = mmukhtarov) AND (status = Open OR status = Reopened OR status = 'In Progress' OR status = 'TO DO LATER' OR status = 'Backlog')"
		jiraSearch(jiraClient, search_string)
	case todoLater.FullCommand():
		search_string := "(assignee = mmukhtarov) AND status = 'TO DO LATER'"
		jiraSearch(jiraClient, search_string)
	case backLog.FullCommand():
		search_string := "(assignee = mmukhtarov) AND status = 'Backlog'"
		jiraSearch(jiraClient, search_string)
	case inProgress.FullCommand():
		search_string := "(assignee = mmukhtarov) AND status = 'In Progress'"
		jiraSearch(jiraClient, search_string)
	case ticket.FullCommand():
		createTicket(jiraClient, login, *ticketDesc, *ticketSum)
	}
}
