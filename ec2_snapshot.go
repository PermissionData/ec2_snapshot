package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/allanliu/easylogger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"gopkg.in/yaml.v2"
)

var (
	awsRegion = flag.String(
		"aws-region",
		"us-east-1",
		"AWS region where instance lives",
	)
	instanceID = flag.String(
		"instance-id",
		"",
		"Instance Id to be copied.  Will crash if not provided.",
	)
	imageName = flag.String(
		"image-name",
		"",
		"Name of backed up image.  Will crash if not provided.",
	)
	timeToSave = flag.Int64(
		"time-to-save",
		604800,
		"Seconds of backups to save",
	)
	configLocation = flag.String(
		"config",
		"./config.yml",
		"Full or relative path to config location",
	)
)

type config struct {
	Filters []struct {
		Key    string   `yaml:"key"`
		Values []string `yaml:"values"`
	} `yaml:"filters"`
}

type deleteError struct {
	imageName string
	msg       string
}

type svcEC2 struct {
	svc                       ec2iface.EC2API
	imageNameWithoutTimestamp string
	imageName                 string
	newImageID                string
	timeToSave                int64
	filter                    []*ec2.Filter
}

func (e *deleteError) Error() string {
	return fmt.Sprintf(
		"Image delete failed for image %s with \"%s\"",
		e.imageName,
		e.msg,
	)
}

func (s *svcEC2) createImage(
	imageMeta *ec2.CreateImageInput,
) (string, error) {
	var (
		outputData *ec2.CreateImageOutput
		err        error
	)
	outputData, err = s.svc.CreateImage(imageMeta)
	easylogger.LogFatal(err)
	s.newImageID = *outputData.ImageId
	if err := s.removeOldImage(
		*outputData.ImageId,
	); err != nil {
		return "", &deleteError{*imageMeta.Name, err.Error()}
	}
	return *outputData.ImageId, nil
}

func (s *svcEC2) removeOldImage(newImageID string) error {
	var (
		resp *ec2.DescribeImagesOutput
		err  error
	)
	resp, err = s.svc.DescribeImages(
		&ec2.DescribeImagesInput{Filters: s.filter},
	)
	if err != nil {
		return &deleteError{
			s.imageName,
			fmt.Sprintf("Failed to describe images with error %s", err.Error()),
		}
	}
	for _, image := range resp.Images {
		if s.newImageID != *image.ImageId && strings.Contains(
			*image.Name,
			s.imageNameWithoutTimestamp,
		) {
			imageCreationTime, timeFormatError := time.Parse(
				time.RFC3339,
				*image.CreationDate,
			)
			easylogger.LogFatal(timeFormatError)
			if time.Now().Unix()-imageCreationTime.Unix() > s.timeToSave {
				params := &ec2.DeregisterImageInput{
					ImageId: image.ImageId,
					DryRun:  aws.Bool(false),
				}
				_, err = s.svc.DeregisterImage(params)
				if err != nil {
					return &deleteError{
						*image.Name,
						fmt.Sprintf("Failed to deregister image b/c of %s", err.Error()),
					}
				}
				if err := s.deleteSnapshotByDescription(*image.ImageId); err != nil {
					return &deleteError{
						*image.Name,
						fmt.Sprintf(
							"Failed to delete snapshot for image b/c of %s",
							err.Error(),
						),
					}
				}
			}
		}
	}
	return nil
}

func (s *svcEC2) deleteSnapshotByDescription(imageID string) error {
	var (
		resp *ec2.DescribeSnapshotsOutput
		err  error
	)
	resp, err = s.svc.DescribeSnapshots(
		&ec2.DescribeSnapshotsInput{Filters: s.filter},
	)
	if err != nil {
		return &deleteError{
			s.imageName,
			fmt.Sprintf(
				"Could not get snapshot list for deletion with msg %s",
				err.Error(),
			),
		}
	}
	for _, snapshot := range resp.Snapshots {
		if snapshot.Description != nil && strings.Contains(
			*snapshot.Description,
			imageID,
		) {
			_, err = s.svc.DeleteSnapshot(
				&ec2.DeleteSnapshotInput{
					SnapshotId: snapshot.SnapshotId,
					DryRun:     aws.Bool(false),
				},
			)
			if err != nil {
				return &deleteError{*snapshot.Description, err.Error()}
			}
			return nil
		}
	}
	return nil
}

func getFilter() []*ec2.Filter {
	dump, err := ioutil.ReadFile(*configLocation)
	easylogger.LogFatal(err)
	c := config{}
	err = yaml.Unmarshal(dump, &c)
	easylogger.LogFatal(err)
	var result = []*ec2.Filter{}
	for _, f := range c.Filters {
		var values = []*string{}
		for _, val := range f.Values {
			values = append(values, aws.String(val))
		}
		result = append(
			result,
			&ec2.Filter{
				Name:   aws.String(f.Key),
				Values: values,
			},
		)
	}
	return result
}

func createNameWithTimestamp(s string) string {
	return fmt.Sprintf(
		"%s.%s",
		s,
		time.Now().Format("20060102150405"),
	)
}
