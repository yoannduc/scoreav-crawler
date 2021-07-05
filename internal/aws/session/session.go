package session

import (
	goenv "github.com/Netflix/go-env"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type config struct {
	Env       string `env:"ENV,default=local"`
	ID        string `env:"AWS_ACCESS_KEY_ID,required=true"`
	AccessKey string `env:"AWS_SECRET_ACCESS_KEY,required=true"`
	SessToken string `env:"AWS_SESSION_TOKEN,required=true"`
	Region    string `env:"AWS_REGION,required=true"`
}

// The persistent session
var s *session.Session

// createSession creates an aws session and assign it to persistent session var
func createSession() error {
	var err error
	var c config
	if _, err = goenv.UnmarshalFromEnviron(&c); err != nil {
		return err
	}

	cfg := &aws.Config{
		Region: aws.String(c.Region),
	}
	if c.Env == "local" || c.Env == "dev" {
		cfg.Credentials = credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     c.ID,
			SecretAccessKey: c.AccessKey,
			SessionToken:    c.SessToken,
		})
	}
	// Real lambda on AWS do not need Credentials, but we need it in local to test
	if c.Env == "local" {
		cfg.Endpoint = aws.String("http://host.docker.internal:4566")
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	s, err = session.NewSession(cfg)
	if err != nil {
		return err
	}

	return nil
}

// GetSession gets the aws session object
//
// Create a session first if none was already instanciated
func GetSession() (*session.Session, error) {
	var err error
	if s == nil {
		err = createSession()
	}

	if err != nil {
		return nil, err
	}

	return s, nil
}
