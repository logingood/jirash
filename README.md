# Jira Shell
## Overview

Allows to query Jira for tickets assigned to user:
- all
- todo 
- backlog
- create a weekly report

Operations with issues:
- Create
- Mark in progress
- Mark as resolved

Please note transition IDs are hardcoded specific for my case, you can change it in the code.

First run jirash will ask you to add username, password, projectId and issue type.

## Install

Use go get:
````
go get github.com/logingood/jirash
````
package will be installed into your $GOBIN path, e.g.: 
```
/usr/local/go/bin
```

## Configuration 

Configuration would be automatically generated on first run and stored at
```
~/.jirashell.json
```

Note: your password from Jira is stored in clear text. If it is not acceptable for you
don't use this script.

## Example of usage:

Get all tickets for a user
```
jirash all
```

Create a new issue
```
jirash create "Test issue" "Test issue description"
```

Mark issue as in progress
```
jirash inprogress "ISSUE-123"
```

Resolve issue
```
jirash close "ISSUE-123"
```

For macOS (Darwin) browser window with issue would be opened for you.

## TODO

1. Do something with password
2. Make transition types configurable
