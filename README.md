# giss

* Generic command interface for various issues management systems.(さまざまな問題管理システム用の汎用コマンドインターフェイス)
* I aim to use various issues management systems with the same operation.(私は同じ操作でさまざまなシステムを使うことを目指しています。)

## Supported services(対応中のサービス)

* Github
	* Type parameter is 'Github'
* Github Enterprise
	* Type parameter is 'Github'
* Gitea
	* Type parameter is 'Gitea'
* Redmine
	* Type parameter is 'Redmine'


---

## how to setup

1. download at the thisrepository/bin/'YourExecutableFile'(このリポジトリのbin/あなたの実行形式 をダウンロードしてください)
2. Create a setting file in ~/.gissrc(~/.gissrcファイルを作成してください)
	* please check to : sample/.gissrc(sample/.gissrc を確認してください)
3. Type 'giss checkin' and press the Enter key('giss checkin)を入力し、実行してください

---

## running sample
```
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
* The Redmine is transforming expressions.(redmineは表現を変換しています。)
	* A 'Category' is 'Label'
	* A 'Tracker' is 'Milestone'
	* A 'Project' is 'Repository'
* For each status ('closed' and 'open'), automatically obtain and automatically input the top of the GUI setting.
* (各ステータス（'closed' と 'open'）については、GUI設定の先頭を自動的に取得して自動的に入力します。)

## Services you want to respond to...(対応したいサービス)

* Gitlab
* Gogs
* Track wiki

---
