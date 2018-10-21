package values

var VersionText string = "Hinoshiba(c) giss command v0.0.1"
var HelpText string = "---------------------------------------------------------\n" + VersionText + `

how to use : giss <subcommand> ...
---------------------------------------------------------
Advance preparation
 checkin			It targets the issue of the remote repository in the current directory.
 checkout			Clear the information acquired with checkin.
 login				Login to current git server and get token.
 status				Display the status of remote repository and login.

Issues operation
 ls [ -a ] [ -l <limit cnt> ]	Display the current issues.
 					-l <limit> : Specify the maximum display line number. By default, 20 lines.
					-a         : Also displays closed issues. By default, only open is displayed.
 less  <issues No>		Display the specified issues detail at the iss,
 add   <issues No>		Add a comment to the end at the specified issues.
 edit  <issues No>		Edit a title and description at the specified issues.
 close <issues No>		Change to the close status at the specified issues.
 open  <issues No>		Change to the open status at the specified issues.

Advance operation
 repo				A mail is automatically generated in which addresses, headers, etc. are automatically inserted.
 				Must be set in advance in "~/.gissrc ".
---------------------------------------------------------

Temporary files are managed under "~/.giss/".

If you have any problems, please contact to <https://github.com/hinoshiba/giss/>.
Whether to create a description is undecided.
---------------------------------------------------------
`