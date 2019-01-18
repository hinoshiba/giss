package values

var Version = "[null]"
var VersionText string = "Hinoshiba(c) giss command "
var TermTitle = "[giss termwindow mode]  " + VersionText
var DevStr string
var HelpText string = `---------------------------------------------------------

how to use : giss <subcommand> ...
---------------------------------------------------------
Advance preparation

Credential & Repository operation
 checkin				Can choose the giss target repository.
 login					Login to current git server and get token.
 status					Display the status of target repository and login.

Issues operation
 create 				Create a issue.
 com   <issue No>			Add a comment to the end at the specified issues.

 edit  <issue No>			Edit a title and description at the specified issues.
 close <issue No>			Change to the close status at the specified issues.
 open  <issue No>			Change to the open status at the specified issues.

 ls [ -a ] [ -l <limit cnt> ]		Display the current issues.
 						-l <limit> : Specify the maximum display line number. By default, 20 lines.
						-a         : Also displays closed issues. By default, only open is displayed.
 show  <issues No>			Display the specified issues detail at the issues.


Milestone operation
 mlls					Display the name of the milestones.
 mlch  <issues No> <milestone name>	Change to the milestone name.
 mldel <issues No>			Unset the milestone.

Labels operation
 lbls					Display the name of the labels.
 lbadd <issue No> <label name>		Add to the label at selected issue.
 lbdel <issue No> <label name>		Delete to the label at selected issue.

Advance operation
 repo					(beta Function) A mail is automatically generated in which addresses, headers, etc. are automatically inserted.
 export [-a] <type>			(beta Function) export all of the issues at stdout.
						-a         : with export the closed issues.
						<type>     : json, xml.
 import [-a] <type>			(beta Function) export all of the issues at stdout.
 					 - can't import comment.
					 - label, milestone can not be imported depending on the situation.
						-a         : with export the closed issues.
						<type>     : json, xml.

---------------------------------------------------------

Must be setup in "~/.gissrc ".
Temporary files are managed under "~/.giss/".

If you have any problems, please contact to <https://github.com/hinoshiba/giss/>.
Whether to create a description is undecided.
---------------------------------------------------------
`

var StartTerm string = `readme

---
<q> to close this window.
---

<giss term> is under development function.
If you find a problem, please report to me at the github.
Issues : https://github.com/hinoshiba/giss/issues

I am aware of the fact that can make the source code more beautiful because it scribbled.
There is also recognition that you need to streamline http requests and make types more efficient.
I do not anticipate reference errors etc.., so I'd like you to tell me.

---
Please <q> to close this window.
---
`

var HelpTerm string = `Basic operation
 q                      quit active window. if there is only one active window, exit.
 <Esc>                  exit giss term.
 j or Ctrl + N          move to up.
 k or Ctrl + P          move to down.
 G                      move to bottom.
 g                      move to top.

issue list operation
 $                      connect to the server and get the latest issues.
 <Eneter>               display the issue detail at current line.
 n                      create new issue.
 c                      comment the issue at current line.
 O                      change to 'open' at current line.
 C                      change to 'closed' at current line.
 M                      the milestone selection screen becomes active.
                          by enter a <space> you can select a milestone.
 L                      the label selection screen becomes active.
                          by enter a <space> you can select a label.

---------------------------------------------------------
 - Must be setup in "~/.gissrc ".
   Temporary files are managed under "~/.giss/".

 - There is also an interactive mode.
   if you want to use interactive mode , please enter the 'giss help' at your prompt.

If you have any problems, please contact to <https://github.com/hinoshiba/giss/>.
Whether to create a description is undecided.
---------------------------------------------------------
`

func init() {
	if DevStr != "" {
		Version += "." + DevStr
	}

	VersionText += Version
	TermTitle += Version
}
