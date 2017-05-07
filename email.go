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
func SendEmail(subject, body, address string, sender map[string]interface{}) error {
	var senderEmail, senderPassword, host, port string
	if sender == nil {
		return errors.New("missing sender information")
	}
	if sender["SenderEmail"] != nil {
		senderEmail = sender["SenderEmail"].(string)
	} else {
		return errors.New("missing SenderEmail parameter")
	}

	if sender["SenderPassword"] != nil {
		senderPassword = sender["SenderPassword"].(string)
	} else {
		return errors.New("missing SenderPassword parameter")
	}
	if sender["Host"] != nil {
		host = sender["Host"].(string)
	} else {
		return errors.New("missing Host parameter")
	}

	if sender["Port"] != nil {
		port = sender["Port"].(string)
	} else {
		return errors.New("missing Port parameter")
	}

	// Set up authentication information.
	auth := smtp.PlainAuth("", senderEmail, senderPassword, host)

	// Connect to the server, authenticate, set the sender and reciepent,
	// and send the email all in one step.
	to := []string{address}
	msg := []byte("To: " + address + "\r\n" +
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
