package values

var Version = "0.2.0"
var VersionText string = "Hinoshiba(c) giss command v"
var DevStr string
var HelpText string = `---------------------------------------------------------

how to use : giss <subcommand> ...
---------------------------------------------------------
Advance preparation

Credential & Repository operation
 checkin			(beta Function) It targets the issue of the remote repository in the current directory.
 login				Login to current git server and get token.
 status				Display the status of remote repository and login.

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
 				Must be set in advance in "~/.gissrc ".
 export [-a] <type>			(test Function) export all of the issues at stdout.
						-a         : with export the closed issues.
						<type>     : json, xml.

---------------------------------------------------------

Temporary files are managed under "~/.giss/".

If you have any problems, please contact to <https://github.com/hinoshiba/giss/>.
Whether to create a description is undecided.
---------------------------------------------------------
`

func init() {
	if DevStr != "" {
		Version += "." + DevStr
	}

	VersionText += Version
}
