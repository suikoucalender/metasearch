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

message = """Dear User,

We are pleased to inform you that your analysis through MetaSearchDB has been successfully completed.
To review your results, please visit the following link: """ + result_url + """weighted_unifrac.output.html.

If you have any inquiries or require further assistance, do not hesitate to reach out. Our team is available to provide support and can be contacted directly at suikou-admin@googlegroups.com .
Thank you for choosing MetaSearchDB for your analytical needs. We look forward to assisting you with any future queries.

Best regards,

The MetaSearchDB Team
"""

send_mail("Completion of Your Analysis via MetaSearchDB: " + original_filename, message, email)
