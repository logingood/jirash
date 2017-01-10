# Jira Shell
## Overview

Allows to query jira for tickets assigned to user:

- all
- todo 
- backlog
- create a weekly report

Also allows to create a JIRA issue in a configured project and issuetype. 

First run jirash will ask you to add username, password, projectId and issue type. 

## Install

Use go get:
````
go get github.com/murat1985/jirash
````
package would be installed into your $GOBIN path, e.g.: 
```
/usr/local/go/bin
```

## Configuration 
Configuration would be automatically generated on first run and stored at
```
~/.jirashell.json
```
Bare in mind your password from JIRA there is in clear text, if it is not acceptable for you
don't use this script.


## TODO

1. Still working on transitions (e.g. mark an issue as done)
2. Do something with password
