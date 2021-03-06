package util

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

// AwsUtil wraps the cli for functionality not in the sdk
type AwsUtil interface {
	Sync(source string, target string) error
}

type awsUtilImpl struct {
	Debug bool
}

// NewAwsUtil creates a new AwsUtil
func NewAwsUtil(debug bool) AwsUtil {
	return &awsUtilImpl{
		Debug: debug,
	}
}

func (u *awsUtilImpl) Sync(source string, target string) error {

	cmd := exec.Command("aws", "s3", "sync", "--delete", source, target)
	if u.Debug {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		log.Error("failed s3 sync")
		return err
	}

	return nil
}
