# giss

* Generic command interface for various issues management systems.
* I aim to use various issues management systems with the same operation.

## giss term demo
<img src="https://raw.githubusercontent.com/wiki/hinoshiba/giss/img/demo.gif" width="620" />

## Supported services (type)

* Github (Github)
* Github Enterprise (Github)
* Gitea (Gitea)
* Redmine (Redmine)

---

## how to setup

1. download at [releases](https://github.com/hinoshiba/giss/releases)
2. create a `~/.gissrc`
	* please check to : `sample/.gissrc`
3. type to '`giss checkin`' and press the Enter key
4. enjoy :)

---

## interactive mode sample
```bash
# print issues list
[hinoshiba@wk01 giss]$ giss ls
   Id [ Milestone  ] Title
----------------------------

# create issue
[hinoshiba@wk01 giss]$ giss create #can use your favorite editor.
Title : test issue
edit option
        t: title, b: body
other option
        p: issue print, done: edit done
Please enter the menu (or cancel) >>done
done...
issue posted : #15

[hinoshiba@wk01 giss]$ giss ls
   Id [ Milestone  ] Title
----------------------------
  #15 [ Bug        ] asfads

# print issue detail
[hinoshiba@wk01 giss]$ giss show 15
# 15 : asfads
## ( New ) user user 2018-12-16 13:10:21 +0000 UTC comments(0)

# add issue comment
[hinoshiba@wk01 giss]$ giss com 15 #can use your favorite editor.
To continue press the enter key....
comment added : #15

[hinoshiba@wk01 giss]$ giss show 15
# 15 : asfads
## ( New ) user user 2018-12-16 13:10:21 +0000 UTC comments(0)

## Comment #34 user user 2018-12-16 13:10:39 +0000 UTC #########################

test comments

# close issue
[hinoshiba@wk01 giss]$ giss close 15
issue posted : #15
state updated : closed

[hinoshiba@wk01 giss]$ giss ls
   Id [ Milestone  ] Title
----------------------------

[hinoshiba@wk01 giss]$
```

---

# other

## Redmine
* The Redmine is transforming expressions.
	* A 'Category' is 'Label'
	* A 'Tracker' is 'Milestone'
	* A 'Project' is 'Repository'
* For each status ('closed' and 'open'), automatically obtain and automatically input the top of the GUI setting.

## Services you want to respond to...

* Gitlab
* Gogs
* Track wiki

---
