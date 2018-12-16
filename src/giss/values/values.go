package values

var Version = "0.1.2"
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
 create 			Create a issue.
 com       <issues No>		Add a comment to the end at the specified issues.

 edit      <issues No>		Edit a title and description at the specified issues.
 close     <issues No>		Change to the close status at the specified issues.
 open      <issues No>		Change to the open status at the specified issues.

 ls [ -a ] [ -l <limit cnt> ]	Display the current issues.
 					-l <limit> : Specify the maximum display line number. By default, 20 lines.
					-a         : Also displays closed issues. By default, only open is displayed.
 show      <issues No>		Display the specified issues detail at the iss,

Advance operation
 repo				(beta Function) A mail is automatically generated in which addresses, headers, etc. are automatically inserted.
 				Must be set in advance in "~/.gissrc ".
---------------------------------------------------------

Temporary files are managed under "~/.giss/".

If you have any problems, please contact to <https://github.com/hinoshiba/giss/>.
Whether to create a description is undecided.
---------------------------------------------------------
`

func init() {
	if DevStr != "" {
		Version += " " + DevStr

	}
	VersionText += Version
}
