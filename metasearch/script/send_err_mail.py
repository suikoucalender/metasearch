# coding: utf8
# Yahooからメールを送信するサンプル

import ssl
import smtplib
from email.mime.text import MIMEText
import sys
import json

# メールを送信
def send_mail(subject, message, user_email):
	json_open = open("config/config.json","r")
	json_load = json.load(json_open)

	from_addr = json_load["account"]["addr"]
	to_addr = user_email
	sender_name = json_load["account"]["addr"]
	passwd = json_load["account"]["pass"]
	
	msg = MIMEText(message)
	msg['Subject'] = subject
	msg['From'] = u'%s<%s>'%("MetaSearch Result",from_addr)
	msg['To'] = to_addr
	
	smtp = smtplib.SMTP_SSL("smtp.mail.yahoo.co.jp", 465, context=ssl.create_default_context())
	smtp.login(sender_name, passwd)
	smtp.sendmail(from_addr, to_addr, msg.as_string())
	smtp.quit()
 
argv = sys.argv
result_url = argv[1] #結果のURL
email = argv[2] #ユーザーのEmailアドレス
original_filename = argv[3] #元のファイル名

message = """This is metasearch.
Your analysis has been finished with an error about""" + original_filename + """

You can contact us via suikou_metasearch@yahoo.co.jp """

send_mail("Your analysis finished with an error: " + original_filename, message, email)
