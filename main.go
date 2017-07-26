package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsEC2 "github.com/aws/aws-sdk-go/service/ec2"
)

func getInstance() (*awsEC2.EC2, string, error) {
	session := awsSession.New()
	metadata := ec2metadata.New(session)
	region, err := metadata.Region()
	if err != nil {
		return nil, "", err
	}
	instanceID, err := metadata.GetMetadata("instance-id")
	if err != nil {
		return nil, "", err
	}
	return awsEC2.New(session, &aws.Config{Region: aws.String(region)}), instanceID, nil
}

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s volume-id aws-device-id [instance-device-to-wait]\n", os.Args[0])
		os.Exit(1)
		return
	}
	volumeID := os.Args[1]
	device := os.Args[2]

	if !strings.HasPrefix(volumeID, "vol-") {
		fmt.Fprintf(os.Stderr, "Usage: %s volume-id aws-device-id [instance-device-to-wait]\n", os.Args[0])
		os.Exit(1)
		return
	}

	ec2, instanceID, err := getInstance()
	if err != nil {
		panic(err)
	}
	_, err = ec2.AttachVolume(&awsEC2.AttachVolumeInput{
		Device:     aws.String(device),
		InstanceId: aws.String(instanceID),
		VolumeId:   aws.String(volumeID),
	})
	if err != nil {
		panic(err)
	}

	if len(os.Args) == 3 {
		// No need to wait for device
		return
	}
	// Wait (forever) for volume to attach
	waitForDevice := os.Args[3]
	for {
		if _, err := os.Lstat(waitForDevice); err == nil {
			return
		}
		time.Sleep(10 * time.Second)
	}
}
