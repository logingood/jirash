package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/andygrunwald/go-jira"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("jirashell", "A command-line jira tool")

	weekly     = app.Command("weekly", "Weekly report")
	allTickets = app.Command("all", "All tickets")
	todoLater  = app.Command("todo", "To Do later tickets")
	backLog    = app.Command("backlog", "Backlogs")
	inProgress = app.Command("inprogress", "In progress tickets")

	ticket     = app.Command("create", "Create a ticket")
	ticketSum  = ticket.Arg("summary", "Summary of the ticket").Required().String()
	ticketDesc = ticket.Arg("desc", "Decription of the ticket").Required().String()

	ticketClose = app.Command("close", "Mark ticket as done")
	ticketNum   = ticketClose.Arg("ticketnum", "Number of ticket to mark as done").Required().String()
)

type Creds struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	ProjectId string `json:"projectid"`
	IssueType string `json:"issuetype"`
	Endpoint  string `json:"endpoint"`
}

func readCreds() (login string, pass string, projectid string, issuetype, endpoint string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter endpoint name, e.g. https://company.atlassian.net: ")
	epoint, _ := reader.ReadString('\n')

	fmt.Print("Enter Project ID (Operations - 11000): ")
	project, _ := reader.ReadString('\n')
	fmt.Print("Enter Issue Type (6 for OPS): ")
	issue, _ := reader.ReadString('\n')

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		panic("Can not read password")
	}

	login = strings.TrimSpace(username)
	pass = strings.TrimSpace(string(bytePassword))
	projectid = strings.TrimSpace(project)
	issuetype = strings.TrimSpace(issue)
	endpoint = strings.TrimSpace(epoint)

	return login, pass, projectid, issuetype, endpoint
}

func handle_error(err error) {
	if err != nil {
		output := fmt.Sprintf("Can't write credentials: %v", err)
		panic(output)
	}
}

func writeCreds() (login, pass, projectid, issuetype, endpoint string) {
	login, pass, projectid, issuetype, endpoint = readCreds()
	creds := Creds{login, pass, projectid, issuetype, endpoint}
	b, err := json.Marshal(creds)
	handle_error(err)

	usr, err := user.Current()
	handle_error(err)

	config := string(usr.HomeDir) + "/.jirashell.json"
	err = ioutil.WriteFile(config, b, 0644)
	handle_error(err)

	return login, pass, projectid, issuetype, endpoint
}

func getCreds() (c Creds) {
	usr, err := user.Current()
	config := string(usr.HomeDir) + "/.jirashell.json"
	raw, err := ioutil.ReadFile(config)
	handle_error(err)
	json.Unmarshal(raw, &c)
	return c
}

func jiraAuth() (jiraClient *jira.Client, login, projectid, issuetype, endpoint string) {
	usr, err := user.Current()
	config := string(usr.HomeDir) + "/.jirashell.json"
	var pass string
	if _, err := os.Stat(config); os.IsNotExist(err) {
		login, pass, projectid, issuetype, endpoint = writeCreds()
	} else {
		lpass := getCreds()
		login, pass, projectid, issuetype, endpoint = lpass.Login, lpass.Password, lpass.ProjectId, lpass.IssueType, lpass.Endpoint
	}

	jiraClient, err = jira.NewClient(nil, endpoint)
	handle_error(err)

	res, err := jiraClient.Authentication.AcquireSessionCookie(login, pass)
	if err != nil || res == false {
		panic(err)
	}
	return jiraClient, login, projectid, issuetype, endpoint
}

func printTicketOutput(issue *jira.Issue, endpoint string) {
	fmt.Printf("We created succesfully issue: %s\n", issue.Key)
	issue_url := fmt.Sprintf("%s/browse/%s", endpoint, issue.Key)
	if runtime.GOOS == "darwin" {
		app := "/usr/bin/open"
		cmd := exec.Command(app, issue_url)
		_, err := cmd.Output()
		handle_error(err)
	}
}

func createTicket(jiraClient *jira.Client, login, desc, summary, issuetype, project, endpoint string) string {
	i := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				Name: login,
			},
			Description: desc,
			Type: jira.IssueType{
				ID: issuetype,
			},
			Project: jira.Project{
				ID: project,
			},
			Summary: summary,
		},
	}
	issue, _, err := jiraClient.Issue.Create(&i)
	handle_error(err)
	printTicketOutput(issue, endpoint)
	return issue.Key
}

func closeTicket(jiraClient *jira.Client, login, ticketnum, issuetype, project, endpoint string) {
	tr, _, _ := jiraClient.Issue.GetTransitions("OP-25793")
	issue, err := jiraClient.Issue.DoTransition("OP-25793", "{\"update\": { \"comment\": [ { \"add\": { \"body\": \"test33333\" } } ] }, \"fields\": { \"assignee\": { \"name\": \"mmukhtarov\" }, \"resolution\": { \"name\": \"Done\" }}, \"transition\": { \"id\": \"821\" } }")
	fmt.Printf("fmt = %+v", tr)
	fmt.Printf("err = %+v", err)
	fmt.Printf("issue = %+v", issue)
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
	jiraClient, login, projectid, issuetype, endpoint := jiraAuth()
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case weekly.FullCommand():
		search_string := fmt.Sprintf("(assignee = %s) AND updatedDate > startOfWeek() ORDER BY updatedDate ASC", login)
		jiraSearch(jiraClient, search_string)
	case allTickets.FullCommand():
		search_string := fmt.Sprintf("(assignee = %s) AND (status = Open OR status = Reopened OR status = 'In Progress' OR status = 'TO DO LATER' OR status = 'Backlog')", login)
		jiraSearch(jiraClient, search_string)
	case todoLater.FullCommand():
		search_string := fmt.Sprintf("(assignee = %s) AND status = 'TO DO LATER'", login)
		jiraSearch(jiraClient, search_string)
	case backLog.FullCommand():
		search_string := fmt.Sprintf("(assignee = %s) AND status = 'Backlog'", login)
		jiraSearch(jiraClient, search_string)
	case inProgress.FullCommand():
		search_string := fmt.Sprintf("(assignee = %s) AND status = 'In Progress'", login)
		jiraSearch(jiraClient, search_string)
	case ticket.FullCommand():
		createTicket(jiraClient, login, *ticketDesc, *ticketSum, issuetype, projectid, endpoint)
	case ticketClose.FullCommand():
		closeTicket(jiraClient, login, *ticketNum, issuetype, projectid, endpoint)
	}
}
