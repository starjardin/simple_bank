package mail

import (
	"testing"

	"github.com/starjardin/simplebank/utils"
	"github.com/stretchr/testify/require"
)

func TestSendMailWithGmail(t *testing.T) {
	config, err := utils.LoadConfig("..")

	require.NoError(t, err)

	sender := NewGmailSender(
		config.EmailSenderName,
		config.EmailSenderAddress,
		config.EmailSenderPassword,
	)

	subject := "Test Email"

	content := `
		<h1>This is a test email</h1>
		<p>Sent from <strong>simplebank</strong> application.</p>
	`
	to := []string{"tantely.and@onja.org"}

	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
