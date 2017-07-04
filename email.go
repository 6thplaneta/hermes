package hermes

import (
	"errors"
	"net/smtp"
)

/*
* sends email to the specified email address
* @param 	string		subject
* @param 	string 		body
* @param 	string		to adress
* @param 	map[string]interface{}		it should has SenderEmail, SenderPassword,Host,Port
* @return	bool 			determines if feild exists or not.
* @return	string 			type of feild
 */
func SendEmail(subject, body, address string, sender map[string]interface{}, html bool) error {
	var senderEmail, senderPassword, host, port string
	if sender == nil {
		return errors.New("missing sender information")
	}
	if sender["sender_email"] != nil {
		senderEmail = sender["sender_email"].(string)
	} else {
		return errors.New("missing SenderEmail parameter")
	}

	if sender["sender_password"] != nil {
		senderPassword = sender["sender_password"].(string)
	} else {
		return errors.New("missing SenderPassword parameter")
	}
	if sender["host"] != nil {
		host = sender["host"].(string)
	} else {
		return errors.New("missing Host parameter")
	}

	if sender["port"] != nil {
		port = sender["port"].(string)
	} else {
		return errors.New("missing Port parameter")
	}

	// Set up authentication information.
	auth := smtp.PlainAuth("", senderEmail, senderPassword, host)

	// Connect to the server, authenticate, set the sender and reciepent,
	// and send the email all in one step.
	mime := ""
	if html {
		mime = "MIME-Version: 1.0" + "\r\n" +
			"Content-type: text/html" + "\r\n"
	}
	to := []string{address}
	msg := []byte("To: " + address + "\r\n" +
		mime +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")
	err := smtp.SendMail(host+":"+port, auth, senderEmail, to, msg)
	if err != nil {
		application.Logger.Error(err.Error())
		return err

	}
	return nil
}
