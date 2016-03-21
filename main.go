package main

import (
	"flag"

	"github.com/allanliu/easylogger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func init() {
	flag.Parse()
	easylogger.InitializeLog()
}

func main() {
	var (
		svc    *svcEC2
		params *ec2.CreateImageInput
		resp   string
		err    error
	)
	if *instanceID == "" {
		panic("Must provide InstanceID")
	}
	if *imageName == "" || len([]rune(*imageName)) < 4 {
		panic("Must provide image Name at least 4 characters in length")
	}
	svc = &svcEC2{
		svc: ec2.New(session.New(), &aws.Config{Region: aws.String(*awsRegion)}),
		imageNameWithoutTimestamp: *imageName,
		imageName:                 createNameWithTimestamp(*imageName),
		timeToSave:                *timeToSave,
		filter:                    getFilter(),
	}
	params = &ec2.CreateImageInput{
		Name:        aws.String(svc.imageName),
		InstanceId:  aws.String(*instanceID),
		Description: aws.String("This is a test"),
		DryRun:      aws.Bool(false),
	}
	resp, err = svc.createImage(params)
	easylogger.Log("Create Successful...")
	easylogger.LogFatal(err)
	easylogger.Log("Success with message: ", resp)
}
